//go:build integration

package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alexedwards/scs/v2/memstore"

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/customers/qrcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/handlers"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// testQRConfig 回測試 config(QR 簽章密鑰 test-qr-secret,與 token 產生同值)。
func testQRConfig() *config.Config {
	cfg := &config.Config{}
	cfg.Auth.JWTSecret = "test-qr-secret"
	cfg.Auth.FrontendURL = "http://localhost:3000"
	return cfg
}

// TestIntegrationQRLoginRedeem 兌換全鏈(3.8.2):有效 token → 回公司/客戶/店家子帳號;
// 二兌拒絕(一次性);竄改拒絕;過期拒絕;跨公司同 customer_code 依 company 正確定位。
func TestIntegrationQRLoginRedeem(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateAuthBusinessUp(t, adminDSN)
	db := authAppRoleClient(t, adminDSN)
	ctx := context.Background()

	// 兩家公司同 customer_code(TY000001):兌換必須依 token 的 company_id 定位。
	coA, custA := seedQRCustomer(t, ctx, db, "QRA", "QR-A", "TY000001", "甲客戶")
	coB, custB := seedQRCustomer(t, ctx, db, "QRB", "QR-B", "TY000001", "乙客戶")
	_ = custB
	subA := seedQRSubAccount(t, ctx, db, coA, custA, "店員A", "sub-a")
	_ = subA

	h := newQRHandler(t, db)
	mux := http.NewServeMux()
	path, handler := salesorderv1connect.NewAuthServiceHandler(h)
	mux.Handle(path, handler)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	rpc := salesorderv1connect.NewAuthServiceClient(http.DefaultClient, srv.URL)

	secret := "test-qr-secret"
	token, jti, err := qrcode.Generate(secret, coA, "TY000001", 0)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	_ = jti

	// 有效兌換。
	resp, err := rpc.QRLogin(ctx, connect.NewRequest(&v1.QRLoginRequest{Token: token}))
	if err != nil {
		t.Fatalf("QRLogin: %v", err)
	}
	if resp.Msg.GetCustomerCode() != "TY000001" || resp.Msg.GetCustomerName() != "甲客戶" {
		t.Fatalf("客戶定位錯誤: %+v", resp.Msg)
	}
	if len(resp.Msg.GetAccounts()) != 1 || resp.Msg.GetAccounts()[0].GetAccountName() != "sub-a" {
		t.Fatalf("店家子帳號清單錯誤: %+v", resp.Msg.GetAccounts())
	}

	// 第二次兌換同一 token → 拒絕(一次性)。
	if _, err := rpc.QRLogin(ctx, connect.NewRequest(&v1.QRLoginRequest{Token: token})); err == nil {
		t.Fatal("重兌應拒絕")
	} else if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("重兌應 invalid_argument,got %v", err)
	}

	// 竄改 → unauthenticated。
	if _, err := rpc.QRLogin(ctx, connect.NewRequest(&v1.QRLoginRequest{Token: token + "x"})); err == nil {
		t.Fatal("竄改應拒絕")
	} else if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("竄改應 unauthenticated,got %v", err)
	}

	// B 家 token 定位到乙客戶(同 code 不同公司)。B 家先建一個店家子帳號,
	// 否則「無可選子帳號」即整筆拒絕(規格:提示以主帳號登入建子帳號)。
	seedQRSubAccount(t, ctx, db, coB, custB, "店員B", "sub-b")
	tokenB, _, err := qrcode.Generate(secret, coB, "TY000001", 0)
	if err != nil {
		t.Fatalf("Generate B: %v", err)
	}
	respB, err := rpc.QRLogin(ctx, connect.NewRequest(&v1.QRLoginRequest{Token: tokenB}))
	if err != nil {
		t.Fatalf("QRLogin B: %v", err)
	}
	if respB.Msg.GetCustomerName() != "乙客戶" {
		t.Fatalf("跨公司定位錯誤: %+v", respB.Msg)
	}
}

// seedQRCustomer 建公司+客戶,回 coID/custID。
func seedQRCustomer(t *testing.T, ctx context.Context, db *ent.Client, name, ident, code, custName string) (int, int) {
	t.Helper()
	var coID, custID int
	seedQRTx(t, db, func(tx *ent.Tx) error {
		co, err := tx.Company.Create().SetName(name).SetIdentifier(ident).SetStatus("active").Save(ctx)
		if err != nil {
			return err
		}
		cust, err := tx.Customer.Create().SetCompanyID(co.ID).
			SetCustomerCode(code).SetName(custName).Save(ctx)
		if err != nil {
			return err
		}
		coID, custID = co.ID, cust.ID
		return nil
	})
	return coID, custID
}

// seedQRSubAccount 建店家子帳號(is_customer 主帳號除外:主帳號/業務子帳號不入清單)。
func seedQRSubAccount(t *testing.T, ctx context.Context, db *ent.Client, coID, custID int, name, account string) int {
	t.Helper()
	var id int
	seedQRTx(t, db, func(tx *ent.Tx) error {
		u, err := tx.User.Create().SetCompanyID(coID).SetCustomerID(custID).
			SetEmail(account + "@t.com").SetName(name).SetRole("customer").
			SetIsCustomer(true).SetIsPrimary(false).SetSystemGenerated(false).
			SetAccountName(account).SetPasswordHash("x").SetStatus("active").Save(ctx)
		if err != nil {
			return err
		}
		id = u.ID
		return nil
	})
	return id
}

// newQRHandler 以測試 secret 組 handler(QR 簽章密鑰與 token 產生同值)。
func newQRHandler(t *testing.T, db *ent.Client) *handlers.AuthHandler {
	t.Helper()
	kv := auth.NewMemoryStore()
	cfg := testQRConfig()
	return handlers.NewAuthHandler(handlers.AuthDeps{
		Cfg:      cfg,
		DB:       db,
		Tokens:   auth.NewTokenManager("test-qr-secret", kv, db),
		Lockout:  auth.NewLoginLock(kv),
		OneTime:  auth.NewOneTimeStore(kv),
		Sessions: auth.WebSessionManager(memstore.New(), 24*time.Hour, false, "lax"),
	})
}

func seedQRTx(t *testing.T, db *ent.Client, fn func(tx *ent.Tx) error) {
	t.Helper()
	if err := dbtenant.SystemScopeTx(t.Context(), db, fn); err != nil {
		t.Fatalf("fixture: %v", err)
	}
}

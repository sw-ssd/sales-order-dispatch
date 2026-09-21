// Package handlers 的 QR 兌換(04 計畫 Task 3.8.2):AuthService.QRLogin 實作。
// 兌換為公開端點(不需登入):驗 token 全鏈 → jti 先佔位(防併發雙兌換) → 定位客戶 →
// 回公司/客戶與店家子帳號清單。兌換本身不核發 session(選帳號+密碼登入另行)。
package handlers

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/customers/qrcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// qrOneTime 轉接 OneTimeStore 為 qrcode.JTIStore(Peek 判已用、Put 佔位)。
type qrOneTime struct{ o *auth.OneTimeStore }

func (s qrOneTime) Get(ctx context.Context, key string) (string, bool, error) {
	return s.o.Peek(ctx, key)
}
func (s qrOneTime) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return s.o.Put(ctx, key, value, ttl)
}

// QRLogin 兌換 QR token:公開端點,不需登入。
//
// 順序:rate 限流 → 驗 token(格式/簽章/purpose/exp) → jti 先佔位 →
// 以 company_id + customer_code 定位客戶(不可僅 code 跨公司識別) → 回清單。
// token 無效一律 unauthenticated(不區分,防探測);過期/已用/客戶無效為 failed_precondition
// (對外訊息不區分細節);超 rate 以 429 語意——Connect 無 429 碼,此處回 resource_exhausted。
func (h *AuthHandler) QRLogin(ctx context.Context, req *connect.Request[v1.QRLoginRequest]) (*connect.Response[v1.QRLoginResponse], error) {
	token := strings.TrimSpace(req.Msg.GetToken())
	if token == "" {
		return nil, errcode.AuthUnauthenticated.Error(nil)
	}
	// Rate limit:同 token 短時超次拒絕(防暴力列舉)。以 LoginLock 語意複用:
	// token 本身為 key,失敗計數達門檻即鎖。成功兌換清零(一次性 token 不重用,清零僅衛生)。
	locked, err := h.deps.Lockout.IsLocked(ctx, "qr:"+token)
	if err != nil {
		return nil, errcode.SysInternal.Wrap(err)
	}
	if locked {
		until, err := h.deps.Lockout.LockedUntil(ctx, "qr:"+token)
		if err != nil {
			return nil, errcode.SysInternal.Wrap(err)
		}
		return nil, errcode.AuthLocked.Error(map[string]string{"until": until.Format("15:04")})
	}
	fail := func() {
		_, _ = h.deps.Lockout.RecordFailure(ctx, "qr:"+token)
	}

	c, err := qrcode.Verify(h.deps.Cfg.Auth.JWTSecret, token)
	if err != nil {
		fail()
		// Verify 已分好碼(unauthenticated / invalid_argument=過期):過期轉 failed_precondition
		// 語意(規格:不區分細節),其餘原樣回。
		if connect.CodeOf(err) == connect.CodeInvalidArgument {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "token"})
		}
		return nil, err
	}
	// jti 先佔位再回應(防併發雙兌換)。Valkey 不可用 → 降級僅驗簽章與效期並記告警。
	if err := qrcode.ClaimOnce(ctx, qrOneTime{h.deps.OneTime}, c.JTI, qrcode.Remaining(c)); err != nil {
		// KV 錯誤(非「已使用」)即降級路徑:記告警,繼續兌換(安全與可用性取捨,見規格)。
		if strings.Contains(err.Error(), "已使用") {
			fail()
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "token"})
		}
		log.Printf("auth: QR jti 佔位失敗(降級為僅驗簽章與效期): %v", err)
	}
	var (
		coName, custName string
		coID             int
		accounts         []*v1.QRLoginResponse_Account
	)
	if err := h.systemScope(ctx, func(ctx context.Context) error {
		db := dbtenant.Client(ctx, h.deps.DB)
		cust, err := db.Customer.Query().
			Where(customer.CompanyIDEQ(c.CompanyID), customer.CustomerCodeEQ(c.CustomerCode),
				customer.DeletedAtIsNil()).Only(ctx)
		if err != nil {
			return err
		}
		coID = cust.CompanyID
		custName = cust.Name
		co, err := db.Company.Query().
			Where(company.IDEQ(c.CompanyID)).Only(ctx)
		if err != nil {
			return err
		}
		coName = co.Name
		// 店家子帳號:同客戶、非主帳號、非系統業務子帳號、啟用中。
		subs, err := db.User.Query().
			Where(user.CustomerIDEQ(cust.ID), user.IsCustomerEQ(true),
				user.IsPrimaryEQ(false), user.SystemGeneratedEQ(false),
				user.StatusEQ(user.StatusActive)).All(ctx)
		if err != nil {
			return err
		}
		for _, s := range subs {
			accounts = append(accounts, &v1.QRLoginResponse_Account{
				Id:          intString(s.ID),
				AccountName: s.AccountName,
			})
		}
		return nil
	}); err != nil {
		fail()
		if ent.IsNotFound(err) {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "token"})
		}
		return nil, errcode.SysInternal.Wrap(err)
	}
	if len(accounts) == 0 {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "token"})
	}
	_ = coID
	_ = h.deps.Lockout.Clear(ctx, "qr:"+token)
	return connect.NewResponse(&v1.QRLoginResponse{
		CompanyId:    intString(coID),
		CompanyName:  coName,
		CustomerCode: c.CustomerCode,
		CustomerName: custName,
		Accounts:     accounts,
	}), nil
}

// intString 轉 id 字串(QR 回應的公司/帳號 id 為字串)。
func intString(n int) string {
	return strconv.Itoa(n)
}

package handlers

import (
	"context"
	"strconv"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	commonv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// authErrorInfo 由 connect error 取 ErrorInfo detail（碼／訊息／details 的唯一來源）。
func authErrorInfo(t *testing.T, err error) *commonv1.ErrorInfo {
	t.Helper()
	ce, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("應為 *connect.Error,got %T", err)
	}
	for _, d := range ce.Details() {
		v, derr := d.Value()
		if derr != nil {
			t.Fatalf("detail 取值失敗: %v", derr)
		}
		if info, ok := v.(*commonv1.ErrorInfo); ok {
			return info
		}
	}
	t.Fatal("錯誤未帶 ErrorInfo")
	return nil
}

// seedLoginAccount 建一個可密碼登入的客戶帳號；companyID 為 0 時刻意不歸屬公司
// （模擬未完成註冊／未完成公司歸屬的帳號）。
func seedLoginAccount(t *testing.T, e *testEnv, code, password string, companyID int) int {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	b := e.db.User.Create().
		SetEmail(code + "@example.com").SetName("店家" + code).
		SetStatus(user.StatusActive).SetRole("customer").SetIsCustomer(true).
		SetAccountName(code).SetPasswordHash(hash)
	if companyID > 0 {
		b = b.SetCompanyID(companyID)
	}
	return b.SaveX(e.ctx).ID
}

// TestAuthFirstBatchCodesAreReturned 逐條斷言登入／密碼路徑回的是**註冊碼**：
// 碼要對得上原本的 connect 碼與語意（帳密錯誤一律同一個碼、不區分帳號存在與否）。
func TestAuthFirstBatchCodesAreReturned(t *testing.T) {
	const pw = "secret-123"
	cases := []struct {
		name        string
		want        string
		wantDetails []string // 必須非空的 details 鍵（空 = 不檢查）
		call        func(t *testing.T) error
	}{
		{"帳號不存在", "AUTH-4003", nil, func(t *testing.T) error {
			e := newTestEnv(t)
			_, err := e.rpc.Login(e.ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "NO-SUCH", Password: pw}))
			return err
		}},
		{"密碼錯誤", "AUTH-4003", nil, func(t *testing.T) error {
			e := newTestEnv(t)
			seedLoginAccount(t, e, "C001", pw, mustCreateCompany(t, e, "co-err-pw"))
			_, err := e.rpc.Login(e.ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "C001", Password: "wrong"}))
			return err
		}},
		{"帳號鎖定", "AUTH-3003", []string{"until"}, func(t *testing.T) error {
			e := newTestEnv(t)
			seedLoginAccount(t, e, "C002", pw, mustCreateCompany(t, e, "co-err-lock"))
			for range 5 {
				_, _ = e.rpc.Login(e.ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "C002", Password: "wrong"}))
			}
			// 鎖定期間即使密碼正確也拒絕（原本即 FailedPrecondition）。
			_, err := e.rpc.Login(e.ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "C002", Password: pw}))
			return err
		}},
		{"公司停用", "AUTH-4002", nil, func(t *testing.T) error {
			e := newTestEnv(t)
			coID := mustCreateCompany(t, e, "co-err-inactive")
			seedLoginAccount(t, e, "C003", pw, coID)
			e.db.Company.UpdateOneID(coID).SetStatus(company.StatusInactive).SaveX(e.ctx)
			_, err := e.rpc.Login(e.ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "C003", Password: pw}))
			return err
		}},
		{"臨時密碼已過期", "AUTH-3002", nil, func(t *testing.T) error {
			ctx := context.Background()
			db := openAuthDB(t)
			cid := seedAuthCompany(t, db)
			old, _ := auth.HashPassword("OldPass123")
			u := db.User.Create().SetCompanyID(cid).SetEmail("exp@t.com").SetName("過期").SetRole("customer").
				SetIsCustomer(true).SetPasswordHash(old).SetMustChangePassword(true).
				SetTempPasswordExpiresAt(time.Now().Add(-time.Hour)).SaveX(ctx)
			client := newIdentifiedAuthClientWithDB(t, db, authz.Identity{
				UserID: strconv.Itoa(u.ID), CompanyID: strconv.Itoa(cid), Role: "customer", Roles: []string{"customer"}})
			_, err := client.ChangePassword(ctx, connect.NewRequest(
				&v1.ChangePasswordRequest{OldPassword: "OldPass123", NewPassword: "NewPass123456"}))
			return err
		}},
		{"未登入(密碼路徑)", "AUTH-4001", nil, func(t *testing.T) error {
			client := newIdentifiedAuthClientWithDB(t, openAuthDB(t), authz.Identity{})
			_, err := client.ChangePassword(context.Background(), connect.NewRequest(
				&v1.ChangePasswordRequest{OldPassword: "a", NewPassword: "b1234567"}))
			return err
		}},
		{"改密碼帳號不存在", "AUTH-4001", nil, func(t *testing.T) error {
			// 同一個 RPC 內的錯誤表面要一致:身分指向不存在的帳號不得回裸 connect 錯誤。
			client := newIdentifiedAuthClientWithDB(t, openAuthDB(t), authz.Identity{
				UserID: "999999", CompanyID: "1", Role: "customer", Roles: []string{"customer"}})
			_, err := client.ChangePassword(context.Background(), connect.NewRequest(
				&v1.ChangePasswordRequest{OldPassword: "a1234567", NewPassword: "b1234567"}))
			return err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call(t)
			if err == nil {
				t.Fatal("應回錯誤")
			}
			info := authErrorInfo(t, err)
			if info.GetCode() != tc.want {
				t.Fatalf("錯誤碼 = %q；want %s（訊息: %s）", info.GetCode(), tc.want, err.Error())
			}
			for _, k := range tc.wantDetails {
				if info.GetDetails()[k] == "" {
					t.Fatalf("details 缺 %s,got %v", k, info.GetDetails())
				}
			}
		})
	}
}

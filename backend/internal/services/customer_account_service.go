// CustomerAccountService:店家自助管理登入帳號(D22/規格 4.2,Task 6.7)。
//
// 為何獨立成一支服務:主帳號被 OpenFGA 排除掉**全部業務能力**(ability 的 primary_account
// exclusion,見 third_party/openfga modelDSL),所以它必須有自己的可達面,否則「主帳號登入後
// 什麼都做不了」——帳號管理正是它唯一被允許的功能(規格 4.2)。
//
// 三道界線:
//   - 只有**客戶主帳號**(is_customer && is_primary)可呼叫;子帳號與員工一律 permission_denied。
//   - 範圍僅限**自己客戶**(self);帳號 ID 即使外洩也跨不出去(查詢以 customer_id 收斂)。
//   - 自動附帶的**業務子帳號**(system_generated=true)可檢視但不可管理(店家並無其密碼,
//     它專供所屬業務使用);後台走 UserService 另有逃生門(規格 4.2 的後台移交)。
package services

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1/customersv1connect"
)

// CustomerAccountService 實作 customers.v1.CustomerAccountService。
type CustomerAccountService struct {
	db *ent.Client
	customersv1connect.UnimplementedCustomerAccountServiceHandler
}

// NewCustomerAccountService 建立 CustomerAccountService。
func NewCustomerAccountService(db *ent.Client) *CustomerAccountService {
	return &CustomerAccountService{db: db}
}

// RegisterCustomerAccountService 掛載店家自助帳號管理(Task 6.7)。
func RegisterCustomerAccountService(mux *http.ServeMux, db *ent.Client) {
	path, handler := customersv1connect.NewCustomerAccountServiceHandler(
		NewCustomerAccountService(db), connect.WithInterceptors(requestid.Interceptor(), dbtenant.Interceptor(db)))
	mux.Handle(path, handler)
}

// primaryCustomerScope 解析並驗證「呼叫者是客戶主帳號」,回傳 (companyID, customerID, userID)。
//
// 為何身分判斷用 id.Role 而非 hasRole:角色繼承展開讓每個後台角色都含 customer
// (company_admin→dept_admin→staff→customer),用展開集會把員工也放進來(同 orderScope 的
// 1ee57cb 迴歸)。主帳號則以 DB 的 is_primary 為準 —— 身分不帶該欄位,不從 token 推測。
func primaryCustomerScope(ctx context.Context, db *ent.Client) (cid, custID, uid int, err error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return 0, 0, 0, err
	}
	if id.Role != "customer" || strings.TrimSpace(id.CustomerID) == "" {
		return 0, 0, 0, errcode.SysPermissionDenied.Error(nil)
	}
	cid, err = parseID(id.CompanyID)
	if err != nil {
		return 0, 0, 0, errcode.SysPermissionDenied.Error(nil)
	}
	custID, err = parseID(id.CustomerID)
	if err != nil {
		return 0, 0, 0, errcode.SysPermissionDenied.Error(nil)
	}
	uid, err = parseID(id.UserID)
	if err != nil {
		return 0, 0, 0, errcode.SysPermissionDenied.Error(nil)
	}
	me, qerr := db.User.Query().Where(user.IDEQ(uid)).Only(ctx)
	if qerr != nil {
		return 0, 0, 0, toConnectError(qerr)
	}
	if !me.IsCustomer || !me.IsPrimary {
		// 子帳號無帳號管理權限(規格 4.2)。
		return 0, 0, 0, errcode.SysPermissionDenied.Error(nil)
	}
	return cid, custID, uid, nil
}

// customerAccountToProto 轉換帳號列;manageable 由呼叫端判定(主帳號恆 false、業務子帳號 false)。
func customerAccountToProto(u *ent.User, manageable bool) *customersv1.CustomerAccount {
	return &customersv1.CustomerAccount{
		Id:              strconv.Itoa(u.ID),
		AccountName:     u.AccountName,
		IsPrimary:       u.IsPrimary,
		SystemGenerated: u.SystemGenerated,
		Manageable:      manageable,
		Status:          string(u.Status),
	}
}

// ListCustomerAccounts 列出自己客戶底下的**全部**帳號(含主帳號與業務子帳號,供 UI 灰化顯示)。
func (s *CustomerAccountService) ListCustomerAccounts(ctx context.Context, _ *connect.Request[customersv1.ListCustomerAccountsRequest]) (*connect.Response[customersv1.ListCustomerAccountsResponse], error) {
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	cid, custID, _, err := primaryCustomerScope(ctx, db)
	if err != nil {
		return nil, err
	}
	rows, err := db.User.Query().
		Where(user.CustomerIDEQ(custID), user.HasCompanyWith(company.IDEQ(cid)), user.IsCustomerEQ(true)).
		Order(ent.Asc(user.FieldID)).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	out := make([]*customersv1.CustomerAccount, 0, len(rows))
	for _, u := range rows {
		// 可管理者＝自建的子帳號(非主帳號、非系統附帶)。業務子帳號店家可檢視但不可管理。
		out = append(out, customerAccountToProto(u, !u.IsPrimary && !u.SystemGenerated))
	}
	return connect.NewResponse(&customersv1.ListCustomerAccountsResponse{Accounts: out}), nil
}

// CreateCustomerAccount 新增子帳號(is_primary=false、24h 臨時密碼、首登強改)。
func (s *CustomerAccountService) CreateCustomerAccount(ctx context.Context, req *connect.Request[customersv1.CreateCustomerAccountRequest]) (*connect.Response[customersv1.CreateCustomerAccountResponse], error) {
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	cid, custID, actor, err := primaryCustomerScope(ctx, db)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Msg.GetAccountName())
	if name == "" {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "account_name"})
	}
	// account_name 在客戶內唯一(規格 4.2);先查再寫,並由 DB 的部分唯一索引兜底(競態)。
	dup, err := db.User.Query().Where(user.CustomerIDEQ(custID), user.AccountNameEQ(name)).
		Exist(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if dup {
		return nil, errcode.SysConflict.Error(map[string]string{"field": "account_name"})
	}
	temp, err := auth.GenerateTempPassword()
	if err != nil {
		return nil, errcode.SysInternal.Wrap(err)
	}
	hash, err := auth.HashPassword(temp)
	if err != nil {
		return nil, errcode.SysInternal.Wrap(err)
	}
	exp := time.Now().UTC().Add(customerTempPasswordTTL)
	// email 佔位沿用建檔慣例(user.email 全域唯一;客戶帳號以 account_name 登入,email 僅佔位)。
	created, err := buildCustomerAccount(ctx, db, accountSpec{
		CompanyID: cid, DepartmentID: nil, CustomerID: custID,
		Email:       newCustomerAccountEmail(custID, name),
		Name:        name,
		AccountName: name,
		IsPrimary:   false, SystemGenerated: false,
		PasswordHash: hash, MustChange: true, TempExpiresAt: exp,
	})
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := recordAudit(ctx, tx, "user", "create", created.ID, cid, nil, actor,
		map[string]any{"account_name": name, "account_type": "customer_sub"}); err != nil {
		return nil, errcode.SysInternal.Wrap(err)
	}
	return connect.NewResponse(&customersv1.CreateCustomerAccountResponse{
		Account:       customerAccountToProto(created, true),
		TempPassword:  temp,
		TempExpiresAt: exp.Format(time.RFC3339),
	}), nil
}

// DeactivateCustomerAccount 停用子帳號(規格 4.2 防呆:主帳號與業務子帳號不可由店家停用)。
func (s *CustomerAccountService) DeactivateCustomerAccount(ctx context.Context, req *connect.Request[customersv1.DeactivateCustomerAccountRequest]) (*connect.Response[customersv1.DeactivateCustomerAccountResponse], error) {
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	cid, custID, actor, err := primaryCustomerScope(ctx, db)
	if err != nil {
		return nil, err
	}
	target, err := s.loadOwnedAccount(ctx, db, req.Msg.GetAccountId(), cid, custID)
	if err != nil {
		return nil, err
	}
	if target.IsPrimary {
		// 防呆:主帳號不可由店家停用(含當前登入的帳號)——避免鎖死自己(規格 4.2)。
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"reason": "primary_account"})
	}
	if target.SystemGenerated {
		// 業務子帳號專供所屬業務,店家無其密碼 → 不可停用。
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"reason": "system_generated"})
	}
	if target.Status == user.StatusInactive {
		return nil, errcode.SysConflict.Error(map[string]string{"reason": "already_inactive"})
	}
	// 停用為關鍵操作:status + tv+1(在途 token 立即失效) + 稽核同交易(D18/D5)。
	if _, err := db.User.UpdateOneID(target.ID).
		SetStatus(user.StatusInactive).AddTokenVersion(1).Save(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := recordAuditBA(ctx, tx, "user", "update", target.ID, cid, nil, actor,
		map[string]any{"status": string(target.Status), "account_name": target.AccountName},
		map[string]any{"status": string(user.StatusInactive), "account_name": target.AccountName}); err != nil {
		return nil, errcode.SysInternal.Wrap(err)
	}
	return connect.NewResponse(&customersv1.DeactivateCustomerAccountResponse{}), nil
}

// ResetCustomerAccountPassword 重置子帳號密碼(重新核發 24h 臨時密碼 + 解除鎖定 + tv+1)。
func (s *CustomerAccountService) ResetCustomerAccountPassword(ctx context.Context, req *connect.Request[customersv1.ResetCustomerAccountPasswordRequest]) (*connect.Response[customersv1.ResetCustomerAccountPasswordResponse], error) {
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	cid, custID, actor, err := primaryCustomerScope(ctx, db)
	if err != nil {
		return nil, err
	}
	target, err := s.loadOwnedAccount(ctx, db, req.Msg.GetAccountId(), cid, custID)
	if err != nil {
		return nil, err
	}
	if target.IsPrimary || target.SystemGenerated {
		// 主帳號不可由店家重置(需後台逃生門);業務子帳號店家無權重置(規格 4.2)。
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"reason": "not_manageable"})
	}
	temp, err := auth.GenerateTempPassword()
	if err != nil {
		return nil, errcode.SysInternal.Wrap(err)
	}
	hash, err := auth.HashPassword(temp)
	if err != nil {
		return nil, errcode.SysInternal.Wrap(err)
	}
	exp := time.Now().UTC().Add(customerTempPasswordTTL)
	// 與後台重置(AuthService.ResetCustomerPassword)同語意:換臨時密碼 + 首登強改 + tv+1
	// (既有 token 立即失效)。登入鎖定存於 Valkey(以 account_name 為鍵,見 auth.LoginLock),
	// 不在 users 表;解鎖由登入路徑的 Lockout 負責,本服務不持有該依賴故不在此處理。
	if _, err := db.User.UpdateOneID(target.ID).
		SetPasswordHash(hash).SetMustChangePassword(true).SetTempPasswordExpiresAt(exp).
		AddTokenVersion(1).Save(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := recordAuditBA(ctx, tx, "user", "update", target.ID, cid, nil, actor,
		map[string]any{"account_name": target.AccountName, "password_reset": true},
		map[string]any{"account_name": target.AccountName, "password_reset": true}); err != nil {
		return nil, errcode.SysInternal.Wrap(err)
	}
	return connect.NewResponse(&customersv1.ResetCustomerAccountPasswordResponse{
		TempPassword:  temp,
		TempExpiresAt: exp.Format(time.RFC3339),
	}), nil
}

// loadOwnedAccount 取「自己客戶底下」的帳號;跨客戶一律 not_found(不以 permission_denied 洩漏存在性)。
func (s *CustomerAccountService) loadOwnedAccount(ctx context.Context, db *ent.Client, idStr string, cid, custID int) (*ent.User, error) {
	aid, err := parseID(idStr)
	if err != nil {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "account_id"})
	}
	u, err := db.User.Query().Where(
		user.IDEQ(aid), user.HasCompanyWith(company.IDEQ(cid)), user.CustomerIDEQ(custID), user.IsCustomerEQ(true),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.SysNotFound.Error(nil)
		}
		return nil, toConnectError(err)
	}
	return u, nil
}

// newCustomerAccountEmail 產生客戶帳號的佔位 email(user.email 全域唯一;登入以 account_name 為準)。
// 以 customer_id + 隨機後綴確保同一客戶可建多個同名不同次帳號(account_name 唯一性另由查詢+索引把關)。
func newCustomerAccountEmail(custID int, accountName string) string {
	suffix := strings.NewReplacer(" ", "-", "@", "-", ".", "-").Replace(accountName)
	return "customer." + strconv.Itoa(custID) + "." + suffix + "." + strconv.FormatInt(time.Now().UnixNano()%1_000_000, 10) + "@system.local"
}

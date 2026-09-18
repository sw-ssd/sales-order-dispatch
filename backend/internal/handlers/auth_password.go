// A3 臨時密碼與首登強制修改(01-auth 1.5.2 / 1.5.4)。
// ChangePassword:登入態改密碼(must_change=true 時唯一可用 RPC);驗舊→強度→更新+清受限態+
// token_version+1 撤銷舊憑證→稽核(D18)同一交易。
// ResetCustomerPassword:dept_admin+ 為客戶帳號重發臨時密碼。
package handlers

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// tempPasswordTTL 為臨時密碼效期(now+24h, 1.5.2)。
const tempPasswordTTL = 24 * time.Hour

// minNewPasswordLen 為新密碼最短長度(1.5.2 強度 ≥ 8)。
const minNewPasswordLen = 8

// ChangePassword 修改密碼:驗舊密碼 → 新密碼 ≥ 8 字元 → 更新雜湊、清 must_change 與臨時效期、
// token_version+1(所有既有 session/憑證失效,以新密碼重登)→ 稽核(action update,不含密碼)同一交易。
func (h *AuthHandler) ChangePassword(ctx context.Context, req *connect.Request[v1.ChangePasswordRequest]) (*connect.Response[v1.ChangePasswordResponse], error) {
	id := authz.IdentityFrom(ctx)
	if id.UserID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	uid, err := parseID(id.UserID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	old, newPw := req.Msg.GetOldPassword(), req.Msg.GetNewPassword()
	if old == "" || newPw == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("舊密碼與新密碼不可為空"))
	}
	if len(newPw) < minNewPasswordLen {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("新密碼至少 8 字元"))
	}
	u, err := h.deps.DB.User.Get(ctx, uid)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("帳號不存在"))
		}
		return nil, internal(err)
	}
	// 臨時密碼過期:must_change 且已過效期 → 僅能由管理員重置(1.5.2)。
	if u.MustChangePassword && u.TempPasswordExpiresAt != nil && time.Now().After(*u.TempPasswordExpiresAt) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("臨時密碼已過期,請聯繫管理員重置"))
	}
	if !auth.VerifyPassword(u.PasswordHash, old) {
		return nil, invalidCredentials()
	}
	hash, err := auth.HashPassword(newPw)
	if err != nil {
		return nil, internal(err)
	}

	tx, err := h.deps.DB.Tx(ctx)
	if err != nil {
		return nil, internal(err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.User.UpdateOneID(uid).
		SetPasswordHash(hash).
		SetMustChangePassword(false).
		ClearTempPasswordExpiresAt().
		AddTokenVersion(1).
		Save(ctx); err != nil {
		return nil, internal(err)
	}
	if err := auditChangePassword(ctx, tx, id, uid); err != nil {
		return nil, internal(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, internal(err)
	}
	return connect.NewResponse(&v1.ChangePasswordResponse{}), nil
}

// ResetCustomerPassword 重發臨時密碼(1.5.4):super 不限 / company_admin 限公司 / dept_admin 限部門;
// 目標須為客戶帳號;同一交易更新 hash+受限態+token_version+1+稽核;明文僅本次回應不落盤。
func (h *AuthHandler) ResetCustomerPassword(ctx context.Context, req *connect.Request[v1.ResetCustomerPasswordRequest]) (*connect.Response[v1.ResetCustomerPasswordResponse], error) {
	id := authz.IdentityFrom(ctx)
	if id.UserID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	targetID, err := parseID(req.Msg.GetUserId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("目標 user_id 格式錯誤"))
	}
	target, err := h.deps.DB.User.Query().Where(user.ID(targetID)).WithCompany().WithDepartment().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("目標帳號不存在"))
		}
		return nil, internal(err)
	}
	if !target.IsCustomer {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("目標非客戶帳號"))
	}
	if err := h.resetScopeOK(id, target); err != nil {
		return nil, err
	}

	temp, err := auth.GenerateTempPassword()
	if err != nil {
		return nil, internal(err)
	}
	hash, err := auth.HashPassword(temp)
	if err != nil {
		return nil, internal(err)
	}
	exp := time.Now().UTC().Add(tempPasswordTTL)

	tx, err := h.deps.DB.Tx(ctx)
	if err != nil {
		return nil, internal(err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.User.UpdateOneID(targetID).
		SetPasswordHash(hash).
		SetMustChangePassword(true).
		SetTempPasswordExpiresAt(exp).
		AddTokenVersion(1).
		Save(ctx); err != nil {
		return nil, internal(err)
	}
	// 稽核公司歸屬取「目標」公司(非操作者公司),符稽核追溯語意(super 跨公司重置時)。
	tgtCompanyID := 0
	if target.Edges.Company != nil {
		tgtCompanyID = target.Edges.Company.ID
	}
	_, actor := auditActor(id)
	if err := auditResetPassword(ctx, tx, actor, tgtCompanyID, targetID); err != nil {
		return nil, internal(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, internal(err)
	}
	return connect.NewResponse(&v1.ResetCustomerPasswordResponse{TempPassword: temp, ExpiresAt: exp.Unix()}), nil
}

// auditChangePassword 於交易內寫一筆「改密碼」稽核(不含密碼, D18)。
func auditChangePassword(ctx context.Context, tx *ent.Tx, id authz.Identity, uid int) error {
	cid, actor := auditActor(id)
	meta := audit.MetaFrom(ctx)
	return audit.Record(ctx, tx, audit.Entry{
		Action:       "update",
		ResourceType: "user",
		ResourceID:   strconv.Itoa(uid),
		CompanyID:    cid,
		UserID:       actor,
		After:        map[string]any{"must_change_password": false},
		IPAddress:    meta.IP,
		UserAgent:    meta.UserAgent,
	})
}

// auditResetPassword 於交易內寫一筆「重置密碼」稽核(操作者 + 目標,不含密碼)。
// companyID 為「目標」所屬公司(稽核追溯對象);actor 為操作者。
func auditResetPassword(ctx context.Context, tx *ent.Tx, actor, companyID, targetID int) error {
	meta := audit.MetaFrom(ctx)
	return audit.Record(ctx, tx, audit.Entry{
		Action:       "update",
		ResourceType: "user",
		ResourceID:   strconv.Itoa(targetID),
		CompanyID:    companyID,
		UserID:       actor,
		After:        map[string]any{"must_change_password": true, "force_reset": true},
		IPAddress:    meta.IP,
		UserAgent:    meta.UserAgent,
	})
}

// auditActor 由身分取稽核 company / actor id(缺則 0,由 audit.Record 拒寫,符合 D18)。
func auditActor(id authz.Identity) (int, int) {
	cid, _ := parseID(id.CompanyID)
	actor, _ := parseID(id.UserID)
	return cid, actor
}

// resetScopeOK 檢查重置範圍:super/developer 不限;company_admin 限同公司;dept_admin 限同部門。
func (h *AuthHandler) resetScopeOK(id authz.Identity, target *ent.User) error {
	if slices.Contains(id.Roles, "super") || slices.Contains(id.Roles, "developer") {
		return nil
	}
	if slices.Contains(id.Roles, "company_admin") {
		actCID, _ := parseID(id.CompanyID)
		tgtCID := 0
		if target.Edges.Company != nil {
			tgtCID = target.Edges.Company.ID
		}
		if actCID == 0 || actCID != tgtCID {
			return connect.NewError(connect.CodePermissionDenied, errors.New("無權限重置此帳號(跨公司)"))
		}
		return nil
	}
	if slices.Contains(id.Roles, "dept_admin") {
		actDID, _ := parseID(id.DepartmentID)
		tgtDID := 0
		if target.Edges.Department != nil {
			tgtDID = target.Edges.Department.ID
		}
		if actDID == 0 || actDID != tgtDID {
			return connect.NewError(connect.CodePermissionDenied, errors.New("無權限重置此帳號(跨部門)"))
		}
		return nil
	}
	return connect.NewError(connect.CodePermissionDenied, errors.New("無重置密碼權限"))
}

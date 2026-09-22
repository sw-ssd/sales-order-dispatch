// 公司 Logo 上傳端點(04 計畫 Task 8 / 細部 2.4.1):POST /api/v1/companies/{company_id}/logo。
// 權限為 company_admin 且限所屬公司(spec 3.1.1 修訂;super/dept_admin/staff 回 permission_denied);
// 存取面複用 FileStore:
// 白名單三重驗證 → 落盤 → 同交易建 file_assets(含檔稽核) → 更新 companies.logo_url → 主檔稽核。
package fileassets

import (
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
)

// logo 上傳/更換公司 Logo(細部 2.4.1,權限修訂:見 spec 3.1.1)。契約:
//   - 上傳者為 company_admin,且**只能上傳自己公司**;super 不經此端點(權限為 company_admin
//     專屬,與 company_admin 可編輯所屬公司識別同語意)。權限先於存在性:非 company_admin
//     恆為 permission_denied(403),不洩漏公司存在與否;
//   - 目標不存在/他公司/已軟刪統一 not_found(404,阻探測);他公司由 target != cid 擋下,
//     與 checkOwner 的 company 分支(ownerID 必須等於自己 cid)同語意;
//   - 同交易:檔記錄 + logo_url + 主檔稽核同成功同失敗(D18);落盤在交易外先行,
//     DB 失敗由 SaveUpload 刪孤兒檔。
func (h *Handler) logo(w http.ResponseWriter, r *http.Request) {
	id := authz.IdentityFrom(r.Context())
	if len(id.Roles) == 0 {
		writeErr(w, r, errcode.AuthUnauthenticated.Error(nil))
		return
	}
	target, err := strconv.Atoi(r.PathValue("company_id"))
	if err != nil || target <= 0 {
		writeErr(w, r, errcode.SysInvalidArgument.Error(map[string]string{"field": "company_id"}))
		return
	}
	cid, did, err := scopeOf(id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if id.Role != "company_admin" {
		writeErr(w, r, errcode.SysPermissionDenied.Error(nil))
		return
	}
	if target != cid {
		writeErr(w, r, notFound())
		return
	}
	// 上限在**讀取階段**就生效:ParseMultipartForm 的參數是記憶體門檻,超量會 spool 到
	// 暫存檔(白名單要等 SaveUpload 的 LimitReader 才擋),故先截斷 body。
	r.Body = http.MaxBytesReader(w, r.Body, MaxImageBytes+(1<<20))
	if err := r.ParseMultipartForm(MaxImageBytes + (1 << 20)); err != nil {
		writeErr(w, r, errcode.SysInvalidArgument.Error(map[string]string{"field": "file"}))
		return
	}
	f, hdr, err := r.FormFile("file")
	if err != nil {
		writeErr(w, r, errcode.SysInvalidArgument.Error(map[string]string{"field": "file"}))
		return
	}
	defer func() { _ = f.Close() }()
	// Logo 只收圖片:副檔名先擋 PDF 等;宣告 MIME ↔ 副檔名 ↔ magic bytes 三重驗證仍在
	// SaveUpload 的白名單內(偽裝副檔名在此之後被檔頭檢查擋下)。
	switch strings.ToLower(filepath.Ext(hdr.Filename)) {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		writeErr(w, r, errcode.SysInvalidArgument.Error(map[string]string{"field": "file"}))
		return
	}
	declared := strings.TrimSpace(r.FormValue("mime_type"))
	if declared == "" {
		declared = mime.TypeByExtension(strings.ToLower(filepath.Ext(hdr.Filename)))
	}

	// 請求交易:同 upload,REST 無 interceptor,手動開交易注入 ctx(RLS 由驅動裝飾器套用)。
	tx, ctx, err := h.tenantTx(r)
	if err != nil {
		writeErr(w, r, errcode.SysInternal.Error(nil))
		return
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	db := tx.Client()

	// 目標公司須存在且未軟刪除(範圍已鎖自己公司,不可見即不存在)。
	prev, err := db.Company.Query().Where(company.ID(target), company.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			writeErr(w, r, notFound())
			return
		}
		writeErr(w, r, errcode.SysInternal.Wrap(err))
		return
	}
	saved, err := h.store.SaveUpload(ctx, id, cid, did, "company", target, f, declared, hdr.Filename)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if _, err := db.Company.UpdateOneID(target).SetLogoURL(saved.URL).Save(ctx); err != nil {
		writeErr(w, r, toConnectErr(err))
		return
	}
	actor, _ := parseActor(id.UserID)
	if err := audit.Record(ctx, tx, audit.Entry{
		Action: "update", ResourceType: "company", ResourceID: strconv.Itoa(target),
		CompanyID: cid, DepartmentID: did, UserID: actor,
		Before:    map[string]any{"logo_url": prev.LogoURL},
		After:     map[string]any{"logo_url": saved.URL},
		IPAddress: audit.MetaFrom(ctx).IP,
		UserAgent: audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		writeErr(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		writeErr(w, r, errcode.SysInternal.Error(nil))
		return
	}
	committed = true
	writeJSON(w, http.StatusCreated, map[string]any{
		"id": saved.ID, "url": saved.URL, "filename": saved.Filename,
		"original_filename": saved.OriginalFilename, "mime_type": saved.MIME, "size_bytes": saved.Size,
	})
}

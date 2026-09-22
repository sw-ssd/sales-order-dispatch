// Package fileassets 的 HTTP 端點(04 計畫 Task 3.6.1/3.6.3):上傳與下載。
// 掛載於 /api/v1 之下(租戶 session + authzMiddleware 身分)。
// 上傳:POST /api/v1/files(multipart:file/owner_type/owner_id);下載:GET /api/v1/files/{id}/download。
package fileassets

import (
	"context"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/ent/fileasset"
	"github.com/salesorder/sales-order-1.0/backend/ent/product"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/resterr"
)

// Handler 為檔案端點依賴(Store + ent client)。
type Handler struct {
	store *Store
	db    *ent.Client
}

// NewHandler 建立 Handler。
func NewHandler(db *ent.Client, root string) *Handler {
	return &Handler{store: NewStore(db, root), db: db}
}

// RegisterRoutes 掛上傳、下載、軟刪除與公司 Logo 路由(呼叫端在 /api/v1 下掛載,此處用相對路徑)。
// apiMux 為標準 ServeMux(Connect handler 同器):REST 四條用方法+路徑模板掛載(Go 1.22+)。
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /files", h.upload)
	mux.HandleFunc("GET /files/{id}/download", h.download)
	mux.HandleFunc("DELETE /files/{id}", h.remove)
	mux.HandleFunc("POST /companies/{company_id}/logo", h.logo)
}

// scopeOf 依身分推導範圍(super/company_admin → 公司層;dept_admin/staff → 本部門;
// customer/guest 無主檔權限,與 deptScope 同語意)。
func scopeOf(id authz.Identity) (int, *int, error) {
	switch {
	case isSuper(id), id.Role == "company_admin":
		cid, err := parseActor(id.CompanyID)
		if err != nil {
			return 0, nil, errcode.SysPermissionDenied.Error(nil)
		}
		return cid, nil, nil
	case id.Role == "dept_admin" || id.Role == "staff":
		cid, err := parseActor(id.CompanyID)
		if err != nil {
			return 0, nil, errcode.SysPermissionDenied.Error(nil)
		}
		did, err := parseActor(id.DepartmentID)
		if err != nil {
			return 0, nil, errcode.SysPermissionDenied.Error(nil)
		}
		return cid, &did, nil
	default:
		return 0, nil, errcode.SysPermissionDenied.Error(nil)
	}
}

// isSuper 含 super/developer 逃生門語意。
func isSuper(id authz.Identity) bool {
	for _, r := range id.Roles {
		if r == "super" || r == "developer" {
			return true
		}
	}
	return id.Role == "super" || id.Role == "developer"
}

// upload 處理 multipart 上傳:需登入;owner 關聯驗證(存在+同租戶)。
func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	id := authz.IdentityFrom(r.Context())
	if len(id.Roles) == 0 {
		writeErr(w, r, errcode.AuthUnauthenticated.Error(nil))
		return
	}
	cid, did, err := scopeOf(id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if err := r.ParseMultipartForm(MaxPDFBytes + (1 << 20)); err != nil {
		writeErr(w, r, errcode.SysInvalidArgument.Error(map[string]string{"field": "file"}))
		return
	}
	f, hdr, err := r.FormFile("file")
	if err != nil {
		writeErr(w, r, errcode.SysInvalidArgument.Error(map[string]string{"field": "file"}))
		return
	}
	defer func() { _ = f.Close() }()
	ownerType := strings.TrimSpace(r.FormValue("owner_type"))
	ownerID, err := parseActor(strings.TrimSpace(r.FormValue("owner_id")))
	if err != nil {
		writeErr(w, r, errcode.SysInvalidArgument.Error(map[string]string{"field": "owner_id"}))
		return
	}
	// 宣告 MIME:先取表單欄位,缺省由副檔名推斷(仍須過白名單+magic 雙檢)。
	declared := strings.TrimSpace(r.FormValue("mime_type"))
	if declared == "" {
		declared = mime.TypeByExtension(strings.ToLower(filepath.Ext(hdr.Filename)))
	}
	// owner 關聯驗證:存在且同租戶。
	if err := h.checkOwner(r.Context(), cid, did, ownerType, ownerID); err != nil {
		writeErr(w, r, err)
		return
	}
	// 請求交易:DB 寫入與稽核同交易(比照 Connect 的 dbtenant.Interceptor;REST 無 interceptor,
	// 此處手動開交易並注入 ctx;driver 裝飾器在 Tx(ctx) 內套用 RLS)。
	tx, err := h.db.Tx(r.Context())
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
	ctx := dbtenant.WithTenantTx(r.Context(), tx)
	saved, err := h.store.SaveUpload(ctx, id, cid, did, ownerType, ownerID, f, declared, hdr.Filename)
	if err != nil {
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

// checkOwner 驗 owner 存在且同租戶(公司/客戶/商品三類;其餘 owner_type 拒絕,避免孤兒關聯)。
func (h *Handler) checkOwner(ctx context.Context, cid int, did *int, ownerType string, ownerID int) error {
	switch ownerType {
	case "company":
		ok, err := h.db.Company.Query().Where(company.ID(ownerID)).Exist(ctx)
		if err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		if !ok {
			return errcode.SysInvalidArgument.Error(map[string]string{"field": "owner_id"})
		}
		// 公司 owner 須為本公司。
		if ownerID != cid {
			return errcode.SysPermissionDenied.Error(nil)
		}
		return nil
	case "customer":
		c, err := h.db.Customer.Query().
			Where(customer.ID(ownerID), customer.DeletedAtIsNil()).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return errcode.SysInvalidArgument.Error(map[string]string{"field": "owner_id"})
			}
			return errcode.SysInternal.Wrap(err)
		}
		if c.CompanyID != cid {
			return errcode.SysPermissionDenied.Error(nil)
		}
		return nil
	case "product":
		p, err := h.db.Product.Query().
			Where(product.ID(pidOf(ownerID)), product.DeletedAtIsNil()).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return errcode.SysInvalidArgument.Error(map[string]string{"field": "owner_id"})
			}
			return errcode.SysInternal.Wrap(err)
		}
		if p.CompanyID != cid {
			return errcode.SysPermissionDenied.Error(nil)
		}
		return nil
	default:
		return errcode.SysInvalidArgument.Error(map[string]string{"field": "owner_type"})
	}
}

// pidOf 轉 ownerID 為 product 查詢鍵(同 parseActor 語意,此處僅轉型)。
func pidOf(n int) int { return n }

// download 串流下載:認證 + 公司/部門範圍 + 未軟刪除;一律 404 隱藏存在性。
func (h *Handler) download(w http.ResponseWriter, r *http.Request) {
	id := authz.IdentityFrom(r.Context())
	if len(id.Roles) == 0 {
		writeErr(w, r, errcode.AuthUnauthenticated.Error(nil))
		return
	}
	cid, did, err := scopeOf(id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	fid, err := parseActor(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, notFound())
		return
	}
	fa, err := scopeFileQuery(r.Context(), h.db, cid, did, fid)
	if err != nil {
		writeErr(w, r, notFound())
		return
	}
	abs := filepath.Join(h.store.root, filepath.Clean(fa.StoragePath))
	// 路徑穿越防線:Clean 後仍須落在 root 之下。
	if !strings.HasPrefix(abs, filepath.Clean(h.store.root)+string(filepath.Separator)) {
		writeErr(w, r, notFound())
		return
	}
	f, err := os.Open(abs)
	if err != nil {
		writeErr(w, r, notFound()) // 檔案遺失視同找不到(另計告警由呼叫端 log)
		return
	}
	defer func() { _ = f.Close() }()
	w.Header().Set("Content-Type", fa.MimeType)
	w.Header().Set("Content-Disposition", contentDisposition(fa.OriginalFilename))
	http.ServeContent(w, r, fa.Filename, fa.CreatedAt, f)
}

// contentDisposition 以 filename* 編碼原始檔名(中文檔名支援)。
func contentDisposition(original string) string {
	return "attachment; filename*=UTF-8''" + urlEscape(original)
}

// urlEscape 百分比編碼(空格編為 %20,非 QueryEscape 的 +)。
func urlEscape(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' || r == '~' {
			b.WriteRune(r)
			continue
		}
		for _, c := range []byte(string(r)) {
			b.WriteString("%" + strings.ToUpper(strconv.FormatInt(int64(c), 16)))
		}
	}
	return b.String()
}

// remove 軟刪除(管理用途):實體檔案保留,下載拒絕。
func (h *Handler) remove(w http.ResponseWriter, r *http.Request) {
	id := authz.IdentityFrom(r.Context())
	if len(id.Roles) == 0 {
		writeErr(w, r, errcode.AuthUnauthenticated.Error(nil))
		return
	}
	cid, did, err := scopeOf(id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	fid, err := parseActor(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, notFound())
		return
	}
	tx, err := h.db.Tx(r.Context())
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
	ctx := dbtenant.WithTenantTx(r.Context(), tx)
	db := tx.Client()
	fa, err := scopeFileQuery(ctx, db, cid, did, fid)
	if err != nil {
		writeErr(w, r, notFound())
		return
	}
	if _, err := db.FileAsset.UpdateOneID(fa.ID).SetDeletedAt(time.Now().UTC()).Save(ctx); err != nil {
		writeErr(w, r, errcode.SysInternal.Error(nil))
		return
	}
	actor, _ := parseActor(id.UserID)
	if err := recordAudit(ctx, tx, fa.ID, cid, did, actor); err != nil {
		writeErr(w, r, errcode.SysInternal.Error(nil))
		return
	}
	if err := tx.Commit(); err != nil {
		writeErr(w, r, errcode.SysInternal.Error(nil))
		return
	}
	committed = true
	writeJSON(w, http.StatusOK, map[string]any{"id": fa.ID})
}

// scopeFileQuery 範圍內查檔(未刪);查無一律回 err(呼叫端轉 404,不洩漏存在性)。
func scopeFileQuery(ctx context.Context, db *ent.Client, cid int, did *int, fid int) (*ent.FileAsset, error) {
	q := db.FileAsset.Query().Where(fileasset.ID(fid), fileasset.DeletedAtIsNil(),
		fileasset.CompanyIDEQ(cid))
	if did != nil {
		q = q.Where(fileasset.Or(fileasset.DepartmentIDIsNil(), fileasset.DepartmentIDEQ(*did)))
	}
	return q.Only(ctx)
}

// writeJSON/writeErr 委派 resterr:REST 錯誤協定的**單一來源**(server.writeConnectError 是
// Connect RPC 的同形實作)。檔名慣例保留,呼叫點不必知道實作落在哪個套件。
func writeJSON(w http.ResponseWriter, code int, v any) {
	resterr.JSON(w, code, v)
}

func writeErr(w http.ResponseWriter, r *http.Request, err error) {
	resterr.Write(w, r, err)
}

// notFound 統一 404(存在性隱藏)。
func notFound() *connect.Error {
	return errcode.SysNotFound.Error(nil)
}

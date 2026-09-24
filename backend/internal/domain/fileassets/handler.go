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
	"github.com/salesorder/sales-order-1.0/backend/ent/logisticsdelivery"
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

// tenantTx 開請求交易並把租戶 ctx 注入其中,回傳 tx 與租戶 ctx。呼叫端 defer tx.Rollback()
// (提交後再 rollback 是 no-op),成功路徑才 Commit。
//
// 為什麼每一條 REST 路徑都必須經過這裡:file_assets / companies 等表在 00036 / 00028 之後是
// ENABLE + FORCE ROW LEVEL SECURITY,而業務連線是 app_rw(非 superuser)——**沒有 SET LOCAL
// app.* 的連線上任何查詢都只看到 0 列**。Connect RPC 由 dbtenant.Interceptor 開交易;REST 端點
// 沒有 interceptor,漏開交易不會報錯,只會讓存在性檢查一律「查無」(download 404、checkOwner
// 判定 owner 不存在),且在 superuser 連線的整合測試下完全看不出來。
func (h *Handler) tenantTx(r *http.Request) (*ent.Tx, context.Context, error) {
	tx, err := h.db.Tx(r.Context())
	if err != nil {
		return nil, nil, err
	}
	return tx, dbtenant.WithTenantTx(r.Context(), tx), nil
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
	// 上限在讀取階段生效(同 logo:ParseMultipartForm 的參數只是記憶體門檻,超量會 spool 到暫存檔)。
	r.Body = http.MaxBytesReader(w, r.Body, MaxPDFBytes+(1<<20))
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
	// 請求交易:owner 驗證、DB 寫入與稽核必須**同一條租戶交易**——owner 驗證在交易外查詢時,
	// RLS 因無 app.* scope 而一律回 0 列(見 tenantTx 說明)。
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
	// owner 關聯驗證:存在且同租戶。
	if err := h.checkOwner(ctx, tx.Client(), cid, did, ownerType, ownerID); err != nil {
		writeErr(w, r, err)
		return
	}
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

// checkOwner 驗 owner 存在且同租戶(公司/客戶/商品/配送四類;其餘 owner_type 拒絕,避免孤兒關聯)。
// db **必須是租戶交易內的 client**(RLS:交易外的查詢一律 0 列 → 誤判 owner 不存在)。
func (h *Handler) checkOwner(ctx context.Context, db *ent.Client, cid int, did *int, ownerType string, ownerID int) error {
	switch ownerType {
	case "logistics_delivery":
		// POD 簽收檔(D32/10.6):只能掛在自己公司、且可見範圍內的配送執行單。
		d, err := db.LogisticsDelivery.Query().
			Where(logisticsdelivery.ID(ownerID), logisticsdelivery.DeletedAtIsNil()).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return errcode.SysInvalidArgument.Error(map[string]string{"field": "owner_id"})
			}
			return errcode.SysInternal.Wrap(err)
		}
		if d.CompanyID != cid {
			return errcode.SysPermissionDenied.Error(nil)
		}
		return nil
	case "company":
		ok, err := db.Company.Query().Where(company.ID(ownerID), company.DeletedAtIsNil()).Exist(ctx)
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
		c, err := db.Customer.Query().
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
		p, err := db.Product.Query().
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
	// 查詢必須在租戶交易內(RLS:交易外一律 0 列 → 存在性檢查誤判為 404)。
	tx, ctx, err := h.tenantTx(r)
	if err != nil {
		writeErr(w, r, errcode.SysInternal.Error(nil))
		return
	}
	fa, err := scopeFileQuery(ctx, tx.Client(), cid, did, r.PathValue("id"))
	_ = tx.Rollback() // 唯讀:不 commit
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
	fa, err := scopeFileQuery(ctx, db, cid, did, r.PathValue("id"))
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
//
// ref 為 {id} 路徑段:整數字串以 id 查,其餘視為系統檔名。**兩種形狀都必須支援** ——
// file_assets.url 是 uuid 檔名(SaveUpload 與 print_helpers 都這樣寫),而列印 API 另以
// 數字 id 組 download_url;呼叫端拿哪一種都必須服務得到,否則 DB 裡的 url 是指不到的。
func scopeFileQuery(ctx context.Context, db *ent.Client, cid int, did *int, ref string) (*ent.FileAsset, error) {
	q := db.FileAsset.Query().Where(fileasset.DeletedAtIsNil(), fileasset.CompanyIDEQ(cid))
	if n, err := strconv.Atoi(ref); err == nil && n > 0 {
		q = q.Where(fileasset.ID(n))
	} else {
		q = q.Where(fileasset.FilenameEQ(ref))
	}
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

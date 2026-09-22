package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/resterr"
)

// Me 為 GET /api/v1/me:回傳目前 session 身分與所屬公司的品牌識別(規格 §8.1 Web 側邊欄
// Logo 顯示;上傳鈕的 super 顯示開關也取自這裡的 role——前端只做顯示,後端仍是唯一決策者)。
// REST(D4)——不經 Connect interceptor,故自行檢查身分、並經 resterr 寫出錯誤協定。
// 公司以**身分**的 company_id 定錨(請求沒有任何可指定公司的參數),跨公司讀取構造不出來;
// 查無與已軟刪同形回 company=null(不區分,前端一律降級為預設圖示)。
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	id := authz.IdentityFrom(r.Context())
	if len(id.Roles) == 0 {
		resterr.Write(w, r, errcode.AuthUnauthenticated.Error(nil))
		return
	}
	var co map[string]any
	if cid, err := strconv.Atoi(id.CompanyID); err == nil && cid > 0 {
		//唯讀請求交易:RLS scope 由驅動裝飾器在 Tx 開立時 SET LOCAL(同 fileassets REST 慣例)。
		tx, terr := h.deps.DB.Tx(r.Context())
		if terr != nil {
			log.Printf("me: 開啟租戶交易失敗: %v", terr)
			resterr.Write(w, r, errcode.SysInternal.Error(nil))
			return
		}
		ctx := dbtenant.WithTenantTx(r.Context(), tx)
		c, qerr := tx.Client().Company.Query().Where(company.ID(cid), company.DeletedAtIsNil()).Only(ctx)
		_ = tx.Rollback() // 唯讀:不 commit
		switch {
		case qerr == nil:
			co = map[string]any{"id": strconv.Itoa(c.ID), "name": c.Name, "logo_url": c.LogoURL}
		case ent.IsNotFound(qerr):
			// 軟刪/不存在同形 → null(不洩漏差異)。
		default:
			log.Printf("me: 讀公司失敗: %v", qerr)
			resterr.Write(w, r, errcode.SysInternal.Error(nil))
			return
		}
	}
	resterr.JSON(w, http.StatusOK, map[string]any{
		"user_id": id.UserID, "role": id.Role, "company": co,
	})
}

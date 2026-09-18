// Package auth 提供 GetAbility RPC:以 OpenFGA ListObjects 列舉目前身分可用的
// (資源, 動作)能力集合,供前端選單/守衛驅動(D32 取代 CASL ability 下發)。
// developer 逃生門(開關啟用)直接回 manage:all。物件狀態條件由 domain 狀態機處理,
// 不進 GetAbility(OpenFGA 只管關係性/角色/租戶範圍)。
package auth

import (
	"context"
	"strings"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// Config 為 AbilityHandler 所需設定子集(server 組裝時由 config.API 注入)。
type Config struct {
	// DeveloperAccountEnabled 對應 config.API.DeveloperAccountEnabled:
	// 啟用時 developer 身分直接取得 manage-all。
	DeveloperAccountEnabled bool
}

// AbilityHandler 實作 salesorder.v1.AbilityService 的 GetAbility。
type AbilityHandler struct {
	db  *ent.Client
	cfg Config
}

// NewAbilityHandler 建立 GetAbility handler。
func NewAbilityHandler(db *ent.Client, cfg Config) *AbilityHandler {
	return &AbilityHandler{db: db, cfg: cfg}
}

// resourceOf 剝除 OpenFGA tuple object 的型別前綴(ability:company → company)。
func resourceOf(obj string) (string, bool) {
	if id, ok := strings.CutPrefix(obj, "ability:"); ok && id != "" {
		return id, true
	}
	return "", false
}

// GetAbility 以 OpenFGA ListObjects 列舉身分可讀/可寫資源,產出 (action, resource) 規則。
// developer(開關啟用)→ manage:all;未登入/engine 未注入 → 空(fail-closed)。
func (h *AbilityHandler) GetAbility(ctx context.Context, req *connect.Request[v1.GetAbilityRequest]) (*connect.Response[v1.GetAbilityResponse], error) {
	id := authz.IdentityFrom(ctx)
	if id.Role == "developer" && h.cfg.DeveloperAccountEnabled {
		return connect.NewResponse(&v1.GetAbilityResponse{
			Rules: []*v1.AbilityRule{{Action: "manage", Subject: "all"}},
		}), nil
	}
	e := authz.EngineFrom(ctx)
	if e == nil || id.UserID == "" {
		// 未注入引擎/未登入 → fail-closed:空能力。
		return connect.NewResponse(&v1.GetAbilityResponse{}), nil
	}
	user := "user:" + id.UserID
	// ability 型別於授權 model 帶 can_read/can_write。
	readObjs, err := e.ListObjects(ctx, user, "can_read", "ability")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	writeObjs, err := e.ListObjects(ctx, user, "can_write", "ability")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	rules := make([]*v1.AbilityRule, 0, len(readObjs)+len(writeObjs))
	seen := map[string]bool{}
	for _, o := range readObjs {
		res, ok := resourceOf(o)
		if !ok {
			continue
		}
		key := "read\x00" + res
		if seen[key] {
			continue
		}
		seen[key] = true
		rules = append(rules, &v1.AbilityRule{Action: "read", Subject: res})
	}
	for _, o := range writeObjs {
		res, ok := resourceOf(o)
		if !ok {
			continue
		}
		key := "write\x00" + res
		if seen[key] {
			continue
		}
		seen[key] = true
		rules = append(rules, &v1.AbilityRule{Action: "write", Subject: res})
	}
	return connect.NewResponse(&v1.GetAbilityResponse{Rules: rules}), nil
}

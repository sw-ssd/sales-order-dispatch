// Package auth 提供 GetAbility RPC:由角色權限規則表產生 CASL JSON 規則,
// 供前端 @casl/ability 初始化(casl-integration T9,D30-3/D30-9)。
package auth

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/ent/rolepermission"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz/casl"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// Config 為 AbilityHandler 所需設定子集(server 組裝時由 config.API 注入)。
type Config struct {
	// DeveloperAccountEnabled 對應 config.API.DeveloperAccountEnabled:
	// 啟用時 developer 身分直接取得 manage-all(跳過規則表)。
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

// GetAbility 由 ctx 身分查詢規則表,輸出依 sort_order 升冪的 CASL JSON 規則。
// 佔位符(如 ${user.department_id})以身分展開為具體值;展開失敗或解析失敗的規則
// fail-closed 不下發(前端亦無從命中)。developer(開關啟用)直接回 manage-all;
// guest 或無角色規則時回空。
func (h *AbilityHandler) GetAbility(ctx context.Context, req *connect.Request[v1.GetAbilityRequest]) (*connect.Response[v1.GetAbilityResponse], error) {
	id := authz.IdentityFrom(ctx)
	if id.Role == "developer" && h.cfg.DeveloperAccountEnabled {
		return connect.NewResponse(&v1.GetAbilityResponse{
			Rules: []*v1.AbilityRule{{Action: "manage", Subject: "all"}},
		}), nil
	}
	rows, err := h.loadRules(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := make([]*v1.AbilityRule, 0, len(rows))
	for _, r := range rows {
		if strings.HasPrefix(r.Subject, "\x00") {
			continue // casl 停用標記:佔位符展開失敗的規則不下發
		}
		rule := &v1.AbilityRule{Action: r.Action, Subject: r.Subject, Inverted: r.Inverted}
		if len(r.Conditions) > 0 {
			s, err := structpb.NewStruct(condsToMap(r.Conditions))
			if err != nil {
				continue // fail-closed:無法序列化 → 不下發
			}
			rule.Conditions = s
		}
		out = append(out, rule)
	}
	return connect.NewResponse(&v1.GetAbilityResponse{Rules: out}), nil
}

// loadRules 查詢身分全部角色的規則(role.CodeIn),依 sort_order 升冪,並以
// casl.NewEvaluator 展開佔位符(與執行層同源,避免雙份展開邏輯漂移)。
// 佔位符展開失敗的規則以 casl 停用標記(\x00 前綴 subject)標記,由 GetAbility 略過。
func (h *AbilityHandler) loadRules(ctx context.Context, id authz.Identity) ([]casl.Rule, error) {
	if h.db == nil || len(id.Roles) == 0 {
		return nil, nil
	}
	rows, err := h.db.RolePermission.Query().
		Where(rolepermission.HasRoleWith(role.CodeIn(id.Roles...))).
		Order(rolepermission.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	rules := make([]casl.Rule, 0, len(rows))
	for _, rp := range rows {
		conds, err := casl.ParseConditions(rp.Conditions)
		if err != nil {
			continue // fail-closed:壞規則視為不存在
		}
		rules = append(rules, casl.Rule{
			Action: rp.Action, Subject: rp.Resource, Conditions: conds, Inverted: rp.Inverted,
		})
	}
	return casl.NewEvaluator(rules, casl.Identity{
		UserID: id.UserID, CompanyID: id.CompanyID,
		DepartmentID: id.DepartmentID, CustomerID: id.CustomerID,
	}).Rules(), nil
}

// condsToMap 將欄位條件還原為 CASL JSON 形:每個運算子輸出 {op: value},
// 同欄位多運算子(含 $eq)合併為一個物件({$eq: v, ...}),@casl/ability 原生支援。
func condsToMap(conds []casl.FieldCondition) map[string]any {
	out := map[string]any{}
	for _, c := range conds {
		m, _ := out[c.Field].(map[string]any)
		if m == nil {
			m = map[string]any{}
			out[c.Field] = m
		}
		m[string(c.Op)] = c.Value
	}
	return out
}

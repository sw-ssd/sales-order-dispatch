// Package authz 提供 CASL 執行層的呼叫面 facade(D30-2/D30-3)。
// 開關狀態與身分由 middleware 注入 ctx(01 計畫 middleware 落地後接線,見 access_test 的 shim);
// repository/usecase 不直接讀 config。
package authz

import (
	"context"
	"sync"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/ent/rolepermission"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz/casl"
)

type ctxKey int

const (
	keyEnabled ctxKey = iota
	keyRegistry
	keyIdentity
	keyDB
)

// Identity 為 CASL 執行層所需的身分(01 middleware 落地後由 rls.Identity 轉換注入)。
type Identity struct {
	UserID       string
	CompanyID    string
	DepartmentID string
	CustomerID   string
	Role         string
	Roles        []string // 全部角色 code(依 Casbin g 規則展開)
}

// WithIdentity 將身分放入 ctx(測試與 middleware 使用)。
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, keyIdentity, id)
}

// IdentityFrom 由 ctx 取身分;未注入時回零值(規則載入會因空 roles 而全略過 → denied)。
func IdentityFrom(ctx context.Context) Identity {
	id, _ := ctx.Value(keyIdentity).(Identity)
	return id
}

// WithDB 將 ent client 放入 ctx(測試與 server 組裝使用)。
func WithDB(ctx context.Context, db *ent.Client) context.Context {
	return context.WithValue(ctx, keyDB, db)
}

func dbFrom(ctx context.Context) *ent.Client {
	db, _ := ctx.Value(keyDB).(*ent.Client)
	return db
}

// WithCASLEnabled 由 middleware 每請求呼叫一次,記錄開關狀態。
func WithCASLEnabled(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, keyEnabled, enabled)
}

func caslEnabled(ctx context.Context) bool {
	v, _ := ctx.Value(keyEnabled).(bool)
	return v
}

// Registry 回傳進程級 FieldRegistry(啟動時註冊全部 subject)。
func Registry(ctx context.Context) *casl.FieldRegistry {
	if r, ok := ctx.Value(keyRegistry).(*casl.FieldRegistry); ok {
		return r
	}
	return defaultRegistry()
}

var (
	registryOnce sync.Once
	registryInst *casl.FieldRegistry
)

func defaultRegistry() *casl.FieldRegistry {
	registryOnce.Do(func() {
		registryInst = casl.NewFieldRegistry()
		casl.RegisterSalesOrder(registryInst)
	})
	return registryInst
}

// loadRules 取當前身分全部角色的規則,SQL ORDER BY sort_order 升冪。
func loadRules(ctx context.Context, db *ent.Client, roles []string) ([]casl.Rule, error) {
	if db == nil || len(roles) == 0 {
		return nil, nil
	}
	rows, err := db.RolePermission.Query().
		Where(rolepermission.HasRoleWith(role.CodeIn(roles...))).
		Order(rolepermission.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	rules := make([]casl.Rule, 0, len(rows))
	for _, rp := range rows {
		conds, err := casl.ParseConditions(rp.Conditions)
		if err != nil {
			// fail-closed:壞規則視為不存在,記 security log(接入 logger 後補)
			continue
		}
		rules = append(rules, casl.Rule{
			Action: rp.Action, Subject: rp.Resource, Conditions: conds, Inverted: rp.Inverted,
		})
	}
	return rules, nil
}

func evaluator(ctx context.Context, db *ent.Client) (*casl.Evaluator, error) {
	id := IdentityFrom(ctx)
	rules, err := loadRules(ctx, db, id.Roles)
	if err != nil {
		return nil, err
	}
	return casl.NewEvaluator(rules, casl.Identity{
		UserID: id.UserID, CompanyID: id.CompanyID,
		DepartmentID: id.DepartmentID, CustomerID: id.CustomerID,
	}), nil
}

// AccessibleFilter 供 repository list 查詢套用。
// 開關關閉 → 空條件;無允許規則 → denied=true(短路回空集)。
func AccessibleFilter(ctx context.Context, action, subject string) (string, []any, bool) {
	if !caslEnabled(ctx) {
		return "", nil, false
	}
	db := dbFrom(ctx)
	e, err := evaluator(ctx, db)
	if err != nil {
		return "", nil, true // 規則載入失敗 fail-closed
	}
	clause, args, denied, err := casl.Translate(e.Rules(), action, subject, Registry(ctx))
	if err != nil {
		return "", nil, true
	}
	return clause, args, denied
}

// Can 供 usecase 對單筆實體做條件檢查。開關關閉 → true。
func Can(ctx context.Context, action, subject string, entity any) bool {
	if !caslEnabled(ctx) {
		return true
	}
	db := dbFrom(ctx)
	e, err := evaluator(ctx, db)
	if err != nil {
		return false
	}
	return e.Can(action, subject, Registry(ctx).Instance(subject, entity))
}

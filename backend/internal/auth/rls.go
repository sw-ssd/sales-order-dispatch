// RLS(T14):以 PostgreSQL SET LOCAL app.* 在交易內切換資料範圍,供 RLS policy 比對
// (設計書 §3.3;00002_rls_policies.sql 已啟用 row_security)。
// brief Step 3 明列 app.current_user_id / app.data_scope;設計書 §3.3 另列
// current_company_id / current_department_id / current_customer_id / current_data_scope ——
// 全部一併寫入(兩組命名皆覆蓋,避免 RLS policy 依任一命名)。
package auth

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// DataScope 資料範圍等級(對齊 roles.data_scope;設計書 §3.2/§3.3)。
type DataScope string

const (
	DataScopeAll        DataScope = "all"
	DataScopeCompany    DataScope = "company"
	DataScopeDepartment DataScope = "department"
	DataScopeSelf       DataScope = "self"
)

// RLSScope 單一請求的 RLS 身分與資料範圍。
type RLSScope struct {
	UserID       string
	CompanyID    string
	DepartmentID string
	CustomerID   string
	DataScope    DataScope
}

type rlsCtxKey struct{}

// WithRLS 將 RLS scope 放入 ctx(middleware 每請求注入一次;brief 介面)。
func WithRLS(ctx context.Context, scope RLSScope) context.Context {
	return context.WithValue(ctx, rlsCtxKey{}, scope)
}

// RLSFrom 由 ctx 取 RLS scope;未注入時回零值。
func RLSFrom(ctx context.Context) RLSScope {
	s, _ := ctx.Value(rlsCtxKey{}).(RLSScope)
	return s
}

// quoteSQL 以單引號包住值並逸出內嵌引號(值來自 session/DB 控制,非使用者輸入)。
func quoteSQL(v string) string {
	return "'" + strings.ReplaceAll(v, "'", "''") + "'"
}

// RLSStatements 產生 scope 的 SET LOCAL 語句(依固定順序,決定性;空值欄位略過)。
func RLSStatements(scope RLSScope) []string {
	var stmts []string
	if scope.UserID != "" {
		stmts = append(stmts, "SET LOCAL app.current_user_id = "+quoteSQL(scope.UserID))
	}
	if scope.CompanyID != "" {
		stmts = append(stmts, "SET LOCAL app.current_company_id = "+quoteSQL(scope.CompanyID))
	}
	if scope.DepartmentID != "" {
		stmts = append(stmts, "SET LOCAL app.current_department_id = "+quoteSQL(scope.DepartmentID))
	}
	if scope.CustomerID != "" {
		stmts = append(stmts, "SET LOCAL app.current_customer_id = "+quoteSQL(scope.CustomerID))
	}
	if scope.DataScope != "" {
		stmts = append(stmts, "SET LOCAL app.current_data_scope = "+quoteSQL(string(scope.DataScope)))
		// brief Step 3 命名別名(與設計書命名並存,相容任一 RLS policy)。
		stmts = append(stmts, "SET LOCAL app.data_scope = "+quoteSQL(string(scope.DataScope)))
	}
	return stmts
}

// Execer 抽象可執行 SQL 的介面(*sql.DB / *sql.Tx)。
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// ApplyRLS 於已開啟的交易/連線上執行 SET LOCAL app.* 切換(brief 介面)。
// SET LOCAL 僅作用於當前交易,故須於交易內呼叫;ent sqlite 測試不支援 SET 語法,
// 語句正確性由 RLSStatements 純函式測試覆蓋(Postgres 端由 02 計畫 RLS policy 驗收)。
func ApplyRLS(ctx context.Context, exec Execer, scope RLSScope) error {
	for _, stmt := range RLSStatements(scope) {
		if _, err := exec.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("rls: %s: %w", stmt, err)
		}
	}
	return nil
}

// ScopeForRole 依內建角色回傳預設資料範圍(設計書 §3.2/§4.4)。
// super/developer = all(繞過 RLS);company_admin = company;dept_admin/staff = department;
// customer/guest = self(客戶僅見自己資料);未知角色回傳空(不注入,依 DB 預設)。
func ScopeForRole(role string) DataScope {
	switch role {
	case "super", "developer":
		return DataScopeAll
	case "company_admin":
		return DataScopeCompany
	case "dept_admin", "staff":
		return DataScopeDepartment
	case "customer", "guest":
		return DataScopeSelf
	default:
		return ""
	}
}

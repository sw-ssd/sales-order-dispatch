package auth

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
)

// TestRLSStatements 驗收 brief Step 3:SET LOCAL app.* 語句產生(順序決定性、空值略過、引號逸出)。
func TestRLSStatements(t *testing.T) {
	scope := RLSScope{
		UserID:       "u1",
		CompanyID:    "c1",
		DepartmentID: "d1",
		DataScope:    DataScopeDepartment,
	}
	got := RLSStatements(scope)
	want := []string{
		"SET LOCAL app.current_user_id = 'u1'",
		"SET LOCAL app.current_company_id = 'c1'",
		"SET LOCAL app.current_department_id = 'd1'",
		"SET LOCAL app.current_data_scope = 'department'",
		"SET LOCAL app.data_scope = 'department'",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RLSStatements = %#v, want %#v", got, want)
	}

	// 空 scope → 無語句(不注入,依 DB 預設)。
	if got := RLSStatements(RLSScope{}); len(got) != 0 {
		t.Fatalf("空 scope 應無語句,got %#v", got)
	}

	// 引號逸出(值含單引號)。
	got = RLSStatements(RLSScope{UserID: "o'brien"})
	if len(got) != 1 || got[0] != "SET LOCAL app.current_user_id = 'o''brien'" {
		t.Fatalf("引號逸出失敗: %#v", got)
	}
}

// TestApplyRLS 以記錄型 executor 驗證 ApplyRLS 依序執行全部語句。
func TestApplyRLS(t *testing.T) {
	var executed []string
	exec := &recordingExec{fn: func(q string) { executed = append(executed, q) }}
	scope := RLSScope{UserID: "u1", CompanyID: "c1", DataScope: DataScopeCompany}
	if err := ApplyRLS(context.Background(), exec, scope); err != nil {
		t.Fatalf("ApplyRLS: %v", err)
	}
	want := []string{
		"SET LOCAL app.current_user_id = 'u1'",
		"SET LOCAL app.current_company_id = 'c1'",
		"SET LOCAL app.current_data_scope = 'company'",
		"SET LOCAL app.data_scope = 'company'",
	}
	if !reflect.DeepEqual(executed, want) {
		t.Fatalf("ApplyRLS 執行 = %#v, want %#v", executed, want)
	}

	// 失敗傳遞。
	boom := &recordingExec{fn: func(string) {}, err: context.DeadlineExceeded}
	if err := ApplyRLS(context.Background(), boom, scope); err == nil {
		t.Fatal("ApplyRLS 應回傳 executor 錯誤")
	}
}

// TestRLSContext WithRLS/RLSFrom 往返;未注入回零值。
func TestRLSContext(t *testing.T) {
	ctx := context.Background()
	if got := RLSFrom(ctx); got != (RLSScope{}) {
		t.Fatalf("未注入應回零值,got %#v", got)
	}
	scope := RLSScope{UserID: "u1", CompanyID: "c1", DataScope: DataScopeSelf}
	got := RLSFrom(WithRLS(ctx, scope))
	if got != scope {
		t.Fatalf("WithRLS/RLSFrom 往返失敗: got %#v want %#v", got, scope)
	}
}

// TestScopeForRole 內建角色資料範圍對映(設計書 3.2/4.4)。
func TestScopeForRole(t *testing.T) {
	cases := map[string]DataScope{
		"super":         DataScopeAll,
		"developer":     DataScopeAll,
		"company_admin": DataScopeCompany,
		"dept_admin":    DataScopeDepartment,
		"staff":         DataScopeDepartment,
		"customer":      DataScopeSelf,
		"guest":         DataScopeSelf,
		"custom_role":   "",
	}
	for role, want := range cases {
		if got := ScopeForRole(role); got != want {
			t.Errorf("ScopeForRole(%q) = %q, want %q", role, got, want)
		}
	}
}

// recordingExec 記錄 ExecContext 呼叫的語句(可注入錯誤)。
type recordingExec struct {
	fn  func(query string)
	err error
}

func (r *recordingExec) ExecContext(_ context.Context, query string, _ ...any) (sql.Result, error) {
	if r.fn != nil {
		r.fn(query)
	}
	return sqlResult{}, r.err
}

type sqlResult struct{}

func (sqlResult) LastInsertId() (int64, error) { return 0, nil }
func (sqlResult) RowsAffected() (int64, error) { return 0, nil }

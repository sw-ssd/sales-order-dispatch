//go:build integration

package services

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx database/sql driver

	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationEntitlementCounterCountsWithinRequestScope 計數器必須看得見**整個公司** ——
// 那是配額的單位（spec §3.2 的 WHERE company_id=?）。可見範圍由 RLS scope 決定，故如下的三分支
// 都是守衛會踩到的情境：
//
//	① 未帶請求交易／無 scope（無身分的系統路徑：OIDC 首次登入、RegisterComplete、CLI）
//	   → 走系統範圍交易（scope=all）拿到公司層真數字。**修前這裡回 0**：00028 的 FORCE RLS 把
//	   users 等表濾成 0 列 → used=0 → 任何上限都不觸發（守衛成了裝飾品）。
//	② 帶請求交易（scope=company A）→ 各 feature 拿到正確筆數（非 0），B 公司的列不計入。
//	③ 部門 scope（dept_admin／staff）→ 仍必須是公司總數（不得低報成「每部門一份」）。
//
// 為什麼一定要 app_rw：容器／測試的 admin 是 superuser，PG 的 superuser 永遠繞過 RLS
// （FORCE 亦然）→ 以 admin 連線計數「怎麼查都對」，測不出漏帶 scope 的實作。
// app_rw 是 00022 的 NOBYPASSRLS 業務角色，正是生產路徑的角色。
func TestIntegrationEntitlementCounterCountsWithinRequestScope(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	defer func() { _ = admin.Close() }()

	// scope=company 只需要公司列＋它自己的列；users/customers/products/departments 都受 RLS。
	coA := insertRLSCompany(t, admin, "A", "CNT-A")
	coB := insertRLSCompany(t, admin, "B", "CNT-B")
	deptA := insertCounterDepartment(t, admin, coA, "部門甲", false)
	insertCounterDepartment(t, admin, coA, "部門乙", true) // 軟刪除
	insertCounterDepartment(t, admin, coB, "部門丙", false)

	insertCounterUser(t, admin, coA, "a1@example.com", "active", 0)
	insertCounterUser(t, admin, coA, "a2@example.com", "pending", 0) // 非 inactive → 佔席位
	insertCounterUser(t, admin, coA, "a3@example.com", "inactive", 0)
	// 部門帳號：department scope 只看得到它，公司 scope 看得到全部 → 用來分辨「每部門上限」。
	insertCounterUser(t, admin, coA, "a4@example.com", "active", deptA)
	insertCounterUser(t, admin, coB, "b1@example.com", "active", 0)

	insertCounterCustomer(t, admin, coA, "CNT000001", false)
	insertCounterCustomer(t, admin, coA, "CNT000002", true) // 軟刪除
	insertCounterCustomer(t, admin, coB, "CNT000003", false)

	insertCounterProduct(t, admin, coA, deptA, "P-1", false)
	insertCounterProduct(t, admin, coA, deptA, "P-2", true) // 軟刪除
	insertCounterProduct(t, admin, coB, deptA, "P-3", false)

	client := dbtenant.NewClient(openAppRoleDB(t, adminDSN))
	t.Cleanup(func() { _ = client.Close() })
	counter := NewEntitlementCounter(client)

	features := []struct {
		feature string
		want    int
	}{
		{entitlements.LimitSeats, 3}, // active ＋ pending ＋ 部門 active（inactive 不佔）
		{entitlements.LimitCustomers, 1},
		{entitlements.LimitProducts, 1},
		{entitlements.LimitDepartments, 1},
	}

	t.Run("未帶請求交易(無 scope)→ 系統範圍交易,仍拿到公司層真數字", func(t *testing.T) {
		for _, tc := range features {
			got, err := counter.Count(context.Background(), coA, tc.feature)
			if err != nil {
				t.Fatalf("Count(%s): %v", tc.feature, err)
			}
			if got != tc.want {
				t.Fatalf("無 scope（無身分系統路徑）時 %s 應為公司層真數字 %d,got %d "+
					"(修前的實作在此回 0 → 守衛形同裝飾品)", tc.feature, tc.want, got)
			}
		}
		// 公司條件仍是**明示**的 company_id：他公司的列不得計入（系統範圍不等於全表亂數）。
		if got, err := counter.Count(context.Background(), coB, entitlements.LimitSeats); err != nil {
			t.Fatalf("Count(coB): %v", err)
		} else if got != 1 {
			t.Fatalf("無 scope 時仍只數該公司（B 應為 1），got %d", got)
		}
	})

	t.Run("請求範圍內(scope=company A)→ 正確數字", func(t *testing.T) {
		ctx := auth.WithRLS(context.Background(), auth.RLSScope{
			DataScope: auth.DataScopeCompany, CompanyID: itoa(coA), CompanyActive: true,
		})
		tx, err := client.Tx(ctx)
		if err != nil {
			t.Fatalf("開租戶交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		ctx = dbtenant.WithTenantTx(ctx, tx)

		for _, tc := range features {
			got, err := counter.Count(ctx, coA, tc.feature)
			if err != nil {
				t.Fatalf("Count(%s): %v", tc.feature, err)
			}
			if got != tc.want {
				t.Errorf("公司 A 的 %s 應為 %d,got %d", tc.feature, tc.want, got)
			}
		}
		// scope 只給 A：B 公司的列不得計入（計數是真的受限，不是無條件全表數）。
		got, err := counter.Count(ctx, coB, entitlements.LimitSeats)
		if err != nil {
			t.Fatalf("Count(coB): %v", err)
		}
		if got != 0 {
			t.Errorf("scope=company A 時不得看到公司 B 的帳號,got %d", got)
		}

		// 守衛在請求交易「之內」被呼叫，必須看見同一交易尚未提交的列（否則同一請求內
		// 連續建立時，後一筆的守衛看不到前一筆 → 超額）。這條同時釘住 Count 走的是
		// 請求交易本身，而不是另外開一條交易（後者因 ctx 仍帶 scope 也能查到已提交的列，
		// 只在這條會露餡）。
		if _, err := tx.Client().Customer.Create().
			SetCompanyID(coA).SetCustomerCode("CNT000009").SetName("同交易客戶").Save(ctx); err != nil {
			t.Fatalf("交易內建客戶: %v", err)
		}
		got, err = counter.Count(ctx, coA, entitlements.LimitCustomers)
		if err != nil {
			t.Fatalf("Count(同交易): %v", err)
		}
		if got != 2 {
			t.Errorf("同一請求交易內剛建立的客戶必須計入(1 筆種子 ＋ 1 筆未提交),got %d", got)
		}
	})

	// 配額是「公司層」的（spec §3.2：WHERE company_id=?），不是「每部門一份」。
	// dept_admin／staff 的請求 scope 是 department，而他們**可以建立**客戶／商品／使用者；
	// 若計數就著請求交易數，RLS 只會暴露本部門 → 配額被低報成每部門一份 → T6 掛上守衛後
	// 超額放行（fail-open）。實測（修前）：company scope 回 3、department scope 回 1。
	t.Run("部門 scope 的身分 → 仍必須拿到公司總數（配額是公司層）", func(t *testing.T) {
		ctx := auth.WithRLS(context.Background(), auth.RLSScope{
			DataScope: auth.DataScopeDepartment, CompanyID: itoa(coA), DepartmentID: itoa(deptA),
			CompanyActive: true,
		})
		tx, err := client.Tx(ctx)
		if err != nil {
			t.Fatalf("開租戶交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		ctx = dbtenant.WithTenantTx(ctx, tx)

		// 先確認這條 scope 真的只暴露本部門（否則本子測就沒有鑑別力）。
		deptOnly, err := tx.Client().User.Query().Count(ctx)
		if err != nil {
			t.Fatalf("部門 scope 下數使用者: %v", err)
		}
		if deptOnly != 1 {
			t.Fatalf("部門 scope 應只看得到 1 位部門帳號(測試前提),got %d", deptOnly)
		}

		for _, tc := range []struct {
			feature string
			want    int
		}{
			{entitlements.LimitSeats, 3},     // 公司全部非 inactive（1＋1＋1），不是部門的 1
			{entitlements.LimitCustomers, 1}, // 客戶的 department_id 為 NULL → 部門 scope 看不到
		} {
			got, err := counter.Count(ctx, coA, tc.feature)
			if err != nil {
				t.Fatalf("Count(%s): %v", tc.feature, err)
			}
			if got != tc.want {
				t.Errorf("部門 scope 下 %s 仍應回公司總數 %d（不可低報成部門用量），got %d",
					tc.feature, tc.want, got)
			}
		}
	})
}

// insertCounterUser 以 admin 連線建一位使用者；status 決定是否佔席位，
// departmentID=0 表示不屬於任何部門（NULL）。
func insertCounterUser(t *testing.T, db *sql.DB, companyID int, email, status string, departmentID int) {
	t.Helper()
	if _, err := db.Exec(
		`INSERT INTO users (email, name, role, status, password_hash, company_users, department_users)
		 VALUES ($1, '使用者', 'staff', $2, 'x', $3, NULLIF($4, 0))`, email, status, companyID, departmentID); err != nil {
		t.Fatalf("建使用者 %s: %v", email, err)
	}
}

// insertCounterDepartment 以 admin 連線建部門；softDeleted=true 時直接標記 deleted_at。
func insertCounterDepartment(t *testing.T, db *sql.DB, companyID int, name string, softDeleted bool) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO departments (name, company_departments, deleted_at)
		 VALUES ($1, $2, CASE WHEN $3 THEN now() ELSE NULL END) RETURNING id`,
		name, companyID, softDeleted).Scan(&id); err != nil {
		t.Fatalf("建部門 %s: %v", name, err)
	}
	return id
}

// insertCounterCustomer 以 admin 連線建客戶；softDeleted=true 時直接標記 deleted_at。
func insertCounterCustomer(t *testing.T, db *sql.DB, companyID int, code string, softDeleted bool) {
	t.Helper()
	if _, err := db.Exec(
		`INSERT INTO customers (company_id, customer_code, name, deleted_at)
		 VALUES ($1, $2, '客戶', CASE WHEN $3 THEN now() ELSE NULL END)`,
		companyID, code, softDeleted); err != nil {
		t.Fatalf("建客戶 %s: %v", code, err)
	}
}

// insertCounterProduct 以 admin 連線建商品；softDeleted=true 時直接標記 deleted_at。
func insertCounterProduct(t *testing.T, db *sql.DB, companyID, departmentID int, code string, softDeleted bool) {
	t.Helper()
	if _, err := db.Exec(
		`INSERT INTO products (company_id, department_id, code, name, deleted_at)
		 VALUES ($1, $2, $3, '商品', CASE WHEN $4 THEN now() ELSE NULL END)`,
		companyID, departmentID, code, softDeleted); err != nil {
		t.Fatalf("建商品 %s: %v", code, err)
	}
}

//go:build integration

// 示範資料 seeder(fake_seed_data.go)的守門探針。
//
// 本檔釘住四件事:
//
//	① production 一律不寫(示範資料絕不可進生產);
//	② 冪等 —— Taskfile 的 `task seed` 會反覆執行,重跑不得累積列、也不得換掉公司 id;
//	③ 不汙染平台自營公司(它是系統自己的租戶);
//	④ **業務角色 app_rw 能完成「該公司該來源的第一張訂單」** —— 這條與 seeder 無關,是生產
//	   契約:建單會 `ensureOrderCounter` 對 order_counters 取 nextval,而 00031 漏授權該序列
//	   (見 00043 檔頭)。它之所以放在本檔,是因為示範 seeder 是唯一會走完整建單流程的探針,
//	   而原本的取號測試走 sqlite(無角色權限概念),測不到。
//
// 連線角色的選擇:
//   - ①②③ 以 **owner(AdminDSN)** 跑:與 cmd/seed/main.go 完全相同。seeder 是維運工具,
//     生產的 owner 是表主人,對 append-only 表(print_logs/notifications)本來就有 DELETE;
//     改用 app_rw 反而是比實況更嚴的假限制,會把 seeder 誤判成壞的。
//     (FORCE RLS 讓 owner 也受 policy 約束,故仍須 SystemScopeTx —— 那一層由
//     seed_rls_integration_test.go 以 app_rw 釘住,此處不重複。)
//   - ④ 以 **app_rw** 跑:被測的是應用程式的真實連線角色。
package main

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationSeedFakeData 驗示範資料的 production 閘門、冪等與平台公司隔離。
func TestIntegrationSeedFakeData(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateForSeed(t, adminDSN)

	admin := openSeedAdmin(t, adminDSN)
	// 以 owner 連線執行(與 cmd/seed/main.go 相同;admin 在容器內是 superuser,但不影響
	// ①②③ 這三條 —— 它們驗的是資料結果,不是 RLS)。
	owner := dbtenant.NewClient(admin)
	t.Cleanup(func() { _ = owner.Close() })
	ctx := t.Context()

	// 平台自營公司由 SeedPlatform 建立;先建好才能驗「示範資料不寫進它」。
	platformID := insertCompany(t, admin, "平台營運", platformCompanyIdentifier)
	platformUsersBefore := seedCount(t, admin,
		`SELECT count(*) FROM users WHERE company_users = `+strconv.Itoa(platformID))

	t.Run("production 不寫任何示範資料", func(t *testing.T) {
		if err := runFake(ctx, owner, "production"); err != nil {
			t.Fatalf("production 應靜默略過, got %v", err)
		}
		if n := seedCount(t, admin,
			`SELECT count(*) FROM companies WHERE identifier LIKE 'FAKE-DEMO%'`); n != 0 {
			t.Fatalf("production 不得建立示範公司, got %d 家", n)
		}
	})

	t.Run("development 建立示範租戶且可重跑", func(t *testing.T) {
		if err := runFake(ctx, owner, "development"); err != nil {
			t.Fatalf("示範資料 seed 失敗: %v", err)
		}
		coID := seedCount(t, admin,
			`SELECT id FROM companies WHERE identifier = '`+fakeCompanyIdentifier+`'`)
		if coID == 0 {
			t.Fatal("應建立示範公司")
		}
		co := strconv.Itoa(coID)
		// 各表都應有列 —— 空狀態頁面就是沒資料才不好看,這是本 seeder 的存在理由。
		for _, tc := range []struct {
			name  string
			query string
			want  int
		}{
			{"部門", `SELECT count(*) FROM departments WHERE company_departments = ` + co, 2},
			{"使用者", `SELECT count(*) FROM users WHERE email = '` + fakeUserEmail + `'`, 1},
			{"客戶", `SELECT count(*) FROM customers WHERE company_id = ` + co, 5},
			{"商品", `SELECT count(*) FROM products WHERE company_id = ` + co, 8},
			{"訂單", `SELECT count(*) FROM sales_orders WHERE company_id = ` + co, 6},
			{"訂單明細", `SELECT count(*) FROM sales_order_items WHERE company_id = ` + co, 11},
			{"退貨申請", `SELECT count(*) FROM return_requests WHERE company_id = ` + co, 3},
			{"列印紀錄", `SELECT count(*) FROM print_logs WHERE company_id = ` + co, 6},
			{"通知", `SELECT count(*) FROM notifications WHERE company_id = ` + co, 5},
			{"取號器", `SELECT count(*) FROM order_counters WHERE company_id = ` + co, 1},
		} {
			if got := seedCount(t, admin, tc.query); got != tc.want {
				t.Errorf("%s 應有 %d 列, got %d", tc.name, tc.want, got)
			}
		}

		// 訂單編號必須是真實格式(來源碼 + 6 位補零),不是自創的示範格式。
		nos := seedStrings(t, admin,
			`SELECT order_no FROM sales_orders WHERE company_id = `+co+` ORDER BY order_no`)
		for _, no := range nos {
			if !strings.HasPrefix(no, "W") || len(no) != 7 {
				t.Errorf("訂單編號應為 W + 6 位數格式, got %q", no)
			}
		}
		// 取號器必須接在已用序號之後,否則 UI 新建訂單會撞既有單號。
		if next := seedCount(t, admin,
			`SELECT next_seq FROM order_counters WHERE company_id = `+co); next != len(nos)+1 {
			t.Errorf("取號器 next_seq 應為 %d(已用 %d 張後), got %d", len(nos)+1, len(nos), next)
		}
		// processing 必然有車次(派車才是轉 processing 的動作);否則看板兩邊都不顯示它。
		if n := seedCount(t, admin, `SELECT count(*) FROM sales_orders
			WHERE company_id = `+co+` AND status = 'processing' AND route_id IS NULL`); n != 0 {
			t.Errorf("processing 訂單不得沒有車次, got %d 張", n)
		}
		// 訂單事件由 FK cascade 隨訂單清除(events 對 app_rw 是 append-only,不直接刪);
		// 若 cascade 沒生效,重跑會累積孤兒事件。
		if n := seedCount(t, admin, `SELECT count(*) FROM sales_order_events e
			WHERE e.company_id = `+co+` AND NOT EXISTS (
				SELECT 1 FROM sales_orders o WHERE o.id = e.sales_order_id)`); n != 0 {
			t.Errorf("不應留下孤兒訂單事件, got %d 筆", n)
		}

		// 冪等:再跑一次不得改變任何列數與公司 id。
		if err := runFake(ctx, owner, "development"); err != nil {
			t.Fatalf("第二次 seed 必須成功(冪等): %v", err)
		}
		if got := seedCount(t, admin,
			`SELECT id FROM companies WHERE identifier = '`+fakeCompanyIdentifier+`'`); got != coID {
			t.Errorf("重跑不得換掉示範公司 id: %d → %d", coID, got)
		}
		if got := seedCount(t, admin,
			`SELECT count(*) FROM sales_orders WHERE company_id = `+co); got != 6 {
			t.Errorf("重跑不得累積訂單: want 6, got %d", got)
		}
		if got := seedCount(t, admin,
			`SELECT count(*) FROM print_logs WHERE company_id = `+co); got != 6 {
			t.Errorf("重跑不得累積列印紀錄: want 6, got %d", got)
		}
		if got := seedCount(t, admin,
			`SELECT count(*) FROM file_assets WHERE company_id = `+co); got != 6 {
			t.Errorf("重跑不得累積列印檔: want 6, got %d", got)
		}
	})

	t.Run("不汙染平台自營公司", func(t *testing.T) {
		if got := seedCount(t, admin,
			`SELECT count(*) FROM users WHERE company_users = `+strconv.Itoa(platformID)); got != platformUsersBefore {
			t.Errorf("示範資料不得寫進平台自營公司: users %d → %d", platformUsersBefore, got)
		}
	})
}

// TestIntegrationAppRoleCanCreateFirstOrder 釘住生產契約:業務角色 app_rw 能完成「該公司該
// 來源的第一張訂單」所必需的那一步 —— 對 order_counters 取 nextval。
//
// 為何要單獨一條:00031 的序列授權漏了 order_counters_id_seq(00043 補上),而
// `CreateSalesOrder` 對每張訂單都會先 `ensureOrderCounter`;counter 列不存在時是 INSERT,
// 需要序列權限。只有「新公司或新來源的第一次建單」會踩到,既有 counter 的公司走 UPDATE
// 不受影響 —— 這種形狀不會在一般測試流程自然浮現(取號測試走 sqlite,無角色權限概念)。
func TestIntegrationAppRoleCanCreateFirstOrder(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateForSeed(t, adminDSN)

	admin := openSeedAdmin(t, adminDSN)
	app := openSeedAppRole(t, adminDSN)
	client := dbtenant.NewClient(app)
	t.Cleanup(func() { _ = client.Close() })
	ctx := t.Context()

	// 不需要先建來源字典:本探針只驗「counter 的 INSERT 需要序列權限」,
	// 那是 ensureOrderCounter 的 SQL;來源字典的驗證屬 validateOrderSource,不在這條契約內。
	coID := insertCompany(t, admin, "建單探針公司", "ORDER-PROBE")

	// 以 app_rw(生產的業務連線角色)在系統範圍交易內建 counter —— 就是 ensureOrderCounter 的 SQL。
	err := dbtenant.SystemScopeTx(ctx, client, func(tx *ent.Tx) error {
		_, err := tx.Client().OrderCounter.Create().
			SetCompanyID(coID).SetSource("W").SetNextSeq(1).SetVersion(0).
			Save(ctx)
		return err
	})
	if err != nil {
		t.Fatalf("app_rw 必須能建立 order_counters(建單前置): %v", err)
	}
}

// runFake 在系統範圍交易內執行示範 seeder(與 cmd/seed/main.go 同一條路徑)。
func runFake(ctx context.Context, client *ent.Client, env string) error {
	return dbtenant.SystemScopeTx(ctx, client, func(tx *ent.Tx) error {
		return SeedFakeData(ctx, tx.Client(), env)
	})
}

// insertCompany 建一家最小公司並回傳 id(示範 seeder 之外的前置條件)。
func insertCompany(t *testing.T, db *sql.DB, name, identifier string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO companies (name, identifier, status) VALUES ($1,$2,'active') RETURNING id`,
		name, identifier).Scan(&id); err != nil {
		t.Fatalf("建公司 %s: %v", identifier, err)
	}
	return id
}

// seedStrings 以 admin 連線取單欄多列真值。
func seedStrings(t *testing.T, db *sql.DB, query string) []string {
	t.Helper()
	rows, err := db.Query(query)
	if err != nil {
		t.Fatalf("查詢 %q: %v", query, err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatalf("掃描 %q: %v", query, err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("走訪 %q: %v", query, err)
	}
	return out
}

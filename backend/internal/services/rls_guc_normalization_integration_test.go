//go:build integration

package services

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// 本檔守門「RLS policy 的 GUC 取值必須是 NULLIF 形式」這個不變式（2026-09-26 修，00051）。
//
// 缺陷形態：policy 寫成 `(current_setting('app.current_company_id', true))::bigint` 而不是
// `(NULLIF(current_setting('app.current_company_id', true), ”))::bigint`。
//
// 為何會爆：`SET LOCAL` 對自訂 GUC 會在 **session** 層留下空字串 placeholder，交易結束後該 GUC
// 還原成 `”` 而非 unset（見 internal/auth/rls.go 的 RLSStatements）。pgx stdlib 的 ResetSession
// 是 noop（pgx/v5 stdlib/sql.go：`ResetSession: func(...) error { return nil }`），不會清掉它。
// 於是同一條**池化連線**上，任何「不需要設該 GUC」的請求形狀 —— 例如公司管理員或無部門的
// dept_admin（不設 app.current_department_id），或 dbtenant.SystemScopeTx（只設
// app.current_data_scope）—— 讀到 `”::bigint` 就讓**整個交易**以 SQLSTATE 22P02 中止。
//
// 兩個容易漏掉的細節：
//  1. PG **不保證** `OR` 的求值順序，故 `data_scope='all'` 不會短路掉那個 cast（實測會爆）。
//  2. 症狀取決於**連線歷史**：同一個請求形狀在乾淨連線上成功、在服務過別的形狀的連線上失敗。
//     單一形狀的測試因此永遠看不到它 —— 本測試刻意在同一條連線上**依序**跑不同形狀。
//
// 為何既有測試擋不住：00025 已把更早的 18 個 policy 正規化，於是 auth 路徑（users/companies/
// customers/…）正常；壞的是 00031 之後新建表的 policy。`assertPoliciesNormalized`
// （rls_masters_integration_test.go）的白名單只含那 18 個，新表不在其中。
// 而容器 superuser 連線永遠繞過 RLS，更看不到任何一種。
//
// 為何必須兩段：catalog 段讓「下一個忘了寫 NULLIF 的 migration」在任何表上立刻變紅；
// 執行段證明那些 NULLIF 真的在 runtime 生效（catalog 有 NULLIF 但位置寫錯也會被執行段抓到）。
func TestIntegrationRLSGucPlaceholderIsNormalized(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	defer func() { _ = admin.Close() }()

	t.Run("catalog:所有 policy 的 GUC cast 都包 NULLIF", func(t *testing.T) {
		// 只要是 public schema 的一般表，其 policy 內出現裸 cast 就是缺陷 —— 不維護白名單，
		// 新增表自動納入（白名單正是這個缺陷當初漏掉 20 張表的原因）。
		rows, err := admin.Query(`
			SELECT c.relname, pol.polname,
			       coalesce(pg_get_expr(pol.polqual, pol.polrelid), '') AS qual,
			       coalesce(pg_get_expr(pol.polwithcheck, pol.polrelid), '') AS wc
			  FROM pg_policy pol
			  JOIN pg_class c ON c.oid = pol.polrelid
			 WHERE c.relnamespace = 'public'::regnamespace AND c.relkind = 'r'`)
		if err != nil {
			t.Fatalf("查 pg_policy: %v", err)
		}
		defer func() { _ = rows.Close() }()

		gucs := []string{"company_id", "department_id", "user_id", "customer_id"}
		checked := 0
		for rows.Next() {
			var tbl, pol, qual, wc string
			if err := rows.Scan(&tbl, &pol, &qual, &wc); err != nil {
				t.Fatalf("scan: %v", err)
			}
			for _, expr := range []string{qual, wc} {
				if expr == "" {
					continue
				}
				checked++
				for _, g := range gucs {
					// 裸形式：`(current_setting('app.current_<g>'::text, true))::bigint`
					// 正規化後：`(NULLIF(current_setting('app.current_<g>'::text, true), ''::text))::bigint`
					raw := "(current_setting('app.current_" + g + "'::text, true))::bigint"
					if strings.Contains(expr, raw) {
						t.Errorf("%s.%s 的 policy 以裸 cast 讀 app.current_%s —— SET LOCAL 後 "+
							"placeholder 會是空字串，同一條池化連線上的下一個請求形狀會以 "+
							"22P02 中止整個交易。改用 (NULLIF(current_setting('app.current_%s', true), ''))::bigint "+
							"（樣板：00025_rls_enable_masters.sql、00051_rls_guc_nullif_normalization.sql）\n"+
							"  運算式：%s", tbl, pol, g, g, expr)
					}
				}
			}
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("rows: %v", err)
		}
		// 沒有 policy 被讀到 = 斷言實際上沒跑（schema 沒建起來 / 查詢寫錯）。
		if checked == 0 {
			t.Fatal("沒有讀到任何 policy 運算式：本測試失去偵測力（migration 是否已套用？）")
		}
		t.Logf("已檢查 %d 個 policy 運算式", checked)
	})

	t.Run("runtime:同一條池化連線上換請求形狀不得 22P02", func(t *testing.T) {
		// MaxOpenConns(1) 強制所有交易落在同一條實體連線 —— 這正是缺陷的觸發條件。
		app, err := sql.Open("pgx", testsupport.AppRoleDSN(t, adminDSN))
		if err != nil {
			t.Fatalf("app_rw 連線: %v", err)
		}
		defer func() { _ = app.Close() }()
		app.SetMaxOpenConns(1)

		var who string
		var bypass bool
		if err := app.QueryRow(`SELECT current_user,
			rolsuper OR rolbypassrls FROM pg_roles WHERE rolname = current_user`).Scan(&who, &bypass); err != nil {
			t.Fatalf("確認身分: %v", err)
		}
		if bypass {
			t.Fatalf("必須以非 superuser／非 bypassrls 角色連線才有 RLS 語意,得到 %s(bypass=%v)", who, bypass)
		}

		coA := insertRLSCompany(t, admin, "GUC-A", "GUC-A")
		actor := insertRLSUser(t, admin, coA, "guc-a@example.com", "company_admin")
		cust := insertRLSCustomer(t, admin, coA, "GUC-C1", "GUC 客戶")
		var deptA int
		if err := admin.QueryRow(`INSERT INTO departments (name, company_departments) VALUES ('GUC 部門', $1) RETURNING id`,
			coA).Scan(&deptA); err != nil {
			t.Fatalf("建部門 fixture: %v", err)
		}

		// 每一張受測表都必須**有列**：policy 的 qual 只對被掃到的列求值，
		// 空表不會評估 qual，缺陷就偵測不到（實測：空表時本子測試在缺 00051 的情況下仍全綠）。
		fixtures := []struct {
			table  string
			insert string
		}{
			{"sales_orders", `INSERT INTO sales_orders (company_id, order_no, customer_id, source, status)
				VALUES ($1, 'GUC-000001', $2, 'manual', 'pending')`},
			{"notifications", `INSERT INTO notifications (company_id, department_id, user_id, channel, title, content)
				VALUES ($1, $2, $3, 'in_app', 'GUC 標題', 'GUC 內容')`},
			{"file_assets", `INSERT INTO file_assets (company_id, owner_type, owner_id, filename,
				original_filename, mime_type, size_bytes, storage_path, url)
				VALUES ($1, 'print_log', 0, 'guc.pdf', 'guc.pdf', 'application/pdf', 1, 'x/y/guc.pdf', '')`},
			{"order_counters", `INSERT INTO order_counters (company_id, source) VALUES ($1, 'manual')`},
		}
		for _, f := range fixtures {
			var err error
			switch f.table {
			case "notifications":
				_, err = admin.Exec(f.insert, coA, deptA, actor)
			case "sales_orders":
				_, err = admin.Exec(f.insert, coA, cust)
			default:
				_, err = admin.Exec(f.insert, coA)
			}
			if err != nil {
				t.Fatalf("建 %s fixture: %v", f.table, err)
			}
		}
		// sales_order_events 有 FK 到 sales_orders,借用上面那筆訂單。
		var orderID int
		if err := admin.QueryRow(`SELECT id FROM sales_orders WHERE company_id = $1 LIMIT 1`, coA).Scan(&orderID); err != nil {
			t.Fatalf("取訂單 id: %v", err)
		}
		if _, err := admin.Exec(`INSERT INTO sales_order_events (sales_order_id, company_id, event_type, actor_id)
			VALUES ($1, $2, 'create', $3)`, orderID, coA, actor); err != nil {
			t.Fatalf("建 sales_order_events fixture: %v", err)
		}

		// 涵蓋兩類受影響的表：三欄式（company＋department）與 only-company 式。
		// 都是 00031 之後新建、且曾用裸 cast 的表。
		tables := []string{"sales_orders", "notifications", "file_assets", "order_counters", "sales_order_events"}

		// shapes 依序在同一條連線上執行；每一個都只用得到部分 GUC，
		// 所以前一個 shape 留下的 placeholder 正是後一個的陷阱。
		//
		// wantRows 只在「公司相符即足夠」的形狀上斷言（見兩個 policy 家族）：
		//   - only-company 家族（order_counters/sales_order_events）：`all OR company 相符`
		//   - 三欄家族（sales_orders/notifications/file_assets）：`all OR (company 相符 AND
		//     (scope=company OR department 相符))`
		// department 範圍（shape 3）刻意**不**斷言列數：該 shape 本來就不設 department GUC，
		// 「看不到別部門的列」是正確行為，斷言列數會變成在測 policy 語意而非本缺陷。
		shapes := []struct {
			name     string
			stmts    []string
			wantRows bool // true 時斷言看得到 fixture 自己的列
		}{
			{
				// 完整三欄 scope：把四個 GUC placeholder 全部建立成 ''。
				name: "full scope（user+company+dept+scope=company）",
				stmts: []string{
					`SET LOCAL app.current_user_id = '` + itoa(actor) + `'`,
					`SET LOCAL app.current_company_id = '` + itoa(coA) + `'`,
					`SET LOCAL app.current_department_id = '` + itoa(deptA) + `'`,
					`SET LOCAL app.current_data_scope = 'company'`,
				},
				wantRows: true,
			},
			{
				// 公司管理員／無部門帳號：不設 department —— 缺陷的頭號觸發形狀。
				name: "company scope（user+company，無 department GUC）",
				stmts: []string{
					`SET LOCAL app.current_user_id = '` + itoa(actor) + `'`,
					`SET LOCAL app.current_company_id = '` + itoa(coA) + `'`,
					`SET LOCAL app.current_data_scope = 'company'`,
				},
				wantRows: true,
			},
			{
				// dbtenant.SystemScopeTx 的形狀：只有 data_scope（登入、身分解析、seed、consumer 都走這裡）。
				name: "system scope=all（僅 data_scope）",
				stmts: []string{
					`SET LOCAL app.current_data_scope = 'all'`,
				},
				wantRows: true,
			},
			{
				// 無部門的 dept_admin：department 範圍但不設 department GUC —— 第二個觸發形狀。
				name: "department scope（無 department GUC）",
				stmts: []string{
					`SET LOCAL app.current_user_id = '` + itoa(actor) + `'`,
					`SET LOCAL app.current_company_id = '` + itoa(coA) + `'`,
					`SET LOCAL app.current_data_scope = 'department'`,
				},
				wantRows: false, // 只驗證不報錯；列數取決於 policy 語意，不是本缺陷
			},
			{
				// 完全沒有 scope：必須 fail-closed 成 0 列，而不是 22P02。
				name:     "無 scope（應回 0 列）",
				stmts:    nil,
				wantRows: false,
			},
		}

		for _, sh := range shapes {
			for _, tbl := range tables {
				tx, err := app.BeginTx(t.Context(), nil)
				if err != nil {
					t.Fatalf("%s / %s: BEGIN: %v", sh.name, tbl, err)
				}
				for _, stmt := range sh.stmts {
					if _, err := tx.Exec(stmt); err != nil {
						t.Fatalf("%s: %s: %v", sh.name, stmt, err)
					}
				}
				var n int
				err = tx.QueryRow(`SELECT count(*) FROM ` + tbl).Scan(&n)
				_ = tx.Rollback()

				if err != nil {
					if strings.Contains(err.Error(), "22P02") || strings.Contains(err.Error(), `invalid input syntax for type bigint: ""`) {
						t.Errorf("%s → %s：查詢以 22P02 中止 —— policy 的 GUC cast 沒包 NULLIF，"+
							"placeholder 的空字串被當成 bigint。這會讓整個請求失敗，不是回 0 列。err=%v",
							sh.name, tbl, err)
						continue
					}
					t.Fatalf("%s → %s: 非預期錯誤: %v", sh.name, tbl, err)
				}
				if sh.name == "無 scope（應回 0 列）" && n != 0 {
					t.Errorf("%s → %s：應 fail-closed 回 0 列，得到 %d", sh.name, tbl, n)
				}
				if sh.wantRows && n == 0 {
					t.Errorf("%s → %s：應看得到 fixture 的列，得到 0 —— 正規化若把有效 scope 也吃掉，"+
						"就會以「查不到」而非報錯的形式壞掉", sh.name, tbl)
				}
			}
		}
	})
}

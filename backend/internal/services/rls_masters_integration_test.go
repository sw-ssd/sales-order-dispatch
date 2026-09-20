//go:build integration

package services

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	mastersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1/mastersv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// mastersRLSTables 是本域四張主檔表的 fixture 寫入語句(company_id 由夾具決定;created_at 等
// 走 DDL 預設,不寫入)。insert 一律以 $1=company_id、$2=code、$3=name 參數化。
var mastersRLSTables = []struct{ table, insert string }{
	{"warehouses", `INSERT INTO warehouses (company_id, code, name) VALUES ($1, $2, $3)`},
	{"routes", `INSERT INTO routes (company_id, code, name) VALUES ($1, $2, $3)`},
	{"processing_specs", `INSERT INTO processing_specs (company_id, code, name) VALUES ($1, $2, $3)`},
	{"product_categories", `INSERT INTO product_categories (company_id, code, name) VALUES ($1, $2, $3)`},
}

// mastersRLSPolicies 是 00025 必須以 NULLIF 形式重建的**全部 18 個** policy:
// 17 個由 00023 建立(Up 區塊)、core_metadicts_scope 由 00011 建立(00023 明文不動它,
// 其 WITH CHECK 刻意較嚴:不含 `department_id IS NULL`)。
var mastersRLSPolicies = []struct{ table, policy string }{
	{"companies", "core_companies_scope"},
	{"departments", "core_departments_scope"},
	{"users", "core_users_scope"},
	{"roles", "core_roles_read"},
	{"role_permissions", "core_role_permissions_read"},
	{"audit_logs", "core_audit_logs_scope"},
	{"metadicts", "core_metadicts_scope"},
	{"customers", "core_customers_scope"},
	{"customer_counters", "core_customer_counters_scope"},
	{"customer_addresses", "core_customer_addresses_scope"},
	{"customer_contacts", "core_customer_contacts_scope"},
	{"warehouses", "core_warehouses_scope"},
	{"routes", "core_routes_scope"},
	{"processing_specs", "core_processing_specs_scope"},
	{"product_categories", "core_product_categories_scope"},
	{"products", "core_products_scope"},
	{"product_units", "core_product_units_scope"},
	{"product_processing_specs", "core_product_processing_specs_scope"},
}

// TestIntegrationRLSMastersIsolation 以 app_rw 直連驗證部門級主檔四表隔離(00025 ENABLE + FORCE):
// 未設 scope → 0 列(fail-closed);設 A 公司 → 只看得到 A;以 A 的身分寫入 B 公司的列 → 被 WITH CHECK 擋。
//
// 為何用 app_rw 而非 admin:容器/測試的 admin 連線是 superuser,PG 的 superuser 永遠繞過 RLS
// (FORCE 亦然)→ 以 admin 連線根本測不到 fail-closed。app_rw 是 00022 建出的 NOBYPASSRLS 非 owner
// 業務角色,正是生產路徑的角色(T5 已為客戶域立下同一作法)。
func TestIntegrationRLSMastersIsolation(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	defer func() { _ = admin.Close() }()
	coA := insertRLSCompany(t, admin, "A", "MST-A")
	coB := insertRLSCompany(t, admin, "B", "MST-B")
	for _, tbl := range mastersRLSTables {
		insertMasterRow(t, admin, tbl.insert, coA, "A-1", "A 列")
		insertMasterRow(t, admin, tbl.insert, coB, "B-1", "B 列")
	}

	// 「未設 scope」與「設了 scope」刻意各用**獨立的連線池**:同一條池化連線若先服務過有 scope
	// 的交易,PG 會在 session 層留下 app.current_* 的空字串 placeholder(見
	// TestIntegrationRLSMastersPooledConnectionNoScope),兩個情境混用會讓「未設 scope 應 0 列」
	// 的斷言偶然變成 22P02。每個池只承擔一種模式,斷言便不依賴連線重用的運氣。
	t.Run("未設 scope → fail-closed", func(t *testing.T) {
		app := openAppRoleDB(t, adminDSN)
		for _, tbl := range mastersRLSTables {
			var n int
			if err := app.QueryRow(`SELECT count(*) FROM ` + tbl.table).Scan(&n); err != nil {
				t.Fatalf("查 %s: %v", tbl.table, err)
			}
			if n != 0 {
				t.Errorf("未設 scope 時 %s 必須 0 列,got %d", tbl.table, n)
			}
		}
	})

	t.Run("scope=company A → 只見 A 且跨租戶寫入被擋", func(t *testing.T) {
		app := openAppRoleDB(t, adminDSN)
		for _, tbl := range mastersRLSTables {
			tx, err := app.Begin()
			if err != nil {
				t.Fatalf("開交易: %v", err)
			}
			setAppScope(t, tx, coA)
			// 以 count + distinct + min 一次斷言「恰好只看到 A 的那一列」:只看 DISTINCT 再取
			// 第一列的話,未 ENABLE 時(兩家公司的列都看得到)可能剛好取到 A 而假通過。
			var n, distinct, minID int
			if err := tx.QueryRow(
				`SELECT count(*), count(DISTINCT company_id), COALESCE(min(company_id), 0) FROM `+tbl.table,
			).Scan(&n, &distinct, &minID); err != nil {
				t.Fatalf("查 %s: %v", tbl.table, err)
			}
			if n != 1 || distinct != 1 || minID != coA {
				t.Errorf("%s 應只看到公司 %d 的 1 列,got count=%d distinct=%d min=%d", tbl.table, coA, n, distinct, minID)
			}
			if _, err := tx.Exec(tbl.insert, coB, "X-1", "別家列"); err == nil {
				t.Errorf("以 A 的身分寫入 B 公司的 %s 必須被擋（WITH CHECK）", tbl.table)
			}
			_ = tx.Rollback()
		}
	})
}

// TestIntegrationMastersUnderAppRole 以 app_rw + 請求層租戶交易跑**真 handler**,逐條走過本域
// 四張主檔表已遷移的全部寫入路徑(Create/Update/Delete/Restore 各 4 支 + List),每條都斷言
// 回傳的列屬身分所屬公司;結尾的負向子測試證明「漏掛 scope/漏掛 dbtenant.Client」這條接線會紅。
//
// 為何既有主檔測試擋不住:master_service_test.go 走 sqlite(不支援 SET LOCAL,且無 RLS 語意),
// 其餘整合測試的連線是容器 superuser(PG 的 superuser 永遠繞過 RLS)→ 路徑漏掛 tenant client 也照樣全綠。
func TestIntegrationMastersUnderAppRole(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	defer func() { _ = admin.Close() }()
	coA := insertRLSCompany(t, admin, "A", "MSVC-A")
	coB := insertRLSCompany(t, admin, "B", "MSVC-B")
	// 稽核列有 user_id FK(00010),操作者必須是真實使用者。
	actorA := insertRLSUser(t, admin, coA, "masters-a@example.com", "company_admin")
	actorB := insertRLSUser(t, admin, coB, "masters-b@example.com", "company_admin")
	fixA := map[string]string{}
	fixB := map[string]string{}
	for _, tbl := range mastersRLSTables {
		fixA[tbl.table] = insertMasterRow(t, admin, tbl.insert, coA, "FIX-A", "A 夾具")
		fixB[tbl.table] = insertMasterRow(t, admin, tbl.insert, coB, "FIX-B", "B 夾具")
	}

	// 業務 client 必須經 dbtenant.NewClient 建立(RLS 裝飾器才生效);DB 連線是 app_rw。
	client := dbtenant.NewClient(openAppRoleDB(t, adminDSN))
	t.Cleanup(func() { _ = client.Close() })
	ctx := t.Context()
	companyA := itoa(coA)

	for _, e := range newMastersAppRoleServer(t, client, actorA, coA, true) {
		t.Run(e.table, func(t *testing.T) {
			// 跨租戶:B 的列在 A 的請求交易裡「不存在」→ 三個寫入動詞都必須 not_found
			// (而非「查到再拒絕」或洩漏存在性)。
			for _, verb := range []struct {
				name string
				call func() error
			}{
				{"update", func() error { _, err := e.update(ctx, fixB[e.table], "改他家的列"); return err }},
				{"delete", func() error { return e.remove(ctx, fixB[e.table]) }},
				{"restore", func() error { _, err := e.restore(ctx, fixB[e.table]); return err }},
			} {
				if err := verb.call(); connect.CodeOf(err) != connect.CodeNotFound {
					t.Errorf("以 A 的身分 %s B 公司的 %s 列應 not_found,got %v", verb.name, e.table, err)
				}
			}

			// Create:回傳的列必須帶身分所屬公司(company_admin → 公司層,department_id 為 NULL)。
			created, err := e.create(ctx, e.code+"-NEW", "新建主檔")
			if err != nil {
				t.Fatalf("Create %s(app_rw + 租戶範圍): %v", e.table, err)
			}
			if created.companyID != companyA {
				t.Fatalf("Create %s 回的列應屬公司 %s,得到 %q", e.table, companyA, created.companyID)
			}

			// List:只見本公司的列(新建 + A 夾具),不得出現 B 的列。
			assertMasterIDs(t, e.table, listMasterRows(ctx, t, e, false, companyA), created.id, fixA[e.table])

			// Update:欄位式更新後仍屬本公司。
			updated, err := e.update(ctx, created.id, "改名後")
			if err != nil {
				t.Fatalf("Update %s: %v", e.table, err)
			}
			if updated.name != "改名後" || updated.companyID != companyA {
				t.Fatalf("Update %s 應回傳改名後、仍屬公司 %s 的列,得到 %+v", e.table, companyA, updated)
			}

			// Delete(軟刪除):預設清單排除,include_deleted 仍看到且屬本公司。
			if err := e.remove(ctx, created.id); err != nil {
				t.Fatalf("Delete %s: %v", e.table, err)
			}
			assertMasterIDs(t, e.table, listMasterRows(ctx, t, e, false, companyA), fixA[e.table])
			deleted := listMasterRows(ctx, t, e, true, companyA)
			assertMasterIDs(t, e.table, deleted, created.id, fixA[e.table])
			for _, row := range deleted {
				if row.id == created.id && row.deletedAt == "" {
					t.Fatalf("include_deleted 的 %s 應帶 deleted_at,得到 %+v", e.table, row)
				}
			}

			// Restore:復原後回到預設清單,且仍屬本公司。
			restored, err := e.restore(ctx, created.id)
			if err != nil {
				t.Fatalf("Restore %s: %v", e.table, err)
			}
			if restored.companyID != companyA || restored.deletedAt != "" {
				t.Fatalf("Restore %s 應回傳已復原、屬公司 %s 的列,得到 %+v", e.table, companyA, restored)
			}
			assertMasterIDs(t, e.table, listMasterRows(ctx, t, e, false, companyA), created.id, fixA[e.table])
		})
	}

	// ── 負向對照:同一組 app_rw server,但**不注入 RLS scope**(模擬漏掛)────────────────
	// 刻意沿用**同一個連線池**(T5 當時因 policy 尚未做 NULLIF 正規化,只能另開乾淨的池;
	// 00025 修好後同一池即可):上面的請求已讓池中連線留下 app.current_* 的空字串 placeholder,
	// 少了 scope 之後 policy 必須把 '' 視為「未設」→ 讀回 0 列(而非 22P02)、寫被 WITH CHECK 擋下。
	t.Run("未注入 RLS scope → fail-closed", func(t *testing.T) {
		noScope := newMastersAppRoleServer(t, client, actorB, coB, false)
		for _, e := range noScope {
			rows, err := e.list(ctx, false)
			if err != nil {
				t.Errorf("無 scope 的 List %s 本身應成功(只是看不到列),got %v", e.table, err)
			} else if len(rows) != 0 {
				t.Errorf("無 scope 時不得看到任何 %s,得到 %d 筆", e.table, len(rows))
			}
			before := countRows(t, admin, `SELECT count(*) FROM `+e.table+` WHERE company_id = $1`, coB)
			if _, err := e.create(ctx, e.code+"-NOPE", "無 scope"); err == nil {
				t.Errorf("無 RLS scope 時 Create %s 必須失敗(寫入被 WITH CHECK 擋下)", e.table)
			}
			if after := countRows(t, admin, `SELECT count(*) FROM `+e.table+` WHERE company_id = $1`, coB); after != before {
				t.Errorf("被擋下的 Create %s 不得落地:B 公司應仍 %d 列,得到 %d", e.table, before, after)
			}
		}
	})
}

// TestIntegrationRLSMastersEnableMigrationDown 00025 的 Up/Down 必須對稱:Up 把 18 個 policy
// 全數改為 NULLIF 形式並對四表 ENABLE+FORCE;Down 把 policy 還原回 00023/00011 的原定義
// (逐條比對 pg_policies 的 qual/with_check 是否含 NULLIF)並關閉四表 RLS,且可重複套用。
//
// 只斷言旗標會漏掉「policy 沒還原」這個半套的 Down —— 那會讓回退後的環境與 00023 不一致,
// 而 NULLIF 的有無在旗標上看不出來。
func TestIntegrationRLSMastersEnableMigrationDown(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)

	admin := openRawDB(t, dsn)
	defer func() { _ = admin.Close() }()

	assertMastersRLSFlags(t, admin, true, true)
	assertPoliciesNormalized(t, admin, true)

	migrateBusinessDownTo(t, dsn, "24") // 僅回退 00025(00024 保留)
	assertMastersRLSFlags(t, admin, false, false)
	assertPoliciesNormalized(t, admin, false)

	migrateBusinessUp(t, dsn) // 回退後必須能重新套用(Up 冪等)
	assertMastersRLSFlags(t, admin, true, true)
	assertPoliciesNormalized(t, admin, true)
}

// TestIntegrationRLSMastersPooledConnectionNoScope 釘住 00025 唯一可觀測的行為改變:同一條
// 池化連線先服務過「有 scope」的請求之後,未設 scope 的查詢必須回 0 列,**不是** SQLSTATE 22P02。
//
// 背景:SET LOCAL 對自訂 GUC 會在 session 層留下空字串 placeholder,交易結束後仍是 ”(而非 unset)
// → 未設 scope 的查詢把 ” 餵進 policy 的 `::bigint` 就報 22P02。兩者都 fail-closed,但錯誤語意
// 與「查不到」混淆;NULLIF 把 ” 正規化成 NULL,語意回到「未設 → 0 列」。
func TestIntegrationRLSMastersPooledConnectionNoScope(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	defer func() { _ = admin.Close() }()
	coA := insertRLSCompany(t, admin, "A", "POOL-A")
	if _, err := admin.Exec(
		`INSERT INTO customers (company_id, customer_code, name) VALUES ($1, 'POOL-1', '客戶')`, coA); err != nil {
		t.Fatalf("建客戶: %v", err)
	}
	for _, tbl := range mastersRLSTables {
		insertMasterRow(t, admin, tbl.insert, coA, "POOL-1", "池化連線測試")
	}

	app := openAppRoleDB(t, adminDSN)
	// 把池收斂到單一連線:否則「後續未設 scope 的查詢」可能落在另一條乾淨連線上,
	// 這個斷言就會隨連線分配而飄。
	app.SetMaxOpenConns(1)
	app.SetMaxIdleConns(1)

	// 模擬一個完整的認證請求:開交易 → SET LOCAL scope → 讀(看得到 A 的列)→ commit。
	tx, err := app.Begin()
	if err != nil {
		t.Fatalf("開交易: %v", err)
	}
	setAppScope(t, tx, coA)
	var n int
	if err := tx.QueryRow(`SELECT count(*) FROM warehouses`).Scan(&n); err != nil {
		t.Fatalf("有 scope 的查詢: %v", err)
	}
	if n != 1 {
		t.Fatalf("有 scope 時應看到 A 的 1 筆倉別,got %d", n)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// 前提檢查:交易結束後該連線的 session 值確實是空字串 placeholder(而非 unset)。
	// 沒有這個前提就沒有本測試要釘的 22P02 情境 → 讓它紅,而不是靜靜失去偵測力。
	var sess sql.NullString
	if err := app.QueryRow(`SELECT current_setting('app.current_company_id', true)`).Scan(&sess); err != nil {
		t.Fatalf("讀 session 的 app.current_company_id: %v", err)
	}
	if !sess.Valid || sess.String != "" {
		t.Fatalf("本測試的前提不成立:交易後 session 的 app.current_company_id 應為空字串 placeholder,got %q(valid=%v)", sess.String, sess.Valid)
	}

	// 未設 scope 的查詢:policy 必須把 '' 正規化為「未設」而回 0 列(客戶域也在本 migration 的
	// 18 個 policy 之內,故一併斷言)。
	tables := []string{"customers"}
	for _, tbl := range mastersRLSTables {
		tables = append(tables, tbl.table)
	}
	for _, table := range tables {
		var got int
		if err := app.QueryRow(`SELECT count(*) FROM ` + table).Scan(&got); err != nil {
			t.Fatalf("同一條池化連線在未設 scope 的 %s 查詢必須回 0 列(00025 的 NULLIF 正規化),得到錯誤: %v", table, err)
		}
		if got != 0 {
			t.Fatalf("未設 scope 時 %s 必須 0 列,got %d", table, got)
		}
	}
}

// masterRow 是四張主檔表共有的可比欄位:斷言「回傳的列屬身分所屬公司」只需這幾個。
type masterRow struct {
	id        string
	companyID string
	name      string
	deletedAt string
}

// masterEntity 以閉包收斂四張表形狀相同的 5 支 RPC,讓「所有已遷移寫入路徑」以同一段流程走過,
// 不會有哪張表的路徑悄悄沒被覆蓋。
type masterEntity struct {
	table   string
	code    string // Create 用的 code(每張表的 (department_id, code) 唯一索引需要不同值)
	create  func(ctx context.Context, code, name string) (masterRow, error)
	list    func(ctx context.Context, includeDeleted bool) ([]masterRow, error)
	update  func(ctx context.Context, id, name string) (masterRow, error)
	remove  func(ctx context.Context, id string) error
	restore func(ctx context.Context, id string) (masterRow, error)
}

func warehouseRow(w *mastersv1.Warehouse) masterRow {
	return masterRow{w.GetId(), w.GetCompanyId(), w.GetName(), w.GetDeletedAt()}
}

func routeRow(r *mastersv1.Route) masterRow {
	return masterRow{r.GetId(), r.GetCompanyId(), r.GetName(), r.GetDeletedAt()}
}

func processingSpecRow(s *mastersv1.ProcessingSpec) masterRow {
	return masterRow{s.GetId(), s.GetCompanyId(), s.GetName(), s.GetDeletedAt()}
}

func productCategoryRow(c *mastersv1.ProductCategory) masterRow {
	return masterRow{c.GetId(), c.GetCompanyId(), c.GetName(), c.GetDeletedAt()}
}

// newMastersAppRoleServer 以業務 client(app_rw + dbtenant.NewClient)掛上主檔四支 service:
// RegisterXService 內含 dbtenant.HandlerOption → 每個 RPC 都在請求交易內執行。
// withScope=false 模擬「漏掛 RLS scope」(負向對照);生產的 scope 由 authzMiddleware 依身分導出。
func newMastersAppRoleServer(t *testing.T, client *ent.Client, actor, companyID int, withScope bool) []masterEntity {
	t.Helper()
	mux := http.NewServeMux()
	RegisterWarehouseService(mux, client)
	RegisterRouteService(mux, client)
	RegisterProcessingSpecService(mux, client)
	RegisterProductCategoryService(mux, client)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), authz.Identity{
			UserID: itoa(actor), CompanyID: itoa(companyID), Role: "company_admin", Roles: []string{"company_admin"},
		})
		ctx = authz.WithDB(ctx, client)
		if withScope {
			ctx = auth.WithRLS(ctx, auth.RLSScope{
				UserID: itoa(actor), CompanyID: itoa(companyID), DataScope: auth.DataScopeCompany, CompanyActive: true,
			})
		}
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	wh := mastersv1connect.NewWarehouseServiceClient(http.DefaultClient, ts.URL)
	rt := mastersv1connect.NewRouteServiceClient(http.DefaultClient, ts.URL)
	ps := mastersv1connect.NewProcessingSpecServiceClient(http.DefaultClient, ts.URL)
	pc := mastersv1connect.NewProductCategoryServiceClient(http.DefaultClient, ts.URL)

	return []masterEntity{
		{
			table: "warehouses", code: "WH",
			create: func(ctx context.Context, code, name string) (masterRow, error) {
				res, err := wh.CreateWarehouse(ctx, connect.NewRequest(&mastersv1.CreateWarehouseRequest{Code: code, Name: name, IsActive: true}))
				if err != nil {
					return masterRow{}, err
				}
				return warehouseRow(res.Msg.GetWarehouse()), nil
			},
			list: func(ctx context.Context, includeDeleted bool) ([]masterRow, error) {
				res, err := wh.ListWarehouses(ctx, connect.NewRequest(&mastersv1.ListWarehousesRequest{Page: 1, PageSize: 50, IncludeDeleted: includeDeleted}))
				if err != nil {
					return nil, err
				}
				rows := make([]masterRow, 0, len(res.Msg.GetWarehouses()))
				for _, e := range res.Msg.GetWarehouses() {
					rows = append(rows, warehouseRow(e))
				}
				return rows, nil
			},
			update: func(ctx context.Context, id, name string) (masterRow, error) {
				res, err := wh.UpdateWarehouse(ctx, connect.NewRequest(&mastersv1.UpdateWarehouseRequest{Id: id, Name: &name}))
				if err != nil {
					return masterRow{}, err
				}
				return warehouseRow(res.Msg.GetWarehouse()), nil
			},
			remove: func(ctx context.Context, id string) error {
				_, err := wh.DeleteWarehouse(ctx, connect.NewRequest(&mastersv1.DeleteWarehouseRequest{Id: id}))
				return err
			},
			restore: func(ctx context.Context, id string) (masterRow, error) {
				res, err := wh.RestoreWarehouse(ctx, connect.NewRequest(&mastersv1.RestoreWarehouseRequest{Id: id}))
				if err != nil {
					return masterRow{}, err
				}
				return warehouseRow(res.Msg.GetWarehouse()), nil
			},
		},
		{
			table: "routes", code: "RT",
			create: func(ctx context.Context, code, name string) (masterRow, error) {
				res, err := rt.CreateRoute(ctx, connect.NewRequest(&mastersv1.CreateRouteRequest{Code: code, Name: name, IsActive: true}))
				if err != nil {
					return masterRow{}, err
				}
				return routeRow(res.Msg.GetRoute()), nil
			},
			list: func(ctx context.Context, includeDeleted bool) ([]masterRow, error) {
				res, err := rt.ListRoutes(ctx, connect.NewRequest(&mastersv1.ListRoutesRequest{Page: 1, PageSize: 50, IncludeDeleted: includeDeleted}))
				if err != nil {
					return nil, err
				}
				rows := make([]masterRow, 0, len(res.Msg.GetRoutes()))
				for _, e := range res.Msg.GetRoutes() {
					rows = append(rows, routeRow(e))
				}
				return rows, nil
			},
			update: func(ctx context.Context, id, name string) (masterRow, error) {
				res, err := rt.UpdateRoute(ctx, connect.NewRequest(&mastersv1.UpdateRouteRequest{Id: id, Name: &name}))
				if err != nil {
					return masterRow{}, err
				}
				return routeRow(res.Msg.GetRoute()), nil
			},
			remove: func(ctx context.Context, id string) error {
				_, err := rt.DeleteRoute(ctx, connect.NewRequest(&mastersv1.DeleteRouteRequest{Id: id}))
				return err
			},
			restore: func(ctx context.Context, id string) (masterRow, error) {
				res, err := rt.RestoreRoute(ctx, connect.NewRequest(&mastersv1.RestoreRouteRequest{Id: id}))
				if err != nil {
					return masterRow{}, err
				}
				return routeRow(res.Msg.GetRoute()), nil
			},
		},
		{
			table: "processing_specs", code: "PS",
			create: func(ctx context.Context, code, name string) (masterRow, error) {
				// 加工規格的旗標至少要有一項 true(validateSpecFlags),否則會在請求驗證就被拒。
				res, err := ps.CreateProcessingSpec(ctx, connect.NewRequest(&mastersv1.CreateProcessingSpecRequest{Code: code, Name: name, AppliesToPicking: true, IsActive: true}))
				if err != nil {
					return masterRow{}, err
				}
				return processingSpecRow(res.Msg.GetProcessingSpec()), nil
			},
			list: func(ctx context.Context, includeDeleted bool) ([]masterRow, error) {
				res, err := ps.ListProcessingSpecs(ctx, connect.NewRequest(&mastersv1.ListProcessingSpecsRequest{Page: 1, PageSize: 50, IncludeDeleted: includeDeleted}))
				if err != nil {
					return nil, err
				}
				rows := make([]masterRow, 0, len(res.Msg.GetProcessingSpecs()))
				for _, e := range res.Msg.GetProcessingSpecs() {
					rows = append(rows, processingSpecRow(e))
				}
				return rows, nil
			},
			update: func(ctx context.Context, id, name string) (masterRow, error) {
				res, err := ps.UpdateProcessingSpec(ctx, connect.NewRequest(&mastersv1.UpdateProcessingSpecRequest{Id: id, Name: &name}))
				if err != nil {
					return masterRow{}, err
				}
				return processingSpecRow(res.Msg.GetProcessingSpec()), nil
			},
			remove: func(ctx context.Context, id string) error {
				_, err := ps.DeleteProcessingSpec(ctx, connect.NewRequest(&mastersv1.DeleteProcessingSpecRequest{Id: id}))
				return err
			},
			restore: func(ctx context.Context, id string) (masterRow, error) {
				res, err := ps.RestoreProcessingSpec(ctx, connect.NewRequest(&mastersv1.RestoreProcessingSpecRequest{Id: id}))
				if err != nil {
					return masterRow{}, err
				}
				return processingSpecRow(res.Msg.GetProcessingSpec()), nil
			},
		},
		{
			table: "product_categories", code: "PC",
			create: func(ctx context.Context, code, name string) (masterRow, error) {
				res, err := pc.CreateProductCategory(ctx, connect.NewRequest(&mastersv1.CreateProductCategoryRequest{Code: code, Name: name, IsActive: true}))
				if err != nil {
					return masterRow{}, err
				}
				return productCategoryRow(res.Msg.GetProductCategory()), nil
			},
			list: func(ctx context.Context, includeDeleted bool) ([]masterRow, error) {
				res, err := pc.ListProductCategories(ctx, connect.NewRequest(&mastersv1.ListProductCategoriesRequest{Page: 1, PageSize: 50, IncludeDeleted: includeDeleted}))
				if err != nil {
					return nil, err
				}
				rows := make([]masterRow, 0, len(res.Msg.GetProductCategories()))
				for _, e := range res.Msg.GetProductCategories() {
					rows = append(rows, productCategoryRow(e))
				}
				return rows, nil
			},
			update: func(ctx context.Context, id, name string) (masterRow, error) {
				res, err := pc.UpdateProductCategory(ctx, connect.NewRequest(&mastersv1.UpdateProductCategoryRequest{Id: id, Name: &name}))
				if err != nil {
					return masterRow{}, err
				}
				return productCategoryRow(res.Msg.GetProductCategory()), nil
			},
			remove: func(ctx context.Context, id string) error {
				_, err := pc.DeleteProductCategory(ctx, connect.NewRequest(&mastersv1.DeleteProductCategoryRequest{Id: id}))
				return err
			},
			restore: func(ctx context.Context, id string) (masterRow, error) {
				res, err := pc.RestoreProductCategory(ctx, connect.NewRequest(&mastersv1.RestoreProductCategoryRequest{Id: id}))
				if err != nil {
					return masterRow{}, err
				}
				return productCategoryRow(res.Msg.GetProductCategory()), nil
			},
		},
	}
}

// insertMasterRow 以 admin 連線(夾具寫入)建一列並回傳 id 字串。
func insertMasterRow(t *testing.T, db *sql.DB, insert string, companyID int, code, name string) string {
	t.Helper()
	var id int
	if err := db.QueryRow(insert+` RETURNING id`, companyID, code, name).Scan(&id); err != nil {
		t.Fatalf("建列(code=%s,公司 %d): %v", code, companyID, err)
	}
	return itoa(id)
}

// listMasterRows 取清單並逐列斷言「回傳的列屬身分所屬公司」:RLS 生效時混入他公司的列,
// 就是服務路徑漏掛 dbtenant.Client(查詢落在無 scope 的池化連線上)。
func listMasterRows(ctx context.Context, t *testing.T, e masterEntity, includeDeleted bool, companyID string) []masterRow {
	t.Helper()
	rows, err := e.list(ctx, includeDeleted)
	if err != nil {
		t.Fatalf("List %s: %v", e.table, err)
	}
	for _, row := range rows {
		if row.companyID != companyID {
			t.Fatalf("List %s 回傳了公司 %q 的列(身分屬公司 %s):%+v", e.table, row.companyID, companyID, row)
		}
	}
	return rows
}

// assertMasterIDs 斷言清單恰好是期望的 id 集合(少了本公司的列或多出他人的列都要紅)。
func assertMasterIDs(t *testing.T, table string, rows []masterRow, want ...string) {
	t.Helper()
	got := make(map[string]bool, len(rows))
	for _, row := range rows {
		got[row.id] = true
	}
	if len(rows) != len(want) {
		t.Fatalf("%s 清單應 %d 筆 %v,得到 %d 筆 %v", table, len(want), want, len(rows), got)
	}
	for _, id := range want {
		if !got[id] {
			t.Fatalf("%s 清單缺少 id=%s(實際 %v)", table, id, got)
		}
	}
}

// assertMastersRLSFlags 斷言主檔四表的 ENABLE/FORCE 旗標。
func assertMastersRLSFlags(t *testing.T, db *sql.DB, enabled, forced bool) {
	t.Helper()
	for _, tbl := range mastersRLSTables {
		var e, f bool
		if err := db.QueryRow(
			`SELECT relrowsecurity, relforcerowsecurity FROM pg_class WHERE oid = $1::regclass`, tbl.table).Scan(&e, &f); err != nil {
			t.Fatalf("查 %s 的 RLS 旗標: %v", tbl.table, err)
		}
		if e != enabled || f != forced {
			t.Fatalf("%s 的 RLS 旗標應為 ENABLE=%v FORCE=%v,got ENABLE=%v FORCE=%v", tbl.table, enabled, forced, e, f)
		}
	}
}

// assertPoliciesNormalized 逐條斷言 18 個 policy 的 USING/WITH CHECK 是否為 NULLIF 形式
// (want=true 對應 00025 的 Up、false 對應其 Down 還原的 00023/00011 原定義)。
// 找不到某個 policy 也視為失敗 —— 重建時漏掉一張表就沒有這條斷言遮蔽的空間。
func assertPoliciesNormalized(t *testing.T, db *sql.DB, want bool) {
	t.Helper()
	for _, p := range mastersRLSPolicies {
		var qual, withCheck sql.NullString
		if err := db.QueryRow(
			`SELECT qual, with_check FROM pg_policies WHERE tablename = $1 AND policyname = $2`,
			p.table, p.policy).Scan(&qual, &withCheck); err != nil {
			t.Fatalf("查 policy %s(%s): %v", p.policy, p.table, err)
		}
		if got := strings.Contains(qual.String, "NULLIF"); got != want {
			t.Errorf("%s.%s 的 USING 應為 NULLIF 形式=%v,got=%v(%s)", p.table, p.policy, want, got, qual.String)
		}
		if got := strings.Contains(withCheck.String, "NULLIF"); got != want {
			t.Errorf("%s.%s 的 WITH CHECK 應為 NULLIF 形式=%v,got=%v(%s)", p.table, p.policy, want, got, withCheck.String)
		}
	}
}

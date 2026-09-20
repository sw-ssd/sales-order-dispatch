//go:build integration

package services

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	productsv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1/productsv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// productsRLSTables 是商品域三張表(00026 必須 ENABLE + FORCE 者)。
// fixtures 的形狀不一致(products 有 company_id,兩張子表靠 product_id),故隔離斷言逐表手寫,
// 這份清單只供旗標斷言使用。
var productsRLSTables = []string{"products", "product_units", "product_processing_specs"}

// TestIntegrationRLSProductsIsolation 以 app_rw 直連驗證商品域三表隔離(00026 ENABLE + FORCE):
// 未設 scope → 0 列(fail-closed);設 A 公司 → 只看得到 A;以 A 的身分寫入 B 的列 → 被 WITH CHECK 擋。
//
// 子表傳遞性(product_units / product_processing_specs 沒有 company_id,00023 的 policy 以
// EXISTS(SELECT 1 FROM products p WHERE p.id = <子表>.product_id) 表達可見性,子查詢本身也受
// products 的 policy 約束)是本測試的重點:父商品不可見時子列必須一起消失,且跨租戶寫入子列
// (掛到他人的商品上)必須被擋 —— 少了 ENABLE 這兩件事都會默默通過。
//
// 為何用 app_rw 而非 admin:容器/測試的 admin 連線是 superuser,PG 的 superuser 永遠繞過 RLS
// (FORCE 亦然)→ 以 admin 連線根本測不到 fail-closed。app_rw 是 00022 建出的 NOBYPASSRLS 非 owner
// 業務角色,正是生產路徑的角色(T5/T6 已為客戶域與主檔域立下同一作法)。
func TestIntegrationRLSProductsIsolation(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	defer func() { _ = admin.Close() }()
	coA := insertRLSCompany(t, admin, "A", "PRD-A")
	coB := insertRLSCompany(t, admin, "B", "PRD-B")
	prodA := insertRLSProduct(t, admin, coA, "PA")
	prodB := insertRLSProduct(t, admin, coB, "PB")
	specA := insertRLSProcessingSpec(t, admin, coA, "SP-A")
	specB := insertRLSProcessingSpec(t, admin, coB, "SP-B")
	unitA := insertRLSProductUnit(t, admin, prodA, "KG")
	insertRLSProductUnit(t, admin, prodB, "KG")
	linkA := insertRLSProductSpecLink(t, admin, prodA, specA)
	insertRLSProductSpecLink(t, admin, prodB, specB)

	// 「未設 scope」與「設了 scope」刻意各用**獨立的連線池**:同一條池化連線若先服務過有 scope
	// 的交易,PG 會在 session 層留下 app.current_* 的空字串 placeholder(見 T6 的池化連線測試),
	// 兩個情境混用會讓「未設 scope 應 0 列」的斷言偶然變成 22P02。
	t.Run("未設 scope → fail-closed", func(t *testing.T) {
		app := openAppRoleDB(t, adminDSN)
		for _, table := range productsRLSTables {
			var n int
			if err := app.QueryRow(`SELECT count(*) FROM ` + table).Scan(&n); err != nil {
				t.Fatalf("查 %s: %v", table, err)
			}
			if n != 0 {
				t.Errorf("未設 scope 時 %s 必須 0 列,got %d", table, n)
			}
		}
	})

	t.Run("scope=company A → 只見 A（含子表傳遞性）", func(t *testing.T) {
		app := openAppRoleDB(t, adminDSN)
		tx, err := app.Begin()
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		setAppScope(t, tx, coA)

		// products:以 count + distinct + min 一次斷言「恰好只看到 A 的那一列」。只看 DISTINCT 再取
		// 第一列的話,未 ENABLE 時(兩家公司的列都看得到)可能剛好取到 A 而假通過。
		var n, distinct, minID int
		if err := tx.QueryRow(
			`SELECT count(*), count(DISTINCT company_id), COALESCE(min(company_id), 0) FROM products`,
		).Scan(&n, &distinct, &minID); err != nil {
			t.Fatalf("查 products: %v", err)
		}
		if n != 1 || distinct != 1 || minID != coA {
			t.Errorf("products 應只看到公司 %d 的 1 列,got count=%d distinct=%d min=%d", coA, n, distinct, minID)
		}

		// 子表:無 company_id,可見性由父表 products 的存在性傳遞 → 只看得到 A 商品的子列。
		for _, child := range []struct {
			table string
			want  int
		}{
			{"product_units", unitA},
			{"product_processing_specs", linkA},
		} {
			var cnt, distinctParents, minChild int
			if err := tx.QueryRow(
				`SELECT count(*), count(DISTINCT product_id), COALESCE(min(id), 0) FROM ` + child.table,
			).Scan(&cnt, &distinctParents, &minChild); err != nil {
				t.Fatalf("查 %s: %v", child.table, err)
			}
			if cnt != 1 || distinctParents != 1 || minChild != child.want {
				t.Errorf("%s 應只看到 A 商品的那一列(id=%d),got count=%d distinct_product=%d min_id=%d",
					child.table, child.want, cnt, distinctParents, minChild)
			}
			// 顯式斷言「父不可見 → 子不可見」:以 B 的商品 id 直接查子表,必須 0 列。
			var cross int
			if err := tx.QueryRow(`SELECT count(*) FROM `+child.table+` WHERE product_id = $1`, prodB).Scan(&cross); err != nil {
				t.Fatalf("查 %s 的 B 商品子列: %v", child.table, err)
			}
			if cross != 0 {
				t.Errorf("父商品不可見時 %s 的子列也不得可見,got %d 列", child.table, cross)
			}
		}

	})

	// 跨租戶寫入:以 A 的身分寫 B 公司的商品,或把子列掛到 B 的商品上 → 皆被 WITH CHECK 擋。
	// 逐條開**獨立交易**(一條失敗會 abort 整個交易,後續語句只剩 25P02,掩蓋真正的守門)且逐條
	// 斷言 SQLSTATE=42501:只斷「有錯」的話,FK 或唯一索引的錯誤也會讓它假通過。子表斷言刻意用
	// B 的父商品 + **A 的** spec,理由是 FK 檢查不受 RLS 約束(實測 (prodA, specB) 在 A 的範圍內
	// 可寫入),故錯誤只可能來自「父不可見」這條 policy。
	t.Run("scope=company A → 跨租戶寫入以 RLS 違反被擋", func(t *testing.T) {
		app := openAppRoleDB(t, adminDSN)
		for _, tc := range []struct {
			name string
			sql  string
			args []any
		}{
			{"products(company_id=B)", `INSERT INTO products (company_id, code, name) VALUES ($1, 'X-1', '別家商品')`, []any{coB}},
			{"product_units(父為 B 的商品)", `INSERT INTO product_units (product_id, unit_code, conversion_rate, is_base) VALUES ($1, 'BX', '1', false)`, []any{prodB}},
			{"product_processing_specs(父為 B 的商品)", `INSERT INTO product_processing_specs (product_id, processing_spec_id) VALUES ($1, $2)`, []any{prodB, specA}},
		} {
			tx, err := app.Begin()
			if err != nil {
				t.Fatalf("開交易: %v", err)
			}
			setAppScope(t, tx, coA)
			_, err = tx.Exec(tc.sql, tc.args...)
			_ = tx.Rollback()
			if !isRLSViolation(err) {
				t.Errorf("以 A 的身分寫入 %s 應以 RLS 違反(SQLSTATE 42501)被擋,got %v", tc.name, err)
			}
		}
	})
}

// isRLSViolation 判斷錯誤是否為 PostgreSQL 的 RLS 違反(42501 insufficient_privilege,
// 「new row violates row-level security policy」)。斷言錯誤碼而非只斷「有錯」,是為了讓
// 「被 FK／唯一索引擋下」不會冒充「被 policy 擋下」。
func isRLSViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42501"
}

// TestIntegrationProductServiceUnderAppRole 以 app_rw + 請求層租戶交易跑**真 handler**,逐條走過
// 商品域已遷移的路徑(Create/Get/List/Update/Delete/Restore 各一支 + 兩張子表的讀寫),每條都
// 斷言回傳的列屬身分所屬公司;結尾的負向子測試證明「漏掛 scope」這條接線會紅。
//
// Create 刻意**引用既有的**分類/倉別/加工規格(00025 已 ENABLE 的表):這正是 product_service.go
// 若仍以裸 s.db 讀那三張表會 fail-closed 的路徑(引用驗證會回 invalid_argument),故本條同時是
// 「收斂完成」的證據。
//
// 為何既有商品測試擋不住:product_service_test.go 走 sqlite(不支援 SET LOCAL,且無 RLS 語意),
// 其餘整合測試的連線是容器 superuser(PG 的 superuser 永遠繞過 RLS)→ 路徑漏掛 tenant client 也照樣全綠。
func TestIntegrationProductServiceUnderAppRole(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	defer func() { _ = admin.Close() }()
	coA := insertRLSCompany(t, admin, "A", "PRDS-A")
	coB := insertRLSCompany(t, admin, "B", "PRDS-B")
	// 稽核列有 user_id FK(00010),操作者必須是真實使用者。
	actorA := insertRLSUser(t, admin, coA, "products-a@example.com", "company_admin")
	actorB := insertRLSUser(t, admin, coB, "products-b@example.com", "company_admin")
	// unit_code KG／BOX 已由 00011 種入系統 unit 字典(validateUnitCode 的標的),不必自建。
	refsA := seedProductAppRefs(t, admin, coA)
	refsB := seedProductAppRefs(t, admin, coB)
	fixA := insertRLSProduct(t, admin, coA, "FIX-A")
	fixB := insertRLSProduct(t, admin, coB, "FIX-B")

	// 業務 client 必須經 dbtenant.NewClient 建立(RLS 裝飾器才生效);DB 連線是 app_rw。
	client := dbtenant.NewClient(openAppRoleDB(t, adminDSN))
	t.Cleanup(func() { _ = client.Close() })
	ctx := t.Context()
	companyA := itoa(coA)
	pc := newProductAppRoleServer(t, client, actorA, coA, true)

	// 跨租戶:B 的商品在 A 的請求交易裡「不存在」→ 四個動詞都必須 not_found
	// (而非「查到再拒絕」或洩漏存在性)。
	for _, verb := range []struct {
		name string
		call func() error
	}{
		{"get", func() error {
			_, err := pc.GetProduct(ctx, connect.NewRequest(&productsv1.GetProductRequest{Id: itoa(fixB)}))
			return err
		}},
		{"update", func() error {
			_, err := pc.UpdateProduct(ctx, connect.NewRequest(&productsv1.UpdateProductRequest{Id: itoa(fixB), Name: strPtr("改他家的商品")}))
			return err
		}},
		{"delete", func() error {
			_, err := pc.DeleteProduct(ctx, connect.NewRequest(&productsv1.DeleteProductRequest{Id: itoa(fixB)}))
			return err
		}},
		{"restore", func() error {
			_, err := pc.RestoreProduct(ctx, connect.NewRequest(&productsv1.RestoreProductRequest{Id: itoa(fixB)}))
			return err
		}},
	} {
		if err := verb.call(); connect.CodeOf(err) != connect.CodeNotFound {
			t.Errorf("以 A 的身分 %s B 公司的商品應 not_found,got %v", verb.name, err)
		}
	}

	// Create:引用 A 的既有主檔(T6 窗口)+ 兩組單位 + 一筆規格關聯;回的列必須屬 A。
	req := &productsv1.CreateProductRequest{
		Code: "P-NEW", Name: "新商品", IsActive: true,
		CategoryId:           refsA.category,
		InventoryWarehouseId: refsA.warehouse,
		PickingWarehouseId:   refsA.warehouse,
		Units: []*productsv1.ProductUnit{
			{UnitCode: "KG", ConversionRate: "1", IsBase: true, SortOrder: 1},
			{UnitCode: "BOX", ConversionRate: "5", IsBase: false, SortOrder: 2},
		},
		ProcessingSpecs: []*productsv1.ProductProcessingSpec{{ProcessingSpecId: refsA.spec}},
	}
	cres, err := pc.CreateProduct(ctx, connect.NewRequest(req))
	if err != nil {
		t.Fatalf("CreateProduct(app_rw + 租戶範圍): %v", err)
	}
	created := cres.Msg.GetProduct()
	if created.GetCompanyId() != companyA {
		t.Fatalf("CreateProduct 回的列應屬公司 %s,得到 %q", companyA, created.GetCompanyId())
	}
	if len(created.GetUnits()) != 2 || len(created.GetProcessingSpecs()) != 1 {
		t.Fatalf("CreateProduct 應回 2 單位 + 1 規格關聯,得到 %d／%d", len(created.GetUnits()), len(created.GetProcessingSpecs()))
	}
	pid := created.GetId()
	// 子表真實落地(admin 真值):子列寫入也在租戶範圍內成功,不是「靜默寫到別處」。
	if got := countRows(t, admin, `SELECT count(*) FROM product_units WHERE product_id = $1`, mustItoa(t, pid)); got != 2 {
		t.Fatalf("product_units 應為新商品落地 2 列,得到 %d", got)
	}
	if got := countRows(t, admin, `SELECT count(*) FROM product_processing_specs WHERE product_id = $1`, mustItoa(t, pid)); got != 1 {
		t.Fatalf("product_processing_specs 應為新商品落地 1 列,得到 %d", got)
	}

	// List:只見本公司的列(新建 + A 夾具),不得出現 B 的列。
	assertProductList(t, companyA, listProducts(ctx, t, pc, false, companyA), pid, itoa(fixA))

	// Get:巢狀單位／規格關聯讀得到 —— fetchNested 走的是兩張子表(父可見則子可見)。
	gres, err := pc.GetProduct(ctx, connect.NewRequest(&productsv1.GetProductRequest{Id: pid}))
	if err != nil {
		t.Fatalf("GetProduct: %v", err)
	}
	if g := gres.Msg.GetProduct(); g.GetCompanyId() != companyA || len(g.GetUnits()) != 2 || len(g.GetProcessingSpecs()) != 1 {
		t.Fatalf("GetProduct 應回屬公司 %s 且含 2 單位／1 規格的商品,得到 %+v", companyA, g)
	}

	// Update:改名 + units 整組替換為單一基本單位。
	ures, err := pc.UpdateProduct(ctx, connect.NewRequest(&productsv1.UpdateProductRequest{
		Id: pid, Name: strPtr("改名後"),
		Units: []*productsv1.ProductUnit{{UnitCode: "KG", ConversionRate: "1", IsBase: true, SortOrder: 1}},
	}))
	if err != nil {
		t.Fatalf("UpdateProduct: %v", err)
	}
	if u := ures.Msg.GetProduct(); u.GetName() != "改名後" || u.GetCompanyId() != companyA || len(u.GetUnits()) != 1 {
		t.Fatalf("UpdateProduct 應回改名後、仍屬公司 %s、1 單位的列,得到 %+v", companyA, u)
	}
	// 整組替換在子表上生效(admin 真值):舊單位列已被刪除。
	if got := countRows(t, admin, `SELECT count(*) FROM product_units WHERE product_id = $1`, mustItoa(t, pid)); got != 1 {
		t.Fatalf("units 整組替換後 product_units 應剩 1 列,得到 %d", got)
	}

	// Delete(軟刪除):預設清單排除,include_deleted 仍看到且屬本公司(帶 deleted_at)。
	if _, err := pc.DeleteProduct(ctx, connect.NewRequest(&productsv1.DeleteProductRequest{Id: pid})); err != nil {
		t.Fatalf("DeleteProduct: %v", err)
	}
	assertProductList(t, companyA, listProducts(ctx, t, pc, false, companyA), itoa(fixA))
	deleted := listProducts(ctx, t, pc, true, companyA)
	assertProductList(t, companyA, deleted, pid, itoa(fixA))
	for _, row := range deleted {
		if row.GetId() == pid && row.GetDeletedAt() == "" {
			t.Fatalf("include_deleted 的商品應帶 deleted_at,得到 %+v", row)
		}
	}

	// Restore:復原後回到預設清單,且仍屬本公司。
	rres, err := pc.RestoreProduct(ctx, connect.NewRequest(&productsv1.RestoreProductRequest{Id: pid}))
	if err != nil {
		t.Fatalf("RestoreProduct: %v", err)
	}
	if r := rres.Msg.GetProduct(); r.GetCompanyId() != companyA || r.GetDeletedAt() != "" {
		t.Fatalf("RestoreProduct 應回已復原、屬公司 %s 的列,得到 %+v", companyA, r)
	}
	assertProductList(t, companyA, listProducts(ctx, t, pc, false, companyA), pid, itoa(fixA))

	// ── 負向對照:同一組 app_rw server,但**不注入 RLS scope**(模擬漏掛)────────────────
	// 刻意沿用**同一個連線池**:上面的請求已讓池中連線留下 app.current_* 的空字串 placeholder,
	// 少了 scope 之後 policy 必須把 '' 視為「未設」→ 讀回 0 列(而非 22P02)、寫被 WITH CHECK 擋下。
	t.Run("未注入 RLS scope → fail-closed", func(t *testing.T) {
		noScope := newProductAppRoleServer(t, client, actorB, coB, false)
		rows, err := noScope.ListProducts(ctx, connect.NewRequest(&productsv1.ListProductsRequest{Page: 1, PageSize: 50}))
		if err != nil {
			t.Errorf("無 scope 的 ListProducts 本身應成功(只是看不到列),got %v", err)
		} else if n := len(rows.Msg.GetProducts()); n != 0 {
			t.Errorf("無 scope 時不得看到任何商品,得到 %d 筆", n)
		}
		before := countRows(t, admin, `SELECT count(*) FROM products WHERE company_id = $1`, coB)
		if _, err := noScope.CreateProduct(ctx, connect.NewRequest(&productsv1.CreateProductRequest{
			Code: "P-NOPE", Name: "無 scope", IsActive: true,
			CategoryId:           refsB.category,
			InventoryWarehouseId: refsB.warehouse,
			PickingWarehouseId:   refsB.warehouse,
			Units:                []*productsv1.ProductUnit{{UnitCode: "KG", ConversionRate: "1", IsBase: true, SortOrder: 1}},
		})); err == nil {
			t.Error("無 RLS scope 時 CreateProduct 必須失敗(寫入被 WITH CHECK 擋下)")
		}
		if after := countRows(t, admin, `SELECT count(*) FROM products WHERE company_id = $1`, coB); after != before {
			t.Errorf("被擋下的 CreateProduct 不得落地:B 公司應仍 %d 列商品,得到 %d", before, after)
		}
	})
}

// TestIntegrationRLSProductsEnableMigrationDown 00026 的 Up/Down 必須對稱:Up 對三表
// ENABLE + FORCE,Down 先 NO FORCE 再 DISABLE(少寫 NO FORCE 會留下 FORCE 旗標、少寫 DISABLE
// 則 RLS 仍生效 —— 兩者都讓回退後的環境與 00025 的狀態不一致),且可重複套用。
//
// 另外斷言三個 policy 在 Down 之後**仍存在**:00026 只負責 ENABLE,policy 由 00023/00025 擁有,
// Down 若順手 DROP POLICY 就會讓回退後的環境失去定義(旗標上看不出來)。
func TestIntegrationRLSProductsEnableMigrationDown(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)

	admin := openRawDB(t, dsn)
	defer func() { _ = admin.Close() }()

	assertProductsRLSFlags(t, admin, true, true)

	migrateBusinessDownTo(t, dsn, "25") // 僅回退 00026(00025 保留)
	assertProductsRLSFlags(t, admin, false, false)
	assertProductPoliciesPresent(t, admin)

	migrateBusinessUp(t, dsn) // 回退後必須能重新套用(Up 冪等)
	assertProductsRLSFlags(t, admin, true, true)
}

// assertProductsRLSFlags 斷言商品域三表的 ENABLE/FORCE 旗標。
func assertProductsRLSFlags(t *testing.T, db *sql.DB, enabled, forced bool) {
	t.Helper()
	for _, table := range productsRLSTables {
		var e, f bool
		if err := db.QueryRow(
			`SELECT relrowsecurity, relforcerowsecurity FROM pg_class WHERE oid = $1::regclass`, table).Scan(&e, &f); err != nil {
			t.Fatalf("查 %s 的 RLS 旗標: %v", table, err)
		}
		if e != enabled || f != forced {
			t.Fatalf("%s 的 RLS 旗標應為 ENABLE=%v FORCE=%v,got ENABLE=%v FORCE=%v", table, enabled, forced, e, f)
		}
	}
}

// assertProductPoliciesPresent 斷言商品域三個 policy 仍存在於 pg_policies(00026 不得動 policy)。
func assertProductPoliciesPresent(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, p := range []struct{ table, policy string }{
		{"products", "core_products_scope"},
		{"product_units", "core_product_units_scope"},
		{"product_processing_specs", "core_product_processing_specs_scope"},
	} {
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM pg_policies WHERE tablename = $1 AND policyname = $2`,
			p.table, p.policy).Scan(&n); err != nil {
			t.Fatalf("查 policy %s(%s): %v", p.policy, p.table, err)
		}
		if n != 1 {
			t.Fatalf("%s.%s 應存在(00026 不得動 policy),got %d 筆", p.table, p.policy, n)
		}
	}
}

// productAppRefs 是一間公司供商品引用的主檔 id(分類／倉別／加工規格)。
type productAppRefs struct{ category, warehouse, spec string }

// seedProductAppRefs 以 admin 連線(夾具)在公司層建分類／倉別／加工規格各一,回傳其 id 字串。
func seedProductAppRefs(t *testing.T, db *sql.DB, companyID int) productAppRefs {
	t.Helper()
	var cat, wh, sp int
	if err := db.QueryRow(
		`INSERT INTO product_categories (company_id, code, name) VALUES ($1, 'CAT', '分類') RETURNING id`,
		companyID).Scan(&cat); err != nil {
		t.Fatalf("建分類(公司 %d): %v", companyID, err)
	}
	if err := db.QueryRow(
		`INSERT INTO warehouses (company_id, code, name) VALUES ($1, 'WH', '倉別') RETURNING id`,
		companyID).Scan(&wh); err != nil {
		t.Fatalf("建倉別(公司 %d): %v", companyID, err)
	}
	// 加工規格需至少一個旗標為 true(services 的 validateSpecFlags)。
	if err := db.QueryRow(
		`INSERT INTO processing_specs (company_id, code, name, applies_to_processing) VALUES ($1, 'SP', '加工規格', true) RETURNING id`,
		companyID).Scan(&sp); err != nil {
		t.Fatalf("建加工規格(公司 %d): %v", companyID, err)
	}
	return productAppRefs{category: itoa(cat), warehouse: itoa(wh), spec: itoa(sp)}
}

// insertRLSProduct 以 admin 連線建一列商品並回傳 id。
func insertRLSProduct(t *testing.T, db *sql.DB, companyID int, code string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO products (company_id, code, name) VALUES ($1, $2, '商品') RETURNING id`,
		companyID, code).Scan(&id); err != nil {
		t.Fatalf("建商品 %s(公司 %d): %v", code, companyID, err)
	}
	return id
}

// insertRLSProductUnit 以 admin 連線為某商品建一列基本單位並回傳 id。
func insertRLSProductUnit(t *testing.T, db *sql.DB, productID int, unitCode string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO product_units (product_id, unit_code, conversion_rate, is_base) VALUES ($1, $2, '1', true) RETURNING id`,
		productID, unitCode).Scan(&id); err != nil {
		t.Fatalf("建商品單位(product=%d): %v", productID, err)
	}
	return id
}

// insertRLSProcessingSpec 以 admin 連線建一列加工規格並回傳 id(子表 fixture 的 FK 標的)。
func insertRLSProcessingSpec(t *testing.T, db *sql.DB, companyID int, code string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO processing_specs (company_id, code, name) VALUES ($1, $2, '規格') RETURNING id`,
		companyID, code).Scan(&id); err != nil {
		t.Fatalf("建加工規格 %s(公司 %d): %v", code, companyID, err)
	}
	return id
}

// insertRLSProductSpecLink 以 admin 連線為某商品建一列加工規格關聯並回傳 id。
func insertRLSProductSpecLink(t *testing.T, db *sql.DB, productID, specID int) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO product_processing_specs (product_id, processing_spec_id) VALUES ($1, $2) RETURNING id`,
		productID, specID).Scan(&id); err != nil {
		t.Fatalf("建商品規格關聯(product=%d,spec=%d): %v", productID, specID, err)
	}
	return id
}

// newProductAppRoleServer 以業務 client(app_rw + dbtenant.NewClient)掛上商品 service:
// RegisterProductService 內含 dbtenant.HandlerOption → 每個 RPC 都在請求交易內執行。
// withScope=false 模擬「漏掛 RLS scope」(負向對照);生產的 scope 由 authzMiddleware 依身分導出。
func newProductAppRoleServer(t *testing.T, client *ent.Client, actor, companyID int, withScope bool) productsv1connect.ProductServiceClient {
	t.Helper()
	mux := http.NewServeMux()
	RegisterProductService(mux, client)
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
	return productsv1connect.NewProductServiceClient(http.DefaultClient, ts.URL)
}

// listProducts 取商品清單並逐列斷言「回傳的列屬身分所屬公司」:RLS 生效時混入他公司的列,
// 就是服務路徑漏掛 dbtenant.Client(查詢落在無 scope 的池化連線上)。
func listProducts(ctx context.Context, t *testing.T, pc productsv1connect.ProductServiceClient, includeDeleted bool, companyID string) []*productsv1.Product {
	t.Helper()
	res, err := pc.ListProducts(ctx, connect.NewRequest(&productsv1.ListProductsRequest{
		Page: 1, PageSize: 50, IncludeDeleted: includeDeleted,
	}))
	if err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
	for _, p := range res.Msg.GetProducts() {
		if p.GetCompanyId() != companyID {
			t.Fatalf("ListProducts 回傳了公司 %q 的商品(身分屬公司 %s):%+v", p.GetCompanyId(), companyID, p)
		}
	}
	return res.Msg.GetProducts()
}

// assertProductList 斷言清單恰好是期望的 id 集合(少了本公司的列或多出他人的列都要紅)。
func assertProductList(t *testing.T, companyID string, rows []*productsv1.Product, want ...string) {
	t.Helper()
	got := make(map[string]bool, len(rows))
	for _, p := range rows {
		got[p.GetId()] = true
	}
	if len(rows) != len(want) {
		t.Fatalf("商品清單應 %d 筆(公司 %s),得到 %d 筆 %v", len(want), companyID, len(rows), got)
	}
	for _, id := range want {
		if !got[id] {
			t.Fatalf("商品清單缺少 id=%s(實際 %v)", id, got)
		}
	}
}

// mustItoa 把 RPC 回傳的 id 字串轉為 admin 真值查詢用的整數(格式錯誤即紅)。
func mustItoa(t *testing.T, s string) int {
	t.Helper()
	n, err := parseID(s)
	if err != nil {
		t.Fatalf("id 格式錯誤(%q): %v", s, err)
	}
	return n
}

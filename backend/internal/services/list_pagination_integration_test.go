//go:build integration

// 三張清單(公司/部門/角色)的「逐頁掃描 == 全量」真 PostgreSQL 整合測試(F1 迴歸)。
//
// F1(T15 波驗收,2026-09-19):排序鍵非唯一時 `ORDER BY <field>` 沒有次序鍵,PostgreSQL
// 對同值群(ties)的順序在不同 LIMIT/OFFSET 下不保證一致(小 LIMIT 走 bounded top-N
// heap sort、大 OFFSET 走完整 quicksort),於是同一列可能在兩頁出現、另一列完全不出現 ——
// 使用者逐頁翻完只看到 71/83 筆相異資料。修法:每個 listSource 在排序鍵後追加
// `ent.Asc(<entity>.FieldID)` 作為次序鍵。
//
// 為何 sqlite(enttest)不足以守住:sqlite 的 sorter 對相同查詢給出穩定的 tie 順序,
// LIMIT/OFFSET 只是同一結果的切片,故本缺陷在 sqlite 上無法重現(已實測:96 筆 tie 資料、
// 每頁 20 筆逐頁掃描,全部白名單欄位 × 升/降冪皆為全綠)。此測試因此必須跑真 PostgreSQL:
// 以 internal/testsupport 起拋棄式容器,再用真實的 service handler 與 ent client 建資料與
// 查詢(不另寫複製的 SQL)。
package services

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"testing"

	"connectrpc.com/connect"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx database/sql driver(ent 以 OpenDB 包裝)

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/department"
	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

const (
	// listScanPageSize 與服務預設每頁筆數(defaultPageSize)一致:足以讓 PostgreSQL 對
	// 前幾頁選 bounded top-N heap sort、對大 OFFSET 選完整排序(重現條件)。
	listScanPageSize = 20
	// listScanRows 每張表的 fixture 筆數(需 > listScanPageSize × 2 才有跨頁同值群)。
	listScanRows = 96
)

// listScanner 以指定排序取回「某一頁」的 id 序列(逐頁掃描與全量掃描共用同一個呼叫點)。
type listScanner func(t *testing.T, sortField string, desc bool, page, pageSize int32) []string

// TestIntegrationListPageScanMatchesFullScan F1 迴歸:對三張清單的每個白名單排序欄位
// (含升/降冪與預設排序)以多筆同值(NULl/重複 tax_id、同名、同 status,遠多於一頁)的資料,
// 逐頁掃描的結果必須與全量一致:
//
//	筆數 == 全量筆數、id 集合 == 全量 id 集合、且無任何重複 id。
//
// 拿掉三個 listSource 的 id 次序鍵 → 本測試必須紅(消去實驗見 f1-fix-report.md)。
func TestIntegrationListPageScanMatchesFullScan(t *testing.T) {
	// fixture 筆數與 id 集合都以「全新空庫」為前提(DB 真值即全量),覆寫模式下 skip。
	testsupport.RequiresContainer(t)
	ctx := t.Context()
	db := openPGEntClient(t, testsupport.Postgres(t))

	co := seedListScanCompanies(t, ctx, db)
	seedListScanDepartments(t, ctx, db, co)
	seedListScanRoles(t, ctx, db)

	cc, dc, rc := newListScanServer(t, db)

	companyScan := func(t *testing.T, sortField string, desc bool, page, pageSize int32) []string {
		t.Helper()
		res, err := cc.ListCompanies(ctx, connect.NewRequest(&v1.ListCompaniesRequest{
			Page: page, PageSize: pageSize, Sort: sortField, Desc: desc,
		}))
		if err != nil {
			t.Fatalf("ListCompanies(page=%d sort=%q desc=%v): %v", page, sortField, desc, err)
		}
		ids := make([]string, 0, len(res.Msg.GetCompanies()))
		for _, c := range res.Msg.GetCompanies() {
			ids = append(ids, c.GetId())
		}
		return ids
	}
	departmentScan := func(t *testing.T, sortField string, desc bool, page, pageSize int32) []string {
		t.Helper()
		res, err := dc.ListDepartments(ctx, connect.NewRequest(&v1.ListDepartmentsRequest{
			Page: page, PageSize: pageSize, Sort: sortField, Desc: desc,
		}))
		if err != nil {
			t.Fatalf("ListDepartments(page=%d sort=%q desc=%v): %v", page, sortField, desc, err)
		}
		ids := make([]string, 0, len(res.Msg.GetDepartments()))
		for _, d := range res.Msg.GetDepartments() {
			ids = append(ids, d.GetId())
		}
		return ids
	}
	roleScan := func(t *testing.T, sortField string, desc bool, page, pageSize int32) []string {
		t.Helper()
		res, err := rc.ListRoles(ctx, connect.NewRequest(&v1.ListRolesRequest{
			Page: page, PageSize: pageSize, Sort: sortField, Desc: desc,
		}))
		if err != nil {
			t.Fatalf("ListRoles(page=%d sort=%q desc=%v): %v", page, sortField, desc, err)
		}
		ids := make([]string, 0, len(res.Msg.GetRoles()))
		for _, r := range res.Msg.GetRoles() {
			ids = append(ids, r.GetId())
		}
		return ids
	}

	// 白名單欄位(空字串 = 服務預設排序)皆須涵蓋,含升/降冪。
	all := map[string][]string{
		"公司": allIDs(t, func() ([]int, error) { return db.Company.Query().Select(company.FieldID).Ints(ctx) }),
		"部門": allIDs(t, func() ([]int, error) { return db.Department.Query().Select(department.FieldID).Ints(ctx) }),
		"角色": allIDs(t, func() ([]int, error) { return db.Role.Query().Select(role.FieldID).Ints(ctx) }),
	}
	for _, tc := range []struct {
		entity string
		field  string
		scan   listScanner
	}{
		{"公司", "", companyScan},
		{"公司", "name", companyScan},
		{"公司", "identifier", companyScan},
		{"公司", "tax_id", companyScan},
		{"公司", "status", companyScan},
		{"公司", "id", companyScan},
		{"部門", "", departmentScan},
		{"部門", "name", departmentScan},
		{"部門", "id", departmentScan},
		{"角色", "", roleScan},
		{"角色", "code", roleScan},
		{"角色", "name", roleScan},
		{"角色", "id", roleScan},
	} {
		for _, desc := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/sort=%q/desc=%v", tc.entity, tc.field, desc), func(t *testing.T) {
				paged := scanPages(t, tc.scan, tc.field, desc)
				assertScanMatchesFull(t, tc.entity, tc.field, desc, paged, all[tc.entity])
			})
		}
	}
}

// scanPages 模擬前端逐頁翻頁:同一個排序鍵反覆查詢 page=1..N,直到某頁不足 listScanPageSize
// 筆為止,回傳依序串接的 id(含重複,交由斷言判別)。
func scanPages(t *testing.T, scan listScanner, sortField string, desc bool) []string {
	t.Helper()
	ids := make([]string, 0, listScanRows)
	for page := int32(1); page <= listScanRows/listScanPageSize+1; page++ {
		got := scan(t, sortField, desc, page, listScanPageSize)
		ids = append(ids, got...)
		if len(got) < listScanPageSize {
			return ids
		}
	}
	t.Fatalf("sort=%q desc=%v:逐頁掃描超過 %d 頁仍未結束(F1 的重複/遺漏會讓掃描不終止)",
		sortField, desc, listScanRows/listScanPageSize+1)
	return nil
}

// assertScanMatchesFull 斷言逐頁掃描與全量一致:無重複 id、筆數相同、id 集合相同。
// 三者任一不符即為 F1(排序鍵非唯一 → LIMIT/OFFSET 的 tie 順序不穩定)。
func assertScanMatchesFull(t *testing.T, entity, sortField string, desc bool, paged, full []string) {
	t.Helper()
	label := fmt.Sprintf("%s sort=%q desc=%v", entity, sortField, desc)

	seen := make(map[string]int, len(paged))
	for _, id := range paged {
		seen[id]++
	}
	var dups []string
	for id, n := range seen {
		if n > 1 {
			dups = append(dups, fmt.Sprintf("%s×%d", id, n))
		}
	}
	if len(dups) > 0 {
		slices.Sort(dups)
		t.Errorf("%s:逐頁掃描出現重複 id %v(共 %d 筆、相異 %d 筆)", label, dups, len(paged), len(seen))
	}
	if len(paged) != len(full) {
		t.Errorf("%s:逐頁掃描 %d 筆 != 全量 %d 筆", label, len(paged), len(full))
	}
	var missing, extra []string
	for _, id := range full {
		if seen[id] == 0 {
			missing = append(missing, id)
		}
	}
	inFull := make(map[string]bool, len(full))
	for _, id := range full {
		inFull[id] = true
	}
	for id := range seen {
		if !inFull[id] {
			extra = append(extra, id)
		}
	}
	if len(missing) > 0 || len(extra) > 0 {
		slices.Sort(missing)
		slices.Sort(extra)
		t.Errorf("%s:id 集合不符 —— 遺漏 %v、多出 %v", label, missing, extra)
	}
}

// allIDs 取 DB 真值(該表 id 全量),作為「全量」基準。
func allIDs(t *testing.T, query func() ([]int, error)) []string {
	t.Helper()
	ids, err := query()
	if err != nil {
		t.Fatalf("查詢全量 id: %v", err)
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, strconv.Itoa(id))
	}
	return out
}

// openPGEntClient 以 pgx 連線建立 ent client 並套用 ent schema(測試只需這些表的欄位形狀,
// 不重跑 goose migration:排序缺陷與 migration 無關,且 ent schema 與 000xx 遷移同源)。
func openPGEntClient(t *testing.T, dsn string) *ent.Client {
	t.Helper()
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open(pgx): %v", err)
	}
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, sqlDB)))
	t.Cleanup(func() { _ = client.Close() })
	if err := client.Schema.Create(t.Context()); err != nil {
		t.Fatalf("建立 ent schema: %v", err)
	}
	return client
}

// newListScanServer 以 super 身分(全權,涵蓋三個 requireScope/requireRole 門檻)把公司 +
// 部門 + 角色 handler 掛在單一 mux 上(與 sqlite 版 newTestServer 同構,只換 DB)。
func newListScanServer(t *testing.T, db *ent.Client) (salesorderv1connect.CompanyServiceClient, salesorderv1connect.DepartmentServiceClient, salesorderv1connect.RoleServiceClient) {
	t.Helper()
	super := authz.Identity{UserID: "1", CompanyID: "c1", Role: "super", Roles: []string{"super"}}
	mux := http.NewServeMux()
	RegisterCompanyServices(mux, db)
	RegisterRoleServices(mux, db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), super)
		ctx = authz.WithCASLEnabled(ctx, true)
		ctx = authz.WithDB(ctx, db)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	return salesorderv1connect.NewCompanyServiceClient(http.DefaultClient, ts.URL),
		salesorderv1connect.NewDepartmentServiceClient(http.DefaultClient, ts.URL),
		salesorderv1connect.NewRoleServiceClient(http.DefaultClient, ts.URL)
}

// seedListScanCompanies 建立 listScanRows 家公司並回傳第一家(供部門掛載)。
// 讓每個排序鍵都有「跨頁同值群」:name 每 24 筆同名、tax_id 每 3 筆一組(含 NULL)、
// status 依 parity 二值(皆大於 listScanPageSize,確保每頁邊界都切在同值群內);
// identifier 為 schema 唯一鍵故無同值群。
func seedListScanCompanies(t *testing.T, ctx context.Context, db *ent.Client) *ent.Company {
	t.Helper()
	taxIDs := []string{"", "11111111", "22222222"} // 空字串 = tax_id 為 NULL
	builders := make([]*ent.CompanyCreate, 0, listScanRows)
	for i := range listScanRows {
		status := company.StatusActive
		if i%2 == 1 {
			status = company.StatusInactive
		}
		b := db.Company.Create().
			SetName(fmt.Sprintf("公司%02d", i/24)).
			SetIdentifier(fmt.Sprintf("P-%03d", i)).
			SetStatus(status)
		if tax := taxIDs[i%len(taxIDs)]; tax != "" {
			b = b.SetTaxID(tax)
		}
		builders = append(builders, b)
	}
	created, err := db.Company.CreateBulk(builders...).Save(ctx)
	if err != nil {
		t.Fatalf("建立公司 fixture: %v", err)
	}
	return created[0]
}

// seedListScanDepartments 在單一公司下建立 listScanRows 個部門:name 每 24 筆同名
// (部門無其他排序欄位,同值群須大於 listScanPageSize 才會被頁邊界切開)。
func seedListScanDepartments(t *testing.T, ctx context.Context, db *ent.Client, co *ent.Company) {
	t.Helper()
	builders := make([]*ent.DepartmentCreate, 0, listScanRows)
	for i := range listScanRows {
		builders = append(builders, db.Department.Create().
			SetName(fmt.Sprintf("部門%02d", i/24)).
			SetCompanyID(co.ID))
	}
	if _, err := db.Department.CreateBulk(builders...).Save(ctx); err != nil {
		t.Fatalf("建立部門 fixture: %v", err)
	}
}

// seedListScanRoles 建立 listScanRows 個角色:name 每 24 筆同名、code 唯一
// (角色預設排序為 id 升冪)。
func seedListScanRoles(t *testing.T, ctx context.Context, db *ent.Client) {
	t.Helper()
	builders := make([]*ent.RoleCreate, 0, listScanRows)
	for i := range listScanRows {
		builders = append(builders, db.Role.Create().
			SetCode(fmt.Sprintf("R-%03d", i)).
			SetName(fmt.Sprintf("角色%02d", i/24)))
	}
	if _, err := db.Role.CreateBulk(builders...).Save(ctx); err != nil {
		t.Fatalf("建立角色 fixture: %v", err)
	}
}

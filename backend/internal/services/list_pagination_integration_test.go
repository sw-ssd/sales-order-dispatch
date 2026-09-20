//go:build integration

// 四張清單(公司/部門/角色/客戶)的「逐頁掃描 == 全量」真 PostgreSQL 整合測試(F1 / P1-A 迴歸)。
//
// F1(T15 波驗收,2026-09-19):排序鍵非唯一時 `ORDER BY <field>` 沒有次序鍵,PostgreSQL
// 對同值群(ties)的順序在不同 LIMIT/OFFSET 下不保證一致(小 LIMIT 走 bounded top-N
// heap sort、大 OFFSET 走完整 quicksort),於是同一列可能在兩頁出現、另一列完全不出現 ——
// 使用者逐頁翻完只看到 71/83 筆相異資料。修法:每個 listSource 在排序鍵後追加
// `ent.Asc(<entity>.FieldID)` 作為次序鍵。
//
// P1-A(Phase 3 波複審,2026-09-20):同一缺陷類別在客戶清單(`customerListSource.Page`,
// `customer_service.go`)仍存在 —— 其排序白名單 `name`/`customer_code`/`created_at` 中
// `name` 與 `created_at` 皆非唯一。修法同 F1。
//
// D2(2026-09-20):客戶端點補上 `desc`(proto 的 ListCustomersRequest 先前無此欄位),
// 故客戶清單與公司/部門/角色一致地掃升冪與降冪。
//
// 降冪另有一層守門:兩個方向都掃的端點,同一排序欄位(非空)**升冪與降冪的頁序必須不同**
// (見 assertDescReversesAsc)——「desc 被忽略」這種退化實作在逐頁/全量集合比對下看不出來
// (兩個方向都會等於全量),只有比對兩個方向的序列才看得到。
//
// F2(Phase 3 波複審,2026-09-20):同型缺陷另外收斂在六個清單端點(六者的 proto 都沒有
// 排序參數,排序鍵固定):加工規格/商品分類/車次以 `sort_order` 排序(欄位預設 0,同值群
// 常遠大於一頁)、字典以 `(sort_order, code)` 排序、商品/倉別以 `code` 排序。後三者的排序鍵
// 在可見集合內不唯一 —— 字典的 code 唯一性只有 `(type, 部門)`,同一 code 在系統預設的不同
// type、以及「系統預設 + 當前部門」的合併視角下都會重複;商品/倉別的 code 唯一性是
// `(department_id, code)`,而 super/company_admin 的 deptScope 回 did=nil → 可見集合跨部門,
// 同 code 一覽無遺。六處同法修:排序鍵後追加 `ent.Asc(<entity>.FieldID)`。
//
// M2(A/B 複審,2026-09-20):稽核清單(ListAuditLogs)同型 —— 只以 `created_at` 降冪排序,
// 而 created_at 是**交易時間**:同一業務交易內寫入的多筆稽核(2.6.2/D18 同事務寫稽核)時間戳
// 完全相同,同值群真實存在且常大於一頁。故排序鍵後追加 `ent.Desc(auditlog.FieldID)`(方向與
// 主鍵一致),並納入本測試(fixture 造出同 created_at 的多筆稽核,且頁邊界切在 tie 群內;
// 稽核的預設時間窗 auditDefaultWindow = 近 3 個月,fixture 時間戳因此取「現在」附近)。
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
	"time"

	"connectrpc.com/connect"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx database/sql driver(ent 以 OpenDB 包裝)

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/auditlog"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/ent/department"
	"github.com/salesorder/sales-order-1.0/backend/ent/metadict"
	"github.com/salesorder/sales-order-1.0/backend/ent/processingspec"
	"github.com/salesorder/sales-order-1.0/backend/ent/product"
	"github.com/salesorder/sales-order-1.0/backend/ent/productcategory"
	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/ent/route"
	"github.com/salesorder/sales-order-1.0/backend/ent/warehouse"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	auditv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/audit/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/audit/v1/auditv1connect"
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1/customersv1connect"
	mastersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1/mastersv1connect"
	metadictv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/metadict/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/metadict/v1/metadictv1connect"
	productsv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1/productsv1connect"
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
	// listScanSortOrderGroups 為固定排序鍵 `sort_order` 的同值群數(加工規格/商品分類/車次):
	// 96 / 4 = 每群 24 筆 > listScanPageSize,且 listScanPageSize 不整除 24 → 頁邊界必落在
	// 同值群內部(tie 被切開才會重複/遺漏)。欄位預設 0,生產中同值群常遠大於此。
	listScanSortOrderGroups = 4
	// listScanCodeScopes 為「唯一鍵含部門、可見集合跨部門」的兩張表(商品/倉別)的同值群數:
	// 同一個 code 在 8 個部門各建一筆 → 同值群 8 筆(listScanPageSize 不整除 8)。
	listScanCodeScopes = 8
	// listScanMetadictCodes 為字典 fixture 每個 type 的 code 數:六個 type(validMetadictTypes
	// 全值)× 16 個 code = 96 筆,系統預設與指定部門各一份 → 合併視角 192 筆。
	listScanMetadictCodes = 16
	// listScanMaxPages 是逐頁掃描的安全上限:正常情況會在最後一頁(不足 listScanPageSize)就結束,
	// 只有「重複/遺漏嚴重到掃不完」時才會撞上此上限而 fail(比照 F1 的終止保護)。
	listScanMaxPages = 32
	// listScanAuditStampGroups 為稽核 fixture 的 created_at 同值群數(M2):96 / 4 = 每群 24 筆
	// (> listScanPageSize 且 listScanPageSize 不整除 24 → 頁邊界必落在同一交易時間的稽核群內)。
	listScanAuditStampGroups = 4
)

// listScanner 以指定排序取回「某一頁」的 id 序列(逐頁掃描與全量掃描共用同一個呼叫點)。
type listScanner func(t *testing.T, sortField string, desc bool, page, pageSize int32) []string

// newListScanner 由「呼叫某端點取第 page 頁」的 closure 造出 listScanner:統一取出 id 與錯誤處理。
// 各清單的參數差異(sort/desc 的有無)由 closure 自行吸收 —— 沒有排序參數的端點(見 F2)忽略
// sortField/desc,一律以服務端固定的排序鍵查詢。
func newListScanner[T any](
	name string,
	call func(sortField string, desc bool, page, pageSize int32) ([]T, error),
	id func(T) string,
) listScanner {
	return func(t *testing.T, sortField string, desc bool, page, pageSize int32) []string {
		t.Helper()
		rows, err := call(sortField, desc, page, pageSize)
		if err != nil {
			t.Fatalf("%s(sort=%q desc=%v page=%d page_size=%d): %v", name, sortField, desc, page, pageSize, err)
		}
		ids := make([]string, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, id(r))
		}
		return ids
	}
}

// TestIntegrationListPageScanMatchesFullScan F1/P1-A/F2 迴歸:對各清單的每個排序欄位
// (含升/降冪與預設排序;F2 的六個端點沒有排序參數,只有服務端固定的排序鍵)
// 以多筆同值(NULl/重複 tax_id、同名、同 status、同 created_at、同 sort_order、同 code,
// 遠多於一頁)的資料,逐頁掃描的結果必須與全量一致:
//
//	筆數 == 全量筆數、id 集合 == 全量 id 集合、且無任何重複 id。
//
// 拿掉各 listSource 的 id 次序鍵 → 本測試必須紅(消去實驗見 f1-fix-report.md /
// p1a-report.md / tiebreak-all-report.md)。
func TestIntegrationListPageScanMatchesFullScan(t *testing.T) {
	// fixture 筆數與 id 集合都以「全新空庫」為前提(DB 真值即全量),覆寫模式下 skip。
	testsupport.RequiresContainer(t)
	ctx := t.Context()
	db := openPGEntClient(t, testsupport.Postgres(t))

	co := seedListScanCompanies(t, ctx, db)
	depts := seedListScanDepartments(t, ctx, db, co)
	seedListScanRoles(t, ctx, db)
	seedListScanCustomers(t, ctx, db, co.ID)
	seedListScanProcessingSpecs(t, ctx, db, co.ID, depts[0].ID)
	seedListScanProductCategories(t, ctx, db, co.ID, depts[0].ID)
	seedListScanRoutes(t, ctx, db, co.ID, depts[0].ID)
	seedListScanMetadicts(t, ctx, db, depts[0].ID)
	seedListScanProducts(t, ctx, db, co.ID, deptIDs(depts[:listScanCodeScopes]))
	seedListScanWarehouses(t, ctx, db, co.ID, deptIDs(depts[:listScanCodeScopes]))
	seedListScanAuditLogs(t, ctx, db, co.ID)

	cl := newListScanServer(t, db, co.ID)

	companyScan := newListScanner("ListCompanies",
		func(sortField string, desc bool, page, pageSize int32) ([]*v1.Company, error) {
			res, err := cl.companies.ListCompanies(ctx, connect.NewRequest(&v1.ListCompaniesRequest{
				Page: page, PageSize: pageSize, Sort: sortField, Desc: desc,
			}))
			if err != nil {
				return nil, err
			}
			return res.Msg.GetCompanies(), nil
		}, func(c *v1.Company) string { return c.GetId() })
	departmentScan := newListScanner("ListDepartments",
		func(sortField string, desc bool, page, pageSize int32) ([]*v1.Department, error) {
			res, err := cl.departments.ListDepartments(ctx, connect.NewRequest(&v1.ListDepartmentsRequest{
				Page: page, PageSize: pageSize, Sort: sortField, Desc: desc,
			}))
			if err != nil {
				return nil, err
			}
			return res.Msg.GetDepartments(), nil
		}, func(d *v1.Department) string { return d.GetId() })
	roleScan := newListScanner("ListRoles",
		func(sortField string, desc bool, page, pageSize int32) ([]*v1.Role, error) {
			res, err := cl.roles.ListRoles(ctx, connect.NewRequest(&v1.ListRolesRequest{
				Page: page, PageSize: pageSize, Sort: sortField, Desc: desc,
			}))
			if err != nil {
				return nil, err
			}
			return res.Msg.GetRoles(), nil
		}, func(r *v1.Role) string { return r.GetId() })
	// 客戶端點(D2 起有 desc)與公司/部門/角色同形;其餘六個端點(F2)連 sort 都沒有,
	// 一律以服務端固定的排序鍵查詢 —— sortField/desc 對這些 closure 只是共用的 listScanner 簽章。
	customerScan := newListScanner("ListCustomers",
		func(sortField string, desc bool, page, pageSize int32) ([]*customersv1.Customer, error) {
			res, err := cl.customers.ListCustomers(ctx, connect.NewRequest(&customersv1.ListCustomersRequest{
				Page: page, PageSize: pageSize, Sort: sortField, Desc: desc,
			}))
			if err != nil {
				return nil, err
			}
			return res.Msg.GetCustomers(), nil
		}, func(c *customersv1.Customer) string { return c.GetId() })
	// F2:以下六個端點無排序參數,排序鍵固定(加工規格/商品分類/車次 sort_order;字典
	// sort_order, code;商品/倉別 code),故 case 表的 field 恆為 ""。
	specScan := newListScanner("ListProcessingSpecs",
		func(_ string, _ bool, page, pageSize int32) ([]*mastersv1.ProcessingSpec, error) {
			res, err := cl.specs.ListProcessingSpecs(ctx, connect.NewRequest(&mastersv1.ListProcessingSpecsRequest{
				Page: page, PageSize: pageSize,
			}))
			if err != nil {
				return nil, err
			}
			return res.Msg.GetProcessingSpecs(), nil
		}, func(p *mastersv1.ProcessingSpec) string { return p.GetId() })
	catScan := newListScanner("ListProductCategories",
		func(_ string, _ bool, page, pageSize int32) ([]*mastersv1.ProductCategory, error) {
			res, err := cl.cats.ListProductCategories(ctx, connect.NewRequest(&mastersv1.ListProductCategoriesRequest{
				Page: page, PageSize: pageSize,
			}))
			if err != nil {
				return nil, err
			}
			return res.Msg.GetProductCategories(), nil
		}, func(c *mastersv1.ProductCategory) string { return c.GetId() })
	routeScan := newListScanner("ListRoutes",
		func(_ string, _ bool, page, pageSize int32) ([]*mastersv1.Route, error) {
			res, err := cl.routes.ListRoutes(ctx, connect.NewRequest(&mastersv1.ListRoutesRequest{
				Page: page, PageSize: pageSize,
			}))
			if err != nil {
				return nil, err
			}
			return res.Msg.GetRoutes(), nil
		}, func(r *mastersv1.Route) string { return r.GetId() })
	metadictScan := newListScanner("ListMetadicts",
		func(_ string, _ bool, page, pageSize int32) ([]*metadictv1.Metadict, error) {
			res, err := cl.metadicts.ListMetadicts(ctx, connect.NewRequest(&metadictv1.ListMetadictsRequest{
				Page: page, PageSize: pageSize,
			}))
			if err != nil {
				return nil, err
			}
			return res.Msg.GetItems(), nil
		}, func(m *metadictv1.Metadict) string { return m.GetId() })
	// metadictDeptScan 是「系統預設 + 當前部門」的合併視角(super 帶 department_id;部門身分的
	// metadictScope 產出等價的 where)。
	metadictDeptScan := newListScanner("ListMetadicts(department_id 指定)",
		func(_ string, _ bool, page, pageSize int32) ([]*metadictv1.Metadict, error) {
			res, err := cl.metadicts.ListMetadicts(ctx, connect.NewRequest(&metadictv1.ListMetadictsRequest{
				Page: page, PageSize: pageSize, DepartmentId: strconv.Itoa(depts[0].ID),
			}))
			if err != nil {
				return nil, err
			}
			return res.Msg.GetItems(), nil
		}, func(m *metadictv1.Metadict) string { return m.GetId() })
	productScan := newListScanner("ListProducts",
		func(_ string, _ bool, page, pageSize int32) ([]*productsv1.Product, error) {
			res, err := cl.products.ListProducts(ctx, connect.NewRequest(&productsv1.ListProductsRequest{
				Page: page, PageSize: pageSize,
			}))
			if err != nil {
				return nil, err
			}
			return res.Msg.GetProducts(), nil
		}, func(p *productsv1.Product) string { return p.GetId() })
	warehouseScan := newListScanner("ListWarehouses",
		func(_ string, _ bool, page, pageSize int32) ([]*mastersv1.Warehouse, error) {
			res, err := cl.warehouses.ListWarehouses(ctx, connect.NewRequest(&mastersv1.ListWarehousesRequest{
				Page: page, PageSize: pageSize,
			}))
			if err != nil {
				return nil, err
			}
			return res.Msg.GetWarehouses(), nil
		}, func(w *mastersv1.Warehouse) string { return w.GetId() })
	// M2:稽核端點同樣沒有排序參數(proto 無 sort/desc),服務端固定 created_at 降冪 + id 降冪。
	auditScan := newListScanner("ListAuditLogs",
		func(_ string, _ bool, page, pageSize int32) ([]*auditv1.AuditLog, error) {
			res, err := cl.audits.ListAuditLogs(ctx, connect.NewRequest(&auditv1.ListAuditLogsRequest{
				Page: page, PageSize: pageSize,
			}))
			if err != nil {
				return nil, err
			}
			return res.Msg.GetItems(), nil
		}, func(a *auditv1.AuditLog) string { return a.GetId() })

	// 白名單欄位(空字串 = 服務預設排序)皆須涵蓋,含升/降冪;F2 的六個端點沒有排序參數,
	// 只有固定排序鍵,故僅 "" 一列。
	//
	// G6:固定排序鍵的八個案例(加工規格/商品分類/車次/字典×2/商品/倉別/稽核)的全量基準
	// **刻意帶上預期方向的 ORDER BY** —— 集合比對對排序方向完全不敏感(服務端 Asc 改成 Desc、
	// 稽核的 created_at/id 改成 Asc 都仍全綠),故另以 assertSequenceMatchesFull 逐位比對
	// 「逐頁序列 == 本基準序列」。方向寫在測試裡、由同一個 PG 計算,建構與 tie 順序都不需
	// 用碼位/字典序猜測(排序鍵為唯一鍵時多一組等價鍵,結果不變)。
	all := map[string][]string{
		"公司": allIDs(t, func() ([]int, error) { return db.Company.Query().Select(company.FieldID).Ints(ctx) }),
		"部門": allIDs(t, func() ([]int, error) { return db.Department.Query().Select(department.FieldID).Ints(ctx) }),
		"角色": allIDs(t, func() ([]int, error) { return db.Role.Query().Select(role.FieldID).Ints(ctx) }),
		// 客戶的可見範圍為身分所屬公司(全部 fixture 客戶皆屬該公司)。
		"客戶": allIDs(t, func() ([]int, error) {
			return db.Customer.Query().Where(customer.CompanyIDEQ(co.ID)).Select(customer.FieldID).Ints(ctx)
		}),
		// F2:以下六個端點的可見範圍依 deptScope —— super 回 did=nil → 公司層(跨部門);
		// 字典另走 metadictScope:super 不過濾部門,系統預設 + 全部部門擴充皆可見(故無公司條件)。
		"加工規格": allIDs(t, func() ([]int, error) {
			return db.ProcessingSpec.Query().Where(processingspec.CompanyIDEQ(co.ID)).
				Order(ent.Asc(processingspec.FieldSortOrder), ent.Asc(processingspec.FieldID)).
				Select(processingspec.FieldID).Ints(ctx)
		}),
		"商品分類": allIDs(t, func() ([]int, error) {
			return db.ProductCategory.Query().Where(productcategory.CompanyIDEQ(co.ID)).
				Order(ent.Asc(productcategory.FieldSortOrder), ent.Asc(productcategory.FieldID)).
				Select(productcategory.FieldID).Ints(ctx)
		}),
		"車次": allIDs(t, func() ([]int, error) {
			return db.Route.Query().Where(route.CompanyIDEQ(co.ID)).
				Order(ent.Asc(route.FieldSortOrder), ent.Asc(route.FieldID)).
				Select(route.FieldID).Ints(ctx)
		}),
		"字典(系統預設)": allIDs(t, func() ([]int, error) {
			return db.Metadict.Query().Where(metadict.DepartmentIDIsNil()).
				Order(ent.Asc(metadict.FieldSortOrder), ent.Asc(metadict.FieldCode), ent.Asc(metadict.FieldID)).
				Select(metadict.FieldID).Ints(ctx)
		}),
		// 系統預設 + 指定部門(與部門身分的 metadictScope 等價的 where)。
		"字典(併當前部門)": allIDs(t, func() ([]int, error) {
			return db.Metadict.Query().
				Where(metadict.Or(metadict.DepartmentIDIsNil(), metadict.DepartmentIDEQ(depts[0].ID))).
				Order(ent.Asc(metadict.FieldSortOrder), ent.Asc(metadict.FieldCode), ent.Asc(metadict.FieldID)).
				Select(metadict.FieldID).Ints(ctx)
		}),
		"商品": allIDs(t, func() ([]int, error) {
			return db.Product.Query().Where(product.CompanyIDEQ(co.ID)).
				Order(ent.Asc(product.FieldCode), ent.Asc(product.FieldID)).
				Select(product.FieldID).Ints(ctx)
		}),
		"倉別": allIDs(t, func() ([]int, error) {
			return db.Warehouse.Query().Where(warehouse.CompanyIDEQ(co.ID)).
				Order(ent.Asc(warehouse.FieldCode), ent.Asc(warehouse.FieldID)).
				Select(warehouse.FieldID).Ints(ctx)
		}),
		// M2:稽核 fixture 全部落在預設時間窗(近 3 個月)內,故全量 = 該表全部列;
		// 契約是「最新在前」,故基準為 created_at 降冪 + id 降冪(G6)。
		"稽核": allIDs(t, func() ([]int, error) {
			return db.AuditLog.Query().
				Order(ent.Desc(auditlog.FieldCreatedAt), ent.Desc(auditlog.FieldID)).
				Select(auditlog.FieldID).Ints(ctx)
		}),
	}
	for _, tc := range []struct {
		entity string
		field  string
		scan   listScanner
		// ascOnly 表示該端點只有升冪一途:F2 的六個端點連 sort 都沒有(排序鍵固定),
		// 故 field 一律為 "";其餘端點皆掃升冪與降冪。
		ascOnly bool
	}{
		{"公司", "", companyScan, false},
		{"公司", "name", companyScan, false},
		{"公司", "identifier", companyScan, false},
		{"公司", "tax_id", companyScan, false},
		{"公司", "status", companyScan, false},
		{"公司", "id", companyScan, false},
		{"部門", "", departmentScan, false},
		{"部門", "name", departmentScan, false},
		{"部門", "id", departmentScan, false},
		{"角色", "", roleScan, false},
		{"角色", "code", roleScan, false},
		{"角色", "name", roleScan, false},
		{"角色", "id", roleScan, false},
		{"客戶", "", customerScan, false},
		{"客戶", "name", customerScan, false},
		{"客戶", "customer_code", customerScan, false},
		{"客戶", "created_at", customerScan, false},
		{"加工規格", "", specScan, true},
		{"商品分類", "", catScan, true},
		{"車次", "", routeScan, true},
		{"字典(系統預設)", "", metadictScan, true},
		{"字典(併當前部門)", "", metadictDeptScan, true},
		{"商品", "", productScan, true},
		{"倉別", "", warehouseScan, true},
		{"稽核", "", auditScan, true},
	} {
		descs := []bool{false, true}
		if tc.ascOnly {
			descs = []bool{false}
		}
		for _, desc := range descs {
			t.Run(fmt.Sprintf("%s/sort=%q/desc=%v", tc.entity, tc.field, desc), func(t *testing.T) {
				paged := scanPages(t, tc.scan, tc.field, desc)
				assertScanMatchesFull(t, tc.entity, tc.field, desc, paged, all[tc.entity])
				// G6:固定排序鍵的端點另釘住**方向**(含 tie 內的次序鍵方向)——
				// 逐頁序列必須逐位等於全量基準(Asc ↔ Desc 對稱變動時集合比對看不出來)。
				if tc.ascOnly {
					assertSequenceMatchesFull(t, tc.entity, tc.field, desc, paged, all[tc.entity])
				}
				// 降冪的方向須有實效:同一欄位(非空)的升冪與降冪頁序必須不同(D2)。
				if desc && !tc.ascOnly && tc.field != "" {
					assertDescReversesAsc(t, tc.entity, tc.field, scanPages(t, tc.scan, tc.field, false), paged)
				}
			})
		}
	}
}

// assertDescReversesAsc 斷言同一排序欄位(非空)的降冪頁序與升冪不同。
// 「desc 被忽略」的退化實作在此必紅 —— 逐頁掃描集合在兩個方向都會等於全量,
// 只有頁序看得出 desc 是否生效。方向本身的正確性由 sqlite 的映射鎖
// (TestListCustomersSortWhitelist)逐值釘住。
func assertDescReversesAsc(t *testing.T, entity, sortField string, asc, descPaged []string) {
	t.Helper()
	if slices.Equal(asc, descPaged) {
		t.Errorf("%s sort=%q:降冪頁序與升冪相同 —— desc 未被套用", entity, sortField)
	}
}

// scanPages 模擬前端逐頁翻頁:同一個排序鍵反覆查詢 page=1..N,直到某頁不足 listScanPageSize
// 筆為止,回傳依序串接的 id(含重複,交由斷言判別)。
func scanPages(t *testing.T, scan listScanner, sortField string, desc bool) []string {
	t.Helper()
	ids := make([]string, 0, listScanRows)
	for page := int32(1); page <= listScanMaxPages; page++ {
		got := scan(t, sortField, desc, page, listScanPageSize)
		ids = append(ids, got...)
		if len(got) < listScanPageSize {
			return ids
		}
	}
	t.Fatalf("sort=%q desc=%v:逐頁掃描超過 %d 頁仍未結束(F1 的重複/遺漏會讓掃描不終止)",
		sortField, desc, listScanMaxPages)
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

// assertSequenceMatchesFull 斷言逐頁序列與全量基準序列**逐位相同**(G6)。
//
// 全量基準帶有預期方向的 ORDER BY(見 `all` 的註解),故本斷言釘住的是排序方向:
// 服務端把 `Asc` 改成 `Desc`、或稽核的 `created_at DESC, id DESC` 改成升冪,逐頁仍會取回
// 同一個集合(assertScanMatchesFull 照樣全綠),但序列會反轉 → 在此必紅。
// 逐位比對同時涵蓋 tie 內的次序鍵方向;兩邊的次序都由同一個 PG 以同一組 ORDER BY 產生,
// 因此不依賴碼位/字典序的假設。
func assertSequenceMatchesFull(t *testing.T, entity, sortField string, desc bool, paged, full []string) {
	t.Helper()
	if len(paged) != len(full) {
		// 筆數不符已由 assertScanMatchesFull 報出原因(重複/遺漏),這裡只說明無法逐位比對。
		t.Errorf("%s sort=%q desc=%v:逐頁 %d 筆 != 全量 %d 筆,無法逐位比對排序方向",
			entity, sortField, desc, len(paged), len(full))
		return
	}
	for i := range paged {
		if paged[i] != full[i] {
			t.Errorf("%s sort=%q desc=%v:第 %d 筆起序列與預期方向不符 —— 逐頁 %v、預期 %v",
				entity, sortField, desc, i, listWindow(paged, i), listWindow(full, i))
			return
		}
	}
}

// listWindow 取序列中索引 i 附近的一小段(錯誤訊息只印出有差異的位置,不整份傾倒)。
func listWindow(ids []string, i int) []string {
	const span = 5
	lo := i
	if lo > len(ids)-span {
		lo = len(ids) - span
	}
	if lo < 0 {
		lo = 0
	}
	return ids[lo:min(lo+span, len(ids))]
}

// allIDs 取 DB 真值(該表 id 全量或指定序列),作為「全量」基準。
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

// listScanClients 為本測試用到的全部清單 client(同一 mux、同一 super 身分)。
// T10 起 users 一併掛上:跨租戶統一守門探針
// (rls_cross_tenant_integration_test.go)重用本 helper 時需要使用者端點。
type listScanClients struct {
	companies   salesorderv1connect.CompanyServiceClient
	departments salesorderv1connect.DepartmentServiceClient
	users       salesorderv1connect.UserServiceClient
	roles       salesorderv1connect.RoleServiceClient
	customers   customersv1connect.CustomerServiceClient
	specs       mastersv1connect.ProcessingSpecServiceClient
	cats        mastersv1connect.ProductCategoryServiceClient
	routes      mastersv1connect.RouteServiceClient
	metadicts   metadictv1connect.MetadictServiceClient
	products    productsv1connect.ProductServiceClient
	warehouses  mastersv1connect.WarehouseServiceClient
	audits      auditv1connect.AuditServiceClient
}

// newListScanServer 以 super 身分(全權,涵蓋 requireScope/requireRole 門檻與 deptScope 的
// 公司範圍)把十一張清單的 handler 掛在單一 mux 上(與 sqlite 版 newTestServer 同構,只換 DB)。
// companyID 為 fixture 所屬公司:各端點以 deptScope 解析身分的 CompanyID 作為可見範圍,
// 故須為數字。
//
// 本函式只決定「原清單測試需要的 scope」:super/公司層 + scope=all。需要別的 scope
// (例如跨租戶守門探針要的部門層級 A 範圍)者用 newListScanServerWithScope。
func newListScanServer(t *testing.T, db *ent.Client, companyID int) listScanClients {
	t.Helper()
	return newListScanServerWithScope(t, db, auth.RLSScope{
		UserID: "1", CompanyID: strconv.Itoa(companyID),
		DataScope: auth.DataScopeAll, CompanyActive: true,
	})
}

// newListScanServerWithScope 與 newListScanServer 同構,只多注入呼叫端指定的 RLS scope:
// 身份固定為 super(ACL 全開,讓紅的一定是 RLS 這條接線),請求層租戶交易由各
// Register*Services 內的 dbtenant.HandlerOption 依 ctx 的 scope 套用(SET LOCAL 在
// driver 裝飾器的 Tx(ctx) 內執行)。
//
// 因此**呼叫端必須傳入 dbtenant.NewClient 建立的 client**(app_rw 連線),否則沒有裝飾器、
// 也就沒有 SET LOCAL。
func newListScanServerWithScope(t *testing.T, db *ent.Client, scope auth.RLSScope) listScanClients {
	t.Helper()
	super := authz.Identity{
		UserID: scope.UserID, CompanyID: scope.CompanyID, Role: "super", Roles: []string{"super"},
	}
	mux := http.NewServeMux()
	RegisterCompanyServices(mux, db)
	RegisterAuditServices(mux, db)
	RegisterUserServices(mux, db)
	RegisterRoleServices(mux, db)
	RegisterCustomerServices(mux, db, "http://localhost:3000")
	RegisterProcessingSpecService(mux, db)
	RegisterProductCategoryService(mux, db)
	RegisterRouteService(mux, db)
	RegisterMetadictServices(mux, db)
	RegisterProductService(mux, db)
	RegisterWarehouseService(mux, db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), super)
		ctx = auth.WithRLS(ctx, scope)
		ctx = authz.WithCASLEnabled(ctx, true)
		ctx = authz.WithDB(ctx, db)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	return listScanClients{
		companies:   salesorderv1connect.NewCompanyServiceClient(http.DefaultClient, ts.URL),
		departments: salesorderv1connect.NewDepartmentServiceClient(http.DefaultClient, ts.URL),
		users:       salesorderv1connect.NewUserServiceClient(http.DefaultClient, ts.URL),
		roles:       salesorderv1connect.NewRoleServiceClient(http.DefaultClient, ts.URL),
		customers:   customersv1connect.NewCustomerServiceClient(http.DefaultClient, ts.URL),
		specs:       mastersv1connect.NewProcessingSpecServiceClient(http.DefaultClient, ts.URL),
		cats:        mastersv1connect.NewProductCategoryServiceClient(http.DefaultClient, ts.URL),
		routes:      mastersv1connect.NewRouteServiceClient(http.DefaultClient, ts.URL),
		metadicts:   metadictv1connect.NewMetadictServiceClient(http.DefaultClient, ts.URL),
		products:    productsv1connect.NewProductServiceClient(http.DefaultClient, ts.URL),
		warehouses:  mastersv1connect.NewWarehouseServiceClient(http.DefaultClient, ts.URL),
		audits:      auditv1connect.NewAuditServiceClient(http.DefaultClient, ts.URL),
	}
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

// seedListScanDepartments 在單一公司下建立 listScanRows 個部門並回傳(供 F2 的商品/倉別/
// 字典 fixture 掛載):name 每 24 筆同名(部門無其他排序欄位,同值群須大於 listScanPageSize
// 才會被頁邊界切開)。
func seedListScanDepartments(t *testing.T, ctx context.Context, db *ent.Client, co *ent.Company) []*ent.Department {
	t.Helper()
	builders := make([]*ent.DepartmentCreate, 0, listScanRows)
	for i := range listScanRows {
		builders = append(builders, db.Department.Create().
			SetName(fmt.Sprintf("部門%02d", i/24)).
			SetCompanyID(co.ID))
	}
	created, err := db.Department.CreateBulk(builders...).Save(ctx)
	if err != nil {
		t.Fatalf("建立部門 fixture: %v", err)
	}
	return created
}

// deptIDs 取出部門 id(seedListScan* 系列的 ent builder 只吃 id)。
func deptIDs(depts []*ent.Department) []int {
	out := make([]int, 0, len(depts))
	for _, d := range depts {
		out = append(out, d.ID)
	}
	return out
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

// seedListScanCustomers 在指定公司下建立 listScanRows 個客戶(P1-A 迴歸):name 每 24 筆同名、
// created_at 每 24 筆同一時間戳(兩者皆非唯一,同值群遠大於 listScanPageSize);customer_code
// 於同公司內唯一(生產端由取號 + migration 00013 的部分唯一索引保證),故為對照組 ——
// 排序鍵唯一時逐頁掃描本就不會重複/遺漏,消去實驗中會維持綠。
func seedListScanCustomers(t *testing.T, ctx context.Context, db *ent.Client, companyID int) {
	t.Helper()
	// 客戶域已 ENABLE(+FORCE):fixture 寫入走明確的系統範圍入口(見 seedTx)。
	seedTx(t, db, func(tx *ent.Tx) error {
		base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		builders := make([]*ent.CustomerCreate, 0, listScanRows)
		for i := range listScanRows {
			builders = append(builders, tx.Customer.Create().
				SetCompanyID(companyID).
				SetCustomerCode(fmt.Sprintf("C-%03d", i)).
				SetName(fmt.Sprintf("客戶%02d", i/24)).
				SetCreatedAt(base.AddDate(0, 0, i/24)))
		}
		_, err := tx.Customer.CreateBulk(builders...).Save(ctx)
		return err
	})
}

// seedListScanProcessingSpecs 建立 listScanRows 個加工規格(掛同一部門):固定排序鍵
// `sort_order` 每 listScanRows/listScanSortOrderGroups 筆同值(共 4 群,群內 code 互異)——
// 生產中 sort_order 欄位預設 0,同值群常遠大於一頁,這裡造出「大於一頁且不整除頁寬」的同值群
// 以確保頁邊界落在群內。
func seedListScanProcessingSpecs(t *testing.T, ctx context.Context, db *ent.Client, companyID, departmentID int) {
	t.Helper()
	builders := make([]*ent.ProcessingSpecCreate, 0, listScanRows)
	for i := range listScanRows {
		builders = append(builders, db.ProcessingSpec.Create().
			SetCompanyID(companyID).
			SetDepartmentID(departmentID).
			SetCode(fmt.Sprintf("SPEC-%03d", i)).
			SetName(fmt.Sprintf("加工規格%02d", i)).
			SetSortOrder(i/(listScanRows/listScanSortOrderGroups)))
	}
	if _, err := db.ProcessingSpec.CreateBulk(builders...).Save(ctx); err != nil {
		t.Fatalf("建立加工規格 fixture: %v", err)
	}
}

// seedListScanProductCategories 建立 listScanRows 個商品分類(掛同一部門):固定排序鍵
// `sort_order` 每 24 筆同值(同 seedListScanProcessingSpecs 的造法);code 於部門內唯一。
func seedListScanProductCategories(t *testing.T, ctx context.Context, db *ent.Client, companyID, departmentID int) {
	t.Helper()
	builders := make([]*ent.ProductCategoryCreate, 0, listScanRows)
	for i := range listScanRows {
		builders = append(builders, db.ProductCategory.Create().
			SetCompanyID(companyID).
			SetDepartmentID(departmentID).
			SetCode(fmt.Sprintf("CAT-%03d", i)).
			SetName(fmt.Sprintf("商品分類%02d", i)).
			SetSortOrder(i/(listScanRows/listScanSortOrderGroups)))
	}
	if _, err := db.ProductCategory.CreateBulk(builders...).Save(ctx); err != nil {
		t.Fatalf("建立商品分類 fixture: %v", err)
	}
}

// seedListScanRoutes 建立 listScanRows 個車次(掛同一部門):固定排序鍵 `sort_order` 每 24 筆
// 同值;code 於部門內唯一。
func seedListScanRoutes(t *testing.T, ctx context.Context, db *ent.Client, companyID, departmentID int) {
	t.Helper()
	builders := make([]*ent.RouteCreate, 0, listScanRows)
	for i := range listScanRows {
		builders = append(builders, db.Route.Create().
			SetCompanyID(companyID).
			SetDepartmentID(departmentID).
			SetCode(fmt.Sprintf("ROUTE-%03d", i)).
			SetName(fmt.Sprintf("車次%02d", i)).
			SetSortOrder(i/(listScanRows/listScanSortOrderGroups)))
	}
	if _, err := db.Route.CreateBulk(builders...).Save(ctx); err != nil {
		t.Fatalf("建立車次 fixture: %v", err)
	}
}

// seedListScanMetadicts 建立 2 × listScanMetadictCodes × 6 個字典:排序鍵是 `(sort_order, code)`,
// 而 code 的生產唯一性只有 `(type, 部門)`(migration 00011 的兩道條件索引)。故同一批 code 各建在
// 六個 type(validMetadictTypes 全值)上,並各再建一份屬於 departmentID 的部門擴充 ——
//   - super 未指定 department_id(管理視角):可見集合 = 系統預設(96 筆),同一 code 有 6 筆同值;
//   - super 指定 department_id / 部門身分:可見集合 = 系統預設 + 該部門(192 筆),同一 code 有
//     12 筆同值(metadictScope 對部門身分的 where 與 super 帶 department_id 時完全相同)。
//
// 兩者的同值群都不整除 listScanPageSize,頁邊界必落在群內。
func seedListScanMetadicts(t *testing.T, ctx context.Context, db *ent.Client, departmentID int) {
	t.Helper()
	types := []string{"unit", "payment_method", "settlement_method", "customer_type", "invoice_type", "order_source"}
	perScope := listScanMetadictCodes * len(types)
	builders := make([]*ent.MetadictCreate, 0, 2*perScope)
	for scope := range 2 { // 0 = 系統預設(department_id NULL);1 = departmentID 的部門擴充
		for i := range perScope {
			b := db.Metadict.Create().
				SetType(types[i%len(types)]).
				SetCode(fmt.Sprintf("MD-%02d", i/len(types))).
				SetDisplayName(fmt.Sprintf("字典%02d-%02d", scope, i))
			if scope == 1 {
				b = b.SetDepartmentID(departmentID)
			}
			builders = append(builders, b)
		}
	}
	if _, err := db.Metadict.CreateBulk(builders...).Save(ctx); err != nil {
		t.Fatalf("建立字典 fixture: %v", err)
	}
}

// seedListScanProducts 建立 listScanRows 個商品:code 的唯一性是 `(department_id, code)`
// (migration 00017 的部分唯一索引),故在 listScanCodeScopes 個部門各建一組相同 code ——
// 每個 code 有 8 筆同值群。super/company_admin 的 deptScope 回 did=nil → 可見集合跨部門,
// 正是生產中「同 code 落在同一頁序」的成因。
func seedListScanProducts(t *testing.T, ctx context.Context, db *ent.Client, companyID int, departmentIDs []int) {
	t.Helper()
	codes := listScanRows / listScanCodeScopes
	builders := make([]*ent.ProductCreate, 0, listScanRows)
	for i := range listScanRows {
		builders = append(builders, db.Product.Create().
			SetCompanyID(companyID).
			SetDepartmentID(departmentIDs[i/codes]).
			SetCode(fmt.Sprintf("PD-%03d", i%codes)).
			SetName(fmt.Sprintf("商品%02d", i)))
	}
	if _, err := db.Product.CreateBulk(builders...).Save(ctx); err != nil {
		t.Fatalf("建立商品 fixture: %v", err)
	}
}

// seedListScanWarehouses 建立 listScanRows 個倉別:同 seedListScanProducts,code 的唯一性是
// `(department_id, code)`(migration 00016),8 個部門各一組相同 code → 每個 code 8 筆同值群。
func seedListScanWarehouses(t *testing.T, ctx context.Context, db *ent.Client, companyID int, departmentIDs []int) {
	t.Helper()
	codes := listScanRows / listScanCodeScopes
	builders := make([]*ent.WarehouseCreate, 0, listScanRows)
	for i := range listScanRows {
		builders = append(builders, db.Warehouse.Create().
			SetCompanyID(companyID).
			SetDepartmentID(departmentIDs[i/codes]).
			SetCode(fmt.Sprintf("WH-%03d", i%codes)).
			SetName(fmt.Sprintf("倉別%02d", i)))
	}
	if _, err := db.Warehouse.CreateBulk(builders...).Save(ctx); err != nil {
		t.Fatalf("建立倉別 fixture: %v", err)
	}
}

// seedListScanAuditLogs 建立 listScanRows 筆稽核(M2):`created_at` 每 24 筆同一時間戳
// (listScanAuditStampGroups 群)—— 同一業務交易會以同一交易時間寫入多筆稽核(2.6.2/D18),
// 故「同一 created_at 的稽核遠多於一頁」是生產資料的常態形狀,而非特例。
//
// 時間戳取「現在附近」:ListAuditLogs 未帶 from/to 時套用預設時間窗(auditDefaultWindow =
// 近 3 個月),fixture 必須落在窗內才看得到(固定日期會隨時間滑出窗外)。
func seedListScanAuditLogs(t *testing.T, ctx context.Context, db *ent.Client, companyID int) {
	t.Helper()
	base := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	builders := make([]*ent.AuditLogCreate, 0, listScanRows)
	for i := range listScanRows {
		builders = append(builders, db.AuditLog.Create().
			SetCompanyID(companyID).
			SetUserID(1).
			SetAction(auditlog.ActionCreate).
			SetResourceType("customer").
			SetResourceID(fmt.Sprintf("AUD-%03d", i)).
			SetCreatedAt(base.Add(time.Duration(i/(listScanRows/listScanAuditStampGroups))*time.Second)))
	}
	if _, err := db.AuditLog.CreateBulk(builders...).Save(ctx); err != nil {
		t.Fatalf("建立稽核 fixture: %v", err)
	}
}

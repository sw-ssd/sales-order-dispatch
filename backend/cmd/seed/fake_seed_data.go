// 示範資料 seeder（僅 development/staging）：把各頁面從空狀態填成有內容的真實樣貌，
// 供 Pixso 設計稿擷取（docs/design）與本機手動驗收使用。
//
// 與 cmd/seed 的職責切分：
//   - platform.go／roles.go 是**生產也要跑**的基礎資料（角色、權限、方案、平台 operator）；
//   - 本檔是**純示範資料**，production 一律跳過（main.go 依 cfg.API.Env 把關），
//     不得被任何生產路徑引用，也不得成為測試的唯一資料來源。
//
// 為何走 SystemScopeTx：這些是業務表（customers/sales_orders/...全數 ENABLE+FORCE RLS），
// 且 seed 不帶請求身分，必須以系統範圍（scope=all）寫入 —— 與 roles.go 同一條路徑
// （未包系統範圍會被 WITH CHECK 以 42501 擋下，見 seed_rls_integration_test.go）。
//
// 冪等：以固定的示範公司識別碼（fakeCompanyIdentifier）為錨，先刪除該公司既有的示範列再重建；
// 重跑不會累積、也不會碰非示範公司的資料。
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/auditlog"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/ent/customerproduct"
	"github.com/salesorder/sales-order-1.0/backend/ent/department"
	"github.com/salesorder/sales-order-1.0/backend/ent/fileasset"
	"github.com/salesorder/sales-order-1.0/backend/ent/notification"
	"github.com/salesorder/sales-order-1.0/backend/ent/notificationtemplate"
	"github.com/salesorder/sales-order-1.0/backend/ent/ordercounter"
	"github.com/salesorder/sales-order-1.0/backend/ent/printlog"
	"github.com/salesorder/sales-order-1.0/backend/ent/processingspec"
	"github.com/salesorder/sales-order-1.0/backend/ent/product"
	"github.com/salesorder/sales-order-1.0/backend/ent/productcategory"
	"github.com/salesorder/sales-order-1.0/backend/ent/returnrequest"
	"github.com/salesorder/sales-order-1.0/backend/ent/returnrequestitem"
	"github.com/salesorder/sales-order-1.0/backend/ent/route"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorderitem"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/ent/warehouse"
)

// fakeCompanyIdentifier 示範公司的識別碼；也是本 seeder 的「這是我的資料」錨點。
// 用固定字串而非名稱，才能在重跑時精準圈出可安全刪除的範圍。
const fakeCompanyIdentifier = "FAKE-DEMO"

// fakeCompanyIdentifier2 第二家示範公司，讓「公司管理」頁有多列可看。
const fakeCompanyIdentifier2 = "FAKE-DEMO-2"

// fakeUserEmail 示範公司的管理者帳號，作為各示範列的 created_by／sales_rep 等操作者。
const fakeUserEmail = "fake-admin@example.test"

// SeedFakeData 建立示範租戶與其業務資料（公司／部門／使用者／主檔／訂單／退貨／列印／通知）。
// 回傳 false 表示環境不允許（production），由呼叫端決定是否提示。
//
// `storageRoot` 是檔案儲存根目錄（`config.Storage.StorageRoot`）：列印紀錄的 file_assets 列
// 必須有對應的 PDF 檔在磁碟上，否則頁面的「下載」連結會 404（只剩資料庫列，見 seedPrintLogs）。
//
// 只建立**設計稿需要看到的狀態**，不追求業務完整性（例如不做庫存、不跑 OpenFGA provisioning）。
func SeedFakeData(ctx context.Context, client *ent.Client, env, storageRoot string) error {
	if env == "production" {
		return nil
	}
	// 平台自營公司是系統自己的租戶，示範資料不得汙染它（firstCompanyID 有同款防護）。
	if err := assertNotPlatformScope(ctx, client); err != nil {
		return err
	}
	return seedFakeTenant(ctx, client, storageRoot)
}

// assertNotPlatformScope 確認連線目前不在平台自營公司的範圍內。
//
// SystemScopeTx 是 scope=all（不限定公司），但本 seeder 全部以 company_id 明確指定寫入對象，
// 唯一風險是「示範公司的識別碼」與平台自營公司的識別碼撞名 —— 撞名會導致下面的清理步驟
// 刪掉平台租戶的資料。識別碼是常數，這裡做一次 belt-and-suspenders 檢查。
func assertNotPlatformScope(ctx context.Context, client *ent.Client) error {
	clash, err := client.Company.Query().
		Where(company.IdentifierIn(fakeCompanyIdentifier, fakeCompanyIdentifier2),
			company.IdentifierEQ(platformCompanyIdentifier)).
		Exist(ctx)
	if err != nil {
		return err
	}
	if clash {
		return fmt.Errorf("示範公司識別碼與平台自營公司（%s）衝突", platformCompanyIdentifier)
	}
	return nil
}

// seedFakeTenant 是 seeder 主體。所有寫入都帶明確的 company_id／department_id。
func seedFakeTenant(ctx context.Context, client *ent.Client, storageRoot string) error {
	// 先清掉上一輪的示範列（子表先刪，避開 FK）。範圍嚴格限定在示範公司內。
	if err := clearFakeData(ctx, client); err != nil {
		return err
	}

	companies, err := ensureFakeCompanies(ctx, client)
	if err != nil {
		return err
	}

	for _, co := range companies {
		if err := seedFakeCompany(ctx, client, co.id, co.identifier, co.withData, storageRoot); err != nil {
			return fmt.Errorf("示範公司 %s: %w", co.identifier, err)
		}
	}
	return nil
}

// fakeCompany 是本次要建立／清理的示範公司。
type fakeCompany struct {
	id         int
	identifier string
	// withData 表示是否要建立完整的業務資料（第二家只放部門，讓公司頁有多列）。
	withData bool
}

// ensureFakeCompanies 冪等建立兩家示範公司，回傳其 id。
func ensureFakeCompanies(ctx context.Context, client *ent.Client) ([]fakeCompany, error) {
	specs := []struct {
		identifier string
		name       string
		withData   bool
	}{
		{fakeCompanyIdentifier, "示範食品股份有限公司", true},
		{fakeCompanyIdentifier2, "大川水產有限公司", false},
	}
	out := make([]fakeCompany, 0, len(specs))
	for _, s := range specs {
		id, err := ensureFakeCompany(ctx, client, s.identifier, s.name)
		if err != nil {
			return nil, err
		}
		out = append(out, fakeCompany{id: id, identifier: s.identifier, withData: s.withData})
	}
	return out, nil
}

// ensureFakeCompany 依 identifier 取回或建立公司（含軟刪除者復原，避免唯一鍵衝突）。
func ensureFakeCompany(ctx context.Context, client *ent.Client, identifier, name string) (int, error) {
	// 唯一鍵是不分軟刪除的，故查詢時**不**過濾 deleted_at：撞到軟刪除列就復原它。
	existing, err := client.Company.Query().
		Where(company.IdentifierEQ(identifier)).
		Only(ctx)
	if err == nil {
		if existing.DeletedAt != nil {
			if _, err := client.Company.UpdateOneID(existing.ID).ClearDeletedAt().Save(ctx); err != nil {
				return 0, err
			}
		}
		return existing.ID, nil
	}
	if !ent.IsNotFound(err) {
		return 0, err
	}
	co, err := client.Company.Create().
		SetName(name).
		SetIdentifier(identifier).
		SetStatus("active").
		SetLogoURL("").
		Save(ctx)
	if err != nil {
		return 0, err
	}
	return co.ID, nil
}

// clearFakeData 刪除示範公司的既有業務列（順序：子表 → 父表）。
// 只碰 company_id 屬於示範公司的列；主檔（公司本身）保留，由 ensureFakeCompanies 復用。
func clearFakeData(ctx context.Context, client *ent.Client) error {
	ids, err := fakeCompanyIDs(ctx, client)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	// 列印紀錄的 file_asset_id 有 FK，先取出來再刪，最後清掉對應的 file_assets。
	printFileIDs, err := client.PrintLog.Query().
		Where(printlog.CompanyIDIn(ids...)).
		Select(printlog.FieldFileAssetID).
		Ints(ctx)
	if err != nil {
		return err
	}

	deleteSteps := []func() error{
		func() error {
			_, err := client.ReturnRequestItem.Delete().Where(returnrequestitem.CompanyIDIn(ids...)).Exec(ctx)
			return err
		},
		func() error {
			_, err := client.ReturnRequest.Delete().Where(returnrequest.CompanyIDIn(ids...)).Exec(ctx)
			return err
		},
		// sales_order_events **不直接刪**：它對 app_rw 只開 SELECT/INSERT（append-only，D10），
		// 直接 DELETE 會是 42501。刪父表 sales_orders 時由 sales_order_events_order_fk 的
		// ON DELETE CASCADE 帶走（已實測）。這也讓「事件流不可竄改」的保證在 seeder 面前同樣成立。
		func() error {
			_, err := client.SalesOrderItem.Delete().Where(salesorderitem.CompanyIDIn(ids...)).Exec(ctx)
			return err
		},
		func() error {
			_, err := client.SalesOrder.Delete().Where(salesorder.CompanyIDIn(ids...)).Exec(ctx)
			return err
		},
		func() error {
			_, err := client.CustomerProduct.Delete().Where(customerproduct.CompanyIDIn(ids...)).Exec(ctx)
			return err
		},
		func() error {
			_, err := client.PrintLog.Delete().Where(printlog.CompanyIDIn(ids...)).Exec(ctx)
			return err
		},
		func() error {
			_, err := client.Notification.Delete().Where(notification.CompanyIDIn(ids...)).Exec(ctx)
			return err
		},
		func() error {
			_, err := client.NotificationTemplate.Delete().Where(notificationtemplate.CompanyIDIn(ids...)).Exec(ctx)
			return err
		},
		func() error {
			_, err := client.Customer.Delete().Where(customer.CompanyIDIn(ids...)).Exec(ctx)
			return err
		},
		func() error {
			_, err := client.Product.Delete().Where(product.CompanyIDIn(ids...)).Exec(ctx)
			return err
		},
		func() error {
			_, err := client.Warehouse.Delete().Where(warehouse.CompanyIDIn(ids...)).Exec(ctx)
			return err
		},
		func() error {
			_, err := client.Route.Delete().Where(route.CompanyIDIn(ids...)).Exec(ctx)
			return err
		},
		func() error {
			_, err := client.ProductCategory.Delete().Where(productcategory.CompanyIDIn(ids...)).Exec(ctx)
			return err
		},
		func() error {
			_, err := client.ProcessingSpec.Delete().Where(processingspec.CompanyIDIn(ids...)).Exec(ctx)
			return err
		},
		func() error {
			_, err := client.OrderCounter.Delete().Where(ordercounter.CompanyIDIn(ids...)).Exec(ctx)
			return err
		},
		func() error {
			if len(printFileIDs) == 0 {
				return nil
			}
			_, err := client.FileAsset.Delete().Where(fileasset.IDIn(printFileIDs...)).Exec(ctx)
			return err
		},
	}
	for i, step := range deleteSteps {
		if err := step(); err != nil {
			return fmt.Errorf("清理示範資料（步驟 %d）: %w", i, err)
		}
	}
	return nil
}

// fakeCompanyIDs 取回示範公司的 id（含軟刪除者，清理時才不會漏）。
func fakeCompanyIDs(ctx context.Context, client *ent.Client) ([]int, error) {
	return client.Company.Query().
		Where(company.IdentifierIn(fakeCompanyIdentifier, fakeCompanyIdentifier2)).
		IDs(ctx)
}

// seedFakeCompany 建立單一示範公司的資料。noData 為 true 時只建立部門（公司頁多列用）。
func seedFakeCompany(ctx context.Context, client *ent.Client, cid int, identifier string, withData bool, storageRoot string) error {
	deptID, err := ensureDepartment(ctx, client, cid, "總公司")
	if err != nil {
		return err
	}
	if !withData {
		if _, err := ensureDepartment(ctx, client, cid, "台中營業所"); err != nil {
			return err
		}
		return nil
	}
	if _, err := ensureDepartment(ctx, client, cid, "台北營業所"); err != nil {
		return err
	}

	userID, err := ensureFakeUser(ctx, client, cid, deptID)
	if err != nil {
		return err
	}
	wh1, wh2, err := seedWarehouses(ctx, client, cid, deptID, userID)
	if err != nil {
		return err
	}
	routes, err := seedRoutes(ctx, client, cid, deptID, userID)
	if err != nil {
		return err
	}
	if err := seedCategories(ctx, client, cid, deptID, userID); err != nil {
		return err
	}
	if _, err := seedProcessingSpecs(ctx, client, cid, deptID, userID); err != nil {
		return err
	}
	products, err := seedProducts(ctx, client, cid, deptID, wh1, wh2, userID)
	if err != nil {
		return err
	}
	customers, err := seedCustomers(ctx, client, cid, deptID, userID)
	if err != nil {
		return err
	}
	if err := seedCustomerProducts(ctx, client, cid, deptID, userID, customers, products); err != nil {
		return err
	}
	if err := seedOrders(ctx, client, cid, deptID, userID, customers, products, routes); err != nil {
		return err
	}
	if err := seedReturns(ctx, client, cid, deptID, userID, customers, products); err != nil {
		return err
	}
	if err := seedNotifications(ctx, client, cid, deptID, userID); err != nil {
		return err
	}
	if err := seedPrintLogs(ctx, client, cid, deptID, userID, routes, storageRoot); err != nil {
		return err
	}
	return seedAuditLogs(ctx, client, cid, deptID, userID)
}

// ensureDepartment 冪等建立部門（軟刪除者復原；部門名稱在同一公司內不重複）。
func ensureDepartment(ctx context.Context, client *ent.Client, cid int, name string) (int, error) {
	existing, err := client.Department.Query().
		Where(department.NameEQ(name), department.HasCompanyWith(company.IDEQ(cid))).
		Only(ctx)
	if err == nil {
		if existing.DeletedAt != nil {
			if _, err := client.Department.UpdateOneID(existing.ID).ClearDeletedAt().Save(ctx); err != nil {
				return 0, err
			}
		}
		return existing.ID, nil
	}
	if !ent.IsNotFound(err) {
		return 0, err
	}
	d, err := client.Department.Create().
		SetName(name).
		SetCompanyID(cid).
		Save(ctx)
	if err != nil {
		return 0, err
	}
	return d.ID, nil
}

// ensureFakeUser 冪等建立示範公司管理者帳號（email 為本 seeder 專屬，不會撞到真人帳號）。
func ensureFakeUser(ctx context.Context, client *ent.Client, cid, deptID int) (int, error) {
	existing, err := client.User.Query().Where(user.EmailEQ(fakeUserEmail)).Only(ctx)
	if err == nil {
		return existing.ID, nil
	}
	if !ent.IsNotFound(err) {
		return 0, err
	}
	u, err := client.User.Create().
		SetEmail(fakeUserEmail).
		SetName("示範管理者").
		SetRole("company_admin").
		SetStatus(user.StatusActive).
		SetPasswordHash("!").
		SetCompanyID(cid).
		SetDepartmentID(deptID).
		Save(ctx)
	if err != nil {
		return 0, err
	}
	return u.ID, nil
}

// seedWarehouses 建立兩個倉別，回傳 1F（庫存）與 2F（揀貨）的 id。
func seedWarehouses(ctx context.Context, client *ent.Client, cid, did, actor int) (int, int, error) {
	specs := []struct{ code, name, addr string }{
		{"W01", "1F 冷藏倉", "台北市中山區建國北路二段 100 號"},
		{"W02", "2F 冷凍倉", "台北市中山區建國北路二段 100 號 2F"},
	}
	ids := make([]int, 0, len(specs))
	for _, s := range specs {
		w, err := client.Warehouse.Create().
			SetCompanyID(cid).SetDepartmentID(did).
			SetCode(s.code).SetName(s.name).SetAddress(s.addr).
			SetIsActive(true).SetCreatedBy(actor).SetUpdatedBy(actor).
			Save(ctx)
		if err != nil {
			return 0, 0, err
		}
		ids = append(ids, w.ID)
	}
	return ids[0], ids[1], nil
}

// seedRoutes 建立三個車次，回傳其 id（依 code 排序，索引對應 1車/2車/3車）。
func seedRoutes(ctx context.Context, client *ent.Client, cid, did, actor int) ([]int, error) {
	specs := []struct {
		code, name, desc string
		sort             int
	}{
		{"R01", "1車", "台北東區上午線", 1},
		{"R02", "2車", "台北西區下午線", 2},
		{"R03", "3車", "新北板橋線", 3},
	}
	ids := make([]int, 0, len(specs))
	for _, s := range specs {
		r, err := client.Route.Create().
			SetCompanyID(cid).SetDepartmentID(did).
			SetCode(s.code).SetName(s.name).SetDescription(s.desc).SetSortOrder(s.sort).
			SetIsActive(true).SetCreatedBy(actor).SetUpdatedBy(actor).
			Save(ctx)
		if err != nil {
			return nil, err
		}
		ids = append(ids, r.ID)
	}
	return ids, nil
}

// seedCategories 建立商品分類。
func seedCategories(ctx context.Context, client *ent.Client, cid, did, actor int) error {
	specs := []struct {
		code, name string
		sort       int
	}{
		{"C01", "豬肉", 1}, {"C02", "雞肉", 2}, {"C03", "海鮮", 3}, {"C04", "加工品", 4},
	}
	for _, s := range specs {
		if _, err := client.ProductCategory.Create().
			SetCompanyID(cid).SetDepartmentID(did).
			SetCode(s.code).SetName(s.name).SetSortOrder(s.sort).
			SetIsActive(true).SetCreatedBy(actor).SetUpdatedBy(actor).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

// seedProcessingSpecs 建立分切規格，回傳 code → id（商品頁的加工欄位用）。
func seedProcessingSpecs(ctx context.Context, client *ent.Client, cid, did, actor int) (map[string]int, error) {
	specs := []struct {
		code, name, kind string
		proc, pick       bool
		attrs            map[string]any
		sort             int
	}{
		{"P01", "薄切", "cut", true, false, map[string]any{"thickness_mm": 2}, 1},
		{"P02", "厚切", "cut", true, false, map[string]any{"thickness_mm": 8}, 2},
		{"P03", "切塊", "cut", true, false, map[string]any{"size_cm": 3}, 3},
		{"P04", "絞肉", "cut", true, false, map[string]any{"grind_mm": 5}, 4},
		{"P05", "去骨", "pick", false, true, map[string]any{}, 5},
	}
	out := make(map[string]int, len(specs))
	for _, s := range specs {
		p, err := client.ProcessingSpec.Create().
			SetCompanyID(cid).SetDepartmentID(did).
			SetCode(s.code).SetName(s.name).SetKind(s.kind).
			SetAppliesToProcessing(s.proc).SetAppliesToPicking(s.pick).
			SetAttributes(s.attrs).SetSortOrder(s.sort).
			SetIsActive(true).SetCreatedBy(actor).SetUpdatedBy(actor).
			Save(ctx)
		if err != nil {
			return nil, err
		}
		out[s.code] = p.ID
	}
	return out, nil
}

// fakeProduct 是示範商品的精簡描述（同時供訂單／退貨明細引用）。
type fakeProduct struct {
	id   int
	code string
	name string
	desc string
}

// seedProducts 建立 8 個商品，回傳清單（索引 0~7 對應 P0001~P0008）。
func seedProducts(ctx context.Context, client *ent.Client, cid, did, wh1, wh2, actor int) ([]fakeProduct, error) {
	specs := []struct{ code, name, desc string }{
		{"P0001", "里肌肉", "CAS 台灣豬"},
		{"P0002", "五花肉", "帶皮五花"},
		{"P0003", "排骨", "腩排"},
		{"P0004", "絞肉", "粗絞"},
		{"P0005", "雞胸肉", "去骨清胸"},
		{"P0006", "雞腿", "骨腿"},
		{"P0007", "培根", "低脂培根"},
		{"P0008", "香腸", "原味香腸"},
	}
	out := make([]fakeProduct, 0, len(specs))
	for _, s := range specs {
		p, err := client.Product.Create().
			SetCompanyID(cid).SetDepartmentID(did).
			SetCode(s.code).SetName(s.name).SetDescription(s.desc).
			SetInventoryWarehouseID(wh1).SetPickingWarehouseID(wh2).
			SetIsActive(true).SetCreatedBy(actor).SetUpdatedBy(actor).
			Save(ctx)
		if err != nil {
			return nil, err
		}
		out = append(out, fakeProduct{id: p.ID, code: s.code, name: s.name, desc: s.desc})
	}
	return out, nil
}

// fakeCustomer 是示範客戶的精簡描述（供訂單／退貨引用）。
type fakeCustomer struct {
	id   int
	code string
	name string
}

// seedCustomers 建立 5 個客戶（客戶編號沿用真實格式：公司前綴 + 6 位序號）。
func seedCustomers(ctx context.Context, client *ent.Client, cid, did, actor int) ([]fakeCustomer, error) {
	specs := []struct{ code, name, tax string }{
		{"FAKE-000001", "永和豆漿", "12345678"},
		{"FAKE-000002", "八方雲集 板橋店", "23456789"},
		{"FAKE-000003", "美而美早餐", "34567890"},
		{"FAKE-000004", "金滿堂婚宴會館", "45678901"},
		{"FAKE-000005", "好味道食堂", "56789012"},
	}
	out := make([]fakeCustomer, 0, len(specs))
	for i, s := range specs {
		// 交替給不同付款條件，讓客戶列表有變化（欄位本身可空，這裡刻意留部分空值）。
		created := client.Customer.Create().
			SetCompanyID(cid).SetDepartmentID(did).
			SetCustomerCode(s.code).SetName(s.name).
			SetPreferredDeliveryDays([]bool{true, false, false, true, false, false}).
			SetCreatedBy(actor).SetUpdatedBy(actor)
		if i%2 == 0 {
			created.SetTaxID(s.tax)
		}
		c, err := created.Save(ctx)
		if err != nil {
			return nil, err
		}
		out = append(out, fakeCustomer{id: c.ID, code: s.code, name: s.name})
	}
	return out, nil
}

// seedCustomerProducts 建立客戶專屬商品（別名），讓客戶詳情頁的「專屬商品」有內容。
func seedCustomerProducts(ctx context.Context, client *ent.Client, cid, did, actor int,
	customers []fakeCustomer, products []fakeProduct,
) error {
	if len(customers) == 0 || len(products) == 0 {
		return nil
	}
	specs := []struct {
		ci    int
		pi    int
		alias string
		qty   string
		note  string
	}{
		{0, 0, "里肌（薄）", "12", "切 2mm"},
		{0, 4, "雞胸（清）", "20", "去骨"},
		{1, 1, "五花（帶皮）", "8", ""},
		{2, 3, "絞肉（粗）", "5", "粗絞"},
	}
	for _, s := range specs {
		if s.ci >= len(customers) || s.pi >= len(products) {
			continue
		}
		if _, err := client.CustomerProduct.Create().
			SetCompanyID(cid).SetDepartmentID(did).
			SetCustomerID(customers[s.ci].id).SetProductID(products[s.pi].id).
			SetAliasName(s.alias).SetDefaultQty(s.qty).SetCutNote(s.note).
			SetCreatedBy(actor).SetUpdatedBy(actor).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

// seedOrders 建立 6 張訂單（涵蓋 pending／processing／completed／cancelled 四態）與明細。
//
// 訂單編號刻意走**真實格式**「來源碼 + 6 位補零」（§53：order_no 由 order_counters 取號，
// 來源碼取自 metadicts order_source 的 W=Web 中台／A=App）—— 先前手寫 SQL 用
// `DEMO-2026-0001` 這種自創格式會讓設計稿與實況不符。這裡也同步建立 order_counters，
// 使後續在 UI 上真正建立訂單時序號接續而不撞號。
func seedOrders(ctx context.Context, client *ent.Client, cid, did, actor int,
	customers []fakeCustomer, products []fakeProduct, routes []int,
) error {
	if len(customers) == 0 || len(products) == 0 {
		return nil
	}
	// 兩個來源軌道各發一些號，全部用 Web（W），並把 counter 推到已用序號之後。
	const source = "W"
	type orderSpec struct {
		status string
		days   int
		// routeIdx 是車次索引（-1 = 未指派）；processing 必然有車次 —— 派車（assignRoute）
		// 才是把訂單轉成 processing 的動作，所以「processing 但沒有車次」在資料上不成立，
		// 而看板的未指派欄只收 pending、車次欄只收有 routeId 的卡片 → 那種列會兩邊都不出現。
		routeIdx int
		custIdx  int
		note     string
		items    []struct {
			prodIdx int
			qty     string
		}
	}
	// 日期刻意讓**今天**就有待派與已派車的卡片：看板預設顯示今天，若示範單全排在未來，
	// 打開派車頁只會看到四個空欄（實測如此）。
	specs := []orderSpec{
		{"pending", 0, -1, 0, "週一早班", []struct {
			prodIdx int
			qty     string
		}{{0, "12"}, {1, "3"}}},
		{"pending", 0, 0, 1, "", []struct {
			prodIdx int
			qty     string
		}{{4, "20"}, {5, "10"}}},
		{"processing", 0, 1, 2, "", []struct {
			prodIdx int
			qty     string
		}{{2, "6"}, {3, "4"}}},
		{"pending", 2, -1, 3, "", []struct {
			prodIdx int
			qty     string
		}{{6, "4"}}},
		{"cancelled", 5, -1, 4, "客戶取消", []struct {
			prodIdx int
			qty     string
		}{{7, "6"}}},
		{"completed", -2, 2, 0, "", []struct {
			prodIdx int
			qty     string
		}{{0, "8"}, {4, "5"}, {6, "2"}}},
	}

	seq := 1
	// 配送順位是**每車次自成一條**（後端 assignRoute 也是算該欄最大順位 +1），所以用
	// per-route 計數器，而不是拿全域單號序號當順位（單號序號跨車次，語意不同）。
	routeSeq := map[int]int{}
	for _, s := range specs {
		if s.custIdx >= len(customers) {
			continue
		}
		orderNo := fmt.Sprintf("%s%06d", source, seq)
		seq++

		created := client.SalesOrder.Create().
			SetCompanyID(cid).SetDepartmentID(did).
			SetOrderNo(orderNo).
			SetCustomerID(customers[s.custIdx].id).
			SetSource(source).
			SetStatus(s.status).
			SetExpectedDeliveryDate(time.Now().AddDate(0, 0, s.days)).
			SetSalesRepID(actor).
			SetNote(s.note).
			SetVersion(0).
			SetCreatedBy(actor).SetUpdatedBy(actor)
		if s.routeIdx >= 0 && s.routeIdx < len(routes) {
			routeSeq[s.routeIdx]++
			created.SetRouteID(routes[s.routeIdx]).SetDeliverySequence(routeSeq[s.routeIdx])
		}
		o, err := created.Save(ctx)
		if err != nil {
			return err
		}

		for i, it := range s.items {
			if it.prodIdx >= len(products) {
				continue
			}
			p := products[it.prodIdx]
			if _, err := client.SalesOrderItem.Create().
				SetSalesOrderID(o.ID).SetCompanyID(cid).SetDepartmentID(did).
				SetProductID(p.id).SetDisplayName(p.name).
				SetQty(it.qty).SetUnit("公斤").SetBaseQty(it.qty).
				SetSortOrder(i).
				Save(ctx); err != nil {
				return err
			}
		}

		// 事件流（append-only）：建立事件；已完成的再補一筆狀態事件，讓訂單詳情頁的時間軸有內容。
		if _, err := client.SalesOrderEvent.Create().
			SetSalesOrderID(o.ID).SetCompanyID(cid).
			SetEventType("created").SetActorID(actor).
			SetPayload(map[string]any{"order_no": orderNo, "items": len(s.items)}).
			Save(ctx); err != nil {
			return err
		}
		switch s.status {
		case "processing":
			if _, err := client.SalesOrderEvent.Create().
				SetSalesOrderID(o.ID).SetCompanyID(cid).
				SetEventType("processing").SetActorID(actor).
				SetPayload(map[string]any{"from": "pending", "to": "processing"}).
				Save(ctx); err != nil {
				return err
			}
		case "completed":
			for _, ev := range []struct{ from, to string }{{"pending", "processing"}, {"processing", "completed"}} {
				if _, err := client.SalesOrderEvent.Create().
					SetSalesOrderID(o.ID).SetCompanyID(cid).
					SetEventType(ev.to).SetActorID(actor).
					SetPayload(map[string]any{"from": ev.from, "to": ev.to}).
					Save(ctx); err != nil {
					return err
				}
			}
		case "cancelled":
			if _, err := client.SalesOrderEvent.Create().
				SetSalesOrderID(o.ID).SetCompanyID(cid).
				SetEventType("cancelled").SetActorID(actor).
				SetReason("客戶取消").
				SetPayload(map[string]any{"from": "pending", "to": "cancelled"}).
				Save(ctx); err != nil {
				return err
			}
		}
	}

	// counter 對齊已用序號：next_seq 指向**下一個要發**的號，故為 seq。
	// 沒有這一列，UI 上新建訂單會從 W000001 開始而撞既有單號。
	if _, err := client.OrderCounter.Create().
		SetCompanyID(cid).SetSource(source).
		SetNextSeq(seq).SetVersion(0).
		Save(ctx); err != nil {
		return err
	}
	return nil
}

// seedReturns 建立 3 筆退貨申請（pending／approved／rejected 各一）與其明細。
func seedReturns(ctx context.Context, client *ent.Client, cid, did, actor int,
	customers []fakeCustomer, products []fakeProduct,
) error {
	if len(customers) == 0 || len(products) == 0 {
		return nil
	}
	specs := []struct {
		status   string
		remark   string
		reject   string
		reviewed bool
	}{
		{"pending", "整箱破損，附照片", "", false},
		{"approved", "數量短少", "", true},
		{"rejected", "客戶改單", "規格與實際不符", true},
	}
	for i, s := range specs {
		if i >= len(customers) {
			break
		}
		created := client.ReturnRequest.Create().
			SetCompanyID(cid).SetDepartmentID(did).
			SetCustomerID(customers[i].id).
			SetCreatedByUserID(actor).
			SetStatus(s.status).
			SetRemark(s.remark).
			SetVersion(0)
		if s.reviewed {
			created.SetReviewedByUserID(actor).SetReviewedAt(time.Now().Add(-time.Hour))
		}
		if s.reject != "" {
			created.SetRejectReason(s.reject)
		}
		rr, err := created.Save(ctx)
		if err != nil {
			return err
		}

		items := []struct {
			pi     int
			qty    string
			spec   string
			reason string
		}{
			{0, "2", "薄切 2mm", "外包裝破損"},
			{1, "1", "厚切 8mm", "到貨短少"},
		}
		for _, it := range items {
			if it.pi >= len(products) {
				continue
			}
			p := products[it.pi]
			if _, err := client.ReturnRequestItem.Create().
				SetReturnRequestID(rr.ID).SetCompanyID(cid).SetDepartmentID(did).
				SetSourceType("order").
				SetProductID(p.id).SetProductName(p.name).
				SetSpec(it.spec).SetUnit("公斤").SetQuantity(it.qty).
				SetReason(it.reason).
				SetPhotoFileIds([]string{}).
				Save(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// seedNotifications 建立 3 個模板與 5 則通知（含未讀／已讀／失敗）。
func seedNotifications(ctx context.Context, client *ent.Client, cid, did, actor int) error {
	tplSpecs := []struct {
		code, name, channel, subject, body string
	}{
		{"ORDER_CREATED", "新訂單通知", "fcm", "新訂單", "您的訂單 {{order_no}} 已建立"},
		{"ORDER_DISPATCHED", "派車通知", "fcm", "已派車", "訂單 {{order_no}} 已排入 {{route}}"},
		{"RETURN_REVIEWED", "退貨審核結果", "fcm", "退貨審核", "您的退貨申請已{{decision}}"},
	}
	tplIDs := make(map[string]int, len(tplSpecs))
	for _, t := range tplSpecs {
		tpl, err := client.NotificationTemplate.Create().
			SetCompanyID(cid).SetDepartmentID(did).
			SetCode(t.code).SetName(t.name).SetChannel(t.channel).
			SetSubject(t.subject).SetBody(t.body).SetLocale("zh-TW").
			SetIsActive(true).
			Save(ctx)
		if err != nil {
			return err
		}
		tplIDs[t.code] = tpl.ID
	}

	now := time.Now()
	notifSpecs := []struct {
		code    string
		title   string
		content string
		status  string
		failure string
		ago     time.Duration
		read    bool
	}{
		{"ORDER_CREATED", "新訂單已建立", "訂單 W000001 已建立，等待派車", "sent", "", 12 * time.Minute, false},
		{"ORDER_DISPATCHED", "已排入車次", "訂單 W000002 已排入 1車（R01）", "sent", "", 95 * time.Minute, false},
		{"RETURN_REVIEWED", "退貨申請待審", "八方雲集 板橋店 提出退貨申請", "sent", "", 4 * time.Hour, true},
		{"ORDER_DISPATCHED", "派車完成", "3車 已完成交付，共 5 站", "sent", "", 25 * time.Hour, true},
		{"ORDER_CREATED", "推播失敗", "訂單 W000004 推播失敗", "failed", "裝置未註冊", 48 * time.Hour, false},
	}
	for _, n := range notifSpecs {
		created := client.Notification.Create().
			SetCompanyID(cid).SetDepartmentID(did).
			SetUserID(actor).
			SetChannel("fcm").
			SetTitle(n.title).SetContent(n.content).
			SetPayload(map[string]any{}).
			SetStatus(n.status).
			SetCreatedAt(now.Add(-n.ago))
		if id, ok := tplIDs[n.code]; ok {
			created.SetTemplateID(id)
		}
		if n.failure != "" {
			created.SetFailureReason(n.failure)
		} else {
			created.SetSentAt(now.Add(-n.ago))
		}
		if n.read {
			created.SetReadAt(now.Add(-n.ago + 2*time.Minute))
		}
		if _, err := created.Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

// seedPrintLogs 建立列印紀錄（4 種單據類型 + 1 筆補印）與其 PDF 檔案的 file_assets 列。
//
// file_assets 是 NOT NULL FK，故必須先建檔列。**同時把檔案寫到磁碟**：下載端點是
// 「查 DB 列 → 開 storage_path 的檔」，只建列會讓列印頁每一條「下載」都 404
// （先前版本只建列,實測 6 筆全部 404）。PDF 內容用最小合法文件即可 —— 示範資料要的是
// 「點得開」，不是可讀的報表。
func seedPrintLogs(ctx context.Context, client *ent.Client, cid, did, actor int, routes []int, storageRoot string) error {
	if len(routes) == 0 {
		return nil
	}
	specs := []struct {
		docType string
		routeIx int
		days    int
		reprint bool
		reason  string
		ago     time.Duration
	}{
		{"dispatch_summary", 0, 0, false, "", 2 * time.Hour},
		{"picking_list", 0, 0, false, "", 3 * time.Hour},
		{"delivery_note", 0, 0, false, "", 4 * time.Hour},
		{"processing_list", 0, 0, false, "", 5 * time.Hour},
		{"picking_list", 1, 1, false, "", 27 * time.Hour},
		{"dispatch_summary", 1, 1, true, "客戶要求重新列印", 50 * time.Hour},
	}
	pdf := minimalPDF()
	for i, s := range specs {
		// 檔名與路徑比照 FileStore 的慣例（`<company>/<yyyy>/<mm>`），但用可讀的固定名，
		// 讓示範資料在磁碟上一眼可辨。
		rel := fmt.Sprintf("%d/%s/%s", cid, time.Now().Format("2006/01"), fmt.Sprintf("print-%d.pdf", i+1))
		abs := filepath.Join(storageRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return fmt.Errorf("建立示範列印檔目錄: %w", err)
		}
		if err := os.WriteFile(abs, pdf, 0o644); err != nil {
			return fmt.Errorf("寫入示範列印檔: %w", err)
		}
		fa, err := client.FileAsset.Create().
			SetCompanyID(cid).SetDepartmentID(did).
			SetOwnerType("print").SetOwnerID(i + 1).
			SetFilename(filepath.Base(rel)).SetOriginalFilename(fmt.Sprintf("示範單據-%d.pdf", i+1)).
			SetMimeType("application/pdf").SetSizeBytes(len(pdf)).
			SetStoragePath(rel).
			SetURL("/api/v1/files/" + filepath.Base(rel) + "/download").
			SetCreatedBy(actor).
			Save(ctx)
		if err != nil {
			return err
		}
		if s.routeIx >= len(routes) {
			continue
		}
		created := client.PrintLog.Create().
			SetCompanyID(cid).SetDepartmentID(did).
			SetDocumentType(s.docType).
			SetRouteID(routes[s.routeIx]).
			SetTargetDate(time.Now().AddDate(0, 0, -s.days)).
			SetIsReprint(s.reprint).
			SetPrintedBy(actor).
			SetPrintedAt(time.Now().Add(-s.ago)).
			SetFileAssetID(fa.ID)
		if s.reason != "" {
			created.SetReprintReason(s.reason)
		}
		if _, err := created.Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

// minimalPDF 回一份最小可開啟的 PDF（單頁、標題「示範單據」）。
//
// 不用套件：示範資料只需要「下載得到一個 Content-Type: application/pdf 的檔案」，
// 為此引入 PDF 產生器不划算。xref 位移以實際長度計算，確保檔案結構合法。
func minimalPDF() []byte {
	content := "BT /F1 24 Tf 72 720 Td (Demo Document) Tj ET"
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	}
	var b bytes.Buffer
	b.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects)+1)
	for i, o := range objects {
		offsets[i+1] = b.Len()
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, o)
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for i := 1; i <= len(objects); i++ {
		fmt.Fprintf(&b, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	return b.Bytes()
}

// seedAuditLogs 建立稽核日誌（動作值域受 audit_logs_action_check 約束：create/update/delete/
// login/logout/print/force_logout/role_change/dispatch_cancel/void —— 自創值會被 CHECK 擋下）。
func seedAuditLogs(ctx context.Context, client *ent.Client, cid, did, actor int) error {
	specs := []struct {
		resource string
		action   auditlog.Action
		resID    string
		ago      time.Duration
	}{
		{"sales_order", auditlog.ActionCreate, "1", 30 * time.Minute},
		{"sales_order", auditlog.ActionUpdate, "2", 90 * time.Minute},
		{"customer", auditlog.ActionCreate, "1", 200 * time.Minute},
		{"product", auditlog.ActionCreate, "1", 260 * time.Minute},
		{"user", auditlog.ActionUpdate, "1", 400 * time.Minute},
		{"sales_order", auditlog.ActionDispatchCancel, "5", 600 * time.Minute},
		{"sales_order", auditlog.ActionPrint, "3", 700 * time.Minute},
		{"user", auditlog.ActionLogin, "1", 900 * time.Minute},
	}
	for _, s := range specs {
		// trace_id 是跨請求關聯鍵（§44），以底線前綴避開業務鍵；示範值固定即可。
		snapshot := map[string]any{"_trace_id": "01a0c9fd-9cc6-7e7b-ad55-5418b0bb345a"}
		if _, err := client.AuditLog.Create().
			SetCompanyID(cid).SetDepartmentID(did).
			SetUserID(actor).
			SetResourceType(s.resource).
			SetAction(s.action).
			SetResourceID(s.resID).
			SetAfterSnapshot(snapshot).
			SetCreatedAt(time.Now().Add(-s.ago)).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

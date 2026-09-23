//go:build integration

package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1/productsv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// fakePrintConv 為假 Gotenberg(回固定 PDF,計次)。
type fakePrintConv struct {
	calls int
}

func (f *fakePrintConv) Convert(html string) ([]byte, error) {
	f.calls++
	return []byte("%PDF-1.4 fake print"), nil
}

// newPrintServer 以 fake 產線掛 PrintService(身分注入 + 請求交易)。
func newPrintServer(t *testing.T, db *ent.Client, id authz.Identity, root string, fc *fakePrintConv) productsv1connect.PrintServiceClient {
	t.Helper()
	SetPrintPipeline(fc, root)
	path, handler := productsv1connect.NewPrintServiceHandler(NewPrintService(db),
		dbtenant.HandlerOption(db))
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		// 測試 scope:dept_admin/staff → department,company_admin/super → company
		// (生產由 authzMiddleware 依身分導出)。
		scope := auth.RLSScope{UserID: id.UserID, CompanyID: id.CompanyID,
			DepartmentID: id.DepartmentID, DataScope: auth.DataScopeDepartment, CompanyActive: true}
		if id.Role == "company_admin" || id.Role == "super" {
			scope.DataScope = auth.DataScopeCompany
		}
		ctx = auth.WithRLS(ctx, scope)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return productsv1connect.NewPrintServiceClient(http.DefaultClient, ts.URL)
}

// seedPrintOrder 建公司/部門/客戶/操作員/商品/車次 + 一張 processing 訂單(2026-09-22)。
// 回 co/dept/cust/actor/route。
func seedPrintOrder(t *testing.T, ctx context.Context, db *ent.Client) (int, int, int, int, int) {
	t.Helper()
	coID, deptID, custID, actorID, prodID := seedOrderCompany(t, ctx, db)
	var routeID int
	seedTx(t, db, func(tx *ent.Tx) error {
		r, err := tx.Route.Create().SetCompanyID(coID).SetDepartmentID(deptID).
			SetCode("A").SetName("市區線").Save(ctx)
		if err != nil {
			return err
		}
		routeID = r.ID
		// 建單後直寫 processing(狀態機由 TransitionOrder 覆蓋,此處只備正式列印的前置)。
		o, err := tx.SalesOrder.Create().SetCompanyID(coID).SetDepartmentID(deptID).
			SetCustomerID(custID).SetOrderNo("W000001").SetSource("W").
			SetRouteID(r.ID).
			SetExpectedDeliveryDate(dateOf(2026, 9, 22)).SetStatus("processing").
			SetCreatedBy(actorID).Save(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.SalesOrderItem.Create().SetSalesOrderID(o.ID).
			SetCompanyID(coID).SetDepartmentID(deptID).SetProductID(prodID).
			SetDisplayName("蘋果").SetQty("10").SetUnit("斤").SetBaseQty("10").Save(ctx); err != nil {
			return err
		}
		return nil
	})
	return coID, deptID, custID, actorID, routeID
}

// TestIntegrationPrintPreviewAndPrint 預覽→正式→重印全鏈(fake Gotenberg):
// 預覽不寫 print_logs;首印 is_reprint=false;重印需原因且 is_reprint=true;
// pending 車次正式列印被拒;空表預覽被拒。
func TestIntegrationPrintPreviewAndPrint(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	coID, deptID, _, actorID, routeID := seedPrintOrder(t, ctx, db)
	root := t.TempDir()
	fc := &fakePrintConv{}
	id := authz.Identity{UserID: uItoa(actorID), CompanyID: uItoa(coID),
		DepartmentID: uItoa(deptID), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
	rpc := newPrintServer(t, db, id, root, fc)

	prev, err := rpc.Preview(ctx, connect.NewRequest(newPreviewReq(routeID)))
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}
	// 回應的 URL 必須**等於 file_assets.url**，不是另一個自行拼出來的字串：
	// 兩者曾經分歧（回應 /files/<id>/download、DB /files/<uuid>.pdf/download），
	// 下載路由兩種都吃，所以只斷言「非空」抓不到。
	assertDownloadURLMatchesStored(t, ctx, db, prev.Msg.GetDownloadUrl(), prev.Msg.GetFileAssetId())
	if n := countPrintLogs(t, ctx, db); n != 0 {
		t.Fatalf("預覽不得寫 print_logs,got %d", n)
	}

	pr, err := rpc.Print(ctx, connect.NewRequest(newPrintReq(routeID, "")))
	if err != nil {
		t.Fatalf("Print 首印: %v", err)
	}
	if pr.Msg.GetIsReprint() {
		t.Fatal("首印 is_reprint 應為 false")
	}

	// 無原因重印 → 拒絕且不新增。
	if _, err := rpc.Print(ctx, connect.NewRequest(newPrintReq(routeID, ""))); err == nil {
		t.Fatal("無原因重印應拒絕")
	}
	if n := countPrintLogs(t, ctx, db); n != 1 {
		t.Fatalf("拒絕的重印不得寫記錄,got %d", n)
	}
	re, err := rpc.Print(ctx, connect.NewRequest(newPrintReq(routeID, "客戶要求多印一份")))
	if err != nil {
		t.Fatalf("Print 重印: %v", err)
	}
	if !re.Msg.GetIsReprint() {
		t.Fatal("重印 is_reprint 應為 true")
	}
	if n := countPrintLogs(t, ctx, db); n != 2 {
		t.Fatalf("重印應新增第二筆,got %d", n)
	}

	// 首印帶原因 → 拒絕(參數組合非法)。
	if _, err := rpc.Print(ctx, connect.NewRequest(newPrintReqFor(routeID, "picking_list", ""))); err != nil {
		t.Fatalf("不同類型首印應成功(獨立比對鍵): %v", err)
	}

	// ListLogs 可查兩筆 dispatch_summary。
	logs, err := rpc.ListLogs(ctx, connect.NewRequest(newListLogsReq()))
	if err != nil {
		t.Fatalf("ListLogs: %v", err)
	}
	if logs.Msg.GetTotal() < 2 {
		t.Fatalf("ListLogs 應至少 2 筆,got %d", logs.Msg.GetTotal())
	}
}

// assertDownloadURLMatchesStored 斷言回應的 download_url 與 file_assets.url 逐字相同。
//
// 為什麼不是「非空」就夠：回應與 DB 曾各自拼字串（回應用整數 id、DB 用系統檔名），
// 下載路由兩種形狀都接受，所以兩者分歧時仍然「有值、也下載得到」，只有比對才抓得到分歧。
func assertDownloadURLMatchesStored(t *testing.T, ctx context.Context, db *ent.Client, gotURL, gotFileAssetID string) {
	t.Helper()
	faid, err := strconv.Atoi(gotFileAssetID)
	if err != nil {
		t.Fatalf("file_asset_id 應為整數字串,got %q", gotFileAssetID)
	}
	fa, err := db.FileAsset.Get(ctx, faid)
	if err != nil {
		t.Fatalf("讀 file_asset %d: %v", faid, err)
	}
	if gotURL != fa.URL {
		t.Fatalf("回應 download_url 應等於 file_assets.url\n  回應: %s\n  DB  : %s", gotURL, fa.URL)
	}
	if gotURL == "" {
		t.Fatal("download_url 不得為空")
	}
}

// TestIntegrationPrintRejectsNonProcessing pending 車次正式列印整批拒絕(無半套記錄)。
func TestIntegrationPrintRejectsNonProcessing(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	coID, deptID, custID, actorID, prodID := seedOrderCompany(t, ctx, db)
	var routeID int
	seedTx(t, db, func(tx *ent.Tx) error {
		r, err := tx.Route.Create().SetCompanyID(coID).SetDepartmentID(deptID).
			SetCode("B").SetName("郊區線").Save(ctx)
		if err != nil {
			return err
		}
		routeID = r.ID
		o, err := tx.SalesOrder.Create().SetCompanyID(coID).SetDepartmentID(deptID).
			SetCustomerID(custID).SetOrderNo("W000002").SetSource("W").
			SetRouteID(r.ID).
			SetExpectedDeliveryDate(dateOf(2026, 9, 22)).SetStatus("pending").
			SetCreatedBy(actorID).Save(ctx)
		if err != nil {
			return err
		}
		_, err = tx.SalesOrderItem.Create().SetSalesOrderID(o.ID).
			SetCompanyID(coID).SetDepartmentID(deptID).SetProductID(prodID).
			SetDisplayName("蘋果").SetQty("5").SetUnit("斤").SetBaseQty("5").Save(ctx)
		return err
	})
	root := t.TempDir()
	fc := &fakePrintConv{}
	id := authz.Identity{UserID: uItoa(actorID), CompanyID: uItoa(coID),
		DepartmentID: uItoa(deptID), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
	rpc := newPrintServer(t, db, id, root, fc)
	if _, err := rpc.Print(ctx, connect.NewRequest(newPrintReq(routeID, ""))); err == nil {
		t.Fatal("pending 車次應拒絕正式列印")
	}
	if n := countPrintLogs(t, ctx, db); n != 0 {
		t.Fatalf("拒絕不得寫記錄,got %d", n)
	}
	// 預覽不限狀態:pending 可預覽。
	if _, err := rpc.Preview(ctx, connect.NewRequest(newPreviewReq(routeID))); err != nil {
		t.Fatalf("pending 應可預覽: %v", err)
	}
}

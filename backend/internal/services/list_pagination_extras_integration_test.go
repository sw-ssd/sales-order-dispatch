//go:build integration

package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/notification"
	"github.com/salesorder/sales-order-1.0/backend/ent/printlog"
	"github.com/salesorder/sales-order-1.0/backend/ent/returnrequest"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	productsv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1/productsv1connect"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// 本檔補上 list_pagination_integration_test.go 未涵蓋的三條清單端點，它們原本只以
// `created_at`／`printed_at` 降冪排序（非唯一鍵）而缺 `ent.Desc(<entity>.FieldID)` 次序鍵
// （2026-09-26 修）。
//
// 為何這三條特別容易踩：排序鍵是**交易時間**，同一交易內寫入的多筆列時間戳完全相同 ——
// 通知的 fan-out（queueNotifications 逐筆 Save）、同一車次一次列印產生的多筆紀錄、
// 以及退貨的批次建立都屬此類。同值群一旦跨頁邊界，PG 會讓某些列在兩頁重複出現、另一些
// 完全不出現（實測：500 列同時間戳、每頁 20 筆逐頁掃描只看到 191 個相異列）。
//
// 為何必須是 integration + 真 PG：sqlite 的 sorter 對同值群給穩定次序、LIMIT/OFFSET 只是
// 切片，同構探針全綠（AGENTS §4 已載明）。
func TestIntegrationListPaginationTieBreak(t *testing.T) {
	testsupport.RequiresContainer(t)
	// 與 list_pagination_integration_test.go 同構：由 ent 建 schema（本測試的重點是 ORDER BY
	// 的 tie 行為，與 RLS 無關，故不需要跑 goose 遷移，也不需要 app_rw 角色）。
	// 不可以先跑 migrateBusinessUp 再 Schema.Create —— goose 建的表與 ent 宣告不一致，
	// Schema.Create 會以「unexpected attribute change」失敗。
	adminDSN := testsupport.Postgres(t)
	db := openPGEntClient(t, adminDSN)
	ctx := t.Context()

	co := seedListScanCompanies(t, ctx, db)
	depts := seedListScanDepartments(t, ctx, db, co)
	seedListScanCustomers(t, ctx, db, co.ID)

	// 三張表都以「同一時間戳」造同值群：這是缺陷唯一的觸發條件。
	const rows = listScanRows
	seedTieBreakFixtures(t, ctx, db, co.ID, depts[0].ID, rows)

	super := authz.Identity{
		UserID: "1", CompanyID: strconv.Itoa(co.ID), Role: "super", Roles: []string{"super"},
	}
	mux := http.NewServeMux()
	RegisterNotificationService(mux, db)
	RegisterPrintService(mux, db)
	RegisterReturnService(mux, db)
	scope := auth.RLSScope{
		UserID: "1", CompanyID: strconv.Itoa(co.ID),
		DataScope: auth.DataScopeAll, CompanyActive: true,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := authz.WithIdentity(r.Context(), super)
		c = auth.WithRLS(c, scope)
		c = authz.WithDB(c, db)
		mux.ServeHTTP(w, r.WithContext(c))
	}))
	t.Cleanup(ts.Close)

	notifications := salesorderv1connect.NewNotificationServiceClient(http.DefaultClient, ts.URL)
	prints := productsv1connect.NewPrintServiceClient(http.DefaultClient, ts.URL)
	returns := salesorderv1connect.NewReturnServiceClient(http.DefaultClient, ts.URL)

	notificationScan := newListScanner("ListNotifications",
		func(_ string, _ bool, page, pageSize int32) ([]*salesorderv1.NotificationView, error) {
			res, err := notifications.ListNotifications(ctx, connect.NewRequest(&salesorderv1.ListNotificationsRequest{
				Page: page, PageSize: pageSize,
			}))
			if err != nil {
				return nil, err
			}
			return res.Msg.GetNotifications(), nil
		}, func(n *salesorderv1.NotificationView) string { return n.GetId() })
	printScan := newListScanner("ListLogs",
		func(_ string, _ bool, page, pageSize int32) ([]*productsv1.LogEntry, error) {
			res, err := prints.ListLogs(ctx, connect.NewRequest(&productsv1.ListLogsRequest{
				Page: page, PageSize: pageSize,
			}))
			if err != nil {
				return nil, err
			}
			return res.Msg.GetEntries(), nil
		}, func(p *productsv1.LogEntry) string { return p.GetId() })
	returnScan := newListScanner("ListReturnRequests",
		func(_ string, _ bool, page, pageSize int32) ([]*salesorderv1.ReturnRequestEntry, error) {
			res, err := returns.ListReturnRequests(ctx, connect.NewRequest(&salesorderv1.ListReturnRequestsRequest{
				Page: page, PageSize: pageSize,
			}))
			if err != nil {
				return nil, err
			}
			return res.Msg.GetEntries(), nil
		}, func(r *salesorderv1.ReturnRequestEntry) string { return r.GetId() })

	// 全量基準刻意帶上服務端的預期方向（與 list_pagination_integration_test.go 的 G6 同理）：
	// 集合比對對方向不敏感，故另以 assertSequenceMatchesFull 逐位比對。
	for _, tc := range []struct {
		entity string
		full   []string
		scan   listScanner
	}{
		{"通知", allIDs(t, func() ([]int, error) {
			return db.Notification.Query().Where(notification.CompanyIDEQ(co.ID)).
				Order(ent.Desc(notification.FieldCreatedAt), ent.Desc(notification.FieldID)).
				Select(notification.FieldID).Ints(ctx)
		}), notificationScan},
		{"列印紀錄", allIDs(t, func() ([]int, error) {
			return db.PrintLog.Query().Where(printlog.CompanyIDEQ(co.ID)).
				Order(ent.Desc(printlog.FieldPrintedAt), ent.Desc(printlog.FieldID)).
				Select(printlog.FieldID).Ints(ctx)
		}), printScan},
		{"退貨申請", allIDs(t, func() ([]int, error) {
			return db.ReturnRequest.Query().Where(returnrequest.CompanyIDEQ(co.ID)).
				Order(ent.Desc(returnrequest.FieldCreatedAt), ent.Desc(returnrequest.FieldID)).
				Select(returnrequest.FieldID).Ints(ctx)
		}), returnScan},
	} {
		t.Run(tc.entity, func(t *testing.T) {
			// 前提檢查：fixture 必須真的造出遠大於一頁的同值群，否則本測試沒有偵測力。
			if len(tc.full) < listScanPageSize*2 {
				t.Fatalf("fixture 只有 %d 列，不足以讓同值群跨頁邊界（需 >= %d）",
					len(tc.full), listScanPageSize*2)
			}
			paged := scanPages(t, tc.scan, "", true)
			assertScanMatchesFull(t, tc.entity, "", true, paged, tc.full)
			// 集合比對對次序鍵方向不敏感，故另逐位比對序列。
			assertSequenceMatchesFull(t, tc.entity, "", true, paged, tc.full)
		})
	}
}

// seedTieBreakFixtures 為三張表各造 n 列，**全部共用同一個時間戳**。
// 時間戳完全相同正是缺陷的觸發條件（非唯一排序鍵 + 同值群跨頁）。
//
// 一律走 ent（與 list_pagination_integration_test.go 的 seed* 同構）：手寫 INSERT 要逐一
// 猜 NOT NULL 與 FK，而 ent 的 schema 就是那份清單；欄位語意也由 ent 的 Set* 表達。
// 寫入走 seedTx 的系統範圍交易，與既有 fixture 一致。
func seedTieBreakFixtures(t *testing.T, ctx context.Context, db *ent.Client, companyID, deptID, n int) {
	t.Helper()
	// 單一固定時間戳（UTC 午夜，避免任何時區邊界解讀）。所有列共用它。
	sameTS := time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC)
	// 依賴列：三者都有 FK 與 NOT NULL 的 department_id／user_id，故先建好真實的使用者、
	// 客戶與車次（ent 的 Set* 會帶上正確的欄位，不必逐一猜 NOT NULL 清單）。
	var userID, custID, routeID, faID int
	seedTx(t, db, func(tx *ent.Tx) error {
		u, err := tx.User.Create().
			SetEmail(fmt.Sprintf("pagination-%d@example.test", companyID)).
			SetName("分頁使用者").SetRole("staff").SetStatus("active").
			SetPasswordHash("x").SetCompanyID(companyID).SetDepartmentID(deptID).Save(ctx)
		if err != nil {
			return err
		}
		userID = u.ID
		c, err := tx.Customer.Create().
			SetCompanyID(companyID).SetDepartmentID(deptID).
			SetCustomerCode(fmt.Sprintf("PG-%d", companyID)).SetName("分頁客戶").Save(ctx)
		if err != nil {
			return err
		}
		custID = c.ID
		r, err := tx.Route.Create().
			SetCompanyID(companyID).SetDepartmentID(deptID).
			SetCode(fmt.Sprintf("RT-%d", companyID)).SetName("分頁車次").Save(ctx)
		if err != nil {
			return err
		}
		routeID = r.ID
		f, err := tx.FileAsset.Create().
			SetCompanyID(companyID).SetOwnerType("print_log").SetOwnerID(0).
			SetFilename("pg.pdf").SetOriginalFilename("pg.pdf").
			SetMimeType("application/pdf").SetSizeBytes(1).
			SetStoragePath("pg/pdf.pdf").SetURL("/files/pg.pdf").Save(ctx)
		if err != nil {
			return err
		}
		faID = f.ID
		return nil
	})
	seedTx(t, db, func(tx *ent.Tx) error {
		nb := make([]*ent.NotificationCreate, 0, n)
		pb := make([]*ent.PrintLogCreate, 0, n)
		rb := make([]*ent.ReturnRequestCreate, 0, n)
		for i := range n {
			nb = append(nb, tx.Notification.Create().
				SetCompanyID(companyID).SetDepartmentID(deptID).SetUserID(userID).
				SetChannel("in_app").SetTitle(fmt.Sprintf("通知 %03d", i)).SetContent("內容").
				SetStatus("pending").SetCreatedAt(sameTS))
			pb = append(pb, tx.PrintLog.Create().
				SetCompanyID(companyID).SetDepartmentID(deptID).
				SetDocumentType("dispatch_summary").SetRouteID(routeID).
				SetTargetDate(sameTS).SetIsReprint(false).
				SetPrintedBy(userID).SetPrintedAt(sameTS).SetFileAssetID(faID))
			rb = append(rb, tx.ReturnRequest.Create().
				SetCompanyID(companyID).SetDepartmentID(deptID).SetCustomerID(custID).
				SetCreatedByUserID(userID).SetStatus("pending").SetVersion(0).
				SetCreatedAt(sameTS))
		}
		if _, err := tx.Notification.CreateBulk(nb...).Save(ctx); err != nil {
			return err
		}
		if _, err := tx.PrintLog.CreateBulk(pb...).Save(ctx); err != nil {
			return err
		}
		_, err := tx.ReturnRequest.CreateBulk(rb...).Save(ctx)
		return err
	})
}

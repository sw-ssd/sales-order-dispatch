//go:build integration

package services

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	productsv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1"
)

// dateOf 組日期(列印 target_date 用)。
func dateOf(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// newPreviewReq 組預覽請求(單車總表)。
func newPreviewReq(routeID int) *productsv1.PreviewRequest {
	return &productsv1.PreviewRequest{
		DocumentType: "dispatch_summary", RouteId: strconv.Itoa(routeID),
		TargetDate: "2026-09-22",
	}
}

// newPrintReq 組正式列印請求(單車總表 + 原因)。
func newPrintReq(routeID int, reason string) *productsv1.PrintRequest {
	return newPrintReqFor(routeID, "dispatch_summary", reason)
}

// newPrintReqFor 組指定類型正式列印請求。
func newPrintReqFor(routeID int, docType, reason string) *productsv1.PrintRequest {
	return &productsv1.PrintRequest{
		DocumentType: docType, RouteId: strconv.Itoa(routeID),
		TargetDate: "2026-09-22", ReprintReason: reason,
	}
}

// newListLogsReq 組記錄查詢(全取)。
func newListLogsReq() *productsv1.ListLogsRequest {
	return &productsv1.ListLogsRequest{Page: 1, PageSize: 20}
}

// countPrintLogs 計 print_logs 筆數。
func countPrintLogs(t *testing.T, ctx context.Context, db *ent.Client) int {
	t.Helper()
	c, err := db.PrintLog.Query().Count(ctx)
	if err != nil {
		t.Fatalf("計 print_logs: %v", err)
	}
	return c
}

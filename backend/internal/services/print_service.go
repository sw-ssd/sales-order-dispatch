// PrintService 列印與預覽(09 計畫 Task 5.5.2–5.5.4):Preview / Print / ListLogs。
// Preview 不限狀態、不寫 print_logs;Print 要求範圍內全 processing,首印/重印推導;
// 重印沿用 Print(有既有記錄即重印分支,必填原因);PDF 經 file_assets 下載 URL 交付。
package services

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/printlog"
	"github.com/salesorder/sales-order-1.0/backend/ent/printpreview"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
	"github.com/salesorder/sales-order-1.0/backend/internal/print"
	productsv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1/productsv1connect"
)

// PrintService 為列印服務(Gotenberg client + 儲存根由建構子注入;測試注入 fake)。
type PrintService struct {
	db   *ent.Client
	conv print.Converter
	root string
}

// printConv 為預設轉換器(延遲綁定:Gotenberg URL 由 server 組裝時設定)。
var printConv print.Converter
var printRoot string

// SetPrintPipeline 設定產線依賴(僅 server 組裝鏈呼叫)。
func SetPrintPipeline(conv print.Converter, root string) {
	printConv = conv
	printRoot = root
}

// NewPrintService 建立 PrintService。
func NewPrintService(db *ent.Client) *PrintService {
	return &PrintService{db: db, conv: printConv, root: printRoot}
}

// RegisterPrintService 掛到 /api/v1(租戶 session + RLS)。
func RegisterPrintService(mux *http.ServeMux, db *ent.Client) {
	path, handler := productsv1connect.NewPrintServiceHandler(NewPrintService(db),
		connect.WithInterceptors(requestid.Interceptor(), dbtenant.Interceptor(db)))
	mux.Handle(path, handler)
}

// parsePrintInput 解析請求(類型/車次/日期/選用選擇器)。
func parsePrintInput(docType, routeID, targetDate, customerID, warehouseID string) (print.AssembleInput, error) {
	var in print.AssembleInput
	switch print.DocumentType(docType) {
	case print.DispatchSummary, print.DeliveryNote, print.PickingList, print.ProcessingList:
		in.Type = print.DocumentType(docType)
	default:
		return in, errcode.SysInvalidArgument.Error(map[string]string{"field": "document_type"})
	}
	rid, err := parseID(routeID)
	if err != nil {
		return in, errcode.SysInvalidArgument.Error(map[string]string{"field": "route_id"})
	}
	in.RouteID = rid
	td, err := time.Parse("2006-01-02", strings.TrimSpace(targetDate))
	if err != nil {
		return in, errcode.SysInvalidArgument.Error(map[string]string{"field": "target_date"})
	}
	in.TargetDate = td
	if strings.TrimSpace(customerID) != "" {
		cid, err := parseID(customerID)
		if err != nil {
			return in, errcode.SysInvalidArgument.Error(map[string]string{"field": "customer_id"})
		}
		in.CustomerID = &cid
	}
	if strings.TrimSpace(warehouseID) != "" {
		wid, err := parseID(warehouseID)
		if err != nil {
			return in, errcode.SysInvalidArgument.Error(map[string]string{"field": "warehouse_id"})
		}
		in.WarehouseID = &wid
	}
	return in, nil
}

// Preview 預覽(任何狀態可印,寫 print_previews,不觸碰 print_logs)。
func (s *PrintService) Preview(ctx context.Context, req *connect.Request[productsv1.PreviewRequest]) (*connect.Response[productsv1.PreviewResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	in, err := parsePrintInput(req.Msg.GetDocumentType(), req.Msg.GetRouteId(),
		req.Msg.GetTargetDate(), req.Msg.GetCustomerId(), req.Msg.GetWarehouseId())
	if err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	asm, err := print.Assemble(ctx, s.db, in)
	if err != nil {
		return nil, toConnectError(err)
	}
	if asm.Empty {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"reason": "無可列印資料"})
	}
	pipe := print.NewPipeline(s.convOrErr(), s.root)
	out, err := pipe.Produce(asm.Model, in.Type, false, cid)
	if err != nil {
		return nil, toConnectError(err)
	}
	committed := false
	defer func() {
		if !committed {
			pipe.Discard(out.RelPath)
		}
	}()
	actor, _ := parseID(id.UserID)
	meta := printFileMeta(cid, did, "print_preview", 0, out, actor, in)
	faid, err := createFileAsset(ctx, db, meta)
	if err != nil {
		return nil, toConnectError(err)
	}
	b := db.PrintPreview.Create().
		SetCompanyID(cid).SetDocumentType(string(in.Type)).SetRouteID(in.RouteID).
		SetTargetDate(in.TargetDate).SetPreviewedBy(actor).SetFileAssetID(faid)
	if did != nil {
		b = b.SetDepartmentID(*did)
	}
	if in.CustomerID != nil {
		b = b.SetCustomerID(*in.CustomerID)
	}
	if in.WarehouseID != nil {
		b = b.SetWarehouseID(*in.WarehouseID)
	}
	pv, err := b.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	committed = true
	return connect.NewResponse(&productsv1.PreviewResponse{
		PreviewId: strconv.Itoa(pv.ID), FileAssetId: strconv.Itoa(faid),
		DownloadUrl: "/api/v1/files/" + strconv.Itoa(faid) + "/download",
	}), nil
}

// Print 正式列印(範圍內全 processing;首印/重印推導;同交易寫 print_logs + file_assets + audit)。
func (s *PrintService) Print(ctx context.Context, req *connect.Request[productsv1.PrintRequest]) (*connect.Response[productsv1.PrintResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	in, err := parsePrintInput(req.Msg.GetDocumentType(), req.Msg.GetRouteId(),
		req.Msg.GetTargetDate(), req.Msg.GetCustomerId(), req.Msg.GetWarehouseId())
	if err != nil {
		return nil, err
	}
	reason := strings.TrimSpace(req.Msg.GetReprintReason())
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	// 狀態檢查:範圍內目標訂單逐筆須 processing(交易內讀取,寫入同交易,無並發窗口)。
	orders, err := ordersInScope(ctx, db, cid, did, in)
	if err != nil {
		return nil, toConnectError(err)
	}
	if len(orders) == 0 {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"reason": "無可列印資料"})
	}
	for _, o := range orders {
		if o.Status != OrderStatusProcessing {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{
				"reason": "訂單 " + o.OrderNo + " 非處理中，不可列印",
			})
		}
	}
	// 首印/重印推導:比對鍵查既有 print_logs。
	isReprint, err := hasPrintLog(ctx, db, cid, did, in)
	if err != nil {
		return nil, toConnectError(err)
	}
	if isReprint && reason == "" {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "reprint_reason"})
	}
	if !isReprint && reason != "" {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "reprint_reason"})
	}
	asm, err := print.Assemble(ctx, s.db, in)
	if err != nil {
		return nil, toConnectError(err)
	}
	if asm.Empty {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"reason": "無可列印資料"})
	}
	pipe := print.NewPipeline(s.convOrErr(), s.root)
	out, err := pipe.Produce(asm.Model, in.Type, false, cid)
	if err != nil {
		return nil, toConnectError(err)
	}
	committed := false
	defer func() {
		if !committed {
			pipe.Discard(out.RelPath)
		}
	}()
	actor, _ := parseID(id.UserID)
	faid, err := createFileAsset(ctx, db, printFileMeta(cid, did, "print_log", 0, out, actor, in))
	if err != nil {
		return nil, toConnectError(err)
	}
	lb := db.PrintLog.Create().
		SetCompanyID(cid).SetDocumentType(string(in.Type)).SetRouteID(in.RouteID).
		SetTargetDate(in.TargetDate).SetIsReprint(isReprint).SetPrintedBy(actor).SetFileAssetID(faid)
	if did != nil {
		lb = lb.SetDepartmentID(*did)
	}
	if in.CustomerID != nil {
		lb = lb.SetCustomerID(*in.CustomerID)
	}
	if in.WarehouseID != nil {
		lb = lb.SetWarehouseID(*in.WarehouseID)
	}
	if reason != "" {
		lb = lb.SetReprintReason(reason)
	}
	pl, err := lb.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	after := map[string]any{"document_type": string(in.Type), "is_reprint": isReprint}
	if reason != "" {
		after["reprint_reason"] = reason
	}
	if err := recordAudit(ctx, tx, "print_log", "print", pl.ID, cid, did, actor, after); err != nil {
		return nil, toConnectError(err)
	}
	committed = true
	return connect.NewResponse(&productsv1.PrintResponse{
		PrintLogId: strconv.Itoa(pl.ID), FileAssetId: strconv.Itoa(faid),
		DownloadUrl: "/api/v1/files/" + strconv.Itoa(faid) + "/download",
		IsReprint:   isReprint,
	}), nil
}

// ListLogs 列印記錄查詢(部門範圍,per_page ≤ 100)。
func (s *PrintService) ListLogs(ctx context.Context, req *connect.Request[productsv1.ListLogsRequest]) (*connect.Response[productsv1.ListLogsResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	q := db.PrintLog.Query().Where(printlog.CompanyIDEQ(cid))
	if did != nil {
		q = q.Where(printlog.DepartmentIDEQ(*did))
	}
	if dt := strings.TrimSpace(req.Msg.GetDocumentType()); dt != "" {
		q = q.Where(printlog.DocumentTypeEQ(dt))
	}
	if rid := strings.TrimSpace(req.Msg.GetRouteId()); rid != "" {
		ridN, err := parseID(rid)
		if err != nil {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "route_id"})
		}
		q = q.Where(printlog.RouteIDEQ(ridN))
	}
	if df := strings.TrimSpace(req.Msg.GetDateFrom()); df != "" {
		d, err := time.Parse("2006-01-02", df)
		if err != nil {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "date_from"})
		}
		q = q.Where(printlog.TargetDateGTE(d))
	}
	if dt := strings.TrimSpace(req.Msg.GetDateTo()); dt != "" {
		d, err := time.Parse("2006-01-02", dt)
		if err != nil {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "date_to"})
		}
		q = q.Where(printlog.TargetDateLTE(d.Add(24 * time.Hour)))
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	rows, err := q.Order(ent.Desc(printlog.FieldPrintedAt)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	resp := &productsv1.ListLogsResponse{
		Page: int32(page), PageSize: int32(pageSize), Total: int32(total),
	}
	for _, r := range rows {
		var faURL string
		if fa, err := printFileURL(ctx, db, r.FileAssetID); err == nil {
			faURL = fa
		}
		resp.Entries = append(resp.Entries, &productsv1.LogEntry{
			Id: strconv.Itoa(r.ID), DocumentType: r.DocumentType,
			RouteId: strconv.Itoa(r.RouteID), TargetDate: r.TargetDate.Format("2006-01-02"),
			PrintedBy: strconv.Itoa(r.PrintedBy), PrintedAt: r.PrintedAt.Format(time.RFC3339),
			IsReprint: r.IsReprint, ReprintReason: r.ReprintReason,
			DownloadUrl: faURL,
		})
	}
	return connect.NewResponse(resp), nil
}

// convOrErr 取轉換器(未設定即 fail-closed:server 未組裝產線時拒絕,不靜默空轉)。
func (s *PrintService) convOrErr() print.Converter {
	if s.conv != nil {
		return s.conv
	}
	return errConverter{}
}

// errConverter 未設定產線時的拒絕轉換器。
type errConverter struct{}

func (errConverter) Convert(string) ([]byte, error) {
	return nil, errcode.SysInternal.Error(nil)
}

var _ = printpreview.FieldID
var _ = authz.Identity{}

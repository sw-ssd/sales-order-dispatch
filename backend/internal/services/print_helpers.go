// Package services 的列印輔助(09 計畫 Task 5.5.3–5.5.4):範圍訂單、重印判斷、file_assets 建檔。
package services

import (
	"context"
	"path/filepath"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/fileasset"
	"github.com/salesorder/sales-order-1.0/backend/ent/printlog"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	"github.com/salesorder/sales-order-1.0/backend/internal/print"
)

// ordersInScope 查列印範圍內訂單(部門 + 車次 + 出貨日期當天;RLS 為最後防線)。
func ordersInScope(ctx context.Context, db *ent.Client, cid int, did *int, in print.AssembleInput) ([]*ent.SalesOrder, error) {
	dayStart := in.TargetDate.Truncate(24 * time.Hour)
	dayEnd := dayStart.Add(24 * time.Hour)
	q := db.SalesOrder.Query().
		Where(salesorder.CompanyIDEQ(cid), salesorder.RouteIDEQ(in.RouteID),
			salesorder.DeletedAtIsNil(),
			salesorder.ExpectedDeliveryDateGTE(dayStart), salesorder.ExpectedDeliveryDateLT(dayEnd))
	if did != nil {
		q = q.Where(salesorder.DepartmentIDEQ(*did))
	}
	return q.All(ctx)
}

// hasPrintLog 依比對鍵查既有正式記錄(部門/document/車次/日期/選用選擇器)。
func hasPrintLog(ctx context.Context, db *ent.Client, cid int, did *int, in print.AssembleInput) (bool, error) {
	q := db.PrintLog.Query().
		Where(printlog.CompanyIDEQ(cid), printlog.DocumentTypeEQ(string(in.Type)),
			printlog.RouteIDEQ(in.RouteID), printlog.TargetDateEQ(in.TargetDate.Truncate(24*time.Hour)))
	if did != nil {
		q = q.Where(printlog.DepartmentIDEQ(*did))
	}
	if in.CustomerID != nil {
		q = q.Where(printlog.CustomerIDEQ(*in.CustomerID))
	}
	if in.WarehouseID != nil {
		q = q.Where(printlog.WarehouseIDEQ(*in.WarehouseID))
	}
	return q.Exist(ctx)
}

// printFileInput 為 file_assets 建檔輸入(產線結果 + 範圍)。
type printFileInput struct {
	cid   int
	did   *int
	owner string
	oid   int
	out   *print.Produced
	actor int
	in    print.AssembleInput
}

// printFileMeta 組建檔輸入(owner 初值 0:呼叫端在記錄落筆後回填,此處先以 route 佔位)。
func printFileMeta(cid int, did *int, owner string, oid int, out *print.Produced, actor int, in print.AssembleInput) printFileInput {
	return printFileInput{cid: cid, did: did, owner: owner, oid: oid, out: out, actor: actor, in: in}
}

// createFileAsset 在請求交易內建 file_assets 元資料(PDF 已落檔;DB 失敗由呼叫端補償刪檔)。
//
// 回傳**已存的 URL**(`saved.URL`,由 `fileassets` 與 `SaveUpload` 共用的格式產生)而非讓呼叫端
// 自行拼字串:先前 Preview/Print 各自內聯 `/api/v1/files/<id>/download`,與 `SaveUpload` 寫入的
// 檔名形式(見 `fileassets`)並存兩套形狀 —— 下載路由兩種都接受,但產生端不該有兩個。
func createFileAsset(ctx context.Context, db *ent.Client, m printFileInput) (id int, url string, err error) {
	b := db.FileAsset.Create().
		SetCompanyID(m.cid).SetOwnerType(m.owner).SetOwnerID(m.in.RouteID).
		SetFilename(filepath.Base(m.out.RelPath)).SetOriginalFilename(printDownloadName(m.in)).
		SetMimeType("application/pdf").SetSizeBytes(int(m.out.Size)).
		SetStoragePath(m.out.RelPath).
		SetURL("/api/v1/files/" + filepath.Base(m.out.RelPath) + "/download")
	if m.did != nil {
		b = b.SetDepartmentID(*m.did)
	}
	if m.actor > 0 {
		b = b.SetCreatedBy(m.actor)
	}
	saved, err := b.Save(ctx)
	if err != nil {
		return 0, "", err
	}
	return saved.ID, saved.URL, nil
}

// printFileURL 取 file_asset 下載 URL(記錄查無回空,不擋列表)。
func printFileURL(ctx context.Context, db *ent.Client, faid int) (string, error) {
	fa, err := db.FileAsset.Query().Where(fileasset.IDEQ(faid)).Only(ctx)
	if err != nil {
		return "", err
	}
	return fa.URL, nil
}

// printDownloadName 組原始檔名(類型_日期.pdf)。
func printDownloadName(in print.AssembleInput) string {
	return string(in.Type) + "_" + in.TargetDate.Format("2006-01-02") + ".pdf"
}

package services

import (
	"context"
	"errors"
	"strings"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
)

// companyStatusReasonManual 為 RPC(管理員手動變更)路徑的原因;平台域的訂閱事件 consumer
// (凍結／復原)自帶自己的原因(例:停用連鎖的觸發事由),稽核欄位格式相同。
const companyStatusReasonManual = "管理員手動變更"

// SetCompanyStatus 是公司狀態的**唯一變更入口**:RPC(UpdateCompany)與平台域的訂閱事件
// consumer(凍結／復原)共用,使「停用連鎖」的語意與稽核格式只有一份。
// actor 為觸發者:平台排程觸發時帶系統 actor(見平台設定),仍落租戶稽核。
//
// 三個約定:
//   - **交易由呼叫端擁有**:平台 consumer 須自己開系統範圍交易後呼叫(RLS 下自開交易會被
//     policy 濾成 0 列,見 dbtenant 檔頭),故 ctx 無請求交易即拒絕,本函式不開交易也不包 scope。
//   - **原因必填**:狀態真的改變時必須說明理由(Platform Global Constraints),空字串即拒絕。
//     寫入的稽核帶得出「誰、何時、為何」,否則平台工具改動的痕跡無法回溯。
//   - **同值為 no-op**:狀態機只有一條路徑,但「已達目標狀態」的重複事件不留痕 ——
//     否則每日重跑的排程會把稽核灌爆。
func SetCompanyStatus(ctx context.Context, db *ent.Client, companyID int,
	status company.Status, reason string, actor authz.Identity) error {
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		// 交易缺失是伺服器端程式錯誤(排程／consumer 沒包交易),非輸入問題 → 走系統碼。
		// 用 errcode 而非 connect.NewError:錯誤建構集中由 registry 管(見 errcode 守門測試)。
		return errcode.SysInternal.Wrap(errors.New("SetCompanyStatus 需在呼叫端的交易內執行"))
	}
	actorID, err := parseID(actor.UserID)
	if err != nil {
		return err // InvalidArgument:稽核必須有可歸屬的觸發者,不接受無來源的 actor
	}
	cur, err := dbtenant.Client(ctx, db).Company.Query().
		Where(company.ID(companyID), company.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return toConnectError(err)
	}
	// 未結項 #1:同值 no-op 排在 reason 檢查之前 —— 「無需動作」不該被原因阻擋。
	// no-op 不寫稽核、無副作用，故不需要原因；真正的狀態變更仍由下一段的 reason 檢查把關。
	if cur.Status == status {
		return nil
	}
	if strings.TrimSpace(reason) == "" {
		// 尚無專碼(新增專碼屬 errcode 註冊範疇),故回一般 error:原因文字原樣可見。
		return errors.New("狀態變更必須提供原因")
	}

	// 條件更新:「已軟刪除的公司不得改狀態」是敘述式條件的一部分,不只是前置查詢 ——
	// 中間沒有可插入的縫(比照 DeleteDepartment 的條件式刪除)。
	updated, err := tx.Client().Company.UpdateOneID(companyID).
		Where(company.DeletedAtIsNil()).SetStatus(status).Save(ctx)
	if err != nil {
		return toConnectError(err)
	}

	if err := recordAuditBA(ctx, tx, "company", "update", companyID, companyID, nil, actorID,
		map[string]any{"status": string(cur.Status)},
		map[string]any{"status": string(updated.Status), "reason": strings.TrimSpace(reason)},
	); err != nil {
		return toConnectError(err)
	}
	return nil
}

// Package services 的訂單取號(05 計畫 Task 2, D7):「來源碼 + 6 位補零序號」,
// order_counters 依 (company_id, source) 一列,樂觀鎖更新,取號與建單同交易。
// 本檔只放取號與來源驗證;建單組裝另票。
package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/metadict"
	"github.com/salesorder/sales-order-1.0/backend/ent/ordercounter"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
)

// maxOrderNoRetries 取號樂觀鎖重試上限(D7,比照客戶取號 5 次)。
const maxOrderNoRetries = 5

// ensureOrderCounter 確保該公司該來源的 counter 列存在(冪等的建前步驟)。
// 併發首次建立撞唯一索引時回錯誤由呼叫端重試(比照 ensureCustomerCounter:
// 請求交易內敘述錯誤會 abort 整交易,無法吞掉衝突再繼續)。
func ensureOrderCounter(ctx context.Context, db *ent.Client, cid int, source string) error {
	client := dbtenant.Client(ctx, db)
	exists, err := client.OrderCounter.Query().
		Where(ordercounter.CompanyIDEQ(cid), ordercounter.SourceEQ(source)).Exist(ctx)
	if err != nil {
		return toConnectError(err)
	}
	if exists {
		return nil
	}
	if _, err := client.OrderCounter.Create().
		SetCompanyID(cid).SetSource(source).SetNextSeq(1).SetVersion(0).Save(ctx); err != nil {
		return toConnectError(err)
	}
	return nil
}

// validateOrderSource 驗證 source 為 metadicts order_source 系統級字典的有效碼且啟用中。
// 來源碼同時是取號軌道與編號前綴,無效碼不得消耗序號(守衛在取號之前執行)。
func validateOrderSource(ctx context.Context, db *ent.Client, source string) error {
	source = strings.TrimSpace(source)
	if source == "" {
		return errcode.SysInvalidArgument.Error(map[string]string{"field": "source"})
	}
	ok, err := dbtenant.Client(ctx, db).Metadict.Query().Where(
		metadict.TypeEQ("order_source"), metadict.CodeEQ(source),
		metadict.IsActiveEQ(true), metadict.DeletedAtIsNil(),
		metadict.DepartmentIDIsNil(), // 系統級固定(D11),不讀部門擴充
	).Exist(ctx)
	if err != nil {
		return toConnectError(err)
	}
	if !ok {
		return errcode.SysInvalidArgument.Error(map[string]string{"field": "source"})
	}
	return nil
}

// nextOrderNo 於請求交易內以樂觀鎖 counter 取號,回傳 order_no(來源碼 + 6 位補零)。
// 呼叫前須先 ensureOrderCounter + validateOrderSource。version 衝突則重試;逾限回 failed_precondition。
// 序號超過 999999 自然增長位數,不截斷。呼叫端交易回滾時序號不消耗(同一交易內)。
func nextOrderNo(ctx context.Context, db *ent.Client, cid int, source string) (string, error) {
	for i := 0; i < maxOrderNoRetries; i++ {
		c, err := db.OrderCounter.Query().
			Where(ordercounter.CompanyIDEQ(cid), ordercounter.SourceEQ(source)).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return "", errcode.SysInternal.Wrap(err)
			}
			return "", err
		}
		n, err := db.OrderCounter.Update().
			Where(ordercounter.CompanyIDEQ(cid), ordercounter.SourceEQ(source), ordercounter.VersionEQ(c.Version)).
			SetNextSeq(c.NextSeq + 1).SetVersion(c.Version + 1).Save(ctx)
		if err != nil {
			return "", err
		}
		if n == 0 {
			continue // version 衝突 → 重試
		}
		return fmt.Sprintf("%s%06d", source, c.NextSeq), nil
	}
	return "", errcode.SysInternal.Error(nil)
}

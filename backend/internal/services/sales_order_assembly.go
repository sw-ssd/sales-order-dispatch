// Package services 的下單組裝(05 計畫 Task 5, 4.2.1–4.2.5):單位換算、手打別名、
// 客戶專屬守衛、來源記錄、偏好送貨日順延。全部落於 Create/Update 交易內,失敗即整單回滾。
package services

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/ent/customerproduct"
	"github.com/salesorder/sales-order-1.0/backend/ent/product"
	"github.com/salesorder/sales-order-1.0/backend/ent/productunit"
	domainproducts "github.com/salesorder/sales-order-1.0/backend/internal/domain/products"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// isCustomerIdentity 呼叫者是否為客戶帳號/customer 主帳號(守衛 4.2.3 用)。
func isCustomerIdentity(roles []string) bool {
	for _, r := range roles {
		if r == "customer" || r == "guest" {
			return true
		}
	}
	return false
}

// resolveBaseQty 依選用單位換算 base_qty(4.2.1):讀 product_units 找 unit 列,
// is_base 直填,否則 qty × conversion_rate(十進位有理數,3 位小數)。
// productID==0 為手打品名:無換算來源,base_qty = qty。
func resolveBaseQty(ctx context.Context, db *ent.Client, productID int, unit, qty string) (string, error) {
	qtyRat, err := domainproducts.ParseQty(qty)
	if err != nil {
		return "", errcode.SysInvalidArgument.Error(map[string]string{"field": "items.qty"})
	}
	if qtyRat.Sign() <= 0 {
		return "", errcode.SysInvalidArgument.Error(map[string]string{"field": "items.qty"})
	}
	if productID == 0 {
		return domainproducts.ToBase(big.NewRat(1, 1), qtyRat), nil
	}
	units, err := db.ProductUnit.Query().
		Where(productunit.ProductIDEQ(productID)).All(ctx)
	if err != nil {
		return "", errcode.SysInternal.Wrap(err)
	}
	for _, u := range units {
		if u.UnitCode != unit {
			continue
		}
		rate, err := domainproducts.ParseRate(u.ConversionRate)
		if err != nil {
			return "", errcode.SysInternal.Wrap(err)
		}
		return domainproducts.ToBase(rate, qtyRat), nil
	}
	return "", errcode.SysInvalidArgument.Error(map[string]string{"field": "items.unit"})
}

// validateOrderItem 驗證單一明細(4.2.1–4.2.3):數量/單位/名稱必填;商品存在且未刪;
// customer_products 未落地(04 Task 3.5)前,客戶帳號不可手打、品項不限清單(放行,RLS 為防線)。
// 回傳 productID(0=手打)、display、baseQty。
func validateOrderItem(ctx context.Context, db *ent.Client, cid int, isCustomer bool,
	item *salesorderv1.OrderItemInput) (productID int, display, baseQty string, err error) {
	qty := strings.TrimSpace(item.GetQty())
	unit := strings.TrimSpace(item.GetUnit())
	if qty == "" || unit == "" {
		return 0, "", "", errcode.SysInvalidArgument.Error(map[string]string{"field": "items"})
	}
	display = strings.TrimSpace(item.GetDisplayName())
	if display == "" {
		display = strings.TrimSpace(item.GetManualName())
	}
	if display == "" {
		return 0, "", "", errcode.SysInvalidArgument.Error(map[string]string{"field": "items.display_name"})
	}
	if strings.TrimSpace(item.GetManualName()) != "" {
		// 手打為業務功能(4.2.2):客戶帳號送出手打即拒。
		if isCustomer {
			return 0, "", "", errcode.SysPermissionDenied.Error(nil)
		}
		baseQty, err := resolveBaseQty(ctx, db, 0, unit, qty)
		if err != nil {
			return 0, "", "", err
		}
		return 0, display, baseQty, nil
	}
	// T4 的最小可用允許「無 product_id、僅 display_name」的明細(歷史相容);
	// T5 起該形狀視為手打(業務限定),不再靜默放行。
	if strings.TrimSpace(item.GetProductId()) == "" {
		if isCustomer {
			return 0, "", "", errcode.SysPermissionDenied.Error(nil)
		}
		baseQty, err := resolveBaseQty(ctx, db, 0, unit, qty)
		if err != nil {
			return 0, "", "", err
		}
		return 0, display, baseQty, nil
	}
	pid, err := parseID(item.GetProductId())
	if err != nil {
		return 0, "", "", err
	}
	exists, err := db.Product.Query().
		Where(product.ID(pid), product.CompanyIDEQ(cid), product.DeletedAtIsNil()).Exist(ctx)
	if err != nil {
		return 0, "", "", toConnectError(err)
	}
	if !exists {
		return 0, "", "", errcode.SysNotFound.Error(nil)
	}
	baseQty, err = resolveBaseQty(ctx, db, pid, unit, qty)
	if err != nil {
		return 0, "", "", err
	}
	return pid, display, baseQty, nil
}

// adjustDeliveryDate 偏好送貨日順延(4.2.5,D26):preferredDays 為週一~六布林陣列(長度 6)。
// 未勾選任何日/陣列異常 → 維持原選擇;所選落勾選日(含週日視同非勾選→順延) → 維持;
// 否則逐日往後找下一個勾選日(至多 7 天)。日期以 UTC+8 營業日曆判星期。
func adjustDeliveryDate(preferredDays []bool, date time.Time) time.Time {
	if len(preferredDays) != 6 {
		return date
	}
	any := false
	for _, b := range preferredDays {
		if b {
			any = true
			break
		}
	}
	if !any {
		return date
	}
	loc := time.FixedZone("UTC+8", 8*3600)
	for i := 0; i < 7; i++ {
		d := date.AddDate(0, 0, i)
		wd := d.In(loc).Weekday() // Sunday=0..Saturday=6
		if wd == time.Sunday {
			continue
		}
		if preferredDays[int(wd)-1] {
			return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
		}
	}
	return date
}

// upsertCustomerAlias 同交易 upsert 別名(4.2.2):存在改 alias,不存在則建。
// 唯一衝突(併發) → 重讀改更新;仍失敗整單回滾(呼叫端交易)。
func upsertCustomerAlias(ctx context.Context, db *ent.Client, cid, custID, pid int, alias string) error {
	if existing, err := db.CustomerProduct.Query().
		Where(customerproduct.CustomerIDEQ(custID), customerproduct.ProductIDEQ(pid),
			customerproduct.DeletedAtIsNil()).Only(ctx); err == nil {
		_, err := db.CustomerProduct.UpdateOneID(existing.ID).SetAliasName(alias).Save(ctx)
		return err
	} else if !ent.IsNotFound(err) {
		return toConnectError(err)
	}
	cust, err := db.Customer.Query().
		Where(customer.ID(custID), customer.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return toConnectError(err)
	}
	build := db.CustomerProduct.Create().
		SetCompanyID(cid).SetCustomerID(custID).SetProductID(pid).
		SetAliasName(alias).SetDefaultQty("0")
	if cust.DepartmentID != nil {
		build = build.SetDepartmentID(*cust.DepartmentID)
	}
	if _, err := build.Save(ctx); err != nil {
		if ent.IsConstraintError(err) {
			if existing, rerr := db.CustomerProduct.Query().
				Where(customerproduct.CustomerIDEQ(custID), customerproduct.ProductIDEQ(pid),
					customerproduct.DeletedAtIsNil()).Only(ctx); rerr == nil {
				_, err := db.CustomerProduct.UpdateOneID(existing.ID).SetAliasName(alias).Save(ctx)
				return err
			}
		}
		return toConnectError(err)
	}
	return nil
}

// ensureCustomerListEntry 選用總表商品自動加入清單(4.2.2):無記錄即建預設列(別名=商品名)。
func ensureCustomerListEntry(ctx context.Context, db *ent.Client, cid, custID, pid int) error {
	exists, err := db.CustomerProduct.Query().
		Where(customerproduct.CustomerIDEQ(custID), customerproduct.ProductIDEQ(pid),
			customerproduct.DeletedAtIsNil()).Exist(ctx)
	if err != nil {
		return toConnectError(err)
	}
	if exists {
		return nil
	}
	prod, err := db.Product.Query().Where(product.ID(pid)).Only(ctx)
	if err != nil {
		return toConnectError(err)
	}
	cust, err := db.Customer.Query().
		Where(customer.ID(custID), customer.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return toConnectError(err)
	}
	build := db.CustomerProduct.Create().
		SetCompanyID(cid).SetCustomerID(custID).SetProductID(pid).
		SetAliasName(prod.Name).SetDefaultQty("0")
	if cust.DepartmentID != nil {
		build = build.SetDepartmentID(*cust.DepartmentID)
	}
	if _, err := build.Save(ctx); err != nil {
		// 併發已建 → 吸收(冪等)。
		if ent.IsConstraintError(err) {
			return nil
		}
		return toConnectError(err)
	}
	return nil
}

// customerPreferredDays 讀客戶偏好陣列(4.2.5):客戶不存在 → not_found。
func customerPreferredDays(ctx context.Context, db *ent.Client, cid, custID int) ([]bool, error) {
	c, err := db.Customer.Query().
		Where(customer.ID(custID), customer.CompanyIDEQ(cid), customer.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.SysNotFound.Error(nil)
		}
		return nil, toConnectError(err)
	}
	return c.PreferredDeliveryDays, nil
}

// orderCustomerGuard 客戶下單守衛(4.2.3):客戶子帳號強制 customer_id 為自己;
// customer_products 未落地前不清單檢查(放行,RLS self 為最後防線);主帳號一律拒絕由 Casbin 層處理。
// 回傳實際落單的 customerID。
//
// 身分未帶 customer_id 時(今日 middleware 未填該欄):請求值為空即報錯,請求值非空則
// 照請求值走—— RLS self 仍是最後防線(跨客戶讀寫不到),不因守衛放行而外洩。
func orderCustomerGuard(customerIDStr, selfCustomerID string, isCustomer bool) (int, error) {
	if !isCustomer || strings.TrimSpace(selfCustomerID) == "" {
		return parseID(customerIDStr)
	}
	// 客戶帳號:忽略請求 customer_id,強制為自己(建議拒絕改為忽略落單自己,API 文件註明)。
	return parseID(selfCustomerID)
}

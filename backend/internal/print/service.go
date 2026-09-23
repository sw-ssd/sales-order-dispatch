// Package print 的資料組合(09 計畫 Task 5.4.2):依單據類型從訂單資料組 view model。
// 狀態過濾不在此層(預覽不限狀態、正式列印限 processing 由 5.5.3 執行);查無明細回空表標記。
package print

import (
	"context"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/ent/customeraddress"
	"github.com/salesorder/sales-order-1.0/backend/ent/customercontact"
	"github.com/salesorder/sales-order-1.0/backend/ent/processingspec"
	"github.com/salesorder/sales-order-1.0/backend/ent/product"
	"github.com/salesorder/sales-order-1.0/backend/ent/productcategory"
	"github.com/salesorder/sales-order-1.0/backend/ent/route"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorderitem"
	"github.com/salesorder/sales-order-1.0/backend/ent/warehouse"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
)

// AssembleInput 為組合輸入:單據類型 + 車次 + 出貨日期 + 選用選擇器。
type AssembleInput struct {
	Type        DocumentType
	RouteID     int
	TargetDate  time.Time
	CustomerID  *int // 對點單單印一店
	WarehouseID *int // 揀貨單單印一倉
}

// Assembled 為組合結果:view model + 空表標記。
type Assembled struct {
	Model any
	Empty bool
}

// Assemble 依類型組 view model(部門範圍由請求交易 RLS 注入;跨部門查不到視同不存在)。
func Assemble(ctx context.Context, db *ent.Client, in AssembleInput) (*Assembled, error) {
	if in.RouteID <= 0 {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "route_id"})
	}
	switch in.Type {
	case DispatchSummary, DeliveryNote, PickingList, ProcessingList:
	default:
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "document_type"})
	}
	if in.Type == PickingList && in.CustomerID != nil {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "customer_id"})
	}
	if in.Type == DeliveryNote && in.WarehouseID != nil {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "warehouse_id"})
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	qdb := tx.Client()
	r, err := qdb.Route.Query().Where(route.ID(in.RouteID)).Only(ctx)
	if err != nil {
		return nil, notFoundErr(err)
	}
	orders, err := ordersFor(qdb, ctx, in)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return &Assembled{Empty: true}, nil
	}
	coName := companyNameOf(ctx, qdb, orders[0].CompanyID)
	switch in.Type {
	case DispatchSummary:
		m, empty := summaryModel(ctx, qdb, coName, r, in, orders)
		return &Assembled{Model: m, Empty: empty}, nil
	case DeliveryNote:
		m, empty := noteModel(ctx, qdb, coName, r, in, orders)
		return &Assembled{Model: m, Empty: empty}, nil
	case PickingList:
		m, empty := pickingModel(ctx, qdb, coName, r, in, orders)
		return &Assembled{Model: m, Empty: empty}, nil
	default:
		m, empty := processingModel(ctx, qdb, coName, r, in, orders)
		return &Assembled{Model: m, Empty: empty}, nil
	}
}

// orderRow 為訂單 + 明細載體。
type orderRow struct {
	*ent.SalesOrder
	items []*ent.SalesOrderItem
}

// ordersFor 查詢範圍內訂單(部門 + 車次 + 出貨日期當天;RLS 為最後防線)。
func ordersFor(qdb *ent.Client, ctx context.Context, in AssembleInput) ([]orderRow, error) {
	dayStart := in.TargetDate.Truncate(24 * time.Hour)
	dayEnd := dayStart.Add(24 * time.Hour)
	q := qdb.SalesOrder.Query().
		Where(salesorder.RouteIDEQ(in.RouteID), salesorder.DeletedAtIsNil(),
			salesorder.ExpectedDeliveryDateGTE(dayStart), salesorder.ExpectedDeliveryDateLT(dayEnd)).
		Order(ent.Asc(salesorder.FieldDeliverySequence), ent.Asc(salesorder.FieldID))
	orders, err := q.All(ctx)
	if err != nil {
		return nil, errcode.SysInternal.Wrap(err)
	}
	var out []orderRow
	for _, o := range orders {
		items, err := qdb.SalesOrderItem.Query().
			Where(salesorderitem.SalesOrderIDEQ(o.ID), salesorderitem.DeletedAtIsNil()).
			Order(ent.Asc(salesorderitem.FieldID)).All(ctx)
		if err != nil {
			return nil, errcode.SysInternal.Wrap(err)
		}
		if len(items) == 0 {
			continue
		}
		out = append(out, orderRow{SalesOrder: o, items: items})
	}
	return out, nil
}

// customerOf 查客戶(軟刪除主檔仍顯示歷史名稱:查不到以編號兜底,D10)。
func customerOf(ctx context.Context, qdb *ent.Client, cid int) (code, name, addr, contact, phone string) {
	c, err := qdb.Customer.Query().Where(customer.ID(cid)).Only(ctx)
	if err != nil {
		return "", "（已刪除客戶）", "", "", ""
	}
	if a, err := qdb.CustomerAddress.Query().
		Where(customeraddress.CustomerIDEQ(cid), customeraddress.IsDefaultEQ(true),
			customeraddress.DeletedAtIsNil()).Only(ctx); err == nil {
		addr = a.AddressLine
	}
	if cc, err := qdb.CustomerContact.Query().
		Where(customercontact.CustomerIDEQ(cid), customercontact.IsDefaultEQ(true),
			customercontact.DeletedAtIsNil()).Only(ctx); err == nil {
		contact, phone = cc.Name, cc.Phone
	}
	return c.CustomerCode, c.Name, addr, contact, phone
}

// itemName 取明細顯示名:有 product 查商品名(刪品顯示歷史編號),無 product(手打品)
// 取下單時快照的 DisplayName。
func itemName(ctx context.Context, qdb *ent.Client, it *ent.SalesOrderItem) string {
	if it.ProductID == nil {
		if it.DisplayName != "" {
			return it.DisplayName
		}
		return "（手打品）"
	}
	p, err := qdb.Product.Query().Where(product.ID(*it.ProductID)).Only(ctx)
	if err != nil {
		return "（已刪除商品）"
	}
	return p.Name
}

// pidOf 解 *int（nil → 0，呼叫端須先判手打品）。
func pidOf(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// specName 查處理規格名稱(無則空)。
func specName(ctx context.Context, qdb *ent.Client, sid *int) string {
	if sid == nil {
		return ""
	}
	s, err := qdb.ProcessingSpec.Query().Where(processingspec.ID(*sid)).Only(ctx)
	if err != nil {
		return ""
	}
	return s.Name
}

// warehouseName 查倉別名稱(無則空)。
func warehouseName(ctx context.Context, qdb *ent.Client, wid *int) string {
	if wid == nil {
		return ""
	}
	w, err := qdb.Warehouse.Query().Where(warehouse.ID(*wid)).Only(ctx)
	if err != nil {
		return ""
	}
	return w.Name
}

// companyNameOf 查公司名稱(查不到回空,不擋列印)。
func companyNameOf(ctx context.Context, qdb *ent.Client, cid int) string {
	co, err := qdb.Company.Query().Where(company.ID(cid)).Only(ctx)
	if err != nil {
		return ""
	}
	return co.Name
}

// summaryModel 組單車總表:店家依 delivery_sequence(查詢已排序),品項原始下單數量與單位。
func summaryModel(ctx context.Context, qdb *ent.Client, coName string, r *ent.Route, in AssembleInput, orders []orderRow) (DispatchSummaryModel, bool) {
	m := DispatchSummaryModel{CompanyName: coName, RouteCode: r.Code, RouteName: r.Name,
		TargetDate: dateStr(in.TargetDate), PrintedAt: nowStr()}
	for _, o := range orders {
		code, name, addr, _, _ := customerOf(ctx, qdb, o.CustomerID)
		seq := 0
		if o.DeliverySequence != nil {
			seq = *o.DeliverySequence
		}
		blk := StoreBlock{CustomerCode: code, Name: name, Address: addr, Sequence: seq}
		for _, it := range o.items {
			blk.Items = append(blk.Items, StoreItem{
				Name: itemName(ctx, qdb, it), Qty: it.Qty, Unit: it.Unit,
				SpecialCut: it.SpecialCutNote, ProcessSpec: specName(ctx, qdb, it.ProcessingSpecID),
				HasCutOrSpec: it.SpecialCutNote != "" || it.ProcessingSpecID != nil,
			})
		}
		m.Stores = append(m.Stores, blk)
	}
	return m, len(m.Stores) == 0
}

// noteModel 組對點單:同車次多店家;指定 customer 時僅取該店。
func noteModel(ctx context.Context, qdb *ent.Client, coName string, r *ent.Route, in AssembleInput, orders []orderRow) (DeliveryNoteModel, bool) {
	m := DeliveryNoteModel{CompanyName: coName, RouteCode: r.Code, RouteName: r.Name,
		TargetDate: dateStr(in.TargetDate)}
	for _, o := range orders {
		if in.CustomerID != nil && o.CustomerID != *in.CustomerID {
			continue
		}
		code, name, addr, contact, phone := customerOf(ctx, qdb, o.CustomerID)
		seq := 0
		if o.DeliverySequence != nil {
			seq = *o.DeliverySequence
		}
		blk := StoreBlock{CustomerCode: code, Name: name, Address: addr,
			Contact: contact, Phone: phone, Sequence: seq}
		for _, it := range o.items {
			blk.Items = append(blk.Items, StoreItem{
				Name: itemName(ctx, qdb, it), Qty: it.Qty, Unit: it.Unit,
				SpecialCut: it.SpecialCutNote, ProcessSpec: specName(ctx, qdb, it.ProcessingSpecID),
				HasCutOrSpec: it.SpecialCutNote != "" || it.ProcessingSpecID != nil,
			})
		}
		m.Stores = append(m.Stores, blk)
	}
	return m, len(m.Stores) == 0
}

// pickingModel 組揀貨單:車次 → 倉別 → 分類 → 品名;同商品跨店合併,base_qty 加總。
func pickingModel(ctx context.Context, qdb *ent.Client, coName string, r *ent.Route, in AssembleInput, orders []orderRow) (PickingListModel, bool) {
	type key struct {
		wid int
		pid int
	}
	agg := map[key]*PickingRow{}
	catOf := map[key]string{}
	for _, o := range orders {
		for _, it := range o.items {
			if in.WarehouseID != nil && (it.WarehouseID == nil || *it.WarehouseID != *in.WarehouseID) {
				continue
			}
			wid := 0
			if it.WarehouseID != nil {
				wid = *it.WarehouseID
			}
			k := key{wid: wid, pid: pidOf(it.ProductID)}
			row, ok := agg[k]
			if !ok {
				row = &PickingRow{Name: itemName(ctx, qdb, it), Unit: it.Unit}
				catOf[k] = categoryOf(ctx, qdb, pidOf(it.ProductID))
				agg[k] = row
			}
			row.BaseQty = addQty(row.BaseQty, it.BaseQty)
			if it.SpecialCutNote != "" || it.ProcessingSpecID != nil {
				row.NeedsCut = true
			}
		}
	}
	m := PickingListModel{CompanyName: coName, RouteCode: r.Code, RouteName: r.Name,
		TargetDate: dateStr(in.TargetDate)}
	if in.WarehouseID != nil {
		m.WarehouseName = warehouseName(ctx, qdb, in.WarehouseID)
	}
	type ranked struct {
		cat string
		row *PickingRow
	}
	var list []ranked
	for k, row := range agg {
		list = append(list, ranked{cat: catOf[k], row: row})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].cat != list[j].cat {
			return list[i].cat < list[j].cat
		}
		return list[i].row.Name < list[j].row.Name
	})
	for _, e := range list {
		e.row.Category = e.cat
		m.Rows = append(m.Rows, *e.row)
	}
	return m, len(m.Rows) == 0
}

// processingModel 組加工單:僅有 spec 或 cut note 的明細;加工室揀在前、配送揀在後。
func processingModel(ctx context.Context, qdb *ent.Client, coName string, r *ent.Route, in AssembleInput, orders []orderRow) (ProcessingListModel, bool) {
	m := ProcessingListModel{CompanyName: coName, RouteCode: r.Code, RouteName: r.Name,
		TargetDate: dateStr(in.TargetDate)}
	for _, o := range orders {
		_, storeName, _, _, _ := customerOf(ctx, qdb, o.CustomerID)
		for _, it := range o.items {
			if it.SpecialCutNote == "" && it.ProcessingSpecID == nil {
				continue
			}
			spec := specName(ctx, qdb, it.ProcessingSpecID)
			m.Workshop = append(m.Workshop, ProcessRow{
				Warehouse: warehouseName(ctx, qdb, it.WarehouseID),
				Name:      itemName(ctx, qdb, it),
				Qty:       it.BaseQty, Unit: it.Unit, Spec: spec,
			})
			m.Delivery = append(m.Delivery, ProcessRow{
				Store: storeName, Name: itemName(ctx, qdb, it),
				Qty: it.Qty, Unit: it.Unit, Spec: it.SpecialCutNote,
			})
		}
	}
	sort.Slice(m.Workshop, func(i, j int) bool {
		if m.Workshop[i].Warehouse != m.Workshop[j].Warehouse {
			return m.Workshop[i].Warehouse < m.Workshop[j].Warehouse
		}
		return m.Workshop[i].Name < m.Workshop[j].Name
	})
	return m, len(m.Workshop) == 0 && len(m.Delivery) == 0
}

// categoryOf 查商品分類名稱(無則空)。
func categoryOf(ctx context.Context, qdb *ent.Client, pid int) string {
	if pid == 0 {
		return ""
	}
	p, err := qdb.Product.Query().Where(product.ID(pid)).Only(ctx)
	if err != nil || p.CategoryID == nil {
		return ""
	}
	c, err := qdb.ProductCategory.Query().Where(productcategory.ID(*p.CategoryID)).Only(ctx)
	if err != nil {
		return ""
	}
	return c.Name
}

// notFoundErr 轉查無為 not_found,其餘 internal。
func notFoundErr(err error) error {
	if ent.IsNotFound(err) {
		return errcode.SysNotFound.Error(nil)
	}
	return errcode.SysInternal.Wrap(err)
}

// addQty 以十進位字串相加(基本單位彙總;解析失敗以後者兜底,不擋列印)。
func addQty(a, b string) string {
	ra, oka := parseDecimal(a)
	rb, okb := parseDecimal(b)
	if !oka {
		return b
	}
	if !okb {
		return a
	}
	return ratToDecimal3(new(big.Rat).Add(ra, rb))
}

// dateStr 取日期字串。
func dateStr(t time.Time) string { return t.Format("2006-01-02") }

// nowStr 取列印時間。
func nowStr() string { return time.Now().Format("2006-01-02 15:04") }

// strings 佔位(匯入檢查)。
var _ = strings.TrimSpace

// parseDecimal 解析十進位字串為有理數(空字串視為缺失,不合法)。
func parseDecimal(s string) (*big.Rat, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, false
	}
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return nil, false
	}
	return r, true
}

// ratToDecimal3 有理數轉十進位文字(去尾零;與 products.conversion 同規則,內聯避免跨層依賴)。
func ratToDecimal3(r *big.Rat) string {
	if r.Sign() == 0 {
		return "0"
	}
	f := new(big.Float).SetRat(r)
	return strings.TrimRight(strings.TrimRight(f.Text('f', 3), "0"), ".")
}

// Package print 單據列印(09 計畫 Task 5.3–5.5):四種 A4 HTML 模板 + view model。
//
// 全文無金額(D12/D15):view model 從源頭不含金額欄位,模板層無從渲染價格。
// 模板以 embed 打包(html/template);渲染層不回 Connect code(資料問題由組合層回 not_found)。
package print

import (
	"bytes"
	"embed"
	"html/template"
	"strings"
)

//go:embed templates/*.html
var templateFS embed.FS

// DocumentType 為四種單據類型。
type DocumentType string

const (
	DispatchSummary DocumentType = "dispatch_summary" // 單車總表(5.3.1)
	DeliveryNote    DocumentType = "delivery_note"    // 對點單(5.3.2)
	PickingList     DocumentType = "picking_list"     // 揀貨單(5.3.3)
	ProcessingList  DocumentType = "processing_list"  // 加工單(5.3.4)
)

// templateFile 對映類型→模板檔。
func templateFile(t DocumentType) string {
	switch t {
	case DispatchSummary:
		return "templates/dispatch_summary.html"
	case DeliveryNote:
		return "templates/delivery_note.html"
	case PickingList:
		return "templates/picking_list.html"
	case ProcessingList:
		return "templates/processing_list.html"
	default:
		return ""
	}
}

// StoreItem 為店家品項摘要(品名、數量、單位、特殊分切備註)。
type StoreItem struct {
	Name         string
	Qty          string
	Unit         string
	SpecialCut   string
	ProcessSpec  string
	HasCutOrSpec bool
}

// StoreBlock 為一店家區塊(單車總表/對點單用)。
type StoreBlock struct {
	CustomerCode string
	Name         string
	Address      string
	Contact      string
	Phone        string
	Sequence     int
	Items        []StoreItem
}

// DispatchSummaryModel 為單車總表 view model:單一車次 + 單一出貨日期,店家依 delivery_sequence 升冪。
type DispatchSummaryModel struct {
	CompanyName string
	RouteCode   string
	RouteName   string
	TargetDate  string
	PrintedAt   string
	Stores      []StoreBlock
}

// DeliveryNoteModel 為對點單 view model:同車次多店家,每店一張 A4(模板層換頁)。
type DeliveryNoteModel struct {
	CompanyName string
	RouteCode   string
	RouteName   string
	TargetDate  string
	Stores      []StoreBlock
}

// PickingRow 為揀貨彙總列(同商品跨店合併,基本單位)。
type PickingRow struct {
	Category string
	Name     string
	BaseQty  string
	Unit     string
	NeedsCut bool
}

// PickingListModel 為揀貨單 view model:單一車次 + 單一倉別 + 單一出貨日期。
type PickingListModel struct {
	CompanyName   string
	RouteCode     string
	RouteName     string
	TargetDate    string
	WarehouseName string
	Rows          []PickingRow
}

// ProcessRow 為加工列(加工後數量一律空白,手寫回填,1.0 不回寫)。
type ProcessRow struct {
	Warehouse string
	Store     string
	Name      string
	Qty       string
	Unit      string
	Spec      string
}

// ProcessingListModel 為加工單 view model:加工室揀在前、配送揀在後。
type ProcessingListModel struct {
	CompanyName string
	RouteCode   string
	RouteName   string
	TargetDate  string
	Workshop    []ProcessRow
	Delivery    []ProcessRow
}

// Render 渲染指定類型模板;未知類型回錯誤(呼叫端轉 invalid_argument)。
func Render(t DocumentType, model any) (string, error) {
	file := templateFile(t)
	if file == "" {
		return "", errUnknownType(string(t))
	}
	tmpl, err := template.ParseFS(templateFS, file)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	name := file[strings.LastIndex(file, "/")+1:]
	if err := tmpl.ExecuteTemplate(&b, name, model); err != nil {
		return "", err
	}
	return b.String(), nil
}

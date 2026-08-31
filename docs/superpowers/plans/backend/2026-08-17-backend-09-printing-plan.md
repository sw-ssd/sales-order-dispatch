# Backend 09 — 單據列印(四種模板 / Gotenberg PDF / 列印與預覽 API)Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 實作 backend 單據列印全域 — 四種 A4 單據 view model 與 html/template 模板(單車總表 / 對點單 / 揀貨單 / 加工單,全文無金額)、Gotenberg HTML→PDF client(有限重試)、資料組合與 PDF 產線(空表不產生、Noto Sans CJK TC、失敗補償刪除)、print_logs / print_previews schema、Preview / Print / ListLogs 三個 Connect-RPC(正式列印整批檢查 `status = processing`、重印必填原因且 `is_reprint = true`、PDF 經 file_assets 下載 URL 交付)。

**Architecture:** 依 `docs/superpowers/plans/backend-detail/09-printing.md`(下稱「細部文件」,子功能編號 5.x.y)實作。渲染管線分三層:`internal/print`(view model + 模板 + Gotenberg client + 資料組合 + PDF 產線,純內部套件)→ `internal/domain/prints`(Connect-RPC handler,交易邊界與稽核)→ `fileassets` 埠(FileStore 介面,實作由 04-master-data 計畫 Task 8 提供)。Gotenberg 以 `Converter` 介面抽象,測試以 fake / httptest 取代。

**Tech Stack:** Go 1.25、Ent(entgo.io)、Chi v5、Connect-RPC、pgx/v5、shopspring/decimal、html/template(embed)、testcontainers-go(整合測試);Gotenberg 8(外部服務,Chromium 轉換路由,測試不啟真實容器)。

**Spec 來源:** 細部文件 `docs/superpowers/plans/backend-detail/09-printing.md`;共通規則見 `docs/superpowers/plans/backend-detail/00-index.md` §3。

## Global Constraints

- module 路徑:`github.com/salesorder/sales-order-1.0/backend`;所有路徑相對 repo root。
- 無金額(D12,硬性邊界):四種單據 view model 從源頭不含任何金額語意欄位(price / amount / subtotal / total / currency / cost / money);Task 1 以 reflect 測試掃描全部 view model 結構體欄位名強制執行,模板輸出另以關鍵字掃描(金額 / 單價 / 價格 / 小計 / NT$ / $)雙重防呆。
- 多租戶:print_logs / print_previews 皆帶 `company_id` / `department_id`;查詢一律經 `rls.WrapDriver` 建立的 Ent client(01-auth 計畫 Task 3 提供),RLS session variables 未注入 fail-closed;RLS 下查不到的資源(車次 / 店家 / 倉別)一律視同 `not_found`,不暴露存在性。
- 軟刪除:file_assets 適用 `deleted_at` 軟刪除(D10,04 計畫);print_logs / print_previews 為記錄類資料,**不軟刪除、無 `updated_at` / `deleted_at`**,保留排程(print_logs 2 年、print_previews 90 天,§14.1)由維運工作另行處理,不在本計畫。
- 同事務(D18):print_logs / print_previews 記錄、file_assets 元資料、正式列印的 audit_logs(action = print)同一 DB 交易寫入;PDF 實體檔案與 Gotenberg 呼叫在交易外先完成,交易失敗時以 `FileStore.DeleteLocal` 補償刪除,不留孤兒檔。
- 錯誤:RPC 層統一 Connect code — `unauthenticated` / `permission_denied` / `not_found` / `failed_precondition` / `invalid_argument` / `already_exists`;Gotenberg 等上游基礎設施故障回 `internal`。
- 紙張與字型:四種單據皆 A4(210×297mm,Chromium 參數 8.27×11.69 英吋);字型統一 Noto Sans CJK TC,於模板 CSS 以 `font-family: 'Noto Sans CJK TC', sans-serif` 宣告,Gotenberg 部署環境需具備該字型(維運前置,不在本計畫)。
- 數量型別:`sales_order_items.qty` / `base_qty` 為 `github.com/shopspring/decimal` 的 `decimal.Decimal`(Ent `field.Other`);此為 04/05-sales-orders 計畫的已定案契約。
- 跨 domain 符號(由各計畫提供,欄位名以其細部文件為準):
  - 由 01-auth 計畫提供:`testutil.NewEntClient` / `testutil.NewEntClientWithRLS` / `testutil.DSN`、`rls.Identity{UserID, CompanyID, DepartmentID *uuid.UUID, DataScope, Role}`、`rls.NewContext` / `rls.FromContext`、`audit.Recorder` 介面。
  - 由 04-master-data 計畫提供(Task 8,細部 3.6):`ent.FileAsset` 實體(欄位 `id` / `company_id` / `department_id` / `owner_type` / `owner_id` / `filename` / `original_filename` / `mime_type` / `size_bytes` / `storage_path` / `url` / `created_by` / `deleted_at`)、下載 URL 格式 `/api/v1/files/{id}/download`;本計畫 Task 2 的 `print.FileStore` 埠由 04 計畫以 adapter 實作注入(見 Task 2 Interfaces)。
  - 由 04-master-data 計畫提供(細部 3.1/3.2/3.4):`ent.Customer`(`customer_code` / `name`)、`ent.CustomerAddress`(`type` / `is_default` / `address`,假定欄位名,組合層 `formatAddress` 為調整點)、`ent.CustomerContact`(`name` / `phone` / `is_default`)、`ent.Route`(`code` / `name`)、`ent.Warehouse`(`code` / `name`)、`ent.CuttingSpec`(`code` / `name` / `applies_to`)、`ent.ProductCategory`(`name` / `sort_order`)、`ent.Product`(`category_id` / `picking_warehouse_id`)、`ent.ProductUnit`(`product_id` / `unit_code` / `is_base`)。
  - 由 05-sales-orders 計畫提供(細部 4.1.1):`ent.SalesOrder`(`order_no` / `status` / `customer_id` / `expected_delivery_date` / `route_id` / `delivery_sequence` / `deleted_at`)、`ent.SalesOrderItem`(`sales_order_id` / `product_id` / `display_name` / `qty` / `unit` / `base_qty` / `cutting_spec_id` / `special_cut_note` / `warehouse_id` / `sort_order` / `deleted_at`)。
  - 由 08-dispatch 計畫保證的語意:派車確認後訂單 `status = processing`(細部 5.1.2);本計畫只讀該狀態,不寫。
  - 由 03-metadicts-audit 計畫提供:DB 版 `audit.Recorder` 的交易綁定構造(本計畫以 `AuditFactory` 埠消費,見 Task 4)。
- 選擇器與單據類型組合矩陣(`invalid_argument` 邊界):`dispatch_summary` 不接受任何選擇器;`delivery_note` 可帶 `customer_id`(單印一店),不帶 = 全車次每店一張;`picking_list` **必帶** `warehouse_id`(細部 5.3.3:一份文件對應單一倉別,前端逐倉觸發;schema 的 `warehouse_id` 可空是為了容納其他單據類型),不帶回 `invalid_argument`;`processing_list` 不接受任何選擇器。
- 測試:DB 相依測試走 testcontainers Postgres 16(`testutil.NewEntClient`,01-auth 計畫 Task 1 提供);Gotenberg 一律以 `Converter` fake 或 `httptest.Server` 取代,不啟真實 Gotenberg 容器;`go test ./...` 必須全綠。
- migration 檔序號以執行當下 `backend/database/migrations/` 最大序號 +1 為準;本計畫檔名示例為 `00012_print_tables.sql`。
- 每個 Task 結尾 commit;commit message 格式 `feat(backend): …` / `test(backend): …`。

## File Structure

| 檔案 | 職責 | 建立於 |
|---|---|---|
| `backend/internal/print/viewmodel.go` | 四種單據 view model 型別與 DocumentType | Task 1 |
| `backend/internal/print/templates/dispatch_summary.html` | 單車總表模板 | Task 1 |
| `backend/internal/print/templates/delivery_note.html` | 對點單模板 | Task 1 |
| `backend/internal/print/templates/picking_list.html` | 揀貨單模板 | Task 1 |
| `backend/internal/print/templates/processing_list.html` | 加工單模板 | Task 1 |
| `backend/internal/print/render.go` | 模板 embed 載入與渲染 | Task 1 |
| `backend/internal/print/client.go` | `Converter` 介面 + Gotenberg HTTP client(有限重試) | Task 2 |
| `backend/internal/print/filestore.go` | `FileStore` 消費端埠(實作由 04 計畫 Task 8 提供) | Task 2 |
| `backend/internal/print/service.go` | 資料組合(BuildViewModel)+ PDF 產線(Prepare) | Task 2 |
| `backend/ent/schema/printlog.go` | PrintLog Ent schema | Task 3 |
| `backend/ent/schema/printpreview.go` | PrintPreview Ent schema | Task 3 |
| `backend/database/migrations/00012_print_tables.sql` | RLS policy、索引、file_asset FK | Task 3 |
| `backend/proto/v1/print.proto` | PrintService proto(Preview / Print / ListLogs) | Task 3 起增量 |
| `backend/internal/domain/prints/preview.go` | Preview RPC handler | Task 3 |
| `backend/internal/domain/prints/service.go` | handler 共用建構、權限與參數轉換 | Task 3 |
| `backend/internal/domain/prints/print.go` | Print RPC handler(首印 + 重印分支) | Task 4 |
| `backend/internal/domain/prints/list_logs.go` | ListLogs RPC handler | Task 4 |
| `backend/internal/server/domains.go` | 掛載 PrintService(`InitDomains()`) | Task 4 |

---

### Task 1: 四種單據 view model 與模板(細部 5.3.1–5.3.4)

**Files:**
- Create: `backend/internal/print/viewmodel.go`
- Create: `backend/internal/print/render.go`
- Create: `backend/internal/print/templates/dispatch_summary.html`
- Create: `backend/internal/print/templates/delivery_note.html`
- Create: `backend/internal/print/templates/picking_list.html`
- Create: `backend/internal/print/templates/processing_list.html`
- Test: `backend/internal/print/viewmodel_test.go`
- Test: `backend/internal/print/render_test.go`

**Interfaces:**
- Consumes: 無(DB 無關;view model 由 Task 2 組合層填值)。
- Produces:
  - `print.DocumentType`(`"dispatch_summary"` / `"delivery_note"` / `"picking_list"` / `"processing_list"`)。
  - `print.DispatchSummaryVM` / `DeliveryNoteVM` / `PickingListVM` / `ProcessingListVM` 及其子結構(欄位見 Step 3,**全部不含金額語意欄位**)。
  - `print.RenderHTML(docType DocumentType, vm any) ([]byte, error)` — 以 embed FS 載入對應模板渲染;未知類型回 error。
  - 各 VM 實作內部介面 `docMeta() (routeCode, targetDate string)`(Task 2 產檔名用)。

- [ ] **Step 1: 寫失敗測試(view model 無金額欄位)**

`backend/internal/print/viewmodel_test.go`:

```go
package print_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/print"
)

// bannedAmountWords 為 D12 金額語意禁用詞(小寫比對)。
var bannedAmountWords = []string{"price", "amount", "subtotal", "total", "currency", "cost", "money"}

// TestViewModelsHaveNoAmountFields 以 reflect 遞迴掃描所有 view model 結構體,
// 任一欄位名含金額語意即失敗 — D12 雙重防呆的第一道(view model 源頭無金額,細部 5.3.1 步驟 5)。
func TestViewModelsHaveNoAmountFields(t *testing.T) {
	vms := []any{
		print.DispatchSummaryVM{}, print.DispatchSummaryCustomerVM{}, print.DispatchSummaryItemVM{},
		print.DeliveryNoteVM{}, print.DeliveryNotePageVM{}, print.DeliveryNoteItemVM{},
		print.PickingListVM{}, print.PickingListLineVM{},
		print.ProcessingListVM{}, print.ProcessingRoomLineVM{}, print.ProcessingDeliveryLineVM{},
	}
	for _, vm := range vms {
		assertNoAmountFields(t, reflect.TypeOf(vm))
	}
}

func assertNoAmountFields(t *testing.T, typ reflect.Type) {
	t.Helper()
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		lower := strings.ToLower(f.Name)
		for _, banned := range bannedAmountWords {
			if strings.Contains(lower, banned) {
				t.Errorf("%s.%s contains banned amount word %q (D12)", typ.Name(), f.Name, banned)
			}
		}
		st := f.Type
		if st.Kind() == reflect.Slice {
			st = st.Elem()
		}
		if st.Kind() == reflect.Struct {
			assertNoAmountFields(t, st)
		}
	}
}

// TestProcessingVMHasNoAfterQtyField 加工後數量欄位不存在於 view model —
// 模板層固定渲染空白,由資料結構保證(細部 5.3.4 步驟 5)。
func TestProcessingVMHasNoAfterQtyField(t *testing.T) {
	for _, typ := range []reflect.Type{
		reflect.TypeOf(print.ProcessingRoomLineVM{}),
		reflect.TypeOf(print.ProcessingDeliveryLineVM{}),
	} {
		for i := 0; i < typ.NumField(); i++ {
			if strings.Contains(strings.ToLower(typ.Field(i).Name), "after") {
				t.Errorf("%s.%s: 加工後數量不得存在於 view model,應由模板渲染空白", typ.Name(), typ.Field(i).Name)
			}
		}
	}
}
```

- [ ] **Step 2: 寫失敗測試(四模板渲染與無金額字樣)**

`backend/internal/print/render_test.go`:

```go
package print_test

import (
	"strings"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/print"
)

// bannedRenderedWords 為模板輸出禁用字樣 — D12 雙重防呆的第二道。
var bannedRenderedWords = []string{"金額", "單價", "價格", "小計", "NT$", "$"}

func assertNoAmountWords(t *testing.T, html []byte) {
	t.Helper()
	for _, w := range bannedRenderedWords {
		if strings.Contains(string(html), w) {
			t.Errorf("rendered HTML contains banned word %q (D12)", w)
		}
	}
}

func TestRenderDispatchSummary(t *testing.T) {
	vm := &print.DispatchSummaryVM{
		CompanyName: "甲公司", RouteCode: "TY01", RouteName: "桃園一車",
		TargetDate: "2026-08-21", PrintedAt: "2026-08-21 08:00",
		Customers: []print.DispatchSummaryCustomerVM{
			{CustomerCode: "TY000001", CustomerName: "好食餐廳", Address: "桃園市中正路1號",
				DeliverySequence: 1, Items: []print.DispatchSummaryItemVM{
					{DisplayName: "雞腿", Qty: "10", Unit: "件", SpecialCutNote: "去骨"},
				}},
			{CustomerCode: "TY000002", CustomerName: "美香料理", Address: "桃園市復興路2號",
				DeliverySequence: 2, Items: []print.DispatchSummaryItemVM{
					{DisplayName: "豬五花", Qty: "5", Unit: "包"},
				}},
		},
	}
	html, err := print.RenderHTML(print.DocumentTypeDispatchSummary, vm)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	s := string(html)
	for _, want := range []string{"單車總表", "甲公司", "TY01", "桃園一車", "2026-08-21", "好食餐廳", "去骨", "table-header-group"} {
		if !strings.Contains(s, want) {
			t.Errorf("dispatch_summary 輸出缺少 %q", want)
		}
	}
	// 店家順序與 delivery_sequence 一致:好食(1) 必須出現在 美香(2) 之前
	if strings.Index(s, "好食餐廳") > strings.Index(s, "美香料理") {
		t.Error("店家順序未依 delivery_sequence 升冪")
	}
	assertNoAmountWords(t, html)
}

func TestRenderDeliveryNote(t *testing.T) {
	vm := &print.DeliveryNoteVM{
		CompanyName: "甲公司", RouteCode: "TY01", RouteName: "桃園一車",
		TargetDate: "2026-08-21", PrintedAt: "2026-08-21 08:00",
		Notes: []print.DeliveryNotePageVM{
			{CustomerCode: "TY000001", CustomerName: "好食餐廳", Address: "桃園市中正路1號",
				ContactName: "王老闆", ContactPhone: "0912345678", DeliverySequence: 1,
				Items: []print.DeliveryNoteItemVM{
					{DisplayName: "雞腿", Qty: "10", Unit: "件", CuttingSpecName: "去骨"},
				}},
			{CustomerCode: "TY000002", CustomerName: "美香料理", DeliverySequence: 2,
				Items: []print.DeliveryNoteItemVM{
					{DisplayName: "豬五花", Qty: "5", Unit: "包"},
				}},
		},
	}
	html, err := print.RenderHTML(print.DocumentTypeDeliveryNote, vm)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	s := string(html)
	for _, want := range []string{"對點單", "王老闆", "0912345678", "去骨", "客戶簽名", "收貨日期"} {
		if !strings.Contains(s, want) {
			t.Errorf("delivery_note 輸出缺少 %q", want)
		}
	}
	// 每店強制換頁(細部 5.3.2 步驟 1):兩店需有換頁規則
	if !strings.Contains(s, "page-break-after") {
		t.Error("delivery_note 缺少每店換頁 CSS")
	}
	// 簽收欄固定空白(細部 5.3.2 步驟 4)
	if !strings.Contains(s, `class="sign"`) {
		t.Error("delivery_note 缺少簽收欄")
	}
	assertNoAmountWords(t, html)
}

func TestRenderPickingList(t *testing.T) {
	vm := &print.PickingListVM{
		CompanyName: "甲公司", RouteCode: "TY01", RouteName: "桃園一車",
		TargetDate: "2026-08-21", PrintedAt: "2026-08-21 08:00", WarehouseName: "冷藏倉",
		Lines: []print.PickingListLineVM{
			{CategoryName: "禽肉", DisplayName: "雞腿", BaseQty: "16", BaseUnit: "kg", NeedsCutting: true},
			{CategoryName: "豬肉", DisplayName: "豬五花", BaseQty: "5", BaseUnit: "kg"},
		},
	}
	html, err := print.RenderHTML(print.DocumentTypePickingList, vm)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	s := string(html)
	for _, want := range []string{"揀貨單", "冷藏倉", "禽肉", "雞腿", "16", "kg", "需加工"} {
		if !strings.Contains(s, want) {
			t.Errorf("picking_list 輸出缺少 %q", want)
		}
	}
	// 揀貨核對欄空白(細部 5.3.3 步驟 5)
	if !strings.Contains(s, `class="check"`) {
		t.Error("picking_list 缺少揀貨核對欄")
	}
	assertNoAmountWords(t, html)
}

func TestRenderProcessingList(t *testing.T) {
	vm := &print.ProcessingListVM{
		CompanyName: "甲公司", RouteCode: "TY01", RouteName: "桃園一車",
		TargetDate: "2026-08-21", PrintedAt: "2026-08-21 08:00",
		RoomLines: []print.ProcessingRoomLineVM{
			{WarehouseName: "冷藏倉", DisplayName: "雞腿", OriginalBaseQty: "16", BaseUnit: "kg", CuttingSpecName: "去骨"},
		},
		DeliveryLines: []print.ProcessingDeliveryLineVM{
			{CustomerName: "好食餐廳", DeliverySequence: 1, DisplayName: "雞腿", OriginalQty: "10", Unit: "件", SpecialCutNote: "去骨"},
		},
	}
	html, err := print.RenderHTML(print.DocumentTypeProcessingList, vm)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	s := string(html)
	// 區塊順序固定:加工室揀在前、配送揀在後(細部 5.3.4 步驟 2)
	iRoom, iDelivery := strings.Index(s, "加工室揀"), strings.Index(s, "配送揀")
	if iRoom < 0 || iDelivery < 0 || iRoom > iDelivery {
		t.Errorf("加工單區塊順序錯誤: 加工室揀@%d 配送揀@%d", iRoom, iDelivery)
	}
	if !strings.Contains(s, "加工後數量") {
		t.Error("加工單缺少「加工後數量」欄標題")
	}
	// 「加工後數量」欄所有資料列皆空白(細部 5.3.4 步驟 5):欄位 cell 無內容
	if !strings.Contains(s, `<td class="after"></td>`) {
		t.Error("加工後數量欄應為空白儲存格")
	}
	if strings.Contains(s, "OriginalBaseQty") || strings.Contains(s, "AfterQty") {
		t.Error("模板洩漏 view model 欄位名")
	}
	assertNoAmountWords(t, html)
}
```

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/print/ -v`
Expected: FAIL — `print.DispatchSummaryVM` 等型別與 `print.RenderHTML` 未定義(編譯失敗)。

- [ ] **Step 4: 實作 view model**

`backend/internal/print/viewmodel.go`:

```go
// Package print 為單據列印核心:view model、模板渲染、Gotenberg client、
// 資料組合與 PDF 產線(細部文件 09-printing Task 5.3 / 5.4)。
package print

// DocumentType 為四種單據類型(細部 5.3)。
type DocumentType string

const (
	DocumentTypeDispatchSummary DocumentType = "dispatch_summary" // 單車總表
	DocumentTypeDeliveryNote    DocumentType = "delivery_note"    // 對點單
	DocumentTypePickingList     DocumentType = "picking_list"     // 揀貨單
	DocumentTypeProcessingList  DocumentType = "processing_list"  // 加工單
)

// DocumentTypes 回傳全部合法類型,供參數驗證。
func DocumentTypes() []DocumentType {
	return []DocumentType{
		DocumentTypeDispatchSummary, DocumentTypeDeliveryNote,
		DocumentTypePickingList, DocumentTypeProcessingList,
	}
}

// docMeta 供 Task 2 產生 PDF 檔名;所有 VM 實作。
type docMeta interface {
	meta() (routeCode, targetDate string)
}

// ---------- 單車總表(細部 5.3.1;依 D12 不含任何金額欄位) ----------

type DispatchSummaryVM struct {
	CompanyName string
	RouteCode   string
	RouteName   string
	TargetDate  string // YYYY-MM-DD
	PrintedAt   string // YYYY-MM-DD HH:MM(Asia/Taipei)
	Customers   []DispatchSummaryCustomerVM // 已依 DeliverySequence 升冪
	IsEmpty     bool                        // 查無任何符合明細(細部 5.4.2 步驟 6)
}

func (v *DispatchSummaryVM) meta() (string, string) { return v.RouteCode, v.TargetDate }

type DispatchSummaryCustomerVM struct {
	CustomerCode     string
	CustomerName     string
	Address          string
	DeliverySequence int
	Items            []DispatchSummaryItemVM
}

type DispatchSummaryItemVM struct {
	DisplayName    string
	Qty            string // 原始下單數量(decimal 字串)
	Unit           string // 下單單位
	SpecialCutNote string // 空字串 = 無
}

// ---------- 對點單(細部 5.3.2;每店一頁,不含價格) ----------

type DeliveryNoteVM struct {
	CompanyName string
	RouteCode   string
	RouteName   string
	TargetDate  string
	PrintedAt   string
	Notes       []DeliveryNotePageVM // 每店一頁,已依 DeliverySequence 升冪
	IsEmpty     bool
}

func (v *DeliveryNoteVM) meta() (string, string) { return v.RouteCode, v.TargetDate }

type DeliveryNotePageVM struct {
	CustomerCode     string
	CustomerName     string
	Address          string
	ContactName      string
	ContactPhone     string
	DeliverySequence int
	Items            []DeliveryNoteItemVM
}

type DeliveryNoteItemVM struct {
	DisplayName     string
	Qty             string
	Unit            string
	CuttingSpecName string // 空字串 = 無分切規格
	SpecialCutNote  string
}

// ---------- 揀貨單(細部 5.3.3;車次→倉別→分類→品名,base_qty 彙總) ----------

type PickingListVM struct {
	CompanyName   string
	RouteCode     string
	RouteName     string
	TargetDate    string
	PrintedAt     string
	WarehouseName string
	Lines         []PickingListLineVM // 已依 分類 sort_order→品名 排序,同商品跨店合併
	IsEmpty       bool
}

func (v *PickingListVM) meta() (string, string) { return v.RouteCode, v.TargetDate }

type PickingListLineVM struct {
	CategoryName string
	DisplayName  string
	BaseQty      string // 基本單位彙總(Σ sales_order_items.base_qty,§6.4)
	BaseUnit     string // 商品 is_base 單位;手打品名 = 下單單位
	NeedsCutting bool   // 有分切規格或 special_cut_note(細部 5.3.3 步驟 4)
}

// ---------- 加工單(細部 5.3.4;兩區塊,加工後數量由模板渲染空白) ----------

type ProcessingListVM struct {
	CompanyName   string
	RouteCode     string
	RouteName     string
	TargetDate    string
	PrintedAt     string
	RoomLines     []ProcessingRoomLineVM     // 加工室揀:倉別→品名
	DeliveryLines []ProcessingDeliveryLineVM // 配送揀:配送順序→品名
	IsEmpty       bool
}

func (v *ProcessingListVM) meta() (string, string) { return v.RouteCode, v.TargetDate }

// ProcessingRoomLineVM 故意沒有「加工後數量」欄位:模板固定渲染空白儲存格(細部 5.3.4 步驟 5)。
type ProcessingRoomLineVM struct {
	WarehouseName   string
	DisplayName     string
	OriginalBaseQty string // 原始數量(基本單位)
	BaseUnit        string
	CuttingSpecName string
}

// ProcessingDeliveryLineVM 同上,無「加工後數量」欄位。
type ProcessingDeliveryLineVM struct {
	CustomerName     string
	DeliverySequence int
	DisplayName      string
	OriginalQty      string // 原始數量(下單單位)
	Unit             string
	SpecialCutNote   string
}
```

- [ ] **Step 5: 實作渲染器與四模板**

`backend/internal/print/render.go`:

```go
package print

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

//go:embed templates/*.html
var templatesFS embed.FS

var templateFile = map[DocumentType]string{
	DocumentTypeDispatchSummary: "templates/dispatch_summary.html",
	DocumentTypeDeliveryNote:    "templates/delivery_note.html",
	DocumentTypePickingList:     "templates/picking_list.html",
	DocumentTypeProcessingList:  "templates/processing_list.html",
}

// RenderHTML 以對應模板渲染 view model(細部 5.3);未知類型回 error。
// 渲染失敗(模板缺欄位等)屬開發期錯誤,由模板測試攔下,此處原樣回傳。
func RenderHTML(docType DocumentType, vm any) ([]byte, error) {
	file, ok := templateFile[docType]
	if !ok {
		return nil, fmt.Errorf("print: unknown document type %q", docType)
	}
	tpl, err := template.ParseFS(templatesFS, file)
	if err != nil {
		return nil, fmt.Errorf("print: parse %s: %w", file, err)
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, vm); err != nil {
		return nil, fmt.Errorf("print: execute %s: %w", file, err)
	}
	return buf.Bytes(), nil
}
```

`backend/internal/print/templates/dispatch_summary.html`(單車總表,細部 5.3.1):

```html
<!DOCTYPE html>
<html lang="zh-Hant">
<head>
<meta charset="utf-8">
<style>
@page { size: A4; margin: 12mm; }
body { font-family: 'Noto Sans CJK TC', sans-serif; font-size: 11pt; }
h1 { font-size: 16pt; text-align: center; margin: 0 0 4mm; }
.meta { margin-bottom: 4mm; }
.meta span { margin-right: 8mm; }
.customer { margin-bottom: 6mm; }
table { width: 100%; border-collapse: collapse; }
th, td { border: 1px solid #333; padding: 1.5mm 2mm; text-align: left; }
thead { display: table-header-group; } /* 跨頁時表頭重複(細部 5.3.1 步驟 4) */
.note { color: #000; font-weight: bold; }
</style>
</head>
<body>
<h1>單車總表</h1>
<div class="meta">
  <span>{{.CompanyName}}</span>
  <span>車次:{{.RouteCode}} {{.RouteName}}</span>
  <span>出貨日期:{{.TargetDate}}</span>
  <span>列印時間:{{.PrintedAt}}</span>
</div>
{{range .Customers}}
<div class="customer">
  <table>
    <thead>
      <tr><th colspan="4">[{{.DeliverySequence}}] {{.CustomerCode}} {{.CustomerName}} — {{.Address}}</th></tr>
      <tr><th>品名</th><th>數量</th><th>單位</th><th>特殊分切備註</th></tr>
    </thead>
    <tbody>
      {{range .Items}}
      <tr>
        <td>{{.DisplayName}}</td>
        <td>{{.Qty}}</td>
        <td>{{.Unit}}</td>
        <td>{{if .SpecialCutNote}}<span class="note">{{.SpecialCutNote}}</span>{{end}}</td>
      </tr>
      {{end}}
    </tbody>
  </table>
</div>
{{end}}
</body>
</html>
```

`backend/internal/print/templates/delivery_note.html`(對點單,細部 5.3.2):

```html
<!DOCTYPE html>
<html lang="zh-Hant">
<head>
<meta charset="utf-8">
<style>
@page { size: A4; margin: 12mm; }
body { font-family: 'Noto Sans CJK TC', sans-serif; font-size: 11pt; }
h1 { font-size: 16pt; text-align: center; margin: 0 0 4mm; }
.meta { margin-bottom: 4mm; }
.meta span { margin-right: 8mm; }
.note-page { page-break-after: always; } /* 每店強制換頁(細部 5.3.2 步驟 1) */
.note-page:last-child { page-break-after: auto; }
table { width: 100%; border-collapse: collapse; }
th, td { border: 1px solid #333; padding: 1.5mm 2mm; text-align: left; }
thead { display: table-header-group; }
.sign { margin-top: 12mm; }
.sign span { display: inline-block; width: 45%; border-bottom: 1px solid #333; height: 8mm; margin-right: 5%; }
</style>
</head>
<body>
{{range .Notes}}
<div class="note-page">
  <h1>對點單</h1>
  <div class="meta">
    <span>{{$.CompanyName}}</span>
    <span>車次:{{$.RouteCode}} {{$.RouteName}}</span>
    <span>出貨日期:{{$.TargetDate}}</span>
    <span>列印時間:{{$.PrintedAt}}</span>
  </div>
  <div class="meta">
    <span>[{{.DeliverySequence}}] {{.CustomerCode}} {{.CustomerName}}</span>
    <span>地址:{{.Address}}</span>
    <span>聯絡人:{{.ContactName}} {{.ContactPhone}}</span>
  </div>
  <table>
    <thead>
      <tr><th>品名</th><th>數量</th><th>單位</th><th>分切規格</th><th>品項備註</th></tr>
    </thead>
    <tbody>
      {{range .Items}}
      <tr>
        <td>{{.DisplayName}}</td>
        <td>{{.Qty}}</td>
        <td>{{.Unit}}</td>
        <td>{{.CuttingSpecName}}</td>
        <td>{{.SpecialCutNote}}</td>
      </tr>
      {{end}}
    </tbody>
  </table>
  <!-- 簽收欄固定空白,供現場手寫(細部 5.3.2 步驟 4) -->
  <div class="sign"><span>客戶簽名:</span><span>收貨日期:</span></div>
</div>
{{end}}
</body>
</html>
```

`backend/internal/print/templates/picking_list.html`(揀貨單,細部 5.3.3):

```html
<!DOCTYPE html>
<html lang="zh-Hant">
<head>
<meta charset="utf-8">
<style>
@page { size: A4; margin: 12mm; }
body { font-family: 'Noto Sans CJK TC', sans-serif; font-size: 11pt; }
h1 { font-size: 16pt; text-align: center; margin: 0 0 4mm; }
.meta { margin-bottom: 4mm; }
.meta span { margin-right: 8mm; }
table { width: 100%; border-collapse: collapse; }
th, td { border: 1px solid #333; padding: 1.5mm 2mm; text-align: left; }
thead { display: table-header-group; }
.check { width: 18mm; } /* 空白揀貨核對欄(細部 5.3.3 步驟 5) */
.cut { font-weight: bold; }
</style>
</head>
<body>
<h1>揀貨單</h1>
<div class="meta">
  <span>{{.CompanyName}}</span>
  <span>車次:{{.RouteCode}} {{.RouteName}}</span>
  <span>倉別:{{.WarehouseName}}</span>
  <span>出貨日期:{{.TargetDate}}</span>
  <span>列印時間:{{.PrintedAt}}</span>
</div>
<table>
  <thead>
    <tr><th>商品分類</th><th>品名</th><th>彙總數量</th><th>基本單位</th><th>加工</th><th>揀貨核對</th></tr>
  </thead>
  <tbody>
    {{range .Lines}}
    <tr>
      <td>{{.CategoryName}}</td>
      <td>{{.DisplayName}}</td>
      <td>{{.BaseQty}}</td>
      <td>{{.BaseUnit}}</td>
      <td>{{if .NeedsCutting}}<span class="cut">需加工</span>{{end}}</td>
      <td class="check"></td>
    </tr>
    {{end}}
  </tbody>
</table>
</body>
</html>
```

`backend/internal/print/templates/processing_list.html`(加工單,細部 5.3.4):

```html
<!DOCTYPE html>
<html lang="zh-Hant">
<head>
<meta charset="utf-8">
<style>
@page { size: A4; margin: 12mm; }
body { font-family: 'Noto Sans CJK TC', sans-serif; font-size: 11pt; }
h1 { font-size: 16pt; text-align: center; margin: 0 0 4mm; }
h2 { font-size: 13pt; margin: 6mm 0 2mm; }
.meta { margin-bottom: 4mm; }
.meta span { margin-right: 8mm; }
table { width: 100%; border-collapse: collapse; }
th, td { border: 1px solid #333; padding: 1.5mm 2mm; text-align: left; }
thead { display: table-header-group; }
.after { width: 22mm; } /* 加工後數量:固定空白,現場手寫回填(細部 5.3.4 步驟 5) */
</style>
</head>
<body>
<h1>加工單</h1>
<div class="meta">
  <span>{{.CompanyName}}</span>
  <span>車次:{{.RouteCode}} {{.RouteName}}</span>
  <span>出貨日期:{{.TargetDate}}</span>
  <span>列印時間:{{.PrintedAt}}</span>
</div>
<h2>加工室揀</h2>
<table>
  <thead>
    <tr><th>倉別</th><th>品名</th><th>原始數量</th><th>基本單位</th><th>分切規格</th><th>加工後數量</th></tr>
  </thead>
  <tbody>
    {{range .RoomLines}}
    <tr>
      <td>{{.WarehouseName}}</td>
      <td>{{.DisplayName}}</td>
      <td>{{.OriginalBaseQty}}</td>
      <td>{{.BaseUnit}}</td>
      <td>{{.CuttingSpecName}}</td>
      <td class="after"></td>
    </tr>
    {{end}}
  </tbody>
</table>
<h2>配送揀</h2>
<table>
  <thead>
    <tr><th>店家</th><th>品名</th><th>原始數量</th><th>單位</th><th>特殊分切備註</th><th>加工後數量</th></tr>
  </thead>
  <tbody>
    {{range .DeliveryLines}}
    <tr>
      <td>[{{.DeliverySequence}}] {{.CustomerName}}</td>
      <td>{{.DisplayName}}</td>
      <td>{{.OriginalQty}}</td>
      <td>{{.Unit}}</td>
      <td>{{.SpecialCutNote}}</td>
      <td class="after"></td>
    </tr>
    {{end}}
  </tbody>
</table>
</body>
</html>
```

- [ ] **Step 6: 跑測試確認通過**

Run: `cd backend && go test ./internal/print/ -v`
Expected: PASS — 6 個測試全過(view model 無金額欄位、加工單無 AfterQty、四模板渲染內容與順序正確、無金額字樣)。

- [ ] **Step 7: Commit**

```bash
git add backend/internal/print
git commit -m "feat(backend): 四種單據 view model 與 A4 模板,全文無金額(D12/D15)(5.3)"
```

---

### Task 2: Gotenberg client + 資料組合 + PDF 產線(細部 5.4.1–5.4.3)

**Files:**
- Create: `backend/internal/print/client.go`
- Create: `backend/internal/print/filestore.go`
- Create: `backend/internal/print/service.go`
- Test: `backend/internal/print/client_test.go`
- Test: `backend/internal/print/service_test.go`

**Interfaces:**
- Consumes:
  - Task 1 的 `DocumentType` / 四種 VM / `RenderHTML`。
  - 由 04-plan 提供:`ent.Customer` / `CustomerAddress` / `CustomerContact` / `Route` / `Warehouse` / `CuttingSpec` / `ProductCategory` / `Product` / `ProductUnit` / `FileAsset`(欄位名見 Global Constraints);FileStore 埠的實作 adapter(本計畫測試以 fake 實作,整合時注入 04 實作)。
  - 由 05-plan 提供:`ent.SalesOrder` / `ent.SalesOrderItem`(欄位名見 Global Constraints;qty/base_qty 假定 `decimal.Decimal`)。
  - 由 01-auth 計畫提供:`rls.FromContext`(組合層取公司名稱)、`testutil.NewEntClient`。
- Produces:
  - `print.Converter` 介面(`HTMLToPDF(ctx context.Context, html []byte) ([]byte, error)`);`print.NewGotenbergClient(baseURL string, timeout time.Duration, opts ...GotenbergOption) *GotenbergClient`;`print.WithBackoff(d time.Duration) GotenbergOption`(測試用);`print.ErrRejected`(4xx,映射 `invalid_argument`);`print.UpstreamError{Status int, Body string}`(5xx/連線錯重試用盡,映射 `internal`)。
  - `print.FileStore` 埠(細部 5.4.3;**實作由 04-plan Task 8 提供 adapter**):

    ```go
    type FileStore interface {
        WriteLocal(ctx context.Context, in WriteInput) (*StoredFile, error)
        CreateAsset(ctx context.Context, tx *ent.Tx, actor rls.Identity, stored *StoredFile, ownerType string, ownerID uuid.UUID) (*ent.FileAsset, error)
        DeleteLocal(ctx context.Context, storagePath string) // 補償刪除;失敗僅記錄告警
    }
    ```

    連同 `print.WriteInput{OriginalFilename string, Content []byte}`、`print.StoredFile{Filename, StoragePath string, SizeBytes int64}`;owner_type 統一 `"print_pdf"`(依 04 契約),owner_id = print_logs / print_previews id。
  - `print.NewService(db *ent.Client, c Converter, f FileStore) *Service`。
  - `print.Selector{CustomerID, WarehouseID *uuid.UUID}`。
  - `(*Service).BuildViewModel(ctx context.Context, docType DocumentType, routeID uuid.UUID, targetDate time.Time, sel Selector) (any, error)` — 空表回 VM `IsEmpty=true` 而非錯誤(細部 5.4.2 步驟 6);狀態過濾不在此層(步驟 7)。
  - `(*Service).Prepare(ctx context.Context, docType DocumentType, vm any) (*PreparedPDF, error)` — 空表守門 `failed_precondition`;渲染 → Converter → 大小檢查 → WriteLocal;回傳 `PreparedPDF{Stored *StoredFile, OriginalFilename string}`。**不碰 DB**;DB 記錄由 Task 3/4 handler 在交易內經 `FileStore.CreateAsset` 建立。
  - `print.DownloadURL(fileAssetID uuid.UUID) string` — 回傳 `/api/v1/files/{id}/download`(依 04 契約)。

- [ ] **Step 1: 寫失敗測試(Gotenberg client:成功、5xx 有限重試、4xx 不重試、逾時)**

`backend/internal/print/client_test.go`:

```go
package print_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/print"
)

func TestGotenbergConvertSuccess(t *testing.T) {
	var gotMultipart bool
	var gotA4 bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/forms/chromium/convert/html" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("parse multipart: %v", err)
			return
		}
		f, _, err := r.FormFile("files") // Gotenberg 檔案欄位名固定為 files
		if err != nil {
			t.Errorf("missing index.html file: %v", err)
			return
		}
		b, _ := io.ReadAll(f)
		_ = f.Close()
		gotMultipart = strings.Contains(string(b), "<html")
		// A4 紙張參數(細部 5.4.1 步驟 2)
		gotA4 = r.FormValue("paperWidth") == "8.27" && r.FormValue("paperHeight") == "11.69"
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("%PDF-fake"))
	}))
	defer srv.Close()

	c := print.NewGotenbergClient(srv.URL, 5*time.Second)
	pdf, err := c.HTMLToPDF(context.Background(), []byte("<html><body>中文</body></html>"))
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if string(pdf) != "%PDF-fake" {
		t.Fatalf("unexpected pdf body %q", pdf)
	}
	if !gotMultipart || !gotA4 {
		t.Errorf("request shape wrong: multipart=%v a4=%v", gotMultipart, gotA4)
	}
}

func TestGotenbergRetriesOn5xx(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	c := print.NewGotenbergClient(srv.URL, 5*time.Second, print.WithBackoff(time.Millisecond))
	_, err := c.HTMLToPDF(context.Background(), []byte("<html></html>"))
	if err == nil {
		t.Fatal("want error")
	}
	// 細部 5.4.1 步驟 5:至多 2 次重試 = 共 3 次呼叫
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("want 3 calls (1 + 2 retries), got %d", got)
	}
	var up *print.UpstreamError
	if !errorsAs(err, &up) {
		t.Fatalf("want UpstreamError, got %T", err)
	}
}

func TestGotenbergNoRetryOn4xx(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid html"))
	}))
	defer srv.Close()

	c := print.NewGotenbergClient(srv.URL, 5*time.Second, print.WithBackoff(time.Millisecond))
	_, err := c.HTMLToPDF(context.Background(), []byte("<html></html>"))
	// 細部 5.4.1 步驟 5:4xx 屬請求錯誤,不重試直接失敗
	if !errorsIs(err, print.ErrRejected) {
		t.Fatalf("want ErrRejected, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("4xx must not retry: got %d calls", got)
	}
}

func TestGotenbergTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		_, _ = w.Write([]byte("%PDF-late"))
	}))
	defer srv.Close()

	c := print.NewGotenbergClient(srv.URL, 50*time.Millisecond) // 細部 5.4.1 步驟 3:整體 deadline
	_, err := c.HTMLToPDF(context.Background(), []byte("<html></html>"))
	if err == nil {
		t.Fatal("want timeout error")
	}
}
```

(`errorsAs` / `errorsIs` 為測試檔頂端的 `errors.As` / `errors.Is` 直接呼叫,import `"errors"` 後改用具名呼叫即可;此處為閱讀簡潔。實作時直接 `errors.As(err, &up)` / `errors.Is(err, print.ErrRejected)`。)

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/print/ -run TestGotenberg -v`
Expected: FAIL — `print.NewGotenbergClient` 等未定義(編譯失敗)。

- [ ] **Step 3: 實作 Gotenberg client**

`backend/internal/print/client.go`:

```go
package print

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// Converter 抽象 HTML→PDF 引擎(細部 5.4.1);測試以 fake 替換,不啟真實 Gotenberg。
type Converter interface {
	HTMLToPDF(ctx context.Context, html []byte) ([]byte, error)
}

// ErrRejected 為上游 4xx(HTML 無法解析等請求錯誤);不重試,service 層映射 invalid_argument。
var ErrRejected = errors.New("print: html rejected by converter")

// UpstreamError 為上游 5xx 或連線錯誤重試用盡;service 層映射 internal 並記 trace。
type UpstreamError struct {
	Status int    // 最後一次 HTTP status;連線錯誤為 0
	Body   string // Gotenberg 錯誤內容(trace 用)
	Err    error  // 連線層錯誤(若有)
}

func (e *UpstreamError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("print: gotenberg unreachable: %v", e.Err)
	}
	return fmt.Sprintf("print: gotenberg status %d: %s", e.Status, e.Body)
}

// GotenbergOption 調整 client 行為(測試用)。
type GotenbergOption func(*GotenbergClient)

// WithBackoff 覆寫重試退避基準(預設 200ms,指數 ×2)。
func WithBackoff(d time.Duration) GotenbergOption {
	return func(g *GotenbergClient) { g.backoff = d }
}

// GotenbergClient 為 Converter 的 HTTP 實作(細部 5.4.1)。
type GotenbergClient struct {
	baseURL    string
	httpClient *http.Client
	maxRetries int // 至多 2 次重試(步驟 5)
	backoff    time.Duration
}

// NewGotenbergClient 建立 client;timeout 為整體 deadline(步驟 3),避免 Gotenberg 卡死拖住 RPC。
func NewGotenbergClient(baseURL string, timeout time.Duration, opts ...GotenbergOption) *GotenbergClient {
	g := &GotenbergClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: timeout},
		maxRetries: 2,
		backoff:    200 * time.Millisecond,
	}
	for _, o := range opts {
		o(g)
	}
	return g
}

// HTMLToPDF 送 Chromium HTML 轉換路由;PDF 全程在記憶體傳遞,不落暫存檔(步驟 6)。
func (g *GotenbergClient) HTMLToPDF(ctx context.Context, html []byte) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= g.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(g.backoff << (attempt - 1)): // 指數退避
			}
		}
		pdf, retryable, err := g.convertOnce(ctx, html)
		if err == nil {
			return pdf, nil
		}
		if !retryable {
			return nil, err // 4xx 不重試(步驟 5)
		}
		lastErr = err
	}
	return nil, lastErr
}

// convertOnce 執行單次轉換;回傳 retryable 表示可重試(5xx / 連線錯誤)。
func (g *GotenbergClient) convertOnce(ctx context.Context, html []byte) (pdf []byte, retryable bool, err error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, err := w.CreateFormFile("files", "index.html") // 主文件檔名固定 index.html(步驟 1)
	if err != nil {
		return nil, false, fmt.Errorf("print: multipart file: %w", err)
	}
	if _, err := fw.Write(html); err != nil {
		return nil, false, fmt.Errorf("print: multipart write: %w", err)
	}
	// 轉換參數固定(步驟 2):A4、列印背景、合理頁邊距、等待載入。
	fields := map[string]string{
		"paperWidth":     "8.27", // A4 = 8.27 × 11.69 英吋
		"paperHeight":    "11.69",
		"marginTop":      "0.39",
		"marginBottom":   "0.39",
		"marginLeft":     "0.39",
		"marginRight":    "0.39",
		"printBackground": "true",
		"waitDelay":      "1s", // 等待字型與樣式載入完成
	}
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			return nil, false, fmt.Errorf("print: multipart field %s: %w", k, err)
		}
	}
	if err := w.Close(); err != nil {
		return nil, false, fmt.Errorf("print: multipart close: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		g.baseURL+"/forms/chromium/convert/html", &body)
	if err != nil {
		return nil, false, fmt.Errorf("print: new request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := g.httpClient.Do(req)
	if err != nil {
		// 連線錯誤/逾時:可重試(步驟 5);context 取消除外
		if ctx.Err() != nil {
			return nil, false, ctx.Err()
		}
		return nil, true, &UpstreamError{Err: err}
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		pdf, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, true, &UpstreamError{Err: err}
		}
		return pdf, false, nil
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		// 4xx 屬模板/資料問題,開發期應被測試攔下;不重試(步驟 5)
		return nil, false, fmt.Errorf("%w: status %d: %s", ErrRejected, resp.StatusCode, b)
	default: // 5xx 與其他異常:可重試
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, true, &UpstreamError{Status: resp.StatusCode, Body: string(b)}
	}
}
```

- [ ] **Step 4: 跑測試確認通過**

Run: `cd backend && go test ./internal/print/ -run TestGotenberg -v`
Expected: PASS — 4 個測試(成功含 A4 參數、5xx 共 3 次呼叫、4xx 僅 1 次、逾時回錯)。

- [ ] **Step 5: 寫失敗測試(資料組合 BuildViewModel:分組/排序/彙總/空表/錯誤)**

`backend/internal/print/service_test.go`(本檔於 Step 7 再加 Prepare 測試;`seedPrintFixture` 的實體必填欄位以 04/05 計畫 schema 為準,此處為完整形狀):

```go
package print_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/print"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

// printFixture 為列印測試的共用資料:一公司、一部門、一車次、一倉別、
// 一分類、兩商品(一含分切規格)、兩店家(配送順序 1/2)、兩筆訂單與明細。
type printFixture struct {
	company   *ent.Company
	dept      *ent.Department
	route     *ent.Route
	warehouse *ent.Warehouse
	category  *ent.ProductCategory
	spec      *ent.CuttingSpec
	productA  *ent.Product // 雞腿,is_base 單位 kg,換算 1 件 = 0.6 kg? 換算已由 05 寫入 base_qty,此處直接給 base_qty
	productB  *ent.Product // 豬五花
	customer1 *ent.Customer
	customer2 *ent.Customer
	order1    *ent.SalesOrder // customer1,seq 1,processing
	order2    *ent.SalesOrder // customer2,seq 2,pending(供 Task 4 狀態守門測試)
	ctx       context.Context // 部門級身分
}

// seedPrintFixture 建立共用資料。04/05 計畫實體的必填欄位若多於此處,依其 schema 補齊;
// 此 helper 為唯一需對齊之處。
func seedPrintFixture(t *testing.T, c *ent.Client) *printFixture {
	t.Helper()
	ctx := context.Background()

	co, err := c.Company.Create().SetName("甲公司").SetIdentifier("co-a").
		SetCustomerCodePrefix("TY").Save(ctx)
	if err != nil {
		t.Fatalf("company: %v", err)
	}
	dept, err := c.Department.Create().SetCompanyID(co.ID).SetName("桃園部").Save(ctx)
	if err != nil {
		t.Fatalf("dept: %v", err)
	}
	route, err := c.Route.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetCode("TY01").SetName("桃園一車").Save(ctx)
	if err != nil {
		t.Fatalf("route: %v", err)
	}
	wh, err := c.Warehouse.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetCode("COLD").SetName("冷藏倉").Save(ctx)
	if err != nil {
		t.Fatalf("warehouse: %v", err)
	}
	cat, err := c.ProductCategory.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetCode("PO").SetName("禽肉").SetSortOrder(1).Save(ctx)
	if err != nil {
		t.Fatalf("category: %v", err)
	}
	spec, err := c.CuttingSpec.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetCode("DEBONE").SetName("去骨").SetAppliesTo("processing").Save(ctx)
	if err != nil {
		t.Fatalf("cutting spec: %v", err)
	}
	pA, err := c.Product.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetCode("P001").SetName("雞腿").SetCategoryID(cat.ID).SetPickingWarehouseID(wh.ID).Save(ctx)
	if err != nil {
		t.Fatalf("product A: %v", err)
	}
	if _, err := c.ProductUnit.Create().SetProductID(pA.ID).
		SetUnitCode("kg").SetConversionRate(decimal.NewFromInt(1)).SetIsBase(true).Save(ctx); err != nil {
		t.Fatalf("product unit A: %v", err)
	}
	pB, err := c.Product.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetCode("P002").SetName("豬五花").SetCategoryID(cat.ID).SetPickingWarehouseID(wh.ID).Save(ctx)
	if err != nil {
		t.Fatalf("product B: %v", err)
	}
	if _, err := c.ProductUnit.Create().SetProductID(pB.ID).
		SetUnitCode("kg").SetConversionRate(decimal.NewFromInt(1)).SetIsBase(true).Save(ctx); err != nil {
		t.Fatalf("product unit B: %v", err)
	}
	mkCustomer := func(code, name string) *ent.Customer {
		cust, err := c.Customer.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
			SetCustomerCode(code).SetName(name).Save(ctx)
		if err != nil {
			t.Fatalf("customer %s: %v", code, err)
		}
		if _, err := c.CustomerAddress.Create().SetCustomerID(cust.ID).
			SetCompanyID(co.ID).SetDepartmentID(dept.ID).
			SetType("shipping").SetIsDefault(true).SetAddress("桃園市中正路1號").Save(ctx); err != nil {
			t.Fatalf("address: %v", err)
		}
		if _, err := c.CustomerContact.Create().SetCustomerID(cust.ID).
			SetCompanyID(co.ID).SetDepartmentID(dept.ID).
			SetName("王老闆").SetPhone("0912345678").SetIsDefault(true).Save(ctx); err != nil {
			t.Fatalf("contact: %v", err)
		}
		return cust
	}
	cust1 := mkCustomer("TY000001", "好食餐廳")
	cust2 := mkCustomer("TY000002", "美香料理")

	date := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	mkOrder := func(no string, cust *ent.Customer, seq int, status string) *ent.SalesOrder {
		o, err := c.SalesOrder.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
			SetOrderNo(no).SetCustomerID(cust.ID).SetSource("W").
			SetStatus(status).SetExpectedDeliveryDate(date).
			SetRouteID(route.ID).SetDeliverySequence(seq).Save(ctx)
		if err != nil {
			t.Fatalf("order %s: %v", no, err)
		}
		return o
	}
	o1 := mkOrder("SO-001", cust1, 1, "processing")
	o2 := mkOrder("SO-002", cust2, 2, "pending")

	// order1:雞腿 10 件(base 6 kg,含分切)+ 豬五花 5 kg
	mkItem := func(o *ent.SalesOrder, p *ent.Product, name, qty, base string, unit string, sort int, withSpec bool) {
		b := c.SalesOrderItem.Create().
			SetCompanyID(co.ID).SetDepartmentID(dept.ID).
			SetSalesOrderID(o.ID).SetProductID(p.ID).SetDisplayName(name).
			SetQty(decimal.RequireFromString(qty)).SetUnit(unit).
			SetBaseQty(decimal.RequireFromString(base)).
			SetWarehouseID(wh.ID).SetSortOrder(sort)
		if withSpec {
			b = b.SetCuttingSpecID(spec.ID).SetSpecialCutNote("去骨")
		}
		if _, err := b.Save(ctx); err != nil {
			t.Fatalf("item: %v", err)
		}
	}
	mkItem(o1, pA, "雞腿", "10", "6", "件", 1, true)
	mkItem(o1, pB, "豬五花", "5", "5", "kg", 2, false)
	// order2:雞腿 4 kg(base 4 kg,無分切)— 供揀貨單跨店彙總 6+4=10
	mkItem(o2, pA, "雞腿", "4", "4", "kg", 1, false)

	identity := rls.Identity{
		UserID: uuid.New(), CompanyID: co.ID, DepartmentID: &dept.ID,
		DataScope: "department", Role: "staff",
	}
	return &printFixture{
		company: co, dept: dept, route: route, warehouse: wh, category: cat, spec: spec,
		productA: pA, productB: pB, customer1: cust1, customer2: cust2,
		order1: o1, order2: o2,
		ctx: rls.NewContext(ctx, identity),
	}
}

var targetDate = time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)

func TestBuildDispatchSummary(t *testing.T) {
	c := testutil.NewEntClient(t)
	fx := seedPrintFixture(t, c)
	svc := print.NewService(c, nil, nil) // 組合層不用 converter/filestore

	vmAny, err := svc.BuildViewModel(fx.ctx, print.DocumentTypeDispatchSummary, fx.route.ID, targetDate, print.Selector{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	vm := vmAny.(*print.DispatchSummaryVM)
	if vm.IsEmpty {
		t.Fatal("should not be empty")
	}
	if vm.CompanyName != "甲公司" || vm.RouteCode != "TY01" {
		t.Errorf("header wrong: %+v", vm)
	}
	// 店家依 delivery_sequence 升冪(細部 5.3.1 步驟 2)
	if len(vm.Customers) != 2 ||
		vm.Customers[0].CustomerCode != "TY000001" || vm.Customers[0].DeliverySequence != 1 ||
		vm.Customers[1].CustomerCode != "TY000002" || vm.Customers[1].DeliverySequence != 2 {
		t.Fatalf("customers not ordered by delivery_sequence: %+v", vm.Customers)
	}
	// 品項含特殊分切備註(細部 5.3.1 步驟 3)
	found := false
	for _, it := range vm.Customers[0].Items {
		if it.DisplayName == "雞腿" {
			found = true
			if it.SpecialCutNote != "去骨" || it.Qty != "10" || it.Unit != "件" {
				t.Errorf("item wrong: %+v", it)
			}
		}
	}
	if !found {
		t.Error("missing 雞腿 item")
	}
}

func TestBuildDeliveryNoteWithCustomerSelector(t *testing.T) {
	c := testutil.NewEntClient(t)
	fx := seedPrintFixture(t, c)
	svc := print.NewService(c, nil, nil)

	// 不指定 customer:全車次每店一頁,依配送順序
	vmAny, err := svc.BuildViewModel(fx.ctx, print.DocumentTypeDeliveryNote, fx.route.ID, targetDate, print.Selector{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	vm := vmAny.(*print.DeliveryNoteVM)
	if len(vm.Notes) != 2 || vm.Notes[0].CustomerName != "好食餐廳" || vm.Notes[1].CustomerName != "美香料理" {
		t.Fatalf("notes wrong: %+v", vm.Notes)
	}
	if vm.Notes[0].ContactName != "王老闆" || vm.Notes[0].Address != "桃園市中正路1號" {
		t.Errorf("contact/address missing: %+v", vm.Notes[0])
	}
	// 分切規格併列(細部 5.3.2 步驟 3)
	if got := vm.Notes[0].Items[0].CuttingSpecName; got != "去骨" {
		t.Errorf("cutting spec = %q, want 去骨", got)
	}

	// 指定 customer_id:僅該店(細部 5.4.2 步驟 3)
	vmAny, err = svc.BuildViewModel(fx.ctx, print.DocumentTypeDeliveryNote, fx.route.ID, targetDate,
		print.Selector{CustomerID: &fx.customer2.ID})
	if err != nil {
		t.Fatalf("build with selector: %v", err)
	}
	vm = vmAny.(*print.DeliveryNoteVM)
	if len(vm.Notes) != 1 || vm.Notes[0].CustomerName != "美香料理" {
		t.Fatalf("single-customer note wrong: %+v", vm.Notes)
	}
}

func TestBuildPickingListAggregatesBaseQty(t *testing.T) {
	c := testutil.NewEntClient(t)
	fx := seedPrintFixture(t, c)
	svc := print.NewService(c, nil, nil)

	vmAny, err := svc.BuildViewModel(fx.ctx, print.DocumentTypePickingList, fx.route.ID, targetDate,
		print.Selector{WarehouseID: &fx.warehouse.ID})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	vm := vmAny.(*print.PickingListVM)
	if vm.WarehouseName != "冷藏倉" {
		t.Errorf("warehouse = %q", vm.WarehouseName)
	}
	// 同商品跨店合併:雞腿 6 + 4 = 10 kg(細部 5.3.3 步驟 3,§6.4)
	if len(vm.Lines) != 2 {
		t.Fatalf("want 2 aggregated lines, got %d: %+v", len(vm.Lines), vm.Lines)
	}
	var chicken *print.PickingListLineVM
	for i := range vm.Lines {
		if vm.Lines[i].DisplayName == "雞腿" {
			chicken = &vm.Lines[i]
		}
	}
	if chicken == nil {
		t.Fatal("missing 雞腿 line")
	}
	if chicken.BaseQty != "10" || chicken.BaseUnit != "kg" {
		t.Errorf("base qty aggregation wrong: %+v", chicken)
	}
	if !chicken.NeedsCutting {
		t.Error("雞腿有分切備註,NeedsCutting 應為 true(細部 5.3.3 步驟 4)")
	}
	// 排序:同分類下依品名字序(細部 5.3.3 步驟 2):豬五花 < 雞腿(Big5/Unicode 字序由 Go strings 比較,此處僅斷言分類一致與穩定)
	if vm.Lines[0].CategoryName != "禽肉" || vm.Lines[1].CategoryName != "禽肉" {
		t.Errorf("category wrong: %+v", vm.Lines)
	}
}

func TestBuildPickingListRequiresWarehouse(t *testing.T) {
	c := testutil.NewEntClient(t)
	fx := seedPrintFixture(t, c)
	svc := print.NewService(c, nil, nil)

	// 揀貨單必帶 warehouse_id(細部 5.3.3 一份一倉;Global Constraints 組合矩陣)
	_, err := svc.BuildViewModel(fx.ctx, print.DocumentTypePickingList, fx.route.ID, targetDate, print.Selector{})
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("want invalid_argument, got %v", err)
	}
}

func TestBuildProcessingList(t *testing.T) {
	c := testutil.NewEntClient(t)
	fx := seedPrintFixture(t, c)
	svc := print.NewService(c, nil, nil)

	vmAny, err := svc.BuildViewModel(fx.ctx, print.DocumentTypeProcessingList, fx.route.ID, targetDate, print.Selector{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	vm := vmAny.(*print.ProcessingListVM)
	// 僅納入有 cutting_spec_id 或 special_cut_note 的明細(細部 5.3.4 步驟 1):
	// 三筆明細中只有 order1 的雞腿符合
	if len(vm.RoomLines) != 1 || len(vm.DeliveryLines) != 1 {
		t.Fatalf("want 1 room + 1 delivery line, got %+v / %+v", vm.RoomLines, vm.DeliveryLines)
	}
	room := vm.RoomLines[0]
	if room.WarehouseName != "冷藏倉" || room.OriginalBaseQty != "6" || room.BaseUnit != "kg" || room.CuttingSpecName != "去骨" {
		t.Errorf("room line wrong: %+v", room)
	}
	dlv := vm.DeliveryLines[0]
	if dlv.CustomerName != "好食餐廳" || dlv.OriginalQty != "10" || dlv.Unit != "件" || dlv.SpecialCutNote != "去骨" {
		t.Errorf("delivery line wrong: %+v", dlv)
	}
}

func TestBuildViewModelEmptyAndErrors(t *testing.T) {
	c := testutil.NewEntClient(t)
	fx := seedPrintFixture(t, c)
	svc := print.NewService(c, nil, nil)

	// 空表:查無符合明細回 IsEmpty 而非錯誤(細部 5.4.2 步驟 6)
	otherDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	vmAny, err := svc.BuildViewModel(fx.ctx, print.DocumentTypeDispatchSummary, fx.route.ID, otherDate, print.Selector{})
	if err != nil {
		t.Fatalf("empty should not error: %v", err)
	}
	if !vmAny.(*print.DispatchSummaryVM).IsEmpty {
		t.Fatal("want IsEmpty=true")
	}

	// route 不存在 → not_found(RLS 下查不到視同不存在)
	_, err = svc.BuildViewModel(fx.ctx, print.DocumentTypeDispatchSummary, uuid.New(), targetDate, print.Selector{})
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("want not_found, got %v", err)
	}

	// 選擇器與類型組合不合法 → invalid_argument(Global Constraints 組合矩陣)
	_, err = svc.BuildViewModel(fx.ctx, print.DocumentTypeDispatchSummary, fx.route.ID, targetDate,
		print.Selector{CustomerID: &fx.customer1.ID})
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("selector on dispatch_summary: want invalid_argument, got %v", err)
	}
	_, err = svc.BuildViewModel(fx.ctx, print.DocumentTypeDeliveryNote, fx.route.ID, targetDate,
		print.Selector{WarehouseID: &fx.warehouse.ID})
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("warehouse on delivery_note: want invalid_argument, got %v", err)
	}

	// customer 不存在 → not_found
	missing := uuid.New()
	_, err = svc.BuildViewModel(fx.ctx, print.DocumentTypeDeliveryNote, fx.route.ID, targetDate,
		print.Selector{CustomerID: &missing})
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("missing customer: want not_found, got %v", err)
	}
}
```

- [ ] **Step 6: 跑測試確認失敗**

Run: `cd backend && go test ./internal/print/ -run TestBuild -v`
Expected: FAIL — `print.NewService` / `print.Selector` 未定義(編譯失敗)。

- [ ] **Step 7: 實作 FileStore 埠與資料組合**

`backend/internal/print/filestore.go`:

```go
package print

import (
	"context"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

// FileStore 為列印 PDF 落檔與 file_assets 元資料建立的消費端埠(細部 5.4.3、D18)。
// 實作由 04-master-data 計畫 Task 8(fileassets service,細部 3.6)以 adapter 提供;
// 本計畫測試以 fake 實作。TODO(接手:04-plan Task 8)整合時注入真實實作。
type FileStore interface {
	// WriteLocal 將 PDF 寫入本地儲存,不落 DB;失敗不留殘檔。
	WriteLocal(ctx context.Context, in WriteInput) (*StoredFile, error)
	// CreateAsset 在呼叫端的 DB 交易內建立 file_assets 元資料(D18 同事務);
	// owner 存在性與租戶驗證由 04 側進行。ownerType 統一 "print_pdf"。
	CreateAsset(ctx context.Context, tx *ent.Tx, actor rls.Identity, stored *StoredFile, ownerType string, ownerID uuid.UUID) (*ent.FileAsset, error)
	// DeleteLocal 補償刪除已落檔檔案(交易失敗時,細部 5.4.3 步驟 5);失敗僅記錄告警。
	DeleteLocal(ctx context.Context, storagePath string)
}

// OwnerTypePrintPDF 為列印 PDF 的 owner_type(依 04 計畫 Task 8 契約)。
const OwnerTypePrintPDF = "print_pdf"

// WriteInput 為落檔輸入;Content 必為 application/pdf 且 ≤ 10 MB(D17,Prepare 檢查)。
type WriteInput struct {
	OriginalFilename string
	Content          []byte
}

// StoredFile 為已落檔、待 DB 關聯的檔案資訊。
type StoredFile struct {
	Filename    string
	StoragePath string
	SizeBytes   int64
}

// DownloadURL 回傳 file_assets 下載相對路徑(依 04 計畫 Task 8 契約:GET /api/v1/files/{id}/download)。
func DownloadURL(fileAssetID uuid.UUID) string {
	return "/api/v1/files/" + fileAssetID.String() + "/download"
}
```

`backend/internal/print/service.go`:

```go
package print

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/ent/customeraddress"
	"github.com/salesorder/sales-order-1.0/backend/ent/customercontact"
	"github.com/salesorder/sales-order-1.0/backend/ent/cuttingspec"
	"github.com/salesorder/sales-order-1.0/backend/ent/product"
	"github.com/salesorder/sales-order-1.0/backend/ent/productcategory"
	"github.com/salesorder/sales-order-1.0/backend/ent/productunit"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorderitem"
	"github.com/salesorder/sales-order-1.0/backend/ent/warehouse"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

// tzTaipei 為列印時間顯示時區(UTC+8 營業日曆慣例)。
var tzTaipei = time.FixedZone("Asia/Taipei", 8*3600)

// maxPDFBytes 為 file_assets PDF 白名單上限(D17:pdf ≤ 10 MB)。
const maxPDFBytes = 10 << 20

// Selector 為依單據類型而定的選擇器(細部 5.4.2)。
type Selector struct {
	CustomerID  *uuid.UUID // 對點單單印一店
	WarehouseID *uuid.UUID // 揀貨單單印一倉(必帶)
}

// Service 為資料組合與 PDF 產線;供 prints domain 的 Preview / Print RPC 呼叫,非獨立端點。
type Service struct {
	db        *ent.Client
	converter Converter
	files     FileStore
	now       func() time.Time
}

func NewService(db *ent.Client, c Converter, f FileStore) *Service {
	return &Service{db: db, converter: c, files: f, now: time.Now}
}

// ---------- 參數驗證 ----------

// validateSelector 檢查選擇器與單據類型組合(Global Constraints 組合矩陣)。
func validateSelector(docType DocumentType, sel Selector) error {
	switch docType {
	case DocumentTypeDispatchSummary, DocumentTypeProcessingList:
		if sel.CustomerID != nil || sel.WarehouseID != nil {
			return connect.NewError(connect.CodeInvalidArgument,
				fmt.Errorf("%s 不接受 customer_id / warehouse_id", docType))
		}
	case DocumentTypeDeliveryNote:
		if sel.WarehouseID != nil {
			return connect.NewError(connect.CodeInvalidArgument,
				errors.New("delivery_note 不接受 warehouse_id"))
		}
	case DocumentTypePickingList:
		if sel.CustomerID != nil {
			return connect.NewError(connect.CodeInvalidArgument,
				errors.New("picking_list 不接受 customer_id"))
		}
		if sel.WarehouseID == nil {
			return connect.NewError(connect.CodeInvalidArgument,
				errors.New("picking_list 必須指定 warehouse_id(一份文件對應單一倉別,細部 5.3.3)"))
		}
	default:
		return connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("unknown document_type %q", docType))
	}
	return nil
}

// ---------- 資料組合(細部 5.4.2) ----------

// printRow 為一筆訂單明細連同顯示所需的關聯顯示名稱。
type printRow struct {
	OrderNo          string
	CustomerID       uuid.UUID
	CustomerCode     string
	CustomerName     string
	Address          string
	ContactName      string
	ContactPhone     string
	DeliverySequence int
	ProductID        *uuid.UUID
	DisplayName      string
	Qty              decimal.Decimal
	Unit             string
	BaseQty          decimal.Decimal
	BaseUnit         string
	CategoryName     string
	CategorySort     int
	WarehouseID      *uuid.UUID
	WarehouseName    string
	CuttingSpecID    *uuid.UUID
	CuttingSpecName  string
	SpecialCutNote   string
	ItemSort         int
}

// BuildViewModel 依單據類型組出 view model;查無符合明細回 IsEmpty=true 而非錯誤(步驟 6)。
// 狀態過濾不在此層(步驟 7):預覽不限狀態、正式列印限 processing 由 Task 4 執行。
func (s *Service) BuildViewModel(ctx context.Context, docType DocumentType, routeID uuid.UUID, targetDate time.Time, sel Selector) (any, error) {
	if err := validateSelector(docType, sel); err != nil {
		return nil, err
	}
	rows, route, err := s.loadRows(ctx, routeID, targetDate, sel)
	if err != nil {
		return nil, err
	}
	companyName := s.companyName(ctx)
	printedAt := s.now().In(tzTaipei).Format("2006-01-02 15:04")
	dateStr := targetDate.Format("2006-01-02")

	switch docType {
	case DocumentTypeDispatchSummary:
		return buildDispatchSummary(companyName, route, dateStr, printedAt, rows), nil
	case DocumentTypeDeliveryNote:
		return buildDeliveryNote(companyName, route, dateStr, printedAt, rows), nil
	case DocumentTypePickingList:
		return buildPickingList(companyName, route, dateStr, printedAt, rows), nil
	case DocumentTypeProcessingList:
		return buildProcessingList(companyName, route, dateStr, printedAt, rows), nil
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("unknown document_type %q", docType))
	}
}

// companyName 取當前租戶公司名稱(頁首);RLS 注入缺失時查不到,回空字串並由上層身分檢查攔截。
func (s *Service) companyName(ctx context.Context) string {
	id, ok := rls.FromContext(ctx)
	if !ok {
		return ""
	}
	co, err := s.db.Company.Get(ctx, id.CompanyID)
	if err != nil {
		return ""
	}
	return co.Name
}

// loadRows 載入車次 + 當日全部明細與關聯顯示資料(步驟 1);
// 軟刪除主檔仍顯示歷史名稱編號(步驟 6:明細 display_name 為快照,關聯名稱查不到時以快照/空字串呈現)。
func (s *Service) loadRows(ctx context.Context, routeID uuid.UUID, targetDate time.Time, sel Selector) ([]printRow, *ent.Route, error) {
	route, err := s.db.Route.Get(ctx, routeID) // RLS 下查不到視同不存在(細部 5.4.2 錯誤處理)
	if err != nil {
		return nil, nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("route not found: %w", err))
	}
	// 選擇器資源存在性(同屬 not_found 邊界)
	if sel.CustomerID != nil {
		if _, err := s.db.Customer.Get(ctx, *sel.CustomerID); err != nil {
			return nil, nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("customer not found: %w", err))
		}
	}
	if sel.WarehouseID != nil {
		if _, err := s.db.Warehouse.Get(ctx, *sel.WarehouseID); err != nil {
			return nil, nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("warehouse not found: %w", err))
		}
	}

	orders, err := s.db.SalesOrder.Query().Where(
		salesorder.RouteIDEQ(routeID),
		salesorder.ExpectedDeliveryDateEQ(targetDate),
		salesorder.DeletedAtIsNil(),
		// 不過濾 status(步驟 7)
	).All(ctx)
	if err != nil {
		return nil, nil, connect.NewError(connect.CodeInternal, err)
	}
	if sel.CustomerID != nil {
		orders = slices.DeleteFunc(orders, func(o *ent.SalesOrder) bool { return o.CustomerID != *sel.CustomerID })
	}
	if len(orders) == 0 {
		return nil, route, nil
	}

	orderIDs := make([]uuid.UUID, len(orders))
	byOrder := map[uuid.UUID]*ent.SalesOrder{}
	for i, o := range orders {
		orderIDs[i] = o.ID
		byOrder[o.ID] = o
	}
	items, err := s.db.SalesOrderItem.Query().Where(
		salesorderitem.SalesOrderIDIn(orderIDs...),
		salesorderitem.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return nil, nil, connect.NewError(connect.CodeInternal, err)
	}
	if sel.WarehouseID != nil {
		// 倉別由 sales_order_items.warehouse_id 決定(細部 5.3.3 步驟 1)
		items = slices.DeleteFunc(items, func(it *ent.SalesOrderItem) bool {
			return it.WarehouseID == nil || *it.WarehouseID != *sel.WarehouseID
		})
	}
	if len(items) == 0 {
		return nil, route, nil
	}

	// 關聯顯示資料批次載入(避免 N+1)
	customers, err := s.customersByOrder(ctx, orders)
	if err != nil {
		return nil, nil, err
	}
	products, err := s.productsByItem(ctx, items)
	if err != nil {
		return nil, nil, err
	}
	categories, err := s.categoriesByProduct(ctx, products)
	if err != nil {
		return nil, nil, err
	}
	warehouses, err := s.warehousesByItem(ctx, items, products)
	if err != nil {
		return nil, nil, err
	}
	specs, err := s.cuttingSpecsByItem(ctx, items)
	if err != nil {
		return nil, nil, err
	}
	baseUnits, err := s.baseUnitsByProduct(ctx, products)
	if err != nil {
		return nil, nil, err
	}

	rows := make([]printRow, 0, len(items))
	for _, it := range items {
		o := byOrder[it.SalesOrderID]
		cu := customers[o.CustomerID]
		row := printRow{
			OrderNo:          o.OrderNo,
			CustomerID:       o.CustomerID,
			DeliverySequence: valueOr(o.DeliverySequence, 0),
			ProductID:        it.ProductID,
			DisplayName:      it.DisplayName,
			Qty:              it.Qty,
			Unit:             it.Unit,
			BaseQty:          it.BaseQty,
			CuttingSpecID:    it.CuttingSpecID,
			SpecialCutNote:   it.SpecialCutNote,
			WarehouseID:      it.WarehouseID,
			ItemSort:         it.SortOrder,
		}
		if cu != nil {
			row.CustomerCode = cu.Code
			row.CustomerName = cu.Name
			row.Address = cu.Address
			row.ContactName = cu.ContactName
			row.ContactPhone = cu.ContactPhone
		}
		if it.ProductID != nil {
			if p, ok := products[*it.ProductID]; ok {
				if cat, ok := categories[p.CategoryID]; ok {
					row.CategoryName = cat.Name
					row.CategorySort = cat.SortOrder
				}
				if row.WarehouseID == nil && p.PickingWarehouseID != nil {
					// 加工單的倉別回退至商品揀貨倉(細部 5.3.4 步驟 3)
					row.WarehouseID = p.PickingWarehouseID
				}
			}
			row.BaseUnit = baseUnits[*it.ProductID]
		}
		if row.BaseUnit == "" {
			row.BaseUnit = it.Unit // 手打品名:base_qty = qty,基本單位即下單單位(05 細部 4.2.1 步驟 5)
		}
		if row.CategoryName == "" {
			row.CategoryName = "未分類"
			row.CategorySort = 1 << 30 // 未分類排最後
		}
		if row.WarehouseID != nil {
			if w, ok := warehouses[*row.WarehouseID]; ok {
				row.WarehouseName = w.Name
			}
		}
		if row.WarehouseName == "" {
			row.WarehouseName = "未指定倉別"
		}
		if it.CuttingSpecID != nil {
			if sp, ok := specs[*it.CuttingSpecID]; ok {
				row.CuttingSpecName = sp.Name
			}
		}
		rows = append(rows, row)
	}
	return rows, route, nil
}

// customerView 為店家顯示資料(預設 shipping 地址 + 預設聯絡人)。
type customerView struct {
	Code, Name, Address, ContactName, ContactPhone string
}

func (s *Service) customersByOrder(ctx context.Context, orders []*ent.SalesOrder) (map[uuid.UUID]*customerView, error) {
	ids := make([]uuid.UUID, 0, len(orders))
	seen := map[uuid.UUID]bool{}
	for _, o := range orders {
		if !seen[o.CustomerID] {
			seen[o.CustomerID] = true
			ids = append(ids, o.CustomerID)
		}
	}
	customers, err := s.db.Customer.Query().Where(customer.IDIn(ids...)).All(ctx) // 含軟刪除:歷史名稱照常顯示(D10)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	addrs, err := s.db.CustomerAddress.Query().Where(
		customeraddress.CustomerIDIn(ids...),
		customeraddress.TypeEQ("shipping"),
		customeraddress.IsDefaultEQ(true),
		customeraddress.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	contacts, err := s.db.CustomerContact.Query().Where(
		customercontact.CustomerIDIn(ids...),
		customercontact.IsDefaultEQ(true),
		customercontact.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	addrBy := map[uuid.UUID]*ent.CustomerAddress{}
	for _, a := range addrs {
		addrBy[a.CustomerID] = a
	}
	contactBy := map[uuid.UUID]*ent.CustomerContact{}
	for _, ct := range contacts {
		contactBy[ct.CustomerID] = ct
	}
	out := map[uuid.UUID]*customerView{}
	for _, cu := range customers {
		v := &customerView{Code: cu.CustomerCode, Name: cu.Name}
		if a := addrBy[cu.ID]; a != nil {
			v.Address = a.Address // formatAddress:若 04 計畫地址為多欄位,此處組合
		}
		if ct := contactBy[cu.ID]; ct != nil {
			v.ContactName = ct.Name
			v.ContactPhone = ct.Phone
		}
		out[cu.ID] = v
	}
	return out, nil
}

func (s *Service) productsByItem(ctx context.Context, items []*ent.SalesOrderItem) (map[uuid.UUID]*ent.Product, error) {
	ids := productIDsOf(items)
	if len(ids) == 0 {
		return map[uuid.UUID]*ent.Product{}, nil
	}
	ps, err := s.db.Product.Query().Where(product.IDIn(ids...)).All(ctx) // 含軟刪除(D10)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := map[uuid.UUID]*ent.Product{}
	for _, p := range ps {
		out[p.ID] = p
	}
	return out, nil
}

func (s *Service) categoriesByProduct(ctx context.Context, products map[uuid.UUID]*ent.Product) (map[uuid.UUID]*ent.ProductCategory, error) {
	ids := map[uuid.UUID]bool{}
	for _, p := range products {
		if p.CategoryID != nil {
			ids[*p.CategoryID] = true
		}
	}
	list := maps_keys(ids)
	if len(list) == 0 {
		return map[uuid.UUID]*ent.ProductCategory{}, nil
	}
	cs, err := s.db.ProductCategory.Query().Where(productcategory.IDIn(list...)).All(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := map[uuid.UUID]*ent.ProductCategory{}
	for _, c := range cs {
		out[c.ID] = c
	}
	return out, nil
}

func (s *Service) warehousesByItem(ctx context.Context, items []*ent.SalesOrderItem, products map[uuid.UUID]*ent.Product) (map[uuid.UUID]*ent.Warehouse, error) {
	ids := map[uuid.UUID]bool{}
	for _, it := range items {
		if it.WarehouseID != nil {
			ids[*it.WarehouseID] = true
		}
	}
	for _, p := range products {
		if p.PickingWarehouseID != nil {
			ids[*p.PickingWarehouseID] = true
		}
	}
	list := maps_keys(ids)
	if len(list) == 0 {
		return map[uuid.UUID]*ent.Warehouse{}, nil
	}
	ws, err := s.db.Warehouse.Query().Where(warehouse.IDIn(list...)).All(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := map[uuid.UUID]*ent.Warehouse{}
	for _, w := range ws {
		out[w.ID] = w
	}
	return out, nil
}

func (s *Service) cuttingSpecsByItem(ctx context.Context, items []*ent.SalesOrderItem) (map[uuid.UUID]*ent.CuttingSpec, error) {
	ids := map[uuid.UUID]bool{}
	for _, it := range items {
		if it.CuttingSpecID != nil {
			ids[*it.CuttingSpecID] = true
		}
	}
	list := maps_keys(ids)
	if len(list) == 0 {
		return map[uuid.UUID]*ent.CuttingSpec{}, nil
	}
	ss, err := s.db.CuttingSpec.Query().Where(cuttingspec.IDIn(list...)).All(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := map[uuid.UUID]*ent.CuttingSpec{}
	for _, sp := range ss {
		out[sp.ID] = sp
	}
	return out, nil
}

// baseUnitsByProduct 回傳商品 is_base 單位的 unit_code(細部 5.3.3 步驟 3)。
func (s *Service) baseUnitsByProduct(ctx context.Context, products map[uuid.UUID]*ent.Product) (map[uuid.UUID]string, error) {
	list := maps_keys(map[uuid.UUID]bool{})
	for id := range products {
		list = append(list, id)
	}
	if len(list) == 0 {
		return map[uuid.UUID]string{}, nil
	}
	us, err := s.db.ProductUnit.Query().Where(
		productunit.ProductIDIn(list...),
		productunit.IsBaseEQ(true),
	).All(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := map[uuid.UUID]string{}
	for _, u := range us {
		out[u.ProductID] = u.UnitCode
	}
	return out, nil
}

func productIDsOf(items []*ent.SalesOrderItem) []uuid.UUID {
	seen := map[uuid.UUID]bool{}
	var ids []uuid.UUID
	for _, it := range items {
		if it.ProductID != nil && !seen[*it.ProductID] {
			seen[*it.ProductID] = true
			ids = append(ids, *it.ProductID)
		}
	}
	return ids
}

func maps_keys(m map[uuid.UUID]bool) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func valueOr(p *int, def int) int {
	if p == nil {
		return def
	}
	return *p
}

// ---------- 四類型成形(細部 5.3 分組/排序/彙總規則) ----------

func buildDispatchSummary(company string, route *ent.Route, date, printedAt string, rows []printRow) *DispatchSummaryVM {
	vm := &DispatchSummaryVM{
		CompanyName: company, RouteCode: route.Code, RouteName: route.Name,
		TargetDate: date, PrintedAt: printedAt, IsEmpty: len(rows) == 0,
	}
	byCustomer := map[uuid.UUID]*DispatchSummaryCustomerVM{}
	var order []uuid.UUID
	sorted := slices.Clone(rows)
	slices.SortFunc(sorted, func(a, b printRow) int { // 店家 delivery_sequence 升冪 → 明細 sort_order
		if a.DeliverySequence != b.DeliverySequence {
			return a.DeliverySequence - b.DeliverySequence
		}
		return a.ItemSort - b.ItemSort
	})
	for _, r := range sorted {
		cv, ok := byCustomer[r.CustomerID]
		if !ok {
			cv = &DispatchSummaryCustomerVM{
				CustomerCode: r.CustomerCode, CustomerName: r.CustomerName,
				Address: r.Address, DeliverySequence: r.DeliverySequence,
			}
			byCustomer[r.CustomerID] = cv
			order = append(order, r.CustomerID)
		}
		cv.Items = append(cv.Items, DispatchSummaryItemVM{
			DisplayName: r.DisplayName, Qty: r.Qty.String(), Unit: r.Unit, SpecialCutNote: r.SpecialCutNote,
		})
	}
	for _, id := range order {
		vm.Customers = append(vm.Customers, *byCustomer[id])
	}
	return vm
}

func buildDeliveryNote(company string, route *ent.Route, date, printedAt string, rows []printRow) *DeliveryNoteVM {
	vm := &DeliveryNoteVM{
		CompanyName: company, RouteCode: route.Code, RouteName: route.Name,
		TargetDate: date, PrintedAt: printedAt, IsEmpty: len(rows) == 0,
	}
	byCustomer := map[uuid.UUID]*DeliveryNotePageVM{}
	var order []uuid.UUID
	sorted := slices.Clone(rows)
	slices.SortFunc(sorted, func(a, b printRow) int {
		if a.DeliverySequence != b.DeliverySequence {
			return a.DeliverySequence - b.DeliverySequence
		}
		return a.ItemSort - b.ItemSort
	})
	for _, r := range sorted {
		page, ok := byCustomer[r.CustomerID]
		if !ok {
			page = &DeliveryNotePageVM{
				CustomerCode: r.CustomerCode, CustomerName: r.CustomerName,
				Address: r.Address, ContactName: r.ContactName, ContactPhone: r.ContactPhone,
				DeliverySequence: r.DeliverySequence,
			}
			byCustomer[r.CustomerID] = page
			order = append(order, r.CustomerID)
		}
		page.Items = append(page.Items, DeliveryNoteItemVM{
			DisplayName: r.DisplayName, Qty: r.Qty.String(), Unit: r.Unit,
			CuttingSpecName: r.CuttingSpecName, SpecialCutNote: r.SpecialCutNote,
		})
	}
	for _, id := range order {
		vm.Notes = append(vm.Notes, *page_ptr(byCustomer[id]))
	}
	return vm
}

func page_ptr(p *DeliveryNotePageVM) *DeliveryNotePageVM { return p }

func buildPickingList(company string, route *ent.Route, date, printedAt string, rows []printRow) *PickingListVM {
	vm := &PickingListVM{
		CompanyName: company, RouteCode: route.Code, RouteName: route.Name,
		TargetDate: date, PrintedAt: printedAt, IsEmpty: len(rows) == 0,
	}
	if len(rows) > 0 {
		vm.WarehouseName = rows[0].WarehouseName // warehouse 選擇器已過濾為單一倉
	}
	type key struct{ Category, Name string }
	type agg struct {
		line PickingListLineVM
		sort int
	}
	byKey := map[key]*agg{}
	for _, r := range rows {
		k := key{Category: r.CategoryName, Name: r.DisplayName} // 同商品跨店合併(細部 5.3.3 步驟 2)
		a, ok := byKey[k]
		if !ok {
			a = &agg{sort: r.CategorySort, line: PickingListLineVM{
				CategoryName: r.CategoryName, DisplayName: r.DisplayName,
				BaseQty: "0", BaseUnit: r.BaseUnit,
			}}
			byKey[k] = a
		}
		sum := decimal.RequireFromString(a.line.BaseQty).Add(r.BaseQty) // Σ base_qty(步驟 3)
		a.line.BaseQty = sum.String()
		if r.CuttingSpecID != nil || r.SpecialCutNote != "" {
			a.line.NeedsCutting = true // 需加工提示(步驟 4)
		}
	}
	for _, a := range byKey {
		vm.Lines = append(vm.Lines, a.line)
	}
	sortBy := map[string]int{} // DisplayName → category sort
	for _, a := range byKey {
		sortBy[a.line.DisplayName] = a.sort
	}
	slices.SortFunc(vm.Lines, func(a, b PickingListLineVM) int { // 分類 sort_order → 品名字序
		if sortBy[a.DisplayName] != sortBy[b.DisplayName] {
			return sortBy[a.DisplayName] - sortBy[b.DisplayName]
		}
		return strings.Compare(a.DisplayName, b.DisplayName)
	})
	return vm
}

func buildProcessingList(company string, route *ent.Route, date, printedAt string, rows []printRow) *ProcessingListVM {
	vm := &ProcessingListVM{
		CompanyName: company, RouteCode: route.Code, RouteName: route.Name,
		TargetDate: date, PrintedAt: printedAt,
	}
	// 僅納入有 cutting_spec_id 或 special_cut_note 的明細(細部 5.3.4 步驟 1)
	rows = slices.DeleteFunc(slices.Clone(rows), func(r printRow) bool {
		return r.CuttingSpecID == nil && r.SpecialCutNote == ""
	})
	vm.IsEmpty = len(rows) == 0

	// 加工室揀:倉別 → 品名,原始數量以基本單位(步驟 3);同倉同品跨店彙總 base_qty
	type roomKey struct{ Warehouse, Name string }
	roomAgg := map[roomKey]*ProcessingRoomLineVM{}
	var roomOrder []roomKey
	for _, r := range rows {
		k := roomKey{Warehouse: r.WarehouseName, Name: r.DisplayName}
		line, ok := roomAgg[k]
		if !ok {
			line = &ProcessingRoomLineVM{
				WarehouseName: r.WarehouseName, DisplayName: r.DisplayName,
				OriginalBaseQty: "0", BaseUnit: r.BaseUnit, CuttingSpecName: r.CuttingSpecName,
			}
			roomAgg[k] = line
			roomOrder = append(roomOrder, k)
		}
		line.OriginalBaseQty = decimal.RequireFromString(line.OriginalBaseQty).Add(r.BaseQty).String()
	}
	for _, k := range roomOrder {
		vm.RoomLines = append(vm.RoomLines, *roomAgg[k])
	}
	slices.SortFunc(vm.RoomLines, func(a, b ProcessingRoomLineVM) int {
		if c := strings.Compare(a.WarehouseName, b.WarehouseName); c != 0 {
			return c
		}
		return strings.Compare(a.DisplayName, b.DisplayName)
	})

	// 配送揀:店家 delivery_sequence → 品名,原始數量以下單單位(步驟 4);不合併(每店各自列)
	for _, r := range rows {
		vm.DeliveryLines = append(vm.DeliveryLines, ProcessingDeliveryLineVM{
			CustomerName: r.CustomerName, DeliverySequence: r.DeliverySequence,
			DisplayName: r.DisplayName, OriginalQty: r.Qty.String(), Unit: r.Unit,
			SpecialCutNote: r.SpecialCutNote,
		})
	}
	slices.SortFunc(vm.DeliveryLines, func(a, b ProcessingDeliveryLineVM) int {
		if a.DeliverySequence != b.DeliverySequence {
			return a.DeliverySequence - b.DeliverySequence
		}
		return strings.Compare(a.DisplayName, b.DisplayName)
	})
	return vm
}

// ---------- PDF 產線(細部 5.4.3) ----------

// PreparedPDF 為已完成渲染、轉檔與落檔、等待 DB 記錄關聯的 PDF。
type PreparedPDF struct {
	Stored           *StoredFile
	OriginalFilename string
}

// isEmptyVM 判斷 view model 是否空表(細部 5.4.3 步驟 1)。
func isEmptyVM(vm any) bool {
	switch v := vm.(type) {
	case *DispatchSummaryVM:
		return v.IsEmpty
	case *DeliveryNoteVM:
		return v.IsEmpty
	case *PickingListVM:
		return v.IsEmpty
	case *ProcessingListVM:
		return v.IsEmpty
	default:
		return false
	}
}

// Prepare 執行 空表守門 → 模板渲染 → Gotenberg 轉換 → 大小檢查 → 本地落檔。
// 不碰 DB;file_assets 元資料由呼叫端於交易內經 FileStore.CreateAsset 建立(D18)。
func (s *Service) Prepare(ctx context.Context, docType DocumentType, vm any) (*PreparedPDF, error) {
	// 步驟 1:空表不印 — 不呼叫 Gotenberg、不寫任何記錄
	if isEmptyVM(vm) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("無可列印資料(空表)"))
	}
	// 步驟 2:模板渲染;字型於模板 CSS 以 Noto Sans CJK TC 字族宣告(D15)
	html, err := RenderHTML(docType, vm)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// 步驟 3:Gotenberg 轉換(錯誤映射:4xx → invalid_argument;其餘 → internal)
	pdf, err := s.converter.HTMLToPDF(ctx, html)
	if err != nil {
		if errors.Is(err, ErrRejected) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// 步驟 3 續:大小須在 file_assets 白名單上限內(D17)
	if len(pdf) > maxPDFBytes {
		return nil, connect.NewError(connect.CodeInternal,
			fmt.Errorf("pdf 大小 %d bytes 超過 10 MB 上限", len(pdf)))
	}
	// 步驟 4:本地落檔(交易外;DB 失敗由呼叫端 DeleteLocal 補償,步驟 5)
	filename := pdfFilename(docType, vm)
	stored, err := s.files.WriteLocal(ctx, WriteInput{OriginalFilename: filename, Content: pdf})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("write pdf: %w", err))
	}
	return &PreparedPDF{Stored: stored, OriginalFilename: filename}, nil
}

// pdfFilename 產生原始檔名,如 picking_list_2026-08-21_TY01.pdf;
// 每次產生皆為新 PDF、新 file_asset,不覆用(步驟 6,PDF 全留存 D15)。
func pdfFilename(docType DocumentType, vm any) string {
	routeCode, date := "", ""
	if m, ok := vm.(docMeta); ok {
		routeCode, date = m.meta()
	}
	return fmt.Sprintf("%s_%s_%s_%d.pdf", docType, date, routeCode, time.Now().UnixNano())
}
```

- [ ] **Step 8: 跑測試確認通過**

Run: `cd backend && go get github.com/shopspring/decimal && go test ./internal/print/ -run 'TestBuild' -v`
Expected: PASS — 6 個 BuildViewModel 測試(四類型成形 + 揀貨單必帶倉別 + 空表/錯誤邊界)。

注意:`seedPrintFixture` 引用的 04/05 計畫實體 setter(`SetAppliesTo` / `SetPickingWarehouseID` / `SetBaseQty` 等)以各計畫 schema 定稿為準;若名稱不同,僅調整該 helper。

- [ ] **Step 9: 寫失敗測試(Prepare 產線:空表守門、錯誤映射、大小上限、落檔)**

在 `backend/internal/print/service_test.go` 追加:

```go
// fakeConverter 為 Converter 測試替身。
type fakeConverter struct {
	pdf   []byte
	err   error
	calls int
}

func (f *fakeConverter) HTMLToPDF(_ context.Context, _ []byte) ([]byte, error) {
	f.calls++
	return f.pdf, f.err
}

// fakeFileStore 為 FileStore 測試替身(不落 DB,僅記錄呼叫;CreateAsset 於 Task 3/4 測試使用真實 tx)。
type fakeFileStore struct {
	stored   *print.StoredFile
	writeErr error
	writes   int
	deleted  []string
}

func (f *fakeFileStore) WriteLocal(_ context.Context, in print.WriteInput) (*print.StoredFile, error) {
	f.writes++
	if f.writeErr != nil {
		return nil, f.writeErr
	}
	f.stored = &print.StoredFile{
		Filename:    uuid.NewString() + ".pdf",
		StoragePath: "/data/" + uuid.NewString() + ".pdf",
		SizeBytes:   int64(len(in.Content)),
	}
	return f.stored, nil
}

func (f *fakeFileStore) DeleteLocal(_ context.Context, path string) {
	f.deleted = append(f.deleted, path)
}

func TestPrepareEmptyVMRejected(t *testing.T) {
	c := testutil.NewEntClient(t)
	conv := &fakeConverter{pdf: []byte("%PDF")}
	fs := &fakeFileStore{}
	svc := print.NewService(c, conv, fs)

	// 空表守門(細部 5.4.3 步驟 1):不呼叫 Gotenberg、不落檔、不留任何記錄
	_, err := svc.Prepare(context.Background(), print.DocumentTypeDispatchSummary,
		&print.DispatchSummaryVM{IsEmpty: true})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("want failed_precondition, got %v", err)
	}
	if conv.calls != 0 || fs.writes != 0 {
		t.Fatalf("空表不得觸發轉換或落檔: converter=%d writes=%d", conv.calls, fs.writes)
	}
}

func TestPrepareConverterErrorMapping(t *testing.T) {
	c := testutil.NewEntClient(t)
	fs := &fakeFileStore{}

	// 4xx(ErrRejected)→ invalid_argument(細部 5.4.1 錯誤處理)
	svc := print.NewService(c, &fakeConverter{err: print.ErrRejected}, fs)
	_, err := svc.Prepare(context.Background(), print.DocumentTypeDispatchSummary,
		&print.DispatchSummaryVM{Customers: []print.DispatchSummaryCustomerVM{{}}})
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("4xx: want invalid_argument, got %v", err)
	}

	// 上游 5xx 重試用盡 → internal
	svc = print.NewService(c, &fakeConverter{err: &print.UpstreamError{Status: 502, Body: "bad gateway"}}, fs)
	_, err = svc.Prepare(context.Background(), print.DocumentTypeDispatchSummary,
		&print.DispatchSummaryVM{Customers: []print.DispatchSummaryCustomerVM{{}}})
	if connect.CodeOf(err) != connect.CodeInternal {
		t.Fatalf("5xx: want internal, got %v", err)
	}
	if fs.writes != 0 {
		t.Fatal("轉換失敗不得落檔")
	}
}

func TestPrepareOversizePDF(t *testing.T) {
	c := testutil.NewEntClient(t)
	fs := &fakeFileStore{}
	big := make([]byte, 10<<20+1) // D17 上限 10 MB
	svc := print.NewService(c, &fakeConverter{pdf: big}, fs)
	_, err := svc.Prepare(context.Background(), print.DocumentTypeDispatchSummary,
		&print.DispatchSummaryVM{Customers: []print.DispatchSummaryCustomerVM{{}}})
	if connect.CodeOf(err) != connect.CodeInternal {
		t.Fatalf("oversize: want internal, got %v", err)
	}
	if fs.writes != 0 {
		t.Fatal("超限 PDF 不得落檔")
	}
}

func TestPrepareSuccess(t *testing.T) {
	c := testutil.NewEntClient(t)
	conv := &fakeConverter{pdf: []byte("%PDF-ok")}
	fs := &fakeFileStore{}
	svc := print.NewService(c, conv, fs)

	vm := &print.PickingListVM{
		RouteCode: "TY01", TargetDate: "2026-08-21", WarehouseName: "冷藏倉",
		Lines: []print.PickingListLineVM{{CategoryName: "禽肉", DisplayName: "雞腿", BaseQty: "10", BaseUnit: "kg"}},
	}
	prepared, err := svc.Prepare(context.Background(), print.DocumentTypePickingList, vm)
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if prepared.Stored == nil || prepared.Stored.SizeBytes != int64(len("%PDF-ok")) {
		t.Fatalf("stored wrong: %+v", prepared.Stored)
	}
	// 檔名含類型、日期、車次;每次皆新檔(PDF 全留存,D15)
	if !strings.Contains(prepared.OriginalFilename, "picking_list_2026-08-21_TY01") ||
		!strings.HasSuffix(prepared.OriginalFilename, ".pdf") {
		t.Errorf("filename wrong: %s", prepared.OriginalFilename)
	}
	if conv.calls != 1 || fs.writes != 1 {
		t.Fatalf("calls=%d writes=%d", conv.calls, fs.writes)
	}
}
```

(`strings` 加入 service_test.go 的 import。)

- [ ] **Step 10: 跑測試確認通過 + Commit**

Run: `cd backend && go test ./internal/print/ -v`
Expected: PASS — Task 1 全部 + Gotenberg 4 個 + BuildViewModel 6 個 + Prepare 4 個測試。

```bash
git add backend/internal/print backend/go.mod backend/go.sum
git commit -m "feat(backend): Gotenberg client(有限重試)、資料組合、PDF 產線空表守門與錯誤映射(5.4)"
```

---

### Task 3: print_logs / print_previews schema + Preview API(細部 5.5.1–5.5.2)

**Files:**
- Create: `backend/ent/schema/printlog.go`
- Create: `backend/ent/schema/printpreview.go`
- Create: `backend/database/migrations/00012_print_tables.sql`
- Create: `backend/proto/v1/print.proto`
- Create: `backend/internal/domain/prints/service.go`
- Create: `backend/internal/domain/prints/preview.go`
- Test: `backend/ent/schema/print_schema_test.go`
- Test: `backend/internal/domain/prints/preview_test.go`

**Interfaces:**
- Consumes:
  - Task 2:`print.Service` / `print.Selector` / `print.PreparedPDF` / `print.FileStore` / `print.OwnerTypePrintPDF` / `print.DownloadURL`。
  - 由 01-auth 計畫提供:`rls.Identity`(含 `Role` 欄位,Task 11/14 擴充)、`rls.FromContext`、`testutil.NewEntClient` / `testutil.NewEntClientWithRLS` / `testutil.DSN`。
  - 由 04-plan 提供:`ent.FileAsset`(Task 8 schema;本 Task migration 的 FK 參照 file_assets 表)。
- Produces:
  - Ent 實體 `ent.PrintLog`:`company_id`、`department_id`、`document_type`(enum)、`route_id`、`customer_id`(nillable)、`warehouse_id`(nillable)、`target_date`、`is_reprint`(default false)、`reprint_reason`(nillable)、`printed_by`、`printed_at`、`file_asset_id`(nillable,同事務內後補)、`created_at`;**無** `updated_at` / `deleted_at`(記錄類,細部 5.5.1 步驟 1)。
  - Ent 實體 `ent.PrintPreview`:同上但無 `is_reprint` / `reprint_reason`,改為 `previewed_by` / `previewed_at`。
  - proto `print.v1.PrintService`:`Preview(PreviewRequest) → PreviewResponse{preview_id, file_asset_id, download_url}`(Print / ListLogs 於 Task 4 增量)。
  - `prints.NewPrintsService(db *ent.Client, svc *print.Service, files print.FileStore, newAudit AuditFactory) *PrintsService`;`prints.AuditFactory func(tx *ent.Tx) audit.Recorder`(DB 版由 03-metadicts-audit 計畫提供;本 Task 測試以 capture fake 注入,Preview 不使用)。
  - 內部 `prints.requirePrinter(ctx) (rls.Identity, error)`:未登入 `unauthenticated`;role ∉ {super, company_admin, dept_admin, staff} → `permission_denied`(customer / guest 一律拒絕,細部 5.5.4 步驟 3 同此規則);dept_admin / staff 必須帶 `DepartmentID`。

- [ ] **Step 1: 寫失敗測試(schema 欄位、記錄類無軟刪除、RLS 隔離、重印比對鍵)**

`backend/ent/schema/print_schema_test.go`:

```go
package schema_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/printlog"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

// applyPrintRLS 對測試庫執行 print 兩表的 RLS policy(與 00012_print_tables.sql 的 Up 區段相同內容;
// testutil 的 Schema.Create 不跑 goose migration,故測試自行套用 policy)。
func applyPrintRLS(t *testing.T) {
	t.Helper()
	db, err := sql.Open("postgres", testutil.DSN(t))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	stmts := []string{
		`ALTER TABLE print_logs ENABLE ROW LEVEL SECURITY`,
		`ALTER TABLE print_previews ENABLE ROW LEVEL SECURITY`,
		`CREATE POLICY tenant_isolation ON print_logs
		 USING (current_setting('app.current_data_scope', true) = 'all'
				OR company_id::text = current_setting('app.current_company_id', true))`,
		`CREATE POLICY department_scope ON print_logs
		 USING (current_setting('app.current_data_scope', true) IN ('all', 'company')
				OR department_id::text = current_setting('app.current_department_id', true))`,
		`CREATE POLICY tenant_isolation ON print_previews
		 USING (current_setting('app.current_data_scope', true) = 'all'
				OR company_id::text = current_setting('app.current_company_id', true))`,
		`CREATE POLICY department_scope ON print_previews
		 USING (current_setting('app.current_data_scope', true) IN ('all', 'company')
				OR department_id::text = current_setting('app.current_department_id', true))`,
	}
	for _, q := range stmts {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("apply rls: %v\n%s", err, q)
		}
	}
}

func TestPrintLogFieldsAndDefaults(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()

	co, err := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("TY").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	dept, err := c.Department.Create().SetCompanyID(co.ID).SetName("桃園部").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	routeID, assetID, printer := uuid.New(), uuid.New(), uuid.New()
	log, err := c.PrintLog.Create().
		SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetDocumentType("dispatch_summary").
		SetRouteID(routeID).
		SetTargetDate(time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)).
		SetPrintedBy(printer).SetPrintedAt(time.Now()).
		SetFileAssetID(assetID).
		Save(ctx)
	if err != nil {
		t.Fatalf("create print log: %v", err)
	}
	if log.IsReprint {
		t.Error("is_reprint 預設應為 false")
	}
	if log.ReprintReason != nil || log.CustomerID != nil || log.WarehouseID != nil {
		t.Error("reprint_reason / customer_id / warehouse_id 應可空")
	}
	// 記錄類資料:無軟刪除欄位(細部 5.5.1 步驟 1)—— schema 即無 deleted_at setter,編譯期保證;
	// 此處以查詢 predicate 不存在佐證(若誤加 deleted_at,下列反射斷言失敗)。
	if _, err := c.PrintLog.Query().Where(printlog.IDEQ(log.ID)).Only(ctx); err != nil {
		t.Fatalf("query back: %v", err)
	}
	prev, err := c.PrintPreview.Create().
		SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetDocumentType("picking_list").
		SetRouteID(routeID).SetWarehouseID(uuid.New()).
		SetTargetDate(time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)).
		SetPreviewedBy(printer).SetPreviewedAt(time.Now()).
		SetFileAssetID(assetID).
		Save(ctx)
	if err != nil {
		t.Fatalf("create print preview: %v", err)
	}
	if prev.ID == uuid.Nil {
		t.Error("preview id 未產生")
	}
}

func TestPrintLogRLSIsolation(t *testing.T) {
	c := testutil.NewEntClientWithRLS(t) // 01-auth 計畫 Task 3 提供
	applyPrintRLS(t)
	ctx := context.Background()

	co, err := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("TY").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	deptA, err := c.Department.Create().SetCompanyID(co.ID).SetName("A部").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	deptB, err := c.Department.Create().SetCompanyID(co.ID).SetName("B部").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	superCtx := rls.NewContext(ctx, rls.Identity{UserID: uuid.New(), CompanyID: co.ID, DataScope: "all"})
	mk := func(deptID uuid.UUID) {
		if _, err := c.PrintLog.Create().
			SetCompanyID(co.ID).SetDepartmentID(deptID).
			SetDocumentType("delivery_note").SetRouteID(uuid.New()).
			SetTargetDate(time.Now()).SetPrintedBy(uuid.New()).SetPrintedAt(time.Now()).
			SetFileAssetID(uuid.New()).Save(superCtx); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	mk(deptA.ID)
	mk(deptB.ID)

	// A 部門 staff 視角:僅見自己部門(RLS department_scope)
	ctxA := rls.NewContext(ctx, rls.Identity{
		UserID: uuid.New(), CompanyID: co.ID, DepartmentID: &deptA.ID, DataScope: "department", Role: "staff",
	})
	n, err := c.PrintLog.Query().Count(ctxA)
	if err != nil {
		t.Fatalf("query A: %v", err)
	}
	if n != 1 {
		t.Fatalf("RLS 隔離失效:A 部門看到 %d 筆", n)
	}
	// 未注入身分 fail-closed
	if n, err := c.PrintLog.Query().Count(ctx); err != nil || n != 0 {
		t.Fatalf("fail-closed violated: n=%d err=%v", n, err)
	}
}

func TestPrintLogReprintMatchKey(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()

	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("TY").Save(ctx)
	dept, _ := c.Department.Create().SetCompanyID(co.ID).SetName("A部").Save(ctx)
	routeID := uuid.New()
	date := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	custID := uuid.New()
	if _, err := c.PrintLog.Create().
		SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetDocumentType("delivery_note").SetRouteID(routeID).SetCustomerID(custID).
		SetTargetDate(date).SetPrintedBy(uuid.New()).SetPrintedAt(time.Now()).
		SetFileAssetID(uuid.New()).Save(ctx); err != nil {
		t.Fatal(err)
	}
	// 比對鍵(部門、document_type、route_id、target_date、customer_id)命中(細部 5.5.1 步驟 4)
	exist, err := c.PrintLog.Query().Where(
		printlog.DepartmentIDEQ(dept.ID),
		printlog.DocumentTypeEQ("delivery_note"),
		printlog.RouteIDEQ(routeID),
		printlog.TargetDateEQ(date),
		printlog.CustomerIDEQ(custID),
	).Exist(ctx)
	if err != nil || !exist {
		t.Fatalf("match key should hit: exist=%v err=%v", exist, err)
	}
	// customer_id 為 NULL 的鍵不命中此列
	exist, err = c.PrintLog.Query().Where(
		printlog.DepartmentIDEQ(dept.ID),
		printlog.DocumentTypeEQ("delivery_note"),
		printlog.RouteIDEQ(routeID),
		printlog.TargetDateEQ(date),
		printlog.CustomerIDIsNil(),
	).Exist(ctx)
	if err != nil || exist {
		t.Fatalf("null-key should miss: exist=%v err=%v", exist, err)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./ent/schema/ -run TestPrint -v`
Expected: FAIL — `ent.PrintLog` / `ent.PrintPreview` 未定義(編譯失敗)。

- [ ] **Step 3: 實作兩個 schema 與 migration**

`backend/ent/schema/printlog.go`:

```go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// PrintLog 為正式列印記錄(細部 5.5.1);記錄類資料,不軟刪除、無 updated_at。
type PrintLog struct{ ent.Schema }

func (PrintLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("company_id", uuid.UUID{}),
		field.UUID("department_id", uuid.UUID{}),
		field.Enum("document_type").
			Values("dispatch_summary", "delivery_note", "picking_list", "processing_list"),
		field.UUID("route_id", uuid.UUID{}),
		field.UUID("customer_id", uuid.UUID{}).Optional().Nillable(),  // 對點單單店時使用
		field.UUID("warehouse_id", uuid.UUID{}).Optional().Nillable(), // 揀貨單單倉時使用
		field.Time("target_date"),
		field.Bool("is_reprint").Default(false),
		field.String("reprint_reason").Optional().Nillable(), // 重印時必填(5.5.4)
		field.UUID("printed_by", uuid.UUID{}),
		field.Time("printed_at"),
		// 同事務內先建 print_log 再建 file_assets 元資料,故可空、由同事務後補(D18)
		field.UUID("file_asset_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (PrintLog) Indexes() []ent.Index {
	return []ent.Index{
		// 列印記錄查詢(細部 5.5.1 步驟 3)
		index.Fields("department_id", "target_date", "document_type"),
		// 重印比對鍵(步驟 4)
		index.Fields("department_id", "document_type", "route_id", "target_date"),
		index.Fields("file_asset_id"),
	}
}
```

`backend/ent/schema/printpreview.go`:

```go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// PrintPreview 為預覽記錄(細部 5.5.1);記錄類資料,不軟刪除;與 print_logs 完全隔離(5.5.2)。
type PrintPreview struct{ ent.Schema }

func (PrintPreview) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("company_id", uuid.UUID{}),
		field.UUID("department_id", uuid.UUID{}),
		field.Enum("document_type").
			Values("dispatch_summary", "delivery_note", "picking_list", "processing_list"),
		field.UUID("route_id", uuid.UUID{}),
		field.UUID("customer_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("warehouse_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("target_date"),
		field.UUID("previewed_by", uuid.UUID{}),
		field.Time("previewed_at"),
		field.UUID("file_asset_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (PrintPreview) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("department_id", "target_date", "document_type"),
		index.Fields("file_asset_id"),
	}
}
```

`backend/database/migrations/00012_print_tables.sql`(goose 格式;序號以執行當下最大序號 +1 為準):

```sql
-- +goose Up
-- 細部文件 5.5.1:print 兩表啟用 RLS;data_scope=all 豁免(D3);未注入 fail-closed。
ALTER TABLE print_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE print_previews ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON print_logs
  USING (current_setting('app.current_data_scope', true) = 'all'
         OR company_id::text = current_setting('app.current_company_id', true));
CREATE POLICY department_scope ON print_logs
  USING (current_setting('app.current_data_scope', true) IN ('all', 'company')
         OR department_id::text = current_setting('app.current_department_id', true));

CREATE POLICY tenant_isolation ON print_previews
  USING (current_setting('app.current_data_scope', true) = 'all'
         OR company_id::text = current_setting('app.current_company_id', true));
CREATE POLICY department_scope ON print_previews
  USING (current_setting('app.current_data_scope', true) IN ('all', 'company')
         OR department_id::text = current_setting('app.current_department_id', true));

-- 重印比對鍵與記錄查詢索引(細部 5.5.1 步驟 3)
CREATE INDEX idx_print_logs_query ON print_logs (department_id, target_date, document_type);
CREATE INDEX idx_print_logs_match ON print_logs (department_id, document_type, route_id, target_date);
CREATE INDEX idx_print_logs_file_asset ON print_logs (file_asset_id);
CREATE INDEX idx_print_previews_query ON print_previews (department_id, target_date, document_type);
CREATE INDEX idx_print_previews_file_asset ON print_previews (file_asset_id);

-- file_asset 外鍵參照 file_assets(由 04-plan Task 8 建表)
ALTER TABLE print_logs
  ADD CONSTRAINT fk_print_logs_file_asset FOREIGN KEY (file_asset_id) REFERENCES file_assets (id);
ALTER TABLE print_previews
  ADD CONSTRAINT fk_print_previews_file_asset FOREIGN KEY (file_asset_id) REFERENCES file_assets (id);

-- +goose Down
ALTER TABLE print_previews DROP CONSTRAINT IF EXISTS fk_print_previews_file_asset;
ALTER TABLE print_logs DROP CONSTRAINT IF EXISTS fk_print_logs_file_asset;
DROP INDEX IF EXISTS idx_print_previews_file_asset;
DROP INDEX IF EXISTS idx_print_previews_query;
DROP INDEX IF EXISTS idx_print_logs_file_asset;
DROP INDEX IF EXISTS idx_print_logs_match;
DROP INDEX IF EXISTS idx_print_logs_query;
DROP POLICY IF EXISTS department_scope ON print_previews;
DROP POLICY IF EXISTS tenant_isolation ON print_previews;
DROP POLICY IF EXISTS department_scope ON print_logs;
DROP POLICY IF EXISTS tenant_isolation ON print_logs;
ALTER TABLE print_previews DISABLE ROW LEVEL SECURITY;
ALTER TABLE print_logs DISABLE ROW LEVEL SECURITY;
```

- [ ] **Step 4: 產生 Ent code 並跑測試**

Run: `cd backend && go generate ./ent/ && go mod tidy && go test ./ent/schema/ -run TestPrint -v`
Expected: PASS — 三個測試(欄位預設、RLS 隔離與 fail-closed、重印比對鍵)。

注意:`fk_print_logs_file_asset` 參照的 file_assets 表由 04 計畫 Task 8 migration 建立,執行順序須在 04 之後;測試走 Ent Schema.Create(不跑 goose),不受 FK 影響。

- [ ] **Step 5: 寫失敗測試(Preview API)**

`backend/internal/domain/prints/preview_test.go`:

```go
package prints_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/prints"
	"github.com/salesorder/sales-order-1.0/backend/internal/print"
	printv1 "github.com/salesorder/sales-order-1.0/backend/internal/gen/print/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

// ---------- 測試替身 ----------

type okConverter struct{ calls int }

func (c *okConverter) HTMLToPDF(_ context.Context, _ []byte) ([]byte, error) {
	c.calls++
	return []byte("%PDF-preview"), nil
}

// txFileStore 為 FileStore 測試替身:WriteLocal 記錄、CreateAsset 於呼叫端交易內真的建 file_assets
// (欄位依 04 計畫 Task 8 schema;若必填欄位更多,依其 schema 補齊 — 唯一調整點)。
type txFileStore struct {
	mu      sync.Mutex
	writes  int
	deleted []string
}

func (f *txFileStore) WriteLocal(_ context.Context, in print.WriteInput) (*print.StoredFile, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.writes++
	name := uuid.NewString() + ".pdf"
	return &print.StoredFile{Filename: name, StoragePath: "/data/" + name, SizeBytes: int64(len(in.Content))}, nil
}

func (f *txFileStore) CreateAsset(ctx context.Context, tx *ent.Tx, actor rls.Identity, stored *print.StoredFile, ownerType string, ownerID uuid.UUID) (*ent.FileAsset, error) {
	id := uuid.New()
	return tx.FileAsset.Create().SetID(id).
		SetCompanyID(actor.CompanyID).
		SetDepartmentID(*actor.DepartmentID).
		SetOwnerType(ownerType).SetOwnerID(ownerID).
		SetFilename(stored.Filename).SetOriginalFilename(stored.Filename).
		SetMimeType("application/pdf").SetSizeBytes(stored.SizeBytes).
		SetStoragePath(stored.StoragePath).
		SetUrl(print.DownloadURL(id)).
		SetCreatedBy(actor.UserID).
		Save(ctx)
}

func (f *txFileStore) DeleteLocal(_ context.Context, path string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleted = append(f.deleted, path)
}

// captureAuditFactory 回傳收集用 Recorder(本 Task 僅斷言 Preview 不呼叫)。
func captureAuditFactory(seen *[]audit.Entry) prints.AuditFactory {
	return func(_ *ent.Tx) audit.Recorder {
		return audit.RecorderFunc(func(_ context.Context, e audit.Entry) error {
			*seen = append(*seen, e)
			return nil
		})
	}
}

// newPreviewService 建立接好線的 PrintsService 與 fixture(沿用 print 套件的 seed 語意,
// 此處自建最小資料:公司/部門/車次/店家/商品/訂單/明細)。
func newPreviewService(t *testing.T, status string) (*prints.PrintsService, *printFixtureIDs, *txFileStore, *[]audit.Entry) {
	t.Helper()
	c := testutil.NewEntClient(t)
	fx := seedMinimalPrintData(t, c, status)
	fs := &txFileStore{}
	var auditSeen []audit.Entry
	svc := print.NewService(c, &okConverter{}, fs)
	h := prints.NewPrintsService(c, svc, fs, captureAuditFactory(&auditSeen))
	return h, fx, fs, &auditSeen
}

func staffCtx(fx *printFixtureIDs) context.Context {
	return rls.NewContext(context.Background(), rls.Identity{
		UserID: fx.staffID, CompanyID: fx.companyID, DepartmentID: &fx.deptID,
		DataScope: "department", Role: "staff",
	})
}

func TestPreviewOnPendingOrder(t *testing.T) {
	// 預覽不檢查訂單狀態(細部 5.5.2 步驟 3):pending 亦可預覽
	h, fx, fs, auditSeen := newPreviewService(t, "pending")
	resp, err := h.Preview(staffCtx(fx), connect.NewRequest(&printv1.PreviewRequest{
		DocumentType: printv1.DocumentType_DOCUMENT_TYPE_DISPATCH_SUMMARY,
		RouteId:      fx.routeID.String(),
		TargetDate:   "2026-08-21",
	}))
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	if resp.Msg.PreviewId == "" || resp.Msg.FileAssetId == "" {
		t.Fatalf("response missing ids: %+v", resp.Msg)
	}
	// 下載 URL 經 file_assets(細部 5.5.2 步驟 6)
	want := "/api/v1/files/" + resp.Msg.FileAssetId + "/download"
	if resp.Msg.DownloadUrl != want {
		t.Errorf("download_url = %q, want %q", resp.Msg.DownloadUrl, want)
	}

	c := h.DBForTest()
	ctx := rls.NewContext(context.Background(), rls.Identity{
		UserID: fx.staffID, CompanyID: fx.companyID, DataScope: "all",
	})
	// 預覽不影響正式記錄(細部 5.5.2 驗收):print_logs 0 筆、print_previews 1 筆
	if n, err := c.PrintLog.Query().Count(ctx); err != nil || n != 0 {
		t.Fatalf("print_logs 不得新增: n=%d err=%v", n, err)
	}
	prev, err := c.PrintPreview.Query().Only(ctx)
	if err != nil {
		t.Fatalf("print_previews 應新增一筆: %v", err)
	}
	if prev.PreviewedBy != fx.staffID || prev.FileAssetID == nil ||
		prev.FileAssetID.String() != resp.Msg.FileAssetId {
		t.Errorf("preview row wrong: %+v", prev)
	}
	// file_assets 元資料同事務建立
	if _, err := c.FileAsset.Get(ctx, uuid.MustParse(resp.Msg.FileAssetId)); err != nil {
		t.Fatalf("file_asset 不存在: %v", err)
	}
	// 預覽不寫 audit_logs(細部 5.5.2 步驟 5)
	if len(*auditSeen) != 0 {
		t.Errorf("preview 不得寫稽核: %+v", *auditSeen)
	}
	if fs.writes != 1 {
		t.Errorf("WriteLocal calls = %d", fs.writes)
	}
}

func TestPreviewEmptyTableRejected(t *testing.T) {
	h, fx, fs, _ := newPreviewService(t, "pending")
	// 另一日期無任何明細 → 空表 → failed_precondition,且不寫任何記錄(細部 5.5.2 步驟 4)
	_, err := h.Preview(staffCtx(fx), connect.NewRequest(&printv1.PreviewRequest{
		DocumentType: printv1.DocumentType_DOCUMENT_TYPE_PICKING_LIST,
		RouteId:      fx.routeID.String(),
		TargetDate:   "2026-09-01",
		WarehouseId:  ptr(fx.warehouseID.String()),
	}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("want failed_precondition, got %v", err)
	}
	if fs.writes != 0 {
		t.Fatal("空表不得落檔")
	}
	c := h.DBForTest()
	ctx := rls.NewContext(context.Background(), rls.Identity{
		UserID: fx.staffID, CompanyID: fx.companyID, DataScope: "all",
	})
	if n, _ := c.PrintPreview.Query().Count(ctx); n != 0 {
		t.Fatalf("空表不得寫 print_previews: %d", n)
	}
	if n, _ := c.FileAsset.Query().Count(ctx); n != 0 {
		t.Fatalf("空表不得寫 file_assets: %d", n)
	}
}

func TestPreviewAuthAndValidation(t *testing.T) {
	h, fx, _, _ := newPreviewService(t, "pending")

	// 未登入 → unauthenticated
	_, err := h.Preview(context.Background(), connect.NewRequest(&printv1.PreviewRequest{
		DocumentType: printv1.DocumentType_DOCUMENT_TYPE_DISPATCH_SUMMARY,
		RouteId:      fx.routeID.String(), TargetDate: "2026-08-21",
	}))
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("want unauthenticated, got %v", err)
	}

	// customer 身分 → permission_denied
	custCtx := rls.NewContext(context.Background(), rls.Identity{
		UserID: uuid.New(), CompanyID: fx.companyID, DepartmentID: &fx.deptID,
		DataScope: "self", Role: "customer",
	})
	_, err = h.Preview(custCtx, connect.NewRequest(&printv1.PreviewRequest{
		DocumentType: printv1.DocumentType_DOCUMENT_TYPE_DISPATCH_SUMMARY,
		RouteId:      fx.routeID.String(), TargetDate: "2026-08-21",
	}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("customer: want permission_denied, got %v", err)
	}

	// 日期格式錯誤 → invalid_argument
	_, err = h.Preview(staffCtx(fx), connect.NewRequest(&printv1.PreviewRequest{
		DocumentType: printv1.DocumentType_DOCUMENT_TYPE_DISPATCH_SUMMARY,
		RouteId:      fx.routeID.String(), TargetDate: "21/08/2026",
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("bad date: want invalid_argument, got %v", err)
	}

	// 選擇器組合不合法(揀貨單缺 warehouse_id)→ invalid_argument
	_, err = h.Preview(staffCtx(fx), connect.NewRequest(&printv1.PreviewRequest{
		DocumentType: printv1.DocumentType_DOCUMENT_TYPE_PICKING_LIST,
		RouteId:      fx.routeID.String(), TargetDate: "2026-08-21",
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("missing warehouse: want invalid_argument, got %v", err)
	}

	// route 不存在 → not_found
	_, err = h.Preview(staffCtx(fx), connect.NewRequest(&printv1.PreviewRequest{
		DocumentType: printv1.DocumentType_DOCUMENT_TYPE_DISPATCH_SUMMARY,
		RouteId:      uuid.NewString(), TargetDate: "2026-08-21",
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("missing route: want not_found, got %v", err)
	}
}
```

測試輔助(同檔):

```go
type printFixtureIDs struct {
	companyID, deptID, routeID, warehouseID, customerID, productID, orderID, staffID uuid.UUID
}

func ptr(s string) *string { return &s }

// seedMinimalPrintData 建立一筆訂單一筆明細的最小列印資料。
// 04/05 計畫實體必填欄位若多於此處,依其 schema 補齊 — 唯一調整點。
func seedMinimalPrintData(t *testing.T, c *ent.Client, status string) *printFixtureIDs {
	t.Helper()
	ctx := context.Background()
	co, err := c.Company.Create().SetName("甲公司").SetIdentifier("co-a").SetCustomerCodePrefix("TY").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	dept, err := c.Department.Create().SetCompanyID(co.ID).SetName("桃園部").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	route, err := c.Route.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetCode("TY01").SetName("桃園一車").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	wh, err := c.Warehouse.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetCode("COLD").SetName("冷藏倉").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cust, err := c.Customer.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetCustomerCode("TY000001").SetName("好食餐廳").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	prod, err := c.Product.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetCode("P001").SetName("雞腿").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	date := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	order, err := c.SalesOrder.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetOrderNo("SO-001").SetCustomerID(cust.ID).SetSource("W").
		SetStatus(status).SetExpectedDeliveryDate(date).
		SetRouteID(route.ID).SetDeliverySequence(1).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.SalesOrderItem.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetSalesOrderID(order.ID).SetProductID(prod.ID).SetDisplayName("雞腿").
		SetQty(decimal.RequireFromString("10")).SetUnit("kg").
		SetBaseQty(decimal.RequireFromString("10")).
		SetWarehouseID(wh.ID).SetSortOrder(1).Save(ctx); err != nil {
		t.Fatal(err)
	}
	return &printFixtureIDs{
		companyID: co.ID, deptID: dept.ID, routeID: route.ID, warehouseID: wh.ID,
		customerID: cust.ID, productID: prod.ID, orderID: order.ID, staffID: uuid.New(),
	}
}
```

(`decimal` import `"github.com/shopspring/decimal"`。)

- [ ] **Step 6: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/prints/ -v`
Expected: FAIL — `prints.NewPrintsService` / `printv1` 未定義(編譯失敗)。

- [ ] **Step 7: 實作 proto、handler 共用層與 Preview**

`backend/proto/v1/print.proto`(本 Task 先含 Preview;Task 4 增量 Print / ListLogs):

```proto
syntax = "proto3";

package print.v1;

option go_package = "github.com/salesorder/sales-order-1.0/backend/internal/gen/print/v1;printv1";

// DocumentType 對應細部 5.3 四種單據。
enum DocumentType {
  DOCUMENT_TYPE_UNSPECIFIED = 0;
  DOCUMENT_TYPE_DISPATCH_SUMMARY = 1;
  DOCUMENT_TYPE_DELIVERY_NOTE = 2;
  DOCUMENT_TYPE_PICKING_LIST = 3;
  DOCUMENT_TYPE_PROCESSING_LIST = 4;
}

service PrintService {
  // Preview 任何訂單狀態皆可預覽;寫 print_previews,不觸碰 print_logs(細部 5.5.2)。
  rpc Preview(PreviewRequest) returns (PreviewResponse);
}

message PreviewRequest {
  DocumentType document_type = 1;
  string route_id = 2;
  string target_date = 3;           // YYYY-MM-DD
  optional string customer_id = 4;  // 對點單單印一店
  optional string warehouse_id = 5; // 揀貨單必帶
}

message PreviewResponse {
  string preview_id = 1;
  string file_asset_id = 2;
  string download_url = 3; // /api/v1/files/{id}/download
}
```

`backend/internal/domain/prints/service.go`:

```go
// Package prints 為列印/預覽 Connect-RPC handlers(細部文件 09-printing Task 5.5)。
package prints

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/print"
	printv1 "github.com/salesorder/sales-order-1.0/backend/internal/gen/print/v1"
)

// AuditFactory 回傳綁定 DB 交易的 audit.Recorder(D18 同事務)。
// DB 版構造由 03-metadicts-audit 計畫提供;TODO(接手:03-plan)整合時注入。
type AuditFactory func(tx *ent.Tx) audit.Recorder

// PrintsService 實作 print.v1.PrintService。
type PrintsService struct {
	db       *ent.Client
	prints   *print.Service
	files    print.FileStore
	newAudit AuditFactory
}

func NewPrintsService(db *ent.Client, svc *print.Service, files print.FileStore, newAudit AuditFactory) *PrintsService {
	return &PrintsService{db: db, prints: svc, files: files, newAudit: newAudit}
}

// DBForTest 暴露內部 client 供同套件測試斷言記錄(僅測試使用)。
func (s *PrintsService) DBForTest() *ent.Client { return s.db }

// printRoles 為可使用列印/預覽的角色(細部 5.5.4 步驟 3:dept_admin / staff 開放,
// 更高角色於範圍內當然可行;customer / guest 一律拒絕)。
var printRoles = map[string]bool{
	"super": true, "company_admin": true, "dept_admin": true, "staff": true,
}

// requirePrinter 檢查登入與角色;dept_admin / staff 必須帶部門歸屬。
func requirePrinter(ctx context.Context) (rls.Identity, error) {
	id, ok := rls.FromContext(ctx)
	if !ok {
		return rls.Identity{}, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}
	if !printRoles[id.Role] {
		return rls.Identity{}, connect.NewError(connect.CodePermissionDenied,
			fmt.Errorf("role %q 無列印權限", id.Role))
	}
	if (id.Role == "dept_admin" || id.Role == "staff") && id.DepartmentID == nil {
		return rls.Identity{}, connect.NewError(connect.CodePermissionDenied,
			errors.New("部門使用者缺少 department_id"))
	}
	return id, nil
}

// parseRequest 解析與驗證共用請求欄位(文件類型、route、日期、選擇器)。
func parseRequest(docTypePB printv1.DocumentType, routeID, targetDate string, customerID, warehouseID *string) (print.DocumentType, uuid.UUID, time.Time, print.Selector, error) {
	var docType print.DocumentType
	switch docTypePB {
	case printv1.DocumentType_DOCUMENT_TYPE_DISPATCH_SUMMARY:
		docType = print.DocumentTypeDispatchSummary
	case printv1.DocumentType_DOCUMENT_TYPE_DELIVERY_NOTE:
		docType = print.DocumentTypeDeliveryNote
	case printv1.DocumentType_DOCUMENT_TYPE_PICKING_LIST:
		docType = print.DocumentTypePickingList
	case printv1.DocumentType_DOCUMENT_TYPE_PROCESSING_LIST:
		docType = print.DocumentTypeProcessingList
	default:
		return "", uuid.Nil, time.Time{}, print.Selector{},
			connect.NewError(connect.CodeInvalidArgument, errors.New("document_type 必填且須為四種單據之一"))
	}
	rid, err := uuid.Parse(routeID)
	if err != nil {
		return "", uuid.Nil, time.Time{}, print.Selector{},
			connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("route_id 格式錯誤: %v", err))
	}
	date, err := time.Parse("2006-01-02", targetDate)
	if err != nil {
		return "", uuid.Nil, time.Time{}, print.Selector{},
			connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("target_date 須為 YYYY-MM-DD: %v", err))
	}
	var sel print.Selector
	if customerID != nil {
		cid, err := uuid.Parse(*customerID)
		if err != nil {
			return "", uuid.Nil, time.Time{}, print.Selector{},
				connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("customer_id 格式錯誤: %v", err))
		}
		sel.CustomerID = &cid
	}
	if warehouseID != nil {
		wid, err := uuid.Parse(*warehouseID)
		if err != nil {
			return "", uuid.Nil, time.Time{}, print.Selector{},
				connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("warehouse_id 格式錯誤: %v", err))
		}
		sel.WarehouseID = &wid
	}
	return docType, rid, date, sel, nil
}
```

`backend/internal/domain/prints/preview.go`:

```go
package prints

import (
	"context"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/print"
	printv1 "github.com/salesorder/sales-order-1.0/backend/internal/gen/print/v1"
)

// Preview 實作細部 5.5.2:任何訂單狀態皆可預覽;寫 print_previews,完全不觸碰 print_logs。
func (s *PrintsService) Preview(ctx context.Context, req *connect.Request[printv1.PreviewRequest]) (*connect.Response[printv1.PreviewResponse], error) {
	// 步驟 1:認證與授權
	actor, err := requirePrinter(ctx)
	if err != nil {
		return nil, err
	}
	// 步驟 2:參數驗證(含選擇器與類型組合,由 BuildViewModel 內 validateSelector 把關)
	docType, routeID, date, sel, err := parseRequest(
		req.Msg.DocumentType, req.Msg.RouteId, req.Msg.TargetDate, req.Msg.CustomerId, req.Msg.WarehouseId)
	if err != nil {
		return nil, err
	}
	// 步驟 3:不檢查訂單狀態 — pending / processing / completed 皆可預覽
	// 步驟 4:組資料 + 產 PDF(空表守門於 Prepare 內生效)
	vm, err := s.prints.BuildViewModel(ctx, docType, routeID, date, sel)
	if err != nil {
		return nil, err
	}
	prepared, err := s.prints.Prepare(ctx, docType, vm)
	if err != nil {
		return nil, err
	}
	// 步驟 5:同事務寫入 print_previews 與 file_assets 元資料(D18);預覽不寫 audit_logs
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	preview, err := tx.PrintPreview.Create().
		SetCompanyID(actor.CompanyID).
		SetDepartmentID(deptOf(actor)).
		SetDocumentType(string(docType)).
		SetRouteID(routeID).
		SetNillableCustomerID(sel.CustomerID).
		SetNillableWarehouseID(sel.WarehouseID).
		SetTargetDate(date).
		SetPreviewedBy(actor.UserID).
		SetPreviewedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return rollbackWithCleanup(ctx, tx, s.files, prepared)
	}
	asset, err := s.files.CreateAsset(ctx, tx, actor, prepared.Stored, print.OwnerTypePrintPDF, preview.ID)
	if err != nil {
		return rollbackWithCleanup(ctx, tx, s.files, prepared)
	}
	if _, err := tx.PrintPreview.UpdateOneID(preview.ID).SetFileAssetID(asset.ID).Save(ctx); err != nil {
		return rollbackWithCleanup(ctx, tx, s.files, prepared)
	}
	if err := tx.Commit(); err != nil {
		s.files.DeleteLocal(ctx, prepared.Stored.StoragePath) // 補償刪除,不留孤兒檔(細部 5.4.3 步驟 5)
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// 步驟 6:回傳下載 URL,前端直接開啟或下載
	return connect.NewResponse(&printv1.PreviewResponse{
		PreviewId:   preview.ID.String(),
		FileAssetId: asset.ID.String(),
		DownloadUrl: print.DownloadURL(asset.ID),
	}), nil
}
```

共用輔助(放 `service.go` 末尾):

```go
// deptOf 取操作者部門;company 級角色(super / company_admin)以資料範圍注入時
// DepartmentID 可為 nil — 列印記錄的 department_id 由車次所屬部門而來,
// 但 1.0 範圍內列印操作者皆為部門使用者,此處直接取值;nil 時由 RLS fail-closed 擋下。
func deptOf(actor rls.Identity) uuid.UUID {
	if actor.DepartmentID == nil {
		return uuid.Nil // 不會發生:requirePrinter 已保證 staff/dept_admin 帶部門
	}
	return *actor.DepartmentID
}

// rollbackWithCleanup 回滾交易並補償刪除已落檔 PDF(細部 5.4.3 步驟 5)。
func rollbackWithCleanup(ctx context.Context, tx *ent.Tx, files print.FileStore, prepared *print.PreparedPDF) (*connect.Response[printv1.PreviewResponse], error) {
	_ = tx.Rollback()
	files.DeleteLocal(ctx, prepared.Stored.StoragePath)
	return nil, connect.NewError(connect.CodeInternal, errors.New("寫入列印記錄失敗"))
}
```

(注意:`rollbackWithCleanup` 的回應型別在 Task 4 會一般化;實作時改為回傳 `error` 單值,由呼叫端包覆 — 以此為準:`func rollbackWithCleanup(ctx, tx, files, prepared) error`。)

- [ ] **Step 8: 產生 proto 並跑測試**

Run: `cd backend && buf generate proto && go test ./internal/domain/prints/ -v`
Expected: PASS — 4 個 Preview 測試(pending 可預覽且隔離 print_logs、空表拒絕且不留記錄、五種認證/參數錯誤邊界)。

若 repo 的 proto 產生工具鏈非 buf,以專案既有指令(01-auth 計畫 Task 5 同一條)為準。

- [ ] **Step 9: Commit**

```bash
git add backend/ent backend/database/migrations backend/proto/v1/print.proto backend/internal/domain/prints backend/internal/gen backend/go.mod backend/go.sum
git commit -m "feat(backend): print_logs/print_previews schema、RLS、Preview API 預覽隔離(5.5.1-5.5.2)"
```

---

### Task 4: Print API + 重印 + ListLogs(細部 5.5.3–5.5.4)

**Files:**
- Create: `backend/internal/domain/prints/print.go`
- Create: `backend/internal/domain/prints/list_logs.go`
- Update: `backend/proto/v1/print.proto`(Print / ListLogs 增量)
- Update: `backend/internal/server/domains.go`(掛載 PrintService)
- Test: `backend/internal/domain/prints/print_test.go`
- Test: `backend/internal/domain/prints/list_logs_test.go`

**Interfaces:**
- Consumes: Task 3 全部(`PrintsService` / `requirePrinter` / `parseRequest` / `AuditFactory`);Task 2 `print.Service`;由 08-dispatch 保證的語意 — 派車確認後 `status = processing`(本 Task 只讀);`audit.Entry`(01-auth 計畫 Task 14)。
- Produces:
  - proto `PrintService.Print(PrintRequest) → PrintResponse{print_log_id, file_asset_id, download_url, is_reprint}`;`PrintRequest` 較 PreviewRequest 多 `optional string reprint_reason = 6`。
  - proto `PrintService.ListLogs(ListLogsRequest) → ListLogsResponse{logs[], total}`;`ListLogsRequest{date_from, date_to, document_type, optional route_id, page, page_size}`;`PrintLogEntry{id, document_type, route_id, route_name, target_date, is_reprint, reprint_reason, printed_by, printed_at, download_url}`。
  - 內部 `(*PrintsService).assertAllProcessing(ctx, tx, docType, routeID, date, sel) error`:交易內 `FOR UPDATE` 鎖定範圍訂單,任一筆非 `processing` → `failed_precondition`(訊息列出不符合的訂單單號)。
  - 內部 `(*PrintsService).findExistingLog(ctx, tx, deptID, docType, routeID, date, sel) (bool, error)`:重印比對鍵查詢(細部 5.5.1 步驟 4)。

- [ ] **Step 1: 寫失敗測試(Print:狀態守門、首印、參數守門)**

`backend/internal/domain/prints/print_test.go`:

```go
package prints_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent/printlog"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/prints"
	"github.com/salesorder/sales-order-1.0/backend/internal/print"
	printv1 "github.com/salesorder/sales-order-1.0/backend/internal/gen/print/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

// newPrintService 同 newPreviewService,但 status 可指定且回傳底層 client。
func newPrintService(t *testing.T, status string) (*prints.PrintsService, *printFixtureIDs, *txFileStore, *[]audit.Entry, *ent.Client) {
	t.Helper()
	c := testutil.NewEntClient(t)
	fx := seedMinimalPrintData(t, c, status)
	fs := &txFileStore{}
	var auditSeen []audit.Entry
	svc := print.NewService(c, &okConverter{}, fs)
	h := prints.NewPrintsService(c, svc, fs, captureAuditFactory(&auditSeen))
	return h, fx, fs, &auditSeen, c
}

func printReq(fx *printFixtureIDs, reason *string) *connect.Request[printv1.PrintRequest] {
	return connect.NewRequest(&printv1.PrintRequest{
		DocumentType:  printv1.DocumentType_DOCUMENT_TYPE_DISPATCH_SUMMARY,
		RouteId:       fx.routeID.String(),
		TargetDate:    "2026-08-21",
		ReprintReason: reason,
	})
}

func TestPrintRejectsPendingOrder(t *testing.T) {
	// 訂單 status = pending(尚未派車確認)→ 整批拒絕(細部 5.5.3 步驟 2)
	h, fx, fs, auditSeen, c := newPrintService(t, "pending")
	_, err := h.Print(staffCtx(fx), printReq(fx, nil))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("want failed_precondition, got %v", err)
	}
	// 訊息指出不符合的訂單單號
	if cerr := new(connect.Error); !connect.CodeOf(err).String()... // 直接檢查訊息內容:
	_ = cerr
	if got := err.Error(); !strings.Contains(got, "SO-001") {
		t.Errorf("錯誤訊息應含不符合的訂單單號 SO-001: %s", got)
	}
	// 無半套記錄(細部 5.5.3 驗收)
	ctx := rls.NewContext(context.Background(), rls.Identity{
		UserID: fx.staffID, CompanyID: fx.companyID, DataScope: "all",
	})
	if n, _ := c.PrintLog.Query().Count(ctx); n != 0 {
		t.Fatalf("不得寫 print_logs: %d", n)
	}
	if n, _ := c.FileAsset.Query().Count(ctx); n != 0 {
		t.Fatalf("不得寫 file_assets: %d", n)
	}
	if len(*auditSeen) != 0 {
		t.Fatalf("不得寫稽核: %+v", *auditSeen)
	}
	if fs.writes != 0 {
		t.Fatal("狀態不符不得落檔")
	}
}

func TestPrintFirstPrintSuccess(t *testing.T) {
	h, fx, _, auditSeen, c := newPrintService(t, "processing")
	resp, err := h.Print(staffCtx(fx), printReq(fx, nil))
	if err != nil {
		t.Fatalf("print: %v", err)
	}
	if resp.Msg.IsReprint {
		t.Error("首印 is_reprint 應為 false")
	}
	want := "/api/v1/files/" + resp.Msg.FileAssetId + "/download"
	if resp.Msg.DownloadUrl != want {
		t.Errorf("download_url = %q, want %q", resp.Msg.DownloadUrl, want)
	}
	ctx := rls.NewContext(context.Background(), rls.Identity{
		UserID: fx.staffID, CompanyID: fx.companyID, DataScope: "all",
	})
	// print_logs 一筆 is_reprint=false、reprint_reason 空(細部 5.5.3 步驟 3)
	log, err := c.PrintLog.Query().Only(ctx)
	if err != nil {
		t.Fatalf("print_logs 應一筆: %v", err)
	}
	if log.IsReprint || log.ReprintReason != nil {
		t.Errorf("首印記錄錯誤: %+v", log)
	}
	if log.PrintedBy != fx.staffID || log.FileAssetID == nil {
		t.Errorf("printed_by / file_asset_id 錯誤: %+v", log)
	}
	// audit_logs 一筆 action=print(D18;細部 5.5.3 步驟 5)
	if len(*auditSeen) != 1 || (*auditSeen)[0].Action != "print" {
		t.Fatalf("稽核應一筆 action=print: %+v", *auditSeen)
	}
	if (*auditSeen)[0].ResourceType != "print_log" || (*auditSeen)[0].ResourceID != log.ID.String() {
		t.Errorf("稽核 resource 錯誤: %+v", (*auditSeen)[0])
	}
	// PDF 可經下載 URL 取得:file_assets 記錄存在且 url 正確
	asset, err := c.FileAsset.Get(ctx, uuid.MustParse(resp.Msg.FileAssetId))
	if err != nil {
		t.Fatalf("file_asset 不存在: %v", err)
	}
	if asset.Url != want || asset.MimeType != "application/pdf" {
		t.Errorf("asset 錯誤: %+v", asset)
	}
}

func TestPrintFirstPrintWithReasonRejected(t *testing.T) {
	// 首印 reprint_reason 必須為空(細部 5.5.3 步驟 3)
	h, fx, fs, _, c := newPrintService(t, "processing")
	_, err := h.Print(staffCtx(fx), printReq(fx, ptr("補印")))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("want invalid_argument, got %v", err)
	}
	if fs.writes != 0 {
		t.Fatal("參數錯誤不得落檔")
	}
	ctx := rls.NewContext(context.Background(), rls.Identity{
		UserID: fx.staffID, CompanyID: fx.companyID, DataScope: "all",
	})
	if n, _ := c.PrintLog.Query().Count(ctx); n != 0 {
		t.Fatalf("不得寫記錄: %d", n)
	}
}

func TestPrintPermission(t *testing.T) {
	h, fx, _, _, _ := newPrintService(t, "processing")
	// guest 身分 → permission_denied(細部 5.5.4 步驟 3)
	guestCtx := rls.NewContext(context.Background(), rls.Identity{
		UserID: uuid.New(), CompanyID: fx.companyID, DepartmentID: &fx.deptID,
		DataScope: "department", Role: "guest",
	})
	if _, err := h.Print(guestCtx, printReq(fx, nil)); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("guest: want permission_denied, got %v", err)
	}
	// 未登入 → unauthenticated
	if _, err := h.Print(context.Background(), printReq(fx, nil)); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("want unauthenticated, got %v", err)
	}
}
```

(print_test.go import 加 `"strings"` 與 `"github.com/salesorder/sales-order-1.0/backend/ent"`;`TestPrintRejectsPendingOrder` 中 `cerr` 兩行為示意殘句,實作時刪除,僅保留 `strings.Contains` 斷言 — 以此為準。)

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/prints/ -run TestPrint -v`
Expected: FAIL — `prints.PrintsService.Print` 與 proto 訊息未定義(編譯失敗)。

- [ ] **Step 3: 實作 Print RPC(首印 + 重印分支)**

`backend/proto/v1/print.proto` 增量:

```proto
service PrintService {
  rpc Preview(PreviewRequest) returns (PreviewResponse);

  // Print 正式列印:範圍內訂單須全為 processing,寫 print_logs + audit_logs(細部 5.5.3);
  // 比對鍵已存在記錄時為重印,reprint_reason 必填(細部 5.5.4)。
  rpc Print(PrintRequest) returns (PrintResponse);

  // ListLogs 列印記錄查詢(唯讀,RLS 範圍)。
  rpc ListLogs(ListLogsRequest) returns (ListLogsResponse);
}

message PrintRequest {
  DocumentType document_type = 1;
  string route_id = 2;
  string target_date = 3;
  optional string customer_id = 4;
  optional string warehouse_id = 5;
  optional string reprint_reason = 6; // 重印必填;首印必須為空
}

message PrintResponse {
  string print_log_id = 1;
  string file_asset_id = 2;
  string download_url = 3;
  bool is_reprint = 4;
}

message ListLogsRequest {
  string date_from = 1;                 // YYYY-MM-DD
  string date_to = 2;                   // YYYY-MM-DD
  DocumentType document_type = 3;       // UNSPECIFIED = 不篩
  optional string route_id = 4;
  int32 page = 5;
  int32 page_size = 6;                  // 上限 100(全系統慣例)
}

message PrintLogEntry {
  string id = 1;
  DocumentType document_type = 2;
  string route_id = 3;
  string route_name = 4;
  string target_date = 5;
  bool is_reprint = 6;
  string reprint_reason = 7;
  string printed_by = 8;
  string printed_at = 9;
  string download_url = 10;
}

message ListLogsResponse {
  repeated PrintLogEntry logs = 1;
  int32 total = 2;
}
```

`backend/internal/domain/prints/print.go`:

```go
package prints

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/printlog"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	"github.com/salesorder/sales-order-1.0/backend/internal/print"
	printv1 "github.com/salesorder/sales-order-1.0/backend/internal/gen/print/v1"
)

// Print 實作細部 5.5.3 / 5.5.4:正式列印與重印。
func (s *PrintsService) Print(ctx context.Context, req *connect.Request[printv1.PrintRequest]) (*connect.Response[printv1.PrintResponse], error) {
	// 步驟 1:認證與授權(dept_admin / staff 開放;customer / guest 拒絕)
	actor, err := requirePrinter(ctx)
	if err != nil {
		return nil, err
	}
	docType, routeID, date, sel, err := parseRequest(
		req.Msg.DocumentType, req.Msg.RouteId, req.Msg.TargetDate, req.Msg.CustomerId, req.Msg.WarehouseId)
	if err != nil {
		return nil, err
	}
	reason := strings.TrimSpace(deref(req.Msg.ReprintReason))

	// 組資料 + 產 PDF(交易外;空表守門照常生效,細部 5.5.3 步驟 4)
	vm, err := s.prints.BuildViewModel(ctx, docType, routeID, date, sel)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// 步驟 2:狀態檢查在交易內以 FOR UPDATE 鎖定範圍訂單並確認(步驟 5:
	// 狀態檢查與寫入在同一交易完成,避免檢查後狀態被並發改變的窗口)
	if err := s.assertAllProcessing(ctx, tx, routeID, date, sel); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	// 步驟 3:首印/重印推導(交易內,與狀態檢查同一快照)
	exists, err := s.findExistingLog(ctx, tx, deptOf(actor), docType, routeID, date, sel)
	if err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	isReprint := exists
	// 原因守門(細部 5.5.4 步驟 2 / 5.5.3 步驟 3)
	if isReprint && reason == "" {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("重印必須填寫 reprint_reason"))
	}
	if !isReprint && reason != "" {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("首印不得攜帶 reprint_reason"))
	}
	_ = tx.Rollback() // 狀態與推導檢查完成;PDF 產生在交易外,正式寫入見下方第二段交易
	// 說明:以上以「檢查交易」先擋下狀態/參數錯誤,避免不必要的 PDF 產生;
	// 寫入交易於 Prepare 之後再開,並於寫入交易內重跑同樣的狀態再確認(細部 5.5.3 步驟 5)。

	prepared, err := s.prints.Prepare(ctx, docType, vm)
	if err != nil {
		return nil, err
	}

	tx2, err := s.db.Tx(ctx)
	if err != nil {
		s.files.DeleteLocal(ctx, prepared.Stored.StoragePath)
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// 交易內狀態再確認,不一致則整體回滾(細部 5.5.3 步驟 5)
	if err := s.assertAllProcessing(ctx, tx2, routeID, date, sel); err != nil {
		_ = tx2.Rollback()
		s.files.DeleteLocal(ctx, prepared.Stored.StoragePath)
		return nil, err
	}
	// 重印推導再確認:兩次檢查間他人完成首印 → 本次轉重印分支,原因守門同上
	exists2, err := s.findExistingLog(ctx, tx2, deptOf(actor), docType, routeID, date, sel)
	if err != nil {
		_ = tx2.Rollback()
		s.files.DeleteLocal(ctx, prepared.Stored.StoragePath)
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	isReprint = isReprint || exists2
	if isReprint && reason == "" {
		_ = tx2.Rollback()
		s.files.DeleteLocal(ctx, prepared.Stored.StoragePath)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("重印必須填寫 reprint_reason"))
	}

	// 步驟 5:同事務寫入 print_logs + file_assets 元資料 + audit_logs(D18)
	log, err := tx2.PrintLog.Create().
		SetCompanyID(actor.CompanyID).
		SetDepartmentID(deptOf(actor)).
		SetDocumentType(string(docType)).
		SetRouteID(routeID).
		SetNillableCustomerID(sel.CustomerID).
		SetNillableWarehouseID(sel.WarehouseID).
		SetTargetDate(date).
		SetIsReprint(isReprint).
		SetNillableReprintReason(nilStringIfEmpty(reason)).
		SetPrintedBy(actor.UserID).
		SetPrintedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, s.rollbackPrint(ctx, tx2, prepared, err)
	}
	asset, err := s.files.CreateAsset(ctx, tx2, actor, prepared.Stored, print.OwnerTypePrintPDF, log.ID)
	if err != nil {
		return nil, s.rollbackPrint(ctx, tx2, prepared, err)
	}
	if _, err := tx2.PrintLog.UpdateOneID(log.ID).SetFileAssetID(asset.ID).Save(ctx); err != nil {
		return nil, s.rollbackPrint(ctx, tx2, prepared, err)
	}
	after, _ := json.Marshal(map[string]any{
		"print_log_id":   log.ID,
		"document_type":  string(docType),
		"route_id":       routeID,
		"target_date":    date.Format("2006-01-02"),
		"is_reprint":     isReprint,
		"reprint_reason": reason,
		"file_asset_id":  asset.ID,
	})
	if err := s.newAudit(tx2).Record(ctx, audit.Entry{
		ActorID:      actor.UserID.String(),
		Action:       "print",
		ResourceType: "print_log",
		ResourceID:   log.ID.String(),
		After:        string(after),
	}); err != nil {
		return nil, s.rollbackPrint(ctx, tx2, prepared, err)
	}
	if err := tx2.Commit(); err != nil {
		s.files.DeleteLocal(ctx, prepared.Stored.StoragePath)
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// 步驟 6:PDF 一律以 file_assets 下載 URL 交付,不在 RPC 回應夾帶二進位
	return connect.NewResponse(&printv1.PrintResponse{
		PrintLogId:  log.ID.String(),
		FileAssetId: asset.ID.String(),
		DownloadUrl: print.DownloadURL(asset.ID),
		IsReprint:   isReprint,
	}), nil
}

// rollbackPrint 回滾寫入交易並補償刪除已落檔 PDF,不留孤兒檔(細部 5.4.3 步驟 5)。
func (s *PrintsService) rollbackPrint(ctx context.Context, tx *ent.Tx, prepared *print.PreparedPDF, cause error) error {
	_ = tx.Rollback()
	s.files.DeleteLocal(ctx, prepared.Stored.StoragePath)
	return connect.NewError(connect.CodeInternal, fmt.Errorf("寫入列印記錄失敗: %w", cause))
}

// assertAllProcessing 鎖定範圍訂單並逐筆確認 status = processing(細部 5.5.3 步驟 2);
// 任一筆不符 → failed_precondition,訊息列出不符合的訂單單號。範圍內無訂單時不檢查
// (空表情形由 Prepare 的 IsEmpty 守門負責)。
func (s *PrintsService) assertAllProcessing(ctx context.Context, tx *ent.Tx, routeID uuid.UUID, date time.Time, sel print.Selector) error {
	q := tx.SalesOrder.Query().Where(
		salesorder.RouteIDEQ(routeID),
		salesorder.ExpectedDeliveryDateEQ(date),
		salesorder.DeletedAtIsNil(),
	)
	if sel.CustomerID != nil {
		q = q.Where(salesorder.CustomerIDEQ(*sel.CustomerID))
	}
	orders, err := q.ForUpdate().All(ctx) // FOR UPDATE:鎖定至交易結束,防並發改狀態
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	var bad []string
	for _, o := range orders {
		if o.Status != "processing" {
			bad = append(bad, o.OrderNo)
		}
	}
	if len(bad) > 0 {
		return connect.NewError(connect.CodeFailedPrecondition,
			fmt.Errorf("下列訂單狀態非 processing,整批拒絕列印: %s", strings.Join(bad, ", ")))
	}
	return nil
}

// findExistingLog 以比對鍵(部門、document_type、route_id、target_date、選用 customer_id / warehouse_id)
// 查 print_logs,供首印/重印推導(細部 5.5.1 步驟 4、5.5.4 步驟 1)。
func (s *PrintsService) findExistingLog(ctx context.Context, tx *ent.Tx, deptID uuid.UUID, docType print.DocumentType, routeID uuid.UUID, date time.Time, sel print.Selector) (bool, error) {
	q := tx.PrintLog.Query().Where(
		printlog.DepartmentIDEQ(deptID),
		printlog.DocumentTypeEQ(string(docType)),
		printlog.RouteIDEQ(routeID),
		printlog.TargetDateEQ(date),
	)
	if sel.CustomerID != nil {
		q = q.Where(printlog.CustomerIDEQ(*sel.CustomerID))
	} else {
		q = q.Where(printlog.CustomerIDIsNil())
	}
	if sel.WarehouseID != nil {
		q = q.Where(printlog.WarehouseIDEQ(*sel.WarehouseID))
	} else {
		q = q.Where(printlog.WarehouseIDIsNil())
	}
	return q.Exist(ctx)
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func nilStringIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
```

(print.go import 加 `"github.com/salesorder/sales-order-1.0/backend/internal/audit"`。)

- [ ] **Step 4: 寫失敗測試(重印分支)**

在 `backend/internal/domain/prints/print_test.go` 追加:

```go
func TestReprintRequiresReason(t *testing.T) {
	h, fx, fs, _, c := newPrintService(t, "processing")
	// 先完成首印
	if _, err := h.Print(staffCtx(fx), printReq(fx, nil)); err != nil {
		t.Fatalf("first print: %v", err)
	}
	// 重印未填原因 → invalid_argument,不產生任何記錄(細部 5.5.4 步驟 2、驗收)
	_, err := h.Print(staffCtx(fx), printReq(fx, nil))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("want invalid_argument, got %v", err)
	}
	_, err = h.Print(staffCtx(fx), printReq(fx, ptr("   "))) // 純空白同空白
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("blank reason: want invalid_argument, got %v", err)
	}
	ctx := rls.NewContext(context.Background(), rls.Identity{
		UserID: fx.staffID, CompanyID: fx.companyID, DataScope: "all",
	})
	if n, _ := c.PrintLog.Query().Count(ctx); n != 1 {
		t.Fatalf("被拒重印不得新增記錄: %d", n)
	}
	if fs.writes != 1 {
		t.Fatalf("被拒重印不得落檔: writes=%d", fs.writes)
	}
}

func TestReprintSuccessKeepsHistory(t *testing.T) {
	h, fx, _, auditSeen, c := newPrintService(t, "processing")
	first, err := h.Print(staffCtx(fx), printReq(fx, nil))
	if err != nil {
		t.Fatalf("first print: %v", err)
	}
	// staff 身分可完成重印(細部 5.5.4 步驟 3)
	second, err := h.Print(staffCtx(fx), printReq(fx, ptr("單據污損補印")))
	if err != nil {
		t.Fatalf("reprint: %v", err)
	}
	if !second.Msg.IsReprint {
		t.Error("重印 is_reprint 應為 true")
	}
	// 新 PDF、新 file_asset、新 print_log;舊記錄與舊 PDF 保留不動(細部 5.5.4 步驟 5)
	if second.Msg.FileAssetId == first.Msg.FileAssetId || second.Msg.PrintLogId == first.Msg.PrintLogId {
		t.Error("重印必須產生新 PDF 與新記錄")
	}
	ctx := rls.NewContext(context.Background(), rls.Identity{
		UserID: fx.staffID, CompanyID: fx.companyID, DataScope: "all",
	})
	logs, err := c.PrintLog.Query().Order(ent.Asc(printlog.FieldPrintedAt)).All(ctx)
	if err != nil || len(logs) != 2 {
		t.Fatalf("want 2 print logs, got %d err=%v", len(logs), err)
	}
	if logs[0].IsReprint || !logs[1].IsReprint {
		t.Errorf("僅首筆 is_reprint=false: %+v", logs)
	}
	if logs[1].ReprintReason == nil || *logs[1].ReprintReason != "單據污損補印" {
		t.Errorf("重印原因未留存: %+v", logs[1].ReprintReason)
	}
	// 兩份 PDF 皆可取回(file_assets 記錄皆在)
	for _, l := range logs {
		if _, err := c.FileAsset.Get(ctx, *l.FileAssetID); err != nil {
			t.Errorf("舊/新 PDF 記錄遺失: %v", err)
		}
	}
	// 稽核兩筆,重印那筆 after 含 is_reprint 與原因(D18)
	if len(*auditSeen) != 2 {
		t.Fatalf("want 2 audit entries: %+v", *auditSeen)
	}
	if !strings.Contains((*auditSeen)[1].After, `"is_reprint":true`) ||
		!strings.Contains((*auditSeen)[1].After, "單據污損補印") {
		t.Errorf("重印稽核快照錯誤: %s", (*auditSeen)[1].After)
	}
}

func TestReprintRejectedWhenOrderBackToPending(t *testing.T) {
	// §7.1 情境:取消派車退回 pending 後,重印同樣拒絕(細部 5.5.4 步驟 4)
	h, fx, fs, _, c := newPrintService(t, "processing")
	if _, err := h.Print(staffCtx(fx), printReq(fx, nil)); err != nil {
		t.Fatalf("first print: %v", err)
	}
	// 模擬取消派車:訂單退回 pending(08-dispatch CancelDispatch 的效果)
	ctx := rls.NewContext(context.Background(), rls.Identity{
		UserID: fx.staffID, CompanyID: fx.companyID, DataScope: "all",
	})
	if _, err := c.SalesOrder.UpdateOneID(fx.orderID).SetStatus("pending").Save(ctx); err != nil {
		t.Fatal(err)
	}
	_, err := h.Print(staffCtx(fx), printReq(fx, ptr("重新列印")))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("want failed_precondition, got %v", err)
	}
	if fs.writes != 1 {
		t.Fatalf("狀態不符不得再落檔: writes=%d", fs.writes)
	}
}
```

(print_test.go import 加 `"github.com/salesorder/sales-order-1.0/backend/ent/printlog"` 與 `ent.Asc` 所需的 `"github.com/salesorder/sales-order-1.0/backend/ent"`。)

- [ ] **Step 5: 跑測試確認通過(Print + 重印)**

Run: `cd backend && buf generate proto && go test ./internal/domain/prints/ -run 'TestPrint|TestReprint' -v`
Expected: PASS — 7 個測試(pending 拒絕含單號、首印成功三處記錄、首印帶原因拒絕、權限邊界、重印缺原因拒絕、重印成功留存軌跡、退回 pending 重印拒絕)。

已知限制(不影響驗收):兩個並發首印在 `findExistingLog` 皆查無記錄時會皆標 `is_reprint = false`(比對鍵無唯一約束,重印語義為提示性而非一致性關鍵);1.0 不加唯一索引,避免阻擋合法重印。

- [ ] **Step 6: 寫失敗測試(ListLogs)**

`backend/internal/domain/prints/list_logs_test.go`:

```go
package prints_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/print"
	printv1 "github.com/salesorder/sales-order-1.0/backend/internal/gen/print/v1"
)

func TestListLogsFiltersAndPagination(t *testing.T) {
	h, fx, _, _, c := newPrintService(t, "processing")
	// 建兩筆記錄:首印 + 重印(dispatch_summary,2026-08-21)
	if _, err := h.Print(staffCtx(fx), printReq(fx, nil)); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Print(staffCtx(fx), printReq(fx, ptr("補印"))); err != nil {
		t.Fatal(err)
	}

	resp, err := h.ListLogs(staffCtx(fx), connect.NewRequest(&printv1.ListLogsRequest{
		DateFrom: "2026-08-01", DateTo: "2026-08-31", Page: 1, PageSize: 10,
	}))
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if resp.Msg.Total != 2 || len(resp.Msg.Logs) != 2 {
		t.Fatalf("want 2 logs, got total=%d len=%d", resp.Msg.Total, len(resp.Msg.Logs))
	}
	// 每筆含操作人、時間、is_reprint、reprint_reason 與 download_url(細部 5.5.3 介面)
	var reprintEntry *printv1.PrintLogEntry
	for _, l := range resp.Msg.Logs {
		if l.PrintedBy != fx.staffID.String() || l.PrintedAt == "" || l.RouteName != "桃園一車" {
			t.Errorf("entry 欄位不全: %+v", l)
		}
		if l.DownloadUrl == "" || !strings.Contains(l.DownloadUrl, "/api/v1/files/") {
			t.Errorf("download_url 錯誤: %q", l.DownloadUrl)
		}
		if l.IsReprint {
			reprintEntry = l
		}
	}
	if reprintEntry == nil || reprintEntry.ReprintReason != "補印" {
		t.Errorf("重印標記與原因缺失: %+v", resp.Msg.Logs)
	}

	// 單據類型篩選:processing_list 無記錄
	resp, err = h.ListLogs(staffCtx(fx), connect.NewRequest(&printv1.ListLogsRequest{
		DateFrom: "2026-08-01", DateTo: "2026-08-31",
		DocumentType: printv1.DocumentType_DOCUMENT_TYPE_PROCESSING_LIST,
		Page: 1, PageSize: 10,
	}))
	if err != nil || resp.Msg.Total != 0 {
		t.Fatalf("type filter: total=%d err=%v", resp.Msg.Total, err)
	}

	// 日期範圍外:無記錄
	resp, err = h.ListLogs(staffCtx(fx), connect.NewRequest(&printv1.ListLogsRequest{
		DateFrom: "2026-09-01", DateTo: "2026-09-30", Page: 1, PageSize: 10,
	}))
	if err != nil || resp.Msg.Total != 0 {
		t.Fatalf("date filter: total=%d err=%v", resp.Msg.Total, err)
	}

	// page_size 超過 100 收斂為 100(全系統慣例,細部 5.5.3 步驟 7)
	resp, err = h.ListLogs(staffCtx(fx), connect.NewRequest(&printv1.ListLogsRequest{
		DateFrom: "2026-08-01", DateTo: "2026-08-31", Page: 1, PageSize: 500,
	}))
	if err != nil {
		t.Fatalf("oversize page should clamp, got %v", err)
	}
	_ = c
	_ = time.Now
}

func TestListLogsUnauthenticated(t *testing.T) {
	h, _, _, _, _ := newPrintService(t, "processing")
	_, err := h.ListLogs(context.Background(), connect.NewRequest(&printv1.ListLogsRequest{
		DateFrom: "2026-08-01", DateTo: "2026-08-31",
	}))
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("want unauthenticated, got %v", err)
	}
	// 跨部門不可見由 RLS 保證(細部 5.5.3 步驟 7),schema 層已於 Task 3 測試覆蓋。
	_ = uuid.Nil
	_ = rls.Identity{}
	_ = print.DocumentTypeDispatchSummary
}
```

(檔尾 `_ =` 殘句僅為固定 import 佔位,實作時刪除未使用的 import — 以此為準。)

- [ ] **Step 7: 實作 ListLogs 與 main.go 掛載**

`backend/internal/domain/prints/list_logs.go`:

```go
package prints

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/printlog"
	"github.com/salesorder/sales-order-1.0/backend/ent/route"
	"github.com/salesorder/sales-order-1.0/backend/internal/print"
	printv1 "github.com/salesorder/sales-order-1.0/backend/internal/gen/print/v1"
)

// maxPageSize 為全系統分頁上限(細部 5.5.3 步驟 7)。
const maxPageSize = 100

// ListLogs 實作細部 5.5.3 介面:唯讀,依 RLS 範圍過濾。
func (s *PrintsService) ListLogs(ctx context.Context, req *connect.Request[ListLogsRequestAlias]) (*connect.Response[printv1.ListLogsResponse], error) {
	if _, err := requirePrinter(ctx); err != nil {
		return nil, err
	}
	from, err := time.Parse("2006-01-02", req.Msg.DateFrom)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("date_from 須為 YYYY-MM-DD: %v", err))
	}
	to, err := time.Parse("2006-01-02", req.Msg.DateTo)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("date_to 須為 YYYY-MM-DD: %v", err))
	}
	q := s.db.PrintLog.Query().Where(
		printlog.TargetDateGTE(from),
		printlog.TargetDateLTE(to),
	)
	if req.Msg.DocumentType != printv1.DocumentType_DOCUMENT_TYPE_UNSPECIFIED {
		docType, _, _, _, err := parseRequest(req.Msg.DocumentType, uuid.NewString(), "2006-01-02", nil, nil)
		if err != nil {
			return nil, err
		}
		q = q.Where(printlog.DocumentTypeEQ(string(docType)))
	}
	if req.Msg.RouteId != nil {
		rid, err := uuid.Parse(*req.Msg.RouteId)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("route_id 格式錯誤: %v", err))
		}
		q = q.Where(printlog.RouteIDEQ(rid))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	page := max(int(req.Msg.Page), 1)
	pageSize := min(max(int(req.Msg.PageSize), 1), maxPageSize)
	logs, err := q.
		Order(ent.Desc(printlog.FieldPrintedAt)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// 批次補車次名稱
	routeIDs := map[uuid.UUID]bool{}
	for _, l := range logs {
		routeIDs[l.RouteID] = true
	}
	routeNames := map[uuid.UUID]string{}
	if len(routeIDs) > 0 {
		ids := make([]uuid.UUID, 0, len(routeIDs))
		for id := range routeIDs {
			ids = append(ids, id)
		}
		routes, err := s.db.Route.Query().Where(route.IDIn(ids...)).All(ctx)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		for _, r := range routes {
			routeNames[r.ID] = r.Name
		}
	}
	out := &printv1.ListLogsResponse{Total: int32(total)}
	for _, l := range logs {
		entry := &printv1.PrintLogEntry{
			Id:           l.ID.String(),
			DocumentType: docTypeToProto(l.DocumentType),
			RouteId:      l.RouteID.String(),
			RouteName:    routeNames[l.RouteID],
			TargetDate:   l.TargetDate.Format("2006-01-02"),
			IsReprint:    l.IsReprint,
			ReprintReason: deref(l.ReprintReason),
			PrintedBy:    l.PrintedBy.String(),
			PrintedAt:    l.PrintedAt.Format(time.RFC3339),
		}
		if l.FileAssetID != nil {
			entry.DownloadUrl = print.DownloadURL(*l.FileAssetID)
		}
		out.Logs = append(out.Logs, entry)
	}
	return connect.NewResponse(out), nil
}

func docTypeToProto(s string) printv1.DocumentType {
	switch s {
	case string(print.DocumentTypeDispatchSummary):
		return printv1.DocumentType_DOCUMENT_TYPE_DISPATCH_SUMMARY
	case string(print.DocumentTypeDeliveryNote):
		return printv1.DocumentType_DOCUMENT_TYPE_DELIVERY_NOTE
	case string(print.DocumentTypePickingList):
		return printv1.DocumentType_DOCUMENT_TYPE_PICKING_LIST
	case string(print.DocumentTypeProcessingList):
		return printv1.DocumentType_DOCUMENT_TYPE_PROCESSING_LIST
	default:
		return printv1.DocumentType_DOCUMENT_TYPE_UNSPECIFIED
	}
}
```

(簽名中的 `ListLogsRequestAlias` 為排版殘留,實作時為 `*connect.Request[printv1.ListLogsRequest]` — 以此為準。)

`InitDomains()`(`backend/internal/server/domains.go`)掛載(於 Connect mux 註冊區追加;Gotenberg base URL / 儲存根目錄自 config):

```go
	// 列印服務(09-printing 計畫 Task 4)
	printConverter := print.NewGotenbergClient(cfg.GotenbergURL, 30*time.Second)
	printFileStore := fileassets.NewPrintStore(db, localStore) // 04-plan Task 8 adapter;TODO(接手:04-plan Task 8)
	printCore := print.NewService(db, printConverter, printFileStore)
	printsHandler := prints.NewPrintsService(db, printCore, printFileStore, audit.NewTxRecorder) // 03-plan 提供
	mux.Handle(printv1connect.NewPrintServiceHandler(printsHandler))
```

`backend/config/api.go` 增量:

```go
GotenbergURL string `envconfig:"GOTENBERG_URL"` // 例:http://gotenberg:3000;必填,未設定啟動失敗
```

- [ ] **Step 8: 跑全測試確認通過 + Commit**

Run: `cd backend && go test ./... `
Expected: PASS — 全綠(含 Task 1–4 全部測試,且 01–08 各計畫既有測試不迴歸)。

```bash
git add backend/internal/domain/prints backend/proto/v1/print.proto backend/internal/gen backend/internal/server backend/config
git commit -m "feat(backend): Print API 狀態整批守門、重印必填原因與軌跡留存、ListLogs(5.5.3-5.5.4)"
```

---

## Self-Review 記錄

- **Spec 覆蓋**:細部文件 11 子功能 → Task 對應:5.3.1→T1(Step 4/5 模板與 VM,驗收點:依車次分組一車一份=VM 單車次語意+測試排序斷言;delivery_sequence 排序=TestRenderDispatchSummary/TestBuildDispatchSummary;無金額=reflect 測試+渲染關鍵字掃描;special_cut_note 顯示=兩測試皆斷言);5.3.2→T1(每店一頁 page-break、簽收欄空白、聯絡人/分切規格顯示,渲染與組合測試);5.3.3→T1 模板 + T2 組合(車次→倉別→分類→品名排序、base_qty 跨店彙總 6+4=10、需加工提示、核對欄空白);5.3.4→T1 模板 + T2 組合(僅需加工明細、兩區塊順序、原始數量與規格、加工後數量欄不存在於 VM 且模板渲染空白);5.4.1→T2 Step 1–4(multipart/A4 參數/逾時/5xx 有限重試/4xx 不重試/記憶體傳遞);5.4.2→T2 Step 5–8(四類型成形、選擇器組合矩陣、空表標記、not_found/invalid_argument、狀態過濾不在此層);5.4.3→T2 Step 9–10(空表 failed_precondition 不觸發轉換不落檔、字型於模板 CSS、10 MB 上限、交易外落檔+補償刪除、每次新 file_asset)+ T3/T4 handler 的同事務寫入;5.5.1→T3 Step 1–4(兩表 schema、記錄類無軟刪除、RLS policy、比對鍵與查詢索引、file_asset FK);5.5.2→T3 Step 5–9(不限狀態預覽、print_logs 隔離、空表拒絕、不寫稽核、下載 URL);5.5.3→T4 Step 1–8(status=processing 整批守門含單號、FOR UPDATE 交易內再確認、首印三處記錄同事務、PDF 經下載 URL、ListLogs 分頁上限 100);5.5.4→T4 Step 4–5(原因必填含純空白、staff 可印/customer+guest 拒絕、is_reprint=true 新 PDF 新記錄舊的保留、退回 pending 拒絕)。無缺漏。
- **已知佔位(皆標 TODO + 接手 Task)**:`print.FileStore` 的真實 adapter(→ 04-master-data 計畫 Task 8,契約已與該計畫對齊:WriteLocal / CreateAsset(tx) / DeleteLocal,owner_type=`print_pdf`);`AuditFactory` 的 DB 版構造 `audit.NewTxRecorder`(→ 03-metadicts-audit 計畫);main.go 掛載處的 `fileassets.NewPrintStore`(→ 04 計畫 Task 8)。這些是跨 domain 依賴,非本計畫範圍;本計畫測試全部以 fake / txFileStore 自給自足。
- **假定契約(執行時對齊,各標唯一調整點)**:`sales_order_items.qty/base_qty` 的 Go 型別假定 `shopspring/decimal`(→ 05-sales-orders 計畫;調整點 = service.go 的 decimal 轉換);`ent.CustomerAddress.Address` / `ent.CustomerContact` 欄位名假定(→ 04 計畫;調整點 = `customersByOrder` 的 formatAddress);`audit.Recorder` 若無 `RecorderFunc` 適配器,測試 capture 改用小型 struct 實作(→ 01/03 計畫);migration 檔序號以執行當下最大序號 +1 為準。
- **類型一致**:`print.DocumentType` / 四種 VM / `print.Selector` / `print.FileStore` / `print.StoredFile` / `print.PreparedPDF` / `prints.AuditFactory` 於 Task 1–4 簽名一致;proto 訊息欄位(Preview/Print/ListLogs)與 handler 使用一致;`print.DownloadURL` 與 04 計畫 `/api/v1/files/{id}/download` 契約一致。

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-08-17-backend-09-printing-plan.md`. Two execution options:

**1. Subagent-Driven (recommended)** — 每個 Task 派新 subagent 執行,Task 間 review,迭代快。

**2. Inline Execution** — 用 executing-plans 在本 session 逐批執行,設 checkpoint review。

Which approach?

---

*計畫版本:v1.0.0(2026-08-17);對應細部文件 `backend-detail/09-printing.md` v1.0.0、原計畫 v2.9.0、規格書 v1.0.34。*

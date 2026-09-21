package print_test

import (
	"strings"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/print"
)

// TestRenderAllTypes 四種模板皆可渲染;店家順序即 delivery_sequence 傳入順序;
// 兩區塊順序固定(加工室揀在前)。
func TestRenderAllTypes(t *testing.T) {
	stores := []print.StoreBlock{
		{CustomerCode: "TY0001", Name: "甲店", Address: "中山路1號", Sequence: 2,
			Items: []print.StoreItem{{Name: "蘋果", Qty: "10", Unit: "斤", SpecialCut: "對半"}}},
		{CustomerCode: "TY0002", Name: "乙店", Address: "中正路2號", Sequence: 1,
			Items: []print.StoreItem{{Name: "梨", Qty: "5", Unit: "顆"}}},
	}
	sum, err := print.Render(print.DispatchSummary, print.DispatchSummaryModel{
		CompanyName: "測試公司", RouteCode: "A", RouteName: "市區線",
		TargetDate: "2026-09-22", PrintedAt: "2026-09-22T08:00:00Z", Stores: stores,
	})
	if err != nil {
		t.Fatalf("dispatch_summary: %v", err)
	}
	if !strings.Contains(sum, "單車總表") || !strings.Contains(sum, "甲店") {
		t.Fatalf("單車總表內容缺失: %q", sum[:200])
	}
	// 傳入順序即渲染順序(排序由組合層負責,見 5.4.2)。
	if strings.Index(sum, "甲店") > strings.Index(sum, "乙店") {
		t.Fatal("店家應依傳入順序渲染")
	}

	note, err := print.Render(print.DeliveryNote, print.DeliveryNoteModel{
		CompanyName: "測試公司", RouteCode: "A", RouteName: "市區線",
		TargetDate: "2026-09-22", Stores: stores,
	})
	if err != nil {
		t.Fatalf("delivery_note: %v", err)
	}
	if !strings.Contains(note, "客戶簽名") {
		t.Fatal("對點單應有簽收欄")
	}

	pick, err := print.Render(print.PickingList, print.PickingListModel{
		CompanyName: "測試公司", RouteCode: "A", RouteName: "市區線",
		TargetDate: "2026-09-22", WarehouseName: "本倉",
		Rows: []print.PickingRow{{Category: "水果", Name: "蘋果", BaseQty: "6", Unit: "斤", NeedsCut: true}},
	})
	if err != nil {
		t.Fatalf("picking_list: %v", err)
	}
	if !strings.Contains(pick, "需加工") {
		t.Fatal("揀貨單需加工提示缺失")
	}

	proc, err := print.Render(print.ProcessingList, print.ProcessingListModel{
		CompanyName: "測試公司", RouteCode: "A", RouteName: "市區線", TargetDate: "2026-09-22",
		Workshop: []print.ProcessRow{{Warehouse: "本倉", Name: "蘋果", Qty: "6", Unit: "斤", Spec: "切片"}},
		Delivery: []print.ProcessRow{{Store: "甲店", Name: "蘋果", Qty: "10", Unit: "斤", Spec: "對半"}},
	})
	if err != nil {
		t.Fatalf("processing_list: %v", err)
	}
	if strings.Index(proc, "加工室揀") > strings.Index(proc, "配送揀") {
		t.Fatal("加工室揀區塊必須在前")
	}
}

// TestRenderHasNoMoneyWords 四單據渲染結果全文無金額字樣(D12/D15 雙重防呆:
// view model 無金額欄 + 模板無金額字)。
func TestRenderHasNoMoneyWords(t *testing.T) {
	models := map[print.DocumentType]any{
		print.DispatchSummary: print.DispatchSummaryModel{CompanyName: "C", Stores: []print.StoreBlock{
			{Name: "S", Items: []print.StoreItem{{Name: "N", Qty: "1", Unit: "斤"}}}}},
		print.DeliveryNote: print.DeliveryNoteModel{CompanyName: "C",
			Stores: []print.StoreBlock{{Name: "S", Items: []print.StoreItem{{Name: "N"}}}}},
		print.PickingList: print.PickingListModel{CompanyName: "C",
			Rows: []print.PickingRow{{Name: "N", BaseQty: "1"}}},
		print.ProcessingList: print.ProcessingListModel{CompanyName: "C",
			Workshop: []print.ProcessRow{{Name: "N"}}, Delivery: []print.ProcessRow{{Name: "N"}}},
	}
	for typ, m := range models {
		out, err := print.Render(typ, m)
		if err != nil {
			t.Fatalf("%s: %v", typ, err)
		}
		for _, w := range []string{"金額", "單價", "小計", "總額", "價格", "price", "amount", "total"} {
			if strings.Contains(out, w) {
				t.Fatalf("%s 含金額字樣 %q", typ, w)
			}
		}
	}
}

// TestRenderUnknownType 未知類型回錯誤(呼叫端轉 invalid_argument)。
func TestRenderUnknownType(t *testing.T) {
	if _, err := print.Render("unknown", nil); err == nil {
		t.Fatal("未知類型應回錯誤")
	}
}

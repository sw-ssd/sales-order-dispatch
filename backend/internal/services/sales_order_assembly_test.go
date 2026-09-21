package services

import (
	"context"
	"testing"
	"time"
)

// TestAdjustDeliveryDate 偏好送貨日順延(4.2.5,D26):週二四偏好選週三→週四;選週四維持;
// 未勾選維持;僅週一選週五→下週一;長度異常視同未勾選。
func TestAdjustDeliveryDate(t *testing.T) {
	tueThu := []bool{false, true, false, true, false, false} // 一~六:二、四
	// 2026-09-23 為週三 → 順延至週四 09-24。
	wed := time.Date(2026, 9, 23, 10, 0, 0, 0, time.UTC)
	if got := adjustDeliveryDate(tueThu, wed); got.Format("2006-01-02") != "2026-09-24" {
		t.Fatalf("週三應順延週四,got %s", got.Format("2006-01-02"))
	}
	thu := time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)
	if got := adjustDeliveryDate(tueThu, thu); got.Format("2006-01-02") != "2026-09-24" {
		t.Fatalf("勾選日應維持,got %s", got.Format("2006-01-02"))
	}
	none := []bool{false, false, false, false, false, false}
	if got := adjustDeliveryDate(none, wed); !got.Equal(wed) {
		t.Fatalf("未勾選應維持,got %s", got.Format("2006-01-02"))
	}
	if got := adjustDeliveryDate([]bool{true}, wed); !got.Equal(wed) {
		t.Fatalf("長度異常視同未勾選,got %s", got.Format("2006-01-02"))
	}
	// 僅週一,選週五 09-25 → 下週一 09-28(跨週)。
	monOnly := []bool{true, false, false, false, false, false}
	fri := time.Date(2026, 9, 25, 10, 0, 0, 0, time.UTC)
	if got := adjustDeliveryDate(monOnly, fri); got.Format("2006-01-02") != "2026-09-28" {
		t.Fatalf("應跨週順延下週一,got %s", got.Format("2006-01-02"))
	}
}

// TestResolveBaseQty 換算(4.2.1):基本單位直填;非基本單位 qty × rate(十進位);
// 無效單位／非正數量拒絕;手打品名(productID=0)無換算來源照填。
func TestResolveBaseQty(t *testing.T) {
	db := newCounterTestDB(t)
	ctx := context.Background()
	p := db.Product.Create().SetCompanyID(7).SetCode("P1").SetName("蘋果").SaveX(ctx)
	db.ProductUnit.Create().SetProductID(p.ID).SetUnitCode("斤").SetConversionRate("1").SetIsBase(true).SaveX(ctx)
	db.ProductUnit.Create().SetProductID(p.ID).SetUnitCode("條").SetConversionRate("0.6").SaveX(ctx)

	if got, err := resolveBaseQty(ctx, db, p.ID, "斤", "10"); err != nil || got != "10" {
		t.Fatalf("基本單位應直填,got %q err=%v", got, err)
	}
	// 10 條 × 0.6 = 6(基本單位)。
	if got, err := resolveBaseQty(ctx, db, p.ID, "條", "10"); err != nil || got != "6" {
		t.Fatalf("10 條應換算 6,got %q err=%v", got, err)
	}
	if _, err := resolveBaseQty(ctx, db, p.ID, "箱", "10"); err == nil {
		t.Fatal("無效單位應拒絕")
	}
	if _, err := resolveBaseQty(ctx, db, p.ID, "斤", "0"); err == nil {
		t.Fatal("非正數量應拒絕")
	}
	if got, err := resolveBaseQty(ctx, db, 0, "斤", "7"); err != nil || got != "7" {
		t.Fatalf("手打品名應照填,got %q err=%v", got, err)
	}
}

// TestOrderCustomerGuard 客戶守衛(4.2.3):客戶帳號強制為自己(忽略請求值);
// 員工照請求值走。
func TestOrderCustomerGuard(t *testing.T) {
	got, err := orderCustomerGuard("999", "42", true)
	if err != nil || got != 42 {
		t.Fatalf("客戶帳號應強制為自己,got %d err=%v", got, err)
	}
	got, err = orderCustomerGuard("77", "", false)
	if err != nil || got != 77 {
		t.Fatalf("員工應照請求值,got %d err=%v", got, err)
	}
	if _, err := orderCustomerGuard("abc", "", false); err == nil {
		t.Fatal("非法 customer_id 應報錯")
	}
}

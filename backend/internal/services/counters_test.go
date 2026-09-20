package services

import (
	"context"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
)

// newCounterTestDB 建立 sqlite 記憶體 client（計數器只讀業務表，免 Docker）。
func newCounterTestDB(t *testing.T) *ent.Client {
	t.Helper()
	db := enttest.Open(t, "sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestEntitlementCounterSeatsExcludesInactive 席位規則（spec §3.2）＝未軟刪除且非 inactive 的
// 帳號數：停用可釋放席位，客戶才能自助降級。users 表無 deleted_at 欄位（00005 DDL），
// 故「未軟刪除」在席位一項是由 schema 保證，唯一要擋的是 inactive。
// pending 仍佔席位（規則是不等於 inactive，不是等於 active）＋ 另一公司的帳號不得計入。
func TestEntitlementCounterSeatsExcludesInactive(t *testing.T) {
	db := newCounterTestDB(t)
	ctx := context.Background()
	co := db.Company.Create().SetName("公司A").SetIdentifier("CNT-A").SaveX(ctx)
	other := db.Company.Create().SetName("公司B").SetIdentifier("CNT-B").SaveX(ctx)

	addUser := func(email string, companyID int, status user.Status) {
		t.Helper()
		db.User.Create().SetEmail(email).SetName(email).SetRole("staff").
			SetPasswordHash("x").SetStatus(status).SetCompanyID(companyID).SaveX(ctx)
	}
	addUser("active@example.com", co.ID, user.StatusActive)
	addUser("pending@example.com", co.ID, user.StatusPending)
	addUser("inactive@example.com", co.ID, user.StatusInactive)
	addUser("other-company@example.com", other.ID, user.StatusActive)

	got, err := NewEntitlementCounter(db).Count(ctx, co.ID, entitlements.LimitSeats)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if got != 2 {
		t.Fatalf("席位應為 2（active ＋ pending，排除 inactive 與他公司），got %d", got)
	}
}

// TestEntitlementCounterRowFeaturesExcludeSoftDeleted 其餘 feature 照「未軟刪除的列數」：
// 軟刪除不算用量（否則客戶刪了資料仍佔額度），且一律限定該公司。
func TestEntitlementCounterRowFeaturesExcludeSoftDeleted(t *testing.T) {
	db := newCounterTestDB(t)
	ctx := context.Background()
	co := db.Company.Create().SetName("公司A").SetIdentifier("CNT-A").SaveX(ctx)
	other := db.Company.Create().SetName("公司B").SetIdentifier("CNT-B").SaveX(ctx)
	gone := time.Now().UTC().Add(-time.Hour)

	dept := db.Department.Create().SetCompanyID(co.ID).SetName("部門甲").SaveX(ctx)
	db.Department.Create().SetCompanyID(co.ID).SetName("部門乙").SetDeletedAt(gone).SaveX(ctx)
	db.Department.Create().SetCompanyID(other.ID).SetName("部門丙").SaveX(ctx)

	db.Customer.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetCustomerCode("CNT000001").SetName("王小明").SaveX(ctx)
	db.Customer.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetCustomerCode("CNT000002").SetName("已刪").SetDeletedAt(gone).SaveX(ctx)
	db.Customer.Create().SetCompanyID(other.ID).
		SetCustomerCode("CNT000003").SetName("他公司").SaveX(ctx)

	db.Product.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetCode("P-1").SetName("商品一").SaveX(ctx)
	db.Product.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetCode("P-2").SetName("已刪").SetDeletedAt(gone).SaveX(ctx)
	db.Product.Create().SetCompanyID(other.ID).
		SetCode("P-3").SetName("他公司").SaveX(ctx)

	counter := NewEntitlementCounter(db)
	for _, tc := range []struct {
		feature string
		want    int
	}{
		{entitlements.LimitCustomers, 1},
		{entitlements.LimitProducts, 1},
		{entitlements.LimitDepartments, 1},
	} {
		got, err := counter.Count(ctx, co.ID, tc.feature)
		if err != nil {
			t.Fatalf("Count(%s): %v", tc.feature, err)
		}
		if got != tc.want {
			t.Errorf("%s 應為 %d（排除軟刪除與他公司），got %d", tc.feature, tc.want, got)
		}
	}
}

// TestEntitlementCounterUnknownFeatureErrors 未定義的 feature 不得靜默回 0:
// 回 0 等於「用量為零」，會讓判定層誤判為未超額而放行。
func TestEntitlementCounterUnknownFeatureErrors(t *testing.T) {
	db := newCounterTestDB(t)
	co := db.Company.Create().SetName("公司A").SetIdentifier("CNT-A").SaveX(context.Background())

	got, err := NewEntitlementCounter(db).Count(context.Background(), co.ID, "limit.nonexistent")
	if err == nil {
		t.Fatalf("未定義 feature 必須回錯誤，got (%d, nil)", got)
	}
	if !strings.Contains(err.Error(), "limit.nonexistent") {
		t.Fatalf("錯誤訊息應點名該 feature，got %q", err.Error())
	}
}

package openfga_test

import (
	"context"
	"testing"
)

// TestPrimaryAccountExcludedFromAbility D22 規格 4.2:客戶主帳號僅供帳號管理,
// **所有業務 API 一律拒絕**;子帳號才是業務登入身分。
//
// 表達方式:ability 的 can_read/can_write 以 `but not primary_account` 排除主帳號
// (exclusion)。這條守住三件事:
//   - 子帳號經 role#assigned 取得能力(不受影響);
//   - 主帳號即使同角色,加了 primary_account tuple 後**一律 deny**;
//   - 移除該 tuple 即恢復(撤銷是資料操作,不必改模型、不必重啟)。
func TestPrimaryAccountExcludedFromAbility(t *testing.T) {
	ctx := context.Background()
	e := newEngine(t)

	const (
		custRole = "role:5#assigned" // 客戶角色的能力集合(role#assigned userset)
		res      = "ability:sales_order"
	)
	// 角色層:customer 對 sales_order 有 can_read/can_write。
	for _, rel := range []string{"can_read", "can_write"} {
		if err := e.WriteTuple(ctx, custRole, rel, res); err != nil {
			t.Fatalf("寫入 %s: %v", rel, err)
		}
	}

	sub := "user:10" // 子帳號
	pri := "user:11" // 主帳號
	for _, u := range []string{sub, pri} {
		if err := e.WriteTuple(ctx, u, "assigned", "role:5"); err != nil {
			t.Fatalf("指派角色: %v", err)
		}
	}

	check := func(u, rel string) bool {
		t.Helper()
		ok, err := e.Check(ctx, u, rel, res)
		if err != nil {
			t.Fatalf("Check(%s,%s): %v", u, rel, err)
		}
		return ok
	}

	// ① 加 primary_account 之前:兩者皆允許(基線)。
	if !check(sub, "can_write") || !check(pri, "can_write") {
		t.Fatal("基線:子/主帳號都應經角色取得 can_write")
	}

	// ② 標記主帳號 → 主帳號兩個動作都被排除,子帳號不受影響。
	if err := e.WriteTuple(ctx, pri, "primary_account", res); err != nil {
		t.Fatalf("寫入 primary_account: %v", err)
	}
	if check(pri, "can_write") {
		t.Fatal("主帳號不得有 can_write(規格 4.2 業務 API 一律 403)")
	}
	if check(pri, "can_read") {
		t.Fatal("主帳號不得有 can_read")
	}
	if !check(sub, "can_write") || !check(sub, "can_read") {
		t.Fatal("子帳號不得受主帳號標記影響")
	}

	// ③ 撤銷標記(例如後台把該帳號改為子帳號)→ 恢復。
	if err := e.DeleteTuple(ctx, pri, "primary_account", res); err != nil {
		t.Fatalf("刪除 primary_account: %v", err)
	}
	if !check(pri, "can_write") {
		t.Fatal("移除 primary_account 後應恢復能力(撤銷是資料操作)")
	}
}

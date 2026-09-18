// Package openfga_test 驗證 internal/authz/openfga 授權引擎 facade(D32)。
// 以 OpenFGA 記憶體 datastore 起內嵌實例,測試 Check/ListObjects/WriteTuple/DeleteTuple。
// 授權 model 採「租戶 userset + 角色指派分離 + 資料驅動」:角色→權限由 role_permissions
// → tuples 承載(role#assigned userset),物件狀態條件由 domain 狀態機處理,不進 CEL。
package openfga_test

import (
	"context"
	"strings"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
	ofga "github.com/salesorder/sales-order-1.0/backend/third_party/openfga"
)

// newEngine 起 OpenFGA 記憶體實例並回傳 facade;defer 關閉由呼叫端負責。
func newEngine(t *testing.T) *openfga.Engine {
	t.Helper()
	client, err := ofga.NewMemory(context.Background(), "facade-test-store")
	if err != nil {
		t.Fatalf("NewMemory: %v", err)
	}
	t.Cleanup(client.Close)
	return openfga.New(client)
}

func TestCheckAllowDeny(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	e := newEngine(t)

	// 角色指派:staff 角色指派給使用者 u1。
	if err := e.WriteTuple(ctx, "user:u1", "assigned", "role:staff"); err != nil {
		t.Fatalf("WriteTuple(role assign): %v", err)
	}
	// 能力授予:role:staff#assigned 可讀取 ability:company(資料驅動)。
	if err := e.WriteTuple(ctx, "role:staff#assigned", "can_read", "ability:company"); err != nil {
		t.Fatalf("WriteTuple(ability read): %v", err)
	}

	t.Run("允許", func(t *testing.T) {
		ok, err := e.Check(ctx, "user:u1", "can_read", "ability:company")
		if err != nil {
			t.Fatalf("Check: %v", err)
		}
		if !ok {
			t.Fatal("Check 應允許 user:u1 can_read ability:company")
		}
	})
	t.Run("拒絕-未授權動作", func(t *testing.T) {
		ok, err := e.Check(ctx, "user:u1", "can_write", "ability:company")
		if err != nil {
			t.Fatalf("Check: %v", err)
		}
		if ok {
			t.Fatal("Check 應拒絕 user:u1 can_write ability:company(未授予)")
		}
	})
	t.Run("拒絕-無指派使用者", func(t *testing.T) {
		ok, err := e.Check(ctx, "user:nobody", "can_read", "ability:company")
		if err != nil {
			t.Fatalf("Check: %v", err)
		}
		if ok {
			t.Fatal("Check 應拒絕未指派角色之使用者")
		}
	})
}

func TestListObjects(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	e := newEngine(t)

	if err := e.WriteTuple(ctx, "user:u1", "assigned", "role:staff"); err != nil {
		t.Fatalf("WriteTuple(role assign): %v", err)
	}
	for _, res := range []string{"company", "department", "role"} {
		if err := e.WriteTuple(ctx, "role:staff#assigned", "can_read", "ability:"+res); err != nil {
			t.Fatalf("WriteTuple(ability read %s): %v", res, err)
		}
	}

	objs, err := e.ListObjects(ctx, "user:u1", "can_read", "ability")
	if err != nil {
		t.Fatalf("ListObjects: %v", err)
	}
	want := map[string]bool{"company": true, "department": true, "role": true}
	// OpenFGA ListObjects 回傳完整 object id(含型別前綴,如 "ability:company")。
	for _, o := range objs {
		if id, ok := strings.CutPrefix(o, "ability:"); ok {
			delete(want, id)
		}
	}
	if len(want) != 0 {
		t.Fatalf("ListObjects 缺 %v,實際 %v", want, objs)
	}
}

func TestDeleteTuple(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	e := newEngine(t)

	if err := e.WriteTuple(ctx, "user:u1", "assigned", "role:staff"); err != nil {
		t.Fatalf("WriteTuple(role assign): %v", err)
	}
	if err := e.WriteTuple(ctx, "role:staff#assigned", "can_read", "ability:company"); err != nil {
		t.Fatalf("WriteTuple(ability read): %v", err)
	}
	if err := e.DeleteTuple(ctx, "role:staff#assigned", "can_read", "ability:company"); err != nil {
		t.Fatalf("DeleteTuple: %v", err)
	}
	ok, err := e.Check(ctx, "user:u1", "can_read", "ability:company")
	if err != nil {
		t.Fatalf("Check after delete: %v", err)
	}
	if ok {
		t.Fatal("刪除 tuple 後 Check 應拒絕")
	}
}

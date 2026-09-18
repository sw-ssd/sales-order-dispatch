package auth_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	authzopenfga "github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
	auth "github.com/salesorder/sales-order-1.0/backend/internal/domain/auth"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	ofga "github.com/salesorder/sales-order-1.0/backend/third_party/openfga"
)

// newTestHandler 建立 enttest sqlite client 與 GetAbility handler。
func newTestHandler(t *testing.T, developerEnabled bool) (*auth.AbilityHandler, *ent.Client, context.Context) {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	return auth.NewAbilityHandler(client, auth.Config{DeveloperAccountEnabled: developerEnabled}), client, context.Background()
}

// newEngine 起 OpenFGA 記憶體引擎(測試用)。
func newEngine(t *testing.T) *authzopenfga.Engine {
	t.Helper()
	ofgaClient, err := ofga.NewMemory(context.Background(), "ability-test-store")
	if err != nil {
		t.Fatalf("NewMemory: %v", err)
	}
	t.Cleanup(ofgaClient.Close)
	return authzopenfga.New(ofgaClient)
}

func TestGetAbilityDeveloper(t *testing.T) {
	t.Run("開關啟用回 manage-all", func(t *testing.T) {
		h, _, ctx := newTestHandler(t, true)
		resp, err := h.GetAbility(
			authz.WithIdentity(ctx, authz.Identity{Role: "developer", Roles: []string{"developer"}}),
			connect.NewRequest(&v1.GetAbilityRequest{}),
		)
		if err != nil {
			t.Fatalf("GetAbility: %v", err)
		}
		rules := resp.Msg.GetRules()
		if len(rules) != 1 || rules[0].GetAction() != "manage" || rules[0].GetSubject() != "all" {
			t.Fatalf("developer 開關啟用應回 manage:all,got %#v", rules)
		}
	})
	t.Run("開關關閉視同一般身分(無能力)", func(t *testing.T) {
		h, _, ctx := newTestHandler(t, false)
		resp, err := h.GetAbility(
			authz.WithIdentity(ctx, authz.Identity{Role: "developer", Roles: []string{"developer"}}),
			connect.NewRequest(&v1.GetAbilityRequest{}),
		)
		if err != nil {
			t.Fatalf("GetAbility: %v", err)
		}
		if rules := resp.Msg.GetRules(); len(rules) != 0 {
			t.Fatalf("開發逃生門關閉時 rules = %d, want 0", len(rules))
		}
	})
}

func TestGetAbilityOpenFGADriven(t *testing.T) {
	h, _, baseCtx := newTestHandler(t, true)
	ctx := context.Background()
	e := newEngine(t)
	// 授予 user:u1 對 sales_order / customer 的 can_read,及 sales_order 的 can_write。
	if err := e.WriteTuple(ctx, "user:u1", "can_read", "ability:sales_order"); err != nil {
		t.Fatalf("WriteTuple(read sales_order): %v", err)
	}
	if err := e.WriteTuple(ctx, "user:u1", "can_read", "ability:customer"); err != nil {
		t.Fatalf("WriteTuple(read customer): %v", err)
	}
	if err := e.WriteTuple(ctx, "user:u1", "can_write", "ability:sales_order"); err != nil {
		t.Fatalf("WriteTuple(write sales_order): %v", err)
	}

	c := authz.WithEngine(authz.WithIdentity(baseCtx, authz.Identity{UserID: "u1", Roles: []string{"staff"}}), e)
	resp, err := h.GetAbility(c, connect.NewRequest(&v1.GetAbilityRequest{}))
	if err != nil {
		t.Fatalf("GetAbility: %v", err)
	}
	rules := resp.Msg.GetRules()
	got := map[string]string{}
	for _, r := range rules {
		got[r.GetAction()+"\x00"+r.GetSubject()] = ""
	}
	for _, want := range []string{"read\x00sales_order", "read\x00customer", "write\x00sales_order"} {
		if _, ok := got[want]; !ok {
			t.Errorf("缺少能力 %q,got %#v", want, rules)
		}
	}
	// 未授予 → 不出現。
	if _, ok := got["write\x00customer"]; ok {
		t.Error("未授予 customer write 不應出現")
	}
}

func TestGetAbilityNoEngineFailClosed(t *testing.T) {
	h, _, ctx := newTestHandler(t, true)
	// 未注入 engine → 空(除 developer 逃生門外)。
	resp, err := h.GetAbility(
		authz.WithIdentity(ctx, authz.Identity{UserID: "u1", Roles: []string{"staff"}}),
		connect.NewRequest(&v1.GetAbilityRequest{}),
	)
	if err != nil {
		t.Fatalf("GetAbility: %v", err)
	}
	if rules := resp.Msg.GetRules(); len(rules) != 0 {
		t.Fatalf("無 engine 應回空,rules = %#v", rules)
	}
}

func TestGetAbilityUnauthenticated(t *testing.T) {
	h, _, ctx := newTestHandler(t, true)
	// 未登入(無身分)→ 空。
	resp, err := h.GetAbility(ctx, connect.NewRequest(&v1.GetAbilityRequest{}))
	if err != nil {
		t.Fatalf("GetAbility: %v", err)
	}
	if rules := resp.Msg.GetRules(); len(rules) != 0 {
		t.Fatalf("未登入應回空,rules = %#v", rules)
	}
}

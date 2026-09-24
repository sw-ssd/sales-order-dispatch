package openfga

import (
	"context"
	"testing"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/language/pkg/go/transformer"
)

func TestNewMemoryClient(t *testing.T) {
	ctx := context.Background()
	c, err := NewMemory(ctx, "test-store")
	if err != nil {
		t.Fatalf("NewMemory: %v", err)
	}
	defer c.Close()
	if c.StoreID == "" {
		t.Fatal("store id 不應為空")
	}
}

// TestWriteDefaultModel 驗證 model.fga 寫入 store 並可讀回(含核心資源型別)。
func TestWriteDefaultModel(t *testing.T) {
	ctx := context.Background()
	c, err := NewMemory(ctx, "model-store")
	if err != nil {
		t.Fatalf("NewMemory: %v", err)
	}
	defer c.Close()
	if c.ModelID == "" {
		t.Fatal("model id 不應為空")
	}
	got, err := c.srv.ReadAuthorizationModel(ctx, &openfgav1.ReadAuthorizationModelRequest{
		StoreId: c.StoreID, Id: c.ModelID,
	})
	if err != nil {
		t.Fatalf("ReadAuthorizationModel: %v", err)
	}
	types := map[string]bool{}
	for _, td := range got.GetAuthorizationModel().GetTypeDefinitions() {
		types[td.GetType()] = true
	}
	for _, want := range []string{"user", "role", "company", "department", "ability"} {
		if !types[want] {
			t.Errorf("model 缺資源型別 %q", want)
		}
	}
}

// TestEnsureDefaultModelReconcilesDSL 是 model 漂移的回歸測試(2026-09-24 實機踩到):
// store 已有 model 但內容與 modelDSL 不同(改了型別/relation)時,必須寫入新版本,
// 否則後續 tuple 寫入與 Check 會以 `type 'x' not found` 失敗,症狀偽裝成「權限沒開」。
func TestEnsureDefaultModelReconcilesDSL(t *testing.T) {
	ctx := context.Background()
	c, err := NewMemory(ctx, "reconcile-store")
	if err != nil {
		t.Fatalf("NewMemory: %v", err)
	}
	defer c.Close()
	stale := c.ModelID

	// 手動寫一版「缺 logistics_delivery」的 model,模擬舊 model 留在 store 的狀態。
	old, err := transformer.TransformDSLToProto(`model
  schema 1.1
type user
`)
	if err != nil {
		t.Fatalf("transform: %v", err)
	}
	resp, err := c.srv.WriteAuthorizationModel(ctx, &openfgav1.WriteAuthorizationModelRequest{
		StoreId:         c.StoreID,
		SchemaVersion:   old.GetSchemaVersion(),
		TypeDefinitions: old.GetTypeDefinitions(),
	})
	if err != nil {
		t.Fatalf("write stale model: %v", err)
	}
	if resp.GetAuthorizationModelId() == stale {
		t.Fatal("測試前置錯誤:stale model id 未變")
	}

	// 對帳後:最新 model 必須重新等價於 modelDSL(含 logistics_delivery),且 ModelID 更新。
	if err := c.EnsureDefaultModel(ctx); err != nil {
		t.Fatalf("EnsureDefaultModel: %v", err)
	}
	if c.ModelID == resp.GetAuthorizationModelId() {
		t.Fatal("DSL 不同時應寫入新版本,而非沿用舊 model")
	}
	got, err := c.srv.ReadAuthorizationModel(ctx, &openfgav1.ReadAuthorizationModelRequest{
		StoreId: c.StoreID, Id: c.ModelID,
	})
	if err != nil {
		t.Fatalf("ReadAuthorizationModel: %v", err)
	}
	types := map[string]bool{}
	for _, td := range got.GetAuthorizationModel().GetTypeDefinitions() {
		types[td.GetType()] = true
	}
	if !types["logistics_delivery"] || !types["driver"] || !types["vehicle"] {
		t.Fatalf("對帳後應含 logistics 型別,got %v", types)
	}

	// 再跑一次:內容已一致 → 不再寫新版本(id 不變,避免每次重啟漂移)。
	before := c.ModelID
	if err := c.EnsureDefaultModel(ctx); err != nil {
		t.Fatalf("EnsureDefaultModel(二次): %v", err)
	}
	if c.ModelID != before {
		t.Fatalf("內容一致時不應改 model id(前 %s 後 %s)", before, c.ModelID)
	}
}

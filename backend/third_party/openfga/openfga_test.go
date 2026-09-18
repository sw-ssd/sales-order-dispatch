package openfga

import (
	"context"
	"testing"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
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

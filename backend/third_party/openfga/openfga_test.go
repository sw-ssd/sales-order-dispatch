package openfga

import (
	"context"
	"testing"
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

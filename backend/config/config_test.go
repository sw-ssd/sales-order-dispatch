package config

import "testing"

func TestNewDefaults(t *testing.T) {
	c := New()
	if c.API.Addr != ":3080" {
		t.Errorf("API.Addr = %q, want :3080", c.API.Addr)
	}
	if c.API.Env != "development" {
		t.Errorf("API.Env = %q, want development", c.API.Env)
	}
	if c.Cache.ValkeyAddr != "localhost:6379" {
		t.Errorf("Cache.ValkeyAddr = %q, want localhost:6379", c.Cache.ValkeyAddr)
	}
	if c.Storage.StorageRoot == "" {
		t.Error("Storage.StorageRoot 應有預設值")
	}
	if c.Observability.LogLevel != "info" {
		t.Errorf("Observability.LogLevel = %q, want info", c.Observability.LogLevel)
	}
}

func TestNewEnvOverride(t *testing.T) {
	t.Setenv("API_ADDR", ":9999")
	t.Setenv("GOOGLE_CLIENT_ID", "gid-123")
	t.Setenv("DATABASE_URL", "postgres://u:p@h:5432/db")

	c := New()
	if c.API.Addr != ":9999" {
		t.Errorf("API.Addr = %q, want :9999", c.API.Addr)
	}
	if c.Auth.GoogleClientID != "gid-123" {
		t.Errorf("Auth.GoogleClientID = %q, want gid-123", c.Auth.GoogleClientID)
	}
	if c.Database.DatabaseURL != "postgres://u:p@h:5432/db" {
		t.Errorf("Database.DatabaseURL = %q, want override", c.Database.DatabaseURL)
	}
}

// TestFCMProviderSelection：FCM 的**環境控制**是這條設定的重點 —— 預設必須是 fake
// （開發不觸外網），而 fcm 模式必須被辨識為需要真憑證。若預設漂成 fcm，開發環境
// 會在每次觸發通知時嘗試對外發送。
func TestFCMProviderSelection(t *testing.T) {
	c := New()
	if c.FCM.Provider != "fake" {
		t.Fatalf("FCM.Provider 預設應為 fake,got %q", c.FCM.Provider)
	}
	if c.FCM.UsesRealFCM() {
		t.Fatal("預設不該要求真發送")
	}
	if c.FCM.TimeoutSeconds <= 0 {
		t.Fatalf("FCM.TimeoutSeconds 應有正的預設值,got %d", c.FCM.TimeoutSeconds)
	}

	t.Setenv("FCM_PROVIDER", "fcm")
	t.Setenv("FCM_PROJECT_ID", "proj-x")
	t.Setenv("FCM_CREDENTIALS_JSON", "/tmp/sa.json")
	c2 := New()
	if !c2.FCM.UsesRealFCM() {
		t.Fatal("FCM_PROVIDER=fcm 應要求真發送")
	}
	if !c2.FCM.Configured() {
		t.Fatal("project id 與憑證都設了,Configured 應為 true")
	}

	// 只設一半：不得被當成「已備妥」（否則啟動看似成功、第一則通知才爆）。
	t.Setenv("FCM_PROJECT_ID", "")
	c3 := New()
	if c3.FCM.Configured() {
		t.Fatal("缺 project id 不該視為 Configured")
	}
}

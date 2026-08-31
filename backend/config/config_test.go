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

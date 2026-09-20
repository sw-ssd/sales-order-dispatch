package config

import "testing"

func TestAdminDSNFallsBackToDatabaseURL(t *testing.T) {
	d := Database{DatabaseURL: "postgres://app@localhost:5432/salesorder"}
	if got := d.AdminDSN(); got != "postgres://app@localhost:5432/salesorder" {
		t.Fatalf("未設 DATABASE_ADMIN_URL 時應沿用 DATABASE_URL,got %q", got)
	}
	d.AdminURL = "postgres://owner@localhost:5432/salesorder"
	if got := d.AdminDSN(); got != "postgres://owner@localhost:5432/salesorder" {
		t.Fatalf("設了 DATABASE_ADMIN_URL 時應採用它,got %q", got)
	}
}

func TestAdminURLFromEnv(t *testing.T) {
	t.Setenv("DATABASE_ADMIN_URL", "postgres://owner@db:5432/salesorder")
	var d Database
	mustProcess(&d)
	if d.AdminURL != "postgres://owner@db:5432/salesorder" {
		t.Fatalf("DATABASE_ADMIN_URL 未綁定,got %q", d.AdminURL)
	}
}

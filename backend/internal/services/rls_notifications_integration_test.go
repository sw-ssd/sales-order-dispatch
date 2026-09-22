//go:build integration

package services

import (
	"strconv"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationRLSNotificationsIsolation 以 app_rw 直連驗證通知域隔離
// （00042 ENABLE + FORCE）：未設 scope → 0 列；設 A 公司 → 只見 A；跨租戶寫入擋。
func TestIntegrationRLSNotificationsIsolation(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	admin := openRawDB(t, dsn)

	coA := insertRLSCompanyPrint(t, admin, "A", "RLSNT-A")
	coB := insertRLSCompanyPrint(t, admin, "B", "RLSNT-B")
	var userA, userB int
	if err := admin.QueryRow(
		`INSERT INTO users (email, name, role, password_hash, company_users, status) VALUES ('a@t.com', '甲', 'staff', 'x', $1, 'active') RETURNING id`,
		coA).Scan(&userA); err != nil {
		t.Fatalf("建使用者 A: %v", err)
	}
	if err := admin.QueryRow(
		`INSERT INTO users (email, name, role, password_hash, company_users, status) VALUES ('b@t.com', '乙', 'staff', 'x', $1, 'active') RETURNING id`,
		coB).Scan(&userB); err != nil {
		t.Fatalf("建使用者 B: %v", err)
	}
	for _, tc := range []struct {
		tbl string
		co  int
	}{
		{"notification_templates", coA}, {"notification_templates", coB},
		{"notifications", coA}, {"notifications", coB},
		{"user_devices", coA}, {"user_devices", coB},
		{"promo_tags", coA}, {"promo_tags", coB},
	} {
		switch tc.tbl {
		case "notification_templates":
			if _, err := admin.Exec(
				`INSERT INTO notification_templates (company_id, code, name, channel, body) VALUES ($1, 'order_created', '下單', 'in_app', 'b')`,
				tc.co); err != nil {
				t.Fatalf("建範本(公司 %d): %v", tc.co, err)
			}
		case "notifications":
			uid := userA
			if tc.co == coB {
				uid = userB
			}
			if _, err := admin.Exec(
				`INSERT INTO notifications (company_id, department_id, user_id, channel, title, content) VALUES ($1, 1, $2, 'in_app', 't', 'c')`,
				tc.co, uid); err != nil {
				t.Fatalf("建通知(公司 %d): %v", tc.co, err)
			}
		case "user_devices":
			uid := userA
			if tc.co == coB {
				uid = userB
			}
			if _, err := admin.Exec(
				`INSERT INTO user_devices (user_id, company_id, platform, fcm_token) VALUES ($1, $2, 'android', $3)`,
				uid, tc.co, "tok-"+strconv.Itoa(tc.co)); err != nil {
				t.Fatalf("建裝置(公司 %d): %v", tc.co, err)
			}
		case "promo_tags":
			if _, err := admin.Exec(
				`INSERT INTO promo_tags (company_id, department_id, code, name) VALUES ($1, 1, 'sale', '促銷')`,
				tc.co); err != nil {
				t.Fatalf("建標籤(公司 %d): %v", tc.co, err)
			}
		}
	}

	app := openAppRoleDB(t, dsn)
	tables := []string{"notification_templates", "notifications", "user_devices", "promo_tags"}

	t.Run("未設 scope → fail-closed", func(t *testing.T) {
		for _, tbl := range tables {
			var n int
			if err := app.QueryRow(`SELECT count(*) FROM ` + tbl).Scan(&n); err != nil {
				t.Fatalf("%s 查詢: %v", tbl, err)
			}
			if n != 0 {
				t.Fatalf("%s 未設 scope 不得可見,got %d", tbl, n)
			}
		}
	})

	t.Run("scope=company A → 只見 A", func(t *testing.T) {
		tx, err := app.Begin()
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		setAppScopePrint(t, tx, coA)
		for _, tbl := range tables {
			var got int
			if err := tx.QueryRow(`SELECT DISTINCT company_id FROM ` + tbl).Scan(&got); err != nil {
				t.Fatalf("%s 查詢: %v", tbl, err)
			}
			if got != coA {
				t.Fatalf("%s 應只看到公司 %d,got %d", tbl, coA, got)
			}
		}
	})

	t.Run("跨租戶 INSERT → 被 WITH CHECK 擋", func(t *testing.T) {
		tx, err := app.Begin()
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		setAppScopePrint(t, tx, coA)
		_, err = tx.Exec(
			`INSERT INTO notifications (company_id, department_id, user_id, channel, title, content) VALUES ($1, 1, $2, 'in_app', 't', 'c')`,
			coB, userB)
		if err == nil {
			t.Fatal("跨租戶寫入應被 WITH CHECK 擋下")
		}
	})
}

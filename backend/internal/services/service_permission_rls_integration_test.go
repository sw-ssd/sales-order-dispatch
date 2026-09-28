//go:build integration

// 服務層逐動作授權在**真 RLS**下的探針(D33 收尾,2026-09-29)。
//
// 這一支補的是單元測試(sqlite,enttest)**結構上驗不到**的一點:role_permissions 自
// 00025 起受 RLS 約束,policy 只要求 `app.current_data_scope` 非空。若逐動作檢查走的是
// 「交易外的 client」,SET LOCAL 不生效 → 查詢回 0 列 → 判定為「無權」——單元測試看不到
// (sqlite 沒有 RLS,照樣讀得到),正式環境則讓**自訂角色靜默失去全部權限**。
//
// 以 app_rw(非 superuser;superuser 繞過 RLS,測不出接線)＋真請求交易驗證:
//
//	① 交易內讀得到 role_permissions;② 逐動作語意成立(有 update 無 create);
//	③ inverted 規則不構成授權。
package services

import (
	"context"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/ent/rolepermission"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

func TestIntegrationServicePermissionUnderAppRole(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	client := openAppRoleEntClient(t, adminDSN)
	admin := openRawDB(t, adminDSN)

	const code = "auditor_it"
	roleID := insertRLSCoreRole(t, admin, code)
	insertRLSCoreRolePermission(t, admin, roleID, "company", "read")
	insertRLSCoreRolePermission(t, admin, roleID, "company", "update")
	// inverted(拒絕)規則:不得構成允許。
	invertedID := insertRLSCoreRolePermission(t, admin, roleID, "role", "read")
	if _, err := admin.Exec(`UPDATE role_permissions SET inverted = true WHERE id = $1`, invertedID); err != nil {
		t.Fatalf("設 inverted: %v", err)
	}

	id := authz.Identity{UserID: "4242", CompanyID: "1", Role: code, Roles: []string{code}}
	// DataScope=all:身分層 ACL 不是本探針標的;SET LOCAL 仍會套用,RLS 條件成立。
	scope := auth.RLSScope{UserID: "4242", CompanyID: "1", DataScope: auth.DataScopeAll, CompanyActive: true}

	// grant 在**真請求交易**內呼叫逐動作檢查(與生產 dbtenant.Interceptor 的路徑同構:
	// middleware 先 WithIdentity＋WithDB,interceptor 開交易後 WithTenantTx)。
	grant := func(t *testing.T, resource, action string) bool {
		t.Helper()
		tx, err := client.Tx(auth.WithRLS(context.Background(), scope))
		if err != nil {
			t.Fatalf("開租戶交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		cctx := authz.WithDB(context.Background(), client)
		cctx = dbtenant.WithTenantTx(authz.WithIdentity(cctx, id), tx)
		got, err := authz.PermissionGrantedTx(cctx, id.Roles, resource, action)
		if err != nil {
			t.Fatalf("PermissionGrantedTx(%s/%s): %v", resource, action, err)
		}
		return got
	}

	t.Run("交易內讀得到 role_permissions", func(t *testing.T) {
		// SystemScopeTx(data_scope=all)保證 scope 非空 → policy 成立;
		// 若接線正確,自訂角色的列讀得到(否則所有自訂角色一律無權)。
		if !grant(t, "company", "read") {
			t.Fatal("交易內查不到 company/read → 逐動作檢查在 RLS 下失效")
		}
	})

	t.Run("逐動作:有 update 無 create|delete", func(t *testing.T) {
		if !grant(t, "company", "update") {
			t.Error("DB 有 company/update,應為 true")
		}
		if grant(t, "company", "create") {
			t.Error("DB 無 company/create,應為 false")
		}
		if grant(t, "company", "delete") {
			t.Error("DB 無 company/delete,應為 false")
		}
	})

	t.Run("inverted 規則不構成授權", func(t *testing.T) {
		if grant(t, "role", "read") {
			t.Error("只有 inverted 規則,不得判為 true")
		}
	})

	t.Run("未授予資源一律拒", func(t *testing.T) {
		if grant(t, "sales_order", "write") {
			t.Error("DB 無 sales_order/write,應為 false")
		}
	})

	// 負向控制:交易**外**的 client 讀不到任何列 —— 證明「必須綁請求交易」不是形式要求。
	// 少了這條,上述正向斷言也可能是「碰巧 RLS 沒生效」而通過。
	t.Run("負向控制:交易外查詢回 0 列", func(t *testing.T) {
		n, err := client.RolePermission.Query().
			Where(rolepermission.ResourceEQ("company")).Count(context.Background())
		if err != nil {
			t.Fatalf("交易外查詢: %v", err)
		}
		if n != 0 {
			t.Fatalf("交易外應因 RLS 回 0 列(unscoped),got %d —— 少了這條,正向斷言無法證明接線", n)
		}
	})
}

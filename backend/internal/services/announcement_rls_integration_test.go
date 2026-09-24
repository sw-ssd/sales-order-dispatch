//go:build integration

package services

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationAnnouncementRLSProbe 公告的真 PG 探針(規則:必須 app_rw + 真 RLS):
//
//  1. schema 面:app_rw 對 announcements 有完整 CRUD + 序列權限;00044 policy 存在、
//     00045 已 ENABLE+FORCE。
//  2. 寫入面:company scope A 建自己公司列成功;建**別家公司**列被 WITH CHECK 擋(42501)
//     —— 這一條是負向控制,少了 RLS 接線不會紅(app_rw 才測得出)。
//  3. 讀取面:無 scope → 公司列 0 列(fail-closed)、全系統列仍可見(company NULL 人人可見);
//     company scope A → 只見自己公司;管理列表(真 handler)對 company_admin A 只回自己公司
//     的列(全系統列由服務層刻意排除 —— CMS「僅可管理」語意)。
func TestIntegrationAnnouncementRLSProbe(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)
	ctx := context.Background()

	// ① schema 面。
	adminSql, adminDB := openPGEntClientFromGoose(t, adminDSN)
	for _, q := range []struct {
		name string
		sql  string
	}{
		{"app_rw CRUD", `SELECT has_table_privilege('app_rw', 'announcements', 'SELECT')
			AND has_table_privilege('app_rw', 'announcements', 'INSERT')
			AND has_table_privilege('app_rw', 'announcements', 'UPDATE')
			AND has_table_privilege('app_rw', 'announcements', 'DELETE')`},
		{"app_rw 序列", `SELECT has_sequence_privilege('app_rw', 'announcements_id_seq', 'USAGE')
			AND has_sequence_privilege('app_rw', 'announcements_id_seq', 'SELECT')`},
		{"ENABLE", `SELECT relrowsecurity FROM pg_class WHERE relname = 'announcements'`},
		{"FORCE", `SELECT relforcerowsecurity FROM pg_class WHERE relname = 'announcements'`},
		{"policy 存在", `SELECT count(*) = 1 FROM pg_policies
			WHERE tablename = 'announcements' AND policyname = 'core_announcements_scope'`},
	} {
		var ok bool
		if err := adminSql.QueryRow(q.sql).Scan(&ok); err != nil {
			t.Fatalf("schema 檢查 %s: %v", q.name, err)
		}
		if !ok {
			t.Fatalf("schema 檢查失敗: %s", q.name)
		}
	}

	// 種子:兩家公司/部門(admin 連線,超級使用者;此時表尚未 ENABLE 前已 migrate 完成,
	// 但 admin 是 superuser → FORCE 亦繞過,寫入不受 policy 影響)。
	coA := adminDB.Company.Create().SetName("探針甲").SetIdentifier("ANP-A").SetStatus("active").SaveX(ctx)
	coB := adminDB.Company.Create().SetName("探針乙").SetIdentifier("ANP-B").SetStatus("active").SaveX(ctx)
	deptA := adminDB.Department.Create().SetName("甲部").SetCompanyID(coA.ID).SaveX(ctx)
	actorA := adminDB.User.Create().SetEmail("anp-a@example.test").SetName("甲管").
		SetRole("company_admin").SetStatus("active").SetPasswordHash("!").SetCompanyID(coA.ID).SaveX(ctx)
	_ = deptA
	_ = actorA

	// 全系統列(scope=all 寫入;之後驗無 scope 仍可見)。
	sysTitle := "全系統公告(探針)"
	if _, err := adminDB.Announcement.Create().SetTitle(sysTitle).SetType("news").Save(ctx); err != nil {
		t.Fatalf("建全系統列: %v", err)
	}

	// 真 handler:app_rw 連線 + dbtenant 裝飾(生產同構),身分與 scope 一併注入(AGENTS §9-16)。
	appClient := openAppRoleEntClient(t, adminDSN)
	newAnnServer := func(id authz.Identity, scope auth.RLSScope) salesorderv1connect.AnnouncementServiceClient {
		t.Helper()
		mux := http.NewServeMux()
		RegisterAnnouncementService(mux, appClient)
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := authz.WithIdentity(r.Context(), id)
			ctx = auth.WithRLS(ctx, scope)
			ctx = authz.WithDB(ctx, appClient)
			mux.ServeHTTP(w, r.WithContext(ctx))
		})
		ts := httptest.NewServer(h)
		t.Cleanup(ts.Close)
		return salesorderv1connect.NewAnnouncementServiceClient(http.DefaultClient, ts.URL)
	}
	scopeCompanyA := auth.RLSScope{
		UserID: uItoa(actorA.ID), CompanyID: uItoa(coA.ID),
		DepartmentID: uItoa(deptA.ID),
		DataScope:    auth.DataScopeCompany, CompanyActive: true,
	}
	idAdminA := authz.Identity{
		UserID: uItoa(actorA.ID), CompanyID: uItoa(coA.ID),
		Role: "company_admin", Roles: auth.RolesFor("company_admin"),
	}

	// 公司 A 列(admin 寫;superuser 繞過 RLS,僅作讀取面素材)。
	if _, err := adminDB.Announcement.Create().
		SetTitle("甲公司公告(探針)").SetType("news").SetCompanyID(coA.ID).Save(ctx); err != nil {
		t.Fatalf("建甲公司列: %v", err)
	}

	// ③' 讀取面(**必須最先跑**:這要是一條沒碰過任何 SET LOCAL 的乾淨連線池;
	// 同池連線被 scoped 交易碰過後,placeholder 會殘留空字串,無 scope 讀取會 22P02 ——
	// 生產無此路徑:每條請求交易都先 SET LOCAL)。無 scope → 公司列 0 列(fail-closed)、
	// 全系統列可見(company NULL 的可見性不依賴 scope)。
	rawApp := openAppRoleDB(t, adminDSN)
	var sysVisible, coRows int
	if err := rawApp.QueryRow(`SELECT
		count(*) FILTER (WHERE company_id IS NULL),
		count(*) FILTER (WHERE company_id IS NOT NULL)
		FROM announcements`).Scan(&sysVisible, &coRows); err != nil {
		t.Fatalf("無 scope 讀取: %v", err)
	}
	if sysVisible != 1 || coRows != 0 {
		t.Fatalf("無 scope 應只見全系統列(1/0),得到 %d/%d", sysVisible, coRows)
	}

	// ② 寫入面:company A 建自己公司列 → 成功(真 handler:dbtenant 交易 + SET LOCAL)。
	rpcA := newAnnServer(idAdminA, scopeCompanyA)
	if _, err := rpcA.CreateAnnouncement(ctx, connect.NewRequest(&v1.CreateAnnouncementRequest{
		Type: "news", Title: "甲公司公告", CompanyId: uItoa(coA.ID),
	})); err != nil {
		t.Fatalf("company scope A 建自己公司列應成功: %v", err)
	}

	// ②' 負向:scope A 的交易直接 SQL 寫**別家公司 B** → WITH CHECK 42501(真 RLS 才會紅)。
	if err := insertAnnouncementUnderScope(t, rawApp, coA.ID, coB.ID, "越權列"); !isRLSViolation(err) {
		t.Fatalf("跨公司寫入應 42501(RLS WITH CHECK),得到: %v", err)
	}

	// ③ 讀取面(管理列表,真 handler):company_admin A 只回**可管理**的自己公司 2 筆
	// (admin 建的 + handler 建的;全系統列由服務層刻意排除 —— CMS「僅可管理」語意)。
	listResp, err := rpcA.ListAnnouncements(ctx, connect.NewRequest(&v1.ListAnnouncementsRequest{}))
	if err != nil {
		t.Fatalf("company_admin A List: %v", err)
	}
	if got := int(listResp.Msg.GetTotal()); got != 2 {
		t.Fatalf("管理列表應只回可管理的 2 筆(不含全系統),got %d", got)
	}

	// ③' 原始可見性(company scope A):全系統列 + 自己公司 2 列 = 3;B 公司不可見。
	withScopeCount(t, rawApp, coA.ID, func(n int) {
		if n != 3 {
			t.Fatalf("company scope A 原始查詢應見 3 列(全系統+自己×2),got %d", n)
		}
	})
}

// withScopeCount 在 company scope A 的交易內數列數。
func withScopeCount(t *testing.T, db *sql.DB, coID int, fn func(int)) {
	t.Helper()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	for _, stmt := range []string{
		"SET LOCAL app.current_data_scope = 'company'",
		"SET LOCAL app.current_company_id = '" + uItoa(coID) + "'",
	} {
		if _, err := tx.Exec(stmt); err != nil {
			t.Fatalf("set scope: %v", err)
		}
	}
	var n int
	if err := tx.QueryRow(`SELECT count(*) FROM announcements`).Scan(&n); err != nil {
		t.Fatalf("scoped count: %v", err)
	}
	fn(n)
}

// insertAnnouncementUnderScope 在 scopeCoID 的 company scope 交易內 INSERT company_id=insertCoID 的列。
func insertAnnouncementUnderScope(t *testing.T, db *sql.DB, scopeCoID, insertCoID int, title string) error {
	t.Helper()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	for _, stmt := range []string{
		"SET LOCAL app.current_data_scope = 'company'",
		"SET LOCAL app.current_company_id = '" + uItoa(scopeCoID) + "'",
	} {
		if _, err := tx.Exec(stmt); err != nil {
			t.Fatalf("set scope: %v", err)
		}
	}
	_, err = tx.Exec(`INSERT INTO announcements (company_id, type, title) VALUES ($1, 'news', $2)`,
		insertCoID, title)
	return err
}

// isRLSViolation 判斷是否為 RLS WITH CHECK 擋下(SQLSTATE 42501;pgx 錯誤字串含碼)。
func isRLSViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "42501") || strings.Contains(msg, "row-level security")
}

//go:build integration

// PlatformAdminService 的真 PostgreSQL 端到端測試:真 admin(owner)連線 ＋ 真 platform schema
// ＋ 真 postgres.Admin。
//
// 為什麼假 store 不夠:平台域不用 ent(手寫 SQL),欄位名寫錯、numeric 掃成 float64、jsonb
// 參數沒轉、LEFT JOIN 把沒有價目的方案濾掉 —— 這些在記憶體假實作上永遠測不出來,而錯的代價是
// operator console 顯示錯誤的帳(金額、期別、誰改的)。
//
// 第二條主線是 RLS:跨租戶投影要讀 companies,而 companies 已 FORCE RLS —— 容器預設的
// postgres 是 superuser 會繞過 policy,所以本檔另建一個 NOSUPERUSER／NOBYPASSRLS 的角色
// (模擬「生產的 owner 不是 superuser」),釘住「admin 連線也必須自帶系統範圍」。
package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/operatorauth"
	platformstore "github.com/salesorder/sales-order-1.0/backend/internal/platform/store/postgres"
	platformv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// platformAdminSeed 為夾具的識別碼與關鍵時間。
type platformAdminSeed struct {
	operatorID int64
	activeID   int64 // 有訂閱的租戶
	noneID     int64 // 沒有訂閱的租戶
	deletedID  int64 // 已軟刪除的租戶(不得出現在任何列表)
	periodEnd  time.Time
}

// TestIntegrationPlatformAdmin 驗五個唯讀 RPC 的真 SQL 投影。
func TestIntegrationPlatformAdmin(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN) // 00029(platform schema)也走這條路徑
	admin := openRawDB(t, adminDSN)
	seed := seedPlatformAdmin(t, admin)

	svc := NewPlatformAdminService(platformstore.NewAdmin(admin))
	ctx := operatorauth.WithIdentity(t.Context(),
		operatorauth.Identity{OperatorID: seed.operatorID, Email: "ops-a@example.com", Role: "admin"})

	t.Run("ListTenants 投影", func(t *testing.T) {
		resp := listTenants(t, svc, ctx, &platformv1.ListTenantsRequest{Page: 1, PageSize: 20})
		if got := resp.GetPagination().GetTotal(); got != 2 {
			t.Fatalf("總數應為 2(已軟刪除的租戶不算),got %d", got)
		}
		byID := map[string]*platformv1.TenantSummary{}
		for _, x := range resp.GetTenants() {
			byID[x.GetCompanyId()] = x
		}
		if _, ok := byID[itoa(int(seed.deletedID))]; ok {
			t.Fatal("已軟刪除的公司不得出現在租戶列表")
		}

		active := byID[itoa(int(seed.activeID))]
		if active == nil {
			t.Fatalf("有訂閱的租戶必須在列表內: %v", byID)
		}
		if active.GetCompanyName() != "甲公司" || active.GetPlanCode() != "std" || active.GetPlanName() != "標準" {
			t.Fatalf("名稱／方案投影錯誤: %v", active)
		}
		if active.GetSubscriptionStatus() != "active" || active.GetSeatCount() != 8 {
			t.Fatalf("訂閱狀態／席位投影錯誤: %v", active)
		}
		// 期別 1 已過期未付 → overdue;本期到期日取「已開始的最後一期」(期別 2)。
		if active.GetCurrentPeriodEnd() != seed.periodEnd.UTC().Format(time.RFC3339) {
			t.Fatalf("本期到期日應為期別 2 的結束時間(%s),got %q",
				seed.periodEnd.UTC().Format(time.RFC3339), active.GetCurrentPeriodEnd())
		}
		if !active.GetOverdue() {
			t.Fatal("有已過期未付的期別 → overdue 必須為 true")
		}

		none := byID[itoa(int(seed.noneID))]
		if none == nil {
			t.Fatalf("沒有訂閱的租戶仍必須在列表內: %v", byID)
		}
		if none.GetSubscriptionStatus() != "none" || none.GetPlanCode() != "" || none.GetSeatCount() != 0 {
			t.Fatalf("未訂閱租戶應為 status=none 且無方案／席位: %v", none)
		}
		if none.GetCurrentPeriodEnd() != "" || none.GetOverdue() {
			t.Fatalf("未訂閱租戶不得有到期日／逾期: %v", none)
		}
	})

	t.Run("ListTenants 篩選與分頁", func(t *testing.T) {
		// 公司名稱模糊搜尋
		if got := listTenants(t, svc, ctx,
			&platformv1.ListTenantsRequest{Keyword: "甲公司", Page: 1, PageSize: 20}).GetPagination().GetTotal(); got != 1 {
			t.Fatalf("以公司名稱搜尋應命中 1 家,got %d", got)
		}
		// 識別碼模糊搜尋(keyword 的契約是「公司名稱／識別碼」)
		if got := listTenants(t, svc, ctx,
			&platformv1.ListTenantsRequest{Keyword: "PA-NONE", Page: 1, PageSize: 20}).GetPagination().GetTotal(); got != 1 {
			t.Fatalf("以識別碼搜尋應命中 1 家,got %d", got)
		}
		// 訂閱狀態篩選(未訂閱者的狀態是 none,不是空字串)
		noneOnly := listTenants(t, svc, ctx,
			&platformv1.ListTenantsRequest{Status: "none", Page: 1, PageSize: 20})
		if noneOnly.GetPagination().GetTotal() != 1 || noneOnly.GetTenants()[0].GetCompanyId() != itoa(int(seed.noneID)) {
			t.Fatalf("status=none 應只回未訂閱那家: %v", noneOnly.GetTenants())
		}
		// LIKE 的萬用字元必須被跳脫:keyword "%" 不得變成「符合全部」
		if got := listTenants(t, svc, ctx,
			&platformv1.ListTenantsRequest{Keyword: "%", Page: 1, PageSize: 20}).GetPagination().GetTotal(); got != 0 {
			t.Fatalf("keyword 的 %% 必須是字面字元(跳脫),got %d 筆", got)
		}
		// 分頁:page_size=1、page=2 → 總數不變、回第二家(證明 OFFSET 生效)
		second := listTenants(t, svc, ctx, &platformv1.ListTenantsRequest{Page: 2, PageSize: 1})
		if second.GetPagination().GetTotal() != 2 || len(second.GetTenants()) != 1 {
			t.Fatalf("分頁應回 1 筆而總數仍為 2,got %d 筆／total %d",
				len(second.GetTenants()), second.GetPagination().GetTotal())
		}
		if second.GetPagination().GetPage() != 2 || second.GetPagination().GetPageSize() != 1 {
			t.Fatalf("分頁回音錯誤: %v", second.GetPagination())
		}
	})

	t.Run("GetTenant 概況與例外", func(t *testing.T) {
		resp, err := svc.GetTenant(ctx, connect.NewRequest(&platformv1.GetTenantRequest{
			CompanyId: itoa(int(seed.activeID)),
		}))
		if err != nil {
			t.Fatalf("GetTenant: %v", err)
		}
		if got := resp.Msg.GetTenant().GetCompanyName(); got != "甲公司" {
			t.Fatalf("租戶名稱投影錯誤: %q", got)
		}
		overrides := map[string]*platformv1.TenantOverride{}
		for _, o := range resp.Msg.GetOverrides() {
			overrides[o.GetFeatureCode()] = o
		}
		if len(overrides) != 2 {
			t.Fatalf("已撤銷的例外不得回傳,應為 2 筆: %v", resp.Msg.GetOverrides())
		}
		printing := overrides["feature.printing"]
		if printing == nil || !printing.GetEnabledSet() || !printing.GetEnabled() {
			t.Fatalf("只關功能的例外(enabled=true、limit=NULL)映射錯誤: %v", printing)
		}
		if printing.GetLimitSet() {
			t.Fatalf("limit_value 為 NULL 時 limit_set 必須是 false(不是 0): %v", printing)
		}
		seats := overrides["limit.seats"]
		if seats == nil || seats.GetEnabledSet() {
			t.Fatalf("只給限額的例外(enabled=NULL)不得被當成 false: %v", seats)
		}
		if !seats.GetLimitSet() || seats.GetLimitValue() != 12 {
			t.Fatalf("限額例外映射錯誤: %v", seats)
		}
		if seats.GetReason() != "加購席位至 12" || seats.GetOwner() != "業務B" || seats.GetExpiresAt() != "" {
			t.Fatalf("例外的原因／承諾者／到期日映射錯誤: %v", seats)
		}

		// 查無此租戶,以及已軟刪除的租戶 → 視同不存在(不得以錯誤碼洩漏已刪除列的存在)。
		for name, id := range map[string]string{
			"不存在的 id":  itoa(int(seed.deletedID) + 100000),
			"已軟刪除的 id": itoa(int(seed.deletedID)),
		} {
			_, err := svc.GetTenant(ctx, connect.NewRequest(&platformv1.GetTenantRequest{CompanyId: id}))
			if connect.CodeOf(err) != connect.CodeNotFound {
				t.Fatalf("%s 應回 NotFound,got %v", name, err)
			}
			if got := errorInfoOf(t, err).GetCode(); got != "SYS-4002" {
				t.Fatalf("%s 必須是註冊碼 SYS-4002,got %q", name, got)
			}
		}
	})

	t.Run("ListPlans 現行價目", func(t *testing.T) {
		resp, err := svc.ListPlans(ctx, connect.NewRequest(&platformv1.ListPlansRequest{}))
		if err != nil {
			t.Fatalf("ListPlans: %v", err)
		}
		if len(resp.Msg.GetPlans()) != 2 {
			t.Fatalf("應回 2 個方案: %v", resp.Msg.GetPlans())
		}
		if resp.Msg.GetPlans()[0].GetCode() != "std" || resp.Msg.GetPlans()[1].GetCode() != "pro" {
			t.Fatalf("方案應依 sort_order 排序: %v", resp.Msg.GetPlans())
		}
		prices := map[string]*platformv1.PlanPrice{}
		for _, p := range resp.Msg.GetPlans()[0].GetPrices() {
			prices[p.GetBillingCycle()] = p
		}
		if len(prices) != 2 {
			t.Fatalf("同一方案每個計費週期只回現行價一筆: %v", resp.Msg.GetPlans()[0].GetPrices())
		}
		if got := prices["monthly"]; got.GetBasePrice() != "1000.00" || got.GetSeatPrice() != "200.00" || got.GetCurrency() != "TWD" {
			t.Fatalf("月繳現行價必須是較新那筆調價(1000.00/200.00),got %v", got)
		}
		if got := prices["yearly"].GetBasePrice(); got != "10000.00" {
			t.Fatalf("年繳價目映射錯誤: %q", got)
		}
		if got := resp.Msg.GetPlans()[1].GetPrices(); len(got) != 0 {
			t.Fatalf("沒有價目的方案仍要回方案本身且 prices 為空,got %v", got)
		}
	})

	t.Run("GetPlanEntitlements 權益與功能清單", func(t *testing.T) {
		resp, err := svc.GetPlanEntitlements(ctx, connect.NewRequest(
			&platformv1.GetPlanEntitlementsRequest{PlanCode: "std"}))
		if err != nil {
			t.Fatalf("GetPlanEntitlements: %v", err)
		}
		ents := map[string]*platformv1.FeatureEntitlement{}
		for _, e := range resp.Msg.GetEntitlements() {
			ents[e.GetFeatureCode()] = e
		}
		if got := ents["limit.seats"]; !got.GetEnabled() || !got.GetLimitSet() || got.GetLimitValue() != 10 {
			t.Fatalf("數值型權益映射錯誤: %v", got)
		}
		if got := ents["feature.printing"]; !got.GetEnabled() || got.GetLimitSet() {
			t.Fatalf("布林型權益(limit=NULL 代表不限)映射錯誤: %v", got)
		}
		// 功能清單必須是**完整**的(含未設權益者),否則權益矩陣顯示不出「未設定的格」。
		features := map[string]bool{}
		for _, f := range resp.Msg.GetFeatures() {
			features[f.GetCode()] = true
		}
		if len(features) != 3 || !features["feature.export"] {
			t.Fatalf("功能清單必須含未設權益的 feature.export: %v", resp.Msg.GetFeatures())
		}

		_, err = svc.GetPlanEntitlements(ctx, connect.NewRequest(
			&platformv1.GetPlanEntitlementsRequest{PlanCode: "nope"}))
		if connect.CodeOf(err) != connect.CodeNotFound {
			t.Fatalf("查無方案應回 NotFound,got %v", err)
		}
		if got := errorInfoOf(t, err).GetCode(); got != "SYS-4002" {
			t.Fatalf("必須是註冊碼 SYS-4002,got %q", got)
		}
	})

	t.Run("ListPlatformAudit 排序與篩選", func(t *testing.T) {
		all := listAudit(t, svc, ctx, &platformv1.ListPlatformAuditRequest{Page: 1, PageSize: 20})
		if got := all.GetPagination().GetTotal(); got != 3 {
			t.Fatalf("夾具應有 3 筆稽核,got %d", got)
		}
		// 新到舊
		for i, e := range all.GetEntries() {
			if want := []string{"login", "subscription.update", "login"}[i]; e.GetAction() != want {
				t.Fatalf("第 %d 筆應為 %s(created_at DESC),got %s", i, want, e.GetAction())
			}
		}
		// operator email 由 platform.operators JOIN 帶出
		if got := all.GetEntries()[1].GetOperatorEmail(); got != "ops-a@example.com" {
			t.Fatalf("稽核必須帶出操作者 email,got %q", got)
		}

		byTarget := listAudit(t, svc, ctx, &platformv1.ListPlatformAuditRequest{
			TargetType: "subscription", TargetId: "1", Page: 1, PageSize: 20,
		})
		if byTarget.GetPagination().GetTotal() != 1 || byTarget.GetEntries()[0].GetAction() != "subscription.update" {
			t.Fatalf("target_type／target_id 篩選錯誤: %v", byTarget.GetEntries())
		}
		if got := listAudit(t, svc, ctx, &platformv1.ListPlatformAuditRequest{
			TargetType: "company", Page: 1, PageSize: 20,
		}).GetPagination().GetTotal(); got != 0 {
			t.Fatalf("沒有此 target_type 的稽核應回 0 筆,got %d", got)
		}
		// 分頁:page_size=2 → 3 筆分兩頁
		first := listAudit(t, svc, ctx, &platformv1.ListPlatformAuditRequest{Page: 1, PageSize: 2})
		second := listAudit(t, svc, ctx, &platformv1.ListPlatformAuditRequest{Page: 2, PageSize: 2})
		if len(first.GetEntries()) != 2 || len(second.GetEntries()) != 1 ||
			first.GetPagination().GetTotal() != 3 || second.GetPagination().GetTotal() != 3 {
			t.Fatalf("稽核分頁錯誤: 第一頁 %d 筆／第二頁 %d 筆／total %d",
				len(first.GetEntries()), len(second.GetEntries()), first.GetPagination().GetTotal())
		}
	})

	t.Run("五個唯讀 RPC 不寫入任何稽核", func(t *testing.T) {
		before := countRows(t, admin, `SELECT count(*) FROM platform.audit_logs`)
		// 含失敗的呼叫(查無方案):失敗路徑同樣不得留痕。
		_, _ = svc.ListTenants(ctx, connect.NewRequest(&platformv1.ListTenantsRequest{Page: 1, PageSize: 20}))
		_, _ = svc.GetTenant(ctx, connect.NewRequest(&platformv1.GetTenantRequest{CompanyId: itoa(int(seed.activeID))}))
		_, _ = svc.ListPlans(ctx, connect.NewRequest(&platformv1.ListPlansRequest{}))
		_, _ = svc.GetPlanEntitlements(ctx, connect.NewRequest(&platformv1.GetPlanEntitlementsRequest{PlanCode: "nope"}))
		_, _ = svc.ListPlatformAudit(ctx, connect.NewRequest(&platformv1.ListPlatformAuditRequest{Page: 1, PageSize: 20}))
		if after := countRows(t, admin, `SELECT count(*) FROM platform.audit_logs`); after != before {
			t.Fatalf("唯讀 RPC 寫入了 %d 筆稽核", after-before)
		}
	})

	t.Run("RecordAudit 落 platform.audit_logs", func(t *testing.T) {
		identity := operatorauth.Identity{OperatorID: seed.operatorID, Email: "ops-a@example.com", Role: "admin"}
		if err := svc.recordPlatformAudit(ctx, identity, "operator.update", "operator",
			itoa(int(seed.operatorID)), "升為管理員",
			map[string]any{"role": "operator"}, map[string]any{"role": "admin"}); err != nil {
			t.Fatalf("recordPlatformAudit: %v", err)
		}

		var (
			operatorID                           int64
			action, targetType, targetID, reason string
			before, after                        []byte
		)
		row := admin.QueryRow(`SELECT operator_id, action, target_type, target_id, reason, before, after
			FROM platform.audit_logs WHERE action = 'operator.update' ORDER BY id DESC LIMIT 1`)
		if err := row.Scan(&operatorID, &action, &targetType, &targetID, &reason, &before, &after); err != nil {
			t.Fatalf("讀回稽核列: %v", err)
		}
		if operatorID != seed.operatorID {
			t.Fatalf("actor 必須是 operator id %d(平台操作沒有租戶使用者),got %d", seed.operatorID, operatorID)
		}
		if action != "operator.update" || targetType != "operator" ||
			targetID != itoa(int(seed.operatorID)) || reason != "升為管理員" {
			t.Fatalf("稽核欄位寫入錯誤: %q/%q/%q/%q", action, targetType, targetID, reason)
		}
		beforeMap, afterMap := map[string]any{}, map[string]any{}
		if err := json.Unmarshal(before, &beforeMap); err != nil {
			t.Fatalf("before 不是合法 jsonb: %v", err)
		}
		if err := json.Unmarshal(after, &afterMap); err != nil {
			t.Fatalf("after 不是合法 jsonb: %v", err)
		}
		if beforeMap["role"] != "operator" || afterMap["role"] != "admin" {
			t.Fatalf("before／after 寫反或未原樣保存: %v / %v", beforeMap, afterMap)
		}

		// 經 RPC 也查得到(把 target 篩選與寫入接起來)。
		byTarget := listAudit(t, svc, ctx, &platformv1.ListPlatformAuditRequest{
			TargetType: "operator", TargetId: itoa(int(seed.operatorID)), Page: 1, PageSize: 20,
		})
		if byTarget.GetPagination().GetTotal() != 2 || byTarget.GetEntries()[0].GetReason() != "升為管理員" {
			t.Fatalf("寫入的稽核必須查得到且在最前面: %v", byTarget.GetEntries())
		}

		// nil 的 before／after 寫 SQL NULL(不是 JSON 的 null:那會被讀成「值就是 null」)。
		if err := svc.recordPlatformAudit(ctx, identity, "operator.disable", "operator", "9", "", nil, nil); err != nil {
			t.Fatalf("recordPlatformAudit(nil): %v", err)
		}
		var beforeNull, afterNull bool
		if err := admin.QueryRow(`SELECT before IS NULL, after IS NULL FROM platform.audit_logs
			WHERE action = 'operator.disable'`).Scan(&beforeNull, &afterNull); err != nil {
			t.Fatalf("讀回稽核列: %v", err)
		}
		if !beforeNull || !afterNull {
			t.Fatal("nil 的 before／after 應寫 SQL NULL")
		}
	})

	t.Run("admin(owner,非 superuser)也必須自帶系統範圍", func(t *testing.T) {
		// companies 已 ENABLE ＋ FORCE RLS(00028):**連 owner 都受 policy 約束**,沒設
		// app.current_data_scope='all' 就靜默回 0 列(不報錯)。容器預設的 postgres 是
		// superuser、會繞過 RLS,故必須自建一個與生產同形的角色才測得到。
		db := openPlatformOwner(t, admin, adminDSN)
		ownerSvc := NewPlatformAdminService(platformstore.NewAdmin(db))

		resp := listTenants(t, ownerSvc, ctx, &platformv1.ListTenantsRequest{Page: 1, PageSize: 20})
		if resp.GetPagination().GetTotal() != 2 {
			t.Fatalf("admin 連線的跨租戶投影必須看得到 2 家租戶(少了系統範圍會靜默回 0),got %d",
				resp.GetPagination().GetTotal())
		}
		if _, err := ownerSvc.GetTenant(ctx, connect.NewRequest(&platformv1.GetTenantRequest{
			CompanyId: itoa(int(seed.activeID)),
		})); err != nil {
			t.Fatalf("admin 連線的租戶詳情: %v", err)
		}
		if _, err := ownerSvc.ListPlans(ctx, connect.NewRequest(&platformv1.ListPlansRequest{})); err != nil {
			t.Fatalf("admin 連線的方案查詢: %v", err)
		}
		// 平台稽核的寫入也必須在同一條路徑上成立(非 superuser 的 admin 連線)。
		if err := ownerSvc.recordPlatformAudit(ctx,
			operatorauth.Identity{OperatorID: seed.operatorID}, "operator.update", "operator", "1", "非 superuser 寫入", nil, nil); err != nil {
			t.Fatalf("admin 連線的稽核寫入: %v", err)
		}
	})
}

// listTenants 呼叫 ListTenants 並斷言成功。
func listTenants(t *testing.T, svc *PlatformAdminService, ctx context.Context,
	req *platformv1.ListTenantsRequest) *platformv1.ListTenantsResponse {
	t.Helper()
	resp, err := svc.ListTenants(ctx, connect.NewRequest(req))
	if err != nil {
		t.Fatalf("ListTenants(%v): %v", req, err)
	}
	return resp.Msg
}

// listAudit 呼叫 ListPlatformAudit 並斷言成功。
func listAudit(t *testing.T, svc *PlatformAdminService, ctx context.Context,
	req *platformv1.ListPlatformAuditRequest) *platformv1.ListPlatformAuditResponse {
	t.Helper()
	resp, err := svc.ListPlatformAudit(ctx, connect.NewRequest(req))
	if err != nil {
		t.Fatalf("ListPlatformAudit(%v): %v", req, err)
	}
	return resp.Msg
}

// seedPlatformAdmin 種出五個查詢各自需要的極端值:
//   - 有訂閱／無訂閱／已軟刪除的租戶各一(列表來源是 companies,不是 subscriptions);
//   - 同一方案同一計費週期的兩次調價(現行價必須取 effective_from 最新者)＋ 一個沒有價目的方案;
//   - 一期已過期未付(overdue)＋ 一期已開始未到期(current_period_end):兩個欄位不得同源;
//   - override 的三種形態:只關功能(enabled)、只給限額(limit)、已撤銷(不得回傳);
//   - 未設權益的 feature(權益矩陣必須仍拿得到它)。
func seedPlatformAdmin(t *testing.T, admin *sql.DB) platformAdminSeed {
	t.Helper()
	ctx := t.Context()

	var seed platformAdminSeed
	if err := admin.QueryRowContext(ctx, `INSERT INTO platform.operators (email, name, role, status)
		VALUES ('ops-a@example.com','甲管理員','admin','active') RETURNING id`).Scan(&seed.operatorID); err != nil {
		t.Fatalf("seed operator: %v", err)
	}
	var opBID int64
	if err := admin.QueryRowContext(ctx, `INSERT INTO platform.operators (email, name, role, status)
		VALUES ('ops-b@example.com','乙操作員','operator','active') RETURNING id`).Scan(&opBID); err != nil {
		t.Fatalf("seed operator: %v", err)
	}

	newCompany := func(name, identifier string, deleted bool) int64 {
		t.Helper()
		var id int64
		if err := admin.QueryRowContext(ctx,
			`INSERT INTO companies (name, identifier) VALUES ($1, $2) RETURNING id`,
			name, identifier).Scan(&id); err != nil {
			t.Fatalf("seed 公司 %s: %v", name, err)
		}
		if deleted {
			if _, err := admin.ExecContext(ctx,
				`UPDATE companies SET deleted_at = now() WHERE id = $1`, id); err != nil {
				t.Fatalf("軟刪除公司 %s: %v", name, err)
			}
		}
		return id
	}
	seed.activeID = newCompany("甲公司", "PA-ACTIVE", false)
	seed.noneID = newCompany("乙公司", "PA-NONE", false)
	seed.deletedID = newCompany("已刪除公司", "PA-DELETED", true)

	mustExec(t, admin, `INSERT INTO platform.features (code, type, unit, description) VALUES
		('limit.seats','integer','席','席位上線'),
		('feature.printing','boolean','','列印'),
		('feature.export','boolean','','匯出')`)

	var stdID, proID int64
	if err := admin.QueryRowContext(ctx,
		`INSERT INTO platform.plans (code, name, sort_order) VALUES ('std','標準',1) RETURNING id`).Scan(&stdID); err != nil {
		t.Fatalf("seed 方案 std: %v", err)
	}
	if err := admin.QueryRowContext(ctx,
		`INSERT INTO platform.plans (code, name, sort_order) VALUES ('pro','專業',2) RETURNING id`).Scan(&proID); err != nil {
		t.Fatalf("seed 方案 pro: %v", err)
	}

	// 價目:月繳兩次調價(較新者為現行價)＋ 年繳一次;pro 刻意沒有價目。
	mustExec(t, admin, `INSERT INTO platform.plan_prices
		(plan_id, billing_cycle, base_price, seat_price, effective_from) VALUES
		($1,'monthly',900.00,100.00, now() - interval '30 days'),
		($1,'monthly',1000.00,200.00, now()),
		($1,'yearly',10000.00,2000.00, now())`, stdID)
	mustExec(t, admin, `INSERT INTO platform.plan_entitlements (plan_id, feature_code, enabled, limit_value) VALUES
		($1,'limit.seats',true,10),
		($1,'feature.printing',true,NULL)`, stdID)

	var subID int64
	if err := admin.QueryRowContext(ctx, `INSERT INTO platform.subscriptions
		(company_id, plan_id, seat_count, billing_cycle, status)
		VALUES ($1, $2, 8, 'monthly', 'active') RETURNING id`, seed.activeID, stdID).Scan(&subID); err != nil {
		t.Fatalf("seed 訂閱: %v", err)
	}

	// 期別 1:已過期未付(overdue);期別 2:已開始、尚未到期(current_period_end)。
	seed.periodEnd = time.Now().Add(30 * 24 * time.Hour).Truncate(time.Second)
	mustExec(t, admin, `INSERT INTO platform.subscription_periods
		(subscription_id, period_no, period_start, period_end, plan_id,
		 unit_price, seat_price, seat_count, amount, status) VALUES
		($1,1, now() - interval '60 days', now() - interval '30 days', $2, 1000.00, 200.00, 8, 2600.00, 'open'),
		($1,2, now() - interval '30 days', $3, $2, 1000.00, 200.00, 8, 2600.00, 'open')`,
		subID, stdID, seed.periodEnd)

	mustExec(t, admin, `INSERT INTO platform.tenant_overrides
		(company_id, feature_code, enabled, limit_value, reason, owner, created_by) VALUES
		($1,'feature.printing',true,NULL,'簽約贈送列印','業務A',$2),
		($1,'limit.seats',NULL,12,'加購席位至 12','業務B',$2)`, seed.activeID, seed.operatorID)
	// 已撤銷的例外不得回傳(同一 feature_code 可再有一列,唯一索引只管未撤銷者)。
	mustExec(t, admin, `INSERT INTO platform.tenant_overrides
		(company_id, feature_code, enabled, limit_value, reason, owner, created_by, revoked_at) VALUES
		($1,'feature.printing',true,9,'已撤銷的例外','業務C',$2, now())`, seed.activeID, seed.operatorID)

	mustExec(t, admin, `INSERT INTO platform.audit_logs (operator_id, action, target_type, target_id, reason, created_at) VALUES
		($1,'login','operator',$3,'', now() - interval '2 hours'),
		($1,'subscription.update','subscription','1','調價', now() - interval '1 hour'),
		($2,'login','operator',$4,'', now())`,
		seed.operatorID, opBID, itoa(int(seed.operatorID)), itoa(int(opBID)))

	return seed
}

// openPlatformOwner 建一個 NOSUPERUSER／NOBYPASSRLS 的角色(模擬生產的 owner)並以它連線。
//
// 為什麼需要它:00028 對業務表施加 FORCE RLS,而 FORCE 讓 owner 也受 policy 約束(00025/00028
// 檔頭:「生產的 owner 不是 superuser」)。容器預設的 postgres 是 superuser,無論有沒有設
// app.current_data_scope 都讀得到全部列 —— 少了這個角色,「admin 連線忘了帶系統範圍」這個
// **靜默回 0 列**的缺陷在測試裡永遠不會紅。
func openPlatformOwner(t *testing.T, admin *sql.DB, adminDSN string) *sql.DB {
	t.Helper()
	const (
		role     = "platform_owner_probe"
		password = "platform-owner-probe"
	)
	mustExec(t, admin, `DROP ROLE IF EXISTS `+role)
	mustExec(t, admin, `CREATE ROLE `+role+` LOGIN PASSWORD '`+password+`' NOSUPERUSER NOBYPASSRLS`)
	mustExec(t, admin, `GRANT USAGE ON SCHEMA public, platform TO `+role)
	mustExec(t, admin, `GRANT SELECT ON companies TO `+role)
	mustExec(t, admin, `GRANT SELECT, INSERT ON ALL TABLES IN SCHEMA platform TO `+role)
	mustExec(t, admin, `GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA platform TO `+role)

	u, err := url.Parse(adminDSN)
	if err != nil {
		t.Fatalf("解析 DSN: %v", err)
	}
	u.User = url.UserPassword(role, password)
	db, err := sql.Open("pgx", u.String())
	if err != nil {
		t.Fatalf("連線(owner 角色): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// mustExec 執行一段 DDL／夾具 SQL。
func mustExec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("執行 SQL 失敗(%s): %v", query, err)
	}
}

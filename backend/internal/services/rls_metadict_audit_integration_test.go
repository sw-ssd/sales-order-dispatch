//go:build integration

package services

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	auditv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/audit/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/audit/v1/auditv1connect"
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1/customersv1connect"
	metadictv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/metadict/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/metadict/v1/metadictv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// metadictAuditRLSTables 是字典/稽核兩表(00027 必須 ENABLE + FORCE 者)。
var metadictAuditRLSTables = []string{"metadicts", "audit_logs"}

// TestIntegrationRLSMetadictAuditIsolation 以 app_rw 直連驗證字典與稽核的隔離(00027 ENABLE + FORCE):
// 未設 scope → 0 列(fail-closed);設 A 公司 → 只看得到 A;跨租戶寫入被 WITH CHECK 擋。
//
// 兩表語意不同,斷言分開:
//   - `metadicts` 單表兩層:`department_id IS NULL` 為系統預設列 —— USING 允許任何 scope 讀取,
//     但 WITH CHECK **不含** `department_id IS NULL`(00011 刻意較嚴),故公司/部門身分只能寫
//     自己部門的擴充列,寫系統預設列必須被擋(只有 super 的 scope=all 才行)。
//   - `audit_logs` 以 company_id 為租戶鍵:任何 scope 的請求都只能寫自己公司的稽核(讀取仍限 all/company)。
//
// 為何用 app_rw 而非 admin:容器/測試的 admin 連線是 superuser,PG 的 superuser 永遠繞過 RLS
// (FORCE 亦然)→ 以它連線根本測不到 fail-closed。app_rw 是 00022 建出的 NOBYPASSRLS 非 owner
// 業務角色,正是生產路徑的角色(T5/T6/T7 已立下同一作法)。
func TestIntegrationRLSMetadictAuditIsolation(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	defer func() { _ = admin.Close() }()

	coA := insertRLSCompany(t, admin, "A", "MDA")
	coB := insertRLSCompany(t, admin, "B", "MDB")
	deptA := insertRLSDepartment(t, admin, coA, "部門A")
	deptB := insertRLSDepartment(t, admin, coB, "部門B")
	userA := insertRLSUser(t, admin, coA, "md-a@t.com", "staff")
	userB := insertRLSUser(t, admin, coB, "md-b@t.com", "staff")
	// 部門擴充字典:A/B 各一列(code 於部門內唯一)。
	insertRLSMetadict(t, admin, "unit", "DA", "部門A單位", &deptA)
	insertRLSMetadict(t, admin, "unit", "DB", "部門B單位", &deptB)
	auditA := insertRLSAuditRow(t, admin, coA, userA, "create")
	auditB := insertRLSAuditRow(t, admin, coB, userB, "create")

	app := openAppRoleDB(t, adminDSN)

	t.Run("未設 scope → 稽核 0 列、字典僅系統預設列(fail-closed)", func(t *testing.T) {
		if n := countRows(t, app, `SELECT count(*) FROM audit_logs`); n != 0 {
			t.Fatalf("未設 scope 時 audit_logs 必須 0 列,got %d", n)
		}
		// metadicts 的 USING **永遠**允許系統預設列(department_id IS NULL,00011 的刻意設計 ——
		// 未帶 scope 的請求仍讀得到系統字典),但不得看到任何部門擴充列。
		var system, visible int
		if err := app.QueryRow(
			`SELECT count(*) FILTER (WHERE department_id IS NULL), count(*) FROM metadicts`).
			Scan(&system, &visible); err != nil {
			t.Fatalf("查字典: %v", err)
		}
		if system == 0 {
			t.Fatal("系統預設字典應存在(00011 seed)")
		}
		if visible != system {
			t.Fatalf("未設 scope 時不得看到部門擴充列,got 可見 %d 系統 %d", visible, system)
		}
		// 寫入:未設 scope 的稽核寫入必須被 WITH CHECK 擋(這是本波之前 auth_password 的處境)。
		if err := appExecScoped(t, app, nil,
			`INSERT INTO audit_logs (company_id, user_id, action, resource_type, resource_id)
			 VALUES ($1,$2,'update','scratch','0')`, coA, userA); !isRLSPolicyViolation(err) {
			t.Fatalf("未設 scope 時寫稽核必須被擋(SQLSTATE 42501),got %v", err)
		}
	})

	t.Run("scope=all:系統預設字典與任何公司稽核皆可讀寫", func(t *testing.T) {
		all := []string{`SET LOCAL app.current_data_scope = 'all'`}
		// 讀:平台範圍看得到所有公司的稽核列。
		tx := appTx(t, app, all)
		defer func() { _ = tx.Rollback() }()
		var cnt, distinct int
		if err := tx.QueryRow(`SELECT count(*), count(DISTINCT company_id) FROM audit_logs`).Scan(&cnt, &distinct); err != nil {
			t.Fatalf("查稽核: %v", err)
		}
		if cnt != 2 || distinct != 2 {
			t.Fatalf("all 應看得到兩家公司的 2 列稽核,got count=%d distinct=%d", cnt, distinct)
		}
		// 系統預設字典(department_id IS NULL)只有 all 可寫(00011 刻意較嚴)。
		if err := appExecScoped(t, app, all,
			`INSERT INTO metadicts (type, code, display_name, department_id, sort_order, is_active)
			 VALUES ('unit','X-ALL','平台別名',NULL,0,true)`); err != nil {
			t.Fatalf("以 all 身分寫入系統預設字典應可通過,got %v", err)
		}
		// super 跨公司:寫任何公司的稽核皆可(平台/維運路徑,如跨公司重置臨時密碼)。
		if err := appExecScoped(t, app, all,
			`INSERT INTO audit_logs (company_id, user_id, action, resource_type, resource_id)
			 VALUES ($1,$2,'update','scratch','5') RETURNING id`, coA, userA); err != nil {
			t.Fatalf("以 all 身分寫公司 A 的稽核應可通過,got %v", err)
		}
		if err := appExecScoped(t, app, all,
			`INSERT INTO audit_logs (company_id, user_id, action, resource_type, resource_id)
			 VALUES ($1,$2,'update','scratch','6') RETURNING id`, coB, userB); err != nil {
			t.Fatalf("以 all 身分寫公司 B 的稽核應可通過,got %v", err)
		}
	})

	t.Run("跨租戶:任何 scope 都讀不到他公司的稽核列(租戶邊界)", func(t *testing.T) {
		// 這是本檔**最重要**的斷言:放寬 USING 是為了讓 RETURNING 可寫自己公司的列,
		// 跨公司可見性絕不能被放寬(單列 by-id 探測與整表 count 兩條都要 0)。
		for _, tc := range []struct {
			name  string
			stmts []string
		}{
			{"company", companyScope(coA)},
			{"department", departmentScope(coA, deptA)},
			{"self", selfScope(coA, userA)},
		} {
			t.Run(tc.name, func(t *testing.T) {
				tx := appTx(t, app, tc.stmts)
				defer func() { _ = tx.Rollback() }()
				if !hasAuditRow(tx, auditA) {
					t.Fatalf("%s:應看得到自己公司(公司 A)的稽核列 id=%d", tc.name, auditA)
				}
				if hasAuditRow(tx, auditB) {
					t.Fatalf("%s(公司 A):不得看到他公司的稽核列 id=%d", tc.name, auditB)
				}
				var foreign int
				if err := tx.QueryRow(`SELECT count(*) FROM audit_logs WHERE company_id = $1`, coB).Scan(&foreign); err != nil {
					t.Fatalf("%s:查他公司稽核: %v", tc.name, err)
				}
				if foreign != 0 {
					t.Fatalf("%s(公司 A):他公司的稽核列必須不可見,got %d 列", tc.name, foreign)
				}
			})
		}
	})

	t.Run("scope=company A → 只見 A(稽核三值斷言)", func(t *testing.T) {
		// count/distinct/min 三值一起斷言:只看 DISTINCT 或取第一列會在未 ENABLE 時假通過(T7 §6.2)。
		tx := appTx(t, app, companyScope(coA))
		defer func() { _ = tx.Rollback() }()
		var cnt, distinct, minCo int
		if err := tx.QueryRow(
			`SELECT count(*), count(DISTINCT company_id), min(company_id) FROM audit_logs`).
			Scan(&cnt, &distinct, &minCo); err != nil {
			t.Fatalf("查稽核: %v", err)
		}
		if cnt != 1 || distinct != 1 || minCo != coA {
			t.Fatalf("公司 A 應只看得到自己那 1 列稽核,got count=%d distinct=%d min=%d(期望 1/1/%d)",
				cnt, distinct, minCo, coA)
		}
		if !hasAuditRow(tx, auditA) {
			t.Fatalf("公司 A 應看得到自己的稽核列 id=%d", auditA)
		}

		// 系統預設字典可讀(USING 含 department_id IS NULL),部門擴充列不可見(無部門脈絡)。
		var system, visible int
		if err := tx.QueryRow(
			`SELECT count(*) FILTER (WHERE department_id IS NULL), count(*) FROM metadicts`).
			Scan(&system, &visible); err != nil {
			t.Fatalf("查字典: %v", err)
		}
		if system == 0 {
			t.Fatal("系統預設字典在 company scope 下應可讀")
		}
		if visible != system {
			t.Fatalf("無部門脈絡時不得看到部門擴充列,got 可見 %d 系統 %d", visible, system)
		}
	})

	t.Run("company scope:寫系統預設字典被擋、寫自己公司稽核可過", func(t *testing.T) {
		// 系統預設字典:WITH CHECK 不含 department_id IS NULL(00011 刻意較嚴)。
		err := appExecScoped(t, app, companyScope(coA),
			`INSERT INTO metadicts (type, code, display_name, department_id, sort_order, is_active)
			 VALUES ('unit','X-SYS','別名',NULL,0,true)`)
		if !isRLSPolicyViolation(err) {
			t.Fatalf("以公司身分寫入系統預設字典必須被 WITH CHECK 擋(SQLSTATE 42501),got %v", err)
		}
		// 公司級身分的稽核寫自己公司 → 必須可寫(生產路徑:company_admin 的每個寫入都落稽核)。
		// 兩種寫法都要斷:PLAIN 與 `RETURNING id` —— ent 的 Create().Save() 一律是後者,而 PG 對
		// RETURNING 會套用 SELECT policy(USING),故「只改 WITH CHECK」不足以讓 dept/self 寫入過關。
		if err := appExecScoped(t, app, companyScope(coA),
			`INSERT INTO audit_logs (company_id, user_id, action, resource_type, resource_id)
			 VALUES ($1,$2,'update','scratch','1')`, coA, userA); err != nil {
			t.Fatalf("以公司身分寫自己公司的稽核(PLAIN)應可通過,got %v", err)
		}
		if err := appExecScoped(t, app, companyScope(coA),
			`INSERT INTO audit_logs (company_id, user_id, action, resource_type, resource_id)
			 VALUES ($1,$2,'update','scratch','1r') RETURNING id`, coA, userA); err != nil {
			t.Fatalf("以公司身分寫自己公司的稽核(RETURNING,ent 路徑)應可通過,got %v", err)
		}
		// 跨租戶:寫公司 B 的稽核 → 被擋(PLAIN 與 RETURNING 都要擋)。
		if err := appExecScoped(t, app, companyScope(coA),
			`INSERT INTO audit_logs (company_id, user_id, action, resource_type, resource_id)
			 VALUES ($1,$2,'update','scratch','2')`, coB, userA); !isRLSPolicyViolation(err) {
			t.Fatalf("以公司 A 的身分寫公司 B 的稽核必須被擋(SQLSTATE 42501),got %v", err)
		}
		if err := appExecScoped(t, app, companyScope(coA),
			`INSERT INTO audit_logs (company_id, user_id, action, resource_type, resource_id)
			 VALUES ($1,$2,'update','scratch','2r') RETURNING id`, coB, userA); !isRLSPolicyViolation(err) {
			t.Fatalf("以公司 A 的身分寫公司 B 的稽核(RETURNING)必須被擋(SQLSTATE 42501),got %v", err)
		}
	})

	t.Run("scope=department A:部門字典可讀可寫、他部門不可見不可寫", func(t *testing.T) {
		tx := appTx(t, app, departmentScope(coA, deptA))
		defer func() { _ = tx.Rollback() }()
		var total, own, foreign int
		if err := tx.QueryRow(
			`SELECT count(*), count(*) FILTER (WHERE department_id = $1),
			        count(*) FILTER (WHERE code = 'DB') FROM metadicts`, deptA).
			Scan(&total, &own, &foreign); err != nil {
			t.Fatalf("查字典: %v", err)
		}
		if own != 1 {
			t.Fatalf("部門 A 應看得到自己部門的 1 列擴充字典,got %d", own)
		}
		if foreign != 0 {
			t.Fatalf("部門 A 不得看到他部門(B 的 DB)的字典,got %d 列", foreign)
		}
		if total <= own {
			t.Fatalf("部門 A 應同時看得到系統預設字典,got 總計 %d(自己部門 %d)", total, own)
		}
		// 寫自己部門的擴充列 → 可(生產路徑:dept_admin 建部門字典)。
		if err := appExecScoped(t, app, departmentScope(coA, deptA),
			`INSERT INTO metadicts (type, code, display_name, department_id, sort_order, is_active)
			 VALUES ('unit','D-A2','部門A新單位',$1,0,true)`, deptA); err != nil {
			t.Fatalf("以部門 A 的身分寫自己部門的字典應可通過,got %v", err)
		}
		// 寫他部門的擴充列 → 被擋。
		if err := appExecScoped(t, app, departmentScope(coA, deptA),
			`INSERT INTO metadicts (type, code, display_name, department_id, sort_order, is_active)
			 VALUES ('unit','D-B2','部門B新單位',$1,0,true)`, deptB); !isRLSPolicyViolation(err) {
			t.Fatalf("以部門 A 的身分寫部門 B 的字典必須被 WITH CHECK 擋(SQLSTATE 42501),got %v", err)
		}
		// 稽核:部門層級的請求同樣會落稽核(D18)→ 寫自己公司可、寫他公司被擋。
		// 這裡刻意用 `RETURNING id`:ent 的 Create().Save() 就是這個形狀,而 PG 對 RETURNING 會
		// 套用 SELECT policy(USING)—— 這正是本檔需要同時放寬 USING 的原因(見 migration 註解)。
		if err := appExecScoped(t, app, departmentScope(coA, deptA),
			`INSERT INTO audit_logs (company_id, user_id, action, resource_type, resource_id)
			 VALUES ($1,$2,'update','scratch','3') RETURNING id`, coA, userA); err != nil {
			t.Fatalf("以部門身分寫自己公司的稽核(ent 路徑)應可通過,got %v", err)
		}
		if err := appExecScoped(t, app, departmentScope(coA, deptA),
			`INSERT INTO audit_logs (company_id, user_id, action, resource_type, resource_id)
			 VALUES ($1,$2,'update','scratch','4') RETURNING id`, coB, userA); !isRLSPolicyViolation(err) {
			t.Fatalf("以部門 A 的身分寫公司 B 的稽核必須被擋(SQLSTATE 42501),got %v", err)
		}
	})

	t.Run("scope=self:自助路徑的稽核可寫,跨公司不可寫", func(t *testing.T) {
		// 客戶(customer/guest)的 data_scope=self:自助寫入(改密碼)也要落自己公司的稽核。
		// 同樣用 `RETURNING id`(ent 路徑):改密碼就是這條路徑,先前 42501 也正是在此。
		if err := appExecScoped(t, app, selfScope(coA, userA),
			`INSERT INTO audit_logs (company_id, user_id, action, resource_type, resource_id)
			 VALUES ($1,$2,'update','user','9') RETURNING id`, coA, userA); err != nil {
			t.Fatalf("以 self 身分寫自己公司的稽核(ent 路徑)應可通過,got %v", err)
		}
		if err := appExecScoped(t, app, selfScope(coA, userA),
			`INSERT INTO audit_logs (company_id, user_id, action, resource_type, resource_id)
			 VALUES ($1,$2,'update','user','8') RETURNING id`, coB, userA); !isRLSPolicyViolation(err) {
			t.Fatalf("以 self 身分寫公司 B 的稽核必須被 WITH CHECK 擋(SQLSTATE 42501),got %v", err)
		}
	})
}

// mdAuditClients 為字典/稽核 domain 的兩個 RPC client(同一 mux、同一身分與 scope)。
type mdAuditClients struct {
	metadicts metadictv1connect.MetadictServiceClient
	audits    auditv1connect.AuditServiceClient
}

// TestIntegrationMetadictAuditUnderAppRole 以 app_rw + 請求層租戶交易跑**真 handler**,逐條走過
// 字典域已遷移的路徑(Create/Get/List/Update/Delete/ListOptions,含系統級與部門級兩種身分)
// 與稽核查詢,並斷言每條寫入的稽核列都真的落地(D18 同事務)。
//
// 為何既有測試擋不住:metadict_service_test.go／audit_service_test.go 走 sqlite(無 RLS 語意);
// 其餘整合測試的連線是容器 superuser(PG 的 superuser 永遠繞過 RLS)→ 路徑漏掛 tenant client 也照樣全綠。
func TestIntegrationMetadictAuditUnderAppRole(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	defer func() { _ = admin.Close() }()

	coA := insertRLSCompany(t, admin, "A", "MDA")
	coB := insertRLSCompany(t, admin, "B", "MDB")
	deptA := insertRLSDepartment(t, admin, coA, "部門A")
	superID := insertRLSUser(t, admin, coA, "sup-md@t.com", "super")
	companyAdminID := insertRLSUser(t, admin, coA, "ca-md@t.com", "company_admin")
	deptAdminID := insertRLSUser(t, admin, coA, "da-md@t.com", "dept_admin")
	userB := insertRLSUser(t, admin, coB, "b-md@t.com", "staff")
	insertRLSAuditRow(t, admin, coB, userB, "create")
	var systemUnitID int
	if err := admin.QueryRow(
		`SELECT id FROM metadicts WHERE department_id IS NULL AND type = 'unit' ORDER BY id LIMIT 1`).
		Scan(&systemUnitID); err != nil {
		t.Fatalf("取系統預設字典: %v", err)
	}

	// app_rw + dbtenant.NewClient(RLS 裝飾器生效);連線池上限 1 見 newMetadictAuditAppRoleServer。
	pool := openAppRoleDB(t, adminDSN)
	pool.SetMaxOpenConns(1)
	client := dbtenant.NewClient(pool)
	t.Cleanup(func() { _ = client.Close() })

	superC := newMetadictAuditAppRoleServer(t, client,
		authz.Identity{UserID: itoa(superID), CompanyID: itoa(coA), Role: "super", Roles: []string{"super"}},
		auth.RLSScope{UserID: itoa(superID), CompanyID: itoa(coA), DataScope: auth.DataScopeAll, CompanyActive: true}, true)
	deptC := newMetadictAuditAppRoleServer(t, client,
		authz.Identity{UserID: itoa(deptAdminID), CompanyID: itoa(coA), DepartmentID: itoa(deptA), Role: "dept_admin", Roles: []string{"dept_admin"}},
		auth.RLSScope{UserID: itoa(deptAdminID), CompanyID: itoa(coA), DepartmentID: itoa(deptA), DataScope: auth.DataScopeDepartment, CompanyActive: true}, true)
	companyC := newMetadictAuditAppRoleServer(t, client,
		authz.Identity{UserID: itoa(companyAdminID), CompanyID: itoa(coA), Role: "company_admin", Roles: []string{"company_admin"}},
		auth.RLSScope{UserID: itoa(companyAdminID), CompanyID: itoa(coA), DataScope: auth.DataScopeCompany, CompanyActive: true}, true)

	// ---- super:系統級字典全生命週期 ----
	created := callRPC(t, func(ctx context.Context) (*metadictv1.Metadict, error) {
		res, err := superC.metadicts.CreateMetadict(ctx, connect.NewRequest(&metadictv1.CreateMetadictRequest{
			Type: "unit", Code: "T8-SYS", DisplayName: "系統噸", SortOrder: 7, IsActive: true,
		}))
		if err != nil {
			return nil, err
		}
		return res.Msg.GetMetadict(), nil
	})
	if created.GetDepartmentId() != "" {
		t.Fatalf("super 建立應為系統級(department 空),got %q", created.GetDepartmentId())
	}
	createdID := mustItoa(t, created.GetId())
	if n := countRows(t, admin, `SELECT count(*) FROM metadicts WHERE id = $1 AND department_id IS NULL`, createdID); n != 1 {
		t.Fatalf("系統級字典應落地(department_id IS NULL),got %d 列", n)
	}
	assertAuditRow(t, admin, "metadict", "create", created.GetId(), coA)

	got := callRPC(t, func(ctx context.Context) (*metadictv1.Metadict, error) {
		res, err := superC.metadicts.GetMetadict(ctx, connect.NewRequest(&metadictv1.GetMetadictRequest{Id: created.GetId()}))
		if err != nil {
			return nil, err
		}
		return res.Msg.GetMetadict(), nil
	})
	if got.GetCode() != "T8-SYS" {
		t.Fatalf("GetMetadict 應回傳剛建立的字典,got %+v", got)
	}

	list := callRPC(t, func(ctx context.Context) ([]*metadictv1.Metadict, error) {
		res, err := superC.metadicts.ListMetadicts(ctx, connect.NewRequest(&metadictv1.ListMetadictsRequest{
			Page: 1, PageSize: 50, Type: "unit",
		}))
		if err != nil {
			return nil, err
		}
		return res.Msg.GetItems(), nil
	})
	if !metadictIDs(list)[created.GetId()] {
		t.Fatalf("ListMetadicts 應包含剛建立的字典(id=%s),got %v", created.GetId(), metadictIDs(list))
	}

	opts := callRPC(t, func(ctx context.Context) ([]*metadictv1.Option, error) {
		res, err := superC.metadicts.ListOptions(ctx, connect.NewRequest(&metadictv1.ListOptionsRequest{Type: "unit", Keyword: "T8-SYS"}))
		if err != nil {
			return nil, err
		}
		return res.Msg.GetOptions(), nil
	})
	if len(opts) != 1 || opts[0].GetCode() != "T8-SYS" {
		t.Fatalf("ListOptions 應只回關鍵字命中的 1 項,got %+v", opts)
	}

	upd := callRPC(t, func(ctx context.Context) (*metadictv1.Metadict, error) {
		name := "系統噸(改)"
		res, err := superC.metadicts.UpdateMetadict(ctx, connect.NewRequest(&metadictv1.UpdateMetadictRequest{
			Id: created.GetId(), DisplayName: &name,
		}))
		if err != nil {
			return nil, err
		}
		return res.Msg.GetMetadict(), nil
	})
	if upd.GetDisplayName() != "系統噸(改)" {
		t.Fatalf("UpdateMetadict 應回傳新名稱,got %q", upd.GetDisplayName())
	}
	if n := countRows(t, admin, `SELECT count(*) FROM metadicts WHERE id = $1 AND display_name = '系統噸(改)'`, createdID); n != 1 {
		t.Fatalf("更新應落地,got %d 列", n)
	}
	assertAuditRow(t, admin, "metadict", "update", created.GetId(), coA)

	if err := callRPCErr(t, func(ctx context.Context) error {
		_, err := superC.metadicts.DeleteMetadict(ctx, connect.NewRequest(&metadictv1.DeleteMetadictRequest{Id: created.GetId()}))
		return err
	}); err != nil {
		t.Fatalf("DeleteMetadict: %v", err)
	}
	if n := countRows(t, admin, `SELECT count(*) FROM metadicts WHERE id = $1 AND deleted_at IS NOT NULL`, createdID); n != 1 {
		t.Fatalf("刪除應為軟刪除,got %d 列", n)
	}
	assertAuditRow(t, admin, "metadict", "delete", created.GetId(), coA)

	// ---- dept_admin:部門級字典 ----
	deptRow := callRPC(t, func(ctx context.Context) (*metadictv1.Metadict, error) {
		res, err := deptC.metadicts.CreateMetadict(ctx, connect.NewRequest(&metadictv1.CreateMetadictRequest{
			Type: "unit", Code: "T8-DEPT", DisplayName: "部門單位", SortOrder: 3, IsActive: true,
		}))
		if err != nil {
			return nil, err
		}
		return res.Msg.GetMetadict(), nil
	})
	if deptRow.GetDepartmentId() != itoa(deptA) {
		t.Fatalf("部門管理員建立應自動帶自己部門 %d,got %q", deptA, deptRow.GetDepartmentId())
	}
	if n := countRows(t, admin, `SELECT count(*) FROM metadicts WHERE id = $1 AND department_id = $2`,
		mustItoa(t, deptRow.GetId()), deptA); n != 1 {
		t.Fatalf("部門級字典應落地於部門 %d,got %d 列", deptA, n)
	}
	assertAuditRow(t, admin, "metadict", "create", deptRow.GetId(), coA)

	deptList := callRPC(t, func(ctx context.Context) ([]*metadictv1.Metadict, error) {
		res, err := deptC.metadicts.ListMetadicts(ctx, connect.NewRequest(&metadictv1.ListMetadictsRequest{Page: 1, PageSize: 50}))
		if err != nil {
			return nil, err
		}
		return res.Msg.GetItems(), nil
	})
	ids := metadictIDs(deptList)
	if !ids[deptRow.GetId()] || !ids[itoa(systemUnitID)] {
		t.Fatalf("部門管理員應看得到系統預設 + 自己部門的列,got %v", ids)
	}

	inactive := false
	if err := callRPCErr(t, func(ctx context.Context) error {
		_, err := deptC.metadicts.UpdateMetadict(ctx, connect.NewRequest(&metadictv1.UpdateMetadictRequest{
			Id: deptRow.GetId(), IsActive: &inactive,
		}))
		return err
	}); err != nil {
		t.Fatalf("部門管理員應能更新自己部門的字典: %v", err)
	}
	assertAuditRow(t, admin, "metadict", "update", deptRow.GetId(), coA)

	// ---- 稽核查詢:company_admin 只看自己公司 ----
	audits := callRPC(t, func(ctx context.Context) ([]*auditv1.AuditLog, error) {
		res, err := companyC.audits.ListAuditLogs(ctx, connect.NewRequest(&auditv1.ListAuditLogsRequest{Page: 1, PageSize: 100}))
		if err != nil {
			return nil, err
		}
		return res.Msg.GetItems(), nil
	})
	if len(audits) == 0 {
		t.Fatal("company_admin 應看得到自己公司的稽核")
	}
	for _, a := range audits {
		if a.GetCompanyId() != itoa(coA) {
			t.Fatalf("稽核清單混入他公司的列:company_id=%s(期望 %d)", a.GetCompanyId(), coA)
		}
	}
	wantAudits := countRows(t, admin, `SELECT count(*) FROM audit_logs WHERE company_id = $1`, coA)
	if len(audits) != wantAudits {
		t.Fatalf("company_admin 應看到公司 A 的全部 %d 列稽核(近 3 個月),got %d", wantAudits, len(audits))
	}

	// 部門管理員不得查稽核(app 層守門,與 RLS 的讀取限制方向一致)。
	if err := callRPCErr(t, func(ctx context.Context) error {
		_, err := deptC.audits.ListAuditLogs(ctx, connect.NewRequest(&auditv1.ListAuditLogsRequest{Page: 1, PageSize: 10}))
		return err
	}); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("dept_admin 查稽核應 permission_denied,got %v", err)
	}

	// ---- 負向對照:同一連線池、不注入 RLS scope ----
	noScope := newMetadictAuditAppRoleServer(t, client,
		authz.Identity{Role: "super", Roles: []string{"super"}}, auth.RLSScope{}, false)
	empty := callRPC(t, func(ctx context.Context) ([]*metadictv1.Metadict, error) {
		res, err := noScope.metadicts.ListMetadicts(ctx, connect.NewRequest(&metadictv1.ListMetadictsRequest{Page: 1, PageSize: 50}))
		if err != nil {
			return nil, err
		}
		return res.Msg.GetItems(), nil
	})
	// metadicts 的 USING 永遠允許系統預設列(00011),故「無 scope」下仍看得到系統字典 ——
	// 但**看不到任何部門擴充列**(本測試 walk 期間建立的 T8-DEPT 屬 deptA)。
	if len(empty) == 0 {
		t.Fatal("無 scope 時仍應看得到系統預設字典(00011 的 USING 允許 department_id IS NULL)")
	}
	for _, m := range empty {
		if m.GetDepartmentId() != "" {
			t.Fatalf("未注入 scope 時不得看到部門擴充列,got id=%s department=%s", m.GetId(), m.GetDepartmentId())
		}
	}
	noScopeAudits := callRPC(t, func(ctx context.Context) ([]*auditv1.AuditLog, error) {
		res, err := noScope.audits.ListAuditLogs(ctx, connect.NewRequest(&auditv1.ListAuditLogsRequest{Page: 1, PageSize: 10}))
		if err != nil {
			return nil, err
		}
		return res.Msg.GetItems(), nil
	})
	if len(noScopeAudits) != 0 {
		t.Fatalf("未注入 scope 時不得看到任何稽核,got %d 筆", len(noScopeAudits))
	}

	auditsBefore := countRows(t, admin, `SELECT count(*) FROM audit_logs`)
	if err := callRPCErr(t, func(ctx context.Context) error {
		_, err := noScope.metadicts.CreateMetadict(ctx, connect.NewRequest(&metadictv1.CreateMetadictRequest{
			Type: "unit", Code: "T8-NOSCOPE", DisplayName: "無 scope", IsActive: true,
		}))
		return err
	}); err == nil {
		t.Fatal("未注入 scope 時建立字典應失敗(稽核寫入被 WITH CHECK 擋)")
	}
	if n := countRows(t, admin, `SELECT count(*) FROM metadicts WHERE code = 'T8-NOSCOPE'`); n != 0 {
		t.Fatalf("未注入 scope 時不得留下任何字典列,got %d", n)
	}
	if n := countRows(t, admin, `SELECT count(*) FROM audit_logs`); n != auditsBefore {
		t.Fatalf("未注入 scope 時不得留下稽核列,got %d(原 %d)", n, auditsBefore)
	}
}

// TestIntegrationDepartmentScopeWriteAuditLanding 以 **department scope**(dept_admin/staff 的生產
// scope)走一條**已啟用**域(客戶域,00024)的寫入路徑:建立客戶必須成功,且稽核列(D18 同交易)
// 必須落地。
//
// 為何需要這一條:T5/T6/T7 的探針全部以 company scope 身分走,「department/self scope 的稽核寫入
// 被 core_audit_logs_scope 擋下」因而一路潛伏(整個缺陷是可觀測的行為差異,沒有這條就看不出來)。
func TestIntegrationDepartmentScopeWriteAuditLanding(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	defer func() { _ = admin.Close() }()

	coA := insertRLSCompany(t, admin, "A", "DEPT-AUDIT")
	setCompanyCodePrefix(t, admin, coA, "DEPA")
	deptA := insertRLSDepartment(t, admin, coA, "部門A")
	actor := insertRLSUser(t, admin, coA, "da-cust@example.com", "dept_admin")
	// 客戶的 default_sales_rep_id 必須是同部門業務 → 把夾具業務掛到 deptA。
	setRLSUserDepartment(t, admin, actor, deptA)

	client := dbtenant.NewClient(openAppRoleDB(t, adminDSN))
	t.Cleanup(func() { _ = client.Close() })

	mux := http.NewServeMux()
	RegisterCustomerServices(mux, client, "http://localhost:3000")
	id := authz.Identity{
		UserID: itoa(actor), CompanyID: itoa(coA), DepartmentID: itoa(deptA),
		Role: "dept_admin", Roles: []string{"dept_admin"},
	}
	scope := auth.RLSScope{
		UserID: itoa(actor), CompanyID: itoa(coA), DepartmentID: itoa(deptA),
		DataScope: auth.DataScopeDepartment, CompanyActive: true,
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = auth.WithRLS(ctx, scope)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	svc := customersv1connect.NewCustomerServiceClient(http.DefaultClient, ts.URL)

	created := callRPC(t, func(ctx context.Context) (*customersv1.Customer, error) {
		res, err := svc.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{
			Name: "部門客戶", DefaultSalesRepId: itoa(actor),
		}))
		if err != nil {
			return nil, err
		}
		return res.Msg.GetCustomer(), nil
	})
	if created.GetCompanyId() != itoa(coA) || created.GetDepartmentId() != itoa(deptA) {
		t.Fatalf("部門管理員建立的客戶應屬公司 %d / 部門 %d,got company=%q department=%q",
			coA, deptA, created.GetCompanyId(), created.GetDepartmentId())
	}
	if n := countRows(t, admin, `SELECT count(*) FROM customers WHERE id = $1 AND department_id = $2`,
		mustItoa(t, created.GetId()), deptA); n != 1 {
		t.Fatalf("客戶應以部門 %d 落地,got %d 列", deptA, n)
	}
	// D18:同一請求交易的稽核列必須落地(政策修正前,這個 INSERT 被 WITH CHECK 擋 → 整筆建檔失敗)。
	assertAuditRow(t, admin, "customer", "create", created.GetId(), coA)
}

// TestIntegrationRLSMetadictAuditEnableMigrationDown 00027 的 Up/Down 必須對稱:Up 對兩表
// ENABLE + FORCE,Down 先 NO FORCE 再 DISABLE(少寫 NO FORCE 會留下 FORCE 旗標、少寫 DISABLE
// 則 RLS 仍生效 —— 兩者都讓回退後的環境與 00026 的狀態不一致),且可重複套用。
//
// 另外斷言兩個 policy 在 Down 之後**仍存在**:policy 由 00011/00023 定義、00025 正規化,
// Down 若順手 DROP POLICY 就會讓回退後的環境失去定義(旗標上看不出來)。
func TestIntegrationRLSMetadictAuditEnableMigrationDown(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)

	admin := openRawDB(t, dsn)
	defer func() { _ = admin.Close() }()

	assertMetadictAuditRLSFlags(t, admin, true, true)
	assertMetadictAuditPoliciesPresent(t, admin)

	migrateBusinessDownTo(t, dsn, "26") // 僅回退 00027
	assertMetadictAuditRLSFlags(t, admin, false, false)
	assertMetadictAuditPoliciesPresent(t, admin)

	migrateBusinessUp(t, dsn) // 回退後必須能重新套用(Up 冪等)
	assertMetadictAuditRLSFlags(t, admin, true, true)
}

// assertMetadictAuditRLSFlags 斷言字典/稽核兩表的 ENABLE/FORCE 旗標。
func assertMetadictAuditRLSFlags(t *testing.T, db *sql.DB, enabled, forced bool) {
	t.Helper()
	for _, table := range metadictAuditRLSTables {
		var e, f bool
		if err := db.QueryRow(
			`SELECT relrowsecurity, relforcerowsecurity FROM pg_class WHERE oid = $1::regclass`, table).Scan(&e, &f); err != nil {
			t.Fatalf("查 %s 的 RLS 旗標: %v", table, err)
		}
		if e != enabled || f != forced {
			t.Fatalf("%s 的 RLS 旗標應為 ENABLE=%v FORCE=%v,got ENABLE=%v FORCE=%v", table, enabled, forced, e, f)
		}
	}
}

// assertMetadictAuditPoliciesPresent 斷言兩表的 policy 仍存在(00027 不得 DROP 或移除 policy)。
func assertMetadictAuditPoliciesPresent(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, p := range []struct{ table, policy string }{
		{"metadicts", "core_metadicts_scope"},
		{"audit_logs", "core_audit_logs_scope"},
	} {
		var n int
		if err := db.QueryRow(
			`SELECT count(*) FROM pg_policies WHERE schemaname = 'public' AND tablename = $1 AND policyname = $2`,
			p.table, p.policy).Scan(&n); err != nil {
			t.Fatalf("查 policy(%s/%s): %v", p.table, p.policy, err)
		}
		if n != 1 {
			t.Fatalf("%s 的 policy %s 必須存在(00027 不得 DROP policy)", p.table, p.policy)
		}
	}
}

// insertRLSDepartment 以 admin 連線(夾具)建一間部門並回傳 id。
func insertRLSDepartment(t *testing.T, db *sql.DB, companyID int, name string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO departments (name, company_departments) VALUES ($1, $2) RETURNING id`,
		name, companyID).Scan(&id); err != nil {
		t.Fatalf("建部門 %s: %v", name, err)
	}
	return id
}

// insertRLSMetadict 以 admin 連線(夾具)建一列字典;departmentID 為 nil 表系統預設列。
func insertRLSMetadict(t *testing.T, db *sql.DB, typ, code, name string, departmentID *int) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO metadicts (type, code, display_name, department_id, sort_order, is_active)
		 VALUES ($1, $2, $3, $4, 0, true) RETURNING id`, typ, code, name, departmentID).Scan(&id); err != nil {
		t.Fatalf("建字典 %s/%s: %v", typ, code, err)
	}
	return id
}

// setRLSUserDepartment 以 admin 連線把使用者掛到指定部門(default_sales_rep_id 的同部門檢查用)。
func setRLSUserDepartment(t *testing.T, db *sql.DB, userID, departmentID int) {
	t.Helper()
	if _, err := db.Exec(`UPDATE users SET department_users = $2 WHERE id = $1`, userID, departmentID); err != nil {
		t.Fatalf("設定使用者 %d 的部門: %v", userID, err)
	}
}

// insertRLSAuditRow 以 admin 連線(夾具)建一列稽核並回傳 id。
func insertRLSAuditRow(t *testing.T, db *sql.DB, companyID, userID int, action string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO audit_logs (company_id, user_id, action, resource_type, resource_id)
		 VALUES ($1, $2, $3, 'customer', 'fixture') RETURNING id`, companyID, userID, action).Scan(&id); err != nil {
		t.Fatalf("建稽核(公司 %d): %v", companyID, err)
	}
	return id
}

// assertAuditRow 以 admin 連線斷言某筆業務寫入確實留下稽核列(D18 同事務寫入的落地證據)。
func assertAuditRow(t *testing.T, db *sql.DB, resourceType, action, resourceID string, companyID int) {
	t.Helper()
	n := countRows(t, db,
		`SELECT count(*) FROM audit_logs
		  WHERE resource_type = $1 AND action = $2 AND resource_id = $3 AND company_id = $4`,
		resourceType, action, resourceID, companyID)
	if n != 1 {
		t.Fatalf("%s/%s(resource_id=%s, 公司 %d)應留下恰 1 列稽核,got %d", resourceType, action, resourceID, companyID, n)
	}
}

// companyScope 回傳公司範圍的 SET LOCAL 語句(與 T5 的 setAppScope 同義,供「一句一交易」的斷言使用)。
func companyScope(companyID int) []string {
	return []string{
		`SET LOCAL app.current_data_scope = 'company'`,
		`SET LOCAL app.current_company_id = '` + itoa(companyID) + `'`,
	}
}

// departmentScope 回傳部門範圍的 SET LOCAL 語句(dept_admin/staff 的生產 scope)。
func departmentScope(companyID, departmentID int) []string {
	return []string{
		`SET LOCAL app.current_data_scope = 'department'`,
		`SET LOCAL app.current_company_id = '` + itoa(companyID) + `'`,
		`SET LOCAL app.current_department_id = '` + itoa(departmentID) + `'`,
	}
}

// selfScope 回傳 self 範圍的 SET LOCAL 語句(customer/guest 的生產 scope)。
func selfScope(companyID, userID int) []string {
	return []string{
		`SET LOCAL app.current_data_scope = 'self'`,
		`SET LOCAL app.current_company_id = '` + itoa(companyID) + `'`,
		`SET LOCAL app.current_user_id = '` + itoa(userID) + `'`,
	}
}

// appTx 以 app_rw 開一個交易並套用 scope(交易由呼叫端回滾)。
func appTx(t *testing.T, app *sql.DB, stmts []string) *sql.Tx {
	t.Helper()
	tx, err := app.Begin()
	if err != nil {
		t.Fatalf("開交易: %v", err)
	}
	execScope(t, tx, stmts)
	return tx
}

// appExecScoped 以 app_rw 開**獨立**交易、套用 scope 後執行單一語句並回傳其錯誤。
// 每個受測語句各自一個交易:一句失敗會 abort 整個交易,共用交易會讓後續語句只剩 25P02(T7 實測),
// 掩蓋真正的原因。
func appExecScoped(t *testing.T, app *sql.DB, stmts []string, query string, args ...any) error {
	t.Helper()
	tx, err := app.Begin()
	if err != nil {
		t.Fatalf("開交易: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	execScope(t, tx, stmts)
	_, err = tx.Exec(query, args...)
	return err
}

// execScope 逐句執行 SET LOCAL。
func execScope(t *testing.T, tx *sql.Tx, stmts []string) {
	t.Helper()
	for _, s := range stmts {
		if _, err := tx.Exec(s); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
	}
}

// hasAuditRow 以指定交易斷言看得到某列稽核(RLS 可見性,與 admin 的真值查詢分開)。
func hasAuditRow(tx *sql.Tx, id int) bool {
	var n int
	return tx.QueryRow(`SELECT count(*) FROM audit_logs WHERE id = $1`, id).Scan(&n) == nil && n == 1
}

// metadictIDs 取字典清單的 id 集合(斷言「包含」用;少了剛建立的列即為未走請求交易)。
func metadictIDs(items []*metadictv1.Metadict) map[string]bool {
	out := make(map[string]bool, len(items))
	for _, m := range items {
		out[m.GetId()] = true
	}
	return out
}

// newMetadictAuditAppRoleServer 以業務 client(app_rw + dbtenant.NewClient)掛載真字典/稽核 handler:
// 兩者的 RegisterXxxServices 內含 dbtenant.HandlerOption → 每個 RPC 都在請求交易內執行。
// withScope=false 模擬「漏掛 RLS scope」(負向對照);生產的 scope 由 authzMiddleware 依身分導出。
//
// **連線池上限 1**(呼叫端設定)是本測試的第二個斷言:請求交易佔住唯一連線,任何自開交易或
// 繞過請求交易的查詢都取不到連線 → 該 RPC 逾時失敗(見 callRPC)。這是「全部 DB 存取都走請求交易」
// 的直接證據 —— 只靠 RLS 可見性抓不到自開交易(wrapped client 對任何 Tx 都會套 ctx 的 SET LOCAL)。
func newMetadictAuditAppRoleServer(t *testing.T, client *ent.Client, id authz.Identity, scope auth.RLSScope, withScope bool) mdAuditClients {
	t.Helper()
	mux := http.NewServeMux()
	RegisterMetadictServices(mux, client)
	RegisterAuditServices(mux, client)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		if withScope {
			ctx = auth.WithRLS(ctx, scope)
		}
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return mdAuditClients{
		metadicts: metadictv1connect.NewMetadictServiceClient(http.DefaultClient, ts.URL),
		audits:    auditv1connect.NewAuditServiceClient(http.DefaultClient, ts.URL),
	}
}

// rpcTimeout 為單一 RPC 的上限:連線池上限 1 之下,卡在請求交易外的存取會以逾時收斂成可讀的紅,
// 而不是讓整個測試無聲掛住(逾時本身就是「未走請求交易」的證據)。
const rpcTimeout = 10 * time.Second

// callRPC 以單一 RPC 的 deadline 執行 fn;錯誤即測試失敗。
func callRPC[T any](t *testing.T, fn func(ctx context.Context) (T, error)) T {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), rpcTimeout)
	defer cancel()
	v, err := fn(ctx)
	if err != nil {
		t.Fatalf("RPC 失敗(若為 deadline 逾期,代表有存取沒走請求交易): %v", err)
	}
	return v
}

// callRPCErr 以單一 RPC 的 deadline 執行 fn,錯誤交由呼叫端斷言。
func callRPCErr(t *testing.T, fn func(ctx context.Context) error) error {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), rpcTimeout)
	defer cancel()
	return fn(ctx)
}

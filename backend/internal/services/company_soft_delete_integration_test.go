//go:build integration

// 公司軟刪除(P2-A)的真 PostgreSQL 整合測試:跑真正的 goose 遷移(含新增的 00019),
// 再以真實 service handler 與 ent client 觀察只有真 PG 才看得見的三件事 ——
//
//	① 00019 的**部分唯一索引**語意:未刪除列之間 identifier 不可重複、已刪除列則可共用
//	   (舊的表層 UNIQUE 不接受 WHERE 條件,軟刪除後識別碼將永遠無法重用);且舊 UNIQUE 已移除。
//	② DeleteCompany 改軟刪除後,audit_logs.company_id 的 FK(00010)不再阻擋刪除 —— 舊行為
//	   在有稽核列時硬刪除即 FK 違反,被舊映射誤報成 AlreadyExists 並把原始 SQL 送給前端。
//	③ 真約束衝突(users.email 的表層 UNIQUE)經 toConnectError 後不得外洩 DB 原文,
//	   也不得再誤報 AlreadyExists。
//
// 為何 sqlite(enttest)不足以守住:sqlite 不報約束名/SQLSTATE、亦不阻擋 FK 的方式與 PG 不同
// (硬刪除的 FK 違反在所有權狀態下無從重現),本測試因此必須真 PG。
// 遷移以 cmd/migrate 相同路徑套用(同 dialect、同目錄、同版本表),不另寫複製的 DDL。
package services

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx database/sql driver
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

const (
	// p2aMigrationsDir 相對套件目錄(go test 以套件目錄為 cwd),與 cmd/migrate 同路徑。
	p2aMigrationsDir = "../../database/migrations"
	// p2aGooseTable 為業務遷移版本表(cmd/migrate 未改動 goose 預設值)。
	p2aGooseTable = "goose_db_version"
)

// TestIntegrationCompanySoftDelete P2-A 回歸(見檔頭說明)。
func TestIntegrationCompanySoftDelete(t *testing.T) {
	// 斷言前提是「全新空庫」(索引/約束由本波遷移決定),覆寫模式下 skip。
	testsupport.RequiresContainer(t)
	ctx := t.Context()
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	sqlDB, db := openPGEntClientFromGoose(t, dsn)

	// 稽核列有 user_id/company_id 的 FK(00010),故操作者必須是真實存在的使用者 ——
	// 以 super 身分操作(sqlite 的 enttest 不建這組 FK,故只有真 PG 會踩到)。
	actorCo := db.Company.Create().SetName("操作者公司").SetIdentifier("P2A-ACTOR").SaveX(ctx)
	actor := db.User.Create().
		SetEmail("p2a-actor@example.com").
		SetName("操作者").
		SetStatus("active").
		SetRole("super").
		SetPasswordHash("x").
		SetCompanyID(actorCo.ID).
		SaveX(ctx)
	cc, uc := newCompanySoftDeleteServer(t, db, authz.Identity{
		UserID: strconv.Itoa(actor.ID), CompanyID: strconv.Itoa(actorCo.ID), Role: "super", Roles: []string{"super"},
	})

	t.Run("00019 部分唯一索引與舊表層 UNIQUE", func(t *testing.T) {
		var indexdef string
		if err := sqlDB.QueryRow(
			`SELECT indexdef FROM pg_indexes WHERE tablename = 'companies' AND indexname = 'companies_identifier_active_unique'`,
		).Scan(&indexdef); err != nil {
			t.Fatalf("查 companies_identifier_active_unique: %v(軟刪除後識別碼重用必須由部分唯一索引保證)", err)
		}
		for _, want := range []string{"UNIQUE", "identifier", "deleted_at IS NULL"} {
			if !strings.Contains(indexdef, want) {
				t.Fatalf("索引定義 %q 應含 %s", indexdef, want)
			}
		}

		// 表層 UNIQUE(00005 的 companies_identifier_key)必須已被移除,否則部分唯一索引形同虛設。
		var tableUnique int
		if err := sqlDB.QueryRow(
			`SELECT count(*) FROM pg_constraint WHERE conrelid = 'companies'::regclass AND contype = 'u'`,
		).Scan(&tableUnique); err != nil {
			t.Fatalf("查 companies 的表層 UNIQUE: %v", err)
		}
		if tableUnique != 0 {
			t.Fatalf("companies 仍有 %d 個表層 UNIQUE 約束(軟刪除後識別碼將無法重用)", tableUnique)
		}

		// ② 的前提:audit_logs.company_id 的 FK 真的存在。
		var fk int
		if err := sqlDB.QueryRow(
			`SELECT count(*) FROM pg_constraint WHERE conname = 'audit_logs_company_id_fkey'`,
		).Scan(&fk); err != nil {
			t.Fatalf("查 audit_logs_company_id_fkey: %v", err)
		}
		if fk != 1 {
			t.Fatalf("audit_logs.company_id 應有 FK(00010),got %d", fk)
		}
	})

	t.Run("識別碼唯一性語意", func(t *testing.T) {
		first := db.Company.Create().SetName("甲").SetIdentifier("P2A-IDX").SaveX(ctx)

		// 未刪除列之間不可重複。
		_, err := db.Company.Create().SetName("乙").SetIdentifier("P2A-IDX").Save(ctx)
		if !isUniqueViolation(err) {
			t.Fatalf("未刪除列之間識別碼重複必須被唯一索引擋下(23505),got %v", err)
		}

		// 軟刪除後可重用同一個識別碼。
		db.Company.UpdateOneID(first.ID).SetDeletedAt(time.Now().UTC()).SaveX(ctx)
		second := db.Company.Create().SetName("丙").SetIdentifier("P2A-IDX").SaveX(ctx)

		// 多筆已刪除列可共用同一識別碼(索引條件只涵蓋未刪除列)。
		db.Company.UpdateOneID(second.ID).SetDeletedAt(time.Now().UTC()).SaveX(ctx)
		db.Company.Create().SetName("丁").SetIdentifier("P2A-IDX").SaveX(ctx)
	})

	t.Run("刪除公司:FK 不再阻擋且稽核留痕", func(t *testing.T) {
		created, err := cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{
			Name: "刪除測試公司", Identifier: "P2A-DEL",
		}))
		if err != nil {
			t.Fatalf("CreateCompany: %v", err)
		}
		idStr := created.Msg.GetCompany().GetId()
		cid, err := strconv.Atoi(idStr)
		if err != nil {
			t.Fatalf("解析公司 id %q: %v", idStr, err)
		}

		// 先寫一筆稽核(UpdateCompany):舊行為(硬刪除)在此條件下必違反 audit_logs 的 FK。
		if _, err := cc.UpdateCompany(ctx, connect.NewRequest(&v1.UpdateCompanyRequest{
			CompanyId: idStr, Status: strPtr("inactive"),
		})); err != nil {
			t.Fatalf("前置 UpdateCompany: %v", err)
		}

		if _, err := cc.DeleteCompany(ctx, connect.NewRequest(&v1.DeleteCompanyRequest{CompanyId: idStr})); err != nil {
			t.Fatalf("刪除有稽核列的公司必須成功(舊行為會 FK 違反): %v", err)
		}

		// 列保留(軟刪除),稽核列仍指向它 —— FK 因此永遠成立。
		row, err := db.Company.Get(ctx, cid)
		if err != nil {
			t.Fatalf("軟刪除後列必須保留: %v", err)
		}
		if row.DeletedAt == nil {
			t.Fatal("刪除必須標記 deleted_at")
		}
		var audits int
		if err := sqlDB.QueryRow(`SELECT count(*) FROM audit_logs WHERE company_id = $1`, cid).Scan(&audits); err != nil {
			t.Fatalf("查稽核列: %v", err)
		}
		if audits < 2 { // update + delete
			t.Fatalf("刪除後應留有該公司的稽核列(update+delete),got %d", audits)
		}
		var deleteAudits int
		if err := sqlDB.QueryRow(
			`SELECT count(*) FROM audit_logs WHERE company_id = $1 AND action = 'delete' AND resource_type = 'company'`, cid,
		).Scan(&deleteAudits); err != nil {
			t.Fatalf("查 delete 稽核列: %v", err)
		}
		if deleteAudits != 1 {
			t.Fatalf("刪除應寫 1 筆 action=delete/resource_type=company 稽核,got %d", deleteAudits)
		}

		// 清單與詳情皆看不到已刪除公司。
		list, err := cc.ListCompanies(ctx, connect.NewRequest(&v1.ListCompaniesRequest{Keyword: "P2A-DEL"}))
		if err != nil {
			t.Fatalf("ListCompanies: %v", err)
		}
		if list.Msg.GetPagination().GetTotal() != 0 || len(list.Msg.GetCompanies()) != 0 {
			t.Fatalf("已刪除公司不得出現在清單,got total=%d items=%d", list.Msg.GetPagination().GetTotal(), len(list.Msg.GetCompanies()))
		}
		if _, err := cc.GetCompany(ctx, connect.NewRequest(&v1.GetCompanyRequest{CompanyId: idStr})); connect.CodeOf(err) != connect.CodeNotFound {
			t.Fatalf("已刪除公司 GetCompany 應 NotFound,got %v", err)
		}
		if _, err := cc.DeleteCompany(ctx, connect.NewRequest(&v1.DeleteCompanyRequest{CompanyId: idStr})); connect.CodeOf(err) != connect.CodeNotFound {
			t.Fatalf("重複刪除應 NotFound,got %v", err)
		}

		// 識別碼釋出後可由新公司重用(部分唯一索引 + 服務層前置查詢皆以未刪除列為準)。
		if _, err := cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{
			Name: "重用識別碼", Identifier: "P2A-DEL",
		})); err != nil {
			t.Fatalf("軟刪除後同識別碼應可再建公司: %v", err)
		}
	})

	t.Run("約束衝突不外洩 SQL", func(t *testing.T) {
		// 真識別碼重複:服務層前置判別 → AlreadyExists,訊息不含 DB 原文。
		if _, err := cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{
			Name: "唯一者", Identifier: "P2A-UNIQ",
		})); err != nil {
			t.Fatalf("前置 CreateCompany: %v", err)
		}
		_, err := cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{
			Name: "重複者", Identifier: "P2A-UNIQ",
		}))
		if connect.CodeOf(err) != connect.CodeAlreadyExists {
			t.Fatalf("識別碼重複應回 AlreadyExists,got %v", err)
		}
		assertNoDBInternals(t, err)

		// 真 DB 約束衝突(users.email 表層 UNIQUE;服務層無前置查詢)→ 由 toConnectError 映射。
		co, err := cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{
			Name: "帳號公司", Identifier: "P2A-USER",
		}))
		if err != nil {
			t.Fatalf("前置 CreateCompany(帳號公司): %v", err)
		}
		userReq := func() *v1.CreateUserRequest {
			return &v1.CreateUserRequest{
				Name:      "重複帳號",
				Email:     "p2a-dup@example.com",
				CompanyId: co.Msg.GetCompany().GetId(),
				Role:      "staff",
			}
		}
		if _, err := uc.CreateUser(ctx, connect.NewRequest(userReq())); err != nil {
			t.Fatalf("前置 CreateUser: %v", err)
		}
		_, err = uc.CreateUser(ctx, connect.NewRequest(userReq()))
		if connect.CodeOf(err) != connect.CodeFailedPrecondition {
			t.Fatalf("約束衝突應映射為 FailedPrecondition(舊映射誤報 AlreadyExists),got %v", err)
		}
		assertNoDBInternals(t, err)
	})
	t.Run("00019 Down 對稱還原", func(t *testing.T) {
		// 專屬容器:Down 會把表層 UNIQUE 加回來,而前面子測試已刻意留下「同識別碼的已刪除列」,
		// 那些列會讓 Down 以 23505 擋下(正確訊號,見 migration 檔頭)。此處驗的是全新庫的對稱性。
		downDSN := testsupport.Postgres(t)
		migrateBusinessUp(t, downDSN)
		downDB := openRawDB(t, downDSN)
		migrateBusinessDownTo(t, downDSN, "18")

		if indexExists(t, downDB, "companies_identifier_active_unique") {
			t.Fatal("Down 必須移除部分唯一索引")
		}
		if columnExists(t, downDB, "companies", "deleted_at") {
			t.Fatal("Down 必須移除 deleted_at 欄位")
		}
		var tableUnique int
		if err := downDB.QueryRow(
			`SELECT count(*) FROM pg_constraint WHERE conrelid = 'companies'::regclass AND contype = 'u'`,
		).Scan(&tableUnique); err != nil {
			t.Fatalf("查 Down 後的表層 UNIQUE: %v", err)
		}
		if tableUnique != 1 {
			t.Fatalf("Down 必須還原表層 UNIQUE(companies_identifier_key),got %d", tableUnique)
		}

		// 再 Up 一次:回到本波語意(索引在、表層 UNIQUE 不在)—— Up/Down 皆可重複套用。
		migrateBusinessUp(t, downDSN)
		var indexdef string
		if err := downDB.QueryRow(
			`SELECT indexdef FROM pg_indexes WHERE tablename = 'companies' AND indexname = 'companies_identifier_active_unique'`,
		).Scan(&indexdef); err != nil {
			t.Fatalf("重跑 Up 後部分唯一索引應復原: %v", err)
		}
		if err := downDB.QueryRow(
			`SELECT count(*) FROM pg_constraint WHERE conrelid = 'companies'::regclass AND contype = 'u'`,
		).Scan(&tableUnique); err != nil {
			t.Fatalf("查重跑 Up 後的表層 UNIQUE: %v", err)
		}
		if tableUnique != 0 {
			t.Fatalf("重跑 Up 後表層 UNIQUE 應再次被移除,got %d", tableUnique)
		}
	})
}

// indexExists 判斷指定索引是否存在。
func indexExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM pg_indexes WHERE indexname = $1`, name).Scan(&n); err != nil {
		t.Fatalf("查索引 %s: %v", name, err)
	}
	return n > 0
}

// columnExists 判斷指定欄位是否存在。
func columnExists(t *testing.T, db *sql.DB, table, column string) bool {
	t.Helper()
	var n int
	if err := db.QueryRow(
		`SELECT count(*) FROM information_schema.columns WHERE table_name = $1 AND column_name = $2`, table, column,
	).Scan(&n); err != nil {
		t.Fatalf("查欄位 %s.%s: %v", table, column, err)
	}
	return n > 0
}

// openRawDB 開一條 database/sql 連線(供 DDL 層級的斷言使用)。
func openRawDB(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open(pgx): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// assertNoDBInternals 斷言錯誤訊息不含驅動層原文(SQLSTATE、約束名、欄位名)。
func assertNoDBInternals(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("預期錯誤,got nil")
	}
	for _, leak := range []string{"SQLSTATE", "23505", "23503", "duplicate key", "constraint", "UNIQUE", "pgx", "users_email"} {
		if strings.Contains(err.Error(), leak) {
			t.Fatalf("錯誤訊息外洩 DB 原文(%q):%v", leak, err)
		}
	}
}

// isUniqueViolation 判斷錯誤鏈中是否有 PostgreSQL 的 23505(unique_violation)。
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// migrateBusinessUp 以 cmd/migrate 相同路徑套用業務遷移(同 dialect、同目錄、同版本表)。
func migrateBusinessUp(t *testing.T, dsn string) {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open(pgx): %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("設定 dialect: %v", err)
	}
	goose.SetTableName(p2aGooseTable)
	goose.SetBaseFS(nil)
	if err := goose.RunContext(t.Context(), "up", db, p2aMigrationsDir); err != nil {
		t.Fatalf("goose up: %v(00019 必須能套用於全新空庫)", err)
	}
}

// migrateBusinessDownTo 以 cmd/migrate 相同路徑回退至指定版本(`goose down-to` 語意:**該版本
// 保留**、其後的遷移全部還原),以版本號釘住要驗的遷移 —— 用「回退最後一筆」會被後續新遷移搶走
// 目標(00020 落地後 `goose down` 只還原 00020,00019 的 Down 便驗不到)。
// 因此要驗 00019 的 Down 就回退至 "18",要驗 00020 的 Down 就回退至 "19"。
func migrateBusinessDownTo(t *testing.T, dsn, version string) {
	t.Helper()
	db := openRawDB(t, dsn)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("設定 dialect: %v", err)
	}
	goose.SetTableName(p2aGooseTable)
	goose.SetBaseFS(nil)
	if err := goose.RunContext(t.Context(), "down-to", db, p2aMigrationsDir, version); err != nil {
		t.Fatalf("goose down-to %s: %v(該遷移的 Down 必須能對稱還原)", version, err)
	}
}

// openPGEntClientFromGoose 以 pgx 連線建立 ent client,但**不**呼叫 ent 的 Schema.Create:
// 本測試要驗的正是 goose 遷移(00019)在真 PG 上建出的索引,不能被 ent 自建的索引掩蓋。
func openPGEntClientFromGoose(t *testing.T, dsn string) (*sql.DB, *ent.Client) {
	t.Helper()
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open(pgx): %v", err)
	}
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, sqlDB)))
	t.Cleanup(func() { _ = client.Close() })
	return sqlDB, client
}

// newCompanySoftDeleteServer 以指定身分掛載公司 + 使用者 handler(與 sqlite 版同構,只換 DB)。
func newCompanySoftDeleteServer(t *testing.T, db *ent.Client, id authz.Identity) (salesorderv1connect.CompanyServiceClient, salesorderv1connect.UserServiceClient) {
	t.Helper()
	mux := http.NewServeMux()
	RegisterCompanyServices(mux, db)
	RegisterUserServices(mux, db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithCASLEnabled(ctx, true)
		ctx = authz.WithDB(ctx, db)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	return salesorderv1connect.NewCompanyServiceClient(http.DefaultClient, ts.URL),
		salesorderv1connect.NewUserServiceClient(http.DefaultClient, ts.URL)
}

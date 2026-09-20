//go:build integration

// customer_counters × ent schema 的真 PostgreSQL 整合測試(E3)。
//
// 缺陷(00013 與 ent 的落差):00013 建出的 customer_counters 以 company_id 為主鍵、沒有 id 欄位,
// 而 ent schema 把 company_id 宣告為一般 Unique 欄位(field.Int("company_id").Unique())
// → ent 的隱含 id 才是主鍵(ent/migrate/schema.go 的 CustomerCountersColumns[0] 為 id 且為
// PrimaryKey)。在 **goose 建出的真 PG** 上,ent 對該表的所有取用(Create/Exist/Only/Update)一律
// `ERROR: column "id" does not exist (SQLSTATE 42703)`;而 CreateCustomer 的必要步驟
// ensureCustomerCounter(Exist/Create)與 nextCustomerCode(Only/Update)都走 ent
// → **任何走到取號的 CreateCustomer 在現行 PG 部署必定失敗**。
//
// 為何 sqlite(enttest)踩不到:enttest 由 ent 依 schema 自建含 id 的表,不是 goose 的 00013,
// 故只有在真 PG + 真 Goose 遷移下才觀測得到;本檔因此以 build tag `integration` 隔離。
//
// 測試三個面向:
//  1. 結構對齊:00021 後 customer_counters 必須有 NOT NULL bigserial id 主鍵,且 company_id 仍唯一
//     (唯一性以行為斷言:重複寫入必 23505,不看約束名字)。
//  2. 取號路徑可用:ent 的 Create/Exist/Only/Update 在 goose 建出的庫上必須可用。
//  3. 真 handler 端到端:dept_admin 走完整 CreateCustomer 兩次 → 取號 TY000001/TY000002、counter 推進。
//  4. Down 對稱 + 再次 Up 的冪等:down-to 20 還原為 company_id 主鍵;帶資料再次 Up 時既有列必須
//     補到不同的 id(ADD COLUMN id bigserial 的 backfill),之後 ent 仍可用。
package services

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/customercounter"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1/customersv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationCustomerCountersAlignEntSchema 00021:goose 建出的 customer_counters 必須符合 ent
// 期待(id 主鍵 + company_id 唯一),且 ent 的 Create/Exist/Only/Update 皆可用。
func TestIntegrationCustomerCountersAlignEntSchema(t *testing.T) {
	testsupport.RequiresContainer(t)
	ctx := t.Context()
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	sqlDB, db := openPGEntClientFromGoose(t, dsn)

	// 結構:id 必須存在、NOT NULL、由 sequence 產生(bigserial)。
	var dataType, isNullable, colDefault string
	if err := sqlDB.QueryRowContext(ctx, `SELECT data_type, is_nullable, coalesce(column_default, '')
		FROM information_schema.columns
		WHERE table_name = 'customer_counters' AND column_name = 'id'`).Scan(&dataType, &isNullable, &colDefault); err != nil {
		t.Fatalf(`讀取 customer_counters.id 定義失敗(%v):00013 的表沒有 id,ent 期待的 id 主鍵必須由後續遷移補上`, err)
	}
	if dataType != "bigint" || isNullable != "NO" || !strings.Contains(colDefault, "nextval(") {
		t.Fatalf("customer_counters.id 應為 NOT NULL bigint 且預設 nextval(sequence),得到 type=%q nullable=%q default=%q",
			dataType, isNullable, colDefault)
	}
	if got := customerCounterPKColumns(t, sqlDB); !slices.Equal(got, []string{"id"}) {
		t.Fatalf("customer_counters 主鍵應為 [id](ent 的隱含主鍵),得到 %v", got)
	}

	// ent 取號路徑用到的四種取用方式(修復前全部因 column "id" does not exist 失敗)。
	// customer_counters 已 ENABLE(+FORCE):fixture 寫入走系統範圍入口(見 seedTx)。
	co := db.Company.Create().SetName("E3 counter 公司").SetIdentifier("E3-COUNTER").SaveX(ctx)
	var created *ent.CustomerCounter
	seedTx(t, db, func(tx *ent.Tx) error {
		var err error
		created, err = tx.CustomerCounter.Create().SetCompanyID(co.ID).SetNextSeq(1).SetVersion(0).Save(ctx)
		return err
	})
	if created.ID <= 0 {
		t.Fatalf("ent Create 應由 sequence 取得 id,得到 %d", created.ID)
	}
	exists, err := db.CustomerCounter.Query().Where(customercounter.CompanyIDEQ(co.ID)).Exist(ctx)
	if err != nil {
		t.Fatalf("ent Exist 失敗: %v", err)
	}
	if !exists {
		t.Fatal("剛建立的 counter 列必須 Exist == true")
	}
	got, err := db.CustomerCounter.Query().Where(customercounter.CompanyIDEQ(co.ID)).Only(ctx)
	if err != nil {
		t.Fatalf("ent Only 失敗: %v", err)
	}
	if got.ID != created.ID || got.NextSeq != 1 {
		t.Fatalf("ent Only 讀回 %+v,應為 id=%d next_seq=1", got, created.ID)
	}
	// nextCustomerCode 的樂觀鎖更新語意:版本相符才推進。
	n, err := db.CustomerCounter.Update().
		Where(customercounter.CompanyIDEQ(co.ID), customercounter.VersionEQ(got.Version)).
		SetNextSeq(got.NextSeq + 1).SetVersion(got.Version + 1).
		Save(ctx)
	if err != nil {
		t.Fatalf("ent Update 失敗: %v", err)
	}
	if n != 1 {
		t.Fatalf("版本相符時樂觀鎖更新應影響 1 列,得到 %d", n)
	}

	// company_id 的唯一性必須保留(PK 讓位給 id 後改以 UNIQUE 表達;行為斷言)。
	if _, err := sqlDB.ExecContext(ctx, `INSERT INTO customer_counters (company_id) VALUES ($1)`, co.ID); !isUniqueViolation(err) {
		t.Fatalf("同一 company_id 重複寫入應 23505(每公司一列),得到 %v", err)
	}
}

// TestIntegrationCreateCustomerIssuesCodeOnPG 真 handler 端到端:真 PG + 真 Goose 遷移下,
// dept_admin 兩次 CreateCustomer 都必須成功取號並推進 counter(修復前第一次即 42703)。
func TestIntegrationCreateCustomerIssuesCodeOnPG(t *testing.T) {
	testsupport.RequiresContainer(t)
	ctx := t.Context()
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)

	co := db.Company.Create().SetName("E3 取號公司").SetIdentifier("E3-CODE").
		SetStatus("active").SetCustomerCodePrefix("TY").SaveX(ctx)
	dep := db.Department.Create().SetName("門市取號").SetCompanyID(co.ID).SaveX(ctx)
	// 稽核列有 user_id FK(00010),操作者必須是真實使用者。
	actor := db.User.Create().SetEmail("e3-actor@example.com").SetName("操作者").SetStatus("active").
		SetRole("dept_admin").SetPasswordHash("x").SetCompanyID(co.ID).SetDepartmentID(dep.ID).SaveX(ctx)
	rep := db.User.Create().SetEmail("e3-rep@example.com").SetName("業務").SetStatus("active").
		SetRole("staff").SetPasswordHash("x").SetCompanyID(co.ID).SetDepartmentID(dep.ID).SaveX(ctx)

	client := newCustomerPGServer(t, db, authz.Identity{
		UserID: strconv.Itoa(actor.ID), CompanyID: strconv.Itoa(co.ID), DepartmentID: strconv.Itoa(dep.ID),
		Role: "dept_admin", Roles: []string{"dept_admin", "staff"},
	})

	for i, want := range []string{"TY000001", "TY000002"} {
		resp, err := client.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{
			Name: "客戶" + strconv.Itoa(i+1), DefaultSalesRepId: strconv.Itoa(rep.ID),
		}))
		if err != nil {
			t.Fatalf("真 PG 上第 %d 次 CreateCustomer 必須成功(取號走 customer_counters),得到 %v", i+1, err)
		}
		if code := resp.Msg.GetCustomer().GetCustomerCode(); code != want {
			t.Fatalf("第 %d 次建檔取號應為 %q,得到 %q", i+1, want, code)
		}
	}

	cnt, err := db.CustomerCounter.Query().Where(customercounter.CompanyIDEQ(co.ID)).Only(ctx)
	if err != nil {
		t.Fatalf("讀取 counter: %v", err)
	}
	if cnt.NextSeq != 3 || cnt.Version != 2 {
		t.Fatalf("counter 應推進至 next_seq=3 version=2,得到 next_seq=%d version=%d", cnt.NextSeq, cnt.Version)
	}
}

// TestIntegrationCustomerCountersMigrationDown 00021 的 Down 必須對稱還原為 company_id 主鍵,
// 且再次 Up 對「已有 counter 列的既有庫」可重複套用(既有列補得不同 id,ent 仍可用)。
func TestIntegrationCustomerCountersMigrationDown(t *testing.T) {
	testsupport.RequiresContainer(t)
	ctx := t.Context()
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)

	// 既有資料:兩家公司的 counter(模擬已跑過 00013/00021 的線上庫)。
	coA := insertCounterRow(t, dsn, 1)
	coB := insertCounterRow(t, dsn, 7)

	migrateBusinessDownTo(t, dsn, "20")
	downDB := openRawDB(t, dsn)
	if columnExists(t, downDB, "customer_counters", "id") {
		t.Fatal("Down 必須移除 customer_counters.id")
	}
	if got := customerCounterPKColumns(t, downDB); !slices.Equal(got, []string{"company_id"}) {
		t.Fatalf("Down 後主鍵應還原為 [company_id](00013 的原貌),得到 %v", got)
	}

	// 再次 Up:既有列必須被 backfill 成不同的 id,且 ent 對該表恢復可用。
	migrateBusinessUp(t, dsn)
	sqlDB, db := openPGEntClientFromGoose(t, dsn)
	var nullIDs, distinctIDs int
	if err := sqlDB.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE id IS NULL), count(DISTINCT id) FROM customer_counters`).Scan(&nullIDs, &distinctIDs); err != nil {
		t.Fatalf("檢查 backfill: %v", err)
	}
	if nullIDs != 0 || distinctIDs != 2 {
		t.Fatalf("再次 Up 後既有 2 列必須各得一個 id(無 NULL),得到 null=%d distinct=%d", nullIDs, distinctIDs)
	}
	for _, coID := range []int{coA, coB} {
		if _, err := db.CustomerCounter.Query().Where(customercounter.CompanyIDEQ(coID)).Only(ctx); err != nil {
			t.Fatalf("再次 Up 後 ent Only(company_id=%d)必須可用: %v", coID, err)
		}
	}
}

// newCustomerPGServer 以指定身分掛載 CustomerService handler(與 sqlite 版同構,只換 DB)。
func newCustomerPGServer(t *testing.T, db *ent.Client, id authz.Identity) customersv1connect.CustomerServiceClient {
	t.Helper()
	mux := http.NewServeMux()
	RegisterCustomerServices(mux, db, "http://localhost:3000")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithDB(ctx, db)
		ctx = audit.WithMeta(ctx, audit.Meta{IP: "1.2.3.4", UserAgent: "t"})
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return customersv1connect.NewCustomerServiceClient(http.DefaultClient, ts.URL)
}

// customerCounterPKColumns 回傳 customer_counters 主鍵欄位(依名稱排序)。
func customerCounterPKColumns(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.Query(`SELECT a.attname FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY (i.indkey)
		WHERE i.indrelid = 'customer_counters'::regclass AND i.indisprimary
		ORDER BY a.attname`)
	if err != nil {
		t.Fatalf("查詢 customer_counters 主鍵: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var cols []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("讀取主鍵欄位: %v", err)
		}
		cols = append(cols, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("走訪主鍵欄位: %v", err)
	}
	return cols
}

// insertCounterRow 以原生 SQL 直寫一列 counter(供 Down/再次 Up 的既有資料情境),
// 略過 ent 以免受本票要修的落差影響。回傳 company id。
func insertCounterRow(t *testing.T, dsn string, nextSeq int) int {
	t.Helper()
	db := openRawDB(t, dsn)
	var coID int
	if err := db.QueryRow(`INSERT INTO companies (name, identifier) VALUES ($1, $1) RETURNING id`,
		fmt.Sprintf("E3 既有公司 %d", nextSeq)).Scan(&coID); err != nil {
		t.Fatalf("建立既有公司: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO customer_counters (company_id, next_seq, version) VALUES ($1, $2, 0)`,
		coID, nextSeq); err != nil {
		t.Fatalf("直寫 counter 列: %v", err)
	}
	return coID
}

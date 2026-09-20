//go:build integration

// 客戶帳號建檔 × 部門刪除的真 PostgreSQL 併發整合測試(E1;D1 報告 §5.1 留下的未關路徑)。
//
// 缺陷:CreateCustomer 會把客戶列與主/業務子帳號(users 角色 customer,建檔連動 D22)一起掛在
// **身分導出的部門**(deptScope → did)。該路徑原本只在 autocommit 下讀部門,而 users.department_id
// 的 FK 插入只取 FOR KEY SHARE、與 DeleteDepartment 的條件式 UPDATE(FOR NO KEY UPDATE)**不衝突**
// → 「建檔驗證通過 → 部門被刪除並提交 → 建檔交易才提交」的交錯會留下活帳號落在已軟刪部門
// (與 D1 已修的三條路徑(CreateUser / UpdateUser / AssignRole)同型,但沒有同一套鎖協定)。
//
// A 側以**生產函式**重現建檔交易內的兩步(與 department_delete_race_integration_test 同一手法):
// validateDepartmentInCompany(E1 後由 CreateCustomer 在自己的交易內呼叫的同一支;含 FOR SHARE)
// → 建客戶列 + buildCustomerAccount(即本票所稱的 createCustomerUser),尚未提交。
// B 側是**真 handler** 的 DeleteDepartment。斷言:
//
//	① A 確實持有 departments 的 FOR SHARE:另一條連線以 FOR NO KEY UPDATE NOWAIT 取同一列必須
//	   失敗(55P03)。刻意不用 FOR UPDATE —— A 的 FK 插入會取得 FOR KEY SHARE,而 FOR UPDATE 與
//	   FOR KEY SHARE 也衝突(會被自己的寫入誤導);FOR NO KEY UPDATE 與 FOR KEY SHARE 相容,
//	   只有 FOR SHARE 會讓它取不到鎖。
//	② 期間 DeleteDepartment 必須真的停在該列鎖上(pg_stat_activity):沒有這步,「尚未返回」可能
//	   只是刪除還沒跑到。
//	③ A 提交後刪除必須回 FailedPrecondition(部門已有成員 = 剛落地的客戶帳號)、部門存活、
//	   帳號仍掛在該部門(活帳號不得落在已軟刪部門)。
//	④ 對照組:帳號調離後同一部門可正常刪除(證明 ③ 來自「仍有成員」,不是鎖把刪除卡死)。
//
// 為何 sqlite(enttest)不足以守住:SQLite 沒有 SELECT ... FOR SHARE/FOR UPDATE(ent 於該 dialect
// 直接讓查詢報錯),列鎖只存在於真 PG。
//
// 為何 A 不直接呼叫 CreateCustomer(真 handler):goose 遷移建出的 customer_counters(00013)只有
// company_id 主鍵,而 ent 的 schema 期待 id 欄位(ent/migrate/schema.go 的 CustomerCountersColumns
// 有 id)—— 在 goose 建出的真 PG 上,CustomerCounters 的 Create/Only 一律失敗
// (`column "id" does not exist`,實測見 e1-report.md),任何走到取號的 CreateCustomer 請求都到不了
// 部門驗證之後的步驟。該落差是本票之外的既有缺陷(修它要動 migration,不在本波範圍),故 A 以
// 交易內的生產函式重現,而「CreateCustomer 真的在自己的交易內呼叫驗證」由預設套件的
// TestCreateCustomerRejectsSoftDeletedDepartment 守住(拿掉呼叫即 RED)。
package services

import (
	"strconv"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent/department"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationCustomerAccountDepartmentDeleteRace E1:建檔交易持有部門列鎖期間,部門刪除必須被
// 擋住;提交後刪除回 FailedPrecondition、部門存活、客戶帳號仍掛在該部門。
func TestIntegrationCustomerAccountDepartmentDeleteRace(t *testing.T) {
	testsupport.RequiresContainer(t)
	ctx := t.Context()
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	sqlDB, db := openPGEntClientFromGoose(t, dsn)

	// 稽核列有 company_id/user_id 的 FK(00010),操作者必須是真實存在的使用者。
	actorCo := db.Company.Create().SetName("E1 操作者公司").SetIdentifier("E1-RACE-ACTOR").SaveX(ctx)
	actor := db.User.Create().
		SetEmail("e1-race-actor@example.com").
		SetName("操作者").
		SetStatus("active").
		SetRole("super").
		SetPasswordHash("x").
		SetCompanyID(actorCo.ID).
		SaveX(ctx)
	_, dc, _ := newDepartmentSoftDeleteServer(t, db, authz.Identity{
		UserID: strconv.Itoa(actor.ID), CompanyID: strconv.Itoa(actorCo.ID), Role: "super", Roles: []string{"super"},
	})

	co := db.Company.Create().SetName("E1 競態公司").SetIdentifier("E1-RACE").SaveX(ctx)
	// 部門當下**無成員** —— 這是競態可達的前提:部門刪除的唯一前置條件是「該部門沒有使用者」,
	// 而建檔請求帶的是較早的身分快照(did),兩者不必一致。
	dep := db.Department.Create().SetName("門市建檔競態").SetCompanyID(co.ID).SaveX(ctx)

	// A:建檔交易內的兩步(生產函式)——驗證部門(FOR SHARE)→ 建客戶列 + 客戶帳號,尚未提交。
	// 註:此處的 customers 寫入刻意**不**走 seedTx 的系統範圍輔助:本測試要驗的正是「這個交易
	// 在提交前一直持有部門列鎖」,把它交給會自行 commit 的輔助會直接毀掉競態本身。
	tx, err := db.Tx(ctx)
	if err != nil {
		t.Fatalf("開啟建檔交易: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := validateDepartmentInCompany(ctx, tx, dep.ID, co.ID); err != nil {
		t.Fatalf("建檔路徑驗證部門(應於本交易內取 FOR SHARE 列鎖): %v", err)
	}
	cust := tx.Customer.Create().
		SetCompanyID(co.ID).SetCustomerCode("E1-000001").SetName("競態客戶").SetDepartmentID(dep.ID).
		SaveX(ctx)
	acct, err := buildCustomerAccount(ctx, tx.Client(), accountSpec{
		CompanyID: co.ID, DepartmentID: &dep.ID, CustomerID: cust.ID,
		Email: "e1-race-account@system.local", Name: "競態客戶", AccountName: "競態客戶",
		IsPrimary: true, PasswordHash: "x", MustChange: true, TempExpiresAt: time.Now().UTC().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("建客戶帳號: %v", err)
	}

	// B:同時刪除同一部門(真 handler、另一條連線)。
	done := make(chan error, 1)
	go func() {
		_, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{
			DepartmentId: strconv.Itoa(dep.ID),
		}))
		done <- err
	}()

	// ① 直接量測 A 的部門列鎖(見檔頭說明為何用 FOR NO KEY UPDATE)。
	var probe int
	err = sqlDB.QueryRowContext(ctx, `SELECT 1 FROM departments WHERE id = $1 FOR NO KEY UPDATE NOWAIT`, dep.ID).Scan(&probe)
	switch {
	case err == nil:
		t.Fatal("建檔交易的部門驗證未取得 FOR SHARE 列鎖:刪除端可在驗證之後提交,產生活帳號落在已軟刪部門")
	case !isRowLockNotAvailable(err):
		t.Fatalf("列鎖探測應因取不到鎖而失敗(55P03),got %v", err)
	}

	// ② 再確認「刪除真的被擋在該列鎖上」:pg_stat_activity 必須看到 B 的敘述停在 Lock 等待。
	waitUntilDeleteBlocked(t, sqlDB, done)

	// ③ 提交建檔 → 刪除重新評估 NOT EXISTS(users) → 看到剛落地的客戶帳號 → FailedPrecondition。
	if err := tx.Commit(); err != nil {
		t.Fatalf("提交建檔交易: %v", err)
	}
	select {
	case err := <-done:
		if connect.CodeOf(err) != connect.CodeFailedPrecondition {
			t.Fatalf("建檔提交後刪除應 FailedPrecondition(部門已有成員),got %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("建檔提交後刪除仍未返回:列鎖未被釋放")
	}

	// 部門存活、客戶帳號仍掛在該部門(活帳號不得落在已軟刪部門)。
	got, err := db.Department.Query().Where(department.ID(dep.ID)).Only(ctx)
	if err != nil {
		t.Fatalf("重新讀取部門: %v", err)
	}
	if got.DeletedAt != nil {
		t.Fatal("部門不得被軟刪除:刪除必須因「仍有成員」失敗")
	}
	attached, err := db.User.Query().
		Where(user.ID(acct.ID), user.HasDepartmentWith(department.ID(dep.ID))).
		Exist(ctx)
	if err != nil {
		t.Fatalf("重新讀取客戶帳號: %v", err)
	}
	if !attached {
		t.Fatalf("客戶帳號 %d 必須仍掛在部門 %d(活帳號不得落在已軟刪部門)", acct.ID, dep.ID)
	}

	// ④ 對照組:帳號調離後同一部門可正常刪除(證明 ③ 來自「仍有成員」,不是列鎖把刪除永久卡住)。
	if _, err := db.User.UpdateOneID(acct.ID).ClearDepartment().Save(ctx); err != nil {
		t.Fatalf("把客戶帳號調離部門: %v", err)
	}
	if _, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{
		DepartmentId: strconv.Itoa(dep.ID),
	})); err != nil {
		t.Fatalf("帳號調離後刪除必須成功: %v", err)
	}
}

//go:build integration

// 部門刪除 × 「把使用者掛進該部門」的真 PostgreSQL 併發整合測試(D1)。
//
// 缺陷:users.department_id 的 FK 插入只取 FOR KEY SHARE,與 DeleteDepartment 的條件式 UPDATE
// (FOR NO KEY UPDATE)**不衝突**。掛載路徑若只在 autocommit 下讀部門(舊行為),「掛載驗證通過 →
// 刪除提交 → 掛載才提交」的交錯會留下活帳號落在已軟刪部門。
//
// D1 的修法是**兩側對同一列取互斥鎖**:
//   - 掛載路徑(validateDepartmentInCompany,在掛載自己的交易內)取 FOR SHARE;
//   - 刪除路徑先對同列取 FOR UPDATE(lockDepartmentForDelete),之後才讀成員快照與寫入。
//
// 只有掛載端取 FOR SHARE 是**不夠**的(本測試的消去實驗實測):刪除的 UPDATE 雖然會被擋住,但
// READ COMMITTED 下它重評條件時仍用該敘述開始的舊快照,看不到等待期間才提交的新成員 → 仍刪成功。
// 把等待移到刪除端的鎖敘述後,UPDATE 才以新快照重評 NOT EXISTS(users) → 看到新成員 →
// FailedPrecondition。反向交錯則由掛載端的 FOR SHARE 等到刪除提交後重讀,看到 deleted_at →
// InvalidArgument(fail-closed;序列版本由 department_soft_delete_integration_test 覆蓋)。
//
// 為何 sqlite(enttest)不足以守住:SQLite 沒有 SELECT ... FOR SHARE/FOR UPDATE(ent 於該 dialect
// 直接讓查詢報錯),列鎖只存在於真 PG,故此驗證必須在 //go:build integration。
//
// A 側走的是**生產函式** (UserService).validateDepartmentInCompany,只是由測試自己開交易並
// 代 CreateUser/UpdateUser/AssignRole 插入成員(服務方法會自行提交,無法在中途停住)。因此
// 消去實驗(拿掉兩側任一列鎖)都會讓本測試 RED,見 d1-report.md。
package services

import (
	"database/sql"
	"errors"
	"strconv"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/salesorder/sales-order-1.0/backend/ent/department"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationDepartmentDeleteRace D1:掛載交易持有部門列鎖期間,刪除必須被擋住,且提交後
// 刪除回 FailedPrecondition、部門存活、成員有效。
func TestIntegrationDepartmentDeleteRace(t *testing.T) {
	testsupport.RequiresContainer(t)
	ctx := t.Context()
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	sqlDB, db := openPGEntClientFromGoose(t, dsn)

	// 稽核列有 company_id/user_id 的 FK(00010),故操作者必須是真實存在的使用者。
	actorCo := db.Company.Create().SetName("競態操作者公司").SetIdentifier("D1-RACE-ACTOR").SaveX(ctx)
	actor := db.User.Create().
		SetEmail("d1-race-actor@example.com").
		SetName("操作者").
		SetStatus("active").
		SetRole("super").
		SetPasswordHash("x").
		SetCompanyID(actorCo.ID).
		SaveX(ctx)
	_, dc, _ := newDepartmentSoftDeleteServer(t, db, authz.Identity{
		UserID: strconv.Itoa(actor.ID), CompanyID: strconv.Itoa(actorCo.ID), Role: "super", Roles: []string{"super"},
	})

	co := db.Company.Create().SetName("競態公司").SetIdentifier("D1-RACE").SaveX(ctx)
	dep := db.Department.Create().SetName("門市競態").SetCompanyID(co.ID).SaveX(ctx)

	// A:掛載路徑 —— 同一交易內讀部門(FOR SHARE)→ 建使用者,尚未提交。
	tx, err := db.Tx(ctx)
	if err != nil {
		t.Fatalf("開啟掛載交易: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := NewUserService(db).validateDepartmentInCompany(ctx, tx, dep.ID, co.ID); err != nil {
		t.Fatalf("掛載路徑驗證部門(應於本交易內取 FOR SHARE 列鎖): %v", err)
	}
	member := tx.User.Create().
		SetEmail("d1-race-member@example.com").
		SetName("併發成員").
		SetStatus("active").
		SetRole("staff").
		SetPasswordHash("x").
		SetCompanyID(co.ID).
		SetDepartmentID(dep.ID).
		SaveX(ctx)

	// B:同時刪除同一部門(真 handler、另一條連線)。
	done := make(chan error, 1)
	go func() {
		_, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{
			DepartmentId: strconv.Itoa(dep.ID),
		}))
		done <- err
	}()

	// ① 直接量測 A 的鎖:另一條連線以 FOR NO KEY UPDATE NOWAIT(與 DeleteDepartment 的條件式
	//    UPDATE 同級)取同一列必須失敗(55P03)。刻意不用 FOR UPDATE —— A 的成員插入使 FK 對該列
	//    取得 FOR KEY SHARE,而 FOR UPDATE 與 FOR KEY SHARE 也衝突,會被自己的插入誤導;
	//    FOR NO KEY UPDATE 與 FOR KEY SHARE **相容**,只有掛載路徑的 FOR SHARE 會讓它取不到鎖。
	var probe int
	err = sqlDB.QueryRowContext(ctx, `SELECT 1 FROM departments WHERE id = $1 FOR NO KEY UPDATE NOWAIT`, dep.ID).Scan(&probe)
	switch {
	case err == nil:
		t.Fatal("掛載交易的部門讀取未取得 FOR SHARE 列鎖:刪除端可在驗證之後提交,產生活帳號落在已軟刪部門")
	case !isRowLockNotAvailable(err):
		t.Fatalf("列鎖探測應因取不到鎖而失敗(55P03),got %v", err)
	}

	// ② 再確認「刪除真的被擋在該列鎖上」:pg_stat_activity 必須看到 B 的 UPDATE 停在 Lock 等待。
	//    沒有這步,「尚未返回」可能只是 B 還沒跑到,於是 B 在掛載提交後才取快照 → 測不到互斥。
	waitUntilDeleteBlocked(t, sqlDB, done)

	// 掛載提交 → 刪除重新評估 NOT EXISTS(users) → 看到新成員 → FailedPrecondition。
	if err := tx.Commit(); err != nil {
		t.Fatalf("提交掛載交易: %v", err)
	}
	select {
	case err := <-done:
		if connect.CodeOf(err) != connect.CodeFailedPrecondition {
			t.Fatalf("掛載提交後刪除應 FailedPrecondition(部門已有成員),got %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("掛載提交後刪除仍未返回:列鎖未被釋放")
	}

	// 部門存活、成員仍掛在該部門(活帳號不得落在已軟刪部門)。
	got, err := db.Department.Query().Where(department.ID(dep.ID)).Only(ctx)
	if err != nil {
		t.Fatalf("重新讀取部門: %v", err)
	}
	if got.DeletedAt != nil {
		t.Fatal("部門不得被軟刪除:刪除必須因「仍有成員」失敗")
	}
	m, err := db.User.Query().
		Where(user.ID(member.ID), user.HasDepartmentWith(department.ID(dep.ID))).
		Exist(ctx)
	if err != nil {
		t.Fatalf("重新讀取成員: %v", err)
	}
	if !m {
		t.Fatalf("成員 %d 必須仍掛在部門 %d(活帳號不得落在已軟刪部門)", member.ID, dep.ID)
	}

	// 對照組:成員調離後同一部門可正常刪除(證明上一段的 FailedPrecondition 來自「仍有成員」,
	// 不是列鎖把刪除永久卡住)。
	if _, err := db.User.UpdateOneID(member.ID).ClearDepartment().Save(ctx); err != nil {
		t.Fatalf("把成員調離部門: %v", err)
	}
	if _, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{
		DepartmentId: strconv.Itoa(dep.ID),
	})); err != nil {
		t.Fatalf("成員調離後刪除必須成功: %v", err)
	}
}

// isRowLockNotAvailable 判斷是否為 PostgreSQL 的 55P03(lock_not_available,FOR ... NOWAIT 取不到鎖)。
func isRowLockNotAvailable(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "55P03"
}

// waitUntilDeleteBlocked 等到 B 的刪除真的停在該部門列鎖上(pg_stat_activity 的 wait_event_type
// 為 Lock,且正在執行的敘述涉及 departments 的 UPDATE/FOR UPDATE);若 B 在等待期間返回則立刻失敗
// (返回即代表列鎖未互斥)。這是下面「提交後 B 才重新評估」的前提:沒有等到 B 卡住,就不能聲稱
// 測到了互斥。
func waitUntilDeleteBlocked(t *testing.T, sqlDB *sql.DB, done <-chan error) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for {
		select {
		case err := <-done:
			t.Fatalf("刪除在掛載交易提交前就返回(%v):部門列鎖未與刪除互斥", err)
		default:
		}
		var blocked bool
		if err := sqlDB.QueryRow(
			`SELECT EXISTS (SELECT 1 FROM pg_stat_activity
			 WHERE wait_event_type = 'Lock' AND state = 'active'
			   AND query ILIKE '%UPDATE%' AND query ILIKE '%departments%')`,
		).Scan(&blocked); err != nil {
			t.Fatalf("查 pg_stat_activity: %v", err)
		}
		if blocked {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("刪除未在 15 秒內停在部門列鎖上:掛載路徑的 FOR SHARE 未生效")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

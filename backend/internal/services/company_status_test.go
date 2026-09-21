package services

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
)

// statusActor 為測試用的觸發者(平台排程觸發時的形狀:系統 actor,仍落租戶稽核)。
var statusActor = authz.Identity{UserID: "1", Role: "super", Roles: []string{"super"}}

// statusTxCtx 開一條租戶交易並注入 ctx。
//
// SetCompanyStatus 的交易邊界由呼叫端擁有(生產為 dbtenant.Interceptor / SystemScopeTx),
// 故測試自備;commit 由呼叫端決定(未 commit 的變更不得被斷言成已落庫)。
func statusTxCtx(t *testing.T, db *ent.Client) (context.Context, *ent.Tx) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.Tx(ctx)
	if err != nil {
		t.Fatalf("開租戶交易: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback() }) // 已 commit 時 Rollback 為 no-op
	return dbtenant.WithTenantTx(ctx, tx), tx
}

// TestSetCompanyStatusWritesAuditAndSkipsNoop：
// ① 狀態真的變更時寫一筆 update/company 稽核(before/after 帶 status、after 帶原因);
// ② 同值再設一次不寫稽核(避免排程每日重跑灌爆稽核)。
func TestSetCompanyStatusWritesAuditAndSkipsNoop(t *testing.T) {
	db := newCounterTestDB(t)
	bg := context.Background()
	co := db.Company.Create().SetName("C").SetIdentifier("SETST-1").
		SetStatus(company.StatusActive).SaveX(bg)

	ctx, tx := statusTxCtx(t, db)
	if err := SetCompanyStatus(ctx, db, co.ID, company.StatusSuspended, "欠費停用", statusActor); err != nil {
		t.Fatalf("停用: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	if got := db.Company.GetX(bg, co.ID); got.Status != company.StatusSuspended {
		t.Fatalf("狀態應為 suspended,got %s", got.Status)
	}
	audits, err := db.AuditLog.Query().All(bg)
	if err != nil {
		t.Fatalf("查稽核: %v", err)
	}
	if len(audits) != 1 {
		t.Fatalf("狀態變更應寫 1 筆稽核,got %d", len(audits))
	}
	if string(audits[0].Action) != "update" || audits[0].ResourceType != "company" {
		t.Fatalf("稽核應為 update/company,got %s/%s", audits[0].Action, audits[0].ResourceType)
	}
	if audits[0].CompanyID != co.ID {
		t.Fatalf("稽核應歸屬被改的公司 %d,got %d", co.ID, audits[0].CompanyID)
	}
	if audits[0].BeforeSnapshot["status"] != string(company.StatusActive) ||
		audits[0].AfterSnapshot["status"] != string(company.StatusSuspended) {
		t.Fatalf("稽核應留 status 前後快照,got %v → %v",
			audits[0].BeforeSnapshot, audits[0].AfterSnapshot)
	}
	if audits[0].AfterSnapshot["reason"] != "欠費停用" {
		t.Fatalf("稽核應留變更原因,got %v", audits[0].AfterSnapshot["reason"])
	}

	// ② 同值再設:no-op(狀態已達標,不寫第二筆稽核)。
	ctx2, tx2 := statusTxCtx(t, db)
	if err := SetCompanyStatus(ctx2, db, co.ID, company.StatusSuspended, "欠費停用", statusActor); err != nil {
		t.Fatalf("同值設定不應報錯: %v", err)
	}
	if err := tx2.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if n := db.AuditLog.Query().CountX(bg); n != 1 {
		t.Fatalf("同值設定不得再寫稽核,got %d 筆", n)
	}
}

// TestSetCompanyStatusRejectsDeletedCompany:已軟刪除的公司不得被改狀態,
// 且錯誤要是「找不到」而非交易缺失(否則本測試會為錯的理由通過)。
func TestSetCompanyStatusRejectsDeletedCompany(t *testing.T) {
	db := newCounterTestDB(t)
	bg := context.Background()
	co := db.Company.Create().SetName("C").SetIdentifier("SETST-2").
		SetStatus(company.StatusActive).SetDeletedAt(time.Now().UTC()).SaveX(bg)

	ctx, _ := statusTxCtx(t, db)
	err := SetCompanyStatus(ctx, db, co.ID, company.StatusSuspended, "欠費停用", statusActor)
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("已刪除公司改狀態應回 NotFound,got %v", err)
	}
	if got := db.Company.GetX(bg, co.ID); got.Status != company.StatusActive {
		t.Fatalf("已刪除公司的狀態不得被改動,got %s", got.Status)
	}
}

// TestSetCompanyStatusRequiresTenantTx:ctx 沒有請求交易即拒絕 —— 靜默退回 fallback client
// 會在 RLS 下變成「自開交易被 policy 濾成 0 列」(本專案吃過的坑),交易邊界一律由呼叫端提供。
func TestSetCompanyStatusRequiresTenantTx(t *testing.T) {
	db := newCounterTestDB(t)
	bg := context.Background()
	co := db.Company.Create().SetName("C").SetIdentifier("SETST-4").
		SetStatus(company.StatusActive).SaveX(bg)

	if err := SetCompanyStatus(bg, db, co.ID, company.StatusSuspended, "欠費停用", statusActor); err == nil {
		t.Fatal("沒有請求交易不得變更狀態")
	} else if connect.CodeOf(err) != connect.CodeInternal {
		t.Fatalf("缺交易應回 Internal（呼叫端未提供交易邊界）,got %v", err)
	}
	if got := db.Company.GetX(bg, co.ID); got.Status != company.StatusActive {
		t.Fatalf("被拒的變更不得改到狀態,got %s", got.Status)
	}
}

// 未結項 #1:reason 檢查排在同值 no-op 之前 —— 同值呼叫（例：排程重跑已達標的公司）
// 若沒帶原因會被拒絕，而「無需動作」本來不該被原因阻擋（no-op 不寫稽核、無副作用）。
func TestSetCompanyStatusNoopSkipsReasonCheck(t *testing.T) {
	db := newCounterTestDB(t)
	bg := context.Background()
	co := db.Company.Create().SetName("C").SetIdentifier("SETST-5").
		SetStatus(company.StatusSuspended).SaveX(bg)

	ctx, tx := statusTxCtx(t, db)
	// 已是 suspended，再設 suspended 且不帶原因 → 應為 no-op 成功（不寫稽核）。
	if err := SetCompanyStatus(ctx, db, co.ID, company.StatusSuspended, "   ", statusActor); err != nil {
		t.Fatalf("同值 no-op 不應被原因阻擋，got %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if n := db.AuditLog.Query().CountX(bg); n != 0 {
		t.Fatalf("no-op 不得寫稽核，got %d 筆", n)
	}
}

// TestSetCompanyStatusRequiresReason:Global Constraints「寫入必附原因」——空字串(含全空白)
// 即拒絕,且不留任何痕跡(狀態未動、無稽核)。
func TestSetCompanyStatusRequiresReason(t *testing.T) {
	db := newCounterTestDB(t)
	bg := context.Background()
	co := db.Company.Create().SetName("C").SetIdentifier("SETST-3").
		SetStatus(company.StatusActive).SaveX(bg)

	ctx, tx := statusTxCtx(t, db)
	if err := SetCompanyStatus(ctx, db, co.ID, company.StatusSuspended, "   ", statusActor); err == nil {
		t.Fatal("未附原因不得變更狀態")
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if got := db.Company.GetX(bg, co.ID); got.Status != company.StatusActive {
		t.Fatalf("被拒的變更不得改到狀態,got %s", got.Status)
	}
	if n := db.AuditLog.Query().CountX(bg); n != 0 {
		t.Fatalf("被拒的變更不得寫稽核,got %d 筆", n)
	}
}

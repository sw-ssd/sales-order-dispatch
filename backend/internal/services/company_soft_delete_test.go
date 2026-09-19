package services

// P2-A(2026-09-20)回歸:公司改軟刪除 + 刪除路徑的錯誤映射。
//
// 缺陷現場:DeleteCompany 以硬刪除刪列,而 audit_logs.company_id 是租戶欄(NOT NULL + FK),
// 只要公司有任何稽核列(UpdateCompany 改一次狀態就會寫一列)硬刪除即違反 FK;toConnectError
// 把該約束錯誤映射成 AlreadyExists,前端顯示「識別碼(identifier)已存在,請換一個」並外洩
// 原始 SQL 訊息(見 task-5/task-9 report 的 F1/F2)。
//
// 本檔守住四件事:軟刪除列保留(識別碼可重用)、刪除寫稽核且 company_id 正確、
// 所有公司查詢排除已刪除列、前置檢查(仍有部門/使用者)不變;另守住約束錯誤不外洩 DB 原文。

import (
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/auditlog"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// TestDeleteCompanySoftDelete P2-A:刪除公司為軟刪除 —— 列保留(稽核 FK 永遠有主可依)、
// 同一交易寫 action=delete 稽核(company_id 指向被刪公司)、查詢不再可見、識別碼可被新公司重用。
func TestDeleteCompanySoftDelete(t *testing.T) {
	ctx := t.Context()
	super := authz.Identity{UserID: "1", CompanyID: "1", Role: "super", Roles: []string{"super"}}
	cc, _, db := newTestServerWithIdentity(t, super)
	co := db.Company.Create().SetName("公司A").SetIdentifier("SOFT-1").SaveX(ctx)
	id := uItoa(co.ID)

	// 先做一次會寫稽核的更新:舊行為(硬刪除)在此條件下必違反 audit_logs.company_id 的 FK。
	if _, err := cc.UpdateCompany(ctx, connect.NewRequest(&v1.UpdateCompanyRequest{CompanyId: id, Status: strPtr("inactive")})); err != nil {
		t.Fatalf("前置 UpdateCompany: %v", err)
	}

	if _, err := cc.DeleteCompany(ctx, connect.NewRequest(&v1.DeleteCompanyRequest{CompanyId: id})); err != nil {
		t.Fatalf("DeleteCompany(有稽核列的公司): %v", err)
	}

	// ① 列保留、deleted_at 已標記(硬刪除會讓這一列消失)。
	row, err := db.Company.Get(ctx, co.ID)
	if err != nil {
		t.Fatalf("公司刪除必須只標記 deleted_at(軟刪除),列不得消失: %v", err)
	}
	if row.DeletedAt == nil {
		t.Fatal("公司刪除必須標記 deleted_at")
	}

	// ② 刪除寫稽核:action=delete、resource_type=company、company_id=被刪公司。
	deletes := db.AuditLog.Query().
		Where(auditlog.ActionEQ(auditlog.ActionDelete), auditlog.ResourceTypeEQ("company")).
		AllX(ctx)
	if len(deletes) != 1 {
		t.Fatalf("刪除公司應寫 1 筆 action=delete 稽核,got %d", len(deletes))
	}
	if deletes[0].CompanyID != co.ID {
		t.Fatalf("稽核列 company_id 應為被刪公司 %d,got %d", co.ID, deletes[0].CompanyID)
	}
	if deletes[0].ResourceID != id {
		t.Fatalf("稽核列 resource_id 應為 %q,got %q", id, deletes[0].ResourceID)
	}
	if name, _ := deletes[0].BeforeSnapshot["name"].(string); name != "公司A" {
		t.Fatalf("刪除稽核應留 before 快照(name=公司A),got %v", deletes[0].BeforeSnapshot)
	}

	// ① 查詢不可見:Get 與 List 皆排除已刪除列。
	if _, err := cc.GetCompany(ctx, connect.NewRequest(&v1.GetCompanyRequest{CompanyId: id})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("刪除後 GetCompany 應回 NotFound,got %v", err)
	}
	list, err := cc.ListCompanies(ctx, connect.NewRequest(&v1.ListCompaniesRequest{}))
	if err != nil {
		t.Fatalf("ListCompanies: %v", err)
	}
	if list.Msg.GetPagination().GetTotal() != 0 || len(list.Msg.GetCompanies()) != 0 {
		t.Fatalf("刪除後清單應為 0 筆,got total=%d items=%d", list.Msg.GetPagination().GetTotal(), len(list.Msg.GetCompanies()))
	}

	// ① UpdateCompany 的存在性檢查同樣排除已刪除列。
	name := "公司A(更名)"
	if _, err := cc.UpdateCompany(ctx, connect.NewRequest(&v1.UpdateCompanyRequest{CompanyId: id, Name: &name})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("更新已刪除公司應回 NotFound,got %v", err)
	}

	// 冪等語意:已刪除的公司再刪一次 → NotFound。
	if _, err := cc.DeleteCompany(ctx, connect.NewRequest(&v1.DeleteCompanyRequest{CompanyId: id})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("重複刪除應回 NotFound,got %v", err)
	}

	// ③ 識別碼在軟刪除後可重用(部分唯一索引 companies_identifier_active_unique):
	// 沒了這條,新公司會被舊列的表層 UNIQUE 擋下。
	created, err := cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{Name: "公司A(新)", Identifier: "SOFT-1"}))
	if err != nil {
		t.Fatalf("軟刪除後同識別碼應可再建公司: %v", err)
	}
	if created.Msg.GetCompany().GetIdentifier() != "SOFT-1" {
		t.Fatalf("新公司識別碼不符:%+v", created.Msg.GetCompany())
	}
	// 未刪除列之間仍唯一:真正的重複識別碼 → AlreadyExists(且訊息不含 DB 原文)。
	_, err = cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{Name: "重複", Identifier: "SOFT-1"}))
	if connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Fatalf("未刪除列之間識別碼重複應回 AlreadyExists,got %v", err)
	}
	for _, leak := range []string{"UNIQUE", "constraint", "SQLSTATE", "sqlite"} {
		if strings.Contains(err.Error(), leak) {
			t.Fatalf("AlreadyExists 訊息不得外洩 DB 原文(%q):%v", leak, err)
		}
	}
}

// TestCompanyQueriesExcludeSoftDeleted P2-A:以 DB 直寫 deleted_at(模擬歷史遺留的軟刪除列),
// ListCompanies(含 total)與 GetCompany 一律看不到 —— 沒有這道過濾,已刪公司會回到清單與詳情頁。
func TestCompanyQueriesExcludeSoftDeleted(t *testing.T) {
	ctx := t.Context()
	super := authz.Identity{UserID: "1", CompanyID: "1", Role: "super", Roles: []string{"super"}}
	cc, _, db := newTestServerWithIdentity(t, super)
	alive := db.Company.Create().SetName("存續").SetIdentifier("KEEP-1").SaveX(ctx)
	gone := db.Company.Create().SetName("已刪").SetIdentifier("GONE-1").SaveX(ctx)
	if _, err := db.Company.UpdateOneID(gone.ID).SetDeletedAt(time.Now().UTC()).Save(ctx); err != nil {
		t.Fatalf("標記軟刪除: %v", err)
	}

	if _, err := cc.GetCompany(ctx, connect.NewRequest(&v1.GetCompanyRequest{CompanyId: uItoa(gone.ID)})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("已刪公司在 GetCompany 應 NotFound,got %v", err)
	}
	list, err := cc.ListCompanies(ctx, connect.NewRequest(&v1.ListCompaniesRequest{}))
	if err != nil {
		t.Fatalf("ListCompanies: %v", err)
	}
	if list.Msg.GetPagination().GetTotal() != 1 || len(list.Msg.GetCompanies()) != 1 {
		t.Fatalf("清單應只剩 1 筆存續公司,got total=%d items=%d", list.Msg.GetPagination().GetTotal(), len(list.Msg.GetCompanies()))
	}
	if got := list.Msg.GetCompanies()[0].GetId(); got != uItoa(alive.ID) {
		t.Fatalf("清單應為存續公司 %d,got %s", alive.ID, got)
	}
	// 關鍵字也只搜得到未刪除的公司(identifier/name 篩選套在同一組 predicate 上)。
	kw, err := cc.ListCompanies(ctx, connect.NewRequest(&v1.ListCompaniesRequest{Keyword: "GONE"}))
	if err != nil {
		t.Fatalf("ListCompanies(keyword): %v", err)
	}
	if kw.Msg.GetPagination().GetTotal() != 0 {
		t.Fatalf("keyword 不應命中已刪除公司,got %d", kw.Msg.GetPagination().GetTotal())
	}
}

// TestDeleteCompanyBlockedByDepartmentsAndUsers P2-A:軟刪除不改變前置檢查 ——
// 仍有部門或使用者時回 FailedPrecondition(且公司必須仍存在,不得被標記刪除)。
func TestDeleteCompanyBlockedByDepartmentsAndUsers(t *testing.T) {
	ctx := t.Context()
	super := authz.Identity{UserID: "1", CompanyID: "1", Role: "super", Roles: []string{"super"}}
	cc, _, db := newTestServerWithIdentity(t, super)

	withDept := db.Company.Create().SetName("有部門").SetIdentifier("BLK-DEPT").SaveX(ctx)
	dept := db.Department.Create().SetName("門市一").SetCompanyID(withDept.ID).SaveX(ctx)
	if _, err := cc.DeleteCompany(ctx, connect.NewRequest(&v1.DeleteCompanyRequest{CompanyId: uItoa(withDept.ID)})); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("仍有部門應回 FailedPrecondition,got %v", err)
	}
	if row := db.Company.GetX(ctx, withDept.ID); row.DeletedAt != nil {
		t.Fatal("前置檢查失敗時不得標記 deleted_at")
	}

	withUser := db.Company.Create().SetName("有使用者").SetIdentifier("BLK-USER").SaveX(ctx)
	db.User.Create().SetEmail("blk@example.com").SetName("員工").SetStatus(user.StatusActive).
		SetRole("staff").SetPasswordHash("x").SetCompanyID(withUser.ID).SaveX(ctx)
	if _, err := cc.DeleteCompany(ctx, connect.NewRequest(&v1.DeleteCompanyRequest{CompanyId: uItoa(withUser.ID)})); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("仍有使用者應回 FailedPrecondition,got %v", err)
	}

	// 前置條件解除後才刪得掉(部門刪除後,部門內不得有使用者)。
	if err := db.Department.DeleteOneID(dept.ID).Exec(ctx); err != nil {
		t.Fatalf("清掉部門: %v", err)
	}
	if _, err := cc.DeleteCompany(ctx, connect.NewRequest(&v1.DeleteCompanyRequest{CompanyId: uItoa(withDept.ID)})); err != nil {
		t.Fatalf("前置條件解除後應可刪除: %v", err)
	}
}

// TestToConnectErrorConstraintSanitized P2-A:約束錯誤的外洩面 —— ent 的 ConstraintError 挾帶
// 驅動層原文(SQLSTATE/約束名/欄位名),映射後必須只剩繁中可行動訊息,且不得誤報 AlreadyExists
// (舊映射把任何約束錯誤都當成識別碼重複,前端才會顯示「識別碼已存在」)。
func TestToConnectErrorConstraintSanitized(t *testing.T) {
	ctx := t.Context()
	_, _, db := newTestServerWithIdentity(t, authz.Identity{UserID: "1", CompanyID: "1", Role: "super", Roles: []string{"super"}})
	db.Company.Create().SetName("甲").SetIdentifier("DUP-1").SaveX(ctx)

	// 繞過服務層前置查詢,取得真實的 ent 約束錯誤。
	_, raw := db.Company.Create().SetName("乙").SetIdentifier("DUP-1").Save(ctx)
	if !ent.IsConstraintError(raw) {
		t.Fatalf("前置條件失敗:預期 ent 約束錯誤,got %v", raw)
	}
	mapped := toConnectError(raw)
	if got := connect.CodeOf(mapped); got != connect.CodeFailedPrecondition {
		t.Fatalf("約束錯誤應映射為 FailedPrecondition,got %v(%v)", got, mapped)
	}
	for _, leak := range []string{"UNIQUE", "unique", "constraint", "SQLSTATE", "companies.identifier", "sqlite"} {
		if strings.Contains(mapped.Error(), leak) {
			t.Fatalf("映射後的訊息不得外洩 DB 原文(%q):%v", leak, mapped.Error())
		}
	}
}

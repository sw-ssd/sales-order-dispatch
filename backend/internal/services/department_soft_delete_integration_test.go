//go:build integration

// 部門軟刪除(00020)的真 PostgreSQL 整合測試:跑真正的 goose 遷移(含新增的 00020),
// 以真實 service handler + ent client 觀察只有真 PG 才看得見的三件事 ——
//
//	① 00020 的欄位成立,且 departments **沒有**表層 UNIQUE(00005 只建 id/name/company_departments)
//	   → 因此不像公司 P2-A 需要把表層 UNIQUE 換成部分唯一索引,本檔只加欄位;若日後新增部門
//	   唯一鍵,必須比照 00019 用 `WHERE deleted_at IS NULL` 的部分唯一索引表達。
//	   該斷言以 pg_constraint(contype='u')與 pg_index(indisunique AND indpred IS NULL)兩者合看 ——
//	   只看前者抓不到 `CREATE UNIQUE INDEX`(複審 M3)。
//	② 原始缺陷在真 PG 上已不可重現:部門成員被稽核(recordUserAudit 以目標使用者的部門寫
//	   department_id)後被調離部門 → DeleteDepartment 成功且稽核留痕。舊行為在此必被
//	   audit_logs_department_id_fkey 擋下 → failed_precondition(訊息與原因無關)。
//	③ 軟刪除後 FK/查詢行為:以 department_id 為 FK 的其他表(倉別)同樣不再阻擋刪除、列保留、
//	   已刪除部門不出現在清單與詳情,且不得再被掛載(CreateUser 帶已刪除的 department_id)。
//
// 為何 sqlite(enttest)不足以守住:ent 未對 audit_logs.department_id 建 edge(純欄位),
// sqlite 不建該 FK,硬刪除的 FK 違反只有在真 PG 才重現。
// 遷移以 cmd/migrate 相同路徑套用(同 dialect、同目錄、同版本表),不另寫複製的 DDL。
package services

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationDepartmentSoftDelete 00020 回歸(見檔頭說明)。
func TestIntegrationDepartmentSoftDelete(t *testing.T) {
	testsupport.RequiresContainer(t)
	ctx := t.Context()
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	sqlDB, db := openPGEntClientFromGoose(t, dsn)

	// 稽核列有 company_id/user_id 的 FK(00010),故操作者必須是真實存在的使用者。
	actorCo := db.Company.Create().SetName("操作者公司").SetIdentifier("D20-ACTOR").SaveX(ctx)
	actor := db.User.Create().
		SetEmail("d20-actor@example.com").
		SetName("操作者").
		SetStatus("active").
		SetRole("super").
		SetPasswordHash("x").
		SetCompanyID(actorCo.ID).
		SaveX(ctx)
	cc, dc, uc := newDepartmentSoftDeleteServer(t, db, authz.Identity{
		UserID: strconv.Itoa(actor.ID), CompanyID: strconv.Itoa(actorCo.ID), Role: "super", Roles: []string{"super"},
	})

	t.Run("00020 欄位與無表層 UNIQUE", func(t *testing.T) {
		if !columnExists(t, sqlDB, "departments", "deleted_at") {
			t.Fatal("00020 必須為 departments 加上 deleted_at(軟刪除的載體)")
		}
		var tableUnique int
		if err := sqlDB.QueryRow(
			`SELECT count(*) FROM pg_constraint WHERE conrelid = 'departments'::regclass AND contype = 'u'`,
		).Scan(&tableUnique); err != nil {
			t.Fatalf("查 departments 的表層 UNIQUE: %v", err)
		}
		if tableUnique != 0 {
			t.Fatalf("departments 不應有表層 UNIQUE(有則必須比照 00019 改為部分唯一索引),got %d", tableUnique)
		}
		// `CREATE UNIQUE INDEX`(含條件式唯一索引)不會出現在 pg_constraint,故另以 pg_index 直查:
		// **無條件**唯一索引只能有 pkey 一個。帶 WHERE 的部分唯一索引是允許的(00019 對公司 identifier
		// 的做法),但必須以 `WHERE deleted_at IS NULL` 表達,否則已刪除列會永久佔用該鍵。
		if got := unconditionalUniqueIndexes(t, sqlDB, "departments"); got != 1 {
			t.Fatalf("departments 的無條件唯一索引只應有 pkey 一個(部分唯一索引須以 WHERE deleted_at IS NULL 表達),got %d", got)
		}
		// ② 的前提:audit_logs.department_id 的 FK 真的存在。
		var fk int
		if err := sqlDB.QueryRow(
			`SELECT count(*) FROM pg_constraint WHERE conname = 'audit_logs_department_id_fkey'`,
		).Scan(&fk); err != nil {
			t.Fatalf("查 audit_logs_department_id_fkey: %v", err)
		}
		if fk != 1 {
			t.Fatalf("audit_logs.department_id 應有 FK(00010),got %d", fk)
		}
	})

	t.Run("刪除部門:稽核 FK 不再阻擋且稽核留痕", func(t *testing.T) {
		co, err := cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{
			Name: "部門測試公司", Identifier: "D20-DEPT",
		}))
		if err != nil {
			t.Fatalf("CreateCompany: %v", err)
		}
		coID := co.Msg.GetCompany().GetId()

		created, err := dc.CreateDepartment(ctx, connect.NewRequest(&v1.CreateDepartmentRequest{CompanyId: coID, Name: "門市一"}))
		if err != nil {
			t.Fatalf("CreateDepartment: %v", err)
		}
		deptID := created.Msg.GetDepartment().GetId()
		did, err := parseID(deptID)
		if err != nil {
			t.Fatalf("解析部門 id %q: %v", deptID, err)
		}

		// 製造缺陷現場:部門成員被稽核(寫下 department_id = 該部門的稽核列)後被調離部門。
		member, err := uc.CreateUser(ctx, connect.NewRequest(&v1.CreateUserRequest{
			Name: "成員", Email: "d20-member@example.com", CompanyId: coID, Role: "staff", DepartmentId: deptID,
		}))
		if err != nil {
			t.Fatalf("前置 CreateUser: %v", err)
		}
		var audits int
		if err := sqlDB.QueryRow(`SELECT count(*) FROM audit_logs WHERE department_id = $1`, did).Scan(&audits); err != nil {
			t.Fatalf("查部門稽核列: %v", err)
		}
		if audits == 0 {
			t.Fatal("前置條件失敗:成員的稽核列應以該部門為 department_id(硬刪除的 FK 陷阱來源)")
		}
		if _, err := uc.UpdateUser(ctx, connect.NewRequest(&v1.UpdateUserRequest{
			UserId: member.Msg.GetUser().GetId(), DepartmentId: strPtr(""),
		})); err != nil {
			t.Fatalf("前置 UpdateUser(把成員調離部門): %v", err)
		}

		// 舊行為在此必 FK 違反 → failed_precondition;軟刪除後必須成功。
		if _, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{DepartmentId: deptID})); err != nil {
			t.Fatalf("刪除有稽核列的部門必須成功(舊行為會被 audit_logs_department_id_fkey 擋下): %v", err)
		}
		var softDeleted int
		if err := sqlDB.QueryRow(`SELECT count(*) FROM departments WHERE id = $1 AND deleted_at IS NOT NULL`, did).Scan(&softDeleted); err != nil {
			t.Fatalf("查軟刪除列: %v", err)
		}
		if softDeleted != 1 {
			t.Fatal("部門刪除必須只標記 deleted_at,列必須保留(所有 FK 才有主可依)")
		}
		// 稽核列仍在且 FK 成立:delete 稽核的 department_id 正是被刪部門。
		var deleteAudits int
		if err := sqlDB.QueryRow(
			`SELECT count(*) FROM audit_logs WHERE department_id = $1 AND action = 'delete' AND resource_type = 'department'`, did,
		).Scan(&deleteAudits); err != nil {
			t.Fatalf("查 delete 稽核列: %v", err)
		}
		if deleteAudits != 1 {
			t.Fatalf("刪除應寫 1 筆 action=delete/resource_type=department 稽核,got %d", deleteAudits)
		}

		list, err := dc.ListDepartments(ctx, connect.NewRequest(&v1.ListDepartmentsRequest{CompanyId: coID}))
		if err != nil {
			t.Fatalf("ListDepartments: %v", err)
		}
		if list.Msg.GetPagination().GetTotal() != 0 || len(list.Msg.GetDepartments()) != 0 {
			t.Fatalf("已刪除部門不得出現在清單,got total=%d items=%d",
				list.Msg.GetPagination().GetTotal(), len(list.Msg.GetDepartments()))
		}
		if _, err := dc.GetDepartment(ctx, connect.NewRequest(&v1.GetDepartmentRequest{DepartmentId: deptID})); connect.CodeOf(err) != connect.CodeNotFound {
			t.Fatalf("已刪除部門 GetDepartment 應 NotFound,got %v", err)
		}
		if _, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{DepartmentId: deptID})); connect.CodeOf(err) != connect.CodeNotFound {
			t.Fatalf("重複刪除應 NotFound,got %v", err)
		}
	})

	t.Run("軟刪除後 FK 與掛載路徑", func(t *testing.T) {
		co, err := cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{
			Name: "FK 測試公司", Identifier: "D20-FK",
		}))
		if err != nil {
			t.Fatalf("CreateCompany: %v", err)
		}
		coID := co.Msg.GetCompany().GetId()
		cid, err := parseID(coID)
		if err != nil {
			t.Fatalf("解析公司 id %q: %v", coID, err)
		}
		created, err := dc.CreateDepartment(ctx, connect.NewRequest(&v1.CreateDepartmentRequest{CompanyId: coID, Name: "有倉別的部門"}))
		if err != nil {
			t.Fatalf("CreateDepartment: %v", err)
		}
		deptID := created.Msg.GetDepartment().GetId()
		did, err := parseID(deptID)
		if err != nil {
			t.Fatalf("解析部門 id %q: %v", deptID, err)
		}

		// 以 department_id 為 FK 的其他表(warehouses_department_fk):舊行為在此必 FK 違反。
		wh := db.Warehouse.Create().SetCompanyID(cid).SetDepartmentID(did).SetCode("D20-WH").SetName("冷藏倉").SaveX(ctx)
		if _, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{DepartmentId: deptID})); err != nil {
			t.Fatalf("部門仍有倉別時刪除必須成功(軟刪除;舊行為會被 warehouses_department_fk 擋下): %v", err)
		}
		var whDept int
		if err := sqlDB.QueryRow(`SELECT department_id FROM warehouses WHERE id = $1`, wh.ID).Scan(&whDept); err != nil {
			t.Fatalf("查倉別: %v", err)
		}
		if whDept != did {
			t.Fatalf("倉別的 department_id 應仍指向被軟刪除的部門 %d,got %d", did, whDept)
		}

		// 掛載路徑:不得再把帳號掛進已軟刪除的部門。
		_, err = uc.CreateUser(ctx, connect.NewRequest(&v1.CreateUserRequest{
			Name: "新人", Email: "d20-new@example.com", CompanyId: coID, Role: "staff", DepartmentId: deptID,
		}))
		if connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Fatalf("掛到已刪除部門應 invalid_argument,got %v", err)
		}
		var users int
		if err := sqlDB.QueryRow(`SELECT count(*) FROM users WHERE email = 'd20-new@example.com'`).Scan(&users); err != nil {
			t.Fatalf("查帳號: %v", err)
		}
		if users != 0 {
			t.Fatalf("掛載被擋後不得留下帳號,got %d 筆", users)
		}
	})

	t.Run("00020 Down 對稱還原", func(t *testing.T) {
		// 專屬容器:Down 會把 deleted_at 移除,故驗的是全新庫的對稱性(Up/Down 皆可重複套用)。
		downDSN := testsupport.Postgres(t)
		migrateBusinessUp(t, downDSN)
		downDB := openRawDB(t, downDSN)
		if !columnExists(t, downDB, "departments", "deleted_at") {
			t.Fatal("Up 後應有 departments.deleted_at")
		}
		migrateBusinessDownTo(t, downDSN, "19")
		if columnExists(t, downDB, "departments", "deleted_at") {
			t.Fatal("Down 必須移除 departments.deleted_at")
		}
		migrateBusinessUp(t, downDSN)
		if !columnExists(t, downDB, "departments", "deleted_at") {
			t.Fatal("重跑 Up 後 departments.deleted_at 應復原")
		}
	})
}

// newDepartmentSoftDeleteServer 以指定身分掛載公司 + 部門 + 使用者 handler(與 sqlite 版同構,只換 DB)。
func newDepartmentSoftDeleteServer(t *testing.T, db *ent.Client, id authz.Identity) (
	salesorderv1connect.CompanyServiceClient,
	salesorderv1connect.DepartmentServiceClient,
	salesorderv1connect.UserServiceClient,
) {
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
		salesorderv1connect.NewDepartmentServiceClient(http.DefaultClient, ts.URL),
		salesorderv1connect.NewUserServiceClient(http.DefaultClient, ts.URL)
}

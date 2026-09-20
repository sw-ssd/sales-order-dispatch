//go:build integration

package services

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationRLSViolationMaskedForClient 釘住計畫 Global Constraints「錯誤一律經
// `toConnectError` 映射,不得回傳 SQLSTATE 或 constraint 名」在 **RLS 違反**上的落實。
//
// 缺陷現場:ent 只把「唯一鍵／FK／CHECK」判為 constraint error(ent v0.14.6
// dialect/sql/sqlgraph/errors.go 以字串比對 `violates unique constraint`…),而 PG 的 RLS 違反是
// **SQLSTATE 42501**、訊息為 `new row violates row-level security policy for table "x"` —— 兩者都
// 對不上 → 落到 `toConnectError` 的 default 分支,以 `CodeInternal` 把驅動層原文(SQLSTATE +
// 表名 + policy 文字)逐字回給客戶端。本測試以真 PG(app_rw)觸發它,再走**實際的映射函式**。
//
// 為何在 ent 層觸發而非挑一條 RPC:無 scope 的 RPC 會在**寫入之前**的範圍讀取就失敗
// (實測 CreateCustomer 回 `invalid_argument: default_sales_rep_id 無效`,users 讀不到)——
// 那是另一條分支,證明不了 42501。故刻意以無 scope 的業務連線直接 INSERT(等同漏掛 SET LOCAL
// 的請求交易),讓 WITH CHECK 擋下,並把 ent 交上來的原始錯誤餵給 `toConnectError`。
//
// 三條斷言各釘一半:
//
//	① 錯誤碼與同檔 `ent.IsConstraintError` 分支一致(對外語意是「資料現況不允許」,非伺服器故障);
//	② 對外訊息不得含 SQLSTATE／表名／policy 文字(遮蔽);
//	③ 根因仍必須落 server log(SQLSTATE + policy 原文),否則維運無從診斷。
func TestIntegrationRLSViolationMaskedForClient(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	defer func() { _ = admin.Close() }()
	coA := insertRLSCompany(t, admin, "A", "MASK-A")

	// 業務 client(app_rw + dbtenant 裝飾器);ctx 不帶 tenant 交易 → 無 SET LOCAL → WITH CHECK 擋。
	client := openAppRoleEntClient(t, adminDSN)
	raw, err := client.Customer.Create().
		SetCompanyID(coA).
		SetCustomerCode("MASK-1").
		SetName("無 scope 客戶").
		Save(t.Context())
	if raw != nil {
		t.Fatalf("無 scope 的業務寫入必須被擋,卻成功建立客戶 %d", raw.ID)
	}
	if !isRLSPolicyViolation(err) {
		t.Fatalf("前提:該錯誤必須是 RLS 違反(42501),got %v", err)
	}
	// 前置驗證的錯誤訊息確實挾帶表名與 policy 原文(否則下面的遮蔽斷言沒有鑑別力)。
	if !strings.Contains(err.Error(), "customers") || !strings.Contains(err.Error(), "row-level security") {
		t.Fatalf("前提:驅動層原文應含表名與 policy 文字,got %q", err.Error())
	}

	var logged bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&logged)
	t.Cleanup(func() { log.SetOutput(prev) })

	// 對外錯誤:走真實映射函式(所有 RPC 路徑的共同出口)。
	out := toConnectError(err)

	if code := connect.CodeOf(out); code != connect.CodeFailedPrecondition {
		t.Fatalf("RLS 違反應映射為 failed_precondition(與約束錯誤分支同語意),got %v(%v)", code, out)
	}
	for _, banned := range []string{"SQLSTATE", "42501", "row-level security", "customers", "policy"} {
		if strings.Contains(out.Error(), banned) {
			t.Fatalf("對外訊息不得含 %q(驅動層原文),got %q", banned, out.Error())
		}
	}
	got := logged.String()
	if !strings.Contains(got, "42501") || !strings.Contains(got, "row-level security") {
		t.Fatalf("根因(SQLSTATE 與 policy 原文)必須落 server log 供維運診斷,got %q", got)
	}
	// 被擋下的寫入不得落地(superuser 真值查詢)。
	if n := countRows(t, admin, `SELECT count(*) FROM customers WHERE company_id = $1`, coA); n != 0 {
		t.Fatalf("被擋下的建檔不得落地,公司 %d 應 0 筆客戶,得到 %d", coA, n)
	}
	// customer 只是被測的其中一張:任一表的 RLS 違反都必須走同一條遮蔽(此處斷言同一分支涵蓋
	// 別張表,避免有人把判斷寫成「表名 == customers」)。
	if _, err := client.Customer.Query().Where(customer.IDEQ(1)).All(t.Context()); err != nil {
		t.Fatalf("同一個無 scope 連線的讀取應為 0 列而非錯誤(證明遮蔽僅作用於寫入違反),got %v", err)
	}
}

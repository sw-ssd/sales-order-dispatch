//go:build integration

package services

import (
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
)

// seedTx 在明確的系統範圍(scope=all)內執行 fixture 寫入 fn。
//
// 為何需要:租戶表 ENABLE(+FORCE)後,未帶 scope 的寫入會被 WITH CHECK 擋下,而該錯誤
// (SQLSTATE 42501)長得像「服務路徑漏掛 dbtenant.Client」,會把測試紅的原因指向錯的地方。
// 把 fixture 建立集中在這一個入口,讓「系統範圍」在呼叫點顯眼可審計(T6–T9 各域的 fixture
// 沿用同一輔助)。刻意**不**用「session 級預設 scope=all」繞過:那會讓漏掛 scope 的服務路徑
// 不再 fail-closed,等於拆掉測試的守門能力。
//
// SET LOCAL 由 dbtenant 的 driver 裝飾器在交易開啟時套用,因此本輔助真正生效的前提是
// fixture 的 client 由 dbtenant.NewClient 建立。現況(2026-09-20,T5)內部整合測試用的是
// 容器的 superuser + 未裝飾的 ent client(PG 的 superuser 永遠繞過 RLS),故此包裝今日
// 只是「顯式宣告系統範圍」;待 fixture client 換成 dbtenant.NewClient + 帶 scope 的請求交易
// (T9/T10 的範圍)之後,它就是 fixture 能否寫入的前提。
func seedTx(t *testing.T, db *ent.Client, fn func(tx *ent.Tx) error) {
	t.Helper()
	if err := dbtenant.SystemScopeTx(t.Context(), db, fn); err != nil {
		t.Fatalf("fixture 系統範圍寫入失敗: %v", err)
	}
}

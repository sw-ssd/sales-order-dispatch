-- 訂單域啟用 RLS(D36 的第五批:sales_orders / sales_order_items / sales_order_events / order_counters,ENABLE + FORCE)。
--
-- 順序:policy(00031 定義,USING 與 WITH CHECK 同條件)+ 服務收斂(05 Task 2-5:取號/狀態機/CRUD/組裝
-- 全經 dbtenant.Client/TxFrom,本檔落地前已驗) → **本檔只做 ENABLE + FORCE**。
-- policy 不動:00031 已建四個 policy(events 拆 SELECT/INSERT 兩條,僅追加);本檔再寫一次只會製造第二份定義。
--
-- 啟用後的行為(呼叫端必須配合):
--   1. **未帶 scope 的查詢一律 fail-closed(0 列)**;服務層任何查詢都必須走
--      `dbtenant.Client(ctx, s.db)`(請求交易內已由 driver 裝飾器 SET LOCAL)。漏掛 → 不是權限
--      錯誤,是「查不到資料」(0 列)。
--   2. FORCE 讓 **table owner** 也受 policy 約束(PG 的 superuser 仍永遠繞過 RLS,FORCE 亦然)。
--      後果:日後對本域做資料回填的 migration 或 seed,必須先
--      `SET LOCAL app.current_data_scope = 'all'`(`dbtenant.SystemScopeTx`),否則以 owner
--      身分寫入會被 WITH CHECK 擋下。
--   3. **子表傳遞性**:sales_order_items 帶冗餘 company_id/department_id,policy 直比對免 JOIN;
--      sales_order_events 僅 SELECT/INSERT(僅追加,無 UPDATE/DELETE)。
--
-- Up/Down 對稱:Up 先 ENABLE 再 FORCE,Down 必須**先 NO FORCE 再 DISABLE**(比照 00024-00028)。
-- 冪等:ALTER TABLE … ENABLE/FORCE/DISABLE ROW LEVEL SECURITY 對已處於該狀態的表為 no-op。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE sales_orders             ENABLE ROW LEVEL SECURITY;
ALTER TABLE sales_orders             FORCE  ROW LEVEL SECURITY;
ALTER TABLE sales_order_items        ENABLE ROW LEVEL SECURITY;
ALTER TABLE sales_order_items        FORCE  ROW LEVEL SECURITY;
ALTER TABLE sales_order_events       ENABLE ROW LEVEL SECURITY;
ALTER TABLE sales_order_events       FORCE  ROW LEVEL SECURITY;
ALTER TABLE order_counters           ENABLE ROW LEVEL SECURITY;
ALTER TABLE order_counters           FORCE  ROW LEVEL SECURITY;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sales_orders             NO FORCE ROW LEVEL SECURITY;
ALTER TABLE sales_orders             DISABLE ROW LEVEL SECURITY;
ALTER TABLE sales_order_items        NO FORCE ROW LEVEL SECURITY;
ALTER TABLE sales_order_items        DISABLE ROW LEVEL SECURITY;
ALTER TABLE sales_order_events       NO FORCE ROW LEVEL SECURITY;
ALTER TABLE sales_order_events       DISABLE ROW LEVEL SECURITY;
ALTER TABLE order_counters           NO FORCE ROW LEVEL SECURITY;
ALTER TABLE order_counters           DISABLE ROW LEVEL SECURITY;
-- +goose StatementEnd

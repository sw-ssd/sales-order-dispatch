-- 客戶域啟用 RLS(D36 的第一批:ENABLE + FORCE)。
--
-- 順序:policy(00013/00015 定義、00023 補 WITH CHECK)→ 本檔 ENABLE + FORCE。
-- 00023 之前這些表只「定義」policy 而沒有 ENABLE,RLS 形同註解(D3 的階段性安排:
-- 待請求層交易與服務層路徑遷移就緒後才啟用 —— T4 的 interceptor／dbtenant.Client 已把
-- 每個 unary RPC 的交易與 SET LOCAL app.* 佈好,T5 才把這四張表的守門真正打開)。
--
-- 啟用後的行為(呼叫端必須配合,已於同一波完成):
--   1. **未帶 scope 的查詢一律 fail-closed(0 列)**,而不是「看得到全部」。因此服務層任何
--      查詢都必須走 `dbtenant.Client(ctx, s.db)`(請求交易內已由 driver 裝飾器 SET LOCAL);
--      寫入路徑同樣(且 WITH CHECK 會擋下跨租戶的新列)。漏掛 → 不是權限錯誤,是「查不到資料」。
--   2. FORCE 讓 **table owner** 也受 policy 約束(PG 的 superuser 仍永遠繞過 RLS,FORCE 亦然;
--      生產環境的 owner 角色並非 superuser,故路徑必須比照業務角色)。後果:日後對本域做資料
--      回填的 migration 或 seed,必須先 `SET LOCAL app.current_data_scope = 'all'`
--      (`dbtenant.SystemScopeTx`),否則以 owner 身分寫入會被 WITH CHECK 擋下。
--   3. customer_counters 的 policy 只看 company_id(該表無 department_id,每公司一列);
--      customers/customer_addresses/customer_contacts 另受 department_id 條件(department_id
--      IS NULL 為公司層共用列)。
--
-- Up/Down 對稱:Up 先 ENABLE 再 FORCE,Down 必須**先 NO FORCE 再 DISABLE**
--(反序雖然也能跑,但會留下一段「ENABLE 且仍然 FORCE」的中間狀態;DISABLE 本身即清除
-- FORCE 旗標,故先 NO FORCE 讓 Down 的每一步都回到上一個穩定狀態)。
--
-- 冪等:ALTER TABLE … ENABLE/FORCE/DISABLE ROW LEVEL SECURITY 對已處於該狀態的表為 no-op,
-- 故 Up、Down 皆可重複套用(比照 00008/00010/00019 的冪等慣例)。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE customers            ENABLE ROW LEVEL SECURITY;
ALTER TABLE customers            FORCE  ROW LEVEL SECURITY;
ALTER TABLE customer_counters    ENABLE ROW LEVEL SECURITY;
ALTER TABLE customer_counters    FORCE  ROW LEVEL SECURITY;
ALTER TABLE customer_addresses   ENABLE ROW LEVEL SECURITY;
ALTER TABLE customer_addresses   FORCE  ROW LEVEL SECURITY;
ALTER TABLE customer_contacts    ENABLE ROW LEVEL SECURITY;
ALTER TABLE customer_contacts    FORCE  ROW LEVEL SECURITY;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE customers            NO FORCE ROW LEVEL SECURITY;
ALTER TABLE customers            DISABLE ROW LEVEL SECURITY;
ALTER TABLE customer_counters    NO FORCE ROW LEVEL SECURITY;
ALTER TABLE customer_counters    DISABLE ROW LEVEL SECURITY;
ALTER TABLE customer_addresses   NO FORCE ROW LEVEL SECURITY;
ALTER TABLE customer_addresses   DISABLE ROW LEVEL SECURITY;
ALTER TABLE customer_contacts    NO FORCE ROW LEVEL SECURITY;
ALTER TABLE customer_contacts    DISABLE ROW LEVEL SECURITY;
-- +goose StatementEnd

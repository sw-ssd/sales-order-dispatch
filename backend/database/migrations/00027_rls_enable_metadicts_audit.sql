-- 字典與稽核啟用 RLS(03 計畫 D36、本計畫 Task 8)。
--
-- 本檔只做兩件事:對 `metadicts`／`audit_logs` 開門(ENABLE + FORCE),並修正
-- `core_audit_logs_scope` 的 WITH CHECK(見下「政策修正」)。policy 的定義/重建仍由
-- 00011(metadicts)、00023(補 WITH CHECK)、00025(NULLIF 正規化)擁有。
--
-- 前置條件(同一個 commit 序列內已完成,T8 Step 3):
--   `internal/services/metadict_service.go`(8 處)、`internal/services/audit_service.go`(2 處)、
--   `internal/handlers/auth_password.go`(改密碼／臨時密碼,2 處自開交易)已全數收斂到
--   `dbtenant.Client(ctx, s.db)` / `dbtenant.TxFrom(ctx)`;服務層與 handler 的任何查詢都落在
--   請求交易內(以 app_rw + 連線池上限 1 的探針斷言:自開交易會取不到連線而逾時)。
--
-- 啟用後的行為(呼叫端必須配合):
--   1. **未帶 scope 的讀取一律 fail-closed(0 列)**:`metadicts` 與 `audit_logs` 皆然。
--   2. **未帶 scope 的稽核寫入被 WITH CHECK 擋**(fail-closed):改密碼等路徑若自開交易寫稽核
--      會整筆失敗,故 Step 3 必須先落地。
--   3. `metadicts` 的系統預設列(`department_id IS NULL`)任何 scope 可讀,但只有 `scope=all`
--      可寫(00011 刻意讓 WITH CHECK 比 USING 嚴)。
--   4. `audit_logs` 的寫入只認「列所屬公司 = 當前公司」:任何已帶身分的 scope(company／
--      department／self)都能為自己公司寫稽核,跨公司寫入仍被擋。
--
-- 政策修正(本檔唯一的語意變更,已由 controller 核准):
--   `core_audit_logs_scope`(00023:92-107、00025 正規化版)的 USING/WITH CHECK 只有
--   `data_scope = 'all'` 與 `data_scope = 'company'` 兩個分支 —— 同一批 policy 中 `users`／
--   `departments` 都有 department/self 分支,`roles`／`role_permissions` 用 `scope <> ''`,
--   **audit_logs 是唯一的例外**。後果(實測,app_rw):`scope=department`／`self` 的請求寫稽核
--   一律 SQLSTATE 42501 → 與 D18(稽核與業務同一交易)直接衝突:凡是「已啟用域 + dept_admin/
--   staff/客戶」的寫入(改密碼、臨時密碼簽發、部門級主檔/商品/客戶的建立與修改)都會因稽核
--   寫入被擋而整筆失敗(T5–T7 的探針全部以 company scope 身分走,故此缺陷一路潛伏)。
--   本檔把 **USING 與 WITH CHECK 改成同一條件**:`'all'` 或「列公司 = 當前公司」
--   (未設 scope 時 `NULLIF` → NULL → 比較為 false,仍 fail-closed);**跨公司讀寫皆被擋**。
--   這也讓 WITH CHECK == USING,滿足本計畫 Global Constraints(「WITH CHECK 等於或嚴於 USING」)。
--
--   為何 USING 也必須改(實測):PostgreSQL 對 `INSERT ... RETURNING` 會**一併套用 SELECT policy
--   (即 USING)**來回傳新列,而 ent 的 `Create().Save()/Exec()` 一律是
--   `INSERT ... RETURNING id`(ent v0.14.6,dialect/sql/sqlgraph graph.go:1475;除非呼叫端自帶 id)。
--   於是「只改 WITH CHECK」在 ent 路徑下依然擋掉 dept/self 的稽核寫入(app_rw 實測:同一 scope、
--   同一列,`INSERT` 成功、`INSERT ... RETURNING id` → 42501)。替代方案(先 `SELECT nextval()`
--   再帶 id 寫入以避開 RETURNING)要把 policy 形狀的問題塞進應用層、寫死 sequence 名,且 rollback
--   會留 id 空洞,故不採。
--   「稽核只有 super/company_admin 可查」由服務層 ACL 承擔(AuditService:super/developer 全系統、
--   company_admin 僅自己公司、其餘角色 permission_denied),RLS 只需保證**跨租戶**不可見不可寫。
--   Down 逐字還原 00025 的 USING/WITH CHECK 形式(00023/00025 為歷史,不回寫)。
--
-- 冪等:`ENABLE/FORCE/NO FORCE/DISABLE ROW LEVEL SECURITY` 與 `DROP POLICY IF EXISTS` +
-- `CREATE POLICY` 皆可重複套用(比照 00008/00010/00019/00024/00025/00026 的冪等慣例)。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE metadicts   ENABLE ROW LEVEL SECURITY;
ALTER TABLE metadicts   FORCE  ROW LEVEL SECURITY;
ALTER TABLE audit_logs  ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs  FORCE  ROW LEVEL SECURITY;
-- +goose StatementEnd

-- +goose StatementBegin
DROP POLICY IF EXISTS core_audit_logs_scope ON audit_logs;
CREATE POLICY core_audit_logs_scope ON audit_logs FOR ALL
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (NULLIF(current_setting('app.current_company_id', true), ''))::bigint = company_id
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (NULLIF(current_setting('app.current_company_id', true), ''))::bigint = company_id
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE metadicts   NO FORCE ROW LEVEL SECURITY;
ALTER TABLE metadicts   DISABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs  NO FORCE ROW LEVEL SECURITY;
ALTER TABLE audit_logs  DISABLE ROW LEVEL SECURITY;
-- +goose StatementEnd

-- policy 還原:核心稽核回 00025 的 WITH CHECK 定義(逐字)。
-- +goose StatementBegin
DROP POLICY IF EXISTS core_audit_logs_scope ON audit_logs;
CREATE POLICY core_audit_logs_scope ON audit_logs FOR ALL
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
            AND (NULLIF(current_setting('app.current_company_id', true), ''))::bigint = company_id
        )
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
            AND (NULLIF(current_setting('app.current_company_id', true), ''))::bigint = company_id
        )
    );
-- +goose StatementEnd

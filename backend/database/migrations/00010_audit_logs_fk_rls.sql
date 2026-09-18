-- 前向修補(審查 I7):audit_logs 補 FK 約束與 RLS policy。
-- 00009 建表時未含 FK;03 計畫 2.6.1 明列 company_id / department_id / user_id 為 FK。
-- 以 DO $$ 冪等對「既有 DB」補約束(全新 DB 由 00009 + 本檔一併成立),不影響 goose 版本表。
-- RLS 比照 00007 模式:僅「定義」policy,不 ENABLE/FORCE(services 尚無每請求交易套用點)。
-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'audit_logs_company_id_fkey') THEN
        ALTER TABLE audit_logs
            ADD CONSTRAINT audit_logs_company_id_fkey
            FOREIGN KEY (company_id) REFERENCES companies(id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'audit_logs_department_id_fkey') THEN
        ALTER TABLE audit_logs
            ADD CONSTRAINT audit_logs_department_id_fkey
            FOREIGN KEY (department_id) REFERENCES departments(id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'audit_logs_user_id_fkey') THEN
        ALTER TABLE audit_logs
            ADD CONSTRAINT audit_logs_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES users(id);
    END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DROP POLICY IF EXISTS core_audit_logs_scope ON audit_logs;
CREATE POLICY core_audit_logs_scope ON audit_logs
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
            AND (current_setting('app.current_company_id', true))::bigint = company_id
        )
    );
-- +goose StatementEnd

-- 注意:同 00007,本波僅「定義」policy 不 ENABLE/FORCE RLS;否則 app.current_* 未設定時
-- audit_logs 將全量不可見。待每請求交易層落定 SET LOCAL app.* 後再以另一次 migration 啟用。

-- +goose Down
-- +goose StatementBegin
DROP POLICY IF EXISTS core_audit_logs_scope ON audit_logs;
ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS audit_logs_company_id_fkey;
ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS audit_logs_department_id_fkey;
ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS audit_logs_user_id_fkey;
-- +goose StatementEnd

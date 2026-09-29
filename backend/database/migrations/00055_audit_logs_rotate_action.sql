-- 擴展 audit_logs.action 列舉:新增 'rotate_external_id'。
-- 背景:公司第三方整合暴露值(external_id)輪換 RPC 需寫一筆業務稽核,
-- 動作語意等同 update(記錄前後快照),但獨立成動作便於稽核檢視與過濾。
-- action 是 text + CHECK 約束(00009 建立,未顯式命名 → PG 自動命名 audit_logs_action_check)。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE audit_logs
    DROP CONSTRAINT IF EXISTS audit_logs_action_check;
ALTER TABLE audit_logs
    ADD CONSTRAINT audit_logs_action_check
    CHECK (action IN (
        'create', 'update', 'delete', 'login', 'logout', 'print',
        'force_logout', 'role_change', 'dispatch_cancel', 'void',
        'rotate_external_id'
    ));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE audit_logs
    DROP CONSTRAINT IF EXISTS audit_logs_action_check;
ALTER TABLE audit_logs
    ADD CONSTRAINT audit_logs_action_check
    CHECK (action IN (
        'create', 'update', 'delete', 'login', 'logout', 'print',
        'force_logout', 'role_change', 'dispatch_cancel', 'void'
    ));
-- +goose StatementEnd

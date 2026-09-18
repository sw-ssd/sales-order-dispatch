-- 稽核日誌表（03 計畫 2.6.1, D18）。不可軟刪除、不可由 API 修改。
-- 寫入一律發生於業務交易內（同一 DB 交易，D18）。只建必要索引，不建 deleted_at/updated_at。
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS audit_logs (
    id              bigserial PRIMARY KEY,
    company_id      bigint NOT NULL,
    department_id   bigint,
    user_id         bigint NOT NULL,
    action          text NOT NULL CHECK (action IN ('create','update','delete','login','logout','print','force_logout','role_change','dispatch_cancel','void')),
    resource_type   text NOT NULL,
    resource_id     text,
    before_snapshot jsonb,
    after_snapshot  jsonb,
    ip_address      text,
    user_agent      text,
    created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS audit_logs_company_created_at_idx  ON audit_logs (company_id, created_at);
CREATE INDEX IF NOT EXISTS audit_logs_resource_idx            ON audit_logs (resource_type, resource_id);
CREATE INDEX IF NOT EXISTS audit_logs_user_created_at_idx     ON audit_logs (user_id, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS audit_logs;
-- +goose StatementEnd

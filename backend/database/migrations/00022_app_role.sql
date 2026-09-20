-- 業務連線角色(D36):非 owner、NOBYPASSRLS，使 RLS 真正生效。
-- 密碼不由 migration 設定(不落版控):本機開發用 `task db:app-password`，
-- 其他環境由部署流程以 ALTER ROLE app_rw PASSWORD ... 提供。
-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_rw') THEN
        CREATE ROLE app_rw LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOBYPASSRLS;
    END IF;
END
$$;
-- +goose StatementEnd

GRANT USAGE ON SCHEMA public TO app_rw;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_rw;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_rw;
-- 之後新增的表/序列自動授權（migration 以 owner 執行，故 default privileges 掛在 owner 上）
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_rw;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO app_rw;

-- +goose Down
-- 必須先逐一撤銷授權再 DROP ROLE:PG 會以既有 ACL 依賴擋下 DROP ROLE
-- (ERROR: role "app_rw" cannot be dropped because some objects depend on it, SQLSTATE 2BP01)。
-- +goose StatementBegin
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    REVOKE SELECT, INSERT, UPDATE, DELETE ON TABLES FROM app_rw;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    REVOKE USAGE, SELECT ON SEQUENCES FROM app_rw;
-- +goose StatementEnd

-- +goose StatementBegin
REVOKE SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public FROM app_rw;
-- +goose StatementEnd

-- +goose StatementBegin
REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public FROM app_rw;
-- +goose StatementEnd

-- +goose StatementBegin
REVOKE USAGE ON SCHEMA public FROM app_rw;
-- +goose StatementEnd

-- +goose StatementBegin
DROP ROLE IF EXISTS app_rw;
-- +goose StatementEnd

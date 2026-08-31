-- 啟用資料庫層級 Row-Level Security,並建立應用程式讀/寫角色。
-- +goose Up
-- +goose StatementBegin
ALTER DATABASE salesorder SET row_security = on;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'app_read') THEN
        CREATE ROLE app_read NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'app_write') THEN
        CREATE ROLE app_write NOLOGIN;
    END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP ROLE IF EXISTS app_write;
DROP ROLE IF EXISTS app_read;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER DATABASE salesorder RESET row_security;
-- +goose StatementEnd

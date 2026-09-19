-- 啟用資料庫層級 Row-Level Security,並建立應用程式讀/寫角色。
-- ALTER DATABASE 以 current_database() 定址(不硬編庫名):同一份 schema 可在任意庫名上套用,
-- 測試庫與自訂部署庫名皆不受限。ALTER DATABASE ... SET 可於交易內執行,故無須 NO TRANSACTION。
-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    EXECUTE format('ALTER DATABASE %I SET row_security = on', current_database());
END
$$;
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
DO $$
BEGIN
    EXECUTE format('ALTER DATABASE %I RESET row_security', current_database());
END
$$;
-- +goose StatementEnd

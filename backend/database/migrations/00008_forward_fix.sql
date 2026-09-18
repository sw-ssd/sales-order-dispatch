-- 向前修補既有 DB 的 migration 缺口(D32 收尾)。
-- 背景:本波將舊 00003(role_permissions 建表,無 FK)/00004(users.token_version)刪併入
-- 00005_core_schema.sql。全新 DB 由 00005 直接建成含 FK 與 token_version 的完整結構;
-- 但「已跑過舊 00001~00004」的既有 DB,00005 以 CREATE TABLE IF NOT EXISTS / ADD COLUMN
-- IF NOT EXISTS 對既有表為 no-op → role_permissions.role_id 缺 FK、users 缺 token_version。
-- 本 migration 以 DO block 冪等補齊,對全新 DB 為 no-op,不影響 goose 版本表。

-- +goose Up
-- +goose StatementBegin
-- role_permissions.role_id → roles(id) FK(既有 DB 由舊 00003 建表時為裸 bigint 無 FK)。
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'role_permissions_role_id_fkey'
          AND conrelid = 'role_permissions'::regclass
    ) THEN
        ALTER TABLE role_permissions
            ADD CONSTRAINT role_permissions_role_id_fkey
            FOREIGN KEY (role_id) REFERENCES roles(id);
    END IF;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
-- users.token_version(既有 DB 由舊 00004 已加;缺則補,與 00005 同形)。
ALTER TABLE users ADD COLUMN IF NOT EXISTS token_version integer NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE role_permissions DROP CONSTRAINT IF EXISTS role_permissions_role_id_fkey;
ALTER TABLE users DROP COLUMN IF EXISTS token_version;
-- +goose StatementEnd

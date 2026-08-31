-- users.token_version(D5):token 撤銷版本號。改密碼 / 停用 / 角色變更 / 強制登出時遞增,
-- 使該使用者既有 access / refresh token 全數失效。ent schema 已含此欄位,此 migration
-- 對齊既有資料庫(00003 role_permissions 先例:以 IF NOT EXISTS 保證冪等)。
-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS token_version integer NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS token_version;

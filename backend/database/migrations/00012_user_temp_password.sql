-- A3 臨時密碼與首登強制修改(01-auth 1.5.2/1.5.4):users 補 must_change_password + 臨時密碼效期。
-- 冪等:ADD COLUMN IF NOT EXISTS,可重複套用於既有 DB;全新 DB 亦由本檔成立。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN IF NOT EXISTS must_change_password boolean NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS temp_password_expires_at timestamptz;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN IF EXISTS temp_password_expires_at;
ALTER TABLE users DROP COLUMN IF EXISTS must_change_password;
-- +goose StatementEnd

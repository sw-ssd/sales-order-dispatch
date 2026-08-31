-- 初始 schema:goose 會自動建立 goose_db_version 版本表;
-- 此檔僅啟用後續 schema(ent,Task 9)所需的 pgcrypto(gen_random_uuid)。
-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP EXTENSION IF EXISTS pgcrypto;
-- +goose StatementEnd

-- 04 客戶 D22(3.1.4):users 補 customer_id / is_primary / system_generated,承載「建檔連動」的
-- 主帳號與業務子帳號。is_primary=true 且 customer_id IS NOT NULL 的部分唯一索引保證「每客戶恰一主帳號」;
-- system_generated 供店家帳號管理清單灰化判斷(對系統自動帳號改名/停用/重置由 01/02 層拒絕)。
-- 冪等:ADD COLUMN IF NOT EXISTS;全新 DB 亦由本檔成立。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN IF NOT EXISTS customer_id bigint;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_primary boolean NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS system_generated boolean NOT NULL DEFAULT false;
-- 查詢索引:依客戶列業務/主帳號清單(LIST 帳號用)。
CREATE INDEX IF NOT EXISTS users_customer_idx ON users (customer_id);
-- 每客戶恰一主帳號(部分唯一索引;employee 等非客戶帳號 customer_id 為 NULL 不受影響)。
CREATE UNIQUE INDEX IF NOT EXISTS users_customer_primary_unique
    ON users (customer_id) WHERE is_primary = true AND customer_id IS NOT NULL;
-- 指向 customers(比照 audit_logs 補 FK 慣例)。
ALTER TABLE users ADD CONSTRAINT users_customer_fk FOREIGN KEY (customer_id) REFERENCES customers(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_customer_fk;
DROP INDEX IF EXISTS users_customer_primary_unique;
DROP INDEX IF EXISTS users_customer_idx;
ALTER TABLE users DROP COLUMN IF EXISTS system_generated;
ALTER TABLE users DROP COLUMN IF EXISTS is_primary;
ALTER TABLE users DROP COLUMN IF EXISTS customer_id;
-- +goose StatementEnd

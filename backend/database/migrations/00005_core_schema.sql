-- 核心表 DDL(對齊 ent/schema 與 ent/migrate/schema.go)。
-- 本檔一次建立 companies / departments / roles / role_permissions / users,
-- 取代先前分段且順序斷裂的 00003(role_permissions 無 FK)與 00004(users 未建即 ALTER)。
-- 所有業務實體以 id bigserial 主鍵,欄位型別/唯一性/外鍵對齊 ent migrate schema。
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS companies (
    id           bigserial PRIMARY KEY,
    name         text NOT NULL,
    tax_id       text,
    status       text NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive','suspended')),
    identifier   text NOT NULL UNIQUE,
    public_info  jsonb,
    capabilities jsonb,
    logo_url     text
);

CREATE TABLE IF NOT EXISTS departments (
    id                 bigserial PRIMARY KEY,
    name               text NOT NULL,
    company_departments bigint NOT NULL REFERENCES companies(id)
);

CREATE TABLE IF NOT EXISTS roles (
    id         bigserial PRIMARY KEY,
    code       text NOT NULL,
    name       text NOT NULL,
    data_scope text NOT NULL DEFAULT 'company' CHECK (data_scope IN ('all','company','department','self')),
    is_system  boolean NOT NULL DEFAULT false,
    is_active  boolean NOT NULL DEFAULT true,
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS role_permissions (
    id         bigserial PRIMARY KEY,
    role_id    bigint NOT NULL REFERENCES roles(id),
    resource   text NOT NULL,
    action     text NOT NULL,
    conditions jsonb,
    inverted   boolean NOT NULL DEFAULT false,
    sort_order integer NOT NULL DEFAULT 0
);
-- 權限列去重:conditions 為 NULL 時視為 '',讓 NULL 條件也能被抓重
-- (表層 UNIQUE 不接受表達式,須以唯一索引表達;md5(jsonb::text) 為 IMMUTABLE 可入索引)。
CREATE UNIQUE INDEX IF NOT EXISTS role_permissions_unique_idx
    ON role_permissions (role_id, resource, action, COALESCE(md5(conditions::text), ''));

CREATE TABLE IF NOT EXISTS users (
    id              bigserial PRIMARY KEY,
    email           text NOT NULL UNIQUE,
    name            text NOT NULL,
    status          text NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive','pending')),
    role            text NOT NULL,
    phone           text,
    employee_no     text,
    is_customer     boolean NOT NULL DEFAULT false,
    account_name    text,
    token_version   integer NOT NULL DEFAULT 0,
    password_hash   text NOT NULL,
    company_users   bigint NOT NULL REFERENCES companies(id),
    department_users bigint REFERENCES departments(id)
);
-- +goose StatementEnd

-- +goose Down
-- 回滾語意(三段,順序不可換):
--   1. 先移除「後續 migration 替本檔 5 張表加上的外部 FK」:audit_logs→users/companies/departments
--      (00010)、customers→companies/departments/users(00013),以及日後任何指向這 5 張表的 FK。
--      goose 正常會先跑那些 migration 的 Down(表連 FK 一起消失),故正常路徑下這一段是 no-op;
--      但版本表若缺該版本列(既有/手改庫的版本表漂移,例如手動 INSERT/DELETE 過版本列),
--      那些表會存活到這裡,並讓下面的 DROP TABLE 以 SQLSTATE 2BP01 失敗:
--      "cannot drop table users because other objects depend on it / constraint
--       audit_logs_user_id_fkey on table audit_logs … customers_sales_rep_fk on table customers"。
--      本檔擁有這 5 張表,故自行釋放指向它們的外部 FK,讓本檔的 Down 不依賴別人先跑;
--      反向(重新 up)時那些 migration 的 Up 會把 FK 補回(00010 為冪等 DO block)。
--   2. 再依序移除本檔建立的物件:role_permissions 須先於 roles、users 先於 departments/companies
--      (references 的 FK 由被 DROP 的表自己帶走)。
--   3. 最後撤銷被本檔合併的 00003/00004 版本紀錄:舊 00003(role_permissions 建表,無 FK)與
--      舊 00004(users.token_version)的內容已併入本檔,回滾本檔即回滾其內容;但兩檔案已刪,
--      goose 無法在版本表仍留著 3/4 的列時繼續往下回滾(會卡在
--      "migration file not found for current version (4)"),故一併清掉這些列(無表/無列皆為 no-op)。
-- +goose StatementBegin
DO $$
DECLARE
    owned oid[];
    dep   record;
BEGIN
    SELECT array_agg(c.oid) INTO owned
    FROM pg_class c
    WHERE c.relkind = 'r'
      AND c.relname IN ('users', 'role_permissions', 'roles', 'departments', 'companies');
    IF owned IS NULL THEN
        RETURN; -- 本檔的表不存在(非本檔的下行狀態),無事可做。
    END IF;
    FOR dep IN
        SELECT con.conrelid::regclass AS tbl, con.conname AS name
        FROM pg_constraint con
        WHERE con.contype = 'f'
          AND con.confrelid = ANY (owned)
          AND NOT (con.conrelid = ANY (owned))
    LOOP
        EXECUTE format('ALTER TABLE %s DROP CONSTRAINT %I', dep.tbl, dep.name);
    END LOOP;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS companies;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    -- goose 版本表名為預設值(salesorder 全庫唯一一組業務版本;OpenFGA 另有 openfga_goose_db_version)
    -- 且可能不存在(no-versioning 情境),故以 to_regclass 守衛。
    IF to_regclass('goose_db_version') IS NOT NULL THEN
        DELETE FROM goose_db_version WHERE version_id IN (3, 4);
    END IF;
END
$$;
-- +goose StatementEnd

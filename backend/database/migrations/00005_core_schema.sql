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
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS companies;
-- +goose StatementEnd

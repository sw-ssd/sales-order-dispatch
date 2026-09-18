-- RLS 政策落地核心表(D32):users / companies / departments 依 app.current_data_scope
-- 做資料範圍隔離(all 繞過;company 限同公司;department 限同部門;self 限本人)。
-- roles / role_permissions 為共享目錄(無租戶欄位)且由 role_service 以 OpenFGA/條件邏輯
-- 控管寫入,採「已驗證身分可讀」許容政策,不作細粒度 WHERE 限制。
-- middleware 於交易內以 SET LOCAL app.* 切換身分與範圍(見 internal/auth/rls.go)。
-- +goose Up

-- 身分/範圍變數預設空字串(current_setting(..., true) 未設定時回 NULL)。
-- 以 CURRENT SETTING + COALESCE 讓各 policy 在「未設定」時 fail-closed(不洩漏)。

-- companies:all → 全見;company → 僅自己公司;departmen/self → 亦依 current_company_id(顯示所屬公司)。
CREATE POLICY core_companies_scope ON companies
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (current_setting('app.current_company_id', true))::bigint = id
    );

-- departments:all → 全見;company → 同公司部門;department → 僅自己部門;self → 不洩漏(無部門)。
CREATE POLICY core_departments_scope ON departments
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
            AND (current_setting('app.current_company_id', true))::bigint = company_departments
        )
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'department'
            AND (current_setting('app.current_department_id', true))::bigint = id
        )
    );

-- users:all → 全見;company → 同公司;department → 同部門;self → 僅本人。
CREATE POLICY core_users_scope ON users
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
            AND (current_setting('app.current_company_id', true))::bigint = company_users
        )
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'department'
            AND (current_setting('app.current_department_id', true))::bigint = department_users
        )
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'self'
            AND (current_setting('app.current_user_id', true))::bigint = id
        )
    );

-- roles / role_permissions:共享目錄,已設定身分(任一範圍)即可讀;無租戶欄位故不細分。
CREATE POLICY core_roles_read ON roles
    USING (COALESCE(current_setting('app.current_data_scope', true), '') <> '');
CREATE POLICY core_role_permissions_read ON role_permissions
    USING (COALESCE(current_setting('app.current_data_scope', true), '') <> '');

-- 對核心表強制啟用 RLS(force 使表 owner 亦受限制)。
ALTER TABLE companies      ENABLE ROW LEVEL SECURITY;
ALTER TABLE companies      FORCE ROW LEVEL SECURITY;
ALTER TABLE departments    ENABLE ROW LEVEL SECURITY;
ALTER TABLE departments    FORCE ROW LEVEL SECURITY;
ALTER TABLE users          ENABLE ROW LEVEL SECURITY;
ALTER TABLE users          FORCE ROW LEVEL SECURITY;
ALTER TABLE roles          ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles          FORCE ROW LEVEL SECURITY;
ALTER TABLE role_permissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE role_permissions FORCE ROW LEVEL SECURITY;

-- +goose Down
-- +goose StatementBegin
ALTER TABLE companies      NO FORCE ROW LEVEL SECURITY;
ALTER TABLE companies      DISABLE ROW LEVEL SECURITY;
ALTER TABLE departments    NO FORCE ROW LEVEL SECURITY;
ALTER TABLE departments    DISABLE ROW LEVEL SECURITY;
ALTER TABLE users          NO FORCE ROW LEVEL SECURITY;
ALTER TABLE users          DISABLE ROW LEVEL SECURITY;
ALTER TABLE roles          NO FORCE ROW LEVEL SECURITY;
ALTER TABLE roles          DISABLE ROW LEVEL SECURITY;
ALTER TABLE role_permissions NO FORCE ROW LEVEL SECURITY;
ALTER TABLE role_permissions DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS core_companies_scope ON companies;
DROP POLICY IF EXISTS core_departments_scope ON departments;
DROP POLICY IF EXISTS core_users_scope ON users;
DROP POLICY IF EXISTS core_roles_read ON roles;
DROP POLICY IF EXISTS core_role_permissions_read ON role_permissions;
-- +goose StatementEnd

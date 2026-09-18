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

-- 注意:本波僅「定義」RLS policy,並不 ENABLE/FORCE RLS。
-- services 目前以 ent client 直接查詢(無每請求交易的 ApplyRLS 套用點);
-- 若此刻 FORCE RLS 而 app.current_* GUC 未設定,會使核心表全量不可見(黑屏)。
-- 待 repository/每請求交易層落定 SET LOCAL app.* 後,再以另一次 migration 啟用 RLS。

-- +goose Down
-- +goose StatementBegin
DROP POLICY IF EXISTS core_companies_scope ON companies;
DROP POLICY IF EXISTS core_departments_scope ON departments;
DROP POLICY IF EXISTS core_users_scope ON users;
DROP POLICY IF EXISTS core_roles_read ON roles;
DROP POLICY IF EXISTS core_role_permissions_read ON role_permissions;
-- +goose StatementEnd

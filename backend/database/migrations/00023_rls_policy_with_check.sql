-- 補齊 RLS 的寫入面：14 個既有 policy 原本只有 USING(D35/D36)。
-- USING 管「可見列」，WITH CHECK 管「可寫入的新列」；缺 WITH CHECK 時 PostgreSQL
-- 以 USING 代替檢查，語意上仍是「只檢查新列」，但表達不完整且審計時看不出意圖。
-- 本波一律明示 WITH CHECK，且**等於原 USING**（不放寬任何既有語意）
-- 例外：core_metadicts_scope 已有 WITH CHECK 且刻意較嚴，不動。
-- +goose Up
-- +goose StatementBegin

-- 核心三表（原 00007）
DROP POLICY IF EXISTS core_companies_scope ON companies;
CREATE POLICY core_companies_scope ON companies FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (current_setting('app.current_company_id', true))::bigint = id
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (current_setting('app.current_company_id', true))::bigint = id
    );

DROP POLICY IF EXISTS core_departments_scope ON departments;
CREATE POLICY core_departments_scope ON departments FOR ALL
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
    )
    WITH CHECK (
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

DROP POLICY IF EXISTS core_users_scope ON users;
CREATE POLICY core_users_scope ON users FOR ALL
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
    )
    WITH CHECK (
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

-- 共享目錄：語意維持「已設定身分即可讀寫」不變（角色/權限的寫入權威仍在服務層 ACL，
-- 本波不在此處收緊，避免與 role_service 的授權矩陣不一致）。
DROP POLICY IF EXISTS core_roles_read ON roles;
CREATE POLICY core_roles_read ON roles FOR ALL
    USING (COALESCE(current_setting('app.current_data_scope', true), '') <> '')
    WITH CHECK (COALESCE(current_setting('app.current_data_scope', true), '') <> '');

DROP POLICY IF EXISTS core_role_permissions_read ON role_permissions;
CREATE POLICY core_role_permissions_read ON role_permissions FOR ALL
    USING (COALESCE(current_setting('app.current_data_scope', true), '') <> '')
    WITH CHECK (COALESCE(current_setting('app.current_data_scope', true), '') <> '');

-- 稽核（原 00010）
DROP POLICY IF EXISTS core_audit_logs_scope ON audit_logs;
CREATE POLICY core_audit_logs_scope ON audit_logs FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
            AND (current_setting('app.current_company_id', true))::bigint = company_id
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
            AND (current_setting('app.current_company_id', true))::bigint = company_id
        )
    );

-- 客戶域（原 00013／00015）：公司 + 部門（department_id IS NULL 為公司層共用列）
DROP POLICY IF EXISTS core_customers_scope ON customers;
CREATE POLICY core_customers_scope ON customers FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

DROP POLICY IF EXISTS core_customer_addresses_scope ON customer_addresses;
CREATE POLICY core_customer_addresses_scope ON customer_addresses FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

DROP POLICY IF EXISTS core_customer_contacts_scope ON customer_contacts;
CREATE POLICY core_customer_contacts_scope ON customer_contacts FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

-- 部門級主檔（原 00016）：四張同型
DROP POLICY IF EXISTS core_warehouses_scope ON warehouses;
CREATE POLICY core_warehouses_scope ON warehouses FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

DROP POLICY IF EXISTS core_routes_scope ON routes;
CREATE POLICY core_routes_scope ON routes FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

DROP POLICY IF EXISTS core_processing_specs_scope ON processing_specs;
CREATE POLICY core_processing_specs_scope ON processing_specs FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

DROP POLICY IF EXISTS core_product_categories_scope ON product_categories;
CREATE POLICY core_product_categories_scope ON product_categories FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

-- 商品（原 00017）
DROP POLICY IF EXISTS core_products_scope ON products;
CREATE POLICY core_products_scope ON products FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

-- 漏網三表：原先完全沒有 policy
-- customer_counters 的 company_id 即主鍵
DROP POLICY IF EXISTS core_customer_counters_scope ON customer_counters;
CREATE POLICY core_customer_counters_scope ON customer_counters FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR company_id = (current_setting('app.current_company_id', true))::bigint
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR company_id = (current_setting('app.current_company_id', true))::bigint
    );

-- product_units / product_processing_specs 無 company_id：以父表 products 的存在性表達。
-- 子查詢本身也受 products 的 policy 約束（同一角色），故隔離具傳遞性。
DROP POLICY IF EXISTS core_product_units_scope ON product_units;
CREATE POLICY core_product_units_scope ON product_units FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR EXISTS (SELECT 1 FROM products p WHERE p.id = product_units.product_id)
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR EXISTS (SELECT 1 FROM products p WHERE p.id = product_units.product_id)
    );

DROP POLICY IF EXISTS core_product_processing_specs_scope ON product_processing_specs;
CREATE POLICY core_product_processing_specs_scope ON product_processing_specs FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR EXISTS (SELECT 1 FROM products p WHERE p.id = product_processing_specs.product_id)
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR EXISTS (SELECT 1 FROM products p WHERE p.id = product_processing_specs.product_id)
    );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- 還原為「只有 USING」的原狀，並移除本波新增的三個 policy
DROP POLICY IF EXISTS core_customer_counters_scope ON customer_counters;
DROP POLICY IF EXISTS core_product_units_scope ON product_units;
DROP POLICY IF EXISTS core_product_processing_specs_scope ON product_processing_specs;
DROP POLICY IF EXISTS core_companies_scope ON companies;
CREATE POLICY core_companies_scope ON companies
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (current_setting('app.current_company_id', true))::bigint = id
    );
DROP POLICY IF EXISTS core_departments_scope ON departments;
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
DROP POLICY IF EXISTS core_users_scope ON users;
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
DROP POLICY IF EXISTS core_roles_read ON roles;
CREATE POLICY core_roles_read ON roles
    USING (COALESCE(current_setting('app.current_data_scope', true), '') <> '');
DROP POLICY IF EXISTS core_role_permissions_read ON role_permissions;
CREATE POLICY core_role_permissions_read ON role_permissions
    USING (COALESCE(current_setting('app.current_data_scope', true), '') <> '');
DROP POLICY IF EXISTS core_audit_logs_scope ON audit_logs;
CREATE POLICY core_audit_logs_scope ON audit_logs
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
            AND (current_setting('app.current_company_id', true))::bigint = company_id
        )
    );
DROP POLICY IF EXISTS core_customers_scope ON customers;
CREATE POLICY core_customers_scope ON customers
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );
DROP POLICY IF EXISTS core_customer_addresses_scope ON customer_addresses;
CREATE POLICY core_customer_addresses_scope ON customer_addresses
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );
DROP POLICY IF EXISTS core_customer_contacts_scope ON customer_contacts;
CREATE POLICY core_customer_contacts_scope ON customer_contacts
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );
DROP POLICY IF EXISTS core_warehouses_scope ON warehouses;
CREATE POLICY core_warehouses_scope ON warehouses
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );
DROP POLICY IF EXISTS core_routes_scope ON routes;
CREATE POLICY core_routes_scope ON routes
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );
DROP POLICY IF EXISTS core_processing_specs_scope ON processing_specs;
CREATE POLICY core_processing_specs_scope ON processing_specs
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );
DROP POLICY IF EXISTS core_product_categories_scope ON product_categories;
CREATE POLICY core_product_categories_scope ON product_categories
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );
DROP POLICY IF EXISTS core_products_scope ON products;
CREATE POLICY core_products_scope ON products
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );
-- +goose StatementEnd

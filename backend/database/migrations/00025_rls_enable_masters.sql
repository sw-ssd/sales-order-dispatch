-- 部門級主檔啟用 RLS(D36,第三批:warehouses/routes/processing_specs/product_categories),
-- 並把 00023/00011 建立的**全部 18 個 policy** 的取值正規化為 NULLIF 形式。
--
-- 為何同一個 migration 要做兩件事(先正規化、再 ENABLE):
--   `SET LOCAL app.current_*` 對**自訂 GUC** 會在 session 層留下空字串 placeholder —— 交易結束後
--   該連線上的值回到 `''`(而非 unset)。於是同一條池化連線在「後續未設 scope」的查詢上,policy 裡的
--   `''::bigint` 直接 22P02 報錯,而不是回 0 列。兩者都 fail-closed(無資料外洩),但錯誤語意與
--   「查不到」混淆,且會讓下游把「未登入/未帶 scope」誤判成系統故障(T5 實測:同一池先服務過有 scope
--   的請求後,無 scope 的查詢以 22P02 失敗)。以 `NULLIF(current_setting(..., true), '')` 把 `''`
--   正規化成 NULL,語意回到「未設 scope → 0 列」。00024 已 ENABLE 的客戶域四表同樣受影響,故本波
--   一次改完 18 個 policy,而不是只改本波 ENABLE 的 4 張表。
--
-- 作法與範圍:
--   1. 18 個 policy 逐表 `DROP POLICY IF EXISTS` 後重建,唯一差異是把每個 `current_setting(...)`
--      包成 `NULLIF(current_setting(...), '')`;USING/WITH CHECK 的其餘條件**逐位元組不變**,
--      不放寬任何既有語意。`core_metadicts_scope`(第 18 個,定義於 00011)的 WITH CHECK 仍刻意
--      不含 `department_id IS NULL`,此處只加 NULLIF,不收緊也不放寬。
--   2. 本波 ENABLE + FORCE 的目標表:部門級主檔四表(葉節點,無子表)。
--
-- 啟用後的行為(呼叫端必須配合,已於同一波完成):
--   1. **未帶 scope 的查詢一律 fail-closed(0 列)** → 服務層任何查詢/寫入都必須走
--      `dbtenant.Client(ctx, s.db)` 或請求交易(`dbtenant.TxFrom`);漏掛不是權限錯誤,是「查不到資料」
--      或被 WITH CHECK 擋下。
--   2. FORCE 讓 **table owner** 也受 policy 約束(PG 的 superuser 仍永遠繞過 RLS,FORCE 亦然;
--      生產的 owner 角色並非 superuser)。日後對本域回填資料的 migration/seed 必須先
--      `SET LOCAL app.current_data_scope = 'all'`(dbtenant.SystemScopeTx),否則寫入會被 WITH CHECK 擋下。
--
-- Up/Down 對稱:Up 先重建 policy 再 ENABLE+FORCE;Down 先 NO FORCE + DISABLE 再把 18 個 policy
-- 逐字還原為 00023 的 Up 定義(metadicts 還原為 00011 的定義),使回退後與 00024 之後的狀態一致。
--
-- 冪等:`DROP POLICY IF EXISTS` + `CREATE POLICY` 與 `ALTER TABLE ... ENABLE/FORCE/DISABLE ROW LEVEL
-- SECURITY` 皆可重複套用(比照 00008/00010/00019/00024 的冪等慣例)。
-- +goose Up
-- +goose StatementBegin
-- 核心三表（原 00007）
DROP POLICY IF EXISTS core_companies_scope ON companies;
CREATE POLICY core_companies_scope ON companies FOR ALL
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (NULLIF(current_setting('app.current_company_id', true), ''))::bigint = id
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (NULLIF(current_setting('app.current_company_id', true), ''))::bigint = id
    );

DROP POLICY IF EXISTS core_departments_scope ON departments;
CREATE POLICY core_departments_scope ON departments FOR ALL
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
            AND (NULLIF(current_setting('app.current_company_id', true), ''))::bigint = company_departments
        )
        OR (
            COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'department'
            AND (NULLIF(current_setting('app.current_department_id', true), ''))::bigint = id
        )
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
            AND (NULLIF(current_setting('app.current_company_id', true), ''))::bigint = company_departments
        )
        OR (
            COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'department'
            AND (NULLIF(current_setting('app.current_department_id', true), ''))::bigint = id
        )
    );

DROP POLICY IF EXISTS core_users_scope ON users;
CREATE POLICY core_users_scope ON users FOR ALL
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
            AND (NULLIF(current_setting('app.current_company_id', true), ''))::bigint = company_users
        )
        OR (
            COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'department'
            AND (NULLIF(current_setting('app.current_department_id', true), ''))::bigint = department_users
        )
        OR (
            COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'self'
            AND (NULLIF(current_setting('app.current_user_id', true), ''))::bigint = id
        )
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
            AND (NULLIF(current_setting('app.current_company_id', true), ''))::bigint = company_users
        )
        OR (
            COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'department'
            AND (NULLIF(current_setting('app.current_department_id', true), ''))::bigint = department_users
        )
        OR (
            COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'self'
            AND (NULLIF(current_setting('app.current_user_id', true), ''))::bigint = id
        )
    );

-- 共享目錄：語意維持「已設定身分即可讀寫」不變（角色/權限的寫入權威仍在服務層 ACL，
-- 本波不在此處收緊，避免與 role_service 的授權矩陣不一致）。
DROP POLICY IF EXISTS core_roles_read ON roles;
CREATE POLICY core_roles_read ON roles FOR ALL
    USING (COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') <> '')
    WITH CHECK (COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') <> '');

DROP POLICY IF EXISTS core_role_permissions_read ON role_permissions;
CREATE POLICY core_role_permissions_read ON role_permissions FOR ALL
    USING (COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') <> '')
    WITH CHECK (COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') <> '');

-- 稽核（原 00010）
DROP POLICY IF EXISTS core_audit_logs_scope ON audit_logs;
CREATE POLICY core_audit_logs_scope ON audit_logs FOR ALL
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
            AND (NULLIF(current_setting('app.current_company_id', true), ''))::bigint = company_id
        )
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
            AND (NULLIF(current_setting('app.current_company_id', true), ''))::bigint = company_id
        )
    );

-- 客戶域（原 00013／00015）：公司 + 部門（department_id IS NULL 為公司層共用列）
DROP POLICY IF EXISTS core_customers_scope ON customers;
CREATE POLICY core_customers_scope ON customers FOR ALL
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
            AND (
                COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
                OR department_id IS NULL
                OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
            AND (
                COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
                OR department_id IS NULL
                OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
            )
        )
    );

DROP POLICY IF EXISTS core_customer_addresses_scope ON customer_addresses;
CREATE POLICY core_customer_addresses_scope ON customer_addresses FOR ALL
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
            AND (
                COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
                OR department_id IS NULL
                OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
            AND (
                COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
                OR department_id IS NULL
                OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
            )
        )
    );

DROP POLICY IF EXISTS core_customer_contacts_scope ON customer_contacts;
CREATE POLICY core_customer_contacts_scope ON customer_contacts FOR ALL
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
            AND (
                COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
                OR department_id IS NULL
                OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
            AND (
                COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
                OR department_id IS NULL
                OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
            )
        )
    );

-- 部門級主檔（原 00016）：四張同型
DROP POLICY IF EXISTS core_warehouses_scope ON warehouses;
CREATE POLICY core_warehouses_scope ON warehouses FOR ALL
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
            AND (
                COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
                OR department_id IS NULL
                OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
            AND (
                COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
                OR department_id IS NULL
                OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
            )
        )
    );

DROP POLICY IF EXISTS core_routes_scope ON routes;
CREATE POLICY core_routes_scope ON routes FOR ALL
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
            AND (
                COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
                OR department_id IS NULL
                OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
            AND (
                COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
                OR department_id IS NULL
                OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
            )
        )
    );

DROP POLICY IF EXISTS core_processing_specs_scope ON processing_specs;
CREATE POLICY core_processing_specs_scope ON processing_specs FOR ALL
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
            AND (
                COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
                OR department_id IS NULL
                OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
            AND (
                COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
                OR department_id IS NULL
                OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
            )
        )
    );

DROP POLICY IF EXISTS core_product_categories_scope ON product_categories;
CREATE POLICY core_product_categories_scope ON product_categories FOR ALL
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
            AND (
                COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
                OR department_id IS NULL
                OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
            AND (
                COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
                OR department_id IS NULL
                OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
            )
        )
    );

-- 商品（原 00017）
DROP POLICY IF EXISTS core_products_scope ON products;
CREATE POLICY core_products_scope ON products FOR ALL
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
            AND (
                COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
                OR department_id IS NULL
                OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR (
            company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
            AND (
                COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'company'
                OR department_id IS NULL
                OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
            )
        )
    );

-- 漏網三表：原先完全沒有 policy
-- customer_counters 的 company_id 即主鍵
DROP POLICY IF EXISTS core_customer_counters_scope ON customer_counters;
CREATE POLICY core_customer_counters_scope ON customer_counters FOR ALL
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR company_id = (NULLIF(current_setting('app.current_company_id', true), ''))::bigint
    );

-- product_units / product_processing_specs 無 company_id：以父表 products 的存在性表達。
-- 子查詢本身也受 products 的 policy 約束（同一角色），故隔離具傳遞性。
DROP POLICY IF EXISTS core_product_units_scope ON product_units;
CREATE POLICY core_product_units_scope ON product_units FOR ALL
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR EXISTS (SELECT 1 FROM products p WHERE p.id = product_units.product_id)
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR EXISTS (SELECT 1 FROM products p WHERE p.id = product_units.product_id)
    );

DROP POLICY IF EXISTS core_product_processing_specs_scope ON product_processing_specs;
CREATE POLICY core_product_processing_specs_scope ON product_processing_specs FOR ALL
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR EXISTS (SELECT 1 FROM products p WHERE p.id = product_processing_specs.product_id)
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR EXISTS (SELECT 1 FROM products p WHERE p.id = product_processing_specs.product_id)
    );

-- 第 18 個 policy:core_metadicts_scope(00011)。00023 明文不動它的 WITH CHECK(刻意較嚴:
-- 不含 `department_id IS NULL`);本波只加 NULLIF,維持原語意。
DROP POLICY IF EXISTS core_metadicts_scope ON metadicts;
CREATE POLICY core_metadicts_scope ON metadicts
    USING (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR department_id IS NULL
        OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
    )
    WITH CHECK (
        COALESCE(NULLIF(current_setting('app.current_data_scope', true), ''), '') = 'all'
        OR department_id = (NULLIF(current_setting('app.current_department_id', true), ''))::bigint
    );
-- +goose StatementEnd

-- 部門級主檔四表:本波才真正打開守門(00016 只定義 policy、不 ENABLE)。
-- +goose StatementBegin
ALTER TABLE warehouses          ENABLE ROW LEVEL SECURITY;
ALTER TABLE warehouses          FORCE  ROW LEVEL SECURITY;
ALTER TABLE routes              ENABLE ROW LEVEL SECURITY;
ALTER TABLE routes              FORCE  ROW LEVEL SECURITY;
ALTER TABLE processing_specs    ENABLE ROW LEVEL SECURITY;
ALTER TABLE processing_specs    FORCE  ROW LEVEL SECURITY;
ALTER TABLE product_categories  ENABLE ROW LEVEL SECURITY;
ALTER TABLE product_categories  FORCE  ROW LEVEL SECURITY;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE warehouses          NO FORCE ROW LEVEL SECURITY;
ALTER TABLE warehouses          DISABLE ROW LEVEL SECURITY;
ALTER TABLE routes              NO FORCE ROW LEVEL SECURITY;
ALTER TABLE routes              DISABLE ROW LEVEL SECURITY;
ALTER TABLE processing_specs    NO FORCE ROW LEVEL SECURITY;
ALTER TABLE processing_specs    DISABLE ROW LEVEL SECURITY;
ALTER TABLE product_categories  NO FORCE ROW LEVEL SECURITY;
ALTER TABLE product_categories  DISABLE ROW LEVEL SECURITY;
-- +goose StatementEnd

-- policy 還原:17 個回 00023 的 Up 定義(逐字)、metadicts 回 00011 的定義(逐字),
-- 即 00025 之前的狀態(00023 補的 WITH CHECK 仍在,只是取值不再有 NULLIF)。
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

-- metadicts 還原為 00011 的定義。
DROP POLICY IF EXISTS core_metadicts_scope ON metadicts;
CREATE POLICY core_metadicts_scope ON metadicts
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR department_id IS NULL
        OR department_id = (current_setting('app.current_department_id', true))::bigint
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR department_id = (current_setting('app.current_department_id', true))::bigint
    );
-- +goose StatementEnd

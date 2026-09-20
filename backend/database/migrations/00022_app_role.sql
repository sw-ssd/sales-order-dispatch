-- 業務連線角色(D36):非 owner、NOBYPASSRLS，使 RLS 真正生效。
-- 密碼不由 migration 設定(不落版控):本機開發用 `task db:app-password`，
-- 其他環境由部署流程以 ALTER ROLE app_rw PASSWORD ... 提供。
--
-- 授權刻意**不用** `GRANT ... ON ALL TABLES/ALL SEQUENCES IN SCHEMA public`、也不用
-- `ALTER DEFAULT PRIVILEGES`(最小權限):OpenFGA 與業務共用同一個 database，其授權表
-- (tuple/authorization_model/store/… 由 cmd/migrate 在同一個 owner DSN 下接著建)就在 public，
-- 全庫授權會讓業務角色取得**授權資料**的 DML，且 default privileges 連未來新增的非業務表
-- (含 goose 版本表)都會一併外溢 —— 外洩 app_rw 密碼或任一業務路徑的 bug 就等於授權可被改寫。
-- 因此只明確列舉 18 張業務表(與 T3 的 policy 覆蓋清單一致);**新增業務表必須在其 migration 內
-- 自行 GRANT**，本 migration 不會自動涵蓋。
-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_rw') THEN
        CREATE ROLE app_rw LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOBYPASSRLS;
    END IF;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
GRANT USAGE ON SCHEMA public TO app_rw;
-- +goose StatementEnd

-- +goose StatementBegin
GRANT SELECT, INSERT, UPDATE, DELETE ON
    companies, departments, users, roles, role_permissions, audit_logs, metadicts,
    customers, customer_counters, customer_addresses, customer_contacts,
    warehouses, routes, processing_specs, product_categories, products,
    product_units, product_processing_specs
TO app_rw;
-- +goose StatementEnd

-- +goose StatementBegin
GRANT USAGE, SELECT ON
    companies_id_seq, departments_id_seq, users_id_seq, roles_id_seq,
    role_permissions_id_seq, audit_logs_id_seq, metadicts_id_seq, customers_id_seq,
    customer_counters_id_seq, customer_addresses_id_seq, customer_contacts_id_seq,
    warehouses_id_seq, routes_id_seq, processing_specs_id_seq, product_categories_id_seq,
    products_id_seq, product_units_id_seq, product_processing_specs_id_seq
TO app_rw;
-- +goose StatementEnd

-- +goose Down
-- 必須先逐一撤銷授權再 DROP ROLE:PG 會以既有 ACL 依賴擋下 DROP ROLE
-- (ERROR: role "app_rw" cannot be dropped because some objects depend on it, SQLSTATE 2BP01)。
-- 撤銷清單與 Up 逐項對稱(未授權過 default privileges,故無需撤銷)。
-- +goose StatementBegin
REVOKE SELECT, INSERT, UPDATE, DELETE ON
    companies, departments, users, roles, role_permissions, audit_logs, metadicts,
    customers, customer_counters, customer_addresses, customer_contacts,
    warehouses, routes, processing_specs, product_categories, products,
    product_units, product_processing_specs
FROM app_rw;
-- +goose StatementEnd

-- +goose StatementBegin
REVOKE USAGE, SELECT ON
    companies_id_seq, departments_id_seq, users_id_seq, roles_id_seq,
    role_permissions_id_seq, audit_logs_id_seq, metadicts_id_seq, customers_id_seq,
    customer_counters_id_seq, customer_addresses_id_seq, customer_contacts_id_seq,
    warehouses_id_seq, routes_id_seq, processing_specs_id_seq, product_categories_id_seq,
    products_id_seq, product_units_id_seq, product_processing_specs_id_seq
FROM app_rw;
-- +goose StatementEnd

-- +goose StatementBegin
REVOKE USAGE ON SCHEMA public FROM app_rw;
-- +goose StatementEnd

-- +goose StatementBegin
DROP ROLE IF EXISTS app_rw;
-- +goose StatementEnd

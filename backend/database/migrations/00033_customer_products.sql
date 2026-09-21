-- 04 客戶專屬商品清單(3.5.1):customer_products,一客戶一商品一筆。
-- alias_name 客戶慣用名稱;default_qty 十進位文字(0 合法:保留不顯示);明確不含單價欄位(D12)。
-- 部分唯一 (customer_id, product_id) WHERE deleted_at IS NULL(D10)。
-- RLS policy 僅「定義」不 ENABLE/FORCE:ENABLE 與服務收斂同批(比照 00031/00032 模式)。
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS customer_products (
    id            bigserial PRIMARY KEY,
    company_id    bigint NOT NULL,
    department_id bigint,
    customer_id   bigint NOT NULL,
    product_id    bigint NOT NULL,
    alias_name    text NOT NULL,
    default_qty   text NOT NULL DEFAULT '0',
    cut_note      text,
    promo_tag_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
    created_by    bigint,
    updated_by    bigint,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS customer_products_customer_product_unique
    ON customer_products (customer_id, product_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS customer_products_company_customer_idx
    ON customer_products (company_id, customer_id) WHERE deleted_at IS NULL;
ALTER TABLE customer_products
    ADD CONSTRAINT customer_products_company_fk    FOREIGN KEY (company_id)    REFERENCES companies(id),
    ADD CONSTRAINT customer_products_department_fk FOREIGN KEY (department_id) REFERENCES departments(id),
    ADD CONSTRAINT customer_products_customer_fk   FOREIGN KEY (customer_id)   REFERENCES customers(id),
    ADD CONSTRAINT customer_products_product_fk    FOREIGN KEY (product_id)    REFERENCES products(id);
-- +goose StatementEnd

-- RLS:依 data_scope + company/department 隔離(僅定義不 ENABLE)。
-- +goose StatementBegin
DROP POLICY IF EXISTS core_customer_products_scope ON customer_products;
CREATE POLICY core_customer_products_scope ON customer_products FOR ALL
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

-- app_rw 授權(比照 00022 明示列舉)。
GRANT SELECT, INSERT, UPDATE, DELETE ON customer_products TO app_rw;
GRANT USAGE, SELECT ON customer_products_id_seq TO app_rw;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
REVOKE USAGE, SELECT ON customer_products_id_seq FROM app_rw;
REVOKE SELECT, INSERT, UPDATE, DELETE ON customer_products FROM app_rw;
DROP POLICY IF EXISTS core_customer_products_scope ON customer_products;
DROP TABLE IF EXISTS customer_products;
-- +goose StatementEnd

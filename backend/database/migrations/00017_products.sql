-- 04 商品主檔(3.3, D10/D18):products / product_units / product_processing_specs。
-- products 承載商品主列(部門內 code 唯一、軟刪除);product_units 一商品多組單位換算
-- (conversion_rate 存十進位文字,禁二進位浮點);product_processing_specs 商品×處理規格多對多。
-- unit/關聯列隨商品整組替換(3.3.2),不設軟刪除。
-- RLS policy 比照 00016:僅「定義」不 ENABLE/FORCE(D3 待每請求交易層接線)。
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS products (
    id                     bigserial PRIMARY KEY,
    company_id             bigint NOT NULL,
    department_id          bigint,
    code                   text NOT NULL,
    name                   text NOT NULL,
    category_id            bigint,
    inventory_warehouse_id bigint,
    picking_warehouse_id   bigint,
    description            text,
    is_active              boolean NOT NULL DEFAULT true,
    created_by             bigint,
    updated_by             bigint,
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now(),
    deleted_at             timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS products_dept_code_unique
    ON products (department_id, code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS products_category_idx ON products (category_id) WHERE deleted_at IS NULL;
ALTER TABLE products
    ADD CONSTRAINT products_company_fk            FOREIGN KEY (company_id) REFERENCES companies(id),
    ADD CONSTRAINT products_department_fk         FOREIGN KEY (department_id) REFERENCES departments(id),
    ADD CONSTRAINT products_category_fk           FOREIGN KEY (category_id) REFERENCES product_categories(id),
    ADD CONSTRAINT products_inventory_warehouse_fk FOREIGN KEY (inventory_warehouse_id) REFERENCES warehouses(id),
    ADD CONSTRAINT products_picking_warehouse_fk  FOREIGN KEY (picking_warehouse_id) REFERENCES warehouses(id);

CREATE TABLE IF NOT EXISTS product_units (
    id              bigserial PRIMARY KEY,
    product_id      bigint NOT NULL,
    unit_code       text NOT NULL,
    conversion_rate text NOT NULL,             -- 十進位文字,換算為基本單位之比率
    is_base         boolean NOT NULL DEFAULT false,
    sort_order      integer NOT NULL DEFAULT 0,
    size_desc       text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS product_units_product_unitcode_unique
    ON product_units (product_id, unit_code);
CREATE UNIQUE INDEX IF NOT EXISTS product_units_single_base_unique
    ON product_units (product_id) WHERE is_base = true;  -- 每商品恰好一個基本單位
ALTER TABLE product_units
    ADD CONSTRAINT product_units_product_fk FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE;

CREATE TABLE IF NOT EXISTS product_processing_specs (
    id                 bigserial PRIMARY KEY,
    product_id         bigint NOT NULL,
    processing_spec_id bigint NOT NULL,
    attributes         jsonb NOT NULL DEFAULT '{}'::jsonb, -- 配對層覆寫指令
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS product_specs_product_spec_unique
    ON product_processing_specs (product_id, processing_spec_id);
CREATE INDEX IF NOT EXISTS product_specs_spec_idx ON product_processing_specs (processing_spec_id);
ALTER TABLE product_processing_specs
    ADD CONSTRAINT product_specs_product_fk FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    ADD CONSTRAINT product_specs_spec_fk    FOREIGN KEY (processing_spec_id) REFERENCES processing_specs(id);
-- +goose StatementEnd

-- RLS:products 依 data_scope + company/department 隔離(僅定義不 ENABLE)。unit/關聯列隨 product 承襲。
-- +goose StatementBegin
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

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS product_processing_specs;
DROP TABLE IF EXISTS product_units;
DROP TABLE IF EXISTS products;
-- +goose StatementEnd

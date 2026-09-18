-- 04 部門級主檔(3.4.1–3.4.4, D10/D18):warehouses / routes / processing_specs / product_categories。
-- 皆 company_id/department_id + 軟刪除 + 稽核欄位;code 部門內部分唯一(軟刪除後可重用)。
-- RLS policy 比照 00013/00015:僅「定義」不 ENABLE/FORCE(D3 待每請求交易層接線)。
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS warehouses (
    id            bigserial PRIMARY KEY,
    company_id    bigint NOT NULL,
    department_id bigint,
    code          text NOT NULL,
    name          text NOT NULL,
    address       text,
    is_active     boolean NOT NULL DEFAULT true,
    created_by    bigint,
    updated_by    bigint,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS warehouses_dept_code_unique
    ON warehouses (department_id, code) WHERE deleted_at IS NULL;
ALTER TABLE warehouses
    ADD CONSTRAINT warehouses_company_fk    FOREIGN KEY (company_id) REFERENCES companies(id),
    ADD CONSTRAINT warehouses_department_fk FOREIGN KEY (department_id) REFERENCES departments(id);

CREATE TABLE IF NOT EXISTS routes (
    id            bigserial PRIMARY KEY,
    company_id    bigint NOT NULL,
    department_id bigint,
    code          text NOT NULL,
    name          text NOT NULL,
    description   text,
    sort_order    integer NOT NULL DEFAULT 0,
    is_active     boolean NOT NULL DEFAULT true,
    created_by    bigint,
    updated_by    bigint,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS routes_dept_code_unique
    ON routes (department_id, code) WHERE deleted_at IS NULL;
ALTER TABLE routes
    ADD CONSTRAINT routes_company_fk    FOREIGN KEY (company_id) REFERENCES companies(id),
    ADD CONSTRAINT routes_department_fk FOREIGN KEY (department_id) REFERENCES departments(id);

CREATE TABLE IF NOT EXISTS processing_specs (
    id                      bigserial PRIMARY KEY,
    company_id              bigint NOT NULL,
    department_id           bigint,
    code                    text NOT NULL,
    name                    text NOT NULL,
    kind                    text NOT NULL DEFAULT 'other',  -- 開放集,metadicts processing_kind 背書
    applies_to_processing   boolean NOT NULL DEFAULT false,
    applies_to_picking      boolean NOT NULL DEFAULT false,
    attributes              jsonb NOT NULL DEFAULT '{}'::jsonb,
    sort_order              integer NOT NULL DEFAULT 0,
    is_active               boolean NOT NULL DEFAULT true,
    created_by              bigint,
    updated_by              bigint,
    created_at              timestamptz NOT NULL DEFAULT now(),
    updated_at              timestamptz NOT NULL DEFAULT now(),
    deleted_at              timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS processing_specs_dept_code_unique
    ON processing_specs (department_id, code) WHERE deleted_at IS NULL;
ALTER TABLE processing_specs
    ADD CONSTRAINT processing_specs_company_fk    FOREIGN KEY (company_id) REFERENCES companies(id),
    ADD CONSTRAINT processing_specs_department_fk FOREIGN KEY (department_id) REFERENCES departments(id);

CREATE TABLE IF NOT EXISTS product_categories (
    id            bigserial PRIMARY KEY,
    company_id    bigint NOT NULL,
    department_id bigint,
    code          text NOT NULL,
    name          text NOT NULL,
    sort_order    integer NOT NULL DEFAULT 0,
    is_active     boolean NOT NULL DEFAULT true,
    created_by    bigint,
    updated_by    bigint,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS product_categories_dept_code_unique
    ON product_categories (department_id, code) WHERE deleted_at IS NULL;
ALTER TABLE product_categories
    ADD CONSTRAINT product_categories_company_fk    FOREIGN KEY (company_id) REFERENCES companies(id),
    ADD CONSTRAINT product_categories_department_fk FOREIGN KEY (department_id) REFERENCES departments(id);
-- +goose StatementEnd

-- RLS:四表依 data_scope + company/department 隔離(僅定義不 ENABLE)。
-- +goose StatementBegin
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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS product_categories;
DROP TABLE IF EXISTS processing_specs;
DROP TABLE IF EXISTS routes;
DROP TABLE IF EXISTS warehouses;
-- +goose StatementEnd

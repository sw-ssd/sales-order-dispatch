-- 04 客戶主檔核心(3.1.1/3.1.3, D7/D10):company.customer_code_prefix + customers + customer_counters。
-- customer_code 取號為「公司前綴+6 位自增」(D7),customers 建檔與 counters 更新同交易(服務層)。
-- RLS policy 比照 00010/00011:僅「定義」不 ENABLE/FORCE(D3 待每請求交易層 SET LOCAL 後啟用)。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE companies ADD COLUMN IF NOT EXISTS customer_code_prefix text;

CREATE TABLE IF NOT EXISTS customers (
    id                     bigserial PRIMARY KEY,
    company_id             bigint NOT NULL,
    department_id          bigint,
    customer_code          text NOT NULL,
    name                   text NOT NULL,
    tax_id                 text,
    payment_method_id      bigint,
    settlement_method_id   bigint,
    customer_type_id       bigint,
    invoice_type_id        bigint,
    default_sales_rep_id   bigint,
    preferred_delivery_days jsonb NOT NULL DEFAULT '[]'::jsonb,
    promo_tag_ids          jsonb NOT NULL DEFAULT '[]'::jsonb,
    created_by             bigint,
    updated_by             bigint,
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now(),
    deleted_at             timestamptz
);
-- 部分唯一索引(D10):公司內 customer_code 唯一,軟刪除後可重用。
CREATE UNIQUE INDEX IF NOT EXISTS customers_company_code_unique
    ON customers (company_id, customer_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS customers_dept_idx       ON customers (department_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS customers_sales_rep_idx  ON customers (default_sales_rep_id);
-- FK(比照 audit_logs 補約束慣例)。
ALTER TABLE customers ADD CONSTRAINT customers_company_fk      FOREIGN KEY (company_id)      REFERENCES companies(id);
ALTER TABLE customers ADD CONSTRAINT customers_department_fk   FOREIGN KEY (department_id)   REFERENCES departments(id);
ALTER TABLE customers ADD CONSTRAINT customers_sales_rep_fk    FOREIGN KEY (default_sales_rep_id) REFERENCES users(id);

CREATE TABLE IF NOT EXISTS customer_counters (
    company_id  bigint PRIMARY KEY,
    next_seq    integer NOT NULL DEFAULT 1,
    version     integer NOT NULL DEFAULT 0,
    updated_at  timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- RLS:customers 依 data_scope + company/department 隔離(僅定義不 ENABLE,同 00010/00011)。
-- +goose StatementBegin
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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS customer_counters;
DROP TABLE IF EXISTS customers;
ALTER TABLE companies DROP COLUMN IF EXISTS customer_code_prefix;
-- +goose StatementEnd

-- 04 客戶地址簿與聯絡人(3.2.1/3.2.2, D10/D18):customer_addresses + customer_contacts。
-- 兩表皆自客戶複寫 company_id/department_id(供 RLS);軟刪除 + 稽核欄位。
-- 「同類型/同客戶至多一筆預設」以部分唯一索引(WITH is_default AND deleted_at IS NULL)兜底;
-- 服務層於設定預設時先清同類型其餘預設、再設目標(先清後設)。
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS customer_addresses (
    id             bigserial PRIMARY KEY,
    company_id     bigint NOT NULL,
    customer_id    bigint NOT NULL,
    department_id  bigint,
    type           text NOT NULL DEFAULT 'other',             -- shipping | billing | other
    recipient_name text NOT NULL,
    phone          text,
    address_line   text NOT NULL,
    city           text,
    postal_code    text,
    is_default     boolean NOT NULL DEFAULT false,
    created_by     bigint,
    updated_by     bigint,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    deleted_at     timestamptz
);
-- 部分唯一索引(D10):同客戶同類型至多一筆預設;軟刪除後可重用。
CREATE UNIQUE INDEX IF NOT EXISTS customer_addresses_default_unique
    ON customer_addresses (customer_id, type) WHERE is_default = true AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS customer_addresses_customer_idx ON customer_addresses (customer_id);
CREATE INDEX IF NOT EXISTS customer_addresses_customer_type_idx ON customer_addresses (customer_id, type);
ALTER TABLE customer_addresses
    ADD CONSTRAINT customer_addresses_company_fk    FOREIGN KEY (company_id)   REFERENCES companies(id),
    ADD CONSTRAINT customer_addresses_customer_fk   FOREIGN KEY (customer_id)  REFERENCES customers(id),
    ADD CONSTRAINT customer_addresses_department_fk FOREIGN KEY (department_id) REFERENCES departments(id);

CREATE TABLE IF NOT EXISTS customer_contacts (
    id            bigserial PRIMARY KEY,
    company_id    bigint NOT NULL,
    customer_id   bigint NOT NULL,
    department_id bigint,
    name          text NOT NULL,
    title         text,
    email         text,
    phone         text,
    is_default    boolean NOT NULL DEFAULT false,
    created_by    bigint,
    updated_by    bigint,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz
);
-- 部分唯一索引(D10):每客戶至多一筆預設聯絡人。
CREATE UNIQUE INDEX IF NOT EXISTS customer_contacts_default_unique
    ON customer_contacts (customer_id) WHERE is_default = true AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS customer_contacts_customer_idx ON customer_contacts (customer_id);
ALTER TABLE customer_contacts
    ADD CONSTRAINT customer_contacts_company_fk    FOREIGN KEY (company_id)   REFERENCES companies(id),
    ADD CONSTRAINT customer_contacts_customer_fk   FOREIGN KEY (customer_id)  REFERENCES customers(id),
    ADD CONSTRAINT customer_contacts_department_fk FOREIGN KEY (department_id) REFERENCES departments(id);
-- +goose StatementEnd

-- RLS:地址/聯絡人依 data_scope + company/department 隔離(僅定義不 ENABLE,同 00013)。
-- +goose StatementBegin
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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS customer_contacts;
DROP TABLE IF EXISTS customer_addresses;
-- +goose StatementEnd

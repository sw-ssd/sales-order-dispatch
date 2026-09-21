-- 06 退貨申請(4.7.1):return_requests / return_request_items,欄位對齊細部 06-returns.md。
-- 業務實體軟刪除(D10);審核全程不修改原訂單(僅參照)。
-- RLS policy 僅「定義」不 ENABLE/FORCE:ENABLE 與服務收斂同批(比照 00033-00038 模式)。
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS return_requests (
    id                   bigserial PRIMARY KEY,
    company_id           bigint NOT NULL,
    department_id        bigint NOT NULL,
    customer_id          bigint NOT NULL,
    created_by_user_id   bigint NOT NULL,
    status               text NOT NULL DEFAULT 'pending',
    remark               text,
    reviewed_by_user_id  bigint,
    reviewed_at          timestamptz,
    reject_reason        text,
    version              integer NOT NULL DEFAULT 0,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz
);
CREATE INDEX IF NOT EXISTS return_requests_customer_status_idx
    ON return_requests (company_id, customer_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS return_requests_dept_status_idx
    ON return_requests (company_id, department_id, status) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS return_request_items (
    id                   bigserial PRIMARY KEY,
    return_request_id    bigint NOT NULL REFERENCES return_requests(id),
    company_id           bigint NOT NULL,
    department_id        bigint NOT NULL,
    source_type          text NOT NULL,
    sales_order_id       bigint,
    sales_order_item_id  bigint,
    customer_product_id  bigint,
    product_id           bigint NOT NULL,
    product_name         text NOT NULL,
    spec                 text,
    unit                 text NOT NULL,
    quantity             text NOT NULL,
    reason               text NOT NULL,
    photo_file_ids       jsonb NOT NULL DEFAULT '[]',
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz
);
CREATE INDEX IF NOT EXISTS return_request_items_request_idx
    ON return_request_items (return_request_id) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- RLS:依 data_scope + company/department 隔離,客戶 self 範圍再比對 customer_id(僅定義不 ENABLE)。
-- +goose StatementBegin
DROP POLICY IF EXISTS core_return_requests_scope ON return_requests;
CREATE POLICY core_return_requests_scope ON return_requests FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id = (current_setting('app.current_department_id', true))::bigint
                OR (
                    COALESCE(current_setting('app.current_data_scope', true), '') = 'self'
                    AND customer_id = NULLIF(current_setting('app.current_customer_id', true), '')::bigint
                )
            )
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id = (current_setting('app.current_department_id', true))::bigint
                OR (
                    COALESCE(current_setting('app.current_data_scope', true), '') = 'self'
                    AND customer_id = NULLIF(current_setting('app.current_customer_id', true), '')::bigint
                )
            )
        )
    );

DROP POLICY IF EXISTS core_return_request_items_scope ON return_request_items;
CREATE POLICY core_return_request_items_scope ON return_request_items FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
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
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

-- app_rw 授權(比照 00022 明示列舉;業務實體 CRUD 全授)。
GRANT SELECT, INSERT, UPDATE, DELETE ON return_requests TO app_rw;
GRANT USAGE, SELECT ON return_requests_id_seq TO app_rw;
GRANT SELECT, INSERT, UPDATE, DELETE ON return_request_items TO app_rw;
GRANT USAGE, SELECT ON return_request_items_id_seq TO app_rw;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
REVOKE USAGE, SELECT ON return_request_items_id_seq FROM app_rw;
REVOKE SELECT, INSERT, UPDATE, DELETE ON return_request_items FROM app_rw;
REVOKE USAGE, SELECT ON return_requests_id_seq FROM app_rw;
REVOKE SELECT, INSERT, UPDATE, DELETE ON return_requests FROM app_rw;
DROP POLICY IF EXISTS core_return_request_items_scope ON return_request_items;
DROP POLICY IF EXISTS core_return_requests_scope ON return_requests;
DROP TABLE IF EXISTS return_request_items;
DROP TABLE IF EXISTS return_requests;
-- +goose StatementEnd

-- 05 銷售訂單(4.1.1, D7/D10/D12/D13):sales_orders / sales_order_items / sales_order_events + order_counters。
-- sales_orders 承載狀態機(D13:pending ⇄ processing → completed、pending → cancelled、completed → voided)、
-- 來源(訂單來源字典 order_source)、派車欄位(08-dispatch 寫入,本表僅定義)與 version 樂觀鎖(D14)。
-- 不含任何金額欄位(D12,硬性邊界)。
-- order_counters 依 (company_id, source) 一列,樂觀鎖取號,取號與建單同交易(D7)。
-- RLS policy 僅「定義」不 ENABLE/FORCE:ENABLE 與收斂同批(比照 00024-00028 模式,服務層先行)。
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sales_orders (
    id                     bigserial PRIMARY KEY,
    company_id             bigint NOT NULL,
    department_id          bigint,
    order_no               text NOT NULL,
    customer_id            bigint NOT NULL,
    source                 text NOT NULL,
    status                 text NOT NULL DEFAULT 'pending',
    expected_delivery_date date,
    sales_rep_id           bigint,
    note                   text,
    dispatched_at          timestamptz,
    dispatched_by          bigint,
    route_id               bigint,
    delivery_sequence      integer,
    version                integer NOT NULL DEFAULT 1,
    created_by             bigint,
    updated_by             bigint,
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now(),
    deleted_at             timestamptz
);
-- 部分唯一索引(D10):公司內 order_no 唯一,軟刪除後序號不回退(由計數器保證)。
CREATE UNIQUE INDEX IF NOT EXISTS sales_orders_company_no_unique
    ON sales_orders (company_id, order_no) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS sales_orders_list_idx
    ON sales_orders (company_id, department_id, status, expected_delivery_date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS sales_orders_customer_idx ON sales_orders (customer_id) WHERE deleted_at IS NULL;
ALTER TABLE sales_orders
    ADD CONSTRAINT sales_orders_company_fk     FOREIGN KEY (company_id)    REFERENCES companies(id),
    ADD CONSTRAINT sales_orders_department_fk  FOREIGN KEY (department_id) REFERENCES departments(id),
    ADD CONSTRAINT sales_orders_customer_fk    FOREIGN KEY (customer_id)   REFERENCES customers(id),
    ADD CONSTRAINT sales_orders_sales_rep_fk   FOREIGN KEY (sales_rep_id)  REFERENCES users(id),
    ADD CONSTRAINT sales_orders_route_fk       FOREIGN KEY (route_id)      REFERENCES routes(id);

CREATE TABLE IF NOT EXISTS sales_order_items (
    id                 bigserial PRIMARY KEY,
    sales_order_id     bigint NOT NULL,
    company_id         bigint NOT NULL,
    department_id      bigint,
    product_id         bigint,
    display_name       text NOT NULL,
    qty                text NOT NULL,   -- 十進位文字,禁二進位浮點
    unit               text NOT NULL,
    base_qty           text NOT NULL,   -- 十進位文字,換算基本單位後數量
    processing_spec_id bigint,
    special_cut_note   text,
    warehouse_id       bigint,
    sort_order         integer NOT NULL DEFAULT 0,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),
    deleted_at         timestamptz
);
CREATE INDEX IF NOT EXISTS sales_order_items_order_idx ON sales_order_items (sales_order_id) WHERE deleted_at IS NULL;
ALTER TABLE sales_order_items
    ADD CONSTRAINT sales_order_items_order_fk   FOREIGN KEY (sales_order_id)     REFERENCES sales_orders(id) ON DELETE CASCADE,
    ADD CONSTRAINT sales_order_items_company_fk FOREIGN KEY (company_id)         REFERENCES companies(id),
    ADD CONSTRAINT sales_order_items_product_fk FOREIGN KEY (product_id)         REFERENCES products(id);

CREATE TABLE IF NOT EXISTS sales_order_events (
    id             bigserial PRIMARY KEY,
    sales_order_id bigint NOT NULL,
    company_id     bigint NOT NULL,
    event_type     text NOT NULL,
    actor_id       bigint NOT NULL,
    reason         text,
    payload        jsonb,
    created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS sales_order_events_order_idx ON sales_order_events (sales_order_id, created_at);
ALTER TABLE sales_order_events
    ADD CONSTRAINT sales_order_events_order_fk   FOREIGN KEY (sales_order_id) REFERENCES sales_orders(id) ON DELETE CASCADE,
    ADD CONSTRAINT sales_order_events_company_fk FOREIGN KEY (company_id)     REFERENCES companies(id);

CREATE TABLE IF NOT EXISTS order_counters (
    company_id bigint NOT NULL,
    source     text NOT NULL,
    next_seq   integer NOT NULL DEFAULT 1,
    version    integer NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (company_id, source)
);
-- +goose StatementEnd

-- RLS:訂單三表 + counters 依 data_scope + company/department 隔離(僅定義不 ENABLE)。
-- events 僅開 SELECT/INSERT,REVOKE UPDATE/DELETE(僅追加,D10)。
-- +goose StatementBegin
DROP POLICY IF EXISTS core_sales_orders_scope ON sales_orders;
CREATE POLICY core_sales_orders_scope ON sales_orders FOR ALL
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

DROP POLICY IF EXISTS core_sales_order_items_scope ON sales_order_items;
CREATE POLICY core_sales_order_items_scope ON sales_order_items FOR ALL
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

DROP POLICY IF EXISTS core_sales_order_events_scope ON sales_order_events;
CREATE POLICY core_sales_order_events_scope ON sales_order_events FOR SELECT
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR company_id = (current_setting('app.current_company_id', true))::bigint
    );
DROP POLICY IF EXISTS core_sales_order_events_insert ON sales_order_events;
CREATE POLICY core_sales_order_events_insert ON sales_order_events FOR INSERT
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR company_id = (current_setting('app.current_company_id', true))::bigint
    );

DROP POLICY IF EXISTS core_order_counters_scope ON order_counters;
CREATE POLICY core_order_counters_scope ON order_counters FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR company_id = (current_setting('app.current_company_id', true))::bigint
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR company_id = (current_setting('app.current_company_id', true))::bigint
    );

-- app_rw 授權(比照 00022 明示列舉,不用 ON ALL TABLES)。
-- events 僅 SELECT/INSERT(僅追加,無 UPDATE/DELETE)。
GRANT SELECT, INSERT, UPDATE, DELETE ON sales_orders, sales_order_items, order_counters TO app_rw;
GRANT SELECT, INSERT ON sales_order_events TO app_rw;
GRANT USAGE, SELECT ON sales_orders_id_seq, sales_order_items_id_seq, sales_order_events_id_seq TO app_rw;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
REVOKE USAGE, SELECT ON sales_orders_id_seq, sales_order_items_id_seq, sales_order_events_id_seq FROM app_rw;
REVOKE SELECT, INSERT ON sales_order_events FROM app_rw;
REVOKE SELECT, INSERT, UPDATE, DELETE ON sales_orders, sales_order_items, order_counters FROM app_rw;
DROP POLICY IF EXISTS core_order_counters_scope ON order_counters;
DROP POLICY IF EXISTS core_sales_order_events_insert ON sales_order_events;
DROP POLICY IF EXISTS core_sales_order_events_scope ON sales_order_events;
DROP POLICY IF EXISTS core_sales_order_items_scope ON sales_order_items;
DROP POLICY IF EXISTS core_sales_orders_scope ON sales_orders;
DROP TABLE IF EXISTS order_counters;
DROP TABLE IF EXISTS sales_order_events;
DROP TABLE IF EXISTS sales_order_items;
DROP TABLE IF EXISTS sales_orders;
-- +goose StatementEnd

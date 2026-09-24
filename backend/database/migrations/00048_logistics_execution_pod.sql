-- D32 logistics 執行層 10.6(配送執行與簽收):執行軌跡表 + POD 表,並補狀態時間欄。
-- logistics_delivery_events:append-only(比照 sales_order_events:只授 SELECT/INSERT)。
-- logistics_proofs:軟刪除(D10);實體檔走既有 file_assets(file_asset_id,D17)。
-- RLS policy 僅「定義」不 ENABLE/FORCE:ENABLE+FORCE 於 00049 獨立檔(比照 00046/00047 模式)。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE logistics_deliveries ADD COLUMN IF NOT EXISTS started_at   timestamptz;
ALTER TABLE logistics_deliveries ADD COLUMN IF NOT EXISTS completed_at timestamptz;

CREATE TABLE IF NOT EXISTS logistics_delivery_events (
    id                     bigserial PRIMARY KEY,
    logistics_delivery_id  bigint NOT NULL REFERENCES logistics_deliveries(id),
    company_id             bigint NOT NULL,
    event_type             text NOT NULL, -- created / started / completed / cancelled
    actor_id               bigint NOT NULL REFERENCES users(id),
    reason                 text,
    payload                jsonb,
    created_at             timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS logistics_delivery_events_delivery_idx
    ON logistics_delivery_events (logistics_delivery_id, created_at);

CREATE TABLE IF NOT EXISTS logistics_proofs (
    id                     bigserial PRIMARY KEY,
    logistics_delivery_id  bigint NOT NULL REFERENCES logistics_deliveries(id),
    company_id             bigint NOT NULL,
    department_id          bigint,
    proof_type             text NOT NULL, -- photo / signature / scan
    file_asset_id          bigint NOT NULL REFERENCES file_assets(id),
    remarks                text,
    captured_by            bigint NOT NULL REFERENCES users(id),
    captured_at            timestamptz NOT NULL,
    deleted_at             timestamptz,
    created_at             timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS logistics_proofs_delivery_idx
    ON logistics_proofs (logistics_delivery_id) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- RLS:部門級隔離(比照 00046:company 比對 + scope/department 條件)。
-- +goose StatementBegin
DROP POLICY IF EXISTS core_logistics_delivery_events_scope ON logistics_delivery_events;
CREATE POLICY core_logistics_delivery_events_scope ON logistics_delivery_events FOR SELECT
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR company_id = (current_setting('app.current_company_id', true))::bigint
    );
DROP POLICY IF EXISTS core_logistics_delivery_events_insert ON logistics_delivery_events;
CREATE POLICY core_logistics_delivery_events_insert ON logistics_delivery_events FOR INSERT
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR company_id = (current_setting('app.current_company_id', true))::bigint
    );

DROP POLICY IF EXISTS core_logistics_proofs_scope ON logistics_proofs;
CREATE POLICY core_logistics_proofs_scope ON logistics_proofs FOR ALL
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

-- app_rw 授權(明示列舉;軌跡表 append-only 只給 SELECT/INSERT)。
GRANT SELECT, INSERT ON logistics_delivery_events TO app_rw;
GRANT SELECT, INSERT, UPDATE, DELETE ON logistics_proofs TO app_rw;
GRANT USAGE, SELECT ON logistics_delivery_events_id_seq, logistics_proofs_id_seq TO app_rw;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
REVOKE USAGE, SELECT ON logistics_delivery_events_id_seq, logistics_proofs_id_seq FROM app_rw;
REVOKE SELECT, INSERT ON logistics_delivery_events FROM app_rw;
REVOKE SELECT, INSERT, UPDATE, DELETE ON logistics_proofs FROM app_rw;
DROP POLICY IF EXISTS core_logistics_proofs_scope ON logistics_proofs;
DROP POLICY IF EXISTS core_logistics_delivery_events_insert ON logistics_delivery_events;
DROP POLICY IF EXISTS core_logistics_delivery_events_scope ON logistics_delivery_events;
DROP TABLE IF EXISTS logistics_proofs;
DROP TABLE IF EXISTS logistics_delivery_events;
ALTER TABLE logistics_deliveries DROP COLUMN IF EXISTS completed_at;
ALTER TABLE logistics_deliveries DROP COLUMN IF EXISTS started_at;
-- +goose StatementEnd

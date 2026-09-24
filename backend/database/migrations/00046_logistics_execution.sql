-- D32 logistics 執行層首批(D32/10.1):logistics_drivers / vehicles / logistics_deliveries。
-- 軟刪除(D10);vehicles.plate_no 部門內部分唯一;logistics_deliveries.version 樂觀鎖(10.4)。
-- RLS policy 僅「定義」不 ENABLE/FORCE:ENABLE+FORCE 於 00047 獨立檔(比照 00031-00045 模式)。
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS logistics_drivers (
    id             bigserial PRIMARY KEY,
    company_id     bigint NOT NULL,
    department_id  bigint,
    user_id        bigint NOT NULL REFERENCES users(id),
    name           text NOT NULL,
    phone          text,
    current_status text NOT NULL DEFAULT 'offline',
    deleted_at     timestamptz,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS logistics_drivers_dept_user_idx
    ON logistics_drivers (department_id, user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS logistics_drivers_user_idx
    ON logistics_drivers (user_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS vehicles (
    id            bigserial PRIMARY KEY,
    company_id    bigint NOT NULL,
    department_id bigint,
    plate_no      text NOT NULL,
    vehicle_type  text,
    status        text NOT NULL DEFAULT 'idle',
    deleted_at    timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS vehicles_plate_unique
    ON vehicles (department_id, plate_no) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS logistics_deliveries (
    id            bigserial PRIMARY KEY,
    company_id    bigint NOT NULL,
    department_id bigint,
    route_id      bigint NOT NULL REFERENCES routes(id),
    driver_id     bigint REFERENCES logistics_drivers(id),
    vehicle_id    bigint REFERENCES vehicles(id),
    assigned_by   bigint NOT NULL REFERENCES users(id),
    status        text NOT NULL DEFAULT 'pending',
    version       integer NOT NULL DEFAULT 1,
    deleted_at    timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS logistics_deliveries_dept_route_idx
    ON logistics_deliveries (department_id, route_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS logistics_deliveries_driver_idx
    ON logistics_deliveries (driver_id) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- RLS:部門級隔離(比照 00041 notifications 家族:company 比對 + scope/department 條件)。
-- +goose StatementBegin
DROP POLICY IF EXISTS core_logistics_drivers_scope ON logistics_drivers;
CREATE POLICY core_logistics_drivers_scope ON logistics_drivers FOR ALL
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

DROP POLICY IF EXISTS core_vehicles_scope ON vehicles;
CREATE POLICY core_vehicles_scope ON vehicles FOR ALL
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

DROP POLICY IF EXISTS core_logistics_deliveries_scope ON logistics_deliveries;
CREATE POLICY core_logistics_deliveries_scope ON logistics_deliveries FOR ALL
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

-- app_rw 授權(比照 00022 明示列舉;含序列 —— 慣例 1-1:表級 DML 與序列是兩份清單)。
GRANT SELECT, INSERT, UPDATE, DELETE ON logistics_drivers, vehicles, logistics_deliveries TO app_rw;
GRANT USAGE, SELECT ON logistics_drivers_id_seq, vehicles_id_seq, logistics_deliveries_id_seq TO app_rw;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
REVOKE USAGE, SELECT ON logistics_drivers_id_seq, vehicles_id_seq, logistics_deliveries_id_seq FROM app_rw;
REVOKE SELECT, INSERT, UPDATE, DELETE ON logistics_drivers, vehicles, logistics_deliveries FROM app_rw;
DROP POLICY IF EXISTS core_logistics_deliveries_scope ON logistics_deliveries;
DROP POLICY IF EXISTS core_vehicles_scope ON vehicles;
DROP POLICY IF EXISTS core_logistics_drivers_scope ON logistics_drivers;
DROP TABLE IF EXISTS logistics_deliveries;
DROP TABLE IF EXISTS vehicles;
DROP TABLE IF EXISTS logistics_drivers;
-- +goose StatementEnd

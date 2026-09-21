-- 09 列印記錄(5.5.1):print_logs / print_previews,欄位對齊規格 §5.2。
-- 記錄類不軟刪除(D10),由保留排程管理(print_logs 2 年、print_previews 90 天,§14.1;排程另案)。
-- RLS policy 僅「定義」不 ENABLE/FORCE:ENABLE 與服務收斂同批(比照 00033-00036 模式)。
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS print_logs (
    id               bigserial PRIMARY KEY,
    company_id       bigint NOT NULL,
    department_id    bigint NOT NULL,
    document_type    text NOT NULL,
    route_id         bigint NOT NULL,
    customer_id      bigint,
    warehouse_id     bigint,
    target_date      date NOT NULL,
    is_reprint       boolean NOT NULL DEFAULT false,
    reprint_reason   text,
    printed_by       bigint NOT NULL,
    printed_at       timestamptz NOT NULL DEFAULT now(),
    file_asset_id    bigint NOT NULL REFERENCES file_assets(id)
);
CREATE INDEX IF NOT EXISTS print_logs_dept_date_type_idx
    ON print_logs (department_id, target_date, document_type);
CREATE INDEX IF NOT EXISTS print_logs_reprint_key_idx
    ON print_logs (department_id, document_type, route_id, target_date);

CREATE TABLE IF NOT EXISTS print_previews (
    id               bigserial PRIMARY KEY,
    company_id       bigint NOT NULL,
    department_id    bigint NOT NULL,
    document_type    text NOT NULL,
    route_id         bigint NOT NULL,
    customer_id      bigint,
    warehouse_id     bigint,
    target_date      date NOT NULL,
    previewed_by     bigint NOT NULL,
    previewed_at     timestamptz NOT NULL DEFAULT now(),
    file_asset_id    bigint NOT NULL REFERENCES file_assets(id)
);
CREATE INDEX IF NOT EXISTS print_previews_dept_date_type_idx
    ON print_previews (department_id, target_date, document_type);
-- +goose StatementEnd

-- RLS:依 data_scope + company/department 隔離(僅定義不 ENABLE)。
-- +goose StatementBegin
DROP POLICY IF EXISTS core_print_logs_scope ON print_logs;
CREATE POLICY core_print_logs_scope ON print_logs FOR ALL
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

DROP POLICY IF EXISTS core_print_previews_scope ON print_previews;
CREATE POLICY core_print_previews_scope ON print_previews FOR ALL
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

-- app_rw 授權(比照 00022 明示列舉;記錄類僅追加,不授 UPDATE/DELETE —— 重印/預覽皆新列)。
GRANT SELECT, INSERT ON print_logs TO app_rw;
GRANT USAGE, SELECT ON print_logs_id_seq TO app_rw;
GRANT SELECT, INSERT ON print_previews TO app_rw;
GRANT USAGE, SELECT ON print_previews_id_seq TO app_rw;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
REVOKE USAGE, SELECT ON print_previews_id_seq FROM app_rw;
REVOKE SELECT, INSERT ON print_previews FROM app_rw;
REVOKE USAGE, SELECT ON print_logs_id_seq FROM app_rw;
REVOKE SELECT, INSERT ON print_logs FROM app_rw;
DROP POLICY IF EXISTS core_print_previews_scope ON print_previews;
DROP POLICY IF EXISTS core_print_logs_scope ON print_logs;
DROP TABLE IF EXISTS print_previews;
DROP TABLE IF EXISTS print_logs;
-- +goose StatementEnd

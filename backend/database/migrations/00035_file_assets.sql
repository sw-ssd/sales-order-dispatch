-- 04 檔案資產(3.6.2):file_assets,驗證通過的檔案元資料。
-- filename 系統產生(uuid + 正規化副檔名);storage_path 本地相對路徑(不對外);url 下載相對路徑。
-- 軟刪除(D10):實體檔案保留,下載拒絕。
-- RLS policy 僅「定義」不 ENABLE/FORCE:ENABLE 與服務收斂同批(比照 00033/00034 模式)。
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS file_assets (
    id                bigserial PRIMARY KEY,
    company_id        bigint NOT NULL,
    department_id     bigint,
    owner_type        text NOT NULL,
    owner_id          bigint NOT NULL,
    filename          text NOT NULL,
    original_filename text NOT NULL,
    mime_type         text NOT NULL,
    size_bytes        integer NOT NULL,
    storage_path      text NOT NULL,
    url               text NOT NULL,
    created_by        bigint,
    created_at        timestamptz NOT NULL DEFAULT now(),
    deleted_at        timestamptz
);
CREATE INDEX IF NOT EXISTS file_assets_company_owner_idx
    ON file_assets (company_id, owner_type, owner_id) WHERE deleted_at IS NULL;
ALTER TABLE file_assets
    ADD CONSTRAINT file_assets_company_fk    FOREIGN KEY (company_id)    REFERENCES companies(id),
    ADD CONSTRAINT file_assets_department_fk FOREIGN KEY (department_id) REFERENCES departments(id);
-- +goose StatementEnd

-- RLS:依 data_scope + company/department 隔離(僅定義不 ENABLE)。
-- +goose StatementBegin
DROP POLICY IF EXISTS core_file_assets_scope ON file_assets;
CREATE POLICY core_file_assets_scope ON file_assets FOR ALL
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
GRANT SELECT, INSERT, UPDATE, DELETE ON file_assets TO app_rw;
GRANT USAGE, SELECT ON file_assets_id_seq TO app_rw;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
REVOKE USAGE, SELECT ON file_assets_id_seq FROM app_rw;
REVOKE SELECT, INSERT, UPDATE, DELETE ON file_assets FROM app_rw;
DROP POLICY IF EXISTS core_file_assets_scope ON file_assets;
DROP TABLE IF EXISTS file_assets;
-- +goose StatementEnd

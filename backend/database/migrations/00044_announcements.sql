-- ANN 公告(spec announcements):banner/news/article 三型別、三層發佈範圍、上下架時間窗、
-- 平台投放(Web/App)。company_id 與 department_id 皆 NULL = 全系統公告。
-- RLS policy 僅「定義」不 ENABLE/FORCE:ENABLE+FORCE 於 00045 獨立檔(比照 00031/00033 模式)。
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS announcements (
    id             bigserial PRIMARY KEY,
    company_id     bigint,              -- NULL = 全系統公告(僅 super 可寫,服務層守衛)
    department_id  bigint,              -- NULL = 全系統或公司層公告
    type           text NOT NULL,       -- banner / news / article
    title          text NOT NULL,
    content        text NOT NULL DEFAULT '',
    image_url      text,
    link_url       text,
    publish_at     timestamptz NOT NULL DEFAULT now(),
    unpublish_at   timestamptz,         -- NULL = 不自動下架
    sort_order     integer NOT NULL DEFAULT 0,
    is_active      boolean NOT NULL DEFAULT true,
    deploy_web     boolean NOT NULL DEFAULT true,
    deploy_app     boolean NOT NULL DEFAULT true,
    created_by     bigint,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    deleted_at     timestamptz,
    CONSTRAINT announcements_type_valid CHECK (type IN ('banner', 'news', 'article'))
);
CREATE INDEX IF NOT EXISTS announcements_scope_idx
    ON announcements (company_id, department_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS announcements_type_sort_idx
    ON announcements (type, sort_order) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- RLS:三層可見性 —— 全系統列(company NULL)人人可見;公司/部門列依 data_scope 比對。
-- 寫入(WITH CHECK)比讀取嚴:'all' 之外只能寫自己公司,且非 company scope 只能寫本部門列
-- (company_admin 可寫公司層與所屬公司任一部門層 —— spec「管理權限依範圍分層」)。
-- +goose StatementBegin
DROP POLICY IF EXISTS core_announcements_scope ON announcements;
CREATE POLICY core_announcements_scope ON announcements
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR company_id IS NULL
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
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

GRANT SELECT, INSERT, UPDATE, DELETE ON announcements TO app_rw;
GRANT USAGE, SELECT ON announcements_id_seq TO app_rw;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
REVOKE USAGE, SELECT ON announcements_id_seq FROM app_rw;
REVOKE SELECT, INSERT, UPDATE, DELETE ON announcements FROM app_rw;
DROP POLICY IF EXISTS core_announcements_scope ON announcements;
DROP TABLE IF EXISTS announcements;
-- +goose StatementEnd

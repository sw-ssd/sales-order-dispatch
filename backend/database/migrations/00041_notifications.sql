-- 07 通知系統(4.3.1/4.3.5):notification_templates / notifications / user_devices / promo_tags。
-- notifications 無 deleted_at(不可刪除,規格 §5.4);其餘三表軟刪除(D10)。
-- promo_tags 僅資料層,CRUD 與選群推播屬 Phase 7 Task 7.4。
-- RLS policy 僅「定義」不 ENABLE/FORCE:ENABLE 與服務收斂同批(比照 00033-00040 模式)。
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS notification_templates (
    id               bigserial PRIMARY KEY,
    company_id       bigint NOT NULL,
    department_id    bigint,
    code             text NOT NULL,
    name             text NOT NULL,
    channel          text NOT NULL,
    subject          text,
    body             text NOT NULL,
    locale           text NOT NULL DEFAULT 'zh-Hant',
    is_active        boolean NOT NULL DEFAULT true,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS notification_templates_scope_unique
    ON notification_templates (company_id, department_id, code, channel, locale)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS notifications (
    id               bigserial PRIMARY KEY,
    company_id       bigint NOT NULL,
    department_id    bigint NOT NULL,
    user_id          bigint NOT NULL,
    template_id      bigint REFERENCES notification_templates(id),
    channel          text NOT NULL,
    title            text NOT NULL,
    content          text NOT NULL,
    payload          jsonb NOT NULL DEFAULT '{}',
    status           text NOT NULL DEFAULT 'pending',
    failure_reason   text,
    sent_at          timestamptz,
    read_at          timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS notifications_user_status_idx
    ON notifications (user_id, status, created_at);

CREATE TABLE IF NOT EXISTS user_devices (
    id               bigserial PRIMARY KEY,
    user_id          bigint NOT NULL,
    company_id       bigint NOT NULL,
    platform         text NOT NULL,
    fcm_token        text NOT NULL,
    device_name      text,
    last_seen_at     timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS user_devices_token_unique
    ON user_devices (fcm_token) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS user_devices_user_idx
    ON user_devices (user_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS promo_tags (
    id               bigserial PRIMARY KEY,
    company_id       bigint NOT NULL,
    department_id    bigint NOT NULL,
    code             text NOT NULL,
    name             text NOT NULL,
    is_active        boolean NOT NULL DEFAULT true,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS promo_tags_scope_unique
    ON promo_tags (company_id, department_id, code) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- RLS:依 data_scope + company/department 隔離(僅定義不 ENABLE)。
-- +goose StatementBegin
DROP POLICY IF EXISTS core_notification_templates_scope ON notification_templates;
CREATE POLICY core_notification_templates_scope ON notification_templates FOR ALL
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

DROP POLICY IF EXISTS core_notifications_scope ON notifications;
CREATE POLICY core_notifications_scope ON notifications FOR ALL
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

DROP POLICY IF EXISTS core_user_devices_scope ON user_devices;
CREATE POLICY core_user_devices_scope ON user_devices FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR company_id = (current_setting('app.current_company_id', true))::bigint
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR company_id = (current_setting('app.current_company_id', true))::bigint
    );

DROP POLICY IF EXISTS core_promo_tags_scope ON promo_tags;
CREATE POLICY core_promo_tags_scope ON promo_tags FOR ALL
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

-- app_rw 授權(比照 00022 明示列舉;notifications 不可刪故無 DELETE)。
GRANT SELECT, INSERT, UPDATE ON notification_templates TO app_rw;
GRANT USAGE, SELECT ON notification_templates_id_seq TO app_rw;
GRANT SELECT, INSERT, UPDATE ON notifications TO app_rw;
GRANT USAGE, SELECT ON notifications_id_seq TO app_rw;
GRANT SELECT, INSERT, UPDATE, DELETE ON user_devices TO app_rw;
GRANT USAGE, SELECT ON user_devices_id_seq TO app_rw;
GRANT SELECT, INSERT, UPDATE, DELETE ON promo_tags TO app_rw;
GRANT USAGE, SELECT ON promo_tags_id_seq TO app_rw;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
REVOKE USAGE, SELECT ON promo_tags_id_seq FROM app_rw;
REVOKE SELECT, INSERT, UPDATE, DELETE ON promo_tags FROM app_rw;
REVOKE USAGE, SELECT ON user_devices_id_seq FROM app_rw;
REVOKE SELECT, INSERT, UPDATE, DELETE ON user_devices FROM app_rw;
REVOKE USAGE, SELECT ON notifications_id_seq FROM app_rw;
REVOKE SELECT, INSERT, UPDATE ON notifications FROM app_rw;
REVOKE USAGE, SELECT ON notification_templates_id_seq FROM app_rw;
REVOKE SELECT, INSERT, UPDATE ON notification_templates FROM app_rw;
DROP POLICY IF EXISTS core_promo_tags_scope ON promo_tags;
DROP POLICY IF EXISTS core_user_devices_scope ON user_devices;
DROP POLICY IF EXISTS core_notifications_scope ON notifications;
DROP POLICY IF EXISTS core_notification_templates_scope ON notification_templates;
DROP TABLE IF EXISTS promo_tags;
DROP TABLE IF EXISTS user_devices;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS notification_templates;
-- +goose StatementEnd

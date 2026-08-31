-- role_permissions CASL 擴充(D30):02 計畫尚未實作,直接以建表 migration 含三欄與唯一鍵。
-- roles 表與 FK 由 02 計畫 2.9.1(roles_seed migration)建立/補上;此處 role_id 先以裸 bigint。
-- 唯一鍵語意:同 role×resource×action 允許多條不同 conditions 規則(以 md5 正規化)。
-- +goose Up
CREATE TABLE IF NOT EXISTS role_permissions (
    id bigserial PRIMARY KEY,
    role_id bigint NOT NULL,
    resource text NOT NULL,
    action text NOT NULL,
    conditions jsonb NULL,
    inverted boolean NOT NULL DEFAULT false,
    sort_order integer NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS role_permissions_rule_key
  ON role_permissions (role_id, resource, action, COALESCE(md5(conditions::text), ''));

-- +goose Down
DROP INDEX IF EXISTS role_permissions_rule_key;
DROP TABLE IF EXISTS role_permissions;

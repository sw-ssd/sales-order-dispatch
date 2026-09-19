-- 權限列去重索引的收斂遷移(F1 收尾):00005 只在「全新 DB」或 goose version < 5 的庫上生效,
-- 凡已記錄 version 5 的既有庫不會重跑 00005,該庫若缺 role_permissions_unique_idx 就永遠缺。
-- 本遷移以 IF NOT EXISTS 表達,是所有環境的收斂入口(索引已在的庫重跑為 no-op)。
--
-- 注意:若既有庫內已有重複列(同 role_id/resource/action 且 conditions 語意相同者,例如
-- conditions 皆為 NULL,或 jsonb 鍵序不同但內容相同),本遷移會以 23505(unique_violation)
-- 擋下升級 —— 那是正確訊號(資料本來就違反去重語意)。修法是先清理重複列再重跑
-- `go run ./cmd/migrate up`,例如(保留 id 最小者):
--   DELETE FROM role_permissions a USING role_permissions b
--    WHERE a.id > b.id
--      AND a.role_id  = b.role_id
--      AND a.resource = b.resource
--      AND a.action   = b.action
--      AND COALESCE(md5(a.conditions::text), '') = COALESCE(md5(b.conditions::text), '');
-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS role_permissions_unique_idx
    ON role_permissions (role_id, resource, action, COALESCE(md5(conditions::text), ''));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS role_permissions_unique_idx;
-- +goose StatementEnd

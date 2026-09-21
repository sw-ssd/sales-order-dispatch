-- 檔案資產啟用 RLS(D36 的第七批:file_assets,ENABLE + FORCE)。
--
-- 順序:policy(00035 定義,USING 與 WITH CHECK 同條件)+ 服務收斂(fileassets 全經
-- dbtenant.Client/TxFrom;REST handler 手動開交易並注入 ctx) → **本檔只做 ENABLE + FORCE**。
-- policy 不動。
--
-- Up/Down 對稱:Up 先 ENABLE 再 FORCE,Down 先 NO FORCE 再 DISABLE(比照 00024-00034)。
-- 冪等:ALTER TABLE … ENABLE/FORCE/DISABLE ROW LEVEL SECURITY 對已處於該狀態的表為 no-op。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE file_assets ENABLE ROW LEVEL SECURITY;
ALTER TABLE file_assets FORCE  ROW LEVEL SECURITY;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE file_assets NO FORCE ROW LEVEL SECURITY;
ALTER TABLE file_assets DISABLE ROW LEVEL SECURITY;
-- +goose StatementEnd

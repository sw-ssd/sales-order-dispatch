-- 列印記錄啟用 RLS(D36 的第八批:print_logs / print_previews,ENABLE + FORCE)。
--
-- 順序:policy(00037 定義,USING 與 WITH CHECK 同條件)+ 服務收斂(PrintService 全經
-- dbtenant.Client/TxFrom;Assemble 交易內讀取) → **本檔只做 ENABLE + FORCE**。
-- policy 不動。
--
-- Up/Down 對稱:Up 先 ENABLE 再 FORCE,Down 先 NO FORCE 再 DISABLE(比照 00024-00036)。
-- 冪等:ALTER TABLE … ENABLE/FORCE/DISABLE ROW LEVEL SECURITY 對已處於該狀態的表為 no-op。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE print_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE print_logs FORCE  ROW LEVEL SECURITY;
ALTER TABLE print_previews ENABLE ROW LEVEL SECURITY;
ALTER TABLE print_previews FORCE  ROW LEVEL SECURITY;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE print_previews NO FORCE ROW LEVEL SECURITY;
ALTER TABLE print_previews DISABLE ROW LEVEL SECURITY;
ALTER TABLE print_logs NO FORCE ROW LEVEL SECURITY;
ALTER TABLE print_logs DISABLE ROW LEVEL SECURITY;
-- +goose StatementEnd

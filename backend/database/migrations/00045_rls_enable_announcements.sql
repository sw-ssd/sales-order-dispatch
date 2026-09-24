-- 公告啟用 RLS(announcements,ENABLE + FORCE)。
--
-- 順序:policy(00044 定義) + 服務收斂(AnnouncementService 全經 dbtenant.Client/TxFrom;
-- scope 由 middleware 注入) → **本檔只做 ENABLE + FORCE**。policy 不動。
--
-- Up/Down 對稱:Up 先 ENABLE 再 FORCE,Down 先 NO FORCE 再 DISABLE(比照 00024-00042)。
-- 冪等:ALTER TABLE … ENABLE/FORCE/DISABLE ROW LEVEL SECURITY 對已處於該狀態的表為 no-op。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE announcements ENABLE ROW LEVEL SECURITY;
ALTER TABLE announcements FORCE  ROW LEVEL SECURITY;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE announcements NO FORCE ROW LEVEL SECURITY;
ALTER TABLE announcements DISABLE ROW LEVEL SECURITY;
-- +goose StatementEnd

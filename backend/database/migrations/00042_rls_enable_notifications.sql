-- 通知系統啟用 RLS(07 Task 4.3.1 收尾:notification_templates / notifications /
-- user_devices / promo_tags,ENABLE + FORCE)。
--
-- 順序:policy(00041 定義) + 服務收斂(Notification/DeviceService 全經
-- dbtenant.Client/TxFrom;scope 由 middleware 注入) → **本檔只做 ENABLE + FORCE**。
-- policy 不動。
--
-- Up/Down 對稱:Up 先 ENABLE 再 FORCE,Down 先 NO FORCE 再 DISABLE(比照 00024-00040)。
-- 冪等:ALTER TABLE … ENABLE/FORCE/DISABLE ROW LEVEL SECURITY 對已處於該狀態的表為 no-op。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE notification_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_templates FORCE  ROW LEVEL SECURITY;
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE notifications FORCE  ROW LEVEL SECURITY;
ALTER TABLE user_devices ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_devices FORCE  ROW LEVEL SECURITY;
ALTER TABLE promo_tags ENABLE ROW LEVEL SECURITY;
ALTER TABLE promo_tags FORCE  ROW LEVEL SECURITY;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE promo_tags NO FORCE ROW LEVEL SECURITY;
ALTER TABLE promo_tags DISABLE ROW LEVEL SECURITY;
ALTER TABLE user_devices NO FORCE ROW LEVEL SECURITY;
ALTER TABLE user_devices DISABLE ROW LEVEL SECURITY;
ALTER TABLE notifications NO FORCE ROW LEVEL SECURITY;
ALTER TABLE notifications DISABLE ROW LEVEL SECURITY;
ALTER TABLE notification_templates NO FORCE ROW LEVEL SECURITY;
ALTER TABLE notification_templates DISABLE ROW LEVEL SECURITY;
-- +goose StatementEnd

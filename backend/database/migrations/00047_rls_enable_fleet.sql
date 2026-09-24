-- fleet 執行層啟用 RLS(D32/10.1:fleet_drivers / vehicles / fleet_deliveries,ENABLE + FORCE)。
--
-- 順序:policy(00046 定義) + 服務收斂(FleetService 全經 dbtenant.Client/TxFrom;
-- scope 由 middleware 注入) → **本檔只做 ENABLE + FORCE**。policy 不動。
--
-- Up/Down 對稱:Up 先 ENABLE 再 FORCE,Down 先 NO FORCE 再 DISABLE(比照 00024-00045)。
-- 冪等:ALTER TABLE … ENABLE/FORCE/DISABLE ROW LEVEL SECURITY 對已處於該狀態的表為 no-op。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE fleet_drivers ENABLE ROW LEVEL SECURITY;
ALTER TABLE fleet_drivers FORCE  ROW LEVEL SECURITY;
ALTER TABLE vehicles ENABLE ROW LEVEL SECURITY;
ALTER TABLE vehicles FORCE  ROW LEVEL SECURITY;
ALTER TABLE fleet_deliveries ENABLE ROW LEVEL SECURITY;
ALTER TABLE fleet_deliveries FORCE  ROW LEVEL SECURITY;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE fleet_deliveries NO FORCE ROW LEVEL SECURITY;
ALTER TABLE fleet_deliveries DISABLE ROW LEVEL SECURITY;
ALTER TABLE vehicles NO FORCE ROW LEVEL SECURITY;
ALTER TABLE vehicles DISABLE ROW LEVEL SECURITY;
ALTER TABLE fleet_drivers NO FORCE ROW LEVEL SECURITY;
ALTER TABLE fleet_drivers DISABLE ROW LEVEL SECURITY;
-- +goose StatementEnd

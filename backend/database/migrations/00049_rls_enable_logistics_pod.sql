-- logistics 執行軌跡與 POD 啟用 RLS(D32/10.6:logistics_delivery_events / logistics_proofs,
-- ENABLE + FORCE)。
--
-- 順序:policy(00048 定義) + 服務收斂(全經 dbtenant.Client/TxFrom;scope 由 middleware 注入)
-- → **本檔只做 ENABLE + FORCE**。policy 不動。
--
-- Up/Down 對稱:Up 先 ENABLE 再 FORCE,Down 先 NO FORCE 再 DISABLE(比照 00024-00047)。
-- 冪等:ALTER TABLE … ENABLE/FORCE/DISABLE ROW LEVEL SECURITY 對已處於該狀態的表為 no-op。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE logistics_delivery_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE logistics_delivery_events FORCE  ROW LEVEL SECURITY;
ALTER TABLE logistics_proofs ENABLE ROW LEVEL SECURITY;
ALTER TABLE logistics_proofs FORCE  ROW LEVEL SECURITY;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE logistics_proofs NO FORCE ROW LEVEL SECURITY;
ALTER TABLE logistics_proofs DISABLE ROW LEVEL SECURITY;
ALTER TABLE logistics_delivery_events NO FORCE ROW LEVEL SECURITY;
ALTER TABLE logistics_delivery_events DISABLE ROW LEVEL SECURITY;
-- +goose StatementEnd

-- 客戶專屬清單啟用 RLS(D36 的第六批:customer_products,ENABLE + FORCE)。
--
-- 順序:policy(00033 定義,USING 與 WITH CHECK 同條件)+ 服務收斂(CustomerProductService 全經
-- dbtenant.Client/TxFrom;下單接線走請求交易) → **本檔只做 ENABLE + FORCE**。policy 不動。
--
-- Up/Down 對稱:Up 先 ENABLE 再 FORCE,Down 先 NO FORCE 再 DISABLE(比照 00024-00032)。
-- 冪等:ALTER TABLE … ENABLE/FORCE/DISABLE ROW LEVEL SECURITY 對已處於該狀態的表為 no-op。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE customer_products ENABLE ROW LEVEL SECURITY;
ALTER TABLE customer_products FORCE  ROW LEVEL SECURITY;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE customer_products NO FORCE ROW LEVEL SECURITY;
ALTER TABLE customer_products DISABLE ROW LEVEL SECURITY;
-- +goose StatementEnd

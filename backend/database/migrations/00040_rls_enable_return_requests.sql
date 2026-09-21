-- 退貨申請啟用 RLS(06 Task 4.7.1 收尾:return_requests / return_request_items,ENABLE + FORCE)。
--
-- 順序:policy(00039 定義,return_requests 含 self→customer_id 分支) + 服務收斂
-- (ReturnService 全經 dbtenant.Client/TxFrom;客戶 self 由 middleware scope 注入)
-- → **本檔只做 ENABLE + FORCE**。policy 不動。
--
-- Up/Down 對稱:Up 先 ENABLE 再 FORCE,Down 先 NO FORCE 再 DISABLE(比照 00024-00038)。
-- 冪等:ALTER TABLE … ENABLE/FORCE/DISABLE ROW LEVEL SECURITY 對已處於該狀態的表為 no-op。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE return_requests ENABLE ROW LEVEL SECURITY;
ALTER TABLE return_requests FORCE  ROW LEVEL SECURITY;
ALTER TABLE return_request_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE return_request_items FORCE  ROW LEVEL SECURITY;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE return_request_items NO FORCE ROW LEVEL SECURITY;
ALTER TABLE return_request_items DISABLE ROW LEVEL SECURITY;
ALTER TABLE return_requests NO FORCE ROW LEVEL SECURITY;
ALTER TABLE return_requests DISABLE ROW LEVEL SECURITY;
-- +goose StatementEnd

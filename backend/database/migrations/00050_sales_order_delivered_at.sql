-- 10.9 送達回寫:訂單記錄實際送達時點。
-- 以 delivered_at 表達「已實際送達」(非空即送達),status 仍沿用 05 的既有狀態機
-- (processing → completed),不新增狀態值(避免第二套狀態語意)。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS delivered_at timestamptz;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sales_orders DROP COLUMN IF EXISTS delivered_at;
-- +goose StatementEnd

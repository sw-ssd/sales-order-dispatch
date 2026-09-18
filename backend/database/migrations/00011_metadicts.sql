-- 字典檔表（03 計畫 2.5.1, D10/D11）。單表兩層：
--   department_id IS NULL = 系統預設；非 NULL = 部門擴充。
-- 軟刪除(deleted_at) + 兩道部分唯一索引，避開 PostgreSQL NULL 不參與唯一比較的陷阱(D10)。
-- RLS policy 比照 00010：僅「定義」不 ENABLE/FORCE（services 尚無每請求交易套用點，D3 待接線）。
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS metadicts (
    id            bigserial PRIMARY KEY,
    type          text NOT NULL,
    code          text NOT NULL,
    display_name  text NOT NULL,
    department_id bigint,
    sort_order    integer NOT NULL DEFAULT 0,
    is_active     boolean NOT NULL DEFAULT true,
    deleted_at    timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS metadicts_type_dept_active_idx
    ON metadicts (type, department_id, is_active);
-- 部分唯一索引(兩道)：系統預設與部門擴充各自獨立去重(D10/D11)。
CREATE UNIQUE INDEX IF NOT EXISTS metadicts_sys_unique
    ON metadicts (type, code) WHERE department_id IS NULL AND deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS metadicts_dept_unique
    ON metadicts (type, code, department_id) WHERE department_id IS NOT NULL AND deleted_at IS NULL;
-- +goose StatementEnd

-- RLS：讀取允許 data_scope=all 或系統預設或等於當前部門；寫入僅 all(super)或部門級擴充。
-- 同 00010,本波僅「定義」policy 不 ENABLE/FORCE RLS;待每請求交易層 SET LOCAL app.* 後再啟用。
-- +goose StatementBegin
DROP POLICY IF EXISTS core_metadicts_scope ON metadicts;
CREATE POLICY core_metadicts_scope ON metadicts
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR department_id IS NULL
        OR department_id = (current_setting('app.current_department_id', true))::bigint
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR department_id = (current_setting('app.current_department_id', true))::bigint
    );
-- +goose StatementEnd

-- 系統預設字典 seed（冪等：以系統級 type+code 是否存在為準，缺則補、有則略）。
-- order_source 為系統級固定(W=Web 中台 / A=App)，供 05-sales-orders 取號，API 不可異動。
-- +goose StatementBegin
INSERT INTO metadicts (type, code, display_name, department_id, sort_order, is_active)
SELECT v.type, v.code, v.display_name, NULL, v.sort_order, true
FROM (VALUES
    ('unit',           'KG',  '公斤',     10),
    ('unit',           'G',   '公克',     20),
    ('unit',           'PCS', '件',       30),
    ('unit',           'BOX', '盒',       40),
    ('unit',           'PKG', '包',       50),
    ('payment_method', 'CASH','現金',     10),
    ('payment_method', 'CARD','信用卡',   20),
    ('payment_method', 'TRANSFER','轉帳', 30),
    ('settlement_method','MONTH','月結',  10),
    ('settlement_method','QUARTER','季結',20),
    ('customer_type',  'WHOLESALE','批發',10),
    ('customer_type',  'RETAIL','零售',   20),
    ('invoice_type',   'UNIFORM','統一發票',10),
    ('invoice_type',   'RECEIPT','收據',  20),
    ('order_source',   'W',   'Web 中台', 10),
    ('order_source',   'A',   'App',      20)
) AS v(type, code, display_name, sort_order)
WHERE NOT EXISTS (
    SELECT 1 FROM metadicts m
    WHERE m.type = v.type AND m.code = v.code AND m.department_id IS NULL AND m.deleted_at IS NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS metadicts;
-- +goose StatementEnd

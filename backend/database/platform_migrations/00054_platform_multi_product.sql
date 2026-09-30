-- +goose Up
-- platform 多產品軸(§ design/2026-09-29-multi-product-key-strategy.md, 決策 B=events.aggregate_id 改 uuid)。
-- 依賴 00053(companies.internal_id 已存在)。
-- 現況:platform schema 所有 company_id 皆裸 bigint 無 FK,跨產品會互撞。改為 (product_id, company_internal_id uuid)。

-- 1) 產品註冊表
CREATE TABLE IF NOT EXISTS platform.products (
    id          text        PRIMARY KEY,          -- 'sales-order' / 'crm' / 'billing'
    type        text        NOT NULL,
    name        text        NOT NULL,
    owner       text        NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- 2) subscriptions: 加 product_id + company_internal_id(公司級, 多產品索引) + internal_id(訂閱級, events.aggregate_id 引用)。
--    company_internal_id / internal_id 皆帶 DEFAULT gen_random_uuid(): 回填時免改既有 raw INSERT。
ALTER TABLE platform.subscriptions
    ADD COLUMN product_id          text NOT NULL DEFAULT 'sales-order',
    ADD COLUMN company_internal_id uuid DEFAULT gen_random_uuid(),
    ADD COLUMN internal_id         uuid DEFAULT gen_random_uuid();

UPDATE platform.subscriptions s
    SET company_internal_id = c.internal_id
    FROM companies c
    WHERE s.company_id = c.id;

ALTER TABLE platform.subscriptions
    ALTER COLUMN company_internal_id SET NOT NULL,
    ALTER COLUMN internal_id SET NOT NULL;
-- 保留既有 subscriptions_active_company_unique(company_id, 部分唯一)供 CreateSubscriptionTx 的
-- ON CONFLICT(company_id) 使用; 本 migration 僅「新增」多產品索引, 不刪舊的(單 product 環境兩者等價)。
CREATE UNIQUE INDEX IF NOT EXISTS subscriptions_active_unique
    ON platform.subscriptions (product_id, company_internal_id) WHERE status <> 'cancelled';

CREATE UNIQUE INDEX IF NOT EXISTS subscriptions_internal_id_unique
    ON platform.subscriptions (internal_id);

-- 3) tenant_overrides: 同構
ALTER TABLE platform.tenant_overrides
    ADD COLUMN product_id          text NOT NULL DEFAULT 'sales-order',
    ADD COLUMN company_internal_id uuid DEFAULT gen_random_uuid();

UPDATE platform.tenant_overrides o
    SET company_internal_id = c.internal_id
    FROM companies c
    WHERE o.company_id = c.id;

ALTER TABLE platform.tenant_overrides
    ALTER COLUMN company_internal_id SET NOT NULL;
-- 保留既有 tenant_overrides_active_unique(company_id, feature_code, 部分唯一)供 admin_writes 的
-- ON CONFLICT(company_id, feature_code) 使用; 本 migration 僅「新增」多產品索引, 不刪舊的。
CREATE UNIQUE INDEX IF NOT EXISTS tenant_overrides_active_unique
    ON platform.tenant_overrides (product_id, company_internal_id, feature_code) WHERE revoked_at IS NULL;


-- 4) events: 加 product_id; aggregate_id bigint 改 uuid。
--    先加 uuid 欄並回填(subscription 經 subscriptions.internal_id; 其餘隨機), 再替換原欄;
--    因 bigint 不能直接 USING 轉 uuid(訂閱的 id 不是 uuid 文字)。
ALTER TABLE platform.events
    ADD COLUMN product_id text NOT NULL DEFAULT 'sales-order';

UPDATE platform.events e
    SET product_id = s.product_id
    FROM platform.subscriptions s
    WHERE e.aggregate_type = 'subscription'
      AND e.aggregate_id = s.id;

ALTER TABLE platform.events ADD COLUMN aggregate_id_uuid uuid;

UPDATE platform.events e
    SET aggregate_id_uuid = s.internal_id
    FROM platform.subscriptions s
    WHERE e.aggregate_type = 'subscription'
      AND e.aggregate_id = s.id;

UPDATE platform.events
    SET aggregate_id_uuid = gen_random_uuid()
    WHERE aggregate_id_uuid IS NULL;

ALTER TABLE platform.events DROP COLUMN aggregate_id;

ALTER TABLE platform.events RENAME COLUMN aggregate_id_uuid TO aggregate_id;

-- 5) 種子銷售單產品(其餘產品由各產品上線時 INSERT platform.products)
INSERT INTO platform.products (id, type, name, owner)
VALUES ('sales-order', 'sales-dispatch', '銷售派單', 'platform')
    ON CONFLICT (id) DO NOTHING;

-- +goose Down
-- 注意: uuid -> bigint 不可逆(訂閱 uuid 非 bigint 文字); 回滾後 aggregate_id 僅作占位,
-- 事件語意失效(屬 migration down 的極少使用路徑, 不保證可讀)。
ALTER TABLE platform.events
    ALTER COLUMN aggregate_id TYPE bigint
    USING 0;

ALTER TABLE platform.events DROP COLUMN IF EXISTS product_id;

DROP INDEX IF EXISTS platform.tenant_overrides_active_unique;
ALTER TABLE platform.tenant_overrides DROP COLUMN IF EXISTS product_id, DROP COLUMN IF EXISTS company_internal_id;

DROP INDEX IF EXISTS platform.subscriptions_active_unique;
DROP INDEX IF EXISTS platform.subscriptions_internal_id_unique;
ALTER TABLE platform.subscriptions DROP COLUMN IF EXISTS product_id, DROP COLUMN IF EXISTS company_internal_id, DROP COLUMN IF EXISTS internal_id;

DROP TABLE IF EXISTS platform.products;

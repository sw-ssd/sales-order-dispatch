-- +goose Up
-- 業務表三鍵策略(§ design/2026-09-29-multi-product-key-strategy.md, 決策 A=全統一):
-- 每張業務實體表加 internal_id(uuid, platform 通用引用, 不可變) + external_id(uuid, 第三方整合暴露, 可輪換)。
-- 既有主鍵 id bigint 保留不動;RLS GUC(app.current_company_id 等)不受影響。
-- ent 不認得這兩欄,INSERT 時由 DB 預設值填寫,ent 不選它們 → 無 codegen break。
-- 不在此加 UNIQUE:internal_id/external_id 由 gen_random_uuid() 預設值保證每列唯一,DB 層 UNIQUE 為 YAGNI,
-- 且 companies/departments 的軟刪除設計(00019/00020)刻意移除表層 UNIQUE 以允許 identifier 重用 ——
-- 加 UNIQUE 會直接撞該測試。唯一真正需要 UNIQUE 的是 platform.subscriptions.internal_id(events.aggregate_id
-- 引用),由 00054 的 subscriptions_internal_id_unique 負責。

-- pgcrypto 確保(PG<13 需要;00001 建完即 DROP,此處重建以相容舊版)。
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 39 張業務表統一加雙鍵。NOT NULL + DEFAULT gen_random_uuid() → ALTER 時自動回填既有列。
ALTER TABLE announcements      ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE audit_logs         ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE companies          ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE customer_addresses ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE customer_contacts  ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE customer_counters  ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE customer_products   ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE customers          ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE departments        ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE file_assets        ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE logistics_deliveries       ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE logistics_delivery_events  ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE logistics_drivers          ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE logistics_proofs           ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE metadicts          ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE notification_templates ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE notifications      ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE order_counters     ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE print_logs         ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE print_previews     ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE processing_specs   ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE product_categories ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE product_processing_specs ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE product_units      ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE products           ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE promo_tags         ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE return_request_items ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE return_requests    ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE role_permissions   ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE roles              ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE routes             ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE sales_order_events ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE sales_order_items  ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE sales_orders       ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE user_devices       ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE users              ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE vehicles           ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE warehouses         ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid(), ADD COLUMN external_id uuid NOT NULL DEFAULT gen_random_uuid();

-- +goose Down
ALTER TABLE announcements      DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE audit_logs         DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE companies          DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE customer_addresses DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE customer_contacts  DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE customer_counters  DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE customer_products   DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE customers          DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE departments        DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE file_assets        DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE logistics_deliveries       DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE logistics_delivery_events  DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE logistics_drivers          DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE logistics_proofs           DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE metadicts          DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE notification_templates DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE notifications      DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE order_counters     DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE print_logs         DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE print_previews     DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE processing_specs   DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE product_categories DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE product_processing_specs DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE product_units      DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE products           DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE promo_tags         DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE return_request_items DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE return_requests    DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE role_permissions   DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE roles              DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE routes             DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE sales_order_events DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE sales_order_items  DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE sales_orders       DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE user_devices       DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE users              DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE vehicles           DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
ALTER TABLE warehouses         DROP COLUMN IF EXISTS internal_id, DROP COLUMN IF EXISTS external_id;
DROP EXTENSION IF EXISTS pgcrypto;

-- 商品域啟用 RLS(D36 的第四批:products / product_units / product_processing_specs,ENABLE + FORCE)。
--
-- 順序:policy(00017 定義、00023 補 WITH CHECK、00025 正規化為 NULLIF)→ **本檔只做 ENABLE + FORCE**。
-- policy 不動:商品三表的 policy 已由 00023 建立並在 00025 的 Up 全數重建為
-- `NULLIF(current_setting('app.current_*', true), '')` 形式;本檔再寫一次只會製造第二份定義。
--
-- 為何收斂必須早於本檔(product_service.go 已於同一波完成):
--   product_service.go 原本以裸 `s.db` 讀 product_categories / warehouses / processing_specs
--   (00025 之後已 fail-closed),商品新增／更新時的引用驗證會回 invalid_argument;該檔的
--   11 處讀取 + 4 處自開交易已全數收斂到 `dbtenant.Client(ctx, s.db)` / `dbtenant.TxFrom(ctx)`
--   (含 `validateUnitCode` 讀 metadicts —— 該表由後續批次啟用,同一支檔的存取必須一起收斂)。
--
-- 啟用後的行為(呼叫端必須配合):
--   1. **未帶 scope 的查詢一律 fail-closed(0 列)**;服務層任何查詢都必須走
--      `dbtenant.Client(ctx, s.db)`(請求交易內已由 driver 裝飾器 SET LOCAL)。漏掛 → 不是權限
--      錯誤,是「查不到資料」(0 列)。
--   2. FORCE 讓 **table owner** 也受 policy 約束(PG 的 superuser 仍永遠繞過 RLS,FORCE 亦然)。
--      後果:日後對本域做資料回填的 migration 或 seed,必須先
--      `SET LOCAL app.current_data_scope = 'all'`(`dbtenant.SystemScopeTx`),否則以 owner
--      身分寫入會被 WITH CHECK 擋下。
--   3. **子表傳遞性**:product_units / product_processing_specs 沒有 company_id / department_id,
--      policy 以 `EXISTS (SELECT 1 FROM products p WHERE p.id = <子表>.product_id)` 表達可見性;
--      子查詢本身也受 products 的 policy 約束(同一角色),故「父商品不可見 → 子列不可見」,
--      把子列掛到他人的商品上同樣被 WITH CHECK 擋下。兩張子表的 DML 與序列權限已由 00022
--      授予 app_rw,不需額外 GRANT。
--
-- Up/Down 對稱:Up 先 ENABLE 再 FORCE,Down 必須**先 NO FORCE 再 DISABLE**
--(反序雖然也能跑,但會留下一段「ENABLE 且仍然 FORCE」的中間狀態;DISABLE 本身即清除
-- FORCE 旗標,故先 NO FORCE 讓 Down 的每一步都回到上一個穩定狀態)。
--
-- 冪等:ALTER TABLE … ENABLE/FORCE/DISABLE ROW LEVEL SECURITY 對已處於該狀態的表為 no-op,
-- 故 Up、Down 皆可重複套用(比照 00024/00025 的冪等慣例)。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE products                  ENABLE ROW LEVEL SECURITY;
ALTER TABLE products                  FORCE  ROW LEVEL SECURITY;
ALTER TABLE product_units             ENABLE ROW LEVEL SECURITY;
ALTER TABLE product_units             FORCE  ROW LEVEL SECURITY;
ALTER TABLE product_processing_specs  ENABLE ROW LEVEL SECURITY;
ALTER TABLE product_processing_specs  FORCE  ROW LEVEL SECURITY;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products                  NO FORCE ROW LEVEL SECURITY;
ALTER TABLE products                  DISABLE ROW LEVEL SECURITY;
ALTER TABLE product_units             NO FORCE ROW LEVEL SECURITY;
ALTER TABLE product_units             DISABLE ROW LEVEL SECURITY;
ALTER TABLE product_processing_specs  NO FORCE ROW LEVEL SECURITY;
ALTER TABLE product_processing_specs  DISABLE ROW LEVEL SECURITY;
-- +goose StatementEnd

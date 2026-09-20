-- customer_counters 對齊 ent schema(2026-09-20,E3):補 id bigserial 主鍵,company_id 的
-- 「每公司一列」改由唯一性表達。
--
-- 背景(可達缺陷,e1-report §7 的殘餘):00013 建出的 customer_counters 以 company_id 為主鍵、
-- **沒有 id 欄位**,而 ent schema 把 company_id 宣告為一般欄位(field.Int("company_id").Unique())
-- → ent 的隱含 id 才是主鍵(ent/migrate/schema.go 的 CustomerCountersColumns[0] 為 id 且為
-- PrimaryKey)。於真 PG 上 ent 對該表的所有取用(Create/Exist/Only/Update)一律
-- `ERROR: column "id" does not exist (SQLSTATE 42703)`;而 CreateCustomer 的必要步驟
-- ensureCustomerCounter(Exist/Create)與 nextCustomerCode(Only/Update)都經 ent
-- → **任何走到取號的 CreateCustomer 在現行 PG 部署必定失敗**(公開 API 上的阻斷級缺陷)。
-- sqlite(enttest)踩不到:ent 依 schema 自建含 id 的表,不是 goose 的 00013。
--
-- 修法(本 repo 慣例是「遷移對齊 ent」,而非改 ent schema 去將就既有 DDL —— 00005 檔頭即
-- 「欄位型別/唯一性/外鍵對齊 ent migrate schema」):為 customer_counters 補 id bigserial 主鍵,
-- 原 company_id 主鍵讓位後改以唯一性約束表達「每公司一列」——名稱沿用 ent 的 PostgreSQL
-- dialect 為 Unique 欄位生成的 `<表>_<欄>_key`(ent/dialect/sql/schema/postgres.go),使 ent 日後
-- 跑 Schema.Create 時不會想再建一次(ent 於該 dialect 對 Unique 欄位建的正是同名唯一索引)。
--
-- 對既有資料的影響:id 由 sequence 產生(bigserial;PG 對帶 volatile 預設的新欄會改寫整表並為
-- 每列取一個新號),故既有列不會有 NULL、主鍵可安全建立;列順序對應的 id 值不保證對應任何外部
-- 識別碼,但該表無任何 FK 指向它(僅 CreateCustomer 內部使用,不對外暴露 RPC),
-- 故改主鍵不影響其他表的關聯;company_id 的唯一性與 next_seq/version 的樂觀鎖語意完全不變。
--
-- 冪等(比照 00008/00010/00019/00020 的 IF EXISTS + DO $$ 模式):對「已跑過 00013 的既有 DB」
-- 與「全新 DB(00013 緊接 00021)」皆同形,重複套用為 no-op(實測 `ADD COLUMN IF NOT EXISTS` 在
-- 欄位已存在時不會重複建立 sequence)。Up/Down 對稱:Up 去 company_id 主鍵 → 補 id 主鍵 →
-- 補 company_id 唯一;Down 完全反向,還原為 00013 的原貌(company_id 主鍵、無 id 欄位)。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE customer_counters DROP CONSTRAINT IF EXISTS customer_counters_pkey;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE customer_counters ADD COLUMN IF NOT EXISTS id bigserial;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'customer_counters_pkey' AND conrelid = 'customer_counters'::regclass
    ) THEN
        ALTER TABLE customer_counters ADD CONSTRAINT customer_counters_pkey PRIMARY KEY (id);
    END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS customer_counters_company_id_key ON customer_counters (company_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS customer_counters_company_id_key;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE customer_counters DROP CONSTRAINT IF EXISTS customer_counters_pkey;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE customer_counters DROP COLUMN IF EXISTS id;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'customer_counters_pkey' AND conrelid = 'customer_counters'::regclass
    ) THEN
        ALTER TABLE customer_counters ADD CONSTRAINT customer_counters_pkey PRIMARY KEY (company_id);
    END IF;
END $$;
-- +goose StatementEnd

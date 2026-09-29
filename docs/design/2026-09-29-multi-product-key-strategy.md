# 多產品鍵策略（internal_id / external_id）

> 狀態：設計草案（2026-09-29）
> 關聯：platform 抽離為多類型產品控制平面（§86 討論）、`docs/superpowers/specs/1.0-requirements/multi-tenancy/spec.md`、`docs/superpowers/specs/architecture/saas-billing-entitlements.md`
> 決策前提：platform 要服務 **N 個不同類型的產品**（銷售單 / CRM / 計費 SaaS…），平台不認得各產品的 `company_id` 型別與 schema。

## 1. 三鍵定義

每張被 platform 引用、或對外暴露給第三方整合的實體表，攜帶三把鍵：

| 鍵 | 型別 | 生成方 | 跨邊界可見 | 可否輪換 |
|---|---|---|---|---|
| `id` | `bigint` / `bigserial` | 產品（現況不變） | 否 | 否（主鍵） |
| `internal_id` | `uuid` | 產品寫入時 | 是（platform 的通用引用鍵） | 否（不可變，等同主鍵別名） |
| `external_id` | `uuid` | 產品寫入時 | 是（第三方整合暴露值） | 可（輪換不改 `id`/`internal_id`） |

**關鍵不變式**：`company_id bigint` 與所有現有 RLS 欄位 **保留原狀**。platform 引用的是新增的
`company_internal_id uuid`，不是 `company_id`。因此 `app.current_company_id` GUC 與既有的 RLS
policy（`internal/auth/rls.go:62` 的 `SET LOCAL app.current_company_id`）**完全不受影響**——
這與前一輪「company_id 改 text」的說法不同，三鍵方案下不需要改 RLS。

## 2. 適用範圍（最小集 vs 全表）

- **最小集（platform 多產品必需）**：`companies`（tenant 根）。platform schema 的
  `subscriptions` / `tenant_overrides` / `events` 改引用 `companies.internal_id`，而非 `companies.id`。
- **對外整合集**：任何經 OAuth / Webhook / OpenFGA / 三方 API 暴露的實體（目前無具體清單）。
- **全表統一（使用者原始要求）**：50 張業務表皆加雙鍵。代價最大（見 §6），但換來「任何表皆可被
  platform 或外部服務引用」的均一性。

**決策點 A**：只做最小集（companies + platform 引用），還是全表統一？
初步建議：最小集先落地（platform 拓撲能跑），全表統一作為後續每張新表的預設慣例 +
既有表視整合需求漸進補。避免一次性 50 表遷移。

## 3. 生成規則

```
internal_id uuid NOT NULL DEFAULT gen_random_uuid()
external_id  uuid NOT NULL DEFAULT gen_random_uuid()
```

- **時機**：實體 `INSERT` 時由 DB 預設值產生（不依賴應用層）。理由：應用層若漏設，跨邊界引用會
  缺鍵；DB 預設值保證不變式。
- **確定性**：`gen_random_uuid()` 隨機，不可推導、不洩漏順序/數量。
- **pgcrypto 陷阱**：`00001_init_schema.sql:5` 建立 `pgcrypto` 後在 `:10` **立即 DROP**。
  後續 migration 用 `gen_random_uuid()` 作預設值會在 PG<13 環境失敗。處理：在 `0003x` 開頭
  `CREATE EXTENSION IF NOT EXISTS pgcrypto;`（PG13+ 該函數已進 core，但仍顯式建立以相容舊版）；
  Down 段對應 `DROP EXTENSION`（僅當本 migration 建立時）。
- **回填**：`ALTER TABLE <t> ADD COLUMN internal_id uuid NOT NULL DEFAULT gen_random_uuid()` 在 ALTER
  當下即自動回填既有列（NOT NULL + DEFAULT 由 PG 就地寫入），**不需**手動 `UPDATE`。Down 段
  `DROP COLUMN` 即丟棄這兩欄與其回填值。

## 4. 輪換語意

- `internal_id`：不可變。等同主鍵別名，用途是「platform 不依賴產品主鍵型別」。
- `external_id`：**可輪換**。洩漏或第三方整合終止時，產品重發 `external_id`，舊值失效；
  `id` / `internal_id` 不變，RLS 與內部引用不受影響。輪換須走產品側事務（同筆更新 + 失效舊 token）。
- 兩者**不加 DB 層 UNIQUE**（`gen_random_uuid()` 預設值已保證每列唯一；UNIQUE 為 YAGNI，且會與
  companies/departments 的軟刪除設計 00019/00020 直接衝突——該設計刻意移除表層 UNIQUE 以允許
  identifier 重用）。唯一需要 UNIQUE 的是 `platform.subscriptions.internal_id`（events.aggregate_id
  引用），由 00054 的 `subscriptions_internal_id_unique` 負責。語意分離仍成立：攻擊者拿到 `external_id`
  也跨不進 RLS scope（scope 綁 `company_id` / `internal_id`，不綁 `external_id`）。

## 5. platform 側對映

```
platform.products(id text PK, type, name, owner, created_at)

platform.subscriptions
  product_id            text  NOT NULL DEFAULT 'sales-order'
  company_internal_id   uuid  NOT NULL          -- 取代 company_id bigint
  ...
  UNIQUE (product_id, company_internal_id) WHERE status <> 'cancelled'

platform.tenant_overrides
  product_id text NOT NULL, company_internal_id uuid NOT NULL, feature_code text ...
  UNIQUE (product_id, company_internal_id, feature_code) WHERE revoked_at IS NULL

platform.events
  product_id text NOT NULL DEFAULT 'sales-order'
  aggregate_id uuid NOT NULL   -- 原 bigint，改 uuid（泛型跨產品引用）
  aggregate_type text NOT NULL
```

- platform **不驗證** `company_internal_id` 指向什麼；它只是 `(product_id, company_internal_id)`
  複合鍵的一部分。無 FK（與現況 `subscriptions.company_id bigint` 無 FK 一致，見 00029）。
- `consumer.SetCompanyStatus` 簽章改 `(productID, companyInternalID uuid, status)`；產品側 worker
  收到後自行用 `companies.internal_id` 反查 `id` 寫入。平台不碰業務表。

## 6. 成本（誠實清單）

- 最小集：1 表加雙鍵 + 3 張 platform 表改引用 + 1 個 extension 指令。機械、低風險。
- 全表統一：50 表 × 2 欄 + `ent` codegen 擴 50 個 field（現 `ent` 無 uuid 欄位，僅 `fileasset.go`
  有自訂 filename）+ 回填。量大但無腦；可獨立 PR。
- `platform.events.aggregate_id` 由 bigint 改 uuid：`billing.go:515` 寫入點與 `:458` JOIN 需同步改
  型別；影響 consumer 讀取（已是 `[]store.Event`，結構改 uuid 即可）。

## 7. 待決事項

- **決策點 A**：最小集 vs 全表統一（§2）。
- **決策點 B**：`platform.events.aggregate_id` 是否跟進改 uuid（建議改，否則異類產品的
  subscription id 型別不一致會讓 outbox 泛型失效）。
- **決策點 C**：`external_id` 輪換的觸發權限（誰能輪換、是否需 audit）。初步：僅 `super` / 該產品
  `company_admin`，輪換寫平台 `audit_logs`。
- **PG 版本**：確認部署 PG ≥ 13（`gen_random_uuid` 進 core），否則 `0003x` 必須 `CREATE EXTENSION pgcrypto`。

## 8. 下一步

決策點 A/B/C 定案後，產出 `0003x_platform_multi_product.sql`（platform 側產品軸 + `company_internal_id`
uuid + 索引重建）。業務表雙鍵與 `ent` codegen 為獨立 PR（§6）。

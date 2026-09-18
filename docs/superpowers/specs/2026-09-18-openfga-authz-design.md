# 授權引擎對齊 D32 — 內嵌 OpenFGA + RLS 設計

> 狀態：✅ 已定案（2026-09-18）
> 性質：本波子專案（backend + frontend 一起切）；對齊決策 **D32**（2026-09-17：Casbin 與 CASL 移除，授權改 OpenFGA + RLS）。
> 稽核基準：`docs/superpowers/specs/1.0-requirements/`（12 份）、決策 D1–D32、設計書 v1.0.34。

---

## 1. 目標與範圍（本波 Done 定義）

1. **授權引擎由 Casbin + CASL 遷移至內嵌 OpenFGA + RLS**（D32）：OpenFGA 負責服務/資源層授權決策（`Check` / `list-objects`）；RLS（`data_scope` all/company/department/self）保留為資料庫層兜底。
2. **Casbin 與 CASL 自後端移除**（engine、執行面、依賴）；前端由「CASL ability」遷移至「OpenFGA 驅動的後端 proxy」。
3. **RLS policy 只建核心表**（users / roles / companies / departments / role_permissions），使後端可端到端運作；未開工領域（03–09、fleet）的表與 policy 各自領域開工時再補，並沿用本波授權介面。
4. 一併補齊地基阻礙：核心表 migration、7 內建角色 seed、`bcrypt → Argon2id`。
5. 範圍**不含**未開工領域的業務 schema/RPC。

## 2. 架構與元件

### 2.1 OpenFGA 落地
- 新 `third_party/openfga/`：以 OpenFGA Go library **內嵌於後端程序**（in-process），datastore 共用現有 PostgreSQL（**單一 store**）；建立 store + authorization model。
- `Server` 持有 OpenFGA store/client；`Server.Init()` 加 fail-fast（store / model 就緒、DB/Valkey/OpenFGA 連線）。
- config：`config/openfga.go`（`OPENFGA_*` keys，內嵌模式以 DSN + store/model id 為主）；於 `config.go` 聚合；`.env.example` 同步。

### 2.2 授權介面
- 新 `internal/authz/openfga/`：
  - `Check(ctx, user, action, resource)` → boolean 決策；
  - `ListObjects(ctx, user, relation, type)` → 資源可見性；
  - 與 Connect middleware 接線。
- `internal/authz/access.go` 的 CASL 執行面（`AccessibleFilter`/`Can`）**移除**，由 OpenFGA Check/ListObjects 提供同等能力；`CASL_ENFORCEMENT_ENABLED` 開關語意退場，由 OpenFGA 啟用與否取代。
- `identityFor` 注入的身分轉為 OpenFGA subject（含 company/department userset 展開）。

### 2.3 role_permissions → OpenFGA
- `role_permissions` **保留為角色/功能權限定義來源**（D32）。
- 異動時同步 translate 成 OpenFGA tuples（`Write`/`Delete`）；`GetAbility`/前端權限由 OpenFGA 驅動。
- developer 逃生門：跳過 OpenFGA `Check`；RLS 注入 `data_scope=all`；`ENV=production` 誤開 → fail-fast（D8/D32 不變）。

## 3. 授權 model（取向 1：租戶 userset + 角色指派分離，資料驅動）

- **固定資源型別與 relations**（`role`、`company`、`department`、`ability`… 各帶 `can_read` / `can_write`…）；`company`/`department` 為租戶型別，`member`/`admin` 以 `[user#assigned]` userset 展開。
- **`role` 型別僅 `assigned: [user]`（角色指派）**；`role → 權限` 的對映**不進 model**，由 `role_permissions` 異動時 translate 成 OpenFGA tuples（資料驅動）→ **彈性自訂角色（D9）不需改 model／發版**。
- **兩層分工（2026-09-18 修訂 D32）**：OpenFGA 只做關係性 / 角色 / 租戶範圍授權；對**物件可變狀態**的條件（如「僅能取消 pending 訂單」）由 **domain 狀態機 / use-case 層**執行（D13），不推入 OpenFGA CEL（避免資料模型洩漏與狀態規則重複漂移）。
- model 以 `.fga` DSL 為唯一來源，版本化；未來領域僅需擴充 model 資源型別與 relations。

## 4. 前端遷移

- 移除 `@casl/ability`。
- 新增「我的權限」RPC（後端 proxy OpenFGA `Check`/`ListObjects`，回傳目前身分可用 (resource, action) 集合）供選單/守衛。
- `guards.ts` `requireAbility` → proxy `check`；`Can.tsx` → `Check` 元件；`context.tsx` / `service.ts` / router / RolesPage 同步改；`queryKey ["ability"]` 失效語意保留。

## 5. 測試

- OpenFGA 授權單元/整合測試：以 OpenFGA 記憶體/embedded 測試環境 + 真實 Postgres 整合測試（D21）。
- 隨引擎移除：刪除 `casbin` / `CASL golden` / `access` / `ability` 舊測試，改為 OpenFGA 對等測試。
- `task check`（fmt+vet+lint+test）通過。

## 6. 交付順序（本波內）

1. 核心表 migration + 7 內建角色 / 角色權限 seed + developer 帳號（dev only）
2. 內嵌 OpenFGA 起實例 + 授權 model DSL + `internal/authz/openfga` Check/ListObjects
3. middleware 接線：受保護 RPC → Check；身分 → OpenFGA + RLS
4. RLS policy 落地核心表（補 00002 為實際 policy）
5. role_permissions 異動 → OpenFGA tuples 同步
6. 移除 Casbin/CASL code；`GetAbility` → OpenFGA 驅動
7. `bcrypt → Argon2id`
8. 前端遷移（guards/Can/context/service/router）
9. `task check` + 整合測試過關

## 7. 已知風險 / 待解
- 屬性/狀態條件「條件欄位」在 OpenFGA condition 的可表達性需於實作時驗證（沿用既有 `casl.ParseConditions` 運算子集合）。
- 內嵌 OpenFGA 依賴量與生命週期綁定後端，需在 `Init()` 完成啟動防護。
- 前端遷移為較大改動，須以「我的權限」proxy 維持選單/守衛語意不變。

---

*最後更新：2026-09-18*

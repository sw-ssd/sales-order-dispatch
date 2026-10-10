# product 側遷移 runbook（platform v0.5.0）

> **本檔是 product repo 內的自帶副本**（`docs/platform/MIGRATION.md`）。維護本 repo 時照著它做；
> 平台的權威版本在 platform repo 的 `docs/superpowers/plans/2026-10-10-product-migration-runbook.md`。
> 兩者內容應一致，若不一致以**本副本的版本號**為準（平台升級時會一併更新副本）。

## 怎麼用（給 AI 的第一動作）

```sh
export GOPRIVATE=github.com/sw-ssd/platform
task platform:check          # 或 scripts/platform-check.sh：列出還缺哪幾步
```

腳本輸出的編號即本檔的章節編號。**逐項做完再跑一次**，全綠即整合完成；然後刪掉 `AGENTS.md` 裡
`PLATFORM-INTEGRATION-PENDING` 的段落。

2026-10-10。**狀態：未執行**（使用者指示：先完成平台側，product 等到維護它時再做）。
本檔把 product 要做的每一步寫成可勾選清單；每一步都對應平台側已實作、已測試的契約。

契約依據：`docs/superpowers/specs/2026-10-10-rpc-only-boundary.md`（設計）＋
`README.md`（平台對外介面與環境變數）。

---

## 0. 前提

| 項目 | 值 |
|---|---|
| 平台模組 | `github.com/sw-ssd/platform`，**最低 v0.5.0**（private GitHub repo）；實際用哪一版由 `go list -m -versions` 查得後**明確 pin** |
| 取用 | `export GOPRIVATE=github.com/sw-ssd/platform`（private repo，proxy 抓不到；**不設會靜默拿到 proxy 上舊架構的 v0.1.0**） |
| service token | `MIZU_SERVICEAUTH_TOKEN`（**兩向共用**：product → platform 的事件拉取面、platform → product 的橋接面） |
| 橋接面位址 | product 自己要有一個可被平台呼叫的位址 → 平台設 `MIZU_PRODUCT_BASEURL` |

## 1. go.mod

```sh
export GOPRIVATE=github.com/sw-ssd/platform
go mod edit -dropreplace=github.com/sw-ssd/platform   # 移除 local replace（若保留，抓到的不是發佈版）
go list -m -versions github.com/sw-ssd/platform       # 看目前有哪些已發佈版本（至少有 v0.5.0）
V=v0.5.0                                              # ← 明確選定，不要寫 latest（理由見下）
go mod edit -require=github.com/sw-ssd/platform@$V
go mod tidy
```

**為什麼不用 `@latest`**：① 私有模組一旦漏設 `GOPRIVATE`，public proxy 上只有舊架構的 `v0.1.0`，
`latest` 會靜默解析到那份舊 code（症狀：編譯出一個「能動但沒有新介面」的版本）；
② 未 pin＝未經 review 的契約變更（RPC 簽章可能已改）；③ 檢查與驗收都要可重現。
若查到的版本高於 v0.5.0，先比對本檔的契約差異（RPC／欄位有無增刪）再決定。

## 2. import 對映（機械替換；括號為 2026-10-10 的實查檔案數）

| 現在 | 改成 |
|---|---|
| `platform/entitlements` (41) | `platform/service/entitlements` |
| `platform/operatorauth` (11) | `platform/service/operatorauth` |
| `platform/store`＋`platform/store/postgres` (20) | `platform/package/store`＋`/postgres` |
| `platform/billing` (8) | `platform/service/billing` |
| `platform/cron` (4) | `platform/service/cronsvc` |
| `platform/money` (2) | `platform/package/money` |
| `platform/server` (2) | **不存在**：轉接層已由 platform binary 自己掛（步驟 4 一併刪） |
| `contracts/{errcode,requestid,cache}` (76) | `platform/package/{errcode,requestid,cache}` |
| `contracts/proto/platform/v1` (21) | `platform/protogen/platform/v1` |
| `contracts/proto/salesorder/v1`（僅 `common.proto` 那支,144） | `platform/protogen/salesorder/v1` |

驗收：`go list -deps ./... \| grep -E 'sw-ssd/platform/(service\|package/(store\|cache))'` 必須無輸出
（白名單：只准 `protogen/**`、`package/errcode`、`package/requestid`）。

## 3. 實作 PlatformBridgeService（product 提供給平台的面）

proto 在平台模組：`product/v1/bridge.proto`（三個 RPC）。

```go
// 掛載:與租戶 API 分開的 mux ＋ service token 認證(三套認證不得混用:operator cookie／
// 租戶 session／service token)。
cs := connect.NewServer(requestid.Interceptor(), serviceauth.Interceptor(token, prevToken))
productv1connect.RegisterPlatformBridgeServiceHandler(cs, impl)
connecthttp.Mount(mux, cs) // mux 掛在 /internal/platform/ 之下,網路層只允許平台來源
```

三個 RPC 的語意（每一條都對應平台側已測試的行為）：

| RPC | product 端要做什麼 | 注意 |
|---|---|---|
| `ApplyEvent` | 把平台推來的事件套用到產品域（凍結／恢復公司），**在系統範圍交易內**做（RLS：公司列需要 `scope=all`）；同時寫稽核 | **幂等**：以 `id` 為鍵。實作建議用**狀態比對**去重（已是目標狀態就不動、不寫稽核，回 `applied=false`）——比另建「已套用事件」表少一個 migration |
| `GetUsage` | 依 `company_internal_id`(+`features`) 回目前用量。feature code 與平台共用：`limit.seats`／`limit.customers`／`limit.products`／`limit.departments`／`limit.storage_gb` | 呼叫端是**平台**（非請求路徑），要用系統範圍交易計數；`map` 中**缺席**代表「沒有這個 feature 的來源」，不要填 0 |
| `ListCompanies` | 分頁回公司的投影欄位：`company_id`/`internal_id`/`name`/`identifier`/`status`/`deleted` | 用 **after_id 游標**（平台逐頁拉）；`deleted` 必須是業務端的軟刪除事實（平台據此把公司排除出租戶清單）；`identifier='platform'` 是平台自營公司 |

平台側的呼叫端：`productclient`（service token、5s 超時、僅唯讀重試）＋ `companyproj`（投影同步）
＋ `cronsvc.PushDispatcher`（事件推送，見步驟 5）。

## 4. 刪除清單（平台已接手）

| 刪除 | 理由 |
|---|---|
| `internal/server/domains.go` 的 `mountPlatformAuth`／`RegisterPlatformAdminService`／`RegisterTenantEntitlementService`、`cmd/platform-server` | 平台 binary 自己提供 `/platform/*`（v0.3.0 起有 `PlatformAdminService` 21 個 RPC） |
| `cmd/platform-cron` | 平台提供 `CronService.TriggerRun`（單飛鎖在平台端）；平台可自行推送事件（步驟 5） |
| `internal/platformhost/consumer`（含 `company_status_worker`） | 事件改由 API 交付（推或拉）；**直讀 `platform.events` 的路徑必須消失** |
| `internal/services/platform_admin_service.go`（平台 admin 的副本） | 平台已提供 |
| `contracts/` 內已收編的套件 | 見步驟 2 的對映 |
| `backend/database/platform_migrations/` | 平台鏈在 platform repo（版本表同名 `platform_goose_db_version` → **零操作**） |
| `entitlements.Invalidate` 的呼叫（consumer／company_status_worker） | 權益快取屬平台：平台在寫入路徑與收到事件後自行失效 |

## 5. 事件交付：推或拉（擇一，或並存）

兩者都用平台的 claim ＋ 租約，**同一筆只會被一邊取得**，所以可以無痛切換。

| 方式 | 誰驅動 | product 要做什麼 |
|---|---|---|
| **推**（建議；已在平台端實作） | 平台的排程（`TriggerRun`） | 步驟 3 的 `ApplyEvent` 即完成；平台失敗時自行重試（attempts 累加、租約釋放） |
| **拉** | product 自己 | 用 `platform/platformevents` 的 client：`ClaimEvents(limit, lease_seconds)` → 處理 → `AckEvent`／`FailEvent`。**不得**再直讀 `platform.events` |

切換順序（重要）：**先停舊的直讀 consumer**（它的認領只看 `dispatched_at`、不看 `claimed_at`，
與新機制並行會重複派送）→ 再啟用推送或拉取。新機制彼此並存是安全的。

## 6. 驗收（product 端）

- `ApplyEvent` 幂等：同一 `id` 送兩次，第二次 `applied=false` 且**沒有第二筆稽核**。
- `ListCompanies`：軟刪除的公司 `deleted=true`；平台自營公司 `identifier='platform'`。
- `GetUsage`：`limit.seats` = 未停用帳號數，與請求路徑 `CheckLimit.current_used` 的口徑一致
  （不一致會出現「守衛說超額、後台說還有位子」）。
- 平台側對應的端到端測試已存在（`service/companyproj` 的整合測試、`service/cronsvc` 的推送測試），
  可用它們的假替身當 product 的對照（`proto/product/v1` 的 handler 介面）。

## 7. 認證與網路

- service token 輪替：平台支援**雙值窗口** —— 平台設 `MIZU_SERVICEAUTH_PREVTOKEN=舊` →
  product 換新 → 平台清掉 prev。全程不停機。
- product 的橋接面**必須**與租戶 API 分開：不同 mux、不同認證、網路層只允許平台來源
  （否則「平台命令」與「租戶請求」可以互相冒充）。

## 8. 已知坑（平台側實測過的）

1. **payload 是 jsonb 正規化過的**：`{"plan":"std"}` 可能變成 `{"plan": "std"}`。不要用位元組相等比對，
   用 JSON 語意比對。
2. **RLS**：`ApplyEvent` 要動公司狀態 → 必須在系統範圍交易內（`scope=all`），否則會被政策濾成 0 列
   而「靜默成功」。
3. **`internal_id`**：平台以 uuid 為跨域鍵（三鍵策略）。product 的 `companies.internal_id` 必須存在
   （product 鏈 00053）；回報的公司若缺它，平台的同步會**跳過該列並記一行 log**。
4. **proxy 陷阱**：沒設 `GOPRIVATE` 的環境會拿到 proxy 上舊架構的 `v0.1.0`。

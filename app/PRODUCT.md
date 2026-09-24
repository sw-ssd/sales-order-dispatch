# Product

<!-- impeccable:product-schema 1 -->

## Platform

adaptive

> 已定案值為 `adaptive`（單一 Flutter 碼基同時出 iOS/Android）。未定案：稿面目前是單一 App 設計語彙（390×844 六個畫面），尚無 per-OS 差異承諾；new-work 決定是否真的分語言前，先當作「同一語彙、雙平台」處理。

## Users

- **業務**（主要）：在客戶店家現場用手機開單。品名可手打或從清單選（業務自行決定輸入方式）；管理該客戶的專屬商品清單；挑選偏好送貨日。
- **客戶店家（一主多子帳號）**：主帳號只管管理、不業務登入；子帳號以 QR Code 登入 App。子帳號由店家以主帳號自助管理（self 範圍），主帳號不可自停、停用會連鎖子帳號。
- **guest**：員工首次 Google 登入後待審核（super / company_admin 審），未完成註冊前不建帳號。

## Product Purpose

多公司訂出貨系統 1.0 的現場端。讓業務不必回辦公室就能在店家現場完成開單、改名、指定送貨日；讓店家自己管理子帳號、查看訂單歷史、發起退貨。

後台（派車、單據列印、主檔維護）見 `../frontend/PRODUCT.md`。產品整體的目的、上線方式（全新 monorepo、不遷移舊資料、Big Bang）與成功定義，兩端一致。

## Positioning

部門自治主檔 + 雙層租戶防護（OpenFGA 管功能、PostgreSQL RLS 管資料範圍）的產品，App 是它在店家現場的介面。相鄰產品做不到的：同一商品在不同部門、不同客戶各自命名，不需大總表；全系統不儲存金額。

## Operating Context

- 業務在店家現場作業，App 是唯一的現場工具；派車與四種紙本單據列印都在後台。
- 專屬商品清單的刪除語意是作業慣例：偶爾不用的品項把數量設 0 保留、不再顯示於單據；長期不用才刪除、不再帶出。
- 偏好送貨日為週一至週六核取；非勾選日下單自動順延。
- 通知（FCM + 站內，**無 Email**）：業務下單推客戶子帳號、後台新增專屬商品推主責業務、退貨審核結果推發起帳號。促銷推播依分類標籤選群，與公告分離。
- 建立客戶時自動附帶業務子帳號（專供所屬業務使用，店家端灰化不可管理）；App 手動表單新增客戶（主檔 + 登入帳號）屬 1.0 範圍（原為 1.1，已提前）。
- 介面語言繁體中文；文件、註解、commit message 一律繁體中文（多語 UI 不在 1.0 範圍）。
- 工程約束：一律以 `fvm` 執行；雙 flavor（dev / prod）與 `--dart-define` 注入組態；OAuth callback scheme `salesorder://` 異動需同步 AndroidManifest / Info.plist；不新增相依套件除非現有套件明確不足。
- 既有 UI 設計稿在 Pixso（`多公司訂出貨系統` UI 稿的 App 六畫面；來源存檔 `docs/design/2026-09-18-pixso-版面美化-進度存檔.md`）。

## Capabilities and Constraints

- 認證：App 用 access JWT 1 小時 + refresh 30 天旋轉；停用／強制登出／改密碼／角色變更以 `token_version` 全數失效。店家子帳號用 QR Code 登入。
- 退貨：客戶發起（歷史訂單勾選與專屬商品清單兩種來源並存）→ 業務審核 → 退貨證明；不建配送頁。
- 1.0 明確不含：App 強制更新、多語 UI、通知重試佇列、上傳病毒掃描。
- 1.1 為獨立迭代（拍照建客戶、語音下單，`docs/superpowers/specs/2026-07-18-app-ai-assist-1.1-design.md` v0.2.0），不得混入 1.0 範圍。
- **全系統不儲存金額**：App 的訂單、商品、專屬商品畫面都沒有價格欄位，也不得為了顯示而新增。

## Evidence on Hand

- 凍結規格書 `docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34，18 章）：欄位與流程的唯一權威。
- 決策記錄 `docs/superpowers/specs/2026-07-19-sales-order-1.0-decisions.md`（D1–D33）。
- 需求規格 12 份 `docs/superpowers/specs/1.0-requirements/`。
- 客戶原始需求 `docs/客戶需求.txt`、`docs/需求備忘_2026-08-03.txt`。
- Pixso App 稿（六畫面）與 `docs/design/2026-09-18-pixso-版面美化-進度存檔.md`。

不存在、未來工作不得捏造：真實客戶名單與推薦語、金額或營收資料、品牌 logo、字體授權憑證、App Store / Play 上架狀態。

## Product Principles

1. **現場優先**：業務在店家現場就能完成開單與改名，不要求回辦公室補資料。
2. **店家自助、後台有逃生門**：子帳號由店家自己管（self 範圍），管理端保留移交與救援路徑。
3. **通知是工作流的一部分**：推播路由（誰下單推給誰、審核結果推給誰）是產品行為，不是附帶功能。
4. **不持有金額**：系統刻意不儲存價格，App 也不得為了顯示而新增。

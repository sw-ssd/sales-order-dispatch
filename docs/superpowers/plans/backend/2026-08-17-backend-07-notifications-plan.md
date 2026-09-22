# Backend 07 — 通知系統 執行計畫（現況對齊版）

> **性質**：原為目標型執行計畫（含內嵌目標程式碼）。經 2026-09-18 盤點（codebase-memory 知識圖譜 + git）重建，為**反映現況的執行計畫**。
>
> **狀態基準**：2026-09-22 實作對齊。後端已落地：四表 schema（00041）＋渲染＋通知中心（List／MarkRead／UnreadCount）＋DeviceService＋Sender／Failmark（FakeSender；FCM 實裝另案）＋三路觸發（下單／專屬／退貨審核）＋00042 RLS ENABLE＋FORCE；派車 adapter 待 08（`DispatchNotifier` 介面未定）；Web/App 頁待。以下保留目標架構供追溯。
>
> **對應設計**：`docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34）、決策 `D16/D23/D24`
> **細部文件**：`docs/superpowers/plans/backend/detail/07-notifications.md`、共通規則 `detail/00-index.md` §3
> **前置依賴**：01-auth（RLS、audit.Recorder）、05（下單）、04（專屬商品）、08（派車通知 adapter）

---

## 執行狀態總覽

| Task | 內容 | 狀態 |
|---|---|---|
| 1 | notification_templates / notifications / user_devices schema + RLS（細部 4.3.1） | ✅ 完成（00041＋00042，白名單 32；另含 promo_tags 資料層） |
| 2 | 範本渲染 + 通知記錄/已讀 API（細部 4.3.2–4.3.3） | ✅ 完成（後端；Web/App 頁待） |
| 3 | DeviceService 註冊/註銷 + 失效 token 清理 + promo_tags（細部 4.3.4–4.3.5） | ✅ 完成（後端資料層＋裝置 API；CRUD/選群推播屬 Phase 7 Task 7.4） |
| 4 | FCM client + 站內發送 + 失敗標記不重試（細部 4.4.1、4.4.2、4.4.5） | 🟡 部分（站內＋Failmark 已落地；FCM 實裝另案，Sender 介面已預留） |
| 5 | 通知路由 — 下單與專屬商品觸發（細部 4.4.3–4.4.4） | ✅ 完成（另含退貨審核觸發） |
| 6 | 派車通知 adapter（對接 08 `DispatchNotifier`） | ⬜ 未開始（待 08 介面） |

**實作範圍**：後端約 85%（FCM 實裝＋派車 adapter＋Web/App 頁待）。

---

## Global Constraints（目標，精煉自原檔）

- 通道僅 `fcm` / `in_app`；通知建檔（pending）與觸發方業務**同一 DB 交易**（D18），FCM 外部呼叫於交易提交後執行，失敗僅標 `failed`、**不重試**（D16）。
- FCM 以 `notification.Sender` 介面抽象，測試注入 `FakeSender`。
- 範本渲染含語系退回；`user_devices` 裝置註冊/註銷 + 失效 token 清理；`promo_tags` 分類標籤（D24）與公告分離。
- 促銷推播依分類標籤選群（D24）；通知路由：業務下單推客戶子帳號、後台新增專屬商品推主責業務（D23）。

## 待辦 Task 詳情

### Task 1: notification_templates / notifications / user_devices schema + RLS（細部 4.3.1）
三表 schema（範本、通知記錄、裝置）＋ RLS。

### Task 2: 範本渲染 + 通知記錄/已讀 API（細部 4.3.2–4.3.3）
範本渲染（含語系退回）、通知中心查詢 + 已讀 API。

### Task 3: DeviceService 註冊/註銷 + 失效 token 清理 + promo_tags 資料層（細部 4.3.4–4.3.5）
裝置註冊/註銷、失效 FCM token 清理、`promo_tags` 資料層（D24）。

### Task 4: FCM client + 站內發送 + 失敗標記不重試（細部 4.4.1、4.4.2、4.4.5）
FCM Admin SDK client + 站內發送；提交後發送、失敗僅標 `failed` 不重試（D16）。

### Task 5: 通知路由 — 下單與專屬商品觸發（細部 4.4.3–4.4.4）
業務下單推客戶子帳號、後台新增專屬商品推主責業務（D23）。

### Task 6: 派車通知 adapter（對接 08 `DispatchNotifier`）
對接 08-dispatch 的 `DispatchNotifier` 介面。

---

*最後更新：2026-09-22（07-notifications 後端對齊；FCM 實裝＋派車 adapter 待）*

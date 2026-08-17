# announcements 需求規格

> 來源：原 OpenSpec delta spec（OpenSpec 工作流已停用，2026-08-03 遷移至 docs/）。


## Purpose

公告 CMS：管理並投放 Banner / 最新消息 / 圖文文章三類公告至 Web 中台與 App 首頁，支援全系統、公司、部門三層發佈範圍、上下架時間窗、平台篩選與排序。

## Requirements

### Requirement: 公告資料模型與類型

系統 SHALL 提供公告實體，區分三種類型 `banner`（首頁輪播）、`news`（最新消息）、`article`（圖文文章），且每筆公告 MUST 包含欄位：`type`、`title`、`content`、`image_url`、`link_url`、`publish_at`、`unpublish_at`、`sort_order`、`is_active`。公告 MUST 採軟刪除（`deleted_at`），列表查詢預設排除已刪除資料。

#### Scenario: 建立各類型公告

- **WHEN** 管理者建立公告並指定 `type` 為 `banner`、`news` 或 `article`，並填寫 `title`、`content` 等欄位
- **THEN** 系統建立公告並儲存全部欄位
- **AND** `type` 不屬於三種合法值時系統拒絕並回傳驗證錯誤

#### Scenario: 軟刪除公告

- **WHEN** 管理者刪除一筆公告
- **THEN** 該公告標記 `deleted_at` 而非實體刪除
- **AND** 前台展示與管理列表預設皆不再顯示該公告

### Requirement: 發佈範圍三層設定

公告 MUST 支援三層發佈範圍：`company_id` 與 `department_id` 皆為 NULL 時為全系統公告；僅指定 `company_id` 時為該公司公告；同時指定 `company_id` 與 `department_id` 時為該部門公告。使用者 SHALL 僅能看到全系統公告，以及其所屬公司、所屬部門的公告。

#### Scenario: 全系統公告

- **WHEN** 一筆公告的 `company_id` 與 `department_id` 皆為 NULL
- **THEN** 所有公司與部門的使用者皆可看到該公告

#### Scenario: 公司層級公告

- **WHEN** 一筆公告指定 `company_id` 且 `department_id` 為 NULL
- **THEN** 僅該公司（含其所有部門）的使用者可看到該公告
- **AND** 其他公司的使用者不可看到

#### Scenario: 部門層級公告

- **WHEN** 一筆公告同時指定 `company_id` 與 `department_id`
- **THEN** 僅該部門的使用者可看到該公告
- **AND** 同公司其他部門的使用者不可看到

### Requirement: 管理權限依範圍分層

公告的 CRUD 管理 SHALL 僅開放 `super`、`company_admin`、`dept_admin`，並依角色範圍限制：`super` 可管理所有範圍的公告；`company_admin` 僅可管理其所屬公司範圍（公司層級與該公司部門層級）的公告；`dept_admin` 僅可管理其所屬部門層級的公告。越權操作 MUST 被拒絕。

#### Scenario: super 管理所有公告

- **WHEN** `super` 角色建立或修改任一範圍（全系統、任一公司、任一部門）的公告
- **THEN** 系統允許操作

#### Scenario: company_admin 依公司範圍管理

- **WHEN** `company_admin` 建立或修改其所屬公司的公司層級或部門層級公告
- **THEN** 系統允許操作
- **AND** 當其操作其他公司或全系統範圍的公告時，系統拒絕並回傳權限錯誤

#### Scenario: dept_admin 依部門範圍管理

- **WHEN** `dept_admin` 建立或修改其所屬部門的公告
- **THEN** 系統允許操作
- **AND** 當其操作其他部門、公司層級或全系統範圍的公告時，系統拒絕並回傳權限錯誤

### Requirement: 平台篩選投放

每筆公告 SHALL 可標記投放平台（Web / App，可複選）；Web 中台 MUST 僅顯示標記投放 Web 的公告，App MUST 僅顯示標記投放 App 的公告。

#### Scenario: 僅投放 Web 的公告

- **WHEN** 一筆公告僅標記投放 Web
- **THEN** Web 中台顯示該公告
- **AND** App 首頁不顯示該公告

#### Scenario: 僅投放 App 的公告

- **WHEN** 一筆公告僅標記投放 App
- **THEN** App 首頁顯示該公告
- **AND** Web 中台不顯示該公告

### Requirement: 上下架時間窗與啟用狀態

前台展示 MUST 僅包含 `is_active` 為 true 且 `publish_at` 早於或等於現在時間、且 `unpublish_at` 晚於現在時間的公告；不滿足時間窗或未啟用的公告 MUST NOT 出現在任何前台列表或輪播中。

#### Scenario: 尚未到發佈時間

- **WHEN** 一筆公告 `is_active` 為 true 但 `publish_at` 晚於現在時間
- **THEN** 前台不顯示該公告

#### Scenario: 已過下架時間

- **WHEN** 一筆公告的 `unpublish_at` 早於或等於現在時間
- **THEN** 前台不顯示該公告

#### Scenario: 停用公告

- **WHEN** 管理者將一筆公告的 `is_active` 設為 false
- **THEN** 即使仍在時間窗內，前台亦不顯示該公告

### Requirement: 前台展示與排序

Web 中台與 App 首頁 SHALL 以輪播展示 `banner` 類型公告，並以列表展示 `news` 與 `article` 類型公告（最新消息）；`article` 類型公告 SHALL 提供圖文內容檢視。同類型公告 MUST 依 `sort_order` 排序；`banner` 設有 `link_url` 時，點擊 SHALL 導向該連結。

#### Scenario: Banner 輪播與最新消息列表

- **WHEN** 使用者開啟 Web 或 App 首頁
- **THEN** 系統以輪播展示符合條件的 `banner` 公告
- **AND** 以列表展示符合條件的 `news` / `article` 公告

#### Scenario: 依 sort_order 排序

- **WHEN** 同類型存在多筆符合條件的公告
- **THEN** 前台依 `sort_order` 排序顯示

#### Scenario: Banner 連結導向

- **WHEN** 使用者點擊設有 `link_url` 的 banner
- **THEN** 系統導向該 `link_url`

### Requirement: 公告圖片上傳與 WebP 處理

公告 SHALL 支援上傳圖片並寫入 `image_url`；上傳的圖片 MUST 經過 WebP 處理（轉換或產生 WebP 版本）後提供前台使用。檔案上傳的 MIME 白名單、大小限制與儲存機制由 `file-assets` capability 定義，本 capability 僅規範公告圖片的引用與呈現行為。

#### Scenario: 上傳公告圖片

- **WHEN** 管理者為公告上傳一張合法圖片
- **THEN** 系統處理為 WebP 版本並將其 URL 寫入該公告的 `image_url`

#### Scenario: 前台顯示公告圖片

- **WHEN** 前台展示設有 `image_url` 的公告
- **THEN** 顯示經 WebP 處理的圖片

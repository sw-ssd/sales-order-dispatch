# file-assets 需求規格

> 來源：原 OpenSpec delta spec（OpenSpec 工作流已停用，2026-08-03 遷移至 docs/）。


## Requirements

### Requirement: 上傳 MIME 與大小白名單

系統 SHALL 對所有檔案上傳套用白名單驗證：圖片僅接受 `image/jpeg`、`image/png`、`image/webp` 且大小 MUST 不超過 5 MB；文件僅接受 `application/pdf` 且大小 MUST 不超過 10 MB。白名單以外的 MIME 類型 MUST 拒絕上傳，超過對應大小上限的檔案 MUST 拒絕上傳。1.0 版本系統 SHALL NOT 實作上傳檔案病毒掃描。

#### Scenario: 接受白名單內的圖片

- **WHEN** 已登入使用者上傳 `image/jpeg`、`image/png` 或 `image/webp` 且大小不超過 5 MB 的圖片
- **THEN** 系統接受上傳並建立 `file_assets` 記錄

#### Scenario: 接受白名單內的 PDF

- **WHEN** 已登入使用者上傳 `application/pdf` 且大小不超過 10 MB 的文件
- **THEN** 系統接受上傳並建立 `file_assets` 記錄

#### Scenario: 拒絕白名單以外的 MIME 類型

- **WHEN** 使用者上傳 MIME 類型為 `image/gif`、`image/svg+xml`、`application/zip` 或其他白名單以外類型的檔案
- **THEN** 系統拒絕上傳，不寫入儲存，也不建立 `file_assets` 記錄，並回傳錯誤

#### Scenario: 拒絕超過大小上限的檔案

- **WHEN** 使用者上傳超過 5 MB 的圖片或超過 10 MB 的 PDF
- **THEN** 系統拒絕上傳，不寫入儲存，也不建立 `file_assets` 記錄，並回傳錯誤

### Requirement: 副檔名與 magic bytes 雙重檢查

系統 SHALL 同時檢查上傳檔案的副檔名與檔案內容的 magic bytes，兩者 MUST 與宣告的 MIME 類型一致；任一檢查不符即拒絕上傳，以防範偽造 MIME 宣告或改副檔名的偽裝檔案。

#### Scenario: 副檔名與 MIME 宣告不符

- **WHEN** 使用者上傳副檔名為 `.exe`（或與宣告 MIME 不符的副檔名）但宣告 MIME 為 `image/png` 的檔案
- **THEN** 系統拒絕上傳並回傳錯誤

#### Scenario: magic bytes 與宣告類型不符

- **WHEN** 使用者將一個實際內容非圖片的檔案改副檔名為 `.jpg` 並宣告為 `image/jpeg` 上傳
- **THEN** 系統以 magic bytes 判定內容不符，拒絕上傳並回傳錯誤

#### Scenario: 三重一致才接受

- **WHEN** 使用者上傳的檔案其宣告 MIME、副檔名、magic bytes 三者一致且均在白名單內
- **THEN** 系統接受上傳

### Requirement: 檔案元資料記錄

系統 SHALL 於每次成功上傳後建立一筆 `file_assets` 記錄，內容 MUST 包含 `company_id`、`department_id`、`owner_type`、`owner_id`、`filename`（儲存檔名）、`original_filename`（使用者原始檔名）、`mime_type`、`size_bytes`、`storage_path`、`url`、`created_by`，以支援多租戶歸屬、用途關聯與稽核追溯。

#### Scenario: 上傳成功寫入完整元資料

- **WHEN** 使用者上傳檔案成功
- **THEN** 系統建立一筆 `file_assets` 記錄，填入該使用者所屬的 `company_id` 與 `department_id`、關聯用途的 `owner_type` 與 `owner_id`、原始檔名與儲存檔名、實際 MIME 類型、檔案大小、`storage_path`、`url` 及上傳者 `created_by`

#### Scenario: 原始檔名與儲存檔名分別保存

- **WHEN** 使用者上傳名為 `產品型錄 2026.pdf` 的檔案
- **THEN** `file_assets.original_filename` 保留 `產品型錄 2026.pdf`，`filename` 為系統產生、不與其他檔案衝突的儲存檔名

### Requirement: 本地儲存與 storage_path

系統 1.0 SHALL 將上傳檔案儲存於本地掛載（NFS / 容器 volume），並於 `file_assets.storage_path` 記錄檔案的實際儲存路徑；檔案備份至 GCS 由 ops 備份排程負責，不屬於本 capability 的上傳/存取行為。

#### Scenario: 上傳後檔案落於本地掛載並記錄路徑

- **WHEN** 檔案上傳驗證通過
- **THEN** 系統將檔案寫入本地掛載的儲存位置，並將實際路徑寫入 `file_assets.storage_path`

#### Scenario: 儲存寫入失敗不留孤兒記錄

- **WHEN** 本地儲存寫入失敗（例如掛載不可用或磁碟錯誤）
- **THEN** 系統回傳錯誤，且不建立 `file_assets` 記錄

### Requirement: 下載 URL 與租戶隔離存取

系統 SHALL 透過後端端點 `GET /api/v1/files/:id/download` 提供檔案下載，不直接暴露本地儲存路徑；下載存取權限 MUST 依 `file_assets` 關聯的 `owner` 資料之租戶範圍判定，使用者 MUST 與檔案所屬 `company_id`（及其資料範圍）相符才可下載，跨公司存取 MUST 拒絕。

#### Scenario: 同公司使用者下載成功

- **WHEN** 已認證使用者對屬於其公司租戶範圍內的 `file_assets` 發出 `GET /api/v1/files/:id/download`
- **THEN** 系統回傳檔案內容與正確的 `mime_type`

#### Scenario: 跨公司存取被拒絕

- **WHEN** A 公司使用者以 B 公司檔案的 id 請求 `GET /api/v1/files/:id/download`
- **THEN** 系統拒絕存取，不回傳檔案內容

#### Scenario: 未認證存取被拒絕

- **WHEN** 未持有有效認證的請求存取 `GET /api/v1/files/:id/download`
- **THEN** 系統拒絕存取並回傳認證錯誤

#### Scenario: 下載不存在或已刪除的檔案

- **WHEN** 使用者請求下載不存在或已軟刪除的 `file_assets` id
- **THEN** 系統回傳找不到錯誤，不回傳任何檔案內容

### Requirement: 檔案軟刪除

`file_assets` SHALL 採軟刪除：刪除時 MUST 標記刪除而非移除記錄；已軟刪除的檔案 MUST 不再出現於查詢結果，且其下載請求 MUST 被拒絕。

#### Scenario: 軟刪除後不可下載

- **WHEN** 一筆 `file_assets` 已被軟刪除，使用者再對該 id 發出下載請求
- **THEN** 系統拒絕該請求並回傳找不到錯誤

#### Scenario: 軟刪除保留記錄供稽核

- **WHEN** 使用者刪除一筆 `file_assets`
- **THEN** 記錄仍保留於資料庫並標記刪除，`company_id`、`owner_type`、`owner_id`、`created_by` 等元資料不被移除

### Requirement: 用途關聯（owner_type / owner_id）

系統 SHALL 以 `owner_type` 與 `owner_id` 將 `file_assets` 關聯至其用途資料（例如公司 Logo、公告圖片、列印 PDF），同一 `file_assets` MUST 對應至多一個 owner；各用途的業務行為（上傳 Logo、建立公告、產生列印 PDF）屬各自 capability，本 capability 僅負責通用關聯的記錄與依 owner 租戶範圍的存取判定。

#### Scenario: 建立關聯後可依 owner 追溯

- **WHEN** 業務流程（如列印或公司識別設定）以上傳結果建立 `owner_type` / `owner_id` 關聯
- **THEN** 系統可經由該關聯查得檔案所屬 `company_id` 與 `department_id`，並據以判定存取權限

#### Scenario: 關聯至不存在的 owner 被拒絕

- **WHEN** 上傳時指定的 `owner_type` / `owner_id` 對應的資料不存在或不屬於上傳者的租戶範圍
- **THEN** 系統拒絕建立該 `file_assets` 記錄並回傳錯誤

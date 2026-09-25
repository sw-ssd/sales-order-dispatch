package config

// FCM 推播設定(07 計畫 Task 4.4.1/4.4.5)。
//
// `Provider` 決定「送出」這一步走哪個實作 —— 這是**環境控制**而非執行期切換:
// 開發/測試用 fake(僅落 DB 狀態,不觸外網)、正式用 fcm(真發 FCM HTTP v1)。
// 為何用顯式開關而非「有憑證就用真的」:後者會讓「憑證沒設好」靜默退化成假發送,
// 而通知永遠停在 pending/sent 的假象上線後才被發現 —— 那正是最難追的一類。
type FCM struct {
	// Provider 為發送實作:fake | fcm。預設 fake(開發友善且不觸外網)。
	// production 下設 fake 會被 Server.Init() 拒絕啟動(上線不該靜默不推播)。
	Provider string `envconfig:"FCM_PROVIDER" default:"fake"`
	// ProjectID 為 Firebase 專案 ID(HTTP v1 的 {parent=projects/*})。fcm 模式必填。
	ProjectID string `envconfig:"FCM_PROJECT_ID"`
	// CredentialsJSON 為 service account JSON 的**檔案路徑**(非內容:內容放環境變數
	// 會出現在 process 環境與 crash dump 裡)。fcm 模式必填。
	CredentialsJSON string `envconfig:"FCM_CREDENTIALS_JSON"`
	// Endpoint 覆寫 FCM 端點;空 = 官方。供測試指向 httptest server(不觸外網)。
	Endpoint string `envconfig:"FCM_ENDPOINT"`
	// Timeout 為單次發送的超時秒數。FCM 屬外部 I/O,不設上限會拖住提交後掛鉤。
	TimeoutSeconds int `envconfig:"FCM_TIMEOUT_SECONDS" default:"10"`
}

// UsesRealFCM 表示設定要求真實 FCM 發送。
func (f FCM) UsesRealFCM() bool { return f.Provider == "fcm" }

// Configured 表示真發送所需設定齊備(僅在 UsesRealFCM 時有意義)。
func (f FCM) Configured() bool { return f.ProjectID != "" && f.CredentialsJSON != "" }

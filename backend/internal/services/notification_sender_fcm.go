// FCM HTTP v1 發送實作(07 計畫 Task 4.4.5)。
//
// 為何自己打 HTTP 而不引入 firebase-admin-go:本功能只需要「帶 OAuth2 token 打一個
// POST」——admin SDK 會連帶拉進 gRPC、Firestore、Storage 等一整套依賴。用的憑證授權
// (`golang.org/x/oauth2/google`)與 FCM 端點都已在依賴樹中,自寫約 100 行。
//
// 時序邊界:本型別的 Send 由 notification_triggers 在**交易提交後**呼叫(AfterCommit),
// 屬外部 I/O,絕不進業務交易 —— 失敗只標 failed,不回滾業務(D16,不重試)。
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/notification"
	"github.com/salesorder/sales-order-1.0/backend/ent/userdevice"
)

// fcmScope 為 FCM HTTP v1 所需的 OAuth2 scope。
const fcmScope = "https://www.googleapis.com/auth/firebase.messaging"

// fcmDefaultEndpoint 為官方端點;測試以 config.FCM.Endpoint 覆寫。
const fcmDefaultEndpoint = "https://fcm.googleapis.com/v1"

// FCMSender 以 FCM HTTP v1 發送通知。
type FCMSender struct {
	projectID string
	endpoint  string
	client    *http.Client
	// tokens 取 OAuth2 access token(service account 換發;oauth2 套件自帶快取與續期)。
	tokens oauth2.TokenSource
}

// NewFCMSender 由 service account JSON 檔建立發送器。
//
// credentialsFile 是**路徑**不是內容(內容放環境變數會出現在 process 環境與 crash dump)。
// endpoint 空 = 官方端點。
func NewFCMSender(projectID, credentialsFile, endpoint string, timeout time.Duration) (*FCMSender, error) {
	if projectID == "" {
		return nil, fmt.Errorf("fcm: project id 不可為空")
	}
	// 讀檔 + 解析:權限不足、JSON 損壞、缺欄位都在這裡就爆(而非第一則通知發不出去才發現)。
	data, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("fcm: 讀取憑證 %q: %w", credentialsFile, err)
	}
	// 用 JWTConfigFromJSON 而非 google.CredentialsFromJSON:後者已被標為 deprecated
	// （不驗證憑證設定，外部來源的設定可能被替換成別的型別）。前者會驗 `type` 必須是
	// `service_account`，正是我們要的憑證種類 —— 設錯檔案的類型會在此直接失敗，
	// 而不是拿一份 authorized_user 設定去換 token 然後在發送時才被 FCM 拒絕。
	jwtCfg, err := google.JWTConfigFromJSON(data, fcmScope)
	if err != nil {
		return nil, fmt.Errorf("fcm: 解析 service account 憑證: %w", err)
	}
	if endpoint == "" {
		endpoint = fcmDefaultEndpoint
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &FCMSender{
		projectID: projectID,
		endpoint:  strings.TrimSuffix(endpoint, "/"),
		client:    &http.Client{Timeout: timeout},
		// jwt.Config 的 TokenSource 處理換發、快取與續期（自簽 JWT 換 access token）。
		tokens: jwtCfg.TokenSource(context.Background()),
	}, nil
}

// fcmMessage 為 FCM HTTP v1 的 message 物件(僅帶 notification,不帶 data:
// 1.0 的 App 以通知中心顯示內容,點擊不需要深層連結參數)。
type fcmMessage struct {
	Message struct {
		Token        string `json:"token"`
		Notification struct {
			Title string `json:"title"`
			Body  string `json:"body"`
		} `json:"notification"`
	} `json:"message"`
}

// Send 逐筆發送通知並落結果。
//
// 一批(同一次觸發的多位收件者)逐筆送出而非平行:一則通知對一顆裝置,數量是「該客戶的
// 子帳號數」量級;平行化換來的吞吐在這種規模看不出來,卻讓失敗歸因複雜化。
//
// 收件者沒有任何 FCM device 時標 failed(missing_device)而非 sent:裝置清單為空就代表
// **這則推播沒有送出去**,記成 sent 會讓「通知已發」變成不實的狀態(且掩蓋 App 未註冊)。
func (s *FCMSender) Send(ctx context.Context, db *ent.Client, notificationIDs []int) []SendResult {
	out := make([]SendResult, 0, len(notificationIDs))
	for _, nid := range notificationIDs {
		out = append(out, s.sendOne(ctx, db, nid))
	}
	return out
}

func (s *FCMSender) sendOne(ctx context.Context, db *ent.Client, nid int) SendResult {
	res := SendResult{NotificationID: nid}
	n, err := db.Notification.Query().Where(notification.IDEQ(nid)).Only(ctx)
	if err != nil {
		res.FailReason = "not_found"
		return res
	}
	// 只處理 fcm 通道:in_app 那筆本來就不該推播(同一次觸發會建兩筆,通道不同)。
	if n.Channel != "fcm" {
		res.Sent = true // 非推播通道無事可做,視為完成(不標 failed)
		applySendResult(ctx, db, n, res)
		return res
	}
	devices, err := db.UserDevice.Query().
		Where(userdevice.UserIDEQ(n.UserID), userdevice.DeletedAtIsNil()).
		All(ctx)
	if err != nil {
		res.FailReason = "device_query_failed"
		applySendResult(ctx, db, n, res)
		return res
	}
	if len(devices) == 0 {
		// 沒有任何裝置可送 → 未送出。不標 sent,讓「沒有 App 裝置」在資料上看得出來。
		res.FailReason = "missing_device"
		applySendResult(ctx, db, n, res)
		return res
	}

	var invalid []string
	sentAny := false
	lastReason := ""
	for _, d := range devices {
		reason, invalidToken := s.post(ctx, d.FcmToken, n.Title, n.Content)
		if reason == "" {
			sentAny = true
			continue
		}
		lastReason = reason
		if invalidToken {
			invalid = append(invalid, d.FcmToken)
		}
	}
	if sentAny {
		res.Sent = true
		// 部分裝置失效仍算送出成功,但失效 token 一併回報供清除
		// (同一個人有多顆裝置時,不該因為舊機失效讓新機的通知被標成失敗)。
		res.InvalidTokens = invalid
		applySendResult(ctx, db, n, res)
		return res
	}
	res.FailReason = lastReason
	if res.FailReason == "" {
		res.FailReason = "send_failed"
	}
	res.InvalidTokens = invalid
	applySendResult(ctx, db, n, res)
	return res
}

// post 對單一 token 發送;回 (失敗原因, 是否為失效 token)。原因為空字串 = 成功。
//
// 失效 token 的判定以 FCM 官方的 ErrorCode 為準(UNREGISTERED/404、INVALID_ARGUMENT/400、
// SENDER_ID_MISMATCH/403):這幾種代表**這顆 token 不該再留**,回報後由呼叫端清除;
// 其餘(503/500/429 等)是暫時性問題,token 本身有效,不得刪。
func (s *FCMSender) post(ctx context.Context, token, title, body string) (string, bool) {
	var msg fcmMessage
	msg.Message.Token = token
	msg.Message.Notification.Title = title
	msg.Message.Notification.Body = body
	payload, err := json.Marshal(msg)
	if err != nil {
		return "marshal_failed", false
	}

	tok, err := s.tokens.Token()
	if err != nil {
		// 憑證換 token 失敗:整批都送不出,但 token 本身沒問題(不刪裝置)。
		return "oauth_token_failed", false
	}

	url := fmt.Sprintf("%s/projects/%s/messages:send", s.endpoint, s.projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "request_build_failed", false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return "network_error", false
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		return "", false
	}
	reason, invalid := fcmFailure(resp.StatusCode)
	return reason, invalid
}

// fcmFailure 依 HTTP 狀態給失敗原因與「是否失效 token」。
//
// 判定集中在一個函式:散在呼叫端會讓「哪個狀態該刪 token」出現第二種說法,
// 而誤刪有效 token 的代價是使用者從此收不到任何推播(直到重新登入)。
func fcmFailure(status int) (string, bool) {
	switch status {
	case http.StatusBadRequest: // 400 INVALID_ARGUMENT:token 格式錯/已失效
		return "invalid_argument", true
	case http.StatusNotFound: // 404 UNREGISTERED:App 已解除註冊或已卸載
		return "unregistered", true
	case http.StatusForbidden: // 403 SENDER_ID_MISMATCH:token 屬於別的 sender
		return "sender_id_mismatch", true
	case http.StatusUnauthorized: // 401 THIRD_PARTY_AUTH_ERROR:APNs 憑證問題,非 token 失效
		return "third_party_auth", false
	case http.StatusTooManyRequests:
		return "quota_exceeded", false
	case http.StatusServiceUnavailable:
		return "unavailable", false
	default:
		return fmt.Sprintf("http_%d", status), false
	}
}

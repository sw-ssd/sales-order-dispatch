package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"

	"golang.org/x/oauth2"
)

// fcmTestSender 以 httptest server 取代 FCM 端點（**不觸外網**），並注入靜態 token
// 來源 —— 憑證換發那條路徑另有 NewFCMSender 的錯誤處理測試覆蓋。
func fcmTestSender(t *testing.T, handler http.HandlerFunc) (*FCMSender, *httptest.Server) {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return &FCMSender{
		projectID: "proj-test",
		endpoint:  ts.URL,
		client:    ts.Client(),
		tokens:    oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "tok"}),
	}, ts
}

func openSenderDB(t *testing.T) *ent.Client {
	t.Helper()
	db := enttest.Open(t, "sqlite3", "file:sender?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func seedSenderNotification(t *testing.T, db *ent.Client, channel string, withDevice bool, fcmToken string) int {
	t.Helper()
	ctx := context.Background()
	co := db.Company.Create().SetName("推播公司").SetIdentifier("FCM-C").SetStatus("active").SaveX(ctx)
	dept := db.Department.Create().SetName("推播部門").SetCompanyID(co.ID).SaveX(ctx)
	u := db.User.Create().SetEmail("fcm-" + t.Name() + "@t.com").SetName("收件者").
		SetRole("customer").SetPasswordHash("x").SetCompanyID(co.ID).
		SetIsCustomer(true).SetStatus("active").SaveX(ctx)
	if withDevice {
		db.UserDevice.Create().SetUserID(u.ID).SetCompanyID(co.ID).
			SetPlatform("ios").SetFcmToken(fcmToken).SetDeviceName("iPhone").
			SaveX(ctx)
	}
	n := db.Notification.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).SetUserID(u.ID).
		SetChannel(channel).SetTitle("標題").SetContent("內容").SetStatus("pending").
		SaveX(ctx)
	return n.ID
}

// TestFCMSenderSuccessAndRequestBody 驗真發送路徑的**請求內容**與狀態落點:
// 端點路徑含 project、Authorization Bearer、body 帶 token/notification。
// 少了任何一項 FCM 會回 4xx,而錯誤會被歸類成 token 失效（誤刪裝置）。
func TestFCMSenderSuccessAndRequestBody(t *testing.T) {
	var gotPath, gotAuth string
	var gotBody map[string]any
	sender, _ := fcmTestSender(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"projects/proj-test/messages/1"}`))
	})
	db := openSenderDB(t)
	nid := seedSenderNotification(t, db, "fcm", true, "device-token-1")

	results := sender.Send(context.Background(), db, []int{nid})

	if len(results) != 1 || !results[0].Sent {
		t.Fatalf("應發送成功,got %+v", results)
	}
	if gotPath != "/projects/proj-test/messages:send" {
		t.Fatalf("請求路徑錯誤: %q", gotPath)
	}
	if gotAuth != "Bearer tok" {
		t.Fatalf("Authorization 應帶 Bearer token,got %q", gotAuth)
	}
	msg, _ := gotBody["message"].(map[string]any)
	if msg["token"] != "device-token-1" {
		t.Fatalf("body 應帶裝置 token,got %v", msg["token"])
	}
	// 狀態落點:通知為 sent 且記 sent_at。
	n := db.Notification.GetX(context.Background(), nid)
	if n.Status != "sent" || n.SentAt == nil {
		t.Fatalf("通知應為 sent 且有 sent_at,got status=%q sent_at=%v", n.Status, n.SentAt)
	}
}

// TestFCMSenderInvalidTokenIsReportedAndDeviceNotDeleted 失效 token 必須被回報
// （供清除）但**不得在此處刪除**（清除需稽核,見 PurgeInvalidTokens）。
func TestFCMSenderInvalidTokenIsReportedAndDeviceNotDeleted(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   int
		invalid  bool
		wantFail string
	}{
		{"404 UNREGISTERED", http.StatusNotFound, true, "unregistered"},
		{"400 INVALID_ARGUMENT", http.StatusBadRequest, true, "invalid_argument"},
		{"403 SENDER_ID_MISMATCH", http.StatusForbidden, true, "sender_id_mismatch"},
		// 暫時性失敗**不得**被當成 token 失效:誤刪會讓使用者從此收不到推播。
		{"503 UNAVAILABLE", http.StatusServiceUnavailable, false, "unavailable"},
		{"429 QUOTA", http.StatusTooManyRequests, false, "quota_exceeded"},
		{"401 THIRD_PARTY_AUTH", http.StatusUnauthorized, false, "third_party_auth"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sender, _ := fcmTestSender(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
			})
			db := openSenderDB(t)
			nid := seedSenderNotification(t, db, "fcm", true, "tok-"+tc.name)

			res := sender.Send(context.Background(), db, []int{nid})[0]

			if res.Sent {
				t.Fatalf("HTTP %d 不該算發送成功", tc.status)
			}
			if res.FailReason != tc.wantFail {
				t.Fatalf("失敗原因應為 %q,got %q", tc.wantFail, res.FailReason)
			}
			got := res.InvalidTokens
			if tc.invalid && len(got) != 1 {
				t.Fatalf("失效 token 應回報 1 個,got %v", got)
			}
			if !tc.invalid && len(got) != 0 {
				t.Fatalf("暫時性失敗不該回報失效 token,got %v", got)
			}
			// 裝置列仍在(清除是另一條路徑)。
			if n := db.UserDevice.Query().CountX(context.Background()); n != 1 {
				t.Fatalf("裝置列不該被發送路徑刪除,got %d", n)
			}
			// 通知標 failed 並記原因。
			n := db.Notification.GetX(context.Background(), nid)
			if n.Status != "failed" || n.FailureReason != tc.wantFail {
				t.Fatalf("通知應為 failed(%s),got status=%q reason=%q", tc.wantFail, n.Status, n.FailureReason)
			}
		})
	}
}

// TestFCMSenderMissingDeviceFailsNotSent 沒有註冊裝置時必須標 failed 而非 sent:
// 記成 sent 會讓「通知已發」成為不實狀態,且掩蓋「App 未註冊推播」這個真問題。
func TestFCMSenderMissingDeviceFailsNotSent(t *testing.T) {
	sender, _ := fcmTestSender(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("沒有裝置時不該呼叫 FCM")
	})
	db := openSenderDB(t)
	nid := seedSenderNotification(t, db, "fcm", false, "")

	res := sender.Send(context.Background(), db, []int{nid})[0]

	if res.Sent {
		t.Fatal("沒有裝置不得算發送成功")
	}
	if res.FailReason != "missing_device" {
		t.Fatalf("失敗原因應為 missing_device,got %q", res.FailReason)
	}
	n := db.Notification.GetX(context.Background(), nid)
	if n.Status != "failed" {
		t.Fatalf("通知應為 failed,got %q", n.Status)
	}
}

// TestFCMSenderSkipsInAppChannel 同一次觸發會建 in_app 與 fcm 兩筆;
// in_app 那筆不該被送去 FCM(會多一則重複推播)。
func TestFCMSenderSkipsInAppChannel(t *testing.T) {
	called := false
	sender, _ := fcmTestSender(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	db := openSenderDB(t)
	nid := seedSenderNotification(t, db, "in_app", true, "tok")

	res := sender.Send(context.Background(), db, []int{nid})[0]

	if called {
		t.Fatal("in_app 通知不該呼叫 FCM")
	}
	if !res.Sent {
		t.Fatalf("in_app 無事可做應視為完成,got %+v", res)
	}
	if n := db.Notification.GetX(context.Background(), nid); n.Status != "sent" {
		t.Fatalf("in_app 應標 sent(否則永遠停在 pending),got %q", n.Status)
	}
}

// TestFCMSenderPartialDeviceFailure 同一人多顆裝置時,舊機失效不該讓新機的通知變失敗。
func TestFCMSenderPartialDeviceFailure(t *testing.T) {
	var calls []string
	sender, _ := fcmTestSender(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		msg, _ := body["message"].(map[string]any)
		tok, _ := msg["token"].(string)
		calls = append(calls, tok)
		if tok == "stale" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	db := openSenderDB(t)
	ctx := context.Background()
	co := db.Company.Create().SetName("多機").SetIdentifier("FCM-M").SetStatus("active").SaveX(ctx)
	dept := db.Department.Create().SetName("多機部門").SetCompanyID(co.ID).SaveX(ctx)
	u := db.User.Create().SetEmail("multi@t.com").SetName("多機").
		SetRole("customer").SetPasswordHash("x").SetCompanyID(co.ID).SaveX(ctx)
	for _, tok := range []string{"stale", "fresh"} {
		db.UserDevice.Create().SetUserID(u.ID).SetCompanyID(co.ID).
			SetPlatform("ios").SetFcmToken(tok).SetDeviceName(tok).SaveX(ctx)
	}
	nid := db.Notification.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).SetUserID(u.ID).
		SetChannel("fcm").SetTitle("t").SetContent("c").SetStatus("pending").SaveX(ctx).ID

	res := sender.Send(ctx, db, []int{nid})[0]

	if len(calls) != 2 {
		t.Fatalf("兩顆裝置都應嘗試,got %v", calls)
	}
	if !res.Sent {
		t.Fatal("至少一顆成功即算送出（否則新機的通知會被舊機拖成失敗）")
	}
	if len(res.InvalidTokens) != 1 || res.InvalidTokens[0] != "stale" {
		t.Fatalf("失效 token 應回報 stale,got %v", res.InvalidTokens)
	}
}

// TestNewFCMSenderRejectsBadCredentials 憑證問題必須在**建立時**爆,
// 而不是第一則通知發不出去才發現（那時使用者已經收不到通知了）。
func TestNewFCMSenderRejectsBadCredentials(t *testing.T) {
	if _, err := NewFCMSender("", "/nonexistent.json", "", 0); err == nil {
		t.Fatal("缺 project id 應回錯")
	}
	if _, err := NewFCMSender("p", "/definitely/not/here.json", "", 0); err == nil {
		t.Fatal("憑證檔不存在應回錯")
	}
	// 存在但不是合法 JSON。
	f := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(f, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFCMSender("p", f, "", 0); err == nil {
		t.Fatal("損壞的憑證應回錯")
	}

	// 合法 JSON 但**型別不對**：FCM 需要 service_account，authorized_user 換不到
	// 正確的 token。這條正是選 JWTConfigFromJSON（而非已 deprecated 的
	// CredentialsFromJSON）的理由 —— 後者不驗型別，會讓設錯的憑證活到發送階段才失敗。
	g := filepath.Join(t.TempDir(), "user.json")
	if err := os.WriteFile(g, []byte(
		`{"type":"authorized_user","client_id":"a","client_secret":"b","refresh_token":"c"}`,
	), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFCMSender("p", g, "", 0); err == nil {
		t.Fatal("authorized_user 憑證應被拒絕（FCM 需要 service_account）")
	}
}

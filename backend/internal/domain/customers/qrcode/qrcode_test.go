package qrcode_test

import (
	"strings"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/domain/customers/qrcode"
)

// TestGenerateVerifyRoundTrip 產生→驗證取回 company/customer/jti。
func TestGenerateVerifyRoundTrip(t *testing.T) {
	token, jti, err := qrcode.Generate("test-secret", 7, "TY000001", 0)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if jti == "" || !strings.Contains(token, ".") {
		t.Fatalf("token 形狀錯誤: %q jti=%q", token, jti)
	}
	c, err := qrcode.Verify("test-secret", token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if c.CompanyID != 7 || c.CustomerCode != "TY000001" || c.JTI != jti || c.Purpose != qrcode.Purpose {
		t.Fatalf("claims 不符: %+v", c)
	}
}

// TestVerifyRejectsTamper 竄改任一欄位即簽章失敗(unauthenticated 語意:有錯即拒,不區分)。
func TestVerifyRejectsTamper(t *testing.T) {
	token, _, err := qrcode.Generate("test-secret", 7, "TY000001", 0)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	payload, _, _ := strings.Cut(token, ".")
	bad := payload[:len(payload)-2] + "xx" + "." + "00"
	if _, err := qrcode.Verify("test-secret", bad); err == nil {
		t.Fatal("竄改應拒絕")
	}
	if _, err := qrcode.Verify("other-secret", token); err == nil {
		t.Fatal("錯密鑰應拒絕")
	}
	if _, err := qrcode.Verify("test-secret", "not-a-token"); err == nil {
		t.Fatal("格式非法應拒絕")
	}
}

// TestVerifyRejectsExpired 過期 token 拒絕(exp 秒級:2 秒效期等 2.5 秒必過期)。
func TestVerifyRejectsExpired(t *testing.T) {
	token, _, err := qrcode.Generate("test-secret", 7, "TY000001", 2*time.Second)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// exp 秒級取整:2 秒效期在毫秒邊界可能差 1 秒才真正過期,等 3.5 秒必過。
	time.Sleep(3500 * time.Millisecond)
	if _, err := qrcode.Verify("test-secret", token); err == nil {
		t.Fatal("過期應拒絕")
	}
}

package server

import (
	"strings"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/config"
)

const (
	tenantSecret   = "test-tenant-jwt-secret"
	operatorSecret = "test-platform-operator-secret"
	consoleURL     = "https://console.example.com"
)

// newServerFromEnv 以 config.New()（envconfig）讀環境後組 Server。
// **必須**走 envconfig：平台守護的條件之所以是「顯式 opt-in」，根因正是 envconfig 會填預設值
// （AllowedEmailDomain 預設 sowinsoft.com、DEVELOPER_ACCOUNT_ENABLED 預設 true）——
// 手寫 config.Config 字面值會把這個現實藏起來，正是 T1 審查 I-1 指出的錯誤。
// infra 指向必然拒絕的位址（port 1），使「未觸發平台守護」的案例能確定走到 DB probe。
func newServerFromEnv(t *testing.T, env, platformSecret, platformConsole string) *Server {
	t.Helper()
	t.Setenv("ENV", env)
	t.Setenv("DEVELOPER_ACCOUNT_ENABLED", "false")
	t.Setenv("JWT_SECRET", tenantSecret)
	t.Setenv("DATABASE_URL", "postgres://probe:probe@127.0.0.1:1/probe")
	t.Setenv("VALKEY_ADDR", "127.0.0.1:1")
	// 平台 env 必須在 config.New() **之前**設定：先組 Server 再 setenv 會讓 envconfig 讀到空值,
	// 測試就測不到真正要驗的條件（空字串視同不設）。
	if platformSecret != "" {
		t.Setenv("PLATFORM_JWT_SECRET", platformSecret)
	}
	if platformConsole != "" {
		t.Setenv("PLATFORM_CONSOLE_URL", platformConsole)
	}
	return New(config.New())
}

// platformGuardErr 表示拒絕來源是平台守護（供「不得由平台守護拒絕」的斷言）。
func platformGuardErr(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "平台工具") || strings.Contains(err.Error(), "PLATFORM_JWT_SECRET"))
}

// TestInitPlatformGuards production 的平台設定守護（D38／任務 1）：
//  1. 平台工具採**顯式 opt-in**：只設了一半（有 secret 無 console，或反之）→ 拒絕啟動，
//     訊息必須給出兩條都可行的出路（補齊，或整組不設以停用）。
//  2. 整組不設 → **不觸發**平台守護（＝不啟用平台工具，與組裝處 `Configured()` 判斷一致）。
//  3. PLATFORM_JWT_SECRET 等於租戶 JWT_SECRET，或等於隨 repo 公開的 dev 預設值 → 拒絕啟動，
//     訊息不得洩漏密鑰值。
//  4. development 不受這些守護影響（本機開發可不設平台工具）。
func TestInitPlatformGuards(t *testing.T) {
	t.Run("全部 PLATFORM_* 不設 → 不觸發平台守護", func(t *testing.T) {
		s := newServerFromEnv(t, "production", "", "")
		err := s.Init()
		if err == nil {
			t.Fatal("production 於 infra 不可達時應被 DB fail-fast 拒絕")
		}
		// 注意：這條不變式與守護在 Init() 中的**位置無關**（條件本身不成立即不觸發），
		// 別把它當成「偵測插隊」的鎖。它保的是 business_role_integration_test.go 的斷言：
		// 該測試不留平台設定，故 production 的拒絕**必須**來自業務連線角色檢查。
		if platformGuardErr(err) {
			t.Fatalf("整組不設不得由平台守護拒絕（否則該守護在實務上無法跳出）,got %q", err.Error())
		}
		if !strings.Contains(err.Error(), "無法連線資料庫") {
			t.Fatalf("整組不設應落到 infra 檢查,got %q", err.Error())
		}
	})

	t.Run("只設 PLATFORM_JWT_SECRET → 拒絕並要求補齊", func(t *testing.T) {
		s := newServerFromEnv(t, "production", operatorSecret, "")
		err := s.Init()
		if err == nil || !strings.Contains(err.Error(), "只設了一部分") {
			t.Fatalf("部分設定必須被平台守護拒絕,got %v", err)
		}
		if !strings.Contains(err.Error(), "PLATFORM_CONSOLE_URL") || !strings.Contains(err.Error(), "PLATFORM_JWT_SECRET") {
			t.Fatalf("拒絕訊息必須給出兩條出路(補齊哪兩個 env／整組不設),got %q", err.Error())
		}
	})

	t.Run("只設 PLATFORM_CONSOLE_URL → 拒絕(反向亦成立)", func(t *testing.T) {
		s := newServerFromEnv(t, "production", "", consoleURL)
		err := s.Init()
		if err == nil || !strings.Contains(err.Error(), "只設了一部分") {
			t.Fatalf("部分設定必須被平台守護拒絕,got %v", err)
		}
	})

	t.Run("設定齊備 → 不觸發平台守護", func(t *testing.T) {
		s := newServerFromEnv(t, "production", operatorSecret, consoleURL)
		err := s.Init()
		if err == nil {
			t.Fatal("production 於 infra 不可達時應被 DB fail-fast 拒絕")
		}
		if platformGuardErr(err) {
			t.Fatalf("設定齊備不得由平台守護拒絕,got %q", err.Error())
		}
	})

	t.Run("PLATFORM_JWT_SECRET 為 dev 預設值 → 拒絕且不洩漏密鑰", func(t *testing.T) {
		s := newServerFromEnv(t, "production", config.DefaultJWTSecret, consoleURL)
		err := s.Init()
		if err == nil || !strings.Contains(err.Error(), "dev-only") {
			t.Fatalf("沿用公開的 dev 預設平台密鑰必須拒絕啟動,got %v", err)
		}
		if strings.Contains(err.Error(), config.DefaultJWTSecret) {
			t.Fatalf("拒絕訊息不得洩漏密鑰值,got %q", err.Error())
		}
	})

	t.Run("PLATFORM_JWT_SECRET 與租戶 JWT_SECRET 共用 → 拒絕且不洩漏密鑰", func(t *testing.T) {
		s := newServerFromEnv(t, "production", tenantSecret, consoleURL)
		err := s.Init()
		if err == nil || !strings.Contains(err.Error(), "不得與 JWT_SECRET 相同") {
			t.Fatalf("共用租戶密鑰必須拒絕啟動,got %v", err)
		}
		if strings.Contains(err.Error(), tenantSecret) {
			t.Fatalf("拒絕訊息不得洩漏密鑰值,got %q", err.Error())
		}
	})

	t.Run("development 不受平台守護影響", func(t *testing.T) {
		// 部分設定:dev 亦不得被擋。
		s := newServerFromEnv(t, "development", operatorSecret, "")
		if err := s.Init(); err != nil {
			t.Fatalf("development 不得被平台守護擋下,got %v", err)
		}
	})
}

package server

import (
	"strings"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/config"
)

// TestInitPlatformGuards production 的平台設定守護（D38／任務 1）：
//  1. 平台工具設定缺件（只設了一半）→ 拒絕啟動，訊息須指出缺哪個 env。
//     否則 operator 登入看似啟用卻走不通，而錯誤只在登入時才爆。
//  2. PLATFORM_JWT_SECRET 與租戶 JWT_SECRET 共用 → 拒絕啟動（共用等於租戶 token 可冒充平台操作者），
//     且訊息**不得洩漏密鑰值**。
//  3. development 不受影響（本機開發可不設平台工具）；整組不設（零值）亦不在此守護的範圍。
func TestInitPlatformGuards(t *testing.T) {
	const shared = "platform-operator-secret-xyz"

	t.Run("production 平台設定缺件拒絕", func(t *testing.T) {
		// AllowedEmailDomain 由 envconfig 填預設值 → 正常啟動的 production 恆為非零值。
		s := New(&config.Config{
			API:      config.API{Env: "production"},
			Auth:     config.Auth{JWTSecret: "tenant-secret"},
			Platform: config.Platform{AllowedEmailDomain: "sowinsoft.com"},
		})
		err := s.Init()
		if err == nil || !strings.Contains(err.Error(), "平台工具設定不完整") {
			t.Fatalf("production 平台設定缺件必須被平台守護拒絕,got %v", err)
		}
		if !strings.Contains(err.Error(), "PLATFORM_JWT_SECRET") || !strings.Contains(err.Error(), "PLATFORM_CONSOLE_URL") {
			t.Fatalf("拒絕訊息必須給修復指引(指出缺哪個 env),got %q", err.Error())
		}
	})

	t.Run("production operator 密鑰與租戶共用拒絕且不洩漏密鑰", func(t *testing.T) {
		s := New(&config.Config{
			API:  config.API{Env: "production"},
			Auth: config.Auth{JWTSecret: shared},
			Platform: config.Platform{
				OperatorJWTSecret: shared, AllowedEmailDomain: "sowinsoft.com",
				ConsoleURL: "https://console.example.com",
			},
		})
		err := s.Init()
		if err == nil || !strings.Contains(err.Error(), "不得與 JWT_SECRET 相同") {
			t.Fatalf("production 共用密鑰必須拒絕啟動,got %v", err)
		}
		if strings.Contains(err.Error(), shared) {
			t.Fatalf("拒絕訊息不得洩漏密鑰值,got %q", err.Error())
		}
	})

	t.Run("development 不受平台守護影響", func(t *testing.T) {
		s := New(&config.Config{
			API:      config.API{Env: "development"},
			Auth:     config.Auth{JWTSecret: config.DefaultJWTSecret},
			Platform: config.Platform{OperatorJWTSecret: config.DefaultJWTSecret},
		})
		if err := s.Init(); err != nil {
			t.Fatalf("development 不得被平台守護擋下,got %v", err)
		}
	})

	// 順序鎖:整組不設(零值)必須落到既有的 infra 檢查,不得由平台守護先攔。
	// business_role_integration_test.go 以「業務連線角色」訊息斷言 production 的拒絕來源 ——
	// 平台守護若提前插隊,該測試會以錯誤的理由變綠(訊息被蓋掉)。
	t.Run("整組不設不觸發平台守護", func(t *testing.T) {
		s := New(&config.Config{
			API:  config.API{Env: "production"},
			Auth: config.Auth{JWTSecret: "tenant-secret"},
		})
		err := s.Init()
		if err == nil {
			t.Fatal("production 於無 infra 環境應被 DB/Valkey fail-fast 拒絕")
		}
		if strings.Contains(err.Error(), "平台工具") || strings.Contains(err.Error(), "PLATFORM_JWT_SECRET") {
			t.Fatalf("零值不得由平台守護拒絕,應落到 infra 檢查,got %q", err.Error())
		}
	})
}

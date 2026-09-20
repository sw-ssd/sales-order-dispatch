//go:build integration

package testsupport

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/salesorder/sales-order-1.0/backend/third_party/cache"
)

const (
	valkeyImage = "docker.io/valkey/valkey:8"
	valkeyPort  = "6379/tcp"
)

// Valkey 回傳一台**全新空** Valkey(redis 協定)的 host:port:起 valkey:8 容器,測試結束終止。
//
// 為什麼需要它:租戶 apiMux 的掛載點在 mountAuth 內,而 mountAuth 對 Valkey 不可用是
// 「log ＋ 略過掛載」(domains.go)。要在**production 的掛載點**上驗租戶 RPC,就必須讓
// 掛載真的發生 —— 否則只能另寫一份組裝,而另寫一份正好驗不到 domains.go 的掛載行
// (掛載行刪掉測試仍綠,那正是這類測試最常見的假通過)。容器執行環境不可用 → skip
// (與 Postgres 同一套判準:連不上才 skip,設定壞掉一律 fail)。
func Valkey(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        valkeyImage,
			ExposedPorts: []string{valkeyPort},
			WaitingFor:   wait.ForLog("Ready to accept connections").WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	if err != nil {
		if isRuntimeUnavailable(err) {
			t.Skipf("無法啟動 %s 容器,跳過整合測試(容器執行環境不可用: %v)\n"+
				"可行作法:以 `task test:integration` 執行(會設定 DOCKER_HOST 指向本機 podman machine socket)。",
				valkeyImage, err)
		}
		t.Fatalf("啟動 %s 容器失敗(容器執行環境可用,但設定有問題): %v\n"+
			"podman 下 testcontainers 的 ryuk 無法運作,需設 TESTCONTAINERS_RYUK_DISABLED=true;"+
			"以 `task test:integration` 執行即可(已固化此 env)。", valkeyImage, err)
	}
	t.Cleanup(func() {
		if err := ctr.Terminate(context.Background()); err != nil {
			t.Logf("終止 %s 容器失敗: %v", valkeyImage, err)
		}
	})

	host, err := ctr.Host(ctx)
	if err != nil {
		t.Fatalf("取得容器 host: %v", err)
	}
	port, err := ctr.MappedPort(ctx, valkeyPort)
	if err != nil {
		t.Fatalf("取得容器對外埠: %v", err)
	}
	addr := fmt.Sprintf("%s:%s", host, port.Port())
	if err := pingValkey(addr); err != nil {
		t.Fatalf("容器已啟動但無法連線(%v)", err)
	}
	return addr
}

// pingValkey 於時限內重試 PING,吸收容器剛啟動的暖機時間(與 postgres 的 ping 同思路)。
func pingValkey(addr string) error {
	client := cache.NewClient(addr)
	defer func() { _ = client.Close() }()
	var last error
	for deadline := time.Now().Add(30 * time.Second); time.Now().Before(deadline); {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		last = cache.Ping(ctx, client)
		cancel()
		if last == nil {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return last
}

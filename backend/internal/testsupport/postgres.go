//go:build integration

// Package testsupport 提供整合測試(//go:build integration)共用的外部依賴樣板。
// 需求容器一律集中在此,個別測試不得自寫 podman/docker 指令或自行拼 DSN。
//
// 容器執行環境的 env(DOCKER_HOST 指向本機 podman machine socket、podman 下停用 ryuk)
// 由 `task test:integration` 固化,測試碼不寫死;若環境已提供現成 PostgreSQL,設
// INTEGRATION_TEST_DSN 即可跳過容器(沿用既有 internal/authz/openfga 整合測試的 gating 名)。
// 兩者皆不可用時 skip(附可行動訊息),不得讓測試因環境而紅。
package testsupport

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	postgresImage = "postgres:16"
	postgresUser  = "postgres"
	postgresPass  = "postgres"
	postgresPort  = "5432/tcp"
	// postgresDB 為容器內建資料庫名。業務遷移 00002 以名定址(`ALTER DATABASE salesorder ...`),
	// 故測試資料庫必須叫 salesorder —— 這也是每次呼叫各自起一顆容器(而非同顆開多庫)的原因。
	postgresDB = "salesorder"
	// INTEGRATION_TEST_DSN 覆寫:指向現成 PostgreSQL(該庫本身即測試用拋棄式庫),跳過容器。
	dsnEnv = "INTEGRATION_TEST_DSN"
)

// Postgres 回傳一台**全新空** PostgreSQL(名為 salesorder)的連線字串:優先沿用
// INTEGRATION_TEST_DSN,未設定時起 postgres:16 容器並於測試結束終止。
// 每次呼叫都是獨立 server/資料庫;容器執行環境不可用 → skip。
//
// 注意:覆寫模式下測試會直接對該 DSN 跑遷移,請指向專用的拋棄式資料庫,
// 且庫名須為 salesorder(00002 以名定址)。
func Postgres(t *testing.T) string {
	t.Helper()
	if dsn := os.Getenv(dsnEnv); dsn != "" {
		if err := ping(dsn); err != nil {
			t.Fatalf("%s 指向的 PostgreSQL 無法連線(%v);請修正 DSN 或移除該變數以改用容器", dsnEnv, err)
		}
		return dsn
	}
	return startContainer(t)
}

// startContainer 起拋棄式 postgres 容器,回傳其 DSN;容器執行環境不可用即 skip。
func startContainer(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        postgresImage,
			ExposedPorts: []string{postgresPort},
			Env: map[string]string{
				"POSTGRES_USER":     postgresUser,
				"POSTGRES_PASSWORD": postgresPass,
				"POSTGRES_DB":       postgresDB,
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	if err != nil {
		// 容器執行環境「連不上」→ skip(本機無 docker/podman 屬正常情境)。
		// 其他失敗(API 可用但設定壞掉,例如 podman 下 ryuk 無法掛載 host socket)一律 fail:
		// 這種情況若也 skip,整個整合測試會靜默全綠 —— 正是本批測試要防的假通過。
		if isRuntimeUnavailable(err) {
			t.Skipf("無法啟動 %s 容器,跳過整合測試(容器執行環境不可用: %v)\n"+
				"可行作法:以 `task test:integration` 執行(會設定 DOCKER_HOST 指向本機 podman machine socket),\n"+
				"或改設 %s 指向現成 PostgreSQL。", postgresImage, err, dsnEnv)
		}
		t.Fatalf("啟動 %s 容器失敗(容器執行環境可用,但設定有問題): %v\n"+
			"podman 下 testcontainers 的 ryuk 無法運作(掛載 host socket、bridge 網路皆不可用),需設 TESTCONTAINERS_RYUK_DISABLED=true;\n"+
			"以 `task test:integration` 執行即可(已固化此 env),或改設 %s 指向現成 PostgreSQL。", postgresImage, err, dsnEnv)
	}
	t.Cleanup(func() {
		if err := ctr.Terminate(context.Background()); err != nil {
			t.Logf("終止 %s 容器失敗: %v", postgresImage, err)
		}
	})

	host, err := ctr.Host(ctx)
	if err != nil {
		t.Fatalf("取得容器 host: %v", err)
	}
	port, err := ctr.MappedPort(ctx, postgresPort)
	if err != nil {
		t.Fatalf("取得容器對外埠: %v", err)
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", postgresUser, postgresPass, host, port.Port(), postgresDB)
	if err := ping(dsn); err != nil {
		t.Fatalf("容器已啟動但無法連線(%v)", err)
	}
	return dsn
}

// isRuntimeUnavailable 判斷錯誤是否屬「容器執行環境連不上」(而非環境可用但設定壞掉)。
// 只認連線類錯誤訊息,不猜其他失敗原因;新增情境時直接把字串加入清單。
func isRuntimeUnavailable(err error) bool {
	msg := err.Error()
	for _, s := range []string{
		"Cannot connect to the Docker daemon", // 無 DOCKER_HOST 且無預設 socket
		"connect: no such file or directory",  // podman machine 未啟動(socket 不存在)
		"connection refused",
		"dial unix",
	} {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

// ping 於時限內重試連線,吸收容器剛啟動的暖機時間。
func ping(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	var last error
	for deadline := time.Now().Add(30 * time.Second); time.Now().Before(deadline); {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		last = db.PingContext(ctx)
		cancel()
		if last == nil {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return last
}

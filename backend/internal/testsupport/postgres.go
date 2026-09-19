//go:build integration

// Package testsupport 提供整合測試(//go:build integration)共用的外部依賴樣板。
// 需求容器一律集中在此,個別測試不得自寫 podman/docker 指令或自行拼 DSN。
//
// 容器執行環境的 env(DOCKER_HOST 指向本機 podman machine socket、podman 下停用 ryuk)
// 由 `task test:integration` 固化,測試碼不寫死;若環境已提供現成 PostgreSQL,設
// INTEGRATION_TEST_DSN 即可跳過容器(沿用既有 internal/authz/openfga 整合測試的 gating 名),
// 但該模式共用同一個既有的庫、**不保證** per-test 隔離:需要全新空庫的測試須先呼叫
// RequiresContainer(覆寫模式下 skip)。
// 兩者皆不可用時 skip(附可行動訊息),不得讓測試因環境而紅。
package testsupport

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	postgresImage = "postgres:16"
	postgresUser  = "postgres"
	postgresPass  = "postgres"
	postgresPort  = "5432/tcp"
	// postgresDB 為容器內建資料庫名(業務 schema 的慣用名)。庫名可由 PostgresNamed 指定,
	// 故 00002 起不再有「測試庫必須叫 salesorder」的限制(見 00002 以 current_database() 定址)。
	postgresDB = "salesorder"
	// INTEGRATION_TEST_DSN 覆寫:指向現成 PostgreSQL(該庫本身即測試用拋棄式庫),跳過容器。
	dsnEnv = "INTEGRATION_TEST_DSN"
)

// Postgres 回傳一台**全新空** PostgreSQL(名為 salesorder)的連線字串:優先沿用
// INTEGRATION_TEST_DSN,未設定時起 postgres:16 容器並於測試結束終止。
// 容器模式下每次呼叫都是獨立 server/資料庫;容器執行環境不可用 → skip。
//
// 注意:覆寫模式回傳的是**同一個** INTEGRATION_TEST_DSN,不保證每次呼叫都是全新庫;
// 需要 per-test 隔離的測試(假設缺表/缺版號/索引不存在…)必須自行先呼叫 RequiresContainer。
// 覆寫模式下測試會直接對該 DSN 跑遷移,請指向專用的拋棄式資料庫。
func Postgres(t *testing.T) string {
	t.Helper()
	if dsn := os.Getenv(dsnEnv); dsn != "" {
		if err := ping(dsn); err != nil {
			t.Fatalf("%s 指向的 PostgreSQL 無法連線(%v);請修正 DSN 或移除該變數以改用容器", dsnEnv, err)
		}
		return dsn
	}
	return startContainer(t, postgresDB)
}

// RequiresContainer 宣告本測試需要 per-test 隔離(自己的容器、自己的全新空庫):覆寫模式
// (INTEGRATION_TEST_DSN)回傳的是同一個既有的庫,無法保證 → skip(附可行動訊息)。
// 容器模式(未設該變數)下不做任何事 —— 隔離來自 Postgres/PostgresNamed 起的容器。
func RequiresContainer(t *testing.T) {
	t.Helper()
	if os.Getenv(dsnEnv) != "" {
		t.Skipf("%s 已設定:此模式共用同一個既有資料庫,無法保證本測試拿到全新空庫;本測試需要 per-test 隔離,請移除該變數並以 `task test:integration` 執行", dsnEnv)
	}
}

// PostgresNamed 回傳一台**全新空** PostgreSQL(庫名為 name)的連線字串,與 Postgres 共用
// 同一容器樣板(唯一差異是 POSTGRES_DB)。供「庫名不叫 salesorder」的情境使用 —— 例如驗證
// 00002 不再以硬編名定址資料庫(該實例內不得存在名為 salesorder 的庫,否則測不出硬編名)。
//
// 覆寫模式(INTEGRATION_TEST_DSN)無法保證庫名 → skip:寧可少跑一條,也不要對名為 salesorder
// 的庫斷言「非預設庫名可遷移」而假通過。
func PostgresNamed(t *testing.T, name string) string {
	t.Helper()
	if dsn := os.Getenv(dsnEnv); dsn != "" {
		t.Skipf("%s 已設定,無法保證資料庫名為 %s;本測試需自建容器(移除該變數,或以 `task test:integration` 執行)", dsnEnv, name)
	}
	return startContainer(t, name)
}

// CreateDatabase 於同一台 PostgreSQL(adminDSN 所指的 server)建立額外空資料庫,回傳其 DSN。
// 供「同一容器內第二個資料庫」情境使用(例如 OpenFGA 專用庫的優先序驗證),
// 不另起容器:單一容器樣板仍只有一處。
func CreateDatabase(t *testing.T, adminDSN, name string) string {
	t.Helper()
	db, err := sql.Open("pgx", adminDSN)
	if err != nil {
		t.Fatalf("連線以建立資料庫 %s: %v", name, err)
	}
	defer func() { _ = db.Close() }()
	// CREATE DATABASE 不接受參數佔位,故以 pgx 的識別字引號處理(不自行拼字串)。
	// 先 DROP ... WITH (FORCE)(PG13+)清掉同名舊庫,使同一顆 server 上重複以同一個庫名呼叫
	// 仍是冪等:容器模式下每測試各起一顆容器不會踩到,覆寫模式/共用 server 會(否則 42P04)。
	nameSQL := pgx.Identifier{name}.Sanitize()
	if _, err := db.Exec("DROP DATABASE IF EXISTS " + nameSQL + " WITH (FORCE)"); err != nil {
		t.Fatalf("清理既有資料庫 %s: %v", name, err)
	}
	if _, err := db.Exec("CREATE DATABASE " + nameSQL); err != nil {
		t.Fatalf("建立資料庫 %s: %v", name, err)
	}
	u, err := url.Parse(adminDSN)
	if err != nil {
		t.Fatalf("解析 DSN(%s): %v", adminDSN, err)
	}
	u.Path = "/" + name
	return u.String()
}

// startContainer 起拋棄式 postgres 容器(庫名 dbName),回傳其 DSN;容器執行環境不可用即 skip。
func startContainer(t *testing.T, dbName string) string {
	t.Helper()
	ctx := context.Background()
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        postgresImage,
			ExposedPorts: []string{postgresPort},
			Env: map[string]string{
				"POSTGRES_USER":     postgresUser,
				"POSTGRES_PASSWORD": postgresPass,
				"POSTGRES_DB":       dbName,
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
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", postgresUser, postgresPass, host, port.Port(), dbName)
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

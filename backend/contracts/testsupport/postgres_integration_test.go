//go:build integration

package testsupport

import (
	"database/sql"
	"errors"
	"testing"
)

// TestIsRuntimeUnavailable 固定「環境不可用 → skip / 環境可用但設定壞 → fail」的分界:
// ryuk 這類設定錯誤若被吞成 skip,整個整合測試會靜默全綠(本機 podman 實例即如此,
// 且會在背景留下 Created 狀態的 postgres 容器)。
func TestIsRuntimeUnavailable(t *testing.T) {
	cases := []struct {
		name string
		msg  string
		want bool
	}{
		{"無預設 socket", "Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?", true},
		{"podman machine 未啟動", "dial unix /tmp/podman/podman-machine-default-api.sock: connect: no such file or directory", true},
		{"API 可用但 ryuk 掛載 socket 失敗", "reaper: new reaper: making volume mountpoint for volume /var/folders/x/podman.sock: operation not supported", false},
		{"API 可用但 ryuk 找不到 bridge 網路", "reaper: new reaper: Error response from daemon: unable to find network with name or ID bridge: network not found", false},
	}
	for _, c := range cases {
		if got := isRuntimeUnavailable(errors.New(c.msg)); got != c.want {
			t.Errorf("%s: isRuntimeUnavailable = %t, want %t", c.name, got, c.want)
		}
	}
}

// TestCreateDatabaseIdempotent 固定 CreateDatabase 的冪等契約:同一顆 server 上以同一個庫名
// 連續呼叫兩次都必須成功,且第二次拿到的是**全新空庫**。
// 呼叫端(如 OpenFGA DSN 優先序測試)用固定庫名,容器模式下每測試各起一顆容器故不變,
// 但覆寫模式或共用 server 時若直接 CREATE DATABASE,第二輪會以 42P04 中斷。
func TestCreateDatabaseIdempotent(t *testing.T) {
	const name = "createdb_idempotent"
	adminDSN := Postgres(t)

	for i := 1; i <= 2; i++ {
		dsn := CreateDatabase(t, adminDSN, name)
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			t.Fatalf("第 %d 次開啟 %s: %v", i, name, err)
		}
		if err := db.PingContext(t.Context()); err != nil {
			_ = db.Close()
			t.Fatalf("第 %d 次連線 %s: %v", i, name, err)
		}
		// 殘留標記表若還在,表示拿到的是舊庫(冪等但不能沿用內容)。
		var leftover bool
		if err := db.QueryRow(
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'createdb_marker')`,
		).Scan(&leftover); err != nil {
			_ = db.Close()
			t.Fatalf("第 %d 次查 createdb_marker: %v", i, err)
		}
		if leftover {
			_ = db.Close()
			t.Fatalf("第 %d 次拿到的 %s 不是全新空庫(殘留 createdb_marker)", i, name)
		}
		if _, err := db.Exec(`CREATE TABLE createdb_marker (id int)`); err != nil {
			_ = db.Close()
			t.Fatalf("第 %d 次建立 createdb_marker: %v", i, err)
		}
		if err := db.Close(); err != nil {
			t.Fatalf("第 %d 次關閉連線: %v", i, err)
		}
		t.Logf("第 %d 次 CreateDatabase(%s) 成功: %s", i, name, dsn)
	}
}

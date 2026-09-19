//go:build integration

package testsupport

import (
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

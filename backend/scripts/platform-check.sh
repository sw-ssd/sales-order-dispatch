#!/usr/bin/env bash
# 平台整合檢查（PLATFORM-INTEGRATION-PENDING 的守門員）。
#
# 用途:在 product repo 維護時，一眼看出「platform（最低 v0.5.0）的整合還缺哪幾步」。
# 它只做**靜態檢查**（不連網、不改檔），每一步對應 docs/platform/MIGRATION.md 的編號。
#
# 用法:
#   scripts/platform-check.sh          # 未整合 → exit 1 並列出缺的步驟
#   PLATFORM_CHECK_STRICT=0 …          # 有缺也不回非零（CI 的 warning 模式）
#
# 整合完成後:本腳本保留為「不得再出現舊路徑」的回歸檢查（屆時全部應該 PASS）。
set -uo pipefail

cd "$(dirname "$0")/.." || exit 2

status=0
note() { printf '  %-4s %s\n' "$1" "$2"; }
fail() { printf '  %-4s %s\n' "✗" "$2"; status=1; }
pass() { printf '  %-4s %s\n' "✓" "$2"; }

echo "平台整合檢查（platform >= v0.5.0）—— 步驟編號對應 docs/platform/MIGRATION.md"

# 1) go.mod 是否仍指向本地 replace（整合後應 require v0.5.0 且無 replace）
if grep -q '^replace github.com/sw-ssd/platform ' go.mod 2>/dev/null; then
  fail "1" "go.mod 仍有 replace github.com/sw-ssd/platform => …"
else
  pass "1" "go.mod 無 platform 的本地 replace"
fi
if grep -q 'github.com/sw-ssd/platform v0\.\(5\|[6-9]\|[1-9][0-9]\)\.' go.mod 2>/dev/null; then
  pass "1" "go.mod require 已是 v0.5.0 以上（已 pin 具體版本）"
else
  fail "1" "go.mod 尚未 require v0.5.0 以上（現在：$(grep -m1 'github.com/sw-ssd/platform ' go.mod | tr -s ' ' | sed 's/^ //')）；用 go list -m -versions 查最新後明確 pin"
fi

# 2) import 白名單：只准 protogen 與 package/{errcode,requestid}
offenders=$(grep -rho '"github.com/sw-ssd/platform/[^"]*"' --include='*.go' . 2>/dev/null |
  sort -u |
  grep -vE '^"(github.com/sw-ssd/platform/(protogen|package/errcode|package/requestid))' || true)
if [ -n "$offenders" ]; then
  fail "2" "仍有非白名單 import（見下）"
  printf '%s\n' "$offenders" | sed 's/^/       /'
else
  pass "2" "platform import 全在白名單內"
fi

# 3) 舊結構殘留
for path in contracts cmd/platform-server cmd/platform-cron internal/platformhost; do
  if [ -e "$path" ]; then
    fail "3" "仍有舊結構：$path/"
  fi
done
[ "$status" -eq 0 ] && pass "3" "無舊結構殘留（contracts／cmd/platform-*／platformhost）"

# 4) 是否仍自己掛平台面（/platform/* 的 handler 註冊）
if grep -rqE 'RegisterPlatformAdminService|RegisterTenantEntitlementService|mountPlatformAuth' --include='*.go' . 2>/dev/null; then
  fail "4" "仍自己掛平台 RPC（平台 binary 已提供，見 MIGRATION §4）"
else
  pass "4" "未自行掛載平台 RPC"
fi

# 5) 橋接面是否已提供（product → platform 的反向面；platform 要呼叫它）
if grep -rq 'PlatformBridgeServiceHandler' --include='*.go' . 2>/dev/null; then
  pass "5" "已提供 PlatformBridgeService"
else
  fail "5" "尚未提供 product/v1.PlatformBridgeService（ApplyEvent／GetUsage／ListCompanies，見 MIGRATION §3）"
fi

# 6) 事件是否仍直讀平台表（跨域 SQL）
if grep -rqE 'FROM[[:space:]]+platform\.events|INTO[[:space:]]+platform\.events' --include='*.go' . 2>/dev/null; then
  fail "6" "仍直讀／直寫 platform.events（改走平台 API，見 MIGRATION §5）"
else
  pass "6" "未直讀平台 outbox 表"
fi

# 7) private module 的可取用性（只提示，不在這裡驗網路）
if [ "${GOPRIVATE:-}" = "github.com/sw-ssd/platform" ] || [ "${GOPRIVATE:-}" = "github.com/sw-ssd/*" ]; then
  pass "7" "GOPRIVATE 已設定"
else
  note "!" "GOPRIVATE 未設（或未含 github.com/sw-ssd/platform）→ 會拿到 proxy 上舊架構的 v0.1.0"
fi

echo
if [ "$status" -ne 0 ] && [ "${PLATFORM_CHECK_STRICT:-1}" = "0" ]; then
  echo "（warning 模式：僅提示，不回非零）"
  exit 0
fi
[ "$status" -eq 0 ] && echo "平台整合檢查：全部通過" || echo "平台整合檢查：尚未整合 → 照 docs/platform/MIGRATION.md 逐項做完"
exit "$status"

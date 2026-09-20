# 企業級錯誤碼系統 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立跨三端穩定的錯誤碼契約：單一 registry（常數即註冊、啟動即驗證）、結構化回應（`code`／`details`／`trace_id`）、集中映射（ent → 碼）、守門測試（既有 **233 處**「只減不增」的基線）、自動產生三端常數與客服可查的碼表。

**Architecture:** `internal/errcode` 為唯一真相來源（Go 常數 + `register()` 於 init 驗證格式／唯一性／區段與 connect code 的對應規則，違反即 panic）。回應以 connect error detail 帶 `common.v1.ErrorInfo`（不引入 googleapis 依賴）。`trace_id` 由 `internal/obs/requestid` 的 interceptor 在回應邊界補進 `ErrorInfo`（`Error`／`Wrap` 不收 ctx）。既有 **233 處** `connect.NewError(`（**98 行基線**，鍵為 `path:歸屬名:筆數`）以**基線檔 + 掃描測試**限制「只減不增」，不要求一次改完。

**Tech Stack:** Go 1.25、connect-go、google/uuid、protobuf（自有 `common.v1.ErrorInfo`）、Vitest/Dart 常數產生

**Spec:** `docs/superpowers/specs/2026-09-20-saas-billing-entitlements-design.md`（§2.2 錯誤語意、§4.3 失敗碼分工）

**排序**：`docs/superpowers/plans/2026-09-20-implementation-order.md` 的 **P0-5**。

**與 Plan A（RLS）的並行範圍（重要）**：**只有 T1–T3 可與 Plan A 並行**（proto／registry／trace_id，完全不重疊）。**T4–T5 必須排在 Plan A 之後**：兩者改到同一批檔案——

| 檔案 | Plan A | Plan D |
|---|---|---|
| `internal/services/company_service.go` | T9：19 處 `s.db.` → `dbtenant.Client` ＋交易三段 | T4：`toConnectError`；T5：`requireScope` |
| `internal/handlers/auth_handler.go` | T9：加 `SystemScopeTx` 包裝 | T5：登入／鎖定／停用的錯誤碼 |

同時改一個檔案會造成無謂的 rebase 衝突；序列化的成本遠低於衝突的處理成本。T6（基線與產生器）與 T7（文件對齊）不受此限。

**依賴**：與 **Plan B／C 有交集**（配額與收款錯誤改碼，見 Task 5、Task 7）。建議順序：**T1–T3 與 Plan A 並行先做完**（骨架），**T4–T5 等 Plan A 完成**（同檔重疊，見上表），Plan B 開始時直接用碼，就不必回頭改。

## Global Constraints

- **碼的形態固定**：`^[A-Z]{2,6}-\d{4}$`；**發佈後不得重用、不得改義**；廢止只加 `Deprecated: true`（測試斷言已 deprecated 的碼不得在新呼叫點出現）
- **區段與 connect 碼的對應是硬規則**（`register()` 驗證）：`1xxx→InvalidArgument`、`2xxx→AlreadyExists`、`3xxx→FailedPrecondition`、`4xxx→{PermissionDenied, Unauthenticated, NotFound}`、`5xxx→FailedPrecondition`、`9xxx→Internal`
- **未註冊的碼無法建構**：`errcode.New(code, …)` 只接受 registry 內的 `Code`（型別即約束）；`register()` 重複 ID 或格式錯誤 → **panic**（啟動就失敗）
- **訊息與參數分離**：後端回 `code` + `message`（繁中樣板渲染後，給 log／fallback）+ `details`（結構化參數，供前端與客服）；**缺參數不得讓錯誤處理爆掉**（回樣板原文 + log warn）
- **5xx 一律 `SYS-9000`**：內部細節（SQLSTATE、constraint 名、stack）只進 log，永不進 message；`trace_id` 同時回給客戶端
- **跨租戶與不存在一律 `SYS-4002`（NotFound）**：不洩漏資源是否存在（防 oracle 探測）。授權**檢查**失敗（角色/範圍不足）才是 `SYS-4001`（PermissionDenied）
- **守門是基線制**：`connect.NewError(` 的既有 **233 處**（98 行基線）列入 `errcode_baseline.txt`；**新增未帶註冊碼者即測試紅**，且基線只能縮小（測試同時斷言「基線不得殘留已消失的呼叫點」）
- 產生檔（碼表與三端常數）**必須與 registry 同步**：CI 跑 generate 後 `git diff --exit-code`
- 註解與 commit message 一律繁體中文

---

## Progress

> 執行記錄與逐任務細節（含每次審查的判定、裁決與更正）見 `.superpowers/sdd/2026-09-20-error-codes-plan/progress.md`。狀態：**7/7 主任務完成**（＋ T3b；T5b 待 Plan B 落地），2026-09-20。下表依**實際執行順序**排列（T1 → T3 → T3b → T2 → T4 → T5 → T6 → T7，理由見「執行期間的計畫更正」第 1 項）。

| # | 任務 | 狀態 | 產出（commit） | 驗證 |
|---|---|---|---|---|
| 1 | `ErrorInfo` proto 與三端生成 | ✅ | `526142b` | 三語言執行期往返皆 62 bytes、重跑生成零 diff |
| 3 | `trace_id` interceptor（15 個生產掛載點） | ✅ | `0e3816b`＋`40a41c7` | `go test ./internal/obs/...`、掛載回歸 25 PASS（log 出現 `rpc: … trace_id=…`） |
| 3b | 回應邊界為帶 `ErrorInfo` 的錯誤補 `trace_id` | ✅ | `5905f40` | 跨網路 `TestErrorInfoTraceIdReachesClient` 7 子測；**實測：就地修改不可行，必須重建** |
| 2 | `internal/errcode` registry 與首批碼 | ✅ | `3e095fe`＋`28a6d4e`＋`d2a254f` | 21 碼逐條落在區段允許集合、五項啟動驗證（含零值與前綴 panic）、`go list -deps` 證明葉節點 |
| 4 | `toConnectError` 改走 registry | ✅ | `943118d`＋`7c1ee8d` | 6 分支表驅動＋訊息外洩／根因落 log 斷言；Plan A 四支整合探針全 PASS（含 42501 遮蔽） |
| 5 | 首批碼落地（auth／權限／樣板域）＋閘門 | ✅ | `374f7ff`＋`c78ff64`／`a52339a`／`2a42db0`／`9003c5e` | P0-5 跨網路驗收 `TestIntegrationErrorInfoReachesClient` PASS；`writeConnectError` 改帶 `ErrorInfo`＋trace_id |
| 6 | 基線守門、碼表與三端常數產生 | ✅ | `b52e3d6`／`4665c36`／`0a0f72b`／`c8af1c7`＋`9a53850`／`17bfb8c` | 三種 RED 齊備；`go generate` 連跑零 diff；CI 漂移與未 commit 產物皆擋下（exit=1） |
| 7 | 慣例與跨文件對齊 | ✅ | `faf0b4c`＋`c0af568`＋本表所在 commit | `git diff --stat` 僅文件；數字全數以指令重驗（見下） |
| 5b | `PLAT-*` 四碼落地（entitlements／billing） | ⬜ 待 Plan B 落地 | — | Plan B／C 已標明碼與落點（見「未結項」第 3 項、Plan B Task 4／Task 6 與 Plan C Task 4／Task 5／Task 9 的註記） |

**現況數字（以程式與產生檔為準，勿抄舊稿）**：碼 **21** 個（SYS 7／AUTH 7／PLAT 4／CUST 3）＝ `docs/error-codes.md`；基線 **98 行／233 呼叫點**＝ `wc -l internal/services/errcode_baseline.txt` 與該檔筆數欄之和；**已落點 14 碼**（`AUTH-3002/3003/3004/4001/4002/4003`、`CUST-1001`、`SYS-1001/2001/3001/3002/4001/4002/9000`）。

### 執行期間的計畫更正（已寫回內文）

| # | 更正 | 理由 |
|---|---|---|
| 1 | 執行順序改為 **T1 → T3 → T2 → T4…** | `code.go` 原本硬相依 `requestid`（T3 產物） |
| 2 | **`trace_id` 改由邊界補**（不再是 `Error(ctx)`） | 省下 176 處呼叫點的機械改動；`SYS-9000` 樣板移除 `{trace}`，消除「缺參數外洩字面 `{trace}`」整類問題 |
| 3 | 首批碼 **18 → 20 → 21** | T4 前置補 `SYS-3001`／`SYS-3002`（保住 Plan A T11 的 42501 語意）；T5 收尾新增 `AUTH-3004`（首登閘門） |
| 4 | `AUTH-1001`／`AUTH-1002` → **`AUTH-4003`／`AUTH-3003`** | 原 ID 落在 `1xxx`（只允許 `InvalidArgument`）→ 照抄會在 init panic |
| 5 | 基線估算 **282 處 → 98 行／233 呼叫點** | T6 的實際掃描才是真相；`grep` 的粒度含行號，與守門鍵不同 |
| 6 | T5 拆為 **T5（可執行）＋ T5b（延後）** | `internal/platform/**` 是 Plan B／C 產物，執行時不存在 |
| 7 | 守門鍵改 **`path:歸屬名:筆數`** | 「同函式新增第二個未註冊呼叫」原本不會被擋（61% 基線鍵屬此類） |

### 未結項（deferred：具名、現況、選項、歸屬）

1. **公司停用閘門的對外碼不一致**（歸屬：**auth／spec 擁有者**，Plan D 不單方面改）
   - 現況：middleware 閘門（`internal/server/server.go` 的 `authzMiddleware`）回**裸 `unauthenticated`／HTTP 401**，`server_test.go` 明文釘住 401，且該處註解說明「不刪 session、公司恢復後可續用」；登入路徑則回 `AUTH-4002` `AuthCompanyInactive`／`permission_denied`／HTTP 403。
   - 選項：(a) 改閘門＋測試（動既有對外 HTTP 狀態與前端 401 處理）；(b) 另立一個 Unauthenticated 語意的公司停用碼（會把不一致固化成兩個碼）。
   - 已寫進 spec §4.3「已知不一致」。**現況更正**：該閘門回的是**裸 connect 錯誤**（不帶 `ErrorInfo`）→ 回應 body 只有 `code`／`message`，**沒有** `details`／`trace_id`（`requestid.Stamp` 只為帶 `ErrorInfo` 的錯誤補 trace_id）。
2. **`httpStatusForCode` 的映射不完整**（歸屬：**auth／spec 擁有者**，與上項一併裁定）
   - 現況：`internal/server/server.go` 只映射 unauthenticated→401／permission_denied→403／invalid_argument→400，其餘**一律 500**；故「首登受限」閘門（`AUTH-3004`）實際回 **HTTP 500**，`not_found`／`already_exists` 亦同。碼與訊息正確，僅 HTTP 狀態不符。
   - **規格值可查證**：Connect 規格（<https://connectrpc.com/docs/protocol>「Error codes」表）與 connect-go v1.21.0 的 `connectCodeToHTTP`（`protocol_connect.go`）都把 `failed_precondition` 映射為 **400**（**不是 412**——原 T5 註記的 412 與規格不符，以規格與 connect-go 為準）；`not_found`→404、`already_exists`→409。
   - 選項：補完對照表（會動既有對外 HTTP 狀態，需前端同步確認）；或維持現狀並在 spec 明載。
3. **未落點的已註冊碼**（歸屬見各項）
   - `AUTH-3001` `AuthRegistrationRequired`：**無語意相符且可達的路徑**（OIDC callback 是 redirect 而非錯誤；`completeGuest` 非 guest 分支語意相反；`subjectFromUser` 的未歸屬公司分支是 FK 不可能出現的防禦分支）→ 歸屬 **auth 流程**（identity-access）；正解是新增碼而非借用，故本計畫保留不落點。
   - `CUST-2001` `CustomerCodeExists`／`CUST-3001` `CustomerDeleted`：**編號由 counter 自動取號**（無「已存在」前置檢查，真撞號只會是 constraint → `SYS-3002`）；已刪除者以 `DeletedAtIsNil` 過濾 → 一律收斂為 `SYS-4002`，這是 Plan A 刻意的 anti-oracle 設計。**保留不落點**，理由已寫在 `codes_customer.go` 的註解；長期不用應考慮標 deprecated（碼不得重用或改義）。
   - `PLAT-*` ×4：**全部未落點**，落地點見 Task 5b（Plan B／C 已標明各碼與應在哪裡用）。
4. **`platform-console/src/lib/errcode.ts`**（歸屬：**Plan C**）
   - 現況：目錄尚不存在 → 產生器跳過並印提示（no-op）；CI 的「Error codes up to date」已把該路徑列入檢查，Plan C 一落地就會自動產生並要求 commit。
   - 已寫進 Plan C Task 14 的 §11 慣例追加（第 14 條）。

---

## File Structure

| 路徑 | 職責 |
|---|---|
| `backend/proto/salesorder/v1/common.proto`（改） | 追加 `ErrorInfo`（code／details／trace_id） |
| `internal/errcode/code.go`（新） | `Code` 型別、`register()`、`Error()`／`Wrap()`／`Render()`；區段規則驗證 |
| `internal/errcode/codes_{sys,auth,platform,customer}.go`（新） | 首批碼（分域檔，一個域一檔） |
| `internal/errcode/registry_test.go`（新） | 格式／唯一性／區段／渲染／panic 行為 |
| `internal/errcode/generate.go`（新） | `go:generate`：產生 `docs/error-codes.md` 與三端常數 |
| `internal/obs/requestid/requestid.go`（新） | `trace_id` interceptor ＋ `From(ctx)` |
| `internal/services/company_service.go`（改） | `toConnectError` 改走 registry（ent 映射、5xx → `SYS-9000`） |
| `internal/handlers/auth_handler.go`、`auth_password.go`、`internal/services/{shared,company}_service.go`（改） | 首批碼落地（auth／權限） |
| `internal/platform/{entitlements,billing}/*.go`（改） | 配額與收款錯誤改碼（`PLAT-*`） |
| `internal/services/customer_service.go`（改） | 樣板域（`CUST-*`） |
| `internal/services/errcode_guard_test.go`＋`errcode_baseline.txt`（新） | 基線守門 |
| `docs/error-codes.md`、`frontend/src/lib/errcode.ts`、`platform-console/src/lib/errcode.ts`、`app/lib/gen/errcode.dart`（產生） | 碼表與三端常數 |
| `backend/AGENTS.md`、`.github/workflows/ci.yml`、spec、Plan B／C、排序文件（改） | 慣例與對齊 |

---

### Task 1: `ErrorInfo` proto 與三端生成

**Files:**
- Modify: `backend/proto/salesorder/v1/common.proto`
- Generated: `backend/internal/proto/salesorder/v1/*`、`frontend/src/lib/proto/salesorder/v1/*`、`app/lib/gen/salesorder/v1/*`

- [ ] **Step 1: 追加訊息（`common.proto`）**

```proto
// ErrorInfo:錯誤的結構化資訊,由 connect error detail 攜帶(不引入 googleapis 依賴)。
// 前端以 code 查本地文案;message 為後端渲染的繁中訊息(log 與 fallback 用);
// trace_id 供客服回報時對照 server log。
message ErrorInfo {
  string code = 1;                    // 例:"CUST-2001"
  string message = 2;                 // 已渲染的繁中訊息
  map<string, string> details = 3;    // 結構化參數,例:{used:"10", limit:"10"}
  string trace_id = 4;                // 伺服器端請求追蹤 id
}
```

- [ ] **Step 2: 生成三端型別**

```bash
cd backend
export PATH="$HOME/fvm/versions/stable/bin:$HOME/.pub-cache/bin:$PATH"
task proto:gen
git status --short   # 應出現三端的 common 生成檔變更
```

- [ ] **Step 3: 確認可編譯**

Run: `cd backend && go build ./... && cd ../frontend && pnpm typecheck`
Expected: 皆通過

- [ ] **Step 4: Commit**

```bash
git add backend/proto/salesorder backend/internal/proto frontend/src/lib/proto app/lib/gen
git commit -m "feat(proto): 追加 ErrorInfo（code/details/trace_id）三端生成"
```

---

### Task 2: `internal/errcode` registry 與首批碼

**Files:**
- Create: `internal/errcode/code.go`、`codes_sys.go`、`codes_auth.go`、`codes_platform.go`、`codes_customer.go`
- Create: `internal/errcode/registry_test.go`

**[已落地：以 `internal/errcode` 實作為準（commit `3e095fe`＋`28a6d4e`）]** 本段下方的程式碼片段是**設計原稿**，實作已在其上收斂；後續任務（T4–T7）請以**程式碼**為唯一真相來源，不要從本文件抄簽章或欄位：
- `Code` 的欄位**全部未匯出**（`id`／`domain`／`connectCode`／`message`／`deprecated`），外部只能以存取子 `ID()`／`Domain()`／`ConnectCode()`／`Message()`／`IsDeprecated()` 讀取 → 「未註冊的碼無法建構」由**編譯期**保證（外部套件寫 `errcode.Code{ID:…}` 會編譯失敗；已實測）。
- **零值守門**：零值 `Code` 是外部唯一還造得出來的，`Render` 對它直接 panic（訊息指明「未經 MustRegister 的零值 Code」）——程式錯誤要在最早可達點爆掉，不要回一個沒有碼、沒有訊息的錯誤。
- `Render` 以**樣板佔位符集合**判定缺參數（不是看渲染結果含不含 `{`），且逐佔位符替換（不會二次替換剛帶入的值）。
- 測試切兩檔：`code_internal_test.go`（package `errcode`，放需要構造不合法碼的 panic 案例）與 `registry_test.go`（外部套件，測契約 1／3／4／6）。
- `AUTH` 首批碼的 ID 為 `AUTH-4003`（帳密錯誤）／`AUTH-3003`（鎖定）——原稿的 `AUTH-1001`／`AUTH-1002` 違反 1xxx 區段規則、會 init panic（見該檔註解）。

**以下是設計原稿（僅供理解意圖）：**

**Interfaces:**
- Produces:

```go
type Domain string

const (
	DomainSys      Domain = "SYS"
	DomainAuth     Domain = "AUTH"
	DomainPlatform Domain = "PLAT"
	DomainCustomer Domain = "CUST"
)

type Code struct {
	ID          string       // "CUST-2001"
	Domain      Domain
	ConnectCode connect.Code // 對外 connect 碼（由區段決定，register 驗證）
	Message     string       // 繁中樣板，可含 {name}
	Deprecated  bool
}

// 注意：Error／Wrap **不收 ctx**——trace_id 由 T3 的 requestid interceptor 在**邊界**補進
// ErrorInfo（見 Task 3 Step 3b）。理由：`toConnectError` 有 176 個呼叫點，逐點傳 ctx 是
// 176 處機械改動且會與 T5 同檔衝突；而「當前請求的 trace_id」本來就只有邊界知道。
func (c Code) Error(params map[string]string) *connect.Error
func (c Code) Wrap(err error, params ...map[string]string) *connect.Error
func (c Code) Render(params map[string]string) string
func Lookup(id string) (Code, bool)
func All() []Code // 已排序，供產生器與文件
```

**[修正] 相依與順序**：本任務**不** import `requestid`（trace_id 改由邊界補，見 Task 3 Step 3b）→ **T2 不再相依 T3**。（T3 已完成，故現況無影響；此註記是為了保留設計理由。）

- [ ] **Step 1: 寫失敗測試（`registry_test.go`）**

```go
package errcode_test

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	commonv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
)

var idFormat = regexp.MustCompile(`^[A-Z]{2,6}-\d{4}$`)

// errorInfoOf 取出 connect error detail 內的 ErrorInfo（找不到即失敗）。
func errorInfoOf(t *testing.T, err *connect.Error) *commonv1.ErrorInfo {
	t.Helper()
	for _, d := range err.Details() {
		if m, ok := d.(proto.Message); ok {
			if ei, ok := m.(*commonv1.ErrorInfo); ok {
				return ei
			}
		}
	}
	t.Fatalf("錯誤未帶 ErrorInfo detail")
	return nil
}

// 契約 1：每個碼格式正確、有訊息、有 domain，且區段與 connect 碼一致。
func TestRegistryInvariants(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range errcode.All() {
		if !idFormat.MatchString(c.ID) {
			t.Fatalf("碼格式錯誤: %q", c.ID)
		}
		if seen[c.ID] {
			t.Fatalf("碼重複: %s", c.ID)
		}
		seen[c.ID] = true
		if strings.TrimSpace(c.Message) == "" {
			t.Fatalf("%s 缺訊息", c.ID)
		}
		if !strings.HasPrefix(c.ID, string(c.Domain)+"-") {
			t.Fatalf("%s 的 domain %s 與碼前綴不符", c.ID, c.Domain)
		}
		if !errcode.SectionAllows(c.ID, c.ConnectCode) {
			t.Fatalf("%s 的 connect 碼 %v 不屬於區段 %s 的允許集合", c.ID, c.ConnectCode, c.ID[:1])
		}
	}
}

// 契約 2：註冊重複 / 格式錯 / 區段不符 → panic（啟動就失敗，不等上線）。
func TestRegisterPanicsOnViolations(t *testing.T) {
	cases := []struct {
		name string
		fn   func()
	}{
		{"重複 ID", func() {
			errcode.MustRegister(errcode.Code{ID: "SYS-1001", Domain: errcode.DomainSys,
				ConnectCode: connect.CodeInvalidArgument, Message: "x"})
		}},
		{"格式錯誤", func() {
			errcode.MustRegister(errcode.Code{ID: "SYS-1", Domain: errcode.DomainSys,
				ConnectCode: connect.CodeInvalidArgument, Message: "x"})
		}},
		{"區段與 connect 碼不符", func() {
			errcode.MustRegister(errcode.Code{ID: "SYS-1234", Domain: errcode.DomainSys,
				ConnectCode: connect.CodeInternal, Message: "x"})
		}},
		{"缺訊息", func() {
			errcode.MustRegister(errcode.Code{ID: "SYS-1234", Domain: errcode.DomainSys,
				ConnectCode: connect.CodeInvalidArgument})
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("應 panic")
				}
			}()
			tc.fn()
		})
	}
}

// 契約 3：Error() 產生帶 ErrorInfo 的 connect error（碼與 details 可被客戶端讀取）。
func TestErrorCarriesErrorInfo(t *testing.T) {
	ctx := t.Context()
	err := errcode.PlatformLimitExceeded.Error(map[string]string{"used": "10", "limit": "10"})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("connect 碼應為 FailedPrecondition，got %v", connect.CodeOf(err))
	}
	if !strings.Contains(err.Message(), "10/10") {
		t.Fatalf("訊息應已渲染參數，got %q", err.Message())
	}
	var info *commonv1.ErrorInfo
	for _, d := range err.Details() {
		if m, ok := d.(proto.Message); ok {
			if ei, ok := m.(*commonv1.ErrorInfo); ok {
				info = ei
			}
		}
	}
	if info == nil || info.GetCode() != "PLAT-5001" {
		t.Fatalf("應帶 ErrorInfo{code: PLAT-5001}，got %+v", info)
	}
	if info.GetDetails()["used"] != "10" {
		t.Fatalf("details 應帶 used=10，got %+v", info.GetDetails())
	}
}

// 契約 4：缺參數不得讓錯誤處理爆掉（回樣板原文），且 Wrap 保留底層錯誤供 log 追查。
func TestRenderMissingParamFallsBackAndWrapKeepsCause(t *testing.T) {
	ctx := t.Context()
	msg := errcode.PlatformLimitExceeded.Render(nil)
	if !strings.Contains(msg, "{used}") {
		t.Fatalf("缺參數時應保留樣板原文，got %q", msg)
	}
	cause := errors.New("db: connection reset")
	err := errcode.SysInternal.Wrap(cause)
	if !errors.Is(err, cause) {
		t.Fatal("Wrap 應保留底層錯誤（Unwrap），否則 log 追不到原因")
	}
	if strings.Contains(err.Message(), "connection reset") {
		t.Fatal("底層錯誤細節不得進對外訊息")
	}
}

// 契約 5（trace_id）已移至 Task 3 的邊界測試：本套件不感知 ctx。
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/errcode/ -v`
Expected: FAIL —`undefined: errcode.All`

- [ ] **Step 3: 實作（`code.go`）**

```go
// Package errcode 為對外錯誤碼的唯一真相來源。
// 設計要點：常數即註冊（register 於 init 驗證格式／唯一性／區段與 connect 碼的一致性，
// 違反即 panic → 啟動就失敗）；碼發佈後不得重用或改義，廢止只標 Deprecated。
package errcode

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"connectrpc.com/connect"

	commonv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
)

type Domain string

const (
	DomainSys      Domain = "SYS"
	DomainAuth     Domain = "AUTH"
	DomainPlatform Domain = "PLAT"
	DomainCustomer Domain = "CUST"
)

// Code 為一筆錯誤碼定義。
type Code struct {
	ID          string
	Domain      Domain
	ConnectCode connect.Code
	Message     string
	Deprecated  bool
}

var (
	idPattern = regexp.MustCompile(`^[A-Z]{2,6}-(\d{4})$`)
	registry  = map[string]Code{}
)

// MustRegister 註冊一個碼；違反下列任一即 panic（啟動時暴露，而非上線後才發現）：
// 格式不符、ID 重複、缺訊息、domain 與前綴不符、connect 碼與區段規則不符。
func MustRegister(c Code) Code {
	m := idPattern.FindStringSubmatch(c.ID)
	if m == nil {
		panic(fmt.Sprintf("errcode: 碼格式錯誤 %q（應為 域-4位數）", c.ID))
	}
	if _, dup := registry[c.ID]; dup {
		panic(fmt.Sprintf("errcode: 碼重複註冊 %q", c.ID))
	}
	if strings.TrimSpace(c.Message) == "" {
		panic(fmt.Sprintf("errcode: %q 缺訊息", c.ID))
	}
	if !strings.HasPrefix(c.ID, string(c.Domain)+"-") {
		panic(fmt.Sprintf("errcode: %q 與 domain %q 不符", c.ID, c.Domain))
	}
	if !allowedConnectCodes(mustSection(m[1], c.ID)).Contains(c.ConnectCode) {
		panic(fmt.Sprintf("errcode: %q 的 connect 碼 %v 與區段 %s 的允許集合不符",
			c.ID, c.ConnectCode, m[1][:1]))
	}
	registry[c.ID] = c
	return c
}

// mustSection 由四位數取出區段（首位數字）；格式問題即 panic（與 register 的契約一致）。
func mustSection(digits, id string) int {
	n, err := strconv.Atoi(digits)
	if err != nil {
		panic(fmt.Sprintf("errcode: %q 的區段非數字", id))
	}
	return n / 1000
}

// allowedConnectCodes 依區段首位數字回傳允許的 connect 碼集合。
// 這是硬規則：碼的分類必須與對外 connect 碼一致，否則前端無法以區段推斷處理方式。
func allowedConnectCodes(section int) connectCodeSet { return sectionRules[section] }

// SectionAllows 供測試與產生器斷言「此碼的 connect code 落在其區段允許集合內」。
// 4xxx 有三個允許值（PermissionDenied／Unauthenticated／NotFound），故不能只比對單一值。
func SectionAllows(id string, cc connect.Code) bool {
	m := idPattern.FindStringSubmatch(id)
	if m == nil {
		return false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return false
	}
	return sectionRules[n/1000].Contains(cc)
}

type connectCodeSet struct {
	codes []connect.Code
}

func (s connectCodeSet) Contains(c connect.Code) bool {
	for _, x := range s.codes {
		if x == c {
			return true
		}
	}
	return false
}

var sectionRules = map[int]connectCodeSet{
	1: {codes: []connect.Code{connect.CodeInvalidArgument}},
	2: {codes: []connect.Code{connect.CodeAlreadyExists}},
	3: {codes: []connect.Code{connect.CodeFailedPrecondition}},
	4: {codes: []connect.Code{connect.CodePermissionDenied, connect.CodeUnauthenticated, connect.CodeNotFound}},
	5: {codes: []connect.Code{connect.CodeFailedPrecondition}},
	9: {codes: []connect.Code{connect.CodeInternal}},
}

// Lookup 以 ID 取碼。
func Lookup(id string) (Code, bool) {
	c, ok := registry[id]
	return c, ok
}

// All 回傳全部碼（依 ID 排序，供產生器與文件）。
func All() []Code {
	out := make([]Code, 0, len(registry))
	for _, c := range registry {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Render 以參數渲染訊息樣板；缺參數時回樣板原文並記 log（不得讓錯誤處理本身爆掉）。
func (c Code) Render(params map[string]string) string {
	msg := c.Message
	missing := false
	for k, v := range params {
		msg = strings.ReplaceAll(msg, "{"+k+"}", v)
	}
	if strings.Contains(msg, "{") {
		missing = true
	}
	if missing {
		log.Printf("errcode: %s 訊息缺參數(已回樣板原文): %q params=%v", c.ID, c.Message, params)
	}
	return msg
}

// Error 產生帶 ErrorInfo 的 connect error（碼、details、trace_id 一併帶出）。
// ctx 用於取 trace_id：requestid interceptor 注入；未注入時 trace_id 為空（僅影響客服追查）。
func (c Code) Error(ctx context.Context, params map[string]string) *connect.Error {
	msg := c.Render(params)
	err := connect.NewError(c.ConnectCode, errors.New(msg))
	info := &commonv1.ErrorInfo{Code: c.ID, Message: msg, TraceId: requestid.From(ctx)}
	if len(params) > 0 {
		info.Details = params
	}
	if detail, derr := connect.NewErrorDetail(info); derr == nil {
		err.AddDetail(detail)
	} else {
		log.Printf("errcode: %s 附掛 ErrorInfo 失敗（回應仍帶 connect 碼）: %v", c.ID, derr)
	}
	return err
}

// Wrap 同 Error，但保留底層錯誤（Unwrap）供 log 追查；**cause 不進入對外 message**。
func (c Code) Wrap(ctx context.Context, cause error, params ...map[string]string) *connect.Error {
	var p map[string]string
	if len(params) > 0 {
		p = params[0]
	}
	msg := c.Render(p)
	// 為什麼不用 fmt.Errorf("%s: %w", msg, cause)：那會把 cause 的文字寫進 message，
	// 而 connect-go 對任何 code 都逐字轉送 connectErr.Message()（T4 review 實證：
	// 客戶端會收到 SQLSTATE／policy 名／SET LOCAL 語句文字）。改以自訂型別承載：
	// 對外只看到 msg，伺服器端仍可 errors.Is／Unwrap 追根因（根因另由 log 記錄）。
	err := connect.NewError(c.ConnectCode, &wrapped{msg: msg, cause: cause})
	info := &commonv1.ErrorInfo{Code: c.ID, Message: msg, TraceId: requestid.From(ctx)}
	if len(p) > 0 {
		info.Details = p
	}
	if detail, derr := connect.NewErrorDetail(info); derr == nil {
		err.AddDetail(detail)
	} else {
		log.Printf("errcode: %s 附掛 ErrorInfo 失敗（回應仍帶 connect 碼）: %v", c.ID, derr)
	}
	return err
}

// wrapped 讓對外訊息與根因分離：Error() 只回對外訊息，Unwrap() 保留根因供 log 追查。
type wrapped struct {
	msg   string
	cause error
}

func (e *wrapped) Error() string { return e.msg }
func (e *wrapped) Unwrap() error { return e.cause }
```

**兩個實作要點**（已反映在上方程式碼，實作時照抄即可）：

1. `Error`／`Wrap` **不收 ctx**；`trace_id` 由 requestid interceptor 在邊界補進 `ErrorInfo`（Task 3 Step 3b）。呼叫端一律 `errcode.SysNotFound.Error(nil)`。
2. 附掛 ErrorInfo 失敗時不得讓錯誤處理失效（記 log，仍回 connect 碼與訊息）——錯誤路徑上的二次失敗最難追。

**[修正] SYS-9000 的樣板**：`SysInternal` 的訊息**不含 `{trace}`**（改為「系統忙碌，請稍後再試」）。理由：`trace_id` 已由 `ErrorInfo.trace_id` 這個**結構化欄位**承載（前端與客服端可直接顯示），把 `{trace}` 塞進訊息樣板反而要求每個呼叫點都先注入參數、否則外洩字面 `{trace}`（原設計的缺口）。客服要看的代碼請由回應的 `ErrorInfo.trace_id` 取。

- [ ] **Step 4: 首批碼（四個分域檔）**

```go
// codes_sys.go
package errcode

import "connectrpc.com/connect"

var (
	// SysInvalidArgument 為無專碼可用的參數驗證失敗（新增端點時優先定義專碼）。
	SysInvalidArgument = MustRegister(Code{ID: "SYS-1001", Domain: DomainSys,
		ConnectCode: connect.CodeInvalidArgument, Message: "參數驗證失敗"})

	// SysConflict 為**已知的**識別碼重複（例：CreateCompany 以 DeletedAtIsNil 前置查詢判定後回此碼）。
	// 為什麼不由 DB 約束錯誤推導：ent 的 constraint 錯誤無法分辨「識別碼重複」與「FK 阻擋」，
	// 把後者回成 AlreadyExists 正是 P2-A 的原始缺陷（見 SysConstraintViolation）。
	SysConflict = MustRegister(Code{ID: "SYS-2001", Domain: DomainSys,
		ConnectCode: connect.CodeAlreadyExists, Message: "資料衝突，請確認識別碼是否已被使用"})

	// SysScopeViolation 為寫入被 RLS 的 WITH CHECK 擋下（資料範圍不符）。
	// 為什麼是 FailedPrecondition 而非 Internal：這是「身分／範圍與該列不匹配」，不是伺服器故障
	// （用 Internal 會讓監控誤判 5xx 並誤導客戶）；SQLSTATE 與 policy 原文只進 log。
	SysScopeViolation = MustRegister(Code{ID: "SYS-3001", Domain: DomainSys,
		ConnectCode: connect.CodeFailedPrecondition, Message: "資料超出目前的存取範圍,無法完成此操作"})

	// SysConstraintViolation 為資料庫約束類錯誤（識別碼重複、FK 阻擋、CHECK 失敗）——無法分辨是哪一種。
	// 為什麼不是 AlreadyExists：P2-A 的原始缺陷正是「FK 阻擋被當成識別碼重複回 AlreadyExists」，
	// 且 ent 給的 constraint 錯誤無法區分兩者；訊息刻意保留「請確認…」的可行動指引但不揭露 DB 細節。
	SysConstraintViolation = MustRegister(Code{ID: "SYS-3002", Domain: DomainSys,
		ConnectCode: connect.CodeFailedPrecondition,
		Message:     "資料違反資料庫約束,無法完成此操作(請確認識別碼是否已被使用、參照對象是否仍存在)"})

	// SysPermissionDenied 為授權檢查失敗（角色/資料範圍不足）。前端應導向「請管理員開權」。
	SysPermissionDenied = MustRegister(Code{ID: "SYS-4001", Domain: DomainSys,
		ConnectCode: connect.CodePermissionDenied, Message: "缺少權限"})

	// SysNotFound 同時代表「不存在」與「不在可見範圍內」——刻意不區分，避免以錯誤碼
	// 探測他租戶資源是否存在（oracle）。跨租戶查詢因範圍過濾而查不到時一律用此碼。
	SysNotFound = MustRegister(Code{ID: "SYS-4002", Domain: DomainSys,
		ConnectCode: connect.CodeNotFound, Message: "資源不存在或無權存取"})

	// SysInternal 為所有 5xx：對外只給碼與訊息；trace_id 由 ErrorInfo 的結構化欄位承載
	// （客服回報用），**不**寫進訊息樣板（否則每個呼叫點都得先注入參數、漏了就外洩字面 {trace}）。
	SysInternal = MustRegister(Code{ID: "SYS-9000", Domain: DomainSys,
		ConnectCode: connect.CodeInternal, Message: "系統忙碌，請稍後再試"})
)
```

```go
// codes_auth.go
package errcode

import "connectrpc.com/connect"

var (
	// AuthBadCredentials 刻意不區分「帳號不存在」與「密碼錯誤」（防帳號列舉）。
	// ID 為 AUTH-4003（不是原稿的 AUTH-1001）：區段規則 1xxx 只允許 InvalidArgument，
	// 而本碼對外是 Unauthenticated → 只有 4xxx 允許，否則 `MustRegister` 於 init panic
	// （T2 實跑證實）。語意與 connect 碼不變。
	AuthBadCredentials = MustRegister(Code{ID: "AUTH-4003", Domain: DomainAuth,
		ConnectCode: connect.CodeUnauthenticated, Message: "帳號或密碼錯誤"})

	// AuthLocked 帶 details.until（解鎖時間）。
	// ID 為 AUTH-3003（不是原稿的 AUTH-1002）：理由同 AuthBadCredentials（FailedPrecondition 屬 3xxx）。
	AuthLocked = MustRegister(Code{ID: "AUTH-3003", Domain: DomainAuth,
		ConnectCode: connect.CodeFailedPrecondition, Message: "帳號已鎖定，請於 {until} 後再試"})

	AuthRegistrationRequired = MustRegister(Code{ID: "AUTH-3001", Domain: DomainAuth,
		ConnectCode: connect.CodeFailedPrecondition, Message: "尚未完成註冊"})

	AuthTempPasswordExpired = MustRegister(Code{ID: "AUTH-3002", Domain: DomainAuth,
		ConnectCode: connect.CodeFailedPrecondition, Message: "臨時密碼已過期，請聯繫管理員重置"})

	AuthUnauthenticated = MustRegister(Code{ID: "AUTH-4001", Domain: DomainAuth,
		ConnectCode: connect.CodeUnauthenticated, Message: "未登入"})

	// AuthCompanyInactive 為公司停用連鎖（含欠費凍結）。
	AuthCompanyInactive = MustRegister(Code{ID: "AUTH-4002", Domain: DomainAuth,
		ConnectCode: connect.CodePermissionDenied, Message: "所屬公司已停用"})
)
```

```go
// codes_platform.go（SaaS 配額與訂閱；Plan B／C 直接使用）
package errcode

import "connectrpc.com/connect"

var (
	// PlatformSubscriptionInactive 為 suspended／cancelled（合約狀態問題，非權限問題）。
	PlatformSubscriptionInactive = MustRegister(Code{ID: "PLAT-3001", Domain: DomainPlatform,
		ConnectCode: connect.CodeFailedPrecondition, Message: "訂閱狀態不允許此操作"})

	// PlatformLimitExceeded 帶 details{feature, used, limit}；前端導向升級方案。
	PlatformLimitExceeded = MustRegister(Code{ID: "PLAT-5001", Domain: DomainPlatform,
		ConnectCode: connect.CodeFailedPrecondition, Message: "已達方案上限（{used}/{limit}），請升級方案"})

	// PlatformFeatureNotInPlan 帶 details{feature}。
	PlatformFeatureNotInPlan = MustRegister(Code{ID: "PLAT-5002", Domain: DomainPlatform,
		ConnectCode: connect.CodeFailedPrecondition, Message: "目前方案未包含此功能，請升級方案"})

	// PlatformPaymentConflict 為收款衝突（期別已付款、金額不符）。
	PlatformPaymentConflict = MustRegister(Code{ID: "PLAT-3002", Domain: DomainPlatform,
		ConnectCode: connect.CodeFailedPrecondition, Message: "收款衝突：{reason}"})
)
```

```go
// codes_customer.go（樣板域：示範 1xxx/2xxx/3xxx 三類寫法，其餘域照抄）
package errcode

import "connectrpc.com/connect"

var (
	CustomerNameRequired = MustRegister(Code{ID: "CUST-1001", Domain: DomainCustomer,
		ConnectCode: connect.CodeInvalidArgument, Message: "客戶名稱不可為空"})

	CustomerCodeExists = MustRegister(Code{ID: "CUST-2001", Domain: DomainCustomer,
		ConnectCode: connect.CodeAlreadyExists, Message: "客戶編號 {code} 已存在"})

	CustomerDeleted = MustRegister(Code{ID: "CUST-3001", Domain: DomainCustomer,
		ConnectCode: connect.CodeFailedPrecondition, Message: "客戶已刪除，無法更新"})
)
```

- [ ] **Step 5: 跑測試確認通過**

Run: `cd backend && go test ./internal/errcode/ -v`
Expected: PASS（含 panic 行為與 ErrorInfo 攜帶）

- [ ] **Step 6: Commit**

```bash
git add backend/internal/errcode
git commit -m "feat(errcode): 錯誤碼 registry（常數即註冊、啟動驗證、區段規則）與首批 20 碼"
```

---

### Task 3: `trace_id` interceptor

**Files:**
- Create: `internal/obs/requestid/requestid.go`、`requestid_test.go`
- Modify: `internal/server/domains.go`（每個 handler 的 options 加上此 interceptor，順序在 dbtenant 之前）

**Interfaces:**
- Produces: `requestid.Interceptor() connect.Interceptor`、`requestid.From(ctx) string`、`requestid.With(ctx, id) context.Context`

- [ ] **Step 1: 寫失敗測試**

```go
package requestid_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
)

// 同一請求內 trace_id 必須一致且有值；不同請求必須不同。
func TestInterceptorInjectsStableTraceID(t *testing.T) {
	seen := []string{}
	handler := func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		id := requestid.From(ctx)
		if id == "" {
			t.Error("trace_id 不得為空")
		}
		seen = append(seen, id)
		return nil, nil
	}
	call := requestid.Interceptor().WrapUnary(handler)
	for range 2 {
		if _, err := call(context.Background(), connect.NewRequest(&emptypb.Empty{})); err != nil {
			t.Fatalf("呼叫: %v", err)
		}
	}
	if seen[0] == seen[1] {
		t.Fatalf("不同請求應有不同 trace_id: %v", seen)
	}
}
```

- [ ] **Step 2: 實作**

```go
// Package requestid 為每個請求產生 trace_id：進 ctx（供錯誤碼與 log 使用）、
// 在錯誤回應中帶出（客服回報時可對照 server log）。
package requestid

import (
	"context"
	"log"

	"connectrpc.com/connect"
	"github.com/google/uuid"
)

type ctxKey struct{}

// With 將 trace_id 放入 ctx（測試與跨服務傳遞用）。
func With(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// From 取 trace_id；未注入時回空字串（呼叫端不得假設非空）。
func From(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

// Interceptor 為每個 unary RPC 產生 trace_id 並記一行 log（method + trace_id），
// 使「客戶回報代碼」能直接對到 log。
func Interceptor() connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			id := uuid.Must(uuid.NewV7()).String()
			ctx = With(ctx, id)
			log.Printf("rpc: %s trace_id=%s", req.Spec().Procedure, id)
			return next(ctx, req)
		}
	})
}
```

（`google/uuid` 已在 `go.mod`；若無則 `go get github.com/google/uuid`。`spec.Procedure` 為 connect-go 的方法全名。）

**Step 2b: 邊界補 `trace_id`（本計畫的關鍵設計）**

`errcode` 刻意不收 ctx（見 Task 2 的 Interfaces 註記），`trace_id` 由本 interceptor 在**回應邊界**補進錯誤的 `ErrorInfo`：

```go
// Interceptor 內：呼叫 next 後補 trace_id。
res, err := next(ctx, req)
return res, stampTraceID(ctx, err)

// stampTraceID 為「已帶 ErrorInfo 的錯誤」補上本請求的 trace_id；其他錯誤原樣回。
// 為什麼在邊界補：錯誤碼有 176 個生產呼叫點，逐點傳 ctx 是無謂的機械改動；且「當前請求的
// trace_id」本來就只有邊界知道。
func stampTraceID(ctx context.Context, err error) error {
	id := From(ctx)
	if id == "" || err == nil {
		return err
	}
	ce, ok := err.(*connect.Error)
	if !ok {
		return err
	}
	var info *commonv1.ErrorInfo
	for _, d := range ce.Details() {
		// 注意：Details() 回傳 []*connect.ErrorDetail（不是 proto.Message）→ 用 Value() 取回 proto。
		if ei, ok := d.Value().(*commonv1.ErrorInfo); ok {
			info = ei
		}
	}
	if info == nil || info.GetTraceId() != "" {
		return err
	}
	// 重建（不賭就地修改）：複製其他 detail 與 Meta，再附上補好 trace 的 ErrorInfo。
	ne := connect.NewError(ce.Code(), errors.New(ce.Message()))
	for k, vs := range ce.Meta() {
		for _, v := range vs {
			ne.Meta().Add(k, v)
		}
	}
	stamped := proto.Clone(info).(*commonv1.ErrorInfo)
	stamped.TraceId = id
	if det, derr := connect.NewErrorDetail(stamped); derr == nil {
		ne.AddDetail(det)
	} else {
		return err
	}
	return ne
}
```

⚠️ **T3b 已實測定案：就地修改在本版本（connect-go v1.21.0）**不可行**——`ErrorDetail.Value()` 回傳 `proto.Clone`、序列化走 `NewErrorDetail` 當下 marshal 的 `pbAny`（先用就地版取 RED：客戶端收到 trace_id=""）。以下疑慮已無懸念，一律採「重建」。

⚠️ **必須以測試確認「就地修改真的傳得出去」**：connect-go 的 `Details()` 是否回傳原指標（改了就生效）或序列化副本（改了沒用）依版本而異。若就地修改無效，改為**重建錯誤**（`connect.NewError(ce.Code(), errors.New(ce.Message()))` ＋ `AddDetail(新 ErrorInfo)`）——不論哪一種，**驗收標準是「客戶端真的收到非空 trace_id」**：

```go
// 驗收（整合）：掛上 interceptor 的 handler 回 errcode.SysNotFound.Error(nil)，
// 客戶端收到的錯誤必須帶 ErrorInfo{code:"SYS-4002", trace_id:非空}，
// 且該 trace_id 與同請求 handler 內 requestid.From(ctx) 相同。
```



- [ ] **Step 3: 掛載（`internal/server/domains.go` 與各 `RegisterXServices`）**

**注意（T3 實測）**：生產掛載點共 **15 處**——`domains.go` 只有 2 個（auth／ability），其餘 13 個在 `internal/services/*.go` 的 `RegisterXServices` 與 `internal/handlers/role_handler.go`。**每一個**都要改（否則那些 RPC 沒有 trace_id）。

每個 `NewXServiceHandler(h, opts...)` 的 options 加上，順序：`connect.WithInterceptors(requestid.Interceptor(), dbtenant.Interceptor(entClient))`（trace_id 先產生，讓 dbtenant 的交易錯誤也帶得到）。

- [ ] **Step 4: 跑測試**

Run: `cd backend && go test ./internal/obs/... -v && task check`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/obs backend/internal/server/domains.go backend/go.mod backend/go.sum
git commit -m "feat(obs): trace_id interceptor（每請求一致、進 log 與錯誤回應）"
```

---

### Task 4: `toConnectError` 改走 registry

**Files:**
- Modify: `internal/services/company_service.go`（`toConnectError`，約 `:618`）
- Create: `internal/services/to_connect_error_test.go`

- [ ] **Step 1: 寫失敗測試（表驅動）**

```go
package services

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	"entgo.io/ent"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
)

// 每個 ent 錯誤型別 → 期望錯誤碼；改動映射表時這條會紅。
func TestToConnectErrorMapsToRegisteredCodes(t *testing.T) {
	cases := []struct {
		name string
		in   error
		code string
		cc   connect.Code
	}{
		{"not found", &ent.NotFoundError{}, "SYS-4002", connect.CodeNotFound},
		{"validation", &ent.ValidationError{}, "SYS-1001", connect.CodeInvalidArgument},
		{"constraint", &ent.ConstraintError{}, "SYS-3002", connect.CodeFailedPrecondition},
		{"rls 違反", &pgconn.PgError{Code: "42501", Message: `new row violates row-level security policy for table "customers"`}, "SYS-3001", connect.CodeFailedPrecondition},
		{"not singular", &ent.NotSingularError{}, "SYS-9000", connect.CodeInternal},
		{"unknown", errors.New("boom"), "SYS-9000", connect.CodeInternal},
	}
	// 另需一條「訊息不得外洩」斷言：rls／constraint／unknown 三種輸入的 Message()
	// 都不得含 "SQLSTATE"／"row-level security"／表名／約束名。
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := toConnectError(tc.in)
			if connect.CodeOf(got) != tc.cc {
				t.Fatalf("connect 碼 = %v；want %v", connect.CodeOf(got), tc.cc)
			}
			if code := errcodeCodeOf(t, got); code != tc.code {
				t.Fatalf("錯誤碼 = %s；want %s", code, tc.code)
			}
		})
	}
}

// errcodeCodeOf 由 connect error 的 detail 取錯誤碼（測試輔助）。
func errcodeCodeOf(t *testing.T, err error) string {
	t.Helper()
	ce, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("應為 *connect.Error，got %T", err)
	}
	for _, d := range ce.Details() {
		// 注意：Details() 回傳 []*connect.ErrorDetail（不是 proto.Message）→ 用 Value() 取回 proto。
		if info, ok := d.Value().(*commonv1.ErrorInfo); ok {
			return info.GetCode()
		}
	}
	return ""
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/services/ -run TestToConnectErrorMaps -v`
Expected: FAIL —目前 `toConnectError` 只回 connect 碼、無 detail

- [ ] **Step 3: 改寫 `toConnectError`**

```go
// toConnectError 將底層錯誤集中映射為「已註冊的錯誤碼」。
// 原則：DB 原始訊息（SQLSTATE／約束名／表名）只進 log，對外一律固定訊息；
// 伺服器端的意外（含 NotSingular 與未知錯誤）一律 SYS-9000（trace_id 由邊界補進 ErrorInfo）。
//
// **簽章不變**（不收 ctx）：本函式有 176 個生產呼叫點，逐點傳 ctx 是無謂的機械改動；
// trace_id 由 requestid interceptor 在回應邊界補（Task 3 Step 3b）。
func toConnectError(err error) error {
	if err == nil {
		return nil
	}
	// 已是錯誤碼（含 registry 產生的）→ 不重複包裝，避免碼被內層蓋掉。
	if ce, ok := err.(*connect.Error); ok {
		return ce
	}
	switch {
	case ent.IsNotFound(err):
		return errcode.SysNotFound.Error(nil)
	case ent.IsValidationError(err):
		return errcode.SysInvalidArgument.Error(nil)
	case isRLSPolicyViolation(err):
		// 保留 Plan A（T11）的語意：範圍不符 ≠ 伺服器故障，故 FailedPrecondition 而非 Internal。
		log.Printf("services: RLS 違反（已映射為 SYS-3001，細節不對外揭露）: %v", err)
		return errcode.SysScopeViolation.Error(nil)
	case ent.IsConstraintError(err):
		log.Printf("services: 資料庫約束錯誤（已映射為 SYS-3002，細節不對外揭露）: %v", err)
		return errcode.SysConstraintViolation.Error(nil)
	default:
		log.Printf("services: 內部錯誤（對外僅回 SYS-9000）: %v", err)
		return errcode.SysInternal.Error(nil)
	}
}
```

**兩個必須遵守的點**：①`isRLSPolicyViolation`（`company_service.go`）**保留**，不可刪——它攔的是 ent 的 constraint 判定**認不出**的 PG 42501（見該函式上方註解）；②本函式**不**把 `err` 附進回應（不用 `Wrap`）——DB 原文一旦掛在 `connect.Error` 上，任何後續路徑（log 中介層、detail 序列化）都有機會把它帶出去；根因已由上面的 `log.Printf` 落 server log。

（原稿要求「`Error(ctx, nil)` 並讓 `toConnectError` 收 ctx」——**已作廢**：簽章不變、trace_id 由邊界補。）

- [ ] **Step 4: 跑測試確認通過並收斂呼叫點**

```bash
cd backend && go build ./... 2>&1 | head -20   # 依編譯錯誤更新 toConnectError 的呼叫端簽章
go test ./internal/services/ -run TestToConnectErrorMaps -v
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/services
git commit -m "refactor(services): toConnectError 改走錯誤碼 registry（ent 映射、5xx 一律 SYS-9000）"
```

---

### Task 5: 首批碼落地（auth／權限／樣板域）

**[範圍修正（controller，開工前實查）]** 原稿的 Step 3 改 `internal/platform/{entitlements,store,billing}`，但**那些是 Plan B/C 的產物、目前不存在**（`backend/internal/platform/` 整個目錄不存在，`newUserTestServerWithEntitlement`／`NewEntitlementCounter` 亦不存在）→ 該步與對應測試 **延後**至 Plan B 落地後另立 **Task 5b** 執行。本任務只做**現在存在**的部分：`requireAuth`（`shared_service.go:26`）、`requireScope`（`company_service.go:86`）、登入／密碼路徑（`auth_handler.go`／`auth_password.go`）、樣板域（`customer_service.go` 三碼）。

**Files:**
- Modify: `internal/handlers/auth_handler.go`、`auth_password.go`
- Modify: `internal/services/shared_service.go`（`requireAuth`）、`company_service.go`（`requireScope`）
- Modify: `internal/services/customer_service.go`（樣板域三碼）

- [ ] **Step 1: 權限與未登入（`shared_service.go`、`company_service.go`）**

```go
// requireAuth：未登入
id := authz.IdentityFrom(ctx)
if id.UserID == "" {
	return authz.Identity{}, errcode.AuthUnauthenticated.Error(nil)
}
```

```go
// requireScope：授權檢查失敗（角色/範圍不足）——與「查不到」不同語意
if !ok {
	return errcode.SysPermissionDenied.Error(map[string]string{"resource": resource, "action": action})
}
```

（訊息樣板若要顯示資源名，把 `SYS-4001` 的 Message 改為 `"缺少 {resource} 資源的 {action} 權限"`；這是註冊時的樣板變更，測試需同步。）

- [ ] **Step 2: 登入路徑（`auth_handler.go`、`auth_password.go`）**

```go
// 帳密錯誤：不區分帳號不存在與密碼錯誤（防列舉）
if user == nil || !auth.CheckPasswordHash(password, user.PasswordHash) {
	return nil, errcode.AuthBadCredentials.Error(nil)
}
// 鎖定（帶解鎖時間）
if locked {
	return nil, errcode.AuthLocked.Error(map[string]string{
		"until": unlockAt.Format("15:04"),
	})
}
// 公司停用連鎖
if !companyActive {
	return nil, errcode.AuthCompanyInactive.Error(nil)
}
// 首登未完成／臨時密碼過期
return nil, errcode.AuthRegistrationRequired.Error(nil)
return nil, errcode.AuthTempPasswordExpired.Error(nil)
```

- [ ] **Step 3: 配額與收款 → 延後（見 Task 5b）**

原稿此步改 `internal/platform/entitlements`（`Allows`／`CheckLimit`）與 `internal/platform/billing`（金額不符、重複入帳），**該目錄目前不存在**（Plan B/C 產物）→ 本任務跳過，待 Plan B 落地後由 **Task 5b** 執行（該步的程式碼片段與 `PLAT-*` 測試案例一併保留在 5b）。

```go
// CheckLimit：未含功能 / 已達上限，各自帶 details 供前端導向升級
if !known || !r.enabled {
	return errcode.PlatformFeatureNotInPlan.Error(map[string]string{"feature": feature})
}
if cur+delta > int(*r.limit) {
	return errcode.PlatformLimitExceeded.Error(map[string]string{
		"feature": feature, "used": strconv.Itoa(cur), "limit": strconv.Itoa(*r.limit),
	})
}
// 訂閱不可用（suspended／cancelled）
return errcode.PlatformSubscriptionInactive.Error(nil)
```

```go
// billing：金額不符（G4）與期別已付款衝突（G3 的 log 分支升級為碼）
return errcode.PlatformPaymentConflict.Error(map[string]string{
	"reason": fmt.Sprintf("輸入金額 %s 與期別金額 %s 不符",
		money.FormatCents(in.AmountCents), money.FormatCents(period.AmountCents)),
})
```

- [ ] **Step 4: 樣板域（`customer_service.go`）**

三處示範：名稱空白 → `CustomerNameRequired`；編號已存在 → `CustomerCodeExists`（details 帶 code）；已刪除 → `CustomerDeleted`。

- [ ] **Step 5: 測試（表驅動：RPC → 期望碼）**

```go
package services

import (
	"context"
	"strconv"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// TestFirstBatchCodesAreReturned 逐條斷言首批碼真的被回傳（不是只註冊了沒用）。
// PLAT-5002（功能未含方案）與 PLAT-3001（訂閱不可用）由 Plan B Task 4 的
// entitlements 單元測試覆蓋（該處直接呼叫 CheckLimit／Allows 斷言碼）。
func TestFirstBatchCodesAreReturned(t *testing.T) {
	cases := []struct {
		name string
		want string
		call func(t *testing.T) error
	}{
		{"未登入", "AUTH-4001", func(t *testing.T) error {
			client, _ := newTestServer(t) // 空身分
			_, err := client.ListCompanies(context.Background(),
				connect.NewRequest(&v1.ListCompaniesRequest{}))
			return err
		}},
		{"角色不足", "SYS-4001", func(t *testing.T) error {
			staff := authz.Identity{UserID: "9", CompanyID: "1", Role: "staff",
				Roles: []string{"staff"}}
			client, _, _ := newTestServerWithIdentity(t, staff)
			_, err := client.DeleteCompany(context.Background(),
				connect.NewRequest(&v1.DeleteCompanyRequest{CompanyId: "1"}))
			return err
		}},
		{"跨租戶查詢", "SYS-4002", func(t *testing.T) error {
			admin := authz.Identity{UserID: "1", CompanyID: "1", Role: "company_admin",
				Roles: []string{"company_admin"}}
			client, _, db := newTestServerWithIdentity(t, admin)
			other := db.Company.Create().SetName("他人公司").
				SetIdentifier("ERR-B").SaveX(context.Background())
			// 以身分所屬範圍查他公司的 id：範圍過濾後查不到 → SYS-4002（不洩漏存在性）
			_, err := client.GetCompany(context.Background(),
				connect.NewRequest(&v1.GetCompanyRequest{CompanyId: strconv.Itoa(other.ID)}))
			return err
		}},
		// 「配額超限 → PLAT-5001」案例屬 Task 5b（需 internal/platform/*，目前不存在）。
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call(t)
			if err == nil {
				t.Fatal("應回錯誤")
			}
			got := errcodeCodeOf(t, err)
			if got != tc.want {
				t.Fatalf("錯誤碼 = %s；want %s（訊息: %s）", got, tc.want, err.Error())
			}
		})
	}
}
```

（`newUserTestServerWithEntitlement`／`ptr` 由 Plan B Task 6 定義；`uItoa`／`seedUserCompany`／`newTestServer*` 為既有輔助。）

**配額碼要帶 details**（前端靠 `used`/`limit` 顯示「10/10」）：

```go
// localCounter 為本檔用的固定用量計數器（不依賴其他測試檔的假實作）。
type localCounter struct{ used int }

func (c localCounter) Count(context.Context, int, string) (int, error) { return c.used, nil }

// errorInfoOf 由 connect error 取 ErrorInfo（本檔共用輔助）。
func errorInfoOf(t *testing.T, err error) *commonv1.ErrorInfo {
	t.Helper()
	var ce *connect.Error
	if !errors.As(err, &ce) {
		t.Fatalf("應為 *connect.Error，got %T", err)
	}
	for _, d := range ce.Details() {
		if m, ok := d.(proto.Message); ok {
			if ei, ok := m.(*commonv1.ErrorInfo); ok {
				return ei
			}
		}
	}
	t.Fatal("錯誤未帶 ErrorInfo")
	return nil
}

// TestLimitExceededCarriesDetails：PLAT-5001 必須帶 feature/used/limit 三個參數，
// 否則前端無法顯示「已用 10／上限 5」。直接測 entitlements 服務（比走 RPC 更精確：
// 這裡驗的是碼與參數的契約，RPC 層的傳遞由 Step 6 的整合測試覆蓋）。
func TestLimitExceededCarriesDetails(t *testing.T) {
	f := store.NewFake()
	f.PutFeature(store.Feature{Code: entitlements.LimitSeats, Type: "integer"})
	f.PutPlan("std", []store.Entitlement{{
		FeatureCode: entitlements.LimitSeats, Enabled: true, Limit: ptr(int64(5))}})
	f.PutSubscription(store.Subscription{CompanyID: 42, PlanCode: "std", Status: "active"})
	svc := entitlements.New(f, localCounter{used: 5}, entitlements.NewMemoryCache(), 0)

	info := errorInfoOf(t, svc.CheckLimit(context.Background(), 42, entitlements.LimitSeats, 1))
	if info.GetCode() != "PLAT-5001" {
		t.Fatalf("錯誤碼 = %s；want PLAT-5001", info.GetCode())
	}
	for _, k := range []string{"feature", "used", "limit"} {
		if info.GetDetails()[k] == "" {
			t.Fatalf("details 缺 %s（前端無法顯示用量），got %v", k, info.GetDetails())
		}
	}
	if info.GetDetails()["used"] != "5" || info.GetDetails()["limit"] != "5" {
		t.Fatalf("用量參數錯誤: %v", info.GetDetails())
	}
}
```

（Task 4 的 `errcodeCodeOf` 應改以 `errorInfoOf(t, err).GetCode()` 實作，避免兩個各自解析 error detail 的輔助。）

- [ ] **Step 6: 整合測試（驗證碼真的跨網路到客戶端）**

```go
//go:build integration

package services

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	commonv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationErrorInfoReachesClient：以 connect client 發出必然失敗的請求，
// 斷言 ClientError 內含 ErrorInfo 且 trace_id 非空 —— 這條驗證「碼跨網路到得了客戶端」，
// 而不只是「碼在 server 端組出來了」。
func TestIntegrationErrorInfoReachesClient(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)
	_, db := openPGEntClientFromGoose(t, adminDSN)
	ctx := t.Context()

	clients := newListScanServer(t, db, 1) // 既有樣板：super 身分、公司範圍 1
	_, err := clients.companies.GetCompany(ctx, connect.NewRequest(
		&v1.GetCompanyRequest{CompanyId: "999999"}))
	if err == nil {
		t.Fatal("查不存在的公司應失敗")
	}
	var ce *connect.Error
	if !errors.As(err, &ce) {
		t.Fatalf("應為 *connect.Error，got %T", err)
	}
	var info *commonv1.ErrorInfo
	for _, d := range ce.Details() {
		if m, ok := d.(proto.Message); ok {
			if ei, ok := m.(*commonv1.ErrorInfo); ok {
				info = ei
			}
		}
	}
	if info == nil {
		t.Fatal("回應未帶 ErrorInfo（碼沒有跨網路到客戶端）")
	}
	if info.GetCode() != "SYS-4002" {
		t.Fatalf("錯誤碼應為 SYS-4002，got %q", info.GetCode())
	}
	if info.GetTraceId() == "" {
		t.Fatal("ErrorInfo 應帶 trace_id（客服回報時對 server log 用）")
	}
}
```

Run: `cd backend && task check && task test:integration -- -run TestIntegrationErrorInfoReachesClient -v`
Expected: 全綠

- [ ] **Step 7: Commit**

```bash
git add backend/internal/handlers backend/internal/services
git commit -m "feat(errcode): 首批碼落地（auth／權限／樣板域）＋ toConnectError 映射"
```

---

### Task 5b（延後至 Plan B 落地後）：配額與收款碼落地

**為什麼延後**：`internal/platform/{store,entitlements,billing}` 是 Plan B/C 的產物，本計畫執行時**不存在**（`backend/internal/platform/` 目錄不存在），故 `PLAT-*` 的落地點與測試 helper（`newUserTestServerWithEntitlement`／`NewEntitlementCounter`／`store.NewFake`）都還沒有。**不要**在本計畫中為了這步而先建 platform 骨架（那是 Plan B 的範圍）。

**Plan B 落地後要做**：
1. `internal/platform/entitlements/service.go` 的 `Allows`／`CheckLimit` → `errcode.PlatformFeatureNotInPlan`（`PLAT-5002`，details 帶 `feature`）／`errcode.PlatformLimitExceeded`（`PLAT-5001`，details 帶 `feature`／`used`／`limit`）／`errcode.PlatformSubscriptionInactive`（`PLAT-3001`）。
2. `internal/platform/billing/billing.go` 的金額不符（G4）與期別已付款衝突（G3）→ `errcode.PlatformPaymentConflict`（`PLAT-3002`，details 帶 `reason`）。
3. 測試：entitlements 的單元測試直接斷言上述碼；並在 `internal/services` 的表驅動測試補回「配額超限 → `PLAT-5001`」案例（原稿片段見 Task 5 Step 3 與其下方程式碼）。

**驗收**：`PLAT-*` 四個碼在 Plan B/C 的對應路徑上真的被回傳（有測試斷言），且碼表產生器（T6）把它們列出來。

---

### Task 6: 基線守門、碼表與三端常數產生

**Files:**
- Create: `internal/services/errcode_guard_test.go`、`internal/services/errcode_baseline.txt`
- Create: `internal/errcode/generate.go`
- 產生器讀 `errcode.All()` ＋ **存取子**（`ID()`／`Domain()`／`ConnectCode()`／`Message()`／`IsDeprecated()`）；`All()` 已保證依 ID 遞增排序（T6 的 `git diff --exit-code` 靠它穩定）。
- Generated: `docs/error-codes.md`、`frontend/src/lib/errcode.ts`、`platform-console/src/lib/errcode.ts`、`app/lib/gen/errcode.dart`
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: 產生基線（一次性；以掃描器本身產生，不要用 grep）**

⚠️ **不要用 grep 產生基線**：Step 2 的守門掃描以 `"<檔名>:<函式名>"` 為鍵（**刻意不用行號**——行號會因無關編輯全部失效，讓守門測試每次都被迫更新），而 `grep | sed 's/:[0-9]*:/:/'` 產生的鍵**粒度不同**（不含函式名），兩者對不上就會變成永遠紅／永遠綠。
作法：讓守門測試支援標準的更新旗標，用它產生基線：

```go
// 用法：go test ./internal/services/ -run TestNoUnregisteredErrorConstruction -update-errcode-baseline
var updateBaseline = flag.Bool("update-errcode-baseline", false,
	"重寫 errcode_baseline.txt 為目前掃描結果（僅在有意縮小基線時使用，並在 PR 說明）")
```

```bash
cd backend
go test ./internal/services/ -run TestNoUnregisteredErrorConstruction -update-errcode-baseline
wc -l internal/services/errcode_baseline.txt   # 以**當時的樹**為準（Plan A／T1–T5 之後已與計畫撰寫時不同，勿抄舊數字）
```

（旗標只在檔案不存在或明確指定時可寫；平時執行必須是唯讀比較。）

- [ ] **Step 2: 寫守門測試**

```go
// TestNoUnregisteredErrorConstruction：新增的 connect.NewError( 必須帶註冊碼；
// 基線只能縮小 —— 已消失的基線行也必須刪除，否則基線會腐化成永久豁免清單。
func TestNoUnregisteredErrorConstruction(t *testing.T) {
	occurrences := scanConnectNewError(t) // 以 go/parser 掃描 internal/**（排除 _test.go 與 errcode 套件）
	baseline := readLines(t, "errcode_baseline.txt")

	inBaseline := map[string]bool{}
	for _, b := range baseline {
		inBaseline[b] = true
	}
	var added, stale []string
	for _, occ := range occurrences {
		if !inBaseline[occ] {
			added = append(added, occ)
		}
	}
	for _, b := range baseline {
		if !slices.Contains(occurrences, b) {
			stale = append(stale, b)
		}
	}
	if len(added) > 0 {
		t.Fatalf("新增的錯誤建構未使用註冊碼（請改用 errcode.<Code>.Error(…)）:\n%s",
			strings.Join(added, "\n"))
	}
	if len(stale) > 0 {
		t.Fatalf("基線含已不存在的呼叫點（請刪除以下行，基線只能縮小）:\n%s", strings.Join(stale, "\n"))
	}
}
```

`scanConnectNewError` 以 `go/parser` 走訪 `internal/**/*.go`，對每個 `connect.NewError(` 呼叫回傳 `"<檔名>:<函式名>"`（**用函式名而非行號**：行號會因無關編輯而全數失效，導致守門測試每次都被迫更新）。

- [ ] **Step 3: 產生器（`internal/errcode/generate.go`）**

```go
//go:generate go run ./cmd/gen-errcodes

// Package errcode 的產生器在 cmd/gen-errcodes：讀 All()，輸出
// ① docs/error-codes.md（客服與文件用：域／碼／connect 碼／訊息／參數）
// ② 三端常數（TS × 2、Dart）
// 產生檔的檔頭標明「由 go generate 產生，請勿手改」與來源檔案。
```

（`cmd/gen-errcodes/main.go` 以 `text/template` 產生四個檔案；TS 輸出 `export const ERR_CUSTOMER_CODE_EXISTS = "CUST-2001";` 形式的常數與 `CODE_MESSAGES` 表；Dart 同理。實作時一併建立 `cmd/gen-errcodes`。）

⚠️ **`platform-console/src/lib/errcode.ts` 目前無處可寫**：`platform-console/` 是 Plan C 的產物、**本計畫執行時不存在**。作法：產生器**只在該目錄存在時**才輸出該檔（否則跳過並印一行提示），CI 的 diff 清單亦比照（路徑不存在時 `git diff --exit-code --` 對該路徑是 no-op，但仍請在 CI 步驟註明「Plan C 落地後會自動納入」）。**不要**為了這步先建 `platform-console/` 骨架。

- [ ] **Step 4: CI 同步檢查**

```yaml
      # 產生檔（碼表與三端常數）必須與 registry 同步：跑完 generate 後不得有 diff
      - name: Error codes up to date
        run: |
          go generate ./internal/errcode
          git diff --exit-code -- ../docs/error-codes.md ../frontend/src/lib/errcode.ts \
            ../platform-console/src/lib/errcode.ts ../app/lib/gen/errcode.dart
        working-directory: backend
```

- [ ] **Step 5: 跑測試與產生**

Run: `cd backend && go generate ./internal/errcode && go test ./internal/services/ -run TestNoUnregisteredErrorConstruction -v`
Expected: PASS（基線外無新增；基線無殘留）

- [ ] **Step 6: Commit**

```bash
git add backend/internal/services/errcode_guard_test.go backend/internal/services/errcode_baseline.txt \
  backend/internal/errcode backend/cmd/gen-errcodes \
  docs/error-codes.md frontend/src/lib/errcode.ts platform-console/src/lib/errcode.ts app/lib/gen/errcode.dart \
  .github/workflows/ci.yml
git commit -m "feat(errcode): 基線守門（只減不增）、碼表與三端常數產生、CI 同步檢查"
```

---

### Task 7: 慣例與跨文件對齊

**Files:**
- Modify: `backend/AGENTS.md`、`docs/superpowers/specs/2026-09-20-saas-billing-entitlements-design.md`
- Modify: `docs/superpowers/plans/2026-09-20-platform-entitlements-plan.md`、`-platform-lifecycle-console-plan.md`
- Modify: `docs/superpowers/plans/2026-09-20-implementation-order.md`

- [ ] **Step 1: `backend/AGENTS.md` 新增錯誤碼小節，並精確化跨租戶慣例**

```markdown
## 10. 錯誤碼（2026-09-20 起）

1. **一律使用 `internal/errcode` 的註冊碼**；不得直接 `connect.NewError(…, errors.New("…"))`（既有呼叫點列於 `internal/services/errcode_baseline.txt`，**基線只能縮小**；新增未帶註冊碼者測試會紅）。
2. **碼發佈後不得重用或改義**；廢止只加 `Deprecated: true`。
3. **區段規則**：`1xxx` 驗證／`2xxx` 衝突／`3xxx` 狀態／`4xxx` 權限（含 NotFound）／`5xxx` 配額與訂閱／`9xxx` 系統。`register()` 會驗證，違反即啟動失敗。
4. **5xx 一律 `SYS-9000`**，內部細節只進 log；`trace_id` 由 requestid interceptor 產生，並在回應的 `ErrorInfo.trace_id` 帶出（**不**寫進訊息樣板）。
5. **跨租戶與不存在一律 `SYS-4002`（NotFound）**：不洩漏資源是否存在。授權**檢查**失敗（角色/範圍不足）才是 `SYS-4001`（PermissionDenied）——兩者語意不同，前端處理也不同（「請管理員開權」vs「找不到」）。
6. **配額不足用 `PLAT-5001/5002`**（帶 `feature`／`used`／`limit`）：前端據此導向升級方案，不得以 `PermissionDenied` 表示額度問題。
7. **`Error`／`Wrap` 不收 ctx**：`trace_id` 由邊界（`internal/obs/requestid` 的 interceptor）補進 `ErrorInfo`，呼叫端不必也不該傳 ctx（`internal/errcode` 因此是純葉節點）。
```

（同時修訂既有第 3 節「跨租戶資料存取失敗回 PermissionDenied」為精確表述：**授權檢查失敗** → `SYS-4001`；**因範圍過濾而查不到** → `SYS-4002`。）

- [ ] **Step 2: spec 對齊**

§2.2 接觸面第 2 條補「錯誤一律經錯誤碼錶（`docs/error-codes.md`）」；§4.3 的失敗碼分工改為引用 `PLAT-5001/5002`、`PLAT-3001`、`SYS-4001`、`SYS-4002`；§9 In 加「錯誤碼骨架（Plan D，P0-5）」。

- [ ] **Step 3: Plan B／C 對齊**

- Plan B Task 4 的 `FailedPrecondition` 文字 → `errcode.PlatformLimitExceeded`／`PlatformFeatureNotInPlan`；Task 6 的守衛測試斷言**碼**（`PLAT-5001`）而非只有 connect 碼。
- Plan C Task 4 的金額不符／重複入帳 → `errcode.PlatformPaymentConflict`；Task 9 的寫入 RPC 錯誤表補上對應碼。

- [ ] **Step 4: 排序文件補 P0-5**

在 P0 表格新增：**P0-5 錯誤碼骨架（Plan D T1–T6）**，並註明與 P0-1 並行、不重疊檔案；`P1` 的說明補一句「配額與收款錯誤已使用 `PLAT-*` 碼」。

- [ ] **Step 5: 跑全部守門**

Run: `cd backend && task check && go test ./internal/services/ -run TestNoUnregistered -v`
Expected: 全綠

- [ ] **Step 6: Commit**

```bash
git add backend/AGENTS.md docs/superpowers/specs docs/superpowers/plans
git commit -m "docs: 錯誤碼慣例（含跨租戶 NotFound）、spec 與 Plan B/C 對齊、排序補 P0-5"
```

---

## 驗收對照（設計決策）

| 決策 | 對應任務 |
|---|---|
| 碼形態 `域-4位數`、區段與 connect 碼一致 | Task 2 |
| 常數即註冊、啟動即驗證（重複/格式/區段 → panic） | Task 2 |
| 回應格式 `ErrorInfo{code,message,details,trace_id}` | Task 1、Task 2、Task 3 |
| `trace_id` 貫穿 ctx／log／回應 | Task 3 |
| 集中映射（ent → 碼、5xx → `SYS-9000`） | Task 4 |
| 首批碼覆蓋 auth／權限／配額／樣板域 | Task 5 |
| 233 呼叫點（98 行基線）既有錯誤「只減不增」基線守門 | Task 6 |
| 碼表與三端常數自動產生、CI 同步 | Task 6 |
| 跨租戶一律 `NotFound`（不洩漏存在性） | Task 2（`SYS-4002`）、Task 5、Task 7 |
| 配額錯誤可與權限錯誤區分（前端導向不同） | Task 5、Task 7 |

## 風險與對策

| # | 風險 | 對策 |
|---|---|---|
| E1 | 語意（碼）與傳輸碼（connect code）漂移 | 區段規則於 `register()` 硬驗證；測試斷言每個碼的 connect 碼符合區段 |
| E2 | 233 呼叫點的基線腐化成「永久豁免清單」 | 守門測試**同時**斷言「基線不得含已消失的呼叫點」→ 基線只能縮小 |
| E3 | 守門以原始碼掃描實作（非行為斷言） | 是不得已的取捨（型別安全需改 233 處）；以「用函式名＋筆數而非行號」降低偽陽性，並在 AGENTS.md 明寫 |
| E4 | `trace_id` 讓 log 量增加 | 每請求一行；必要時改為 `log/slog` 結構化欄位並調整 level |
| E5 | 錯誤訊息洩漏內部細節 | 5xx 一律 `SYS-9000`；`Wrap` 保留 cause 但訊息不含 cause 內容（測試斷言） |
| E6 | 跨租戶改 `NotFound` 後，前端把「無權限」誤判為「不存在」 | `SYS-4002` 的訊息同時含兩種可能（「資源不存在或無權存取」）；需要區分時由服務層主動回 `SYS-4001`（授權檢查失敗） |

---

*建立：2026-09-20（SaaS spec 的第四份實作計畫；排序為 P0-5，與 Plan A 並行）*

// Package cron 為平台排程的**單趟**執行:一趟 = 一次呼叫,不內建迴圈。
//
// 為什麼是單趟(cmd/platform-cron):「什麼時候跑」交給觸發器(k8s CronJob、Taskfile、測試),
// 「跑什麼」只有這一份 —— 若把 ticker 寫進服務,補跑要另外寫一套、時間也只能是現在。
// now 一律由呼叫端給(可注入 → 可測、可補跑);重跑不得產生第二個期別／第二個事件／第二次轉移
// (冪等由 Task 5 的掃描查詢保證),重疊的兩趟則由單飛鎖(RunGuarded)擋掉。
//
// 分工(spec §2.2 規則 3):本套件**不直寫產品域**,也不碰業務表 —— 狀態掃描走 billing
// (platform.subscriptions／events),凍結是 consumer 把事件翻成 services.SetCompanyStatus。
//
// 稽核:排程本身**不寫** platform.audit_logs(platform.audit_logs.operator_id 是 NOT NULL 且 FK 到
// platform.operators,而排程沒有 operator 主體;見 store.SystemActor 的說明)。它的紀錄是
// platform.events(每次轉移一筆)＋ consumer 經 SetCompanyStatus 落的租戶稽核(actor = 系統 actor)。
//
// 錯誤:回的是**給人看的**錯誤(fmt.Errorf ＋ %w,底層已帶 errcode 的錯誤原樣留在鏈上)——
// cmd/platform-cron 把它直接印進 log,而 errcode 的對外訊息只留固定字串、cause 藏在
// ErrorInfo.detail,panic 的 stack 會因此從 log 裡消失。
package cron

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// isBillingDataError 判斷 EnsureNextPeriod 的失敗是否為「資料問題」(缺價目／週期非法)：
// 兩者都落在 PlatformSubscriptionInactive → 對外 connect 碼 FailedPrecondition。基礎設施問題
// (連線中斷／死鎖)是 errcode.SysInternal → Internal，**不得**計入 Unbilled —— 否則一次連線
// 抖動會讓 Unbilled誤報(而真正的基礎設施錯誤已經在 periodErr 裡)。
// 用 connect 碼而非字串比對：同一個碼家族代表同一類可行動性(見 PLAT-3002 的字串比對教訓，
// console 曾依 details.reason 的中文關鍵詞分流 —— 脆弱)。
func isBillingDataError(err error) bool {
	return connect.CodeOf(err) == connect.CodeFailedPrecondition
}

// Billing 為排程驅動的帳務掃描(實作:billing.Billing)。宣告成介面(C-08)是為了讓 RunOnce
// 能用假 deps 測 —— 掃描語意屬 Task 5,編排語意屬本套件,兩者不該互相綁進容器。
type Billing interface {
	// ExpireTrials 把「試用已到期」的 trialing 訂閱轉 past_due 並設寬限期,回實際筆數。
	// 轉過去之後由 SuspendOverdue／RecordPayment 接手(同一張狀態機,不另立第二套轉移)。
	ExpireTrials(ctx context.Context, now time.Time, graceDays int) (int, error)
	// MarkPastDue 把「期末已過且當期仍 open」的 active 訂閱轉 past_due 並設寬限期,回實際筆數。
	MarkPastDue(ctx context.Context, now time.Time, graceDays int) (int, error)
	// SuspendOverdue 把寬限期已過的 past_due 訂閱轉 suspended 並發 subscription.suspended,回筆數。
	SuspendOverdue(ctx context.Context, now time.Time) (int, error)
	// ExpireCancelled 對「已取消且期末已過」發 subscription.expired(G7;不改訂閱狀態),回筆數。
	ExpireCancelled(ctx context.Context, now time.Time) (int, error)
	// EnsureNextPeriod 在提前窗內為服務中的訂閱開下一期;已開過或未進窗回 false。
	EnsureNextPeriod(ctx context.Context, companyID int, now time.Time, leadDays int) (bool, error)
}

// Dispatcher 為 outbox 派送(實作:consumer.Consumer)。
//
// 回傳值是**本趟認領的事件數**,不是「副作用發生次數」:沒有產品域動作的型別(period.opened…)
// 只被認領,被別的執行搶先認領的也不算。摘要因此寫「認領事件 N 筆」。
type Dispatcher interface {
	DispatchOnce(ctx context.Context, limit int) (int, error)
}

// Store 為排程需要的平台讀取(實作:store/postgres 的 admin 連線)。
//
// 刻意只列用得到的三個方法:排程只需要營運參數、服務中的訂閱、待收款期別,
// 不是整個 BillingStore(介面越窄,假 deps 越小、越不容易與真 store 漂移)。
type Store interface {
	// Setting 讀 platform.settings 的營運參數(grace_days／lead_days)。
	Setting(ctx context.Context, key string) (string, error)
	// ActiveOrTrialingSubscriptions 回仍在服務中的訂閱(逐租戶產生下一期)。
	ActiveOrTrialingSubscriptions(ctx context.Context) ([]store.Subscription, error)
	// OverdueReceivablePeriods 回待收款期別：open 且期末已過，且**排除 G5 平台自營公司**
	// (console 的 ListReceivables 用 `identifier <> 'platform'` 排除，兩處的「租戶」
	// 定義必須一致)。now 由呼叫端給（補跑時的「過期」跟著呼叫端走，不跟 DB 時鐘）。
	OverdueReceivablePeriods(ctx context.Context, now time.Time) ([]store.Period, error)
	// TrialingSubscriptionsWithoutTrialEnd 回 trialing 但沒有到期日的訂閱（未結項 #40）：
	// ExpireTrials 刻意不碰它們，只計數進摘要 StuckTrialing，不做任何轉移。
	TrialingSubscriptionsWithoutTrialEnd(ctx context.Context) ([]store.Subscription, error)
}

// Locker 為單飛鎖:同一時間只允許一趟排程處理(k8s CronJob 與手動補跑會重疊)。
// 生產實作:AdvisoryLocker(PostgreSQL advisory lock)。
type Locker interface {
	// TryLock 嘗試取得鎖。取不到回 (nil, false, nil):別的執行正在跑,**不是錯誤**。
	// 成功時回 unlock(nil 不該發生);unlock 必須被呼叫,否則鎖會留到行程結束。
	TryLock(ctx context.Context) (unlock func(context.Context) error, ok bool, err error)
}

// Deps 為排程的外部依賴(C-08:欄位一律是介面,RunOnce 因此不需要容器就能測)。
type Deps struct {
	Billing  Billing
	Consumer Dispatcher
	Store    Store
	// Lock 只在 RunGuarded 用到(RunOnce 本身是純編排,不含鎖)。
	Lock Locker
}

// Params 為一趟排程的營運參數。全部由 platform.settings 來(LoadParams),不硬編在程式碼裡:
// 寬限與提前天數會隨客戶與季節調整,寫死等於每次調整都要重新部署。
type Params struct {
	GraceDays  int
	LeadDays   int
	EventBatch int
}

// EventBatch 為每趟派送的事件上限。它不是營運參數(沒有「這個客戶要派送幾筆」這種事),
// 故留在程式碼:目的是讓單趟有界,不讓積壓的事件把一趟拉成無限期。
const EventBatch = 200

// Settings 為營運參數的來源(store.BillingStore 的 Setting 子集)。
type Settings interface {
	Setting(ctx context.Context, key string) (string, error)
}

// LoadParams 由 platform.settings 讀取排程參數(key:grace_days／lead_days)。
//
// 缺席或值不合理**一律回錯誤**,不得默默用預設值:那會讓「忘記 seed」變成無聲的錯誤寬限期
// (帳務參數的預設值不能是猜的)。修復方式是把 seed 跑起來(它會補寫這兩個 key)。
func LoadParams(ctx context.Context, st Settings) (Params, error) {
	parse := func(key string) (int, error) {
		raw, err := st.Setting(ctx, key)
		if err != nil {
			return 0, fmt.Errorf("缺少設定 %s: %w", key, err)
		}
		n, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			return 0, fmt.Errorf("設定 %s 不是整數: %q", key, raw)
		}
		if n < 0 || n > 365 {
			return 0, fmt.Errorf("設定 %s 的值不合理(應為 0..365 天): %d", key, n)
		}
		return n, nil
	}
	var p Params
	var err error
	if p.GraceDays, err = parse("grace_days"); err != nil {
		return Params{}, err
	}
	if p.LeadDays, err = parse("lead_days"); err != nil {
		return Params{}, err
	}
	p.EventBatch = EventBatch
	return p, nil
}

// Summary 為一趟排程的結果(C-20／C-21:排程只有這一種摘要,log 行與 JSON 都是它)。
//
// 兩個欄位語意要分清楚:
//   - 前五個是「**本趟**做了什麼」(轉移／開期／認領的**筆數**);
//   - Receivables 是「**還欠多少**」(已過期未付的 open 期別數):它不是本趟的動作,
//     故重跑時不會歸零 —— 一律拿它當「這一趟處理了幾筆待收款」就是誤讀。
type Summary struct {
	// Locked 為本趟是否取得單飛鎖(RunGuarded 填;直接呼叫 RunOnce 時恆為 false)。false 代表
	// 另一個執行正在跑、本趟未處理任何事(不是錯誤),其餘計數必為 0 —— 有了它,log 才分得出
	// 「跳過」與「跑了但沒事可做」。
	Locked           bool `json:"locked"`
	TrialsExpired    int  `json:"trials_expired"`
	PastDue          int  `json:"past_due"`
	Suspended        int  `json:"suspended"`
	ExpiredCancelled int  `json:"expired_cancelled"`
	PeriodsOpened    int  `json:"periods_opened"`
	Dispatched       int  `json:"dispatched"`
	Receivables      int  `json:"receivables"`
	// Unbilled 為「服務中卻沒有 open 期別」的租戶數(未結項 #14):最新一期已 paid／void 且
	// 下一期開不出來(價目缺失／週期非法)的訂閱，會永遠停在 active —— 既不被催收(只掃 open)、
	// 也不再被開帳(逐租戶失敗只記進 log)。這個數字是 operator 唯一能從摘要看見它的地方。
	Unbilled int `json:"unbilled"`
	// StuckTrialing 為「試用中但沒有到期日」的訂閱數(未結項 #40):ExpireTrials 的謂詞
	// 刻意只認 `trial_ends_at IS NOT NULL`（不讓排程猜），故這種列永遠停在 trialing
	// （可用、不催收、不凍結）。寫入路徑已擋（開通要求未來日期），只剩繞過開通的手工列
	// 與歷史殘留 —— 這個數字是 operator 唯一能從摘要看見它的地方。
	StuckTrialing int `json:"stuck_trialing"`
}

// RunOnce 執行一趟完整排程:**試用到期 → 逾期 → 凍結(停用欠費、取消到期) → 產生期別 → 派送事件**。
//
// 順序有依賴,不能重排:
//   - 試用到期放**最前面**:它是生命週期最早的階段(還在試用的租戶不該被當成逾期),而且轉成
//     past_due 之後就不在「服務中」→ 本趟的產生期別不會替一個試用已到期的租戶再開一期;
//   - 先轉移狀態再產生期別:剛被停用的租戶不該被開新期(而且停用是**在 DB 裡**先發生,
//     掃描本身也排除非 active/trialing,兩層一致);
//   - 派送放最後:本趟產生的每一個事件(含凍結)都在同一趟內被認領,不留「事件寫了但沒送」的窗口。
//
// **掃描失敗不得讓派送被跳過**(派出的事件到不了產品域,狀態會長期與帳務不一致)。故失敗分兩類:
//   - 四段帳務掃描(試用到期／逾期／停用／取消到期)失敗即中止:它們成敗相連(同一個狀態機、同一條平台寫入
//     路徑),失敗代表這條路徑本身壞了(連線／權限),硬走後面的步驟只會多幾個一樣的錯誤;
//   - **逐租戶**的產生期別失敗則記下錯誤、跑完其餘租戶,**照樣派送**:那是單一租戶的資料問題
//     (缺當期生效價目、billing_cycle 壞掉),一個租戶的髒資料不該讓**所有**租戶的
//     subscription.suspended／expired 永遠到不了產品域 —— 而 console 那頭早已顯示 suspended。
//
// 中止或部分完成都回**已完成的計數**:每個掃描各自是一個交易,已完成的轉移不會被撤銷,
// 故部分完成是已落地的帳務事實,摘要必須如實回報。now 必為呼叫端提供的時間(不得在此讀時鐘,
// 否則補跑與測試都不可控)。
//
// 錯誤刻意**不經 internal/errcode**:本行程是 CLI,錯誤只進 log 不跨網路,errcode 的穩定對外碼
// 在此沒有價值(且 errcode 的對外訊息會把 cause 藏進 detail,panic 的 stack 會從 log 裡消失)。
// 底層已帶碼的錯誤原樣留在 %w 鏈上,基線未加寬(已獲 controller 核准,T14 補文件)。
func RunOnce(ctx context.Context, deps Deps, now time.Time, p Params) (Summary, error) {
	return runOnceGuarded(ctx, deps, now, p)
}

// runOnceGuarded 是 RunOnce 的本體,唯一的差別是它自己 recover:panic 發生在**這個框**裡時,
// 具名回傳的 s 已經帶著前面步驟完成的計數(若把 recover 留在 RunGuarded,那裡只看得到零值)。
func runOnceGuarded(ctx context.Context, deps Deps, now time.Time, p Params) (s Summary, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("排程 panic: %v\n%s", r, debug.Stack())
		}
	}()

	if s.TrialsExpired, err = deps.Billing.ExpireTrials(ctx, now, p.GraceDays); err != nil {
		return s, fmt.Errorf("試用到期: %w", err)
	}
	// 未結項 #40:trialing 但沒有到期日的訂閱永遠停在試用（ExpireTrials 刻意不碰）。
	// 只計數、不轉移 —— 這個數字是 operator 唯一能從摘要看見它的地方。
	// 失敗不中止整趟（可觀測性查詢壞了不該擋住帳務掃描；錯誤仍往外傳，由呼叫端決定）。
	if stuck, serr := deps.Store.TrialingSubscriptionsWithoutTrialEnd(ctx); serr != nil {
		return s, fmt.Errorf("列出無到期日試用: %w", serr)
	} else {
		s.StuckTrialing = len(stuck)
	}
	if s.PastDue, err = deps.Billing.MarkPastDue(ctx, now, p.GraceDays); err != nil {
		return s, fmt.Errorf("標記逾期: %w", err)
	}
	if s.Suspended, err = deps.Billing.SuspendOverdue(ctx, now); err != nil {
		return s, fmt.Errorf("停用欠費: %w", err)
	}
	// G7:已取消且期末已過 → 發 subscription.expired(consumer 據此凍結公司)。
	if s.ExpiredCancelled, err = deps.Billing.ExpireCancelled(ctx, now); err != nil {
		return s, fmt.Errorf("取消到期: %w", err)
	}

	subs, err := deps.Store.ActiveOrTrialingSubscriptions(ctx)
	if err != nil {
		return s, fmt.Errorf("列出服務中的訂閱: %w", err)
	}
	var periodErr error
	for _, sub := range subs {
		created, err := deps.Billing.EnsureNextPeriod(ctx, sub.CompanyID, now, p.LeadDays)
		if err != nil {
			// 記下**第一個**錯誤就好(其餘同型錯誤只會讓 log 膨脹),但不中斷迴圈:
			// 每個租戶各自一個交易,其他租戶的期別照開。
			if periodErr == nil {
				periodErr = fmt.Errorf("產生期別(company=%d): %w", sub.CompanyID, err)
			}
			// 未結項 #14:服務中卻沒有 open 期別 → 摘要必須看得見。下一期開不出來的兩種
			// 「資料問題」(缺價目／週期非法)都落在 errcode.PlatformSubscriptionInactive，
			// 而「基礎設施問題」(連線中斷／死鎖)是 errcode.SysInternal —— 只數前者，
			// 否則一次連線抖動會讓 Unbilled 誤報(而真正的基礎設施錯誤已經在 periodErr 裡)。
			if isBillingDataError(err) {
				s.Unbilled++
			}
			continue
		}
		if created {
			s.PeriodsOpened++
		}
	}

	if s.Dispatched, err = deps.Consumer.DispatchOnce(ctx, p.EventBatch); err != nil {
		dispatchErr := fmt.Errorf("派送事件: %w", err)
		if periodErr != nil {
			return s, errors.Join(periodErr, dispatchErr)
		}
		return s, dispatchErr
	}

	// 待收款清單(spec §5.4):已過期未付的 open 期別，且排除 G5 平台自營公司
	// (console 的 ListReceivables 用同一謂詞，兩處的「租戶」定義必須一致)。
	// 過期判定用呼叫端的 now（補跑跟著呼叫端走，不跟 DB 時鐘）。
	overdue, err := deps.Store.OverdueReceivablePeriods(ctx, now)
	if err != nil {
		return s, fmt.Errorf("列出待收款期別: %w", err)
	}
	s.Receivables = len(overdue)
	return s, periodErr
}

// RunGuarded 為排程的**唯一入口** = 單飛鎖 + RunOnce + panic 復原。
//
// 取不到鎖 → 回零值摘要、nil 錯誤(別的執行正在跑,本趟什麼都不做;把它當錯誤會讓 k8s 把
// 正常的重疊記成失敗)。panic → 收斂成錯誤(含 stack):排程是無人看管的行程,panic 裸奔之後
// 症狀只有 CrashLoopBackOff,看起來像部署問題而不是「這一趟沒做成」。
func RunGuarded(ctx context.Context, deps Deps, now time.Time, p Params) (s Summary, err error) {
	// recover 註冊在**取鎖之前**:TryLock 本身也可能 panic(例:typed nil 的 Locker),那時還沒動
	// 資料也沒持鎖,但「排程不得裸奔」的立意要一致。RunOnce 自己另有一層 recover(它才看得到
	// 那個框裡的已完成計數),這一層負責取鎖與解鎖這一段。
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("排程 panic: %v\n%s", r, debug.Stack())
		}
	}()
	if deps.Lock == nil {
		// 少了鎖就沒有單飛:寧可這一趟不跑,也不要兩個執行一起改帳。
		return Summary{}, errors.New("排程缺少單飛鎖(未設定時不得執行:兩個執行會一起改帳)")
	}
	unlock, ok, err := deps.Lock.TryLock(ctx)
	if err != nil {
		return Summary{}, fmt.Errorf("取得單飛鎖: %w", err)
	}
	if !ok {
		return Summary{}, nil
	}
	defer func() {
		// 用 WithoutCancel 解鎖:ctx 逾時／被取消不該讓鎖留在連線上,否則下一趟全被擋住。
		// 解鎖失敗在 log 之外還會**丟棄那條連線**(見 AdvisoryLocker.TryLock),故這裡只記錄。
		if uerr := unlock(context.WithoutCancel(ctx)); uerr != nil {
			log.Printf("platform-cron: 釋放單飛鎖失敗(已丟棄該連線,連線關閉即釋放鎖): %v", uerr)
		}
	}()

	s, err = RunOnce(ctx, deps, now, p)
	// 走得到這裡就代表鎖在手上(包括 RunOnce 回錯誤的部分完成)。
	s.Locked = true
	return s, err
}

// LockKey 為單飛鎖的 key(pg_try_advisory_lock 的 bigint):常數即識別,不另建表。
// 值取自 ASCII "PLATCRON"(0x504C415443524F4E),避免與其他用途的 advisory lock 撞號。
const LockKey int64 = 0x504C415443524F4E

// AdvisoryLocker 以 PostgreSQL advisory lock 做單飛(key 見 LockKey)。
//
// 為什麼用 advisory lock 而不是「在 platform.settings 記一筆執行中」:單飛要的是**排他性與
// 自動釋放**,不是資料。多一張表就多一份要清理的狀態(而且清理本身也要單飛),行程被 kill 還會
// 留下永遠「執行中」的殘骸;advisory lock 在連線結束時由資料庫自動釋放。
//
// 鎖是 **session 級**、綁在取得它的那一條連線上,故取鎖用一條專屬連線(*sql.Conn,從池中固定
// 一條),解鎖也在**同一條**連線上 —— 用池裡隨機的連線解鎖會解到別的連線,鎖就洩漏到行程結束。
// 解鎖失敗(或 `pg_advisory_unlock` 回報未持有)時那條連線會被**丟棄**(不還池):連線還活著就
// 可能仍持有鎖,還回池裡會讓同一行程的下一趟擋住自己(從外面看起來就是排程靜默停擺)。
type AdvisoryLocker struct {
	db  *sql.DB
	key int64
}

// NewAdvisoryLocker 建立 advisory lock 單飛鎖;db 為平台(admin)連線,key 一般用 LockKey。
func NewAdvisoryLocker(db *sql.DB, key int64) *AdvisoryLocker {
	return &AdvisoryLocker{db: db, key: key}
}

// TryLock 見 Locker.TryLock。
func (l *AdvisoryLocker) TryLock(ctx context.Context) (func(context.Context) error, bool, error) {
	conn, err := l.db.Conn(ctx)
	if err != nil {
		return nil, false, err
	}
	var got bool
	// pg_try_advisory_lock 不等待:取不到就回 false(排程不該為了鎖排隊,下一趟再來)。
	if err := conn.QueryRowContext(ctx, `SELECT pg_try_advisory_lock($1)`, l.key).Scan(&got); err != nil {
		_ = conn.Close()
		return nil, false, err
	}
	if !got {
		_ = conn.Close()
		return nil, false, nil
	}
	return func(ctx context.Context) error {
		var unlocked bool
		err := conn.QueryRowContext(ctx, `SELECT pg_advisory_unlock($1)`, l.key).Scan(&unlocked)
		if err == nil && !unlocked {
			// 回 false 表示這條 session 根本沒持有鎖(正常路徑不該發生)。
			err = errors.New("pg_advisory_unlock 回報未持有鎖")
		}
		if err != nil {
			// 解鎖沒有確認成功 → **丟棄**這條連線(不要還池):連線還活著就可能仍持有 session 級鎖,
			// 還回池裡會讓同一行程的下一趟擋住自己。*sql.Conn.Close 只是**還池**,故在 Raw 內直接
			// 關掉底層 driver 連線(driver.Conn.Close;連線關閉 = PostgreSQL 自動釋放該 session 的鎖),
			// 再回 driver.ErrBadConn,讓 database/sql 把這個池位標成壞的、不再重用。
			_ = conn.Raw(func(any) error { return driver.ErrBadConn })
			return err
		}
		// 解鎖已確認:*sql.Conn.Close 是**還回池**(不是關掉 TCP 連線)——這裡安全,因為鎖已經
		// 明確放掉;還池讓下一趟不必重新建連線。
		return conn.Close()
	}, true, nil
}

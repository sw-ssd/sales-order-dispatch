// Package consumer 派送 platform.events 的跨域副作用:把平台域的事件變成產品域的公司狀態變更。
//
// 為什麼需要它(spec §2.2 規則 3):平台域**不直寫產品域**。排程只寫事件(platform.events),
// 由本套件把事件翻成 services.SetCompanyStatus 的呼叫 —— 平台域因此沒有第二條改 companies 的路,
// 「停用連鎖」的語意與稽核格式也只有 SetCompanyStatus 那一份。
//
// 交易形狀(**同一條連線的同一個交易**,缺一不可):
//
//  1. 取待派送事件(EventStore.UndispatchedEvents,admin 唯讀);
//  2. 每筆開一個系統範圍(scope=all)的 ent 交易(SystemTx.Run);
//  3. 交易內**先條件式認領**:UPDATE platform.events SET dispatched_at = now()
//     WHERE id = $1 AND dispatched_at IS NULL —— 0 列代表別的執行已處理,直接跳過(不做事、不算失敗);
//  4. 同一交易內呼叫產品域的唯一入口 services.SetCompanyStatus;
//  5. commit。
//
// 為什麼兩者真的同一個 commit:platform schema 與業務表在同一個 PostgreSQL 資料庫,而
// SystemTx 的認領與 ent 寫入都走**同一條連線的那一個交易**(見 DBSystemTx)。於是
// 「認領成功但狀態寫入失敗」整筆回滾(下趟重試),不會出現「公司沒被凍結、事件卻已派送」的孤兒。
//
// 冪等(排程可重跑、可補跑):認領是條件式 UPDATE(一筆事件只會被認領一次),而 SetCompanyStatus
// 對同值為 no-op(不留稽核)—— 兩層加起來使重跑不產生第二個副作用。
package consumer

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"connectrpc.com/connect"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
	"github.com/salesorder/sales-order-1.0/backend/internal/services"
)

// eventAction 為事件型別在產品域的動作。
type eventAction struct {
	status company.Status
	reason string
}

// actions 為會改動產品域的事件映射表。
//
// subscription.expired **必須在表上**(G7,spec §5.6):訂閱已取消且期末已過 → 凍結公司。
// 漏了它,consumer 會把它當「未對應型別」只認領不動作 —— 整條 G7(已取消租戶期末停用)靜默失效,
// 而排程、事件、狀態看起來全部正常。
//
// 未列表的型別(subscription.past_due、period.opened、period.payment_recorded…)不動公司狀態:
// 寬限期內仍提供服務(凍結由 subscription.suspended 負責),期別事件目前只供通知。它們走未對應分支:
// 記一行 log 後認領 —— 不認領的話排程每趟都會重掃同一筆(無限循環)。
var actions = map[string]eventAction{
	"subscription.suspended":   {company.StatusSuspended, "訂閱逾期未付（排程凍結）"},
	"subscription.expired":     {company.StatusSuspended, "訂閱已取消且期末已過（排程凍結）"},
	"subscription.reactivated": {company.StatusActive, "訂閱補款復原（排程解除凍結）"},
}

// EventStore 為 consumer 需要的平台讀取(實作:platform/store/postgres 的 admin 連線)。
type EventStore interface {
	// UndispatchedEvents 取未派送事件(依 id 排序,先寫先派送)。
	UndispatchedEvents(ctx context.Context, limit int) ([]store.Event, error)
	// SystemActor 讀 platform.settings.system_actor_user_id:稽核主體是**租戶 users.id**
	// (audit_logs.user_id 的 FK 指向它),不是平台 operator。
	SystemActor(ctx context.Context) (int64, error)
}

// SystemTx 在系統範圍(scope=all)的 ent 交易內執行 fn。生產實作:NewDBSystemTx;
// 單元測試以假實作把「條件式認領」與「fn 失敗即回滾」建模出來。
type SystemTx interface {
	// Run 開一個系統範圍交易執行 fn:ctx 已注入該交易(產品域入口由 ctx 取交易),
	// tx 提供認領與交易內的 client。fn 回錯誤即整筆回滾,否則 commit。
	Run(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error
}

// Tx 為一筆事件在交易內可用的能力(兩者都在同一條連線的同一個交易)。
type Tx interface {
	// Claim 條件式認領事件:false = 別的執行已認領(或已處理),呼叫端應跳過且不算失敗。
	Claim(ctx context.Context, eventID int64) (bool, error)
	// Client 回傳交易內的 ent client(產品域寫入一律走它;不得用交易外的 fallback)。
	Client() *ent.Client
}

// CompanyStatusSetter 為產品域**唯一**公司狀態入口的窄介面。
type CompanyStatusSetter interface {
	// SetStatus 在 ctx 的交易內變更公司狀態(client 為同一交易的 ent client)。
	SetStatus(ctx context.Context, client *ent.Client, companyID int,
		status company.Status, reason string, actor authz.Identity) error
}

// ProductDomain 把 CompanyStatusSetter 接到產品域的唯一入口 services.SetCompanyStatus。
// 不得自己 UPDATE companies:停用連鎖的語意與稽核格式只有 SetCompanyStatus 那一份。
type ProductDomain struct{}

func (ProductDomain) SetStatus(ctx context.Context, client *ent.Client, companyID int,
	status company.Status, reason string, actor authz.Identity) error {
	return services.SetCompanyStatus(ctx, client, companyID, status, reason, actor)
}

// Consumer 派送懸而未決的事件(排程每趟呼叫 DispatchOnce)。
type Consumer struct {
	events EventStore
	sysTx  SystemTx
	setter CompanyStatusSetter
}

// New 建立 consumer:events 為平台讀取,setter 為產品域入口(生產:ProductDomain{},
// 單元測試:假實作),sysTx 提供「認領 ＋ 狀態變更」所在的系統範圍交易。
func New(events EventStore, sysTx SystemTx, setter CompanyStatusSetter) *Consumer {
	return &Consumer{events: events, sysTx: sysTx, setter: setter}
}

// DispatchOnce 依序派送未處理事件,回傳**本趟認領**的事件數。
//
// 未對應型別也算已派送(它已被認領,不再重掃);被其他執行搶先認領者不計。單筆失敗即停
// (回已完成數與錯誤):事件依 id 排序、先寫先派送,壞掉的那筆留待下趟重試,不吞掉錯誤,
// 也不讓後續事件被靜默跳過。
func (c *Consumer) DispatchOnce(ctx context.Context, limit int) (int, error) {
	events, err := c.events.UndispatchedEvents(ctx, limit)
	if err != nil {
		return 0, errcode.SysInternal.Wrap(err)
	}
	// actor 只在**真的要寫產品域**時解析:未設定的 platform.settings 不該讓「整批都是未對應型別」
	// 的一趟失敗(那會讓那些事件每趟被重掃)。零值代表還沒解析過。
	var actor authz.Identity
	done := 0
	for _, ev := range events {
		act, mapped := actions[ev.EventType]
		var companyID int
		if mapped {
			if companyID, err = companyIDOf(ev); err != nil {
				return done, err
			}
			if actor.UserID == "" {
				if actor, err = c.systemActor(ctx); err != nil {
					return done, err
				}
			}
		} else {
			log.Printf("platform consumer: 事件 %d 型別 %q 沒有產品域動作(僅認領)", ev.ID, ev.EventType)
		}
		claimed, err := c.dispatch(ctx, ev, act, mapped, companyID, actor)
		if err != nil {
			return done, err
		}
		if claimed {
			done++
		}
	}
	return done, nil
}

// dispatch 在同一交易內認領事件,並(對應型別時)變更公司狀態。回傳本筆是否由這趟認領。
func (c *Consumer) dispatch(ctx context.Context, ev store.Event, act eventAction,
	mapped bool, companyID int, actor authz.Identity) (bool, error) {
	claimed := false
	err := c.sysTx.Run(ctx, func(ctx context.Context, tx Tx) error {
		ok, err := tx.Claim(ctx, ev.ID)
		claimed = ok
		if err != nil || !ok {
			// 沒認領到就是別的執行處理了:不做事、不算失敗(回去 commit 空的交易)。
			return err
		}
		if !mapped {
			return nil
		}
		return c.setter.SetStatus(ctx, tx.Client(), companyID, act.status, act.reason, actor)
	})
	return claimed, err
}

// systemActor 組出稽核主體。Role 只是標記(seed 的系統 actor 是 super);稽核只取 UserID,
// 而 UserID 必須是**真實存在的租戶 users.id**(audit_logs.user_id 是 FK)。
func (c *Consumer) systemActor(ctx context.Context) (authz.Identity, error) {
	id, err := c.events.SystemActor(ctx)
	if err != nil {
		return authz.Identity{}, errcode.SysInternal.Wrap(err)
	}
	return authz.Identity{
		UserID: strconv.FormatInt(id, 10),
		Role:   "super",
		Roles:  []string{"super"},
	}, nil
}

// companyIDOf 取事件 payload 的 company_id。Task 4／5 保證一定有(payload 帶足以識別的欄位),
// 故 consumer **不回查 DB 推**(推出來的可能是另一家公司)。
func companyIDOf(ev store.Event) (int, error) {
	var payload struct {
		CompanyID int `json:"company_id"`
	}
	if err := json.Unmarshal(ev.Payload, &payload); err != nil {
		return 0, errcode.SysInternal.Wrap(
			fmt.Errorf("事件 %d(%s) 的 payload 不是合法 JSON: %w", ev.ID, ev.EventType, err))
	}
	if payload.CompanyID == 0 {
		return 0, errcode.SysInternal.Wrap(
			fmt.Errorf("事件 %d(%s) 的 payload 缺 company_id", ev.ID, ev.EventType))
	}
	return payload.CompanyID, nil
}

// DBSystemTx 為 SystemTx 的生產實作:admin(owner)連線上的系統範圍 ent 交易。
//
// 為什麼是 admin 連線:platform schema 對業務角色零權限(00029),而業務表(companies／audit_logs)
// 已 FORCE RLS —— 這條路徑兩者都要碰,只有 owner 連線做得到,且交易內一定要有
// scope=all(dbtenant.SystemScopeTx 負責;FORCE 讓 owner 也受 policy 約束)。
type DBSystemTx struct{ db *sql.DB }

// NewDBSystemTx 以 admin DSN 的連線建立(與平台的 store 共用同一個 *sql.DB 即可)。
func NewDBSystemTx(db *sql.DB) *DBSystemTx { return &DBSystemTx{db: db} }

// Run 見 SystemTx.Run。每趟自建 ent client 與 driver:見 txTracker 的說明。
func (s *DBSystemTx) Run(ctx context.Context, fn func(context.Context, Tx) error) error {
	track := &txTracker{inner: entsql.OpenDB(dialect.Postgres, s.db)}
	client := ent.NewClient(ent.Driver(dbtenant.Wrap(track)))
	return wrapUncoded(dbtenant.SystemScopeTx(ctx, client, func(tx *ent.Tx) error {
		raw, err := track.current()
		if err != nil {
			return err
		}
		return fn(dbtenant.WithTenantTx(ctx, tx), dbTx{raw: raw, client: tx.Client()})
	}))
}

// dbTx 為 DBSystemTx 開出的交易:認領走 driver 記下的 dialect.Tx,產品域寫入走 ent client,
// 兩者都在同一條連線的同一個交易。
type dbTx struct {
	raw    dialect.Tx
	client *ent.Client
}

func (t dbTx) Client() *ent.Client { return t.client }

// claimSQL 條件式認領:只有「尚未派送」的那一列會被更新,0 列 = 別的執行已處理。
// attempts 與 store.MarkEventDispatchedTx 同義(同一張表的可觀測性計數,含重試)。
const claimSQL = `UPDATE platform.events
   SET dispatched_at = now(), attempts = attempts + 1
 WHERE id = $1 AND dispatched_at IS NULL`

// Claim 見 Tx.Claim。
//
// ent 的 *ent.Tx 沒有 raw SQL 的公開入口(未開 gen 的 ExecQuery feature),故用 driver 記下的
// dialect.Tx:v 傳 *sql.Result 是 ent SQL driver 的約定(同 dbtenant 套 SET LOCAL 的寫法)——
// 也正因為如此,這條 UPDATE 與 ent 的寫入才在同一條連線上。
func (t dbTx) Claim(ctx context.Context, eventID int64) (bool, error) {
	var res sql.Result
	if err := t.raw.Exec(ctx, claimSQL, []any{eventID}, &res); err != nil {
		return false, errcode.SysInternal.Wrap(fmt.Errorf("認領事件 %d: %w", eventID, err))
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, errcode.SysInternal.Wrap(fmt.Errorf("認領事件 %d 的列數: %w", eventID, err))
	}
	return n == 1, nil
}

// txTracker 包住 ent 的 SQL driver,記下它剛開好的交易(見 DBSystemTx.Run)。
//
// 為什麼需要它:認領(platform.events 的條件式 UPDATE)與 services.SetCompanyStatus 必須落在
// 同一條連線的同一個交易才可能有原子性,而 ent 沒有公開的 raw SQL 入口。
// 每趟 Run 各自建立一個 tracker(連同 ent client),故槽位不會被同時進行的另一趟覆蓋。
type txTracker struct {
	inner dialect.Driver
	tx    dialect.Tx
}

// Tx 開交易並記下它;dbtenant 的 RLS 裝飾器隨後會在**同一條**交易上套 scope=all。
func (t *txTracker) Tx(ctx context.Context) (dialect.Tx, error) {
	tx, err := t.inner.Tx(ctx)
	if err != nil {
		return nil, err
	}
	t.tx = tx
	return tx, nil
}

// current 回傳本趟剛開好的交易。SystemScopeTx 先開交易才呼叫 fn,故 fn 內必定拿得到。
func (t *txTracker) current() (dialect.Tx, error) {
	if t.tx == nil {
		return nil, errcode.SysInternal.Wrap(errors.New("consumer: 系統範圍交易未開啟"))
	}
	return t.tx, nil
}

func (t *txTracker) Exec(ctx context.Context, query string, args, v any) error {
	return t.inner.Exec(ctx, query, args, v)
}
func (t *txTracker) Query(ctx context.Context, query string, args, v any) error {
	return t.inner.Query(ctx, query, args, v)
}
func (t *txTracker) Close() error    { return t.inner.Close() }
func (t *txTracker) Dialect() string { return t.inner.Dialect() }

// wrapUncoded 對未帶碼的錯誤補上系統碼;產品域已帶碼的錯誤原樣往外(蓋掉專碼等於丟失語意)。
func wrapUncoded(err error) error {
	if err == nil {
		return nil
	}
	var coded *connect.Error
	if errors.As(err, &coded) {
		return err
	}
	return errcode.SysInternal.Wrap(err)
}

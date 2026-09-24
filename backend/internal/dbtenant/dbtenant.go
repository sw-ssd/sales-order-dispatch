// Package dbtenant 提供「請求層租戶交易」：每個 unary RPC 開一個交易，交易內以
// SET LOCAL app.* 套用 RLS scope，服務層統一由 Client(ctx, s.db) 取得被約束的 client。
//
// 為何不是逐呼叫點包交易：服務層有 124 處直呼查詢與 42 處自開交易，逐點包會漏；
// 交易邊界改由請求擁有（spec §6.3）。串流 RPC 與長時工作（未來 WatchBoard／PDF）
// 不得沿用此法，屆時另立短交易邊界。
//
// 為何用 driver 裝飾器而不是 auth.ApplyRLS(ctx, tx, scope)：本專案產生的 ent 型別
// 沒有 ExecContext，ent 也沒有 OpenTx 可把 *sql.Tx 綁成 client。ent 的 client.Tx(ctx)
// 會呼叫 driver.Tx(ctx)，我們就在那裡把 ctx 的 RLS scope 套進剛開好的交易 ——
// 於是**服務層任何 client.Tx(ctx) 都自動 RLS 安全**（42 處不必逐一改）。
package dbtenant

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"connectrpc.com/connect"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
)

type txCtxKey struct{}

// NewClient 建立業務 ent client：**必須**經此建立，RLS 裝飾器才會生效。
// 其他 ent client（CLI／seed／測試 fixture）走原本的 entsql.OpenDB，不受影響。
func NewClient(db *sql.DB) *ent.Client {
	return ent.NewClient(ent.Driver(Wrap(entsql.OpenDB(dialect.Postgres, db))))
}

// Wrap 以 RLS 裝飾器包住 dialect driver。
func Wrap(inner dialect.Driver) dialect.Driver { return &rlsDriver{inner: inner} }

type rlsDriver struct{ inner dialect.Driver }

func (d *rlsDriver) Exec(ctx context.Context, query string, args, v any) error {
	return d.inner.Exec(ctx, query, args, v)
}

func (d *rlsDriver) Query(ctx context.Context, query string, args, v any) error {
	return d.inner.Query(ctx, query, args, v)
}

func (d *rlsDriver) Close() error    { return d.inner.Close() }
func (d *rlsDriver) Dialect() string { return d.inner.Dialect() }

// Tx 開交易並**立刻**套用 ctx 的 RLS scope：SET LOCAL 只在當前交易有效，
// 而這裡正是交易剛開好、任何業務查詢之前 —— 錯過這個點就再也補不上。
func (d *rlsDriver) Tx(ctx context.Context) (dialect.Tx, error) {
	tx, err := d.inner.Tx(ctx)
	if err != nil {
		return nil, err
	}
	for _, stmt := range auth.RLSStatements(auth.RLSFrom(ctx)) {
		// dialect.Tx.Exec 的 v 參數對 SQL driver 而言是 *sql.Result（見 ent dialect 文件）：
		// entsql.Result 是 sql.Result 的別名（介面），故取變數位址而非複合字面值。
		var res sql.Result
		if err := tx.Exec(ctx, stmt, []any{}, &res); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("dbtenant: 套用 RLS 失敗(%s): %w", stmt, err)
		}
	}
	return tx, nil
}

// WithTenantTx 把請求交易放進 ctx。
func WithTenantTx(ctx context.Context, tx *ent.Tx) context.Context {
	return context.WithValue(ctx, txCtxKey{}, tx)
}

// afterCommitKey 為請求交易內的 post-commit 掛鉤收集器（見 AfterCommit）。
type afterCommitKey struct{}

// postCommitClientKey 承載「提交後掛鉤可安全寫入的 client」（見 PostCommitClient）。
type postCommitClientKey struct{}

// PostCommitClient 回傳提交後掛鉤應該使用的 client：**基礎 client + 當前 ctx 的 RLS scope**
// （裝飾器會在 client.Tx(ctx) 時自動套用 SET LOCAL），而不是請求交易本身的 client ——
// 後者在掛鉤執行時已經提交，對它寫入只會得到 ErrTxDone。
//
// 為何不能沿用請求交易 client：通知的發送排在提交後（FCM 是外部 I/O，不得進交易），
// 發送結果要把 pending 改 sent/failed —— 用已提交的 client 會靜默失敗（呼叫端忽略錯誤），
// 通知就永遠停在 pending：使用者看不到「已發送」，MarkRead 的 sent 分支也永不生效
// （2026-09-24 由 10.11 通知工作以真 PG 迴歸測試抓到）。
//
// 沒有請求交易（CLI／seed／單測直呼服務）→ 退回 fallback，維持原語意。
func PostCommitClient(ctx context.Context, fallback *ent.Client) *ent.Client {
	if c, ok := ctx.Value(postCommitClientKey{}).(*ent.Client); ok {
		return c
	}
	return fallback
}

// pendingAfterCommit 收集「請求交易 commit 成功後才執行」的工作。
type pendingAfterCommit struct{ fns []func(context.Context) error }

// withAfterCommit 把收集器放進 ctx（僅 Interceptor 內部使用）。
func withAfterCommit(ctx context.Context, p *pendingAfterCommit) context.Context {
	return context.WithValue(ctx, afterCommitKey{}, p)
}

// withPostCommitClient 把基礎 client 放進 ctx（僅 Interceptor 內部使用）。
func withPostCommitClient(ctx context.Context, c *ent.Client) context.Context {
	return context.WithValue(ctx, postCommitClientKey{}, c)
}

// AfterCommit 註冊一項「請求交易 commit 成功後才執行」的工作（目前用於 OpenFGA tuple 同步）。
//
// 為何需要這個掛鉤：OpenFGA 與業務 DB 不是同一個交易，同步必須發生在業務交易提交**之後** ——
// 在交易內同步一旦該交易其後被回滾（同函式後續讀取失敗、或 Interceptor 的 Commit 失敗），
// OpenFGA 就停在**被回滾的 DB 狀態**：新增方向 = 多授權（fail-open）、移除方向 = 多收縮，兩者都要
// 等下次 authz.Provision reconcile 才修正；而且業務交易的列鎖會跨越外部寫入。反之，handler 自行
// commit 又違反「交易邊界由請求擁有」與「同一請求不得對同一列開第二條交易」。
// 因此由 Interceptor 在 handler 之前放收集器、**Commit 成功後**依序執行；回滾則整個丟棄不執行。
//
// 沒有請求交易（CLI／單元測試直接呼叫 service）→ 回錯誤：呼叫端的同步沒被安排，fail-closed。
func AfterCommit(ctx context.Context, fn func(context.Context) error) error {
	p, ok := ctx.Value(afterCommitKey{}).(*pendingAfterCommit)
	if !ok {
		return errors.New("dbtenant: 沒有請求交易，無法註冊 post-commit 工作")
	}
	p.fns = append(p.fns, fn)
	return nil
}

// TxFrom 取出請求交易；未注入時回 false（CLI／seed／單元測試）。
func TxFrom(ctx context.Context) (*ent.Tx, bool) {
	tx, ok := ctx.Value(txCtxKey{}).(*ent.Tx)
	return tx, ok
}

// Client 回傳當前請求的 scoped client；沒有請求交易時退回 fallback。
func Client(ctx context.Context, fallback *ent.Client) *ent.Client {
	if tx, ok := TxFrom(ctx); ok {
		return tx.Client()
	}
	return fallback
}

// Interceptor 為 unary RPC 開租戶交易：開交易（RLS 由 Wrap 的 driver 裝飾器在
// Tx(ctx) 內套用，此處只負責交易邊界）→ 呼叫 handler → err == nil 則 commit、否則 rollback。
// 以 interceptor 而非 HTTP middleware 的理由：interceptor 看得到 domain error
// （HTTP 狀態碼在 connect 下與錯誤碼的對應是間接的）。
func Interceptor(client *ent.Client) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			tx, err := client.Tx(ctx)
			if err != nil {
				// 對外只回固定訊息(不洩漏 SQLSTATE／policy 名／SET LOCAL 語句):connect-go 會
				// 逐字轉送 Message() 給客戶端。根因必須落 server log —— chi Logger 不記錄
				// handler error,不落 log 就完全無從追查(同 company_service 既有慣例)。
				log.Printf("dbtenant: 開啟租戶交易失敗: %v", err)
				return nil, connect.NewError(connect.CodeInternal, errors.New("開啟租戶交易失敗"))
			}
			pending := &pendingAfterCommit{}
			resp, err := next(withAfterCommit(WithTenantTx(ctx, tx), pending), req)
			if err != nil {
				if rbErr := tx.Rollback(); rbErr != nil {
					log.Printf("dbtenant: 回滾租戶交易失敗: %v", rbErr)
				}
				// 回滾 → 掛鉤整個丟棄(AFTER 的同步只允許發生在已提交的狀態上)。
				return nil, err
			}
			if err := tx.Commit(); err != nil {
				log.Printf("dbtenant: 提交租戶交易失敗: %v", err)
				return nil, connect.NewError(connect.CodeInternal, errors.New("提交交易失敗"))
			}
			// commit 成功後才執行掛鉤(列鎖已釋放、DB 已定案)。掛鉤失敗即回該錯誤 ——
			// 與「handler 自行 commit 後同步失敗」的舊語意一致(DB 已提交,不回滾)。
			// 掛鉤的 ctx 帶上基礎 client(PostCommitClient),讓掛鉤能另開受 RLS 約束的新交易
			// 落庫其結果 —— 請求交易在此已提交,不可再寫。
			hookCtx := withPostCommitClient(ctx, client)
			for _, fn := range pending.fns {
				if err := fn(hookCtx); err != nil {
					return nil, err
				}
			}
			return resp, nil
		}
	})
}

// HandlerOption 供 NewXServiceHandler 掛載租戶交易。
func HandlerOption(client *ent.Client) connect.HandlerOption {
	return connect.WithInterceptors(Interceptor(client))
}

// SystemScopeTx 在明確的系統範圍（scope=all）內執行 fn：供「尚無身分」的路徑使用
// （登入憑證查詢、authzMiddleware 的身分解析、seed）。刻意獨立成一個入口，
// 讓「系統範圍」在呼叫點顯眼可審計，而不是散落的 SET LOCAL。
func SystemScopeTx(ctx context.Context, client *ent.Client, fn func(*ent.Tx) error) error {
	// scope=all 必須在**開交易之前**注入 ctx：driver 裝飾器在 Tx(ctx) 內讀它。
	ctx = auth.WithRLS(ctx, auth.RLSScope{DataScope: auth.DataScopeAll})
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

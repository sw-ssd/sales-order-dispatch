// 平台排程入口(單趟):config → admin 連線 → 組 deps → RunGuarded → 摘要。
//
// 為什麼極薄:排程的內容(順序、可重跑、單飛、panic 復原)全在 internal/platform/cron,
// 本檔只負責組裝與把結果印出來 —— 同一份組裝不該在 main、Taskfile 與日後的 HTTP 端點各寫一遍。
//
// 不內建迴圈:重複執行交給觸發器(k8s CronJob;本機用 `task backend:platform:cron`)。
// --date 可覆寫「現在」(RFC3339)以便手動補跑或驗證特定日期的那一趟。
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx database/sql driver

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/consumer"
	platformcron "github.com/salesorder/sales-order-1.0/backend/internal/platform/cron"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store/postgres"
	"github.com/salesorder/sales-order-1.0/backend/third_party/database"
)

func main() {
	dateFlag := flag.String("date", "", "覆寫執行時間(RFC3339;預設為現在)")
	flag.Parse()

	now := time.Now().UTC()
	if *dateFlag != "" {
		parsed, err := time.Parse(time.RFC3339, *dateFlag)
		if err != nil {
			log.Fatalf("--date 格式錯誤(需 RFC3339,例 2026-10-01T03:00:00Z): %v", err)
		}
		now = parsed.UTC()
	}

	cfg := config.New()
	// 平台域與 consumer 都走 owner(admin)連線:platform schema 對業務角色 app_rw 零權限(00029/S9),
	// 而凍結要同時寫 platform.events 與業務表(companies／audit_logs,FORCE RLS 靠交易內的 scope=all)。
	db, err := database.OpenSQL(cfg.Database.AdminDSN())
	if err != nil {
		log.Fatalf("admin 連線: %v", err)
	}
	defer func() { _ = db.Close() }()

	st := postgres.New(db)
	deps := platformcron.Deps{
		Billing:  billing.NewBilling(st),
		Consumer: consumer.New(st, consumer.NewDBSystemTx(db), consumer.ProductDomain{}),
		Store:    st,
		Lock:     platformcron.NewAdvisoryLocker(db, platformcron.LockKey),
	}
	// 營運參數一律來自 platform.settings,不用程式碼裡的預設值:缺席即失敗(見 LoadParams)。
	ctx := context.Background()
	params, err := platformcron.LoadParams(ctx, st)
	if err != nil {
		log.Fatalf("讀取排程參數失敗: %v\n"+
			"修復:執行 `task backend:seed`(冪等)寫入預設值(grace_days／lead_days)與系統 actor。", err)
	}

	summary, err := platformcron.RunGuarded(ctx, deps, now, params)
	if err != nil {
		// 唯一以非零離開碼收場的情況:真有沒做成的事(掃描失敗、派送有事件沒派送成功、panic)。
		log.Fatalf("排程執行失敗(now=%s): %v", now.Format(time.RFC3339), err)
	}
	if !summary.Locked {
		// 取不到單飛鎖:另一個執行正在跑。**不是錯誤** —— 以非零離開碼收場會讓 k8s 把正常的重疊
		// 記成失敗(而它其實什麼都沒做,也沒什麼可重試)。
		log.Printf("platform-cron 跳過(now=%s):另一個排程執行中", now.Format(time.RFC3339))
		return
	}
	raw, err := json.Marshal(summary)
	if err != nil {
		log.Fatalf("摘要序列化失敗: %v", err)
	}
	// 同一份摘要兩種呈現:繁中一行給人看、JSON 一行給機器(grep／告警)。
	// 語意要準:dispatched 是「**本趟認領**的事件數」,不是「已派送 N 筆」—— 沒有產品域動作的
	// 型別(period.opened…)只被認領,被別的執行搶先認領的也不算。
	log.Printf("platform-cron 完成(now=%s):逾期 %d 筆、停用 %d 筆、取消到期 %d 筆、產生期別 %d 筆、認領事件 %d 筆、待收款 %d 筆",
		now.Format(time.RFC3339), summary.PastDue, summary.Suspended, summary.ExpiredCancelled,
		summary.PeriodsOpened, summary.Dispatched, summary.Receivables)
	log.Printf("platform-cron 摘要(now=%s):%s", now.Format(time.RFC3339), raw)
}

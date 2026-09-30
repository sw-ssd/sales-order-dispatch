-- 平台 KV 設定(platform.settings)。單一職責:放「不該寫死在程式碼、但營運要能改」的參數 ——
-- 系統 actor 的 user id(排程／consumer 改公司狀態時的稽核主體,G5)與試用／寬限／提前天數
-- (cmd/platform-cron 的參數來源)。
--
-- 為何是 KV 而不是一張有欄位的設定表:這些值由營運工具(store.UpsertSettingTx)與 cmd/seed
-- 兩邊寫入,新增一個參數不該再開一次 migration;型別一律 text,解析(整數與否)由讀取端負責
-- (store.SystemActor 把非數字視為錯誤,不得靜默當 0)。
--
-- 與 cmd/seed/platform.go 的契約(**逐字對應**,改欄位名會讓 seed 寫得進去但讀不到):
--   key 為主鍵 —— seed 的 `ON CONFLICT (key) DO UPDATE/DO NOTHING` 以此為衝突目標;
--   value 為 text NOT NULL;updated_at 有預設值(seed 的 INSERT 不帶它,少了預設值即 23502)。
--   四鍵:system_actor_user_id／trial_days／grace_days／lead_days。
--
-- 硬邊界(00029／S9):platform schema 對業務角色 app_rw 零權限,本表亦**不套 RLS** —— RLS 只
-- 服務租戶資料,平台域靠 schema 隔離 ＋ admin(owner)連線把關。故本檔沒有
-- `SET LOCAL app.current_data_scope`:那是業務表在租戶 interceptor 交易內的慣例,抄來這裡
-- 只會誤導(平台寫入不經租戶交易)。

-- +goose Up
CREATE TABLE IF NOT EXISTS platform.settings (
    key        text PRIMARY KEY,
    value      text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- belt-and-braces(與 00029 檔頭同理):正常狀態下是 no-op —— 00022 是白名單式授權且沒有
-- ALTER DEFAULT PRIVILEGES,本表從未被 GRANT 給 app_rw。留著只為「未來有人手滑加了授權」時
-- 仍關得上門。
REVOKE ALL ON platform.settings FROM app_rw;

-- +goose Down
DROP TABLE IF EXISTS platform.settings;

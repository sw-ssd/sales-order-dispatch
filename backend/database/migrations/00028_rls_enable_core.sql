-- 核心域啟用 RLS（D36）。這是最後一批：核心五表（companies／departments／users／roles／
-- role_permissions）被幾乎所有服務讀取 —— 每個 RPC 的身分解析、稽核的 FK、角色查詢都會摸到 ——
-- 故必須等各域路徑都已走請求層租戶交易與系統範圍交易之後才啟用。任何一條未收斂的存取在此
-- 之後就是全站黑屏（查不到資料、寫不進去、登入全滅），而它不會報錯、只會靜默回 0 列。
--
-- 啟用後的行為（以 app_rw 實測，見 internal/services/rls_core_integration_test.go 與
-- internal/server/rls_core_integration_test.go）：
--   未設 scope              → 五表皆 0 列（fail-closed；寫入被 WITH CHECK 擋，SQLSTATE 42501）
--   scope=all              → 全庫可讀可寫（dbtenant.SystemScopeTx 用的就是這個等級）
--   scope=company+company  → companies（id=當前公司）、departments／users（company 欄比對）
--   scope=department       → departments（id=當前部門）、users（department 欄比對）
--   scope=self             → users 僅自己那一列；departments 無 self 分支故 0 列
--   scope 非空即可讀寫      → roles／role_permissions（共享目錄，寫入權威在服務層 ACL）
--
-- policy 一律不動：五張表的 policy 由 00023 建立、00025 以 NULLIF 重建（USING 與 WITH CHECK
-- 同條件），本檔只切換 ENABLE／FORCE 旗標。FORCE 是必要的 —— 否則 owner 連線（migration／維運）
-- 會繞過 policy，測試與正式環境的行為分歧。
--
-- 對應的程式碼收斂（必須早於本檔，順序見 plans/2026-09-20-rls-tenant-isolation-plan）：
--   company_service／user_service／role_service → dbtenant.Client(ctx, s.db) 與 dbtenant.TxFrom(ctx)
--   server.identityFor／dataScopeForUser、handlers 的登入/註冊/OIDC、auth.TokenManager、
--   authz.Provision → dbtenant.SystemScopeTx（或 auth 內等價的系統範圍交易）
-- +goose Up
ALTER TABLE companies         ENABLE ROW LEVEL SECURITY;
ALTER TABLE companies         FORCE  ROW LEVEL SECURITY;
ALTER TABLE departments       ENABLE ROW LEVEL SECURITY;
ALTER TABLE departments       FORCE  ROW LEVEL SECURITY;
ALTER TABLE users             ENABLE ROW LEVEL SECURITY;
ALTER TABLE users             FORCE  ROW LEVEL SECURITY;
ALTER TABLE roles             ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles             FORCE  ROW LEVEL SECURITY;
ALTER TABLE role_permissions  ENABLE ROW LEVEL SECURITY;
ALTER TABLE role_permissions  FORCE  ROW LEVEL SECURITY;

-- +goose Down
-- 反序對稱：先 NO FORCE 再 DISABLE（少寫 NO FORCE 會留下 FORCE 旗標、少寫 DISABLE 則 RLS 仍生效，
-- 兩者都讓回退後的環境與 00027 的狀態不一致，而 policy 不變故只有旗標看得出來）。
ALTER TABLE companies         NO FORCE ROW LEVEL SECURITY;
ALTER TABLE companies         DISABLE ROW LEVEL SECURITY;
ALTER TABLE departments       NO FORCE ROW LEVEL SECURITY;
ALTER TABLE departments       DISABLE ROW LEVEL SECURITY;
ALTER TABLE users             NO FORCE ROW LEVEL SECURITY;
ALTER TABLE users             DISABLE ROW LEVEL SECURITY;
ALTER TABLE roles             NO FORCE ROW LEVEL SECURITY;
ALTER TABLE roles             DISABLE ROW LEVEL SECURITY;
ALTER TABLE role_permissions  NO FORCE ROW LEVEL SECURITY;
ALTER TABLE role_permissions  DISABLE ROW LEVEL SECURITY;

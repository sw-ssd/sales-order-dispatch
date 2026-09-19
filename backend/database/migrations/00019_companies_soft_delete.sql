-- 公司軟刪除(P2-A,2026-09-20):companies 加 deleted_at,並把識別碼的唯一性從
-- 「表層 UNIQUE」改為「僅未刪除的公司之間唯一」的部分唯一索引。
--
-- 背景(P2-A 缺陷):DeleteCompany 原以硬刪除刪列,而 audit_logs.company_id 是租戶欄
-- (NOT NULL + FK→companies,00009/00010)。只要該公司有任何稽核列(例如曾改名/改狀態,
-- UpdateCompany 就會寫一列 audit_logs),硬刪除即違反 FK;服務層把約束錯誤映射成
-- AlreadyExists,前端於是顯示「識別碼(identifier)已存在,請換一個」並外洩原始 SQL 訊息。
-- 公司改軟刪除後稽核列永遠有主可依,FK 不再阻擋刪除。
--
-- 影響(呼叫端必須配合,已於同一波修改):
--   1. 00005 建表時的識別碼表層 UNIQUE(PG 自動命名為 companies_identifier_key)無法表達
--      `WHERE deleted_at IS NULL`,故本檔移除該約束,改以部分唯一索引表達同一語意:
--      未刪除的公司之間 identifier 仍不可重複;已刪除的公司不佔用識別碼,可被新公司重用。
--   2. 所有公司查詢都必須排除軟刪除列(`deleted_at IS NULL`):ListCompanies、
--      GetCompany、UpdateCompany/DeleteCompany 的存在性檢查、CreateDepartment 與 user_service
--      以公司為條件的存在性檢查、auth_handler 的公司解析(登入/註冊)、seed 的 firstCompanyID。
--
-- 冪等(比照 00008/00010 的 DO $$ 模式):對「已跑過 00005」的既有 DB 補欄位、換約束;
-- 對全新 DB 亦同形,且重複套用為 no-op。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE companies ADD COLUMN IF NOT EXISTS deleted_at timestamptz;
-- +goose StatementEnd

-- +goose StatementBegin
-- 移除表層 UNIQUE(既有 DB 由 00005 的 `identifier text NOT NULL UNIQUE` 建出,名稱由 PG
-- 以 <表>_<欄>_key 自動產生)。DROP CONSTRAINT IF EXISTS 讓全新 DB/重複套用皆為 no-op。
ALTER TABLE companies DROP CONSTRAINT IF EXISTS companies_identifier_key;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS companies_identifier_active_unique
    ON companies (identifier) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS companies_identifier_active_unique;
-- +goose StatementEnd

-- +goose StatementBegin
-- 還原表層 UNIQUE。若庫中已存在「同 identifier 的已刪除列 + 未刪除列」,此處會以
-- 23505(unique_violation)擋下 —— 那是正確訊號(資料本來就違反原約束語意);清掉重複的
-- 已刪除列後再重跑即可。
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'companies_identifier_key' AND conrelid = 'companies'::regclass
    ) THEN
        ALTER TABLE companies ADD CONSTRAINT companies_identifier_key UNIQUE (identifier);
    END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE companies DROP COLUMN IF EXISTS deleted_at;
-- +goose StatementEnd

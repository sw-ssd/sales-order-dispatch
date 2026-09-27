-- 公司層角色的部門可空:notifications / print_logs / print_previews。
--
-- 背景(2026-09-26 以 app_rw 在真 PG 實測):
--   `deptScope`(services/department_master.go)對 company_admin 回 (cid, nil) —— 公司層角色
--   **沒有部門**。但這三張表的 department_id 是 NOT NULL,而寫入端一律是
--   `if did != nil { b = b.SetDepartmentID(*did) }`(notification_triggers.go:36、
--   print_service.go:140/233)。於是 did==nil 時欄位根本沒被設值:
--
--     ent: missing required field "PrintLog.department_id"
--
--   ent 的 required 檢查在 client 端,所以**連 DB 都碰不到**;toConnectError 把它映射成
--   `ent.IsValidationError → SYS-1001`,且 detail 為空 —— 使用者與前端都只看到「參數驗證失敗」,
--   沒有任何線索指向部門欄位。
--
--   已實測的兩個症狀(company_admin,公司 178):
--     - `PrintService.Preview` / `Print` 一律 400 SYS-1001(dept_admin 同一請求 200)。
--     - `SalesOrderService/CreateOrder`:客戶**沒有**子帳號 → 200;客戶**有**子帳號
--       → OnOrderCreated 要寫通知 → 400 SYS-1001(整筆下單回滾)。
--   即「多店客戶收不到下單通知、而且下單本身直接失敗」。
--
-- 為什麼是改 schema 而不是改 Go:
--   1. 寫入端的 `if did != nil` 已經表達了「可能沒有部門」的意圖 —— NOT NULL 才是那個偏差。
--   2. 同一批工作的 sibling 表早就是這個形狀:file_assets / audit_logs / sales_orders /
--      customers 的 department_id 皆可空,且 RLS policy 用 `department_id IS NULL` 表達
--      「公司層共用列」(見 00013/00023/00035)。print_logs/print_previews 的 policy 少了那條
--      `IS NULL` 分支,正是同一份 00037 沒跟上。
--   3. 00041 內部就自相矛盾:notification_templates.department_id 可空(NULL = 公司層範本),
--      notifications 卻 NOT NULL —— 同一支遷移、同一種語意,兩種寫法。
--
-- 語意:department_id IS NULL = 公司層記錄(由無部門的 company_admin/系統觸發),可見範圍與
-- customers/file_assets 的公司層列一致 —— data_scope='company' 者看得到,department 範圍者
-- 看不到(它不屬於任何部門)。
-- +goose Up
-- +goose StatementBegin
ALTER TABLE notifications    ALTER COLUMN department_id DROP NOT NULL;
ALTER TABLE print_logs       ALTER COLUMN department_id DROP NOT NULL;
ALTER TABLE print_previews   ALTER COLUMN department_id DROP NOT NULL;

-- policy 補上 `department_id IS NULL` 分支(比照 file_assets 的 core_file_assets_scope)。
-- 從 pg_policy 讀回 PG 正規化過的運算式再改寫,不手抄內文:NOW 形式固定,
-- 若哪天 policy 被改過而不再是這個形狀,這裡會改不到,故結尾斷言必須命中 3 個。
DO $$
DECLARE
    targets text[][] := ARRAY[
        ['notifications',  'core_notifications_scope'],
        ['print_logs',     'core_print_logs_scope'],
        ['print_previews', 'core_print_previews_scope']
    ];
    rec       record;
    new_qual  text;
    new_check text;
    -- 整段比較式（含左運算元）。只替換右邊的 cast 會產生
    -- `department_id = department_id IS NULL OR ...`（42804：OR 的運算元必須是 boolean）。
    needle    text := 'department_id = (NULLIF(current_setting(''app.current_department_id''::text, true), ''''::text))::bigint';
    ddl       text;
    has_null  boolean;
    n_touched int := 0;
BEGIN
    FOR i IN 1 .. array_length(targets, 1) LOOP
        SELECT c.relname, p.polname, p.polcmd, p.polpermissive,
               coalesce(pg_get_expr(p.polqual, p.polrelid), '')      AS qual,
               coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '') AS withcheck,
               (SELECT string_agg(quote_ident(r.rolname), ', ' ORDER BY r.rolname)
                  FROM unnest(p.polroles) AS ro
                  JOIN pg_roles r ON r.oid = ro
                 WHERE ro <> 0)                                        AS roles
          INTO rec
          FROM pg_policy p
          JOIN pg_class c ON c.oid = p.polrelid
         WHERE c.relnamespace = 'public'::regnamespace
           AND c.relname = targets[i][1]
           AND p.polname = targets[i][2];

        IF NOT FOUND THEN
            RAISE EXCEPTION '部門可空: 找不到 %.%（清單與資料庫不同步）', targets[i][1], targets[i][2];
        END IF;

        -- 已有分支就跳過（可重跑）。
        IF position('department_id IS NULL' in rec.qual) > 0 THEN
            CONTINUE;
        END IF;

        new_qual  := replace(rec.qual,  needle, 'department_id IS NULL OR ' || needle);
        new_check := replace(rec.withcheck, needle, 'department_id IS NULL OR ' || needle);
        IF new_qual = rec.qual AND new_check = rec.withcheck THEN
            RAISE EXCEPTION '部門可空: %.% 找不到部門比較條件可改寫——policy 是否已被改寫？',
                rec.relname, rec.polname;
        END IF;

        ddl := 'DROP POLICY ' || quote_ident(rec.polname) || ' ON ' || quote_ident(rec.relname) || ';'
            || ' CREATE POLICY ' || quote_ident(rec.polname) || ' ON ' || quote_ident(rec.relname)
            || CASE rec.polcmd WHEN 'r' THEN ' FOR SELECT'
                               WHEN 'a' THEN ' FOR INSERT'
                               WHEN 'w' THEN ' FOR UPDATE'
                               WHEN 'd' THEN ' FOR DELETE'
                               ELSE ' FOR ALL' END
            || CASE WHEN rec.polpermissive THEN ' ' ELSE ' AS RESTRICTIVE ' END
            || ' TO ' || coalesce(rec.roles, 'PUBLIC')
            || CASE WHEN new_qual  <> '' THEN ' USING (' || new_qual || ')' ELSE '' END
            || CASE WHEN new_check <> '' THEN ' WITH CHECK (' || new_check || ')' ELSE '' END
            || ';';
        EXECUTE ddl;
        n_touched := n_touched + 1;
    END LOOP;

    -- 後置條件而非「改了幾個」：Up 可重跑（部分已改、部分未改都要能收斂），
    -- 因此斷言的是**結果**——三張表的 policy 都必須有 IS NULL 分支。
    -- 用計數斷言會在「部分已改」時誤判失敗（Round-trip 測試 down→up 就是這個情形）。
    FOR i IN 1 .. array_length(targets, 1) LOOP
        SELECT position('department_id IS NULL' in coalesce(pg_get_expr(p.polqual, p.polrelid), '')) > 0
          INTO has_null
          FROM pg_policy p
          JOIN pg_class c ON c.oid = p.polrelid
         WHERE c.relnamespace = 'public'::regnamespace
           AND c.relname = targets[i][1]
           AND p.polname = targets[i][2];
        IF NOT coalesce(has_null, false) THEN
            RAISE EXCEPTION '部門可空: %.% 改寫後仍無 IS NULL 分支', targets[i][1], targets[i][2];
        END IF;
    END LOOP;
    RAISE NOTICE '部門可空: 已改寫 % 個 policy（三張表皆有 IS NULL 分支）', n_touched;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
DECLARE
    targets text[][] := ARRAY[
        ['notifications',  'core_notifications_scope'],
        ['print_logs',     'core_print_logs_scope'],
        ['print_previews', 'core_print_previews_scope']
    ];
    rec       record;
    new_qual  text;
    new_check text;
    needle    text := 'department_id = (NULLIF(current_setting(''app.current_department_id''::text, true), ''''::text))::bigint';
    ddl       text;
BEGIN
    FOR i IN 1 .. array_length(targets, 1) LOOP
        SELECT c.relname, p.polname, p.polcmd, p.polpermissive,
               coalesce(pg_get_expr(p.polqual, p.polrelid), '')      AS qual,
               coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '') AS withcheck,
               (SELECT string_agg(quote_ident(r.rolname), ', ' ORDER BY r.rolname)
                  FROM unnest(p.polroles) AS ro
                  JOIN pg_roles r ON r.oid = ro
                 WHERE ro <> 0)                                        AS roles
          INTO rec
          FROM pg_policy p
          JOIN pg_class c ON c.oid = p.polrelid
         WHERE c.relnamespace = 'public'::regnamespace
           AND c.relname = targets[i][1]
           AND p.polname = targets[i][2];
        IF NOT FOUND THEN
            RAISE EXCEPTION '部門可空 Down: 找不到 %.%', targets[i][1], targets[i][2];
        END IF;

        -- PG 會把條件重新括號成 `((department_id IS NULL) OR (department_id = X))`，
        -- 所以要連同外層括號一起比對；只比對 `department_id IS NULL OR ` 會找不到而靜默無效。
        new_qual  := replace(rec.qual,
            '((department_id IS NULL) OR (' || needle || '))', '(' || needle || ')');
        new_check := replace(rec.withcheck,
            '((department_id IS NULL) OR (' || needle || '))', '(' || needle || ')');
        IF new_qual = rec.qual AND new_check = rec.withcheck THEN
            RAISE EXCEPTION '部門可空 Down: %.% 找不到已加上的 IS NULL 分支——policy 是否已被改回？',
                rec.relname, rec.polname;
        END IF;

        ddl := 'DROP POLICY ' || quote_ident(rec.polname) || ' ON ' || quote_ident(rec.relname) || ';'
            || ' CREATE POLICY ' || quote_ident(rec.polname) || ' ON ' || quote_ident(rec.relname)
            || CASE rec.polcmd WHEN 'r' THEN ' FOR SELECT'
                               WHEN 'a' THEN ' FOR INSERT'
                               WHEN 'w' THEN ' FOR UPDATE'
                               WHEN 'd' THEN ' FOR DELETE'
                               ELSE ' FOR ALL' END
            || CASE WHEN rec.polpermissive THEN ' ' ELSE ' AS RESTRICTIVE ' END
            || ' TO ' || coalesce(rec.roles, 'PUBLIC')
            || CASE WHEN new_qual  <> '' THEN ' USING (' || new_qual || ')' ELSE '' END
            || CASE WHEN new_check <> '' THEN ' WITH CHECK (' || new_check || ')' ELSE '' END
            || ';';
        EXECUTE ddl;
    END LOOP;
END $$;
-- +goose StatementEnd

-- Down 只在資料庫真的沒有公司層列時才還原 NOT NULL;有 NULL 列就直接失敗,
-- 讓「還原」不會靜默把資料改成違反約束(或反過來悄悄刪列)。
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM notifications  WHERE department_id IS NULL)
        OR EXISTS (SELECT 1 FROM print_logs     WHERE department_id IS NULL)
        OR EXISTS (SELECT 1 FROM print_previews WHERE department_id IS NULL) THEN
        RAISE EXCEPTION '部門可空 Down: 仍有 department_id IS NULL 的列存在(公司層記錄)，不還原 NOT NULL';
    END IF;
END $$;
ALTER TABLE notifications    ALTER COLUMN department_id SET NOT NULL;
ALTER TABLE print_logs       ALTER COLUMN department_id SET NOT NULL;
ALTER TABLE print_previews   ALTER COLUMN department_id SET NOT NULL;
-- +goose StatementEnd

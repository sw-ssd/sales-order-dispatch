-- 補正 RLS policy 的 GUC 取值形式：`current_setting('app.current_*', true)` 一律包 NULLIF。
--
-- 背景（2026-09-26 以 app_rw 在真 PG 實測）：
--   `SET LOCAL` 對自訂 GUC 會在 **session** 層留下空字串 placeholder，交易結束後該 GUC 還原成
--   `''` 而不是 unset（見 internal/auth/rls.go 的 RLSStatements）。pgx stdlib 的 ResetSession
--   是 noop，不會清除它。於是**同一條池化連線**上，任何「不需要設該 GUC」的請求形狀（例如
--   company_admin 或無部門的 dept_admin：不設 app.current_department_id；或
--   dbtenant.SystemScopeTx：只設 app.current_data_scope）在讀到 `''::bigint` 時，policy 直接以
--   ```
--   ERROR:  invalid input syntax for type bigint: ""   (SQLSTATE 22P02)
--   ```
--   中止**整個交易**。PG 不保證 `OR` 的求值順序，故 `data_scope='all'` 也不會短路掉這個 cast。
--
--   症狀：不是「查不到」，是整個請求炸掉，且**取決於連線歷史**（同一 shape 在乾淨連線成功、
--   在服務過別的 shape 的連線失敗）。00025 已把更早的 18 個 policy 正規化為 NULLIF 形式，
--   但 00031 之後新建表的 policy 又用了舊式寫法，於是 auth 路徑（users/companies/customers/…）
--   正常、業務表全滅。
--
-- 為何用 catalog 驅動改寫而不是手抄 policy 內文：手抄 22 份必須與當下的定義逐字一致，任何
-- 一處漂移都會**靜默改變授權語意**（可能是放寬）。這裡從 pg_policy 讀回 PG 自己正規化過的
-- 運算式，只做那一個字串替換，其餘（USING／WITH CHECK／FOR 子句／PERMISSIVE／roles）原樣重建，
-- 並在結尾斷言「實際改到的數量 == 預期」，數量不符即讓整個遷移失敗。
--
-- 另見內部慣例：policy 隨建表、ENABLE＋FORCE 另開一檔（00032/00034/…/00049）。本遷移只動
-- policy 內文，**不碰** relrowsecurity／relforcerowsecurity。
-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    -- 本遷移負責的表/政策清單（00031 之後新建、且 policy 用了未正規化 cast 者）。
    targets text[][] := ARRAY[
        ['announcements',               'core_announcements_scope'],
        ['customer_products',           'core_customer_products_scope'],
        ['file_assets',                 'core_file_assets_scope'],
        ['logistics_deliveries',        'core_logistics_deliveries_scope'],
        ['logistics_delivery_events',   'core_logistics_delivery_events_insert'],
        ['logistics_delivery_events',   'core_logistics_delivery_events_scope'],
        ['logistics_drivers',           'core_logistics_drivers_scope'],
        ['logistics_proofs',            'core_logistics_proofs_scope'],
        ['notification_templates',      'core_notification_templates_scope'],
        ['notifications',               'core_notifications_scope'],
        ['order_counters',              'core_order_counters_scope'],
        ['print_logs',                  'core_print_logs_scope'],
        ['print_previews',              'core_print_previews_scope'],
        ['promo_tags',                  'core_promo_tags_scope'],
        ['return_request_items',        'core_return_request_items_scope'],
        ['return_requests',             'core_return_requests_scope'],
        ['sales_order_events',          'core_sales_order_events_insert'],
        ['sales_order_events',          'core_sales_order_events_scope'],
        ['sales_order_items',           'core_sales_order_items_scope'],
        ['sales_orders',                'core_sales_orders_scope'],
        ['user_devices',                'core_user_devices_scope'],
        ['vehicles',                    'core_vehicles_scope']
    ];
    rec        record;
    new_qual   text;
    new_check  text;
    ddl        text;
    n_touched  int := 0;
BEGIN
    FOR i IN 1 .. array_length(targets, 1) LOOP
        SELECT c.relname, p.polname, p.polcmd,
               p.polpermissive,
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
            RAISE EXCEPTION 'RLS 正規化: 找不到 %.%（清單與資料庫不同步；policy 是否已被更名或刪除？）',
                targets[i][1], targets[i][2];
        END IF;

        -- 只做這一個替換：把裸 cast 包成 NULLIF。四個 GUC 都是同一形狀。
        new_qual  := rec.qual;
        new_check := rec.withcheck;
        FOREACH ddl IN ARRAY ARRAY['app.current_company_id', 'app.current_department_id',
                                   'app.current_user_id', 'app.current_customer_id'] LOOP
            new_qual := replace(new_qual,
                '(current_setting(''' || ddl || '''::text, true))::bigint',
                '(NULLIF(current_setting(''' || ddl || '''::text, true), ''''::text))::bigint');
            new_check := replace(new_check,
                '(current_setting(''' || ddl || '''::text, true))::bigint',
                '(NULLIF(current_setting(''' || ddl || '''::text, true), ''''::text))::bigint');
        END LOOP;

        -- 該 policy 若本來就是 NULLIF 形式，這裡會是空操作 —— 不該發生，屬清單維護錯誤。
        IF new_qual = rec.qual AND new_check = rec.withcheck THEN
            RAISE EXCEPTION 'RLS 正規化: %.% 沒有任何裸 cast 可補（已是 NULLIF 形式？）——請修正清單',
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

    IF n_touched <> array_length(targets, 1) THEN
        RAISE EXCEPTION 'RLS 正規化: 預期改寫 % 個 policy，實際 %', array_length(targets, 1), n_touched;
    END IF;

    RAISE NOTICE 'RLS 正規化: 已補正 % 個 policy 的 GUC NULLIF 取值', n_touched;
END $$;
-- +goose StatementEnd

-- +goose Down
-- 只把本遷移負責的清單還原回裸 cast 形式（不碰 00025 等早已是 NULLIF 的 policy）。
-- +goose StatementBegin
DO $$
DECLARE
    targets text[][] := ARRAY[
        ['announcements',               'core_announcements_scope'],
        ['customer_products',           'core_customer_products_scope'],
        ['file_assets',                 'core_file_assets_scope'],
        ['logistics_deliveries',        'core_logistics_deliveries_scope'],
        ['logistics_delivery_events',   'core_logistics_delivery_events_insert'],
        ['logistics_delivery_events',   'core_logistics_delivery_events_scope'],
        ['logistics_drivers',           'core_logistics_drivers_scope'],
        ['logistics_proofs',            'core_logistics_proofs_scope'],
        ['notification_templates',      'core_notification_templates_scope'],
        ['notifications',               'core_notifications_scope'],
        ['order_counters',              'core_order_counters_scope'],
        ['print_logs',                  'core_print_logs_scope'],
        ['print_previews',              'core_print_previews_scope'],
        ['promo_tags',                  'core_promo_tags_scope'],
        ['return_request_items',        'core_return_request_items_scope'],
        ['return_requests',             'core_return_requests_scope'],
        ['sales_order_events',          'core_sales_order_events_insert'],
        ['sales_order_events',          'core_sales_order_events_scope'],
        ['sales_order_items',           'core_sales_order_items_scope'],
        ['sales_orders',                'core_sales_orders_scope'],
        ['user_devices',                'core_user_devices_scope'],
        ['vehicles',                    'core_vehicles_scope']
    ];
    rec       record;
    new_qual  text;
    new_check text;
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
            RAISE EXCEPTION 'RLS 正規化 Down: 找不到 %.%', targets[i][1], targets[i][2];
        END IF;

        new_qual  := rec.qual;
        new_check := rec.withcheck;
        FOREACH ddl IN ARRAY ARRAY['app.current_company_id', 'app.current_department_id',
                                   'app.current_user_id', 'app.current_customer_id'] LOOP
            new_qual := replace(new_qual,
                '(NULLIF(current_setting(''' || ddl || '''::text, true), ''''::text))::bigint',
                '(current_setting(''' || ddl || '''::text, true))::bigint');
            new_check := replace(new_check,
                '(NULLIF(current_setting(''' || ddl || '''::text, true), ''''::text))::bigint',
                '(current_setting(''' || ddl || '''::text, true))::bigint');
        END LOOP;

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

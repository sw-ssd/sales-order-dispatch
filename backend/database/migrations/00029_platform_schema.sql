-- 平台域 schema(D34–D39)。平台域(方案／訂閱／帳務／操作者)與業務域**不 JOIN**、不共用
-- 任何一張表,只以 company_id 對照租戶;故獨立成一個 schema,讓「業務連線碰不到帳務」成為
-- 結構事實而非紀律。
--
-- 硬邊界(S9／§3.3):業務連線角色 app_rw 對本 schema **零權限**。本檔沒有任何 GRANT 給
-- app_rw 的敘述;平台域的存取一律走 DATABASE_ADMIN_URL(owner 連線)。
-- 真正的邊界來自兩件事實(不是靠下面的 REVOKE):
--   ① 00022 對 app_rw 是**白名單式**授權 —— 只逐一列出 18 張業務表(table)與其 sequence,
--      且**沒有** ALTER DEFAULT PRIVILEGES(故新 schema、新表一律不會外溢到 app_rw);
--   ② platform schema 從建立到現在**從未被 GRANT** 給 app_rw,本檔亦不授權。
-- 下面的 REVOKE 只是 belt-and-braces:萬一未來有人手滑加了 GRANT,或 PUBLIC 的預設權限
-- 被環境改動,這裡會再把門關上;它們在正常狀態下是 no-op(無權限可撤)。
--
-- 後人規則(承 00022 的同一條慣例):**新增業務表必須在其 migration 內自行 GRANT 給 app_rw**;
-- `platform` schema 是刻意的例外 —— 永不授權、也不套 RLS(policy 只服務租戶資料,
-- 平台域靠 schema 隔離 + owner 連線把關)。

-- +goose Up
CREATE SCHEMA IF NOT EXISTS platform;

-- 方案（價目的容器）；被引用後只歸檔不刪除，避免歷史帳斷鏈。
CREATE TABLE IF NOT EXISTS platform.plans (
    id           bigserial PRIMARY KEY,
    code         text        NOT NULL UNIQUE,
    name         text        NOT NULL,
    status       text        NOT NULL DEFAULT 'active',   -- active | archived
    sort_order   integer     NOT NULL DEFAULT 0,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

-- 價格史：同一方案可有多次調價，以 effective_from 取「當期生效價」。
CREATE TABLE IF NOT EXISTS platform.plan_prices (
    id             bigserial PRIMARY KEY,
    plan_id        bigint      NOT NULL REFERENCES platform.plans(id),
    billing_cycle  text        NOT NULL,                  -- monthly | yearly
    base_price     numeric(12,2) NOT NULL,
    seat_price     numeric(12,2) NOT NULL,
    currency       text        NOT NULL DEFAULT 'TWD',
    effective_from timestamptz NOT NULL DEFAULT now(),
    created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS plan_prices_plan_effective_idx
    ON platform.plan_prices (plan_id, billing_cycle, effective_from DESC);

-- 可賣的功能與限額清單（**與 OpenFGA 的 resource/action 是不同軸**：這裡管「買了沒有」）。
CREATE TABLE IF NOT EXISTS platform.features (
    code        text        PRIMARY KEY,                  -- limit.seats / feature.printing …
    type        text        NOT NULL,                     -- boolean | integer
    unit        text        NOT NULL DEFAULT '',           -- 席 / 客戶 / 商品 / 部門 / GB
    description text        NOT NULL DEFAULT ''
);

-- 方案 × 功能：boolean 用 enabled，數值上限制用 limit_value（NULL = 不限）。
CREATE TABLE IF NOT EXISTS platform.plan_entitlements (
    plan_id     bigint NOT NULL REFERENCES platform.plans(id),
    feature_code text  NOT NULL REFERENCES platform.features(code),
    enabled     boolean NOT NULL DEFAULT false,
    limit_value bigint,
    PRIMARY KEY (plan_id, feature_code)
);

-- 一租戶一份合約；UNIQUE 保證同時只有一份未取消的訂閱。
CREATE TABLE IF NOT EXISTS platform.subscriptions (
    id               bigserial PRIMARY KEY,
    company_id       bigint      NOT NULL,
    plan_id          bigint      NOT NULL REFERENCES platform.plans(id),
    seat_count       integer     NOT NULL DEFAULT 1,
    -- 計費週期（G1）：期別產生必須依它決定「加一個月」或「加一年」，
    -- 否則年繳方案每次只會產生一個月期別（少收 11 個月）。
    billing_cycle    text        NOT NULL DEFAULT 'monthly', -- monthly | yearly
    status           text        NOT NULL,                -- trialing | active | past_due | suspended | cancelled
    trial_ends_at    timestamptz,
    grace_until      timestamptz,
    dunning_attempts integer     NOT NULL DEFAULT 0,
    payment_provider text        NOT NULL DEFAULT 'manual',
    invoice_provider text        NOT NULL DEFAULT 'manual',
    external_ref     text,
    started_at       timestamptz NOT NULL DEFAULT now(),
    cancelled_at     timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS subscriptions_active_company_unique
    ON platform.subscriptions (company_id) WHERE status <> 'cancelled';

-- 帳的單位：金額／期間／價格快照一律在此（四個「不可回填」欄位）。
CREATE TABLE IF NOT EXISTS platform.subscription_periods (
    id               bigserial PRIMARY KEY,
    subscription_id  bigint      NOT NULL REFERENCES platform.subscriptions(id),
    period_no        integer     NOT NULL,
    period_start     timestamptz NOT NULL,
    period_end       timestamptz NOT NULL,
    plan_id          bigint      NOT NULL,                -- 快照
    unit_price       numeric(12,2) NOT NULL,              -- 快照
    seat_price       numeric(12,2) NOT NULL,              -- 快照
    seat_count       integer     NOT NULL,                -- 快照
    amount           numeric(12,2) NOT NULL,
    currency         text        NOT NULL DEFAULT 'TWD',
    status           text        NOT NULL DEFAULT 'open', -- open | paid | void
    paid_at          timestamptz,
    invoice_no       text,
    invoice_status   text,
    buyer_tax_id     text,
    carrier          text,
    payment_provider text        NOT NULL DEFAULT 'manual',
    external_ref     text,
    -- 短收／溢收等人工註記（G8）：不改變期別金額，只留對帳線索。
    note             text        NOT NULL DEFAULT '',
    created_at       timestamptz NOT NULL DEFAULT now(),
    UNIQUE (subscription_id, period_no)
);
-- webhook 冪等：同一 provider 的交易號只能入帳一次（v1 人工時 external_ref 為 NULL）。
CREATE UNIQUE INDEX IF NOT EXISTS periods_provider_ref_unique
    ON platform.subscription_periods (payment_provider, external_ref)
    WHERE external_ref IS NOT NULL;

-- 例外（簽約承諾）；欄位強制可追溯：誰承諾、為何、何時到期。
CREATE TABLE IF NOT EXISTS platform.tenant_overrides (
    id           bigserial PRIMARY KEY,
    company_id   bigint      NOT NULL,
    feature_code text        NOT NULL REFERENCES platform.features(code),
    enabled      boolean,
    limit_value  bigint,
    reason       text        NOT NULL,
    owner        text        NOT NULL,                    -- 承諾者（平台側人員）
    expires_at   timestamptz,
    revoked_at   timestamptz,
    created_by   bigint      NOT NULL,                    -- platform.operators.id
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS tenant_overrides_active_unique
    ON platform.tenant_overrides (company_id, feature_code) WHERE revoked_at IS NULL;

-- outbox：跨域副作用由此驅動（凍結公司、通知）。
CREATE TABLE IF NOT EXISTS platform.events (
    id             bigserial PRIMARY KEY,
    aggregate_type text        NOT NULL,
    aggregate_id   bigint      NOT NULL,
    event_type     text        NOT NULL,
    payload        jsonb       NOT NULL DEFAULT '{}'::jsonb,
    dispatched_at  timestamptz,
    attempts       integer     NOT NULL DEFAULT 0,
    created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS events_undispatched_idx ON platform.events (created_at) WHERE dispatched_at IS NULL;

-- 平台操作者白名單（S8）：不與租戶 users 有任何關聯。
CREATE TABLE IF NOT EXISTS platform.operators (
    id            bigserial PRIMARY KEY,
    email         text        NOT NULL UNIQUE,
    name          text        NOT NULL DEFAULT '',
    role          text        NOT NULL DEFAULT 'operator', -- operator | admin
    status        text        NOT NULL DEFAULT 'active',   -- active | disabled
    last_login_at timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

-- 平台側稽核（S9）：actor 為 operator_id，**不 FK 租戶 users**（對方無此列）。
CREATE TABLE IF NOT EXISTS platform.audit_logs (
    id          bigserial PRIMARY KEY,
    operator_id bigint      NOT NULL REFERENCES platform.operators(id),
    action      text        NOT NULL,
    target_type text        NOT NULL,                      -- company | plan | subscription | operator
    target_id   text        NOT NULL,
    reason      text        NOT NULL DEFAULT '',
    before      jsonb,
    after       jsonb,
    ip_address  text        NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS platform_audit_created_idx ON platform.audit_logs (created_at DESC);

-- 硬邊界(見檔頭):正常狀態下以下三句都是 no-op —— 本 schema 從未被 GRANT 給 app_rw、
-- 00022 也沒有 ALTER DEFAULT PRIVILEGES 可外溢。留著只為「未來有人手滑加了授權」時仍關得上門。
REVOKE ALL ON SCHEMA platform FROM PUBLIC;
REVOKE ALL ON ALL TABLES IN SCHEMA platform FROM app_rw;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA platform FROM app_rw;

-- +goose Down
-- +goose StatementBegin
DROP SCHEMA IF EXISTS platform CASCADE;
-- +goose StatementEnd

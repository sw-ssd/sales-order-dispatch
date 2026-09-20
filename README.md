# sales-order-dispatch

多公司訂出貨系統 1.0 monorepo：backend（Go）/ frontend（SolidJS）/ app（Flutter）/ infra。

## 啟動

```bash
task infra:start         # 起 PostgreSQL / Valkey / Gotenberg（docker-compose.dev.yml）
# 環境變數由 shell 讀取（全部 key 見 backend/.env.example，未載入 .env 檔）：
#   DATABASE_URL       業務連線 —— 非 owner 的 app_rw（RLS 對它生效）
#   DATABASE_ADMIN_URL owner 連線 —— 遷移／seed／OpenFGA／平台域；未設時沿用 DATABASE_URL
# 未設 DATABASE_URL 時會落到 config 的預設值（postgres superuser）→ RLS 全被繞過，務必自行 export。
export DATABASE_URL=postgres://app_rw:app_rw@localhost:5432/salesorder?sslmode=disable
export DATABASE_ADMIN_URL=postgres://postgres:postgres@localhost:5432/salesorder?sslmode=disable
task backend:db:app-password   # 設定 app_rw 密碼（00022 只建角色與授權、不設密碼）
                               # 需要本機有 psql；沒有時可 `podman exec <postgres 容器> psql -U postgres -d salesorder -c "ALTER ROLE app_rw WITH PASSWORD 'app_rw'"`
task backend:migrate:up  # 建 schema（含 OpenFGA datastore 版本表）
task backend:seed        # 內建角色／權限 與開發者帳號（冪等；owner 連線＋系統範圍交易）
task backend:dev         # 開發模式（backend air hot reload 等）
```

## 文件

- 計畫總索引：`docs/superpowers/plans/README.md`
- 設計規格書：`docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`

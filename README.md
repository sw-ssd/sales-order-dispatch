# sales-order-dispatch

多公司訂出貨系統 1.0 monorepo：backend（Go）/ frontend（SolidJS）/ app（Flutter）/ infra。

## 啟動

```bash
task infra:start   # 起 PostgreSQL / Valkey / Gotenberg（docker-compose.dev.yml）
task dev           # 開發模式（backend air hot reload 等）
```

## 文件

- 計畫總索引：`docs/superpowers/plans/README.md`
- 設計規格書：`docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`

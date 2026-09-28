# 現役設計文件（architecture）

現役架構／設計決策的解釋文件。檔名不再帶日期（2026-09-29 重整）；歷史版本由 git history 保存。

| 檔案 | 涵蓋 |
|---|---|
| `openfga-authz.md` | 授權引擎 Casbin/CASL → OpenFGA + RLS（D32） |
| `backend-go8-structure.md` | 後端目錄／DI／config 對齊 go8（D31） |
| `saas-billing-entitlements.md` | 平台域、計費、權益（S1–S11） |
| `app-flutter-stack.md` | App 技術棧（D29） |
| `frontend-ui-library.md` | UI 元件庫化（Phase 1） |
| `frontend-forms-tanstack.md` | 表單改 TanStack Form（Phase 2） |
| `frontend-tables.md` | 表格與資料層（Phase 3） |
| `frontend-tailkit-ark.md` | Ark 行為 × Tailkit 視覺 |

領域需求與可驗證行為見 `../1.0-requirements/`；範圍與欄位合約見 `../1.0-contract.md`；決策見 `../decisions.md`。

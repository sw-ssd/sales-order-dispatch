// RBACModel 與 RBACPolicy 為 Casbin model / 內建角色 policy 的編譯期內嵌
// (internal/auth 的 Enforcer 使用;go:embed 保證測試與正式 binary 皆可直接讀取)。
// 原始檔:internal/auth/rbac_model.conf 與 internal/auth/rbac_policy.csv。
// 依 D31 慣例,授權設定位居 internal/auth,不放 config/(config 僅逐檔 envconfig struct)。
package auth

import _ "embed"

//go:embed rbac_model.conf
var RBACModel string

//go:embed rbac_policy.csv
var RBACPolicy string

package config

import _ "embed"

// RBACModel 與 RBACPolicy 為 Casbin model / 內建角色 policy 的編譯期內嵌
// (internal/auth 的 Enforcer 使用;go:embed 保證測試與正式 binary 皆可直接讀取)。
// 原始檔:config/rbac_model.conf 與 config/rbac_policy.csv。
//
//go:embed rbac_model.conf
var RBACModel string

//go:embed rbac_policy.csv
var RBACPolicy string

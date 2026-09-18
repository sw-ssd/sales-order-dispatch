package openfga

// modelDSL 為授權 model.fga 的 Go 原始字串版本(避免依賴環境 go:embed 之工具鏈 bug;
// model 與 repo 同版本控管)。資源型別固定、角色→權限對映由 role_permissions→tuples 承載(資料驅動)。
// 物件狀態條件由 domain 狀態機處理,不進 CEL。
const modelDSL = `model
  schema 1.1
type user
type role
  relations
    define assigned: [user]
type company
  relations
    define member: [role#assigned, user]
    define admin: [role#assigned, user]
    define can_read: [role#assigned, company#member, user]
    define can_write: [role#assigned, company#admin, user]
type department
  relations
    define company: [company]
    define member: [role#assigned, user]
    define admin: [role#assigned, user]
    define can_read: [role#assigned, department#member, user]
    define can_write: [role#assigned, department#admin, user]
type ability
  relations
    define can_read: [role#assigned, user]
    define can_write: [role#assigned, user]
`

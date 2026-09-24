package openfga

// modelDSL 為授權 model.fga 的 Go 原始字串版本(避免依賴環境 go:embed 之工具鏈 bug;
// model 與 repo 同版本控管)。資源型別固定、角色→權限對映由 role_permissions→tuples 承載(資料驅動)。
// 物件狀態條件由 domain 狀態機處理,不進 CEL。DSL 字串內不可放 // 註解(解析器不收,實測)。
//
// logistics 執行層(D32/10.8):vehicle/driver/logistics_delivery 帶 company/department 租戶 parent 邊
// (物件建立時經 AfterCommit 寫 tuple);logistics_delivery#driver 寫 userset 主體
// 「driver:<id>#assignee」,讓被指派司機本人沿 driver#assignee 到達 can_read/can_write
// (10.8:被指派司機可操作其 delivery、他人 403)。
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
    define primary_account: [user]
    define can_read: [role#assigned, user] but not primary_account
    define can_write: [role#assigned, user] but not primary_account

type vehicle
  relations
    define company: [company#member]
    define department: [department#member]

type driver
  relations
    define company: [company#member]
    define department: [department#member]
    define assignee: [user]

type logistics_delivery
  relations
    define company: [company#member]
    define department: [department#member]
    define manager: [department#admin]
    define driver: [driver#assignee]
    define assigned_by: [user]
    define can_read: company or department or manager or driver or assigned_by
    define can_write: manager or driver or assigned_by
`

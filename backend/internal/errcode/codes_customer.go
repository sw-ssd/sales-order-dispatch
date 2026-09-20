package errcode

import "connectrpc.com/connect"

// 客戶域：樣板域，示範 1xxx／2xxx／3xxx 三類寫法，其餘域照抄。
var (
	// CustomerNameRequired 用於建檔／更新時的名稱空白（customer_service.go 兩處）。
	CustomerNameRequired = MustRegister(Code{id: "CUST-1001", domain: DomainCustomer,
		connectCode: connect.CodeInvalidArgument, message: "客戶名稱不可為空"})

	// CustomerCodeExists **目前無呼叫點**（Task 5 實查）：customer_code 由 counter 自動取號
	// （nextCustomerCode），沒有任何「客戶端提供編號」或「編號重複」的前置檢查；真撞號只會是
	// DB 約束錯誤 → SYS-3002（ent 無法分辨是哪一種約束）。
	// 保留理由：編號是對外識別，若日後開放「手動指定編號」或「匯入客戶」就需要此碼；長期不使用
	// 應考慮移除或標 deprecated（碼一旦發佈不得重用或改義，故先留著比事後借用別的碼安全）。
	CustomerCodeExists = MustRegister(Code{id: "CUST-2001", domain: DomainCustomer,
		connectCode: connect.CodeAlreadyExists, message: "客戶編號 {code} 已存在"})

	// CustomerDeleted **目前無呼叫點**（Task 5 實查）：已軟刪除的客戶在更新／查詢路徑以
	// `DeletedAtIsNil()` 過濾 → 與「不存在」「不在可見範圍」一律收斂為 SYS-4002，這是 Plan A
	// 刻意的 anti-oracle 設計（不以錯誤碼洩漏他租戶／已刪列的存在性）；改回本碼會破壞該性質。
	// 保留理由：若日後有「明確知道自己操作的是已刪列」的情境（例：復原流程的前置拒絕）才需要；
	// 長期不使用應考慮移除或標 deprecated。
	CustomerDeleted = MustRegister(Code{id: "CUST-3001", domain: DomainCustomer,
		connectCode: connect.CodeFailedPrecondition, message: "客戶已刪除，無法更新"})
)

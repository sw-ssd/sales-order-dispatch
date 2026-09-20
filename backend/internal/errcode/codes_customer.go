package errcode

import "connectrpc.com/connect"

// 客戶域：樣板域，示範 1xxx／2xxx／3xxx 三類寫法，其餘域照抄。
var (
	CustomerNameRequired = MustRegister(Code{id: "CUST-1001", domain: DomainCustomer,
		connectCode: connect.CodeInvalidArgument, message: "客戶名稱不可為空"})

	CustomerCodeExists = MustRegister(Code{id: "CUST-2001", domain: DomainCustomer,
		connectCode: connect.CodeAlreadyExists, message: "客戶編號 {code} 已存在"})

	CustomerDeleted = MustRegister(Code{id: "CUST-3001", domain: DomainCustomer,
		connectCode: connect.CodeFailedPrecondition, message: "客戶已刪除，無法更新"})
)

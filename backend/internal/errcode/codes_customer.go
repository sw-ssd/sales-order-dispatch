package errcode

import "connectrpc.com/connect"

// 客戶域：樣板域，示範 1xxx／2xxx／3xxx 三類寫法，其餘域照抄。
var (
	CustomerNameRequired = MustRegister(Code{ID: "CUST-1001", Domain: DomainCustomer,
		ConnectCode: connect.CodeInvalidArgument, Message: "客戶名稱不可為空"})

	CustomerCodeExists = MustRegister(Code{ID: "CUST-2001", Domain: DomainCustomer,
		ConnectCode: connect.CodeAlreadyExists, Message: "客戶編號 {code} 已存在"})

	CustomerDeleted = MustRegister(Code{ID: "CUST-3001", Domain: DomainCustomer,
		ConnectCode: connect.CodeFailedPrecondition, Message: "客戶已刪除，無法更新"})
)

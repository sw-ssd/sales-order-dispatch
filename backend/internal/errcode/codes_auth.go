package errcode

import "connectrpc.com/connect"

// 認證域：登入、鎖定、註冊流程與公司狀態。
var (
	// AuthBadCredentials 刻意不區分「帳號不存在」與「密碼錯誤」（防帳號列舉）。
	// ID 為 4xxx（非 1xxx）：憑證錯誤必須是對外 Unauthenticated，而 1xxx 區段只允許
	// InvalidArgument——區段規則是硬規則，碼必須落在語意相符的區段。
	AuthBadCredentials = MustRegister(Code{ID: "AUTH-4003", Domain: DomainAuth,
		ConnectCode: connect.CodeUnauthenticated, Message: "帳號或密碼錯誤"})

	// AuthLocked 帶 details.until（解鎖時間）。
	// ID 為 3xxx（非 1xxx）：帳號鎖定是狀態而非參數問題（FailedPrecondition），
	// 1xxx 只允許 InvalidArgument。
	AuthLocked = MustRegister(Code{ID: "AUTH-3003", Domain: DomainAuth,
		ConnectCode: connect.CodeFailedPrecondition, Message: "帳號已鎖定，請於 {until} 後再試"})

	AuthRegistrationRequired = MustRegister(Code{ID: "AUTH-3001", Domain: DomainAuth,
		ConnectCode: connect.CodeFailedPrecondition, Message: "尚未完成註冊"})

	AuthTempPasswordExpired = MustRegister(Code{ID: "AUTH-3002", Domain: DomainAuth,
		ConnectCode: connect.CodeFailedPrecondition, Message: "臨時密碼已過期，請聯繫管理員重置"})

	AuthUnauthenticated = MustRegister(Code{ID: "AUTH-4001", Domain: DomainAuth,
		ConnectCode: connect.CodeUnauthenticated, Message: "未登入"})

	// AuthCompanyInactive 為公司停用連鎖（含欠費凍結）。
	AuthCompanyInactive = MustRegister(Code{ID: "AUTH-4002", Domain: DomainAuth,
		ConnectCode: connect.CodePermissionDenied, Message: "所屬公司已停用"})
)

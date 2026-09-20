package errcode

import "connectrpc.com/connect"

// 認證域：登入、鎖定、註冊流程與公司狀態。
var (
	// AuthBadCredentials 刻意不區分「帳號不存在」與「密碼錯誤」（防帳號列舉）。
	// 與 AuthUnauthenticated（AUTH-4001）同為 Unauthenticated 但語意不同，**不可合併**：
	// 本碼＝「登入嘗試的憑證錯誤」（登入端點回此碼，前端留在登入頁顯示訊息）；
	// AUTH-4001＝「未帶有效身分就存取需登入的端點」（前端導向登入頁）。
	// ID 為 4xxx（非 1xxx）：憑證錯誤必須是對外 Unauthenticated，而 1xxx 區段只允許
	// InvalidArgument——區段規則是硬規則，碼必須落在語意相符的區段。
	AuthBadCredentials = MustRegister(Code{id: "AUTH-4003", domain: DomainAuth,
		connectCode: connect.CodeUnauthenticated, message: "帳號或密碼錯誤"})

	// AuthLocked 帶 details.until（解鎖時間）。
	// ID 為 3xxx（非 1xxx）：帳號鎖定是狀態而非參數問題（FailedPrecondition），
	// 1xxx 只允許 InvalidArgument。
	AuthLocked = MustRegister(Code{id: "AUTH-3003", domain: DomainAuth,
		connectCode: connect.CodeFailedPrecondition, message: "帳號已鎖定，請於 {until} 後再試"})

	AuthRegistrationRequired = MustRegister(Code{id: "AUTH-3001", domain: DomainAuth,
		connectCode: connect.CodeFailedPrecondition, message: "尚未完成註冊"})

	AuthTempPasswordExpired = MustRegister(Code{id: "AUTH-3002", domain: DomainAuth,
		connectCode: connect.CodeFailedPrecondition, message: "臨時密碼已過期，請聯繫管理員重置"})

	// AuthUnauthenticated 為「未帶有效身分就存取需登入的端點」（前端導向登入頁）。
	// 與 AuthBadCredentials（AUTH-4003）同為 Unauthenticated 但語意不同，**不可合併**：
	// 本碼沒有登入嘗試的上下文，前端不應在登入頁顯示表單錯誤。
	AuthUnauthenticated = MustRegister(Code{id: "AUTH-4001", domain: DomainAuth,
		connectCode: connect.CodeUnauthenticated, message: "未登入"})

	// AuthCompanyInactive 為公司停用連鎖（含欠費凍結）。
	AuthCompanyInactive = MustRegister(Code{id: "AUTH-4002", domain: DomainAuth,
		connectCode: connect.CodePermissionDenied, message: "所屬公司已停用"})
)

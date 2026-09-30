package errcode

import "connectrpc.com/connect"

// 平台域：SaaS 訂閱、配額與收款（Plan B／C 直接使用）。
var (
	// PlatformSubscriptionInactive 為 suspended／cancelled（合約狀態問題，非權限問題）。
	PlatformSubscriptionInactive = MustRegister(Code{id: "PLAT-3001", domain: DomainPlatform,
		connectCode: connect.CodeFailedPrecondition, message: "訂閱狀態不允許此操作"})

	// PlatformLimitExceeded 帶 details{feature, used, limit}；前端導向升級方案。
	PlatformLimitExceeded = MustRegister(Code{id: "PLAT-5001", domain: DomainPlatform,
		connectCode: connect.CodeFailedPrecondition, message: "已達方案上限（{used}/{limit}），請升級方案"})

	// PlatformFeatureNotInPlan 帶 details{feature}。
	PlatformFeatureNotInPlan = MustRegister(Code{id: "PLAT-5002", domain: DomainPlatform,
		connectCode: connect.CodeFailedPrecondition, message: "目前方案未包含此功能，請升級方案"})

	// PlatformPaymentConflict 為收款衝突（期別已付款、金額不符）。
	PlatformPaymentConflict = MustRegister(Code{id: "PLAT-3002", domain: DomainPlatform,
		connectCode: connect.CodeFailedPrecondition, message: "收款衝突：{reason}"})

	// PlatformOperatorGovernance 為**操作者治理的不變式**被違反（停用自己、停用最後一位 admin）。
	//
	// 為什麼要有專碼而不是共用 SYS-3002：這不是「資料庫約束」也不是「參數錯」，而是「這個操作
	// 會讓平台**失去可管理性**」（沒有任何 admin 能再管理白名單），處置方式與其他失敗都不同
	// （換一個 admin 來做，而不是改參數或重試）；訊息帶著可行動的 reason（例：請先新增另一位
	// admin），前端可以直接顯示。
	PlatformOperatorGovernance = MustRegister(Code{id: "PLAT-3003", domain: DomainPlatform,
		connectCode: connect.CodeFailedPrecondition, message: "此操作會讓平台失去可管理性：{reason}"})
)

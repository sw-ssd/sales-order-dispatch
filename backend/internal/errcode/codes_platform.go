package errcode

import "connectrpc.com/connect"

// 平台域：SaaS 訂閱、配額與收款（Plan B／C 直接使用）。
var (
	// PlatformSubscriptionInactive 為 suspended／cancelled（合約狀態問題，非權限問題）。
	PlatformSubscriptionInactive = MustRegister(Code{ID: "PLAT-3001", Domain: DomainPlatform,
		ConnectCode: connect.CodeFailedPrecondition, Message: "訂閱狀態不允許此操作"})

	// PlatformLimitExceeded 帶 details{feature, used, limit}；前端導向升級方案。
	PlatformLimitExceeded = MustRegister(Code{ID: "PLAT-5001", Domain: DomainPlatform,
		ConnectCode: connect.CodeFailedPrecondition, Message: "已達方案上限（{used}/{limit}），請升級方案"})

	// PlatformFeatureNotInPlan 帶 details{feature}。
	PlatformFeatureNotInPlan = MustRegister(Code{ID: "PLAT-5002", Domain: DomainPlatform,
		ConnectCode: connect.CodeFailedPrecondition, Message: "目前方案未包含此功能，請升級方案"})

	// PlatformPaymentConflict 為收款衝突（期別已付款、金額不符）。
	PlatformPaymentConflict = MustRegister(Code{ID: "PLAT-3002", Domain: DomainPlatform,
		ConnectCode: connect.CodeFailedPrecondition, Message: "收款衝突：{reason}"})
)

package errcode

import "connectrpc.com/connect"

// 系統域：通用參數、衝突、權限、不存在與 5xx。無專碼可用的錯誤才用本域碼。
var (
	// SysInvalidArgument 為無專碼可用的參數驗證失敗（新增端點時優先定義專碼）。
	SysInvalidArgument = MustRegister(Code{ID: "SYS-1001", Domain: DomainSys,
		ConnectCode: connect.CodeInvalidArgument, Message: "參數驗證失敗"})

	// SysConflict 為**已知的**識別碼重複（例：CreateCompany 以 DeletedAtIsNil 前置查詢判定後回此碼）。
	// 為什麼不由 DB 約束錯誤推導：ent 的 constraint 錯誤無法分辨「識別碼重複」與「FK 阻擋」，
	// 把後者回成 AlreadyExists 正是 P2-A 的原始缺陷（見 SysConstraintViolation）。
	SysConflict = MustRegister(Code{ID: "SYS-2001", Domain: DomainSys,
		ConnectCode: connect.CodeAlreadyExists, Message: "資料衝突，請確認識別碼是否已被使用"})

	// SysScopeViolation 為寫入被 RLS 的 WITH CHECK 擋下（資料範圍不符）。
	// 為什麼是 FailedPrecondition 而非 Internal：這是「身分／範圍與該列不匹配」，不是伺服器故障
	// （用 Internal 會讓監控誤判 5xx 並誤導客戶）；SQLSTATE 與 policy 原文只進 log。
	SysScopeViolation = MustRegister(Code{ID: "SYS-3001", Domain: DomainSys,
		ConnectCode: connect.CodeFailedPrecondition, Message: "資料超出目前的存取範圍,無法完成此操作"})

	// SysConstraintViolation 為資料庫約束類錯誤（識別碼重複、FK 阻擋、CHECK 失敗）——無法分辨是哪一種。
	// 為什麼不是 AlreadyExists：P2-A 的原始缺陷正是「FK 阻擋被當成識別碼重複回 AlreadyExists」，
	// 且 ent 給的 constraint 錯誤無法區分兩者；訊息刻意保留可行動指引但不揭露 DB 細節。
	SysConstraintViolation = MustRegister(Code{ID: "SYS-3002", Domain: DomainSys,
		ConnectCode: connect.CodeFailedPrecondition,
		Message:     "資料違反資料庫約束,無法完成此操作(請確認識別碼是否已被使用、參照對象是否仍存在)"})

	// SysPermissionDenied 為授權檢查失敗（角色／資料範圍不足）。前端應導向「請管理員開權」。
	SysPermissionDenied = MustRegister(Code{ID: "SYS-4001", Domain: DomainSys,
		ConnectCode: connect.CodePermissionDenied, Message: "缺少權限"})

	// SysNotFound 同時代表「不存在」與「不在可見範圍內」——刻意不區分，避免以錯誤碼
	// 探測他租戶資源是否存在（oracle）。跨租戶查詢因範圍過濾而查不到時一律用此碼。
	SysNotFound = MustRegister(Code{ID: "SYS-4002", Domain: DomainSys,
		ConnectCode: connect.CodeNotFound, Message: "資源不存在或無權存取"})

	// SysInternal 為所有 5xx：對外只給碼與訊息；trace_id 由 ErrorInfo 的結構化欄位承載
	// （客服回報用），**不**寫進訊息樣板（否則每個呼叫點都得先注入參數、漏了就外洩字面 {trace}）。
	SysInternal = MustRegister(Code{ID: "SYS-9000", Domain: DomainSys,
		ConnectCode: connect.CodeInternal, Message: "系統忙碌，請稍後再試"})
)

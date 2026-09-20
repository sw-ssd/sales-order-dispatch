package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/operatorauth"
)

// Operators 以 admin(owner)連線存取 platform.operators 白名單與登入稽核。
//
// 與 Store(唯讀的權益／訂閱查詢)分開:本型別**有寫入**(last_login_at、platform.audit_logs)。
// 平台 schema 對業務角色 app_rw 零權限(00029/S9),故只接受 admin 連線。
type Operators struct{ db *sql.DB }

var _ operatorauth.Store = (*Operators)(nil)

// NewOperators 建立 Operators(呼叫端負責 admin DSN)。
func NewOperators(db *sql.DB) *Operators { return &Operators{db: db} }

// OperatorByEmail 查白名單;查無此 email 回 (nil, nil) —— 「不在名單」是預期結果,不是錯誤
// (回錯誤會讓呼叫端把「陌生人嘗試登入」當成系統故障)。
//
// 比對為大小寫敏感,與 00029 的 UNIQUE 約束一致:同一個 email 只會有一列,不存在
// 「大小寫不同的兩列」而挑錯人的情況;Google 回傳的 email 亦為小寫正規形。
func (s *Operators) OperatorByEmail(ctx context.Context, email string) (*operatorauth.Operator, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, email, name, role, status
		  FROM platform.operators
		 WHERE email = $1`, email)
	var op operatorauth.Operator
	err := row.Scan(&op.ID, &op.Email, &op.Name, &op.Role, &op.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &op, nil
}

// TouchOperatorLogin 記錄最後登入時間(updated_at 一併更新,與其他寫入慣例一致)。
func (s *Operators) TouchOperatorLogin(ctx context.Context, id int64, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE platform.operators SET last_login_at = $2, updated_at = now() WHERE id = $1`, id, at)
	return err
}

// AuditOperatorLogin 寫入登入稽核(S9:actor 為 operator_id,不 FK 租戶 users)。
// target_id 為該 operator 的 id(文字欄);email 與 User-Agent 進 after —— 表上沒有 user_agent
// 欄位,但不留痕就答不出「這次登入是不是本人」。
func (s *Operators) AuditOperatorLogin(ctx context.Context, operatorID int64, email, ip, ua string) error {
	after, err := json.Marshal(map[string]string{"email": email, "user_agent": ua})
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO platform.audit_logs (operator_id, action, target_type, target_id, reason, after, ip_address)
		VALUES ($1, 'login', 'operator', $2, '', $3::jsonb, $4)`,
		operatorID, strconv.FormatInt(operatorID, 10), string(after), ip)
	return err
}

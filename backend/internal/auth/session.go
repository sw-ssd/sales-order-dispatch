package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

// Web Session 常數。
const (
	SessionCookieName = "session"
	SessionUserIDKey  = "user_id"
	SessionRoleKey    = "role"
)

// SessionStore 以 KVStore 實作 scs.Store 與 scs.CtxStore(Valkey 為正式儲存;測試用 scs memstore)。
// scs.SessionManager 偵測到 CtxStore 時一律走 context 變體(DeleteCtx / FindCtx / CommitCtx)。
type SessionStore struct {
	kv KVStore
}

// NewSessionStore 建立 SessionStore。
func NewSessionStore(kv KVStore) *SessionStore {
	return &SessionStore{kv: kv}
}

func sessionKey(token string) string {
	return "scs:session:" + token
}

// DeleteCtx 實作 scs.CtxStore.DeleteCtx(不存在為 no-op)。
func (s *SessionStore) DeleteCtx(ctx context.Context, token string) error {
	return s.kv.Delete(ctx, sessionKey(token))
}

// FindCtx 實作 scs.CtxStore.FindCtx;不存在或已過期 found=false。
func (s *SessionStore) FindCtx(ctx context.Context, token string) ([]byte, bool, error) {
	v, ok, err := s.kv.Get(ctx, sessionKey(token))
	if err != nil || !ok {
		return nil, ok, err
	}
	return []byte(v), true, nil
}

// CommitCtx 實作 scs.CtxStore.CommitCtx。
func (s *SessionStore) CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) error {
	ttl := time.Until(expiry)
	if ttl <= 0 {
		ttl = time.Minute
	}
	return s.kv.Set(ctx, sessionKey(token), string(b), ttl)
}

// 以下為 scs.Store 純函式變體(無 context),僅供型別滿足 scs.Store。

// Delete 實作 scs.Store.Delete。
func (s *SessionStore) Delete(token string) error {
	return s.kv.Delete(context.Background(), sessionKey(token))
}

// Find 實作 scs.Store.Find。
func (s *SessionStore) Find(token string) ([]byte, bool, error) {
	return s.FindCtx(context.Background(), token)
}

// Commit 實作 scs.Store.Commit。
func (s *SessionStore) Commit(token string, b []byte, expiry time.Time) error {
	return s.CommitCtx(context.Background(), token, b, expiry)
}

// WebSessionManager 建立 scs session manager(Web httpOnly cookie;D5 Web 軌)。
// sameSite 接受 lax / strict / none。
func WebSessionManager(store scs.Store, lifetime time.Duration, secure bool, sameSite string) *scs.SessionManager {
	m := scs.New()
	m.Store = store
	m.Lifetime = lifetime
	m.Cookie.Name = SessionCookieName
	m.Cookie.HttpOnly = true
	m.Cookie.Secure = secure
	m.Cookie.Persist = true
	switch sameSite {
	case "strict":
		m.Cookie.SameSite = http.SameSiteStrictMode
	case "none":
		m.Cookie.SameSite = http.SameSiteNoneMode
	default:
		m.Cookie.SameSite = http.SameSiteLaxMode
	}
	return m
}

// EstablishWebSession 於 scs session 寫入使用者身分(供 callback / RPC 呼叫;由 LoadAndSave middleware 提交)。
func EstablishWebSession(ctx context.Context, m *scs.SessionManager, userID int, role string) {
	m.Put(ctx, SessionUserIDKey, userID)
	m.Put(ctx, SessionRoleKey, role)
}

// SessionUserID 讀取 scs session 中的使用者 ID;未登入回傳 0。
func SessionUserID(ctx context.Context, m *scs.SessionManager) int {
	return m.GetInt(ctx, SessionUserIDKey)
}

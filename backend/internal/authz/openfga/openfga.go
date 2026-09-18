// Package openfga 為授權引擎對齊 D32 的 internal/authz facade。
// 包裝 third_party/openfga 的內嵌 OpenFGA client,對 authz 呼叫端提供
// Check / ListObjects / WriteTuple / DeleteTuple 的一致介面(D30 執行面退場,
// 授權決策改由 OpenFGA 提供;RLS 保留為資料庫層兜底)。
// 條件兩層分工:OpenFGA 只管關係性/角色/租戶範圍授權;物件狀態條件由 domain 狀態機處理。
package openfga

import (
	"context"
	"errors"

	ofga "github.com/salesorder/sales-order-1.0/backend/third_party/openfga"
)

// Engine 持有內嵌 OpenFGA client,為 authz 呼叫端提供授權決策介面(D32)。
// 以 context 注入(server 組裝於啟動時單一建立),測試以記憶體 datastore 建置。
type Engine struct {
	client *ofga.Client
}

// New 以既有內嵌 OpenFGA client 建立 Engine facade。
func New(client *ofga.Client) *Engine {
	return &Engine{client: client}
}

// Client 回傳底層 client(供 store id/model id 等進階操作)。
func (e *Engine) Client() *ofga.Client {
	return e.client
}

// Check 判斷 user 是否具 relation 於 object(OpenFGA 內嵌 Check)。
// user/object 為 OpenFGA tuple 語法(如 "user:u1"、"company:c1"、"role#assigned")。
func (e *Engine) Check(ctx context.Context, user, relation, object string) (bool, error) {
	if e == nil || e.client == nil {
		return false, errors.New("authz/openfga: engine 未初始化")
	}
	return e.client.Check(ctx, user, relation, object)
}

// ListObjects 回傳 user 具 relation 於指定 objectType 的所有 object id。
func (e *Engine) ListObjects(ctx context.Context, user, relation, objectType string) ([]string, error) {
	if e == nil || e.client == nil {
		return nil, errors.New("authz/openfga: engine 未初始化")
	}
	return e.client.ListObjects(ctx, user, relation, objectType)
}

// WriteTuple 寫入一筆授權 tuple(user relation object)。
func (e *Engine) WriteTuple(ctx context.Context, user, relation, object string) error {
	if e == nil || e.client == nil {
		return errors.New("authz/openfga: engine 未初始化")
	}
	return e.client.WriteTuple(ctx, user, relation, object)
}

// DeleteTuple 刪除一筆授權 tuple。
func (e *Engine) DeleteTuple(ctx context.Context, user, relation, object string) error {
	if e == nil || e.client == nil {
		return errors.New("authz/openfga: engine 未初始化")
	}
	return e.client.DeleteTuple(ctx, user, relation, object)
}

// ListTuples 列舉 store 全部 tuples(供 reconcile 差異比對)。回傳 (user, relation, object) 清單。
func (e *Engine) ListTuples(ctx context.Context) ([][3]string, error) {
	if e == nil || e.client == nil {
		return nil, errors.New("authz/openfga: engine 未初始化")
	}
	return e.client.ListTuples(ctx)
}

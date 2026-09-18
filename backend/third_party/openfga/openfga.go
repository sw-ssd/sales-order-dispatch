// Package openfga 以 OpenFGA Go library 內嵌授權引擎(D32)。
// 於後端程序內起一個 in-process OpenFGA server,datastore 共用 PostgreSQL(單一 store);
// 測試以記憶體 datastore。授權決策 Check / ListObjects 由此封裝提供。
package openfga

import (
	"context"
	"fmt"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/language/pkg/go/transformer"
	"github.com/openfga/openfga/pkg/server"
	"github.com/openfga/openfga/pkg/storage"
	"github.com/openfga/openfga/pkg/storage/memory"
	"github.com/openfga/openfga/pkg/storage/postgres"
	"github.com/openfga/openfga/pkg/storage/sqlcommon"
)

// SchemaVersion 對齊 model.fga 的 schema 版本。
const SchemaVersion = "1.1"

// Client 包裝內嵌 OpenFGA server 與 store id/model id,提供授權決策介面。
type Client struct {
	srv     *server.Server
	StoreID string
	ModelID string
}

// NewPostgres 以 PostgreSQL datastore 起內嵌 server 並建立 store(生產單一 store)。
func NewPostgres(ctx context.Context, dsn, storeName string) (*Client, error) {
	ds, err := postgres.New(dsn, &sqlcommon.Config{})
	if err != nil {
		return nil, fmt.Errorf("openfga: postgres datastore: %w", err)
	}
	return newClient(ctx, ds, storeName)
}

// NewMemory 以記憶體 datastore 起內嵌 server(測試用)。
func NewMemory(ctx context.Context, storeName string) (*Client, error) {
	return newClient(ctx, memory.New(), storeName)
}

func newClient(ctx context.Context, ds storage.OpenFGADatastore, storeName string) (*Client, error) {
	srv, err := server.NewServerWithOpts(server.WithDatastore(ds))
	if err != nil {
		return nil, fmt.Errorf("openfga: server 建置失敗: %w", err)
	}
	// 依名稱複用既有 store(D32 單一 store):僅在不存在時才建立,避免每次重啟重建 store
	// 使已寫入的 tuples/model 消失。ListStores 依 name 過濾回傳相符 store。
	existing, err := findStoreByName(ctx, srv, storeName)
	if err != nil {
		return nil, err
	}
	if existing == "" {
		created, err := srv.CreateStore(ctx, &openfgav1.CreateStoreRequest{Name: storeName})
		if err != nil {
			return nil, fmt.Errorf("openfga: create store: %w", err)
		}
		existing = created.GetId()
	}
	c := &Client{srv: srv, StoreID: existing}
	// 僅在 store 尚未有 authorization model 時寫入預設 model(避免每次重啟覆寫/更替 model id)。
	if err := c.EnsureDefaultModel(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

// findStoreByName 依名稱查詢既有 store;不存在回傳空字串。ListStores 分頁,逐頁掃描。
func findStoreByName(ctx context.Context, srv *server.Server, name string) (string, error) {
	token := ""
	for {
		resp, err := srv.ListStores(ctx, &openfgav1.ListStoresRequest{Name: name, ContinuationToken: token})
		if err != nil {
			return "", fmt.Errorf("openfga: list stores: %w", err)
		}
		for _, st := range resp.GetStores() {
			if st.GetName() == name {
				return st.GetId(), nil
			}
		}
		token = resp.GetContinuationToken()
		if token == "" {
			return "", nil
		}
	}
}

// WriteDefaultModel 將 modelDSL 寫為 store 的 authorization model,並記錄 ModelID。
func (c *Client) WriteDefaultModel(ctx context.Context) error {
	model, err := transformer.TransformDSLToProto(modelDSL)
	if err != nil {
		return fmt.Errorf("openfga: model DSL 解析失敗: %w", err)
	}
	resp, err := c.srv.WriteAuthorizationModel(ctx, &openfgav1.WriteAuthorizationModelRequest{
		StoreId:         c.StoreID,
		SchemaVersion:   SchemaVersion,
		TypeDefinitions: model.GetTypeDefinitions(),
		Conditions:      model.GetConditions(),
	})
	if err != nil {
		return fmt.Errorf("openfga: write model: %w", err)
	}
	c.ModelID = resp.GetAuthorizationModelId()
	return nil
}

// EnsureDefaultModel 於 store 尚未有 authorization model 時寫入預設 model,並記錄 ModelID。
// 已存在 model 時僅選取最新 model id,避免重啟後每次覆寫、造成 store/model id 漂移。
func (c *Client) EnsureDefaultModel(ctx context.Context) error {
	list, err := c.srv.ReadAuthorizationModels(ctx, &openfgav1.ReadAuthorizationModelsRequest{StoreId: c.StoreID})
	if err != nil {
		return fmt.Errorf("openfga: list models: %w", err)
	}
	if len(list.GetAuthorizationModels()) > 0 {
		// 取現有 model 之一記錄(Check/ListObjects 未傳 AuthorizationModelId,OpenFGA 自用最新)。
		c.ModelID = list.GetAuthorizationModels()[0].GetId()
		return nil
	}
	return c.WriteDefaultModel(ctx)
}

// Close 關閉內嵌 server 資源。
func (c *Client) Close() {
	if c.srv != nil {
		c.srv.Close()
	}
}

// Check 判斷 user 是否具 relation 於 object 上。
func (c *Client) Check(ctx context.Context, user, relation, object string) (bool, error) {
	resp, err := c.srv.Check(ctx, &openfgav1.CheckRequest{
		StoreId: c.StoreID,
		TupleKey: &openfgav1.CheckRequestTupleKey{
			User:     user,
			Relation: relation,
			Object:   object,
		},
	})
	if err != nil {
		return false, err
	}
	return resp.GetAllowed(), nil
}

// ListObjects 回傳 user 具 relation 於指定 type 上的所有 object id。
func (c *Client) ListObjects(ctx context.Context, user, relation, objectType string) ([]string, error) {
	resp, err := c.srv.ListObjects(ctx, &openfgav1.ListObjectsRequest{
		StoreId:  c.StoreID,
		Type:     objectType,
		Relation: relation,
		User:     user,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetObjects(), nil
}

// WriteTuple 寫入一筆授權 tuple(user relation object)。
func (c *Client) WriteTuple(ctx context.Context, user, relation, object string) error {
	_, err := c.srv.Write(ctx, &openfgav1.WriteRequest{
		StoreId: c.StoreID,
		Writes: &openfgav1.WriteRequestWrites{TupleKeys: []*openfgav1.TupleKey{
			{User: user, Relation: relation, Object: object},
		}},
	})
	return err
}

// DeleteTuple 刪除一筆授權 tuple。
func (c *Client) DeleteTuple(ctx context.Context, user, relation, object string) error {
	_, err := c.srv.Write(ctx, &openfgav1.WriteRequest{
		StoreId: c.StoreID,
		Deletes: &openfgav1.WriteRequestDeletes{TupleKeys: []*openfgav1.TupleKeyWithoutCondition{
			{User: user, Relation: relation, Object: object},
		}},
	})
	return err
}

// ListTuples 列舉 store 全部 tuples(分頁全量),供 reconcile 差異比對。
// OpenFGA Read 於僅給 user 時要求 object type,故此處以空 filter 全量讀取。
func (c *Client) ListTuples(ctx context.Context) ([][3]string, error) {
	var out [][3]string
	token := ""
	for {
		resp, err := c.srv.Read(ctx, &openfgav1.ReadRequest{
			StoreId:           c.StoreID,
			ContinuationToken: token,
		})
		if err != nil {
			return nil, fmt.Errorf("openfga: read tuples: %w", err)
		}
		for _, t := range resp.GetTuples() {
			k := t.GetKey()
			if k == nil {
				continue
			}
			out = append(out, [3]string{k.GetUser(), k.GetRelation(), k.GetObject()})
		}
		token = resp.GetContinuationToken()
		if token == "" {
			return out, nil
		}
	}
}

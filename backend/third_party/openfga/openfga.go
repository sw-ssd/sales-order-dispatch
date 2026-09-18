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
	store, err := srv.CreateStore(ctx, &openfgav1.CreateStoreRequest{Name: storeName})
	if err != nil {
		return nil, fmt.Errorf("openfga: create store: %w", err)
	}
	c := &Client{srv: srv, StoreID: store.GetId()}
	if err := c.WriteDefaultModel(ctx); err != nil {
		return nil, err
	}
	return c, nil
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

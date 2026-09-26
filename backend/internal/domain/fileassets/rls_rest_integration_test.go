//go:build integration

package fileassets_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/fileassets"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// appRoleFileClient 以 **app_rw**(非 superuser,RLS 對其生效)建立業務 client。
// 必須經 dbtenant.NewClient,RLS 裝飾器才會在 Tx(ctx) 內套用 SET LOCAL。
// 池上限 4:請求交易之外若還有第二條交易(本套件不該有)才不會互鎖成死結。
func appRoleFileClient(t *testing.T, adminDSN string) *ent.Client {
	t.Helper()
	pool, err := sql.Open("pgx", testsupport.AppRoleDSN(t, adminDSN))
	if err != nil {
		t.Fatalf("app_rw 連線: %v", err)
	}
	pool.SetMaxOpenConns(4)
	client := dbtenant.NewClient(pool)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// TestIntegrationFileEndpointsUnderAppRole 以 **app_rw 連線**驅動真實 handler,驗證 REST 路徑
// 真的落在租戶交易內。
//
// 為什麼需要這一條:file_assets / companies 是 ENABLE + FORCE RLS,而業務連線是非 superuser ——
// **交易外**的查詢一律回 0 列(不是報錯,是靜默「查無」)。在 superuser 連線下的整合測試完全看不出來
// (superuser 繞過 RLS),實際症狀只在 smoke 時現形:上傳成功、DB 有列,下載卻 404;通用上傳的
// owner 驗證也誤判「owner 不存在」。此測試即為那個情境的回歸鎖。
//
// 覆蓋:上傳(含 owner 驗證)→ 以回傳 url 下載 → 以數字 id 下載 → 軟刪除後下載 404。
func TestIntegrationFileEndpointsUnderAppRole(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUpFile(t, dsn)
	_, adminDB := openPGEntClientFile(t, dsn)
	ctx := context.Background()
	coID, deptID, actorID := seedFileCompany(t, ctx, adminDB)
	root := t.TempDir()

	// 業務 client 走 app_rw;fixture 仍以 admin(owner)寫入。
	client := appRoleFileClient(t, dsn)
	id := authz.Identity{UserID: itoaFile(actorID), CompanyID: itoaFile(coID),
		DepartmentID: itoaFile(deptID), Role: "dept_admin", Roles: []string{"dept_admin"}}
	// 身分注入**必須連同 RLS scope**——production 的 authzMiddleware 兩者都做(server.go):
	// 只注入身分會讓驅動裝飾器沒有 SET LOCAL 可用,交易內照樣 0 列。
	scope := auth.RLSScope{UserID: id.UserID, CompanyID: id.CompanyID,
		DepartmentID: id.DepartmentID, DataScope: auth.ScopeForRole(id.Role), CompanyActive: true}
	mux := http.NewServeMux()
	fileassets.NewHandler(client, root).RegisterRoutes(mux)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		mux.ServeHTTP(w, r.WithContext(auth.WithRLS(ctx, scope)))
	}))
	defer srv.Close()

	png := append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, bytes.Repeat([]byte{0x01}, 64)...)
	req := uploadHelper(t, srv.URL+"/files", "a.png", "image/png", "company", itoaFile(coID), png)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("上傳: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	// owner 驗證必須看得到公司列(app_rw 下交易外查詢會 0 列 → 這裡會拿到 400)。
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("app_rw 上傳應 201(owner 驗證須在租戶交易內),got %d body=%s", resp.StatusCode, raw)
	}
	var out struct {
		ID  int    `json:"id"`
		URL string `json:"url"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("解回應: %v,body=%s", err, raw)
	}

	// 以回傳 url 下載(測試 mux 無 StripPrefix,剝掉對外前綴)。
	byURL, err := http.Get(srv.URL + strings.TrimPrefix(out.URL, "/api/v1"))
	if err != nil {
		t.Fatalf("以 url 下載: %v", err)
	}
	defer func() { _ = byURL.Body.Close() }()
	if byURL.StatusCode != http.StatusOK {
		t.Fatalf("app_rw 下以回傳 url 下載應 200(下載查詢須在租戶交易內),got %d", byURL.StatusCode)
	}

	// 數字 id 形狀(列印 API 的 download_url)。
	byID, err := http.Get(srv.URL + "/files/" + itoaFile(out.ID) + "/download")
	if err != nil {
		t.Fatalf("以 id 下載: %v", err)
	}
	defer func() { _ = byID.Body.Close() }()
	if byID.StatusCode != http.StatusOK {
		t.Fatalf("app_rw 下以數字 id 下載應 200,got %d", byID.StatusCode)
	}

	// 軟刪除後下載拒絕(同一條 RLS 交易路徑上)。
	delReq, _ := http.NewRequest(http.MethodDelete, srv.URL+"/files/"+itoaFile(out.ID), nil)
	delResp, err := http.DefaultClient.Do(delReq)
	if err != nil {
		t.Fatalf("刪除: %v", err)
	}
	defer func() { _ = delResp.Body.Close() }()
	if delResp.StatusCode != http.StatusOK {
		t.Fatalf("app_rw 下軟刪除應 200,got %d", delResp.StatusCode)
	}
	after, err := http.Get(srv.URL + "/files/" + itoaFile(out.ID) + "/download")
	if err != nil {
		t.Fatalf("刪後下載: %v", err)
	}
	defer func() { _ = after.Body.Close() }()
	if after.StatusCode != http.StatusNotFound {
		t.Fatalf("軟刪除後下載應 404,got %d", after.StatusCode)
	}
}

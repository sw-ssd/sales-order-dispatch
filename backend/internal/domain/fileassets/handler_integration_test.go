//go:build integration

package fileassets_test

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/fileassets"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// newFileMux 以 dept_admin 身分掛檔案路由(走 authzMiddleware 同形的身分注入;
// 交易由 handler 內手動開,此處只注入身分)。
func newFileMux(t *testing.T, db *ent.Client, root string, id authz.Identity) http.Handler {
	t.Helper()
	mux := http.NewServeMux()
	fileassets.NewHandler(db, root).RegisterRoutes(mux)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
}

// seedFileCompany 建公司+部門+操作者,回 coID/deptID/actorID。
func seedFileCompany(t *testing.T, ctx context.Context, db *ent.Client) (int, int, int) {
	t.Helper()
	var coID, deptID, actorID int
	seedTxFile(t, db, func(tx *ent.Tx) error {
		co, err := tx.Company.Create().SetName("檔案公司").SetIdentifier("F-" + t.Name()).
			SetStatus("active").Save(ctx)
		if err != nil {
			return err
		}
		d, err := tx.Department.Create().SetCompanyID(co.ID).SetName("門市一").Save(ctx)
		if err != nil {
			return err
		}
		op, err := tx.User.Create().SetCompanyID(co.ID).SetDepartmentID(d.ID).
			SetEmail("op-" + t.Name() + "@t.com").SetName("操作員").
			SetRole("dept_admin").SetPasswordHash("x").Save(ctx)
		if err != nil {
			return err
		}
		coID, deptID, actorID = co.ID, d.ID, op.ID
		return nil
	})
	return coID, deptID, actorID
}

func seedTxFile(t *testing.T, db *ent.Client, fn func(tx *ent.Tx) error) {
	t.Helper()
	if err := dbtenant.SystemScopeTx(t.Context(), db, fn); err != nil {
		t.Fatalf("fixture 系統範圍寫入失敗: %v", err)
	}
}

// uploadHelper 組 multipart 上傳請求。
func uploadHelper(t *testing.T, url, filename, mimeType, ownerType, ownerID string, content []byte) *http.Request {
	t.Helper()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("建表單檔: %v", err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatalf("寫表單檔: %v", err)
	}
	_ = w.WriteField("owner_type", ownerType)
	_ = w.WriteField("owner_id", ownerID)
	_ = w.WriteField("mime_type", mimeType)
	if err := w.Close(); err != nil {
		t.Fatalf("關表單: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, &b)
	if err != nil {
		t.Fatalf("建請求: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

// TestIntegrationFileUploadDownload 上傳→下載→跨公司拒絕→軟刪除後拒絕,全鏈真 PG + 真磁碟。
func TestIntegrationFileUploadDownload(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUpFile(t, dsn)
	_, db := openPGEntClientFile(t, dsn)
	ctx := context.Background()
	coID, deptID, actorID := seedFileCompany(t, ctx, db)
	root := t.TempDir()

	id := authz.Identity{UserID: itoaFile(actorID), CompanyID: itoaFile(coID),
		DepartmentID: itoaFile(deptID), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
	srv := httptest.NewServer(newFileMux(t, db, root, id))
	defer srv.Close()

	png := append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, bytes.Repeat([]byte{0x01}, 64)...)
	// owner 為公司本身。
	req := uploadHelper(t, srv.URL+"/files", "a.png", "image/png", "company", itoaFile(coID), png)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("上傳: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("上傳應 201,got %d", resp.StatusCode)
	}
	// 解析回應取 id 與 url。
	var out struct {
		ID  int    `json:"id"`
		URL string `json:"url"`
	}
	if err := jsonDecode(resp, &out); err != nil {
		t.Fatalf("解回應: %v", err)
	}
	// 檔案確實落盤(相對路徑拼 root)。
	matches, _ := filepath.Glob(filepath.Join(root, "*", "*", "*", "*"))
	if len(matches) != 1 {
		t.Fatalf("應恰落盤 1 檔,got %v", matches)
	}
	// 下載:同公司可得正確 MIME。
	dl, err := http.Get(srv.URL + "/files/" + itoaFile(out.ID) + "/download")
	if err != nil {
		t.Fatalf("下載: %v", err)
	}
	defer dl.Body.Close()
	if dl.StatusCode != http.StatusOK || dl.Header.Get("Content-Type") != "image/png" {
		t.Fatalf("下載應 200 + image/png,got %d %q", dl.StatusCode, dl.Header.Get("Content-Type"))
	}
	if cd := dl.Header.Get("Content-Disposition"); cd == "" {
		t.Fatal("下載應帶 Content-Disposition")
	}

	// 跨公司拒絕:另公司身分下載同一 id → 404。
	coID2 := seedSecondCompany(t, ctx, db)
	id2 := authz.Identity{UserID: itoaFile(actorID), CompanyID: itoaFile(coID2),
		Role: "company_admin", Roles: []string{"company_admin"}}
	srv2 := httptest.NewServer(newFileMux(t, db, root, id2))
	defer srv2.Close()
	dl2, err := http.Get(srv2.URL + "/files/" + itoaFile(out.ID) + "/download")
	if err != nil {
		t.Fatalf("跨公司下載: %v", err)
	}
	defer dl2.Body.Close()
	if dl2.StatusCode != http.StatusNotFound {
		t.Fatalf("跨公司應 404,got %d", dl2.StatusCode)
	}

	// 偽裝檔被拒(exe 改名 jpg)。
	reqBad := uploadHelper(t, srv.URL+"/files", "evil.jpg", "image/jpeg", "company", itoaFile(coID),
		[]byte("MZ........"))
	respBad, err := http.DefaultClient.Do(reqBad)
	if err != nil {
		t.Fatalf("偽裝上傳: %v", err)
	}
	defer respBad.Body.Close()
	if respBad.StatusCode != http.StatusBadRequest {
		t.Fatalf("偽裝檔應 400,got %d", respBad.StatusCode)
	}

	// 軟刪除後下載拒絕。
	delReq, _ := http.NewRequest(http.MethodDelete, srv.URL+"/files/"+itoaFile(out.ID), nil)
	delResp, err := http.DefaultClient.Do(delReq)
	if err != nil {
		t.Fatalf("刪除: %v", err)
	}
	defer delResp.Body.Close()
	if delResp.StatusCode != http.StatusOK {
		t.Fatalf("刪除應 200,got %d", delResp.StatusCode)
	}
	dl3, err := http.Get(srv.URL + "/files/" + itoaFile(out.ID) + "/download")
	if err != nil {
		t.Fatalf("刪後下載: %v", err)
	}
	defer dl3.Body.Close()
	if dl3.StatusCode != http.StatusNotFound {
		t.Fatalf("刪後下載應 404,got %d", dl3.StatusCode)
	}
	// 實體檔案保留(稽核用)。
	if _, err := os.Stat(matches[0]); err != nil {
		t.Fatalf("軟刪除應保留實體檔: %v", err)
	}
}

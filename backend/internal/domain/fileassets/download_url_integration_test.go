//go:build integration

package fileassets_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationSavedURLIsDownloadable 守住「DB 裡的 url 指得到東西」這個契約:
// file_assets.url 是**系統檔名**形狀(/api/v1/files/<uuid>.<ext>/download,SaveUpload 與
// print_helpers 都這樣寫),列印 API 另以數字 id 組 download_url —— 兩種形狀都必須服務得到。
//
// 這個測試若退化成「斷言 url 字串等於某值」就會失去意義:它必須把該 url **實際發一次請求**,
// 因為原本的 bug 正是 url 字串自洽、卻沒有任何路由能解析它(下載路由只吃整數 id → 一律 404)。
func TestIntegrationSavedURLIsDownloadable(t *testing.T) {
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
	req := uploadHelper(t, srv.URL+"/files", "logo.png", "image/png", "company", itoaFile(coID), png)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("上傳: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("上傳應 201,got %d", resp.StatusCode)
	}
	var out struct {
		ID  int    `json:"id"`
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("解回應: %v", err)
	}

	// 上傳回應帶回的 url 是**對外**路徑(/api/v1/...);測試 mux 未掛 StripPrefix(比照 handler
	// 在 apiMux 上的相對路徑),故比對前先剝掉前綴——production 由 domains.go 的
	// http.StripPrefix("/api/v1", …) 負責,登錄路徑本身是相對的。
	rel := strings.TrimPrefix(out.URL, "/api/v1")
	byURL, err := http.Get(srv.URL + rel)
	if err != nil {
		t.Fatalf("以 url 下載: %v", err)
	}
	defer byURL.Body.Close()
	if byURL.StatusCode != http.StatusOK || byURL.Header.Get("Content-Type") != "image/png" {
		t.Fatalf("以 DB 記錄的 url 下載應 200 + image/png,got %d %q(url=%s)",
			byURL.StatusCode, byURL.Header.Get("Content-Type"), out.URL)
	}

	// 數字 id 形狀(列印 API 的 download_url)維持可用。
	byID, err := http.Get(srv.URL + "/files/" + itoaFile(out.ID) + "/download")
	if err != nil {
		t.Fatalf("以 id 下載: %v", err)
	}
	defer byID.Body.Close()
	if byID.StatusCode != http.StatusOK {
		t.Fatalf("以數字 id 下載應 200,got %d", byID.StatusCode)
	}

	// 未知檔名不得變成「查無即回第一筆」之類的寬鬆查詢。
	missing, err := http.Get(srv.URL + "/files/deadbeef.png/download")
	if err != nil {
		t.Fatalf("未知檔名下載: %v", err)
	}
	defer missing.Body.Close()
	if missing.StatusCode != http.StatusNotFound {
		t.Fatalf("未知檔名應 404,got %d", missing.StatusCode)
	}
}

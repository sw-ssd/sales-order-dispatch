//go:build integration

package fileassets_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/ent/auditlog"
	"github.com/salesorder/sales-order-1.0/backend/ent/fileasset"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// logoIdentity 組指定角色的身分(同 newFileMux 的注入形狀)。
func logoIdentity(actorID, coID, deptID int, role string) authz.Identity {
	return authz.Identity{
		UserID: itoaFile(actorID), CompanyID: itoaFile(coID), DepartmentID: itoaFile(deptID),
		Role: role, Roles: []string{role},
	}
}

// TestIntegrationCompanyLogoUpload 細部 2.4.1 驗收(權限:上傳者為 company_admin,見 spec 3.1.1):
// 未登入 401、super 403、非圖片副檔名 400、**他公司 404**(存在與不存在皆同形)、
// company_admin 成功(logo_url 更新＋檔記錄＋雙稽核＋回傳 url 可下載)、軟刪公司 404。
// 真 PG + 真磁碟。
func TestIntegrationCompanyLogoUpload(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUpFile(t, dsn)
	_, db := openPGEntClientFile(t, dsn)
	ctx := context.Background()
	coID, deptID, actorID := seedFileCompany(t, ctx, db)
	otherCoID := seedSecondCompany(t, ctx, db)
	root := t.TempDir()

	png := append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, bytes.Repeat([]byte{0x01}, 64)...)
	logoPath := "/companies/" + itoaFile(coID) + "/logo"

	// post 以給定身分打 Logo 端點,回狀態碼與 body(讀畢即關,調用端不須 Close)。
	post := func(id authz.Identity, filename, mime string, content []byte) (int, []byte) {
		t.Helper()
		srv := httptest.NewServer(newFileMux(t, db, root, id))
		defer srv.Close()
		req := uploadHelper(t, srv.URL+logoPath, filename, mime, "company", itoaFile(coID), content)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("上傳請求: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("讀回應: %v", err)
		}
		return resp.StatusCode, b
	}

	// 未登入(Roles 空)→ 401。
	if code, _ := post(authz.Identity{}, "logo.png", "image/png", png); code != http.StatusUnauthorized {
		t.Fatalf("未登入應 401,got %d", code)
	}
	// super 不經此端點(權限為 company_admin 專屬)→ 403,logo_url 不動。
	if code, _ := post(logoIdentity(actorID, coID, deptID, "super"), "logo.png", "image/png", png); code != http.StatusForbidden {
		t.Fatalf("super 應 403,got %d", code)
	}
	if co, err := db.Company.Get(ctx, coID); err != nil || co.LogoURL != "" {
		t.Fatalf("403 後 logo_url 應保持空,err=%v", err)
	}
	// dept_admin/staff 亦 403(非公司層管理角色)。
	if code, _ := post(logoIdentity(actorID, coID, deptID, "dept_admin"), "logo.png", "image/png", png); code != http.StatusForbidden {
		t.Fatalf("dept_admin 應 403,got %d", code)
	}
	// PDF 副檔名 → 400(Logo 只收圖片),且不建檔。
	if code, _ := post(logoIdentity(actorID, coID, deptID, "company_admin"), "logo.pdf", "application/pdf", []byte("%PDF-1.4 fake")); code != http.StatusBadRequest {
		t.Fatalf("PDF 副檔名應 400,got %d", code)
	}
	if n, err := db.FileAsset.Query().Where(fileasset.OwnerType("company")).Count(ctx); err != nil || n != 0 {
		t.Fatalf("拒絕路徑不應建檔,n=%d err=%v", n, err)
	}
	// 偽裝副檔名(exe 改名 png)→ 400,由 magic bytes 擋下。
	if code, _ := post(logoIdentity(actorID, coID, deptID, "company_admin"), "evil.png", "image/png", []byte("MZ........")); code != http.StatusBadRequest {
		t.Fatalf("偽裝副檔名應 400,got %d", code)
	}

	// 他公司 → 404。**存在**的他公司與**不存在**的公司同形,才證明是範圍判斷而非查無。
	for _, tc := range []struct{ name, id string }{
		{"存在的他公司", itoaFile(otherCoID)},
		{"不存在的公司", itoaFile(coID + 9999)},
	} {
		srv := httptest.NewServer(newFileMux(t, db, root, logoIdentity(actorID, coID, deptID, "company_admin")))
		req := uploadHelper(t, srv.URL+"/companies/"+tc.id+"/logo", "logo.png", "image/png", "company", tc.id, png)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			srv.Close()
			t.Fatalf("%s 請求: %v", tc.name, err)
		}
		code := resp.StatusCode
		_ = resp.Body.Close()
		srv.Close()
		if code != http.StatusNotFound {
			t.Fatalf("%s 應 404,got %d", tc.name, code)
		}
	}
	if co, err := db.Company.Get(ctx, otherCoID); err != nil || co.LogoURL != "" {
		t.Fatalf("他公司 logo_url 不得被寫入,err=%v", err)
	}

	// company_admin 成功:201 + 回傳 url + logo_url 更新 + 檔記錄 + 兩種稽核。
	code, raw := post(logoIdentity(actorID, coID, deptID, "company_admin"), "logo.png", "image/png", png)
	if code != http.StatusCreated {
		t.Fatalf("company_admin 上傳應 201,got %d,body=%s", code, raw)
	}
	var body struct {
		ID  int    `json:"id"`
		URL string `json:"url"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("解譯回應: %v,body=%s", err, raw)
	}
	if body.URL == "" || body.ID <= 0 {
		t.Fatalf("回應缺 file_asset_id/url: %+v", body)
	}
	co, err := db.Company.Get(ctx, coID)
	if err != nil {
		t.Fatalf("讀公司: %v", err)
	}
	if co.LogoURL != body.URL {
		t.Fatalf("companies.logo_url 應等於回傳 url,got %q want %q", co.LogoURL, body.URL)
	}
	if n, err := db.FileAsset.Query().Where(
		fileasset.OwnerType("company"), fileasset.OwnerID(coID), fileasset.ID(body.ID),
	).Count(ctx); err != nil || n != 1 {
		t.Fatalf("應有 1 筆 company 檔記錄,n=%d err=%v", n, err)
	}
	// 回傳的 url 必須真的下載得到(細部 2.4.1 驗收:「可經 url 下載」)。測試 mux 未掛
	// StripPrefix,故先剝掉對外前綴(production 由 domains.go 的 http.StripPrefix 負責)。
	srv := httptest.NewServer(newFileMux(t, db, root, logoIdentity(actorID, coID, deptID, "company_admin")))
	defer srv.Close()
	dl, err := http.Get(srv.URL + strings.TrimPrefix(body.URL, "/api/v1"))
	if err != nil {
		t.Fatalf("以回傳 url 下載: %v", err)
	}
	defer func() { _ = dl.Body.Close() }()
	if dl.StatusCode != http.StatusOK || dl.Header.Get("Content-Type") != "image/png" {
		t.Fatalf("回傳 url 應可下載(200 + image/png),got %d %q", dl.StatusCode, dl.Header.Get("Content-Type"))
	}
	// 稽核:檔上傳 create(SaveUpload)+ 公司 update(logo_url 變更)。
	if n, err := db.AuditLog.Query().Where(
		auditlog.ResourceTypeEQ("company"), auditlog.ActionEQ(auditlog.ActionUpdate),
	).Count(ctx); err != nil || n != 1 {
		t.Fatalf("公司 update 稽核應 1 筆,n=%d err=%v", n, err)
	}
	if n, err := db.AuditLog.Query().Where(
		auditlog.ResourceTypeEQ("file_asset"), auditlog.ActionEQ(auditlog.ActionCreate),
	).Count(ctx); err != nil || n != 1 {
		t.Fatalf("檔 create 稽核應 1 筆,n=%d err=%v", n, err)
	}
	// 軟刪除公司 → 404(spec 2.4.1:不存在或已軟刪一律 not_found),且不建檔。
	if _, err := db.Company.UpdateOneID(coID).SetDeletedAt(time.Now()).Save(ctx); err != nil {
		t.Fatalf("軟刪公司: %v", err)
	}
	if code, _ := post(logoIdentity(actorID, coID, deptID, "company_admin"), "logo.png", "image/png", png); code != http.StatusNotFound {
		t.Fatalf("軟刪公司應 404,got %d", code)
	}
	if n, err := db.FileAsset.Query().Where(fileasset.OwnerType("company")).Count(ctx); err != nil || n != 1 {
		t.Fatalf("404 路徑不應建檔,n=%d err=%v", n, err)
	}
}

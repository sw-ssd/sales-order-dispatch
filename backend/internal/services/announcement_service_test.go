package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"
	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/announcement"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
)

// annDBSeq 讓每個測試拿到獨立的記憶體庫(同 dsn 會共用同一顆,seed 會互相汙染)。
var annDBSeq atomic.Int64

func openAnnDB(t *testing.T) *ent.Client {
	t.Helper()
	dsn := fmt.Sprintf("file:ann%d?mode=memory&cache=shared&_fk=1", annDBSeq.Add(1))
	db := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// newAnnouncementClient 把 AnnouncementService 掛到 httptest,注入指定身分與 DB。
func newAnnouncementClient(t *testing.T, db *ent.Client, id authz.Identity) salesorderv1connect.AnnouncementServiceClient {
	t.Helper()
	mux := http.NewServeMux()
	RegisterAnnouncementService(mux, db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithDB(ctx, db)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return salesorderv1connect.NewAnnouncementServiceClient(http.DefaultClient, ts.URL)
}

// annIdentity 造指定角色的身分(Roles 走內建繼承展開,與生產 middleware 同構)。
func annIdentity(role, companyID, deptID string) authz.Identity {
	return authz.Identity{
		UserID: "1", CompanyID: companyID, DepartmentID: deptID,
		Role: role, Roles: auth.RolesFor(role),
	}
}

// annSeed 兩家公司與各自部門,回傳 (coA, coB, deptA, deptB)。
func annSeed(t *testing.T, db *ent.Client) (int, int, int, int) {
	t.Helper()
	ctx := context.Background()
	coA := db.Company.Create().SetName("公告甲公司").SetIdentifier("ANN-A").SetStatus("active").SaveX(ctx)
	coB := db.Company.Create().SetName("公告乙公司").SetIdentifier("ANN-B").SetStatus("active").SaveX(ctx)
	deptA := db.Department.Create().SetName("甲部門").SetCompanyID(coA.ID).SaveX(ctx)
	deptB := db.Department.Create().SetName("乙部門").SetCompanyID(coB.ID).SaveX(ctx)
	return coA.ID, coB.ID, deptA.ID, deptB.ID
}

// TestAnnouncementScopeMatrix spec「管理權限依範圍分層」:super 全範圍、
// company_admin 限自己公司(含該公司任一部門層)、dept_admin 僅本部門層、其餘一律拒絕;
// 非法 type / 空標題 → invalid_argument。
func TestAnnouncementScopeMatrix(t *testing.T) {
	ctx := context.Background()

	// 成功路徑各以獨立 db 跑(身分要引用 seed 出來的 id)。
	t.Run("super 全範圍", func(t *testing.T) {
		db := openAnnDB(t)
		cc := newAnnouncementClient(t, db, annIdentity("super", "", ""))
		if _, err := cc.CreateAnnouncement(ctx, connect.NewRequest(&v1.CreateAnnouncementRequest{
			Type: "banner", Title: "全系統公告",
		})); err != nil {
			t.Fatalf("super 建全系統公告應成功: %v", err)
		}
	})
	t.Run("company_admin 自己公司含部門層", func(t *testing.T) {
		db := openAnnDB(t)
		coA, _, deptA, _ := annSeed(t, db)
		cc := newAnnouncementClient(t, db, annIdentity("company_admin", uItoa(coA), ""))
		for _, req := range []*v1.CreateAnnouncementRequest{
			{Type: "news", Title: "公司層", CompanyId: uItoa(coA)},
			{Type: "news", Title: "部門層", CompanyId: uItoa(coA), DepartmentId: uItoa(deptA)},
		} {
			if _, err := cc.CreateAnnouncement(ctx, connect.NewRequest(req)); err != nil {
				t.Fatalf("company_admin 建 %q 應成功: %v", req.Title, err)
			}
		}
	})
	t.Run("dept_admin 本部門層", func(t *testing.T) {
		db := openAnnDB(t)
		coA, _, deptA, _ := annSeed(t, db)
		cc := newAnnouncementClient(t, db, annIdentity("dept_admin", uItoa(coA), uItoa(deptA)))
		if _, err := cc.CreateAnnouncement(ctx, connect.NewRequest(&v1.CreateAnnouncementRequest{
			Type: "news", Title: "部門公告", CompanyId: uItoa(coA), DepartmentId: uItoa(deptA),
		})); err != nil {
			t.Fatalf("dept_admin 建本部門公告應成功: %v", err)
		}
	})
	t.Run("空欄位自動歸屬", func(t *testing.T) {
		db := openAnnDB(t)
		coA, _, deptA, _ := annSeed(t, db)
		// company_admin 完全不帶範圍 → 公司層(company=自己、dept 空)。
		admin := newAnnouncementClient(t, db, annIdentity("company_admin", uItoa(coA), ""))
		if _, err := admin.CreateAnnouncement(ctx, connect.NewRequest(&v1.CreateAnnouncementRequest{
			Type: "news", Title: "預設公司層",
		})); err != nil {
			t.Fatalf("company_admin 無欄位應自動歸公司層並成功: %v", err)
		}
		row := db.Announcement.Query().Where(announcement.TitleEQ("預設公司層")).OnlyX(ctx)
		if row.CompanyID == nil || *row.CompanyID != coA || row.DepartmentID != nil {
			t.Fatalf("company_admin 預設應 company=自己/dept=空,got company=%v dept=%v",
				row.CompanyID, row.DepartmentID)
		}
		// dept_admin 完全不帶範圍 → 部門層(補自己部門)。
		dadmin := newAnnouncementClient(t, db, annIdentity("dept_admin", uItoa(coA), uItoa(deptA)))
		if _, err := dadmin.CreateAnnouncement(ctx, connect.NewRequest(&v1.CreateAnnouncementRequest{
			Type: "news", Title: "預設部門層",
		})); err != nil {
			t.Fatalf("dept_admin 無欄位應自動歸本部門並成功: %v", err)
		}
		row2 := db.Announcement.Query().Where(announcement.TitleEQ("預設部門層")).OnlyX(ctx)
		if row2.CompanyID == nil || *row2.CompanyID != coA ||
			row2.DepartmentID == nil || *row2.DepartmentID != deptA {
			t.Fatalf("dept_admin 預設應 company=自己/dept=自己,got company=%v dept=%v",
				row2.CompanyID, row2.DepartmentID)
		}
	})

	// 拒絕路徑:同一 db、逐案換身分。
	// (「company_admin 不帶範圍」「dept_admin 不帶部門」不再是拒絕例 —— 空欄位
	// 自動歸屬見下方成功路徑;可表達的越權只剩「指向別人的範圍」。)
	db := openAnnDB(t)
	coA, coB, deptA, deptB := annSeed(t, db)
	deny := []struct {
		name     string
		identity authz.Identity
		req      *v1.CreateAnnouncementRequest
	}{
		{"company_admin 別家公司", annIdentity("company_admin", uItoa(coA), ""),
			&v1.CreateAnnouncementRequest{Type: "news", Title: "x", CompanyId: uItoa(coB)}},
		{"company_admin 部門掛別家公司", annIdentity("company_admin", uItoa(coA), ""),
			&v1.CreateAnnouncementRequest{Type: "news", Title: "x", CompanyId: uItoa(coA), DepartmentId: uItoa(deptB)}},
		{"dept_admin 別部門", annIdentity("dept_admin", uItoa(coA), uItoa(deptA)),
			&v1.CreateAnnouncementRequest{Type: "news", Title: "x", CompanyId: uItoa(coA), DepartmentId: uItoa(deptB)}},
		{"staff 一律拒", annIdentity("staff", uItoa(coA), uItoa(deptA)),
			&v1.CreateAnnouncementRequest{Type: "news", Title: "x"}},
	}
	for _, tc := range deny {
		t.Run("拒絕:"+tc.name, func(t *testing.T) {
			cc := newAnnouncementClient(t, db, tc.identity)
			_, err := cc.CreateAnnouncement(ctx, connect.NewRequest(tc.req))
			if connect.CodeOf(err) != connect.CodePermissionDenied {
				t.Fatalf("應 permission_denied,得到 %v", err)
			}
		})
	}

	t.Run("非法型別", func(t *testing.T) {
		cc := newAnnouncementClient(t, openAnnDB(t), annIdentity("super", "", ""))
		_, err := cc.CreateAnnouncement(ctx, connect.NewRequest(&v1.CreateAnnouncementRequest{
			Type: "video", Title: "x",
		}))
		if connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Fatalf("非法 type 應 invalid_argument,得到 %v", err)
		}
	})
	t.Run("空標題", func(t *testing.T) {
		cc := newAnnouncementClient(t, openAnnDB(t), annIdentity("super", "", ""))
		_, err := cc.CreateAnnouncement(ctx, connect.NewRequest(&v1.CreateAnnouncementRequest{
			Type: "news", Title: "  ",
		}))
		if connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Fatalf("空標題應 invalid_argument,得到 %v", err)
		}
	})
}

// TestAnnouncementListScoping 管理列表只列可管理範圍(全系統列只給 super;公司/部門各自收斂)。
func TestAnnouncementListScoping(t *testing.T) {
	ctx := context.Background()
	db := openAnnDB(t)
	coA, coB, deptA, _ := annSeed(t, db)

	super := newAnnouncementClient(t, db, annIdentity("super", "", ""))
	for _, r := range []*v1.CreateAnnouncementRequest{
		{Type: "news", Title: "sys"},
		{Type: "news", Title: "co-a", CompanyId: uItoa(coA)},
		{Type: "news", Title: "dept-a", CompanyId: uItoa(coA), DepartmentId: uItoa(deptA)},
		{Type: "news", Title: "co-b", CompanyId: uItoa(coB)},
	} {
		if _, err := super.CreateAnnouncement(ctx, connect.NewRequest(r)); err != nil {
			t.Fatalf("seed %q: %v", r.Title, err)
		}
	}

	countFor := func(id authz.Identity) int {
		t.Helper()
		c := newAnnouncementClient(t, db, id)
		resp, err := c.ListAnnouncements(ctx, connect.NewRequest(&v1.ListAnnouncementsRequest{}))
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		return int(resp.Msg.GetTotal())
	}

	if got := countFor(annIdentity("super", "", "")); got != 4 {
		t.Fatalf("super 應看到全部 4 筆,got %d", got)
	}
	if got := countFor(annIdentity("company_admin", uItoa(coA), "")); got != 2 {
		t.Fatalf("company_admin A 應看到公司層+部門層 2 筆(不含全系統/B),got %d", got)
	}
	if got := countFor(annIdentity("dept_admin", uItoa(coA), uItoa(deptA))); got != 1 {
		t.Fatalf("dept_admin 應只看到本部門 1 筆,got %d", got)
	}
}

// TestAnnouncementUpdateDeleteGuards 更新/刪除同時守「舊範圍」與「新範圍」:
// 全系統公告僅 super 可動;他公司列一律拒絕。
func TestAnnouncementUpdateDeleteGuards(t *testing.T) {
	ctx := context.Background()
	db := openAnnDB(t)
	coA, coB, _, _ := annSeed(t, db)

	super := newAnnouncementClient(t, db, annIdentity("super", "", ""))
	sysResp, err := super.CreateAnnouncement(ctx, connect.NewRequest(&v1.CreateAnnouncementRequest{
		Type: "news", Title: "sys",
	}))
	if err != nil {
		t.Fatalf("seed sys: %v", err)
	}
	sys := sysResp.Msg.GetAnnouncement()
	coRowBResp, err := super.CreateAnnouncement(ctx, connect.NewRequest(&v1.CreateAnnouncementRequest{
		Type: "news", Title: "b", CompanyId: uItoa(coB),
	}))
	if err != nil {
		t.Fatalf("seed co-b: %v", err)
	}
	coRowB := coRowBResp.Msg.GetAnnouncement()

	adminA := newAnnouncementClient(t, db, annIdentity("company_admin", uItoa(coA), ""))
	_, err = adminA.UpdateAnnouncement(ctx, connect.NewRequest(&v1.UpdateAnnouncementRequest{
		Id: sys.GetId(), Type: "news", Title: "改", PublishAt: sys.GetPublishAt(),
		IsActive: true, DeployWeb: true, DeployApp: true,
	}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("company_admin 改全系統公告應 permission_denied,got %v", err)
	}
	_, err = adminA.UpdateAnnouncement(ctx, connect.NewRequest(&v1.UpdateAnnouncementRequest{
		Id: coRowB.GetId(), Type: "news", Title: "改",
		PublishAt: coRowB.GetPublishAt(), IsActive: true, DeployWeb: true, DeployApp: true,
	}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("company_admin 改別家公司公告應 permission_denied,got %v", err)
	}
	if _, err = adminA.DeleteAnnouncement(ctx, connect.NewRequest(&v1.DeleteAnnouncementRequest{
		Id: sys.GetId(),
	})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("company_admin 刪全系統公告應 permission_denied,got %v", err)
	}

	// super 更新(全量替換路徑) → 成功且值真的換掉。
	if _, err := super.UpdateAnnouncement(ctx, connect.NewRequest(&v1.UpdateAnnouncementRequest{
		Id: sys.GetId(), Type: "news", Title: "改後", PublishAt: sys.GetPublishAt(),
		IsActive: true, DeployWeb: true, DeployApp: true,
	})); err != nil {
		t.Fatalf("super 更新應成功: %v", err)
	}
	if got := db.Announcement.Query().Where(announcement.TitleEQ("改後")).CountX(ctx); got != 1 {
		t.Fatalf("更新後標題應為「改後」,count=%d", got)
	}
	// super 刪除 → 軟刪除。
	if _, err := super.DeleteAnnouncement(ctx, connect.NewRequest(&v1.DeleteAnnouncementRequest{
		Id: sys.GetId(),
	})); err != nil {
		t.Fatalf("super 刪除應成功: %v", err)
	}
	row := db.Announcement.GetX(ctx, mustAnnID(t, sys.GetId()))
	if row.DeletedAt == nil {
		t.Fatal("刪除應為軟刪除(deleted_at 非空)")
	}
	// 軟刪除後不再出現在管理列表。
	listResp, err := super.ListAnnouncements(ctx, connect.NewRequest(&v1.ListAnnouncementsRequest{}))
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range listResp.Msg.GetAnnouncements() {
		if a.GetId() == sys.GetId() {
			t.Fatal("已刪除公告不得再出現於列表")
		}
	}
}

// mustAnnID 把 proto id 字串轉 int。
func mustAnnID(t *testing.T, raw string) int {
	t.Helper()
	v, err := parseID(raw)
	if err != nil {
		t.Fatalf("id %q: %v", raw, err)
	}
	return v
}

// TestAnnouncementActiveListVisibility spec「上下架時間窗與啟用狀態」+「平台篩選投放」
// +「前台展示與排序」:前台列表只回當下可見者,依 sort_order(同值依 id),並依型別分組。
//
// 這是管理列表**不能**取代的一條路徑:管理面看得到未上架/停用(要能預覽與編輯),
// 前台只能看到已上架的 —— 少了前台過濾,未到時間或已停用的公告會直接曝光。
func TestAnnouncementActiveListVisibility(t *testing.T) {
	ctx := context.Background()
	db := openAnnDB(t)
	coA, _, _, _ := annSeed(t, db)

	super := newAnnouncementClient(t, db, annIdentity("super", "", ""))
	create := func(req *v1.CreateAnnouncementRequest) *v1.Announcement {
		t.Helper()
		resp, err := super.CreateAnnouncement(ctx, connect.NewRequest(req))
		if err != nil {
			t.Fatalf("create %+v: %v", req, err)
		}
		return resp.Msg.GetAnnouncement()
	}
	now := time.Now().UTC()

	// 可見:已上架、在時間窗內、投放兩邊（sort_order=2,用來驗排序。
	create(&v1.CreateAnnouncementRequest{
		Type: "banner", Title: "可見輪播", SortOrder: 2,
		PublishAt: now.Add(-time.Hour).Format(time.RFC3339),
		IsActive:  true, DeployWeb: true, DeployApp: true, CompanyId: uItoa(coA),
	})
	visibleFirst := create(&v1.CreateAnnouncementRequest{
		Type: "banner", Title: "排序在前", SortOrder: 1,
		PublishAt: now.Add(-time.Hour).Format(time.RFC3339),
		IsActive:  true, DeployWeb: true, DeployApp: true, CompanyId: uItoa(coA),
	})
	// 不可見:尚未到發佈時間。
	create(&v1.CreateAnnouncementRequest{
		Type: "banner", Title: "未來才發", SortOrder: 0,
		PublishAt: now.Add(time.Hour).Format(time.RFC3339),
		IsActive:  true, DeployWeb: true, DeployApp: true, CompanyId: uItoa(coA),
	})
	// 不可見:已過下架時間。
	create(&v1.CreateAnnouncementRequest{
		Type: "banner", Title: "已下架", SortOrder: 0,
		PublishAt:   now.Add(-2 * time.Hour).Format(time.RFC3339),
		UnpublishAt: now.Add(-time.Hour).Format(time.RFC3339),
		IsActive:    true, DeployWeb: true, DeployApp: true, CompanyId: uItoa(coA),
	})
	// 不可見:停用。
	create(&v1.CreateAnnouncementRequest{
		Type: "banner", Title: "停用中", SortOrder: 0,
		PublishAt: now.Add(-time.Hour).Format(time.RFC3339),
		IsActive:  false, DeployWeb: true, DeployApp: true, CompanyId: uItoa(coA),
	})
	// 平台過濾:只投 Web。
	webOnly := create(&v1.CreateAnnouncementRequest{
		Type: "news", Title: "只投Web", SortOrder: 1,
		PublishAt: now.Add(-time.Hour).Format(time.RFC3339),
		IsActive:  true, DeployWeb: true, DeployApp: false, CompanyId: uItoa(coA),
	})
	// 平台過濾:只投 App。
	appOnly := create(&v1.CreateAnnouncementRequest{
		Type: "news", Title: "只投App", SortOrder: 1,
		PublishAt: now.Add(-time.Hour).Format(time.RFC3339),
		IsActive:  true, DeployWeb: false, DeployApp: true, CompanyId: uItoa(coA),
	})

	// customer 也看得到(前台):RLS 可見 + 角色有 announcement read。
	cust := newAnnouncementClient(t, db, annIdentity("customer", uItoa(coA), ""))

	webResp, err := cust.ListActiveAnnouncements(ctx, connect.NewRequest(
		&v1.ListActiveAnnouncementsRequest{Platform: "web"}))
	if err != nil {
		t.Fatalf("customer ListActive(web): %v", err)
	}
	// 排序:sort_order 升冪,故 visibleFirst(1) 在 visible(2) 前。
	if got := webResp.Msg.GetBanners(); len(got) != 2 {
		titles := make([]string, 0, len(got))
		for _, b := range got {
			titles = append(titles, b.GetTitle())
		}
		t.Fatalf("web banner 應 2 筆(可見者),got %d: %v", len(got), titles)
	} else if got[0].GetTitle() != visibleFirst.GetTitle() {
		t.Fatalf("banner 應依 sort_order 排序,got 首位 %q", got[0].GetTitle())
	}
	// 平台過濾:web 只回 deploy_web=true → 只投 Web 的那筆在、只投 App 的不在。
	webTitles := map[string]bool{}
	for _, n := range webResp.Msg.GetNews() {
		webTitles[n.GetTitle()] = true
	}
	if !webTitles[webOnly.GetTitle()] {
		t.Fatal("web 列表應含只投 Web 的公告")
	}
	if webTitles[appOnly.GetTitle()] {
		t.Fatal("web 列表不得含只投 App 的公告（spec「平台篩選投放」）")
	}

	appResp, err := cust.ListActiveAnnouncements(ctx, connect.NewRequest(
		&v1.ListActiveAnnouncementsRequest{Platform: "app"}))
	if err != nil {
		t.Fatalf("customer ListActive(app): %v", err)
	}
	appTitles := map[string]bool{}
	for _, n := range appResp.Msg.GetNews() {
		appTitles[n.GetTitle()] = true
	}
	if !appTitles[appOnly.GetTitle()] {
		t.Fatal("app 列表應含只投 App 的公告")
	}
	if appTitles[webOnly.GetTitle()] {
		t.Fatal("app 列表不得含只投 Web 的公告（spec「平台篩選投放」）")
	}

	// 非法/未帶平台:不得靜默回全部(否則一邊會看到不該投放的公告)。
	if _, err := cust.ListActiveAnnouncements(ctx, connect.NewRequest(
		&v1.ListActiveAnnouncementsRequest{})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("未帶 platform 應 invalid_argument,got %v", err)
	}
	if _, err := cust.ListActiveAnnouncements(ctx, connect.NewRequest(
		&v1.ListActiveAnnouncementsRequest{Platform: "desktop"})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("非法 platform 應 invalid_argument,got %v", err)
	}

	// 軟刪除後不再出現(spec「軟刪除公告」:前台預設不再顯示)。
	if _, err := super.DeleteAnnouncement(ctx, connect.NewRequest(
		&v1.DeleteAnnouncementRequest{Id: visibleFirst.GetId()})); err != nil {
		t.Fatalf("delete: %v", err)
	}
	after, err := cust.ListActiveAnnouncements(ctx, connect.NewRequest(
		&v1.ListActiveAnnouncementsRequest{Platform: "web"}))
	if err != nil {
		t.Fatalf("ListActive after delete: %v", err)
	}
	for _, b := range after.Msg.GetBanners() {
		if b.GetId() == visibleFirst.GetId() {
			t.Fatal("軟刪除的公告不得出現在前台")
		}
	}
}

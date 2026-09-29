//go:build integration

// RotateCompanyExternalID 真 PostgreSQL 整合測試:輪換後 external_id 確實變更、業務稽核
// (audit_logs)留下 rotate_external_id 一筆、且無 company/update 權限者被拒。
//
// 為何必須真 PG:external_id 欄由 00053 加、ent 由本變更補認;輪換寫入與稽核同筆租戶交易、
// 以及 RLS 下的寫入路徑,記憶體假實作驗不出。
package services

import (
	"strconv"
	"testing"

	"connectrpc.com/connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationRotateCompanyExternalID 驗證輪換路徑的三個不變式。
func TestIntegrationRotateCompanyExternalID(t *testing.T) {
	testsupport.RequiresContainer(t)
	ctx := t.Context()
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	sqlDB, db := openPGEntClientFromGoose(t, dsn)

	actorCo := db.Company.Create().SetName("操作者公司").SetIdentifier("ROT-ACTOR").SaveX(ctx)
	actor := db.User.Create().
		SetEmail("rot-actor@example.com").
		SetName("操作者").
		SetStatus("active").
		SetRole("super").
		SetPasswordHash("x").
		SetCompanyID(actorCo.ID).
		SaveX(ctx)
	target := db.Company.Create().SetName("目標公司").SetIdentifier("ROT-TARGET").SaveX(ctx)
	before := target.ExternalID.String()

	// super 身分具備 company/update。
	cc, _ := newCompanySoftDeleteServer(t, db, authz.Identity{
		UserID: strconv.Itoa(actor.ID), CompanyID: strconv.Itoa(actorCo.ID), Role: "super", Roles: []string{"super"},
	})

	t.Run("輪換後 external_id 變更且回傳新值", func(t *testing.T) {
		resp, err := cc.RotateCompanyExternalID(ctx, connect.NewRequest(&v1.RotateCompanyExternalIDRequest{
			CompanyId: strconv.Itoa(target.ID),
		}))
		if err != nil {
			t.Fatalf("輪換: %v", err)
		}
		if resp.Msg.Company.ExternalId == before {
			t.Fatalf("輪換後 external_id 應變更,仍為 %s", before)
		}
		// DB 實際值也變更。
		reload, err := db.Company.Get(ctx, target.ID)
		if err != nil {
			t.Fatalf("重載公司: %v", err)
		}
		if reload.ExternalID.String() != resp.Msg.Company.ExternalId {
			t.Fatalf("DB external_id %s 與回傳 %s 不符", reload.ExternalID.String(), resp.Msg.Company.ExternalId)
		}
	})

	t.Run("業務稽核留 rotate_external_id 一筆", func(t *testing.T) {
		// 輪換前後各查一次 audit_logs(業務稽核,非 platform.audit_logs)。
		var before, after int
		if err := sqlDB.QueryRowContext(ctx,
			`SELECT count(*) FROM audit_logs WHERE action = 'rotate_external_id' AND resource_id = $1`,
			strconv.Itoa(target.ID)).Scan(&before); err != nil {
			t.Fatalf("查稽核前: %v", err)
		}
		if _, err := cc.RotateCompanyExternalID(ctx, connect.NewRequest(&v1.RotateCompanyExternalIDRequest{
			CompanyId: strconv.Itoa(target.ID),
		})); err != nil {
			t.Fatalf("二次輪換: %v", err)
		}
		if err := sqlDB.QueryRowContext(ctx,
			`SELECT count(*) FROM audit_logs WHERE action = 'rotate_external_id' AND resource_id = $1`,
			strconv.Itoa(target.ID)).Scan(&after); err != nil {
			t.Fatalf("查稽核後: %v", err)
		}
		if after != before+1 {
			t.Fatalf("rotate_external_id 稽核應恰增一筆,before=%d after=%d", before, after)
		}
	})

	t.Run("無 company/update 權限者被拒", func(t *testing.T) {
		viewer := db.User.Create().
			SetEmail("rot-viewer@example.com").
			SetName("旁觀者").
			SetStatus("active").
			SetRole("customer").
			SetPasswordHash("x").
			SetCompanyID(actorCo.ID).
			SaveX(ctx)
		vc, _ := newCompanySoftDeleteServer(t, db, authz.Identity{
			UserID: strconv.Itoa(viewer.ID), CompanyID: strconv.Itoa(actorCo.ID), Role: "customer", Roles: []string{"customer"},
		})
		_, err := vc.RotateCompanyExternalID(ctx, connect.NewRequest(&v1.RotateCompanyExternalIDRequest{
			CompanyId: strconv.Itoa(target.ID),
		}))
		if err == nil {
			t.Fatalf("無權限者輪換應被拒")
		}
	})
}


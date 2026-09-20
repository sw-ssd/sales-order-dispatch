package services

// T9:平台寫入 RPC 的**服務層契約**(免容器的假 store ＋ 假帳務 ＋ 記錄版快取)。
//
// 這裡驗的是三件在真 SQL 上看不出來的事:
//  1. **每個寫入 RPC 的三條共同契約**(無 operator 身分 → AUTH-4001;reason 全空白 → SYS-1001;
//     成功 → 恰一筆稽核且 actor 是真的 operator);
//  2. **一次寫入 = 一個交易**:失敗時資料與稽核**都不留**(假 store 的 WithTx 回滾模型),
//     提交失敗時**不失效快取**(失效必須跟在提交之後,否則回滾時白刪);
//  3. **快取失效的範圍**:override → 該公司;方案／價目 → 全量(且「不支援全量」的哨兵不得讓
//     RPC 失敗 —— 那是降級部署,不是故障)。
//
// 寫入的 SQL 欄位對應(含開票三欄真的落進 subscription_periods)由
// platform_admin_write_integration_test.go 以真容器把關。

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/operatorauth"
	platformstore "github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
	platformv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1"
)

// recordingCache 記錄失效呼叫;其餘 Cache 方法為 no-op。
type recordingCache struct {
	deleted  []string
	failNext error
	// scans 非 nil 即支援全量失效(Scanner);nil 時 InvalidateAll 會回 ErrScanUnsupported。
	scans []string
}

func (c *recordingCache) Get(context.Context, string) ([]byte, bool, error) { return nil, false, nil }

func (c *recordingCache) Set(context.Context, string, []byte, time.Duration) error { return nil }

func (c *recordingCache) Delete(_ context.Context, key string) error {
	if c.failNext != nil {
		err := c.failNext
		c.failNext = nil
		return err
	}
	c.deleted = append(c.deleted, key)
	return nil
}

func (c *recordingCache) Keys(context.Context, string) ([]string, error) { return c.scans, nil }

// fakeSeatCounter 為 entitlements.Counter 的假實作(席位用量)。
type fakeSeatCounter struct {
	used int
	err  error
}

func (c *fakeSeatCounter) Count(context.Context, int, string) (int, error) { return c.used, c.err }

// newReadOnlyPlatformService 建構「讀取面」的服務(既有測試用):真的帳務狀態機(假
// BillingStore)但沒有快取、沒有計數器 —— 讀取 RPC 不會走到那兩個依賴。
func newReadOnlyPlatformService(st platformStore) *PlatformAdminService {
	return NewPlatformAdminService(st, billing.NewBilling(platformstore.NewFakeBilling()), nil, nil)
}

// writeRecord 為假 store 記下的一次寫入(資料面)。所有寫入共用一個結構,是為了讓「交易回滾」
// 只需要截斷一個切片 —— 分開的型別會讓每個 slice 各自要一份 mark／truncate。
type writeRecord struct {
	kind        string
	companyID   int64
	featureCode string
	enabled     *bool
	limit       *int64
	planCode    string
	cycle       string
	baseCents   int64
	seatCents   int64
	operatorID  int64
	email       string
	name        string
	role        string
	key, value  string
}

// fakeWrites 為假 store 的**已提交**寫入與稽核;writes 是在交易內累積的緩衝。
//
// 交易語意:WithTx 記下長度,fn 回錯誤即截斷回該長度(記憶體裡沒有真的交易,這是唯一能表達
// 「失敗不留半成品」的方式)。它**不會**擋下「沒開交易就直接寫」的誤用 —— 那條界線由真容器的
// 整合測試把關(與 store.FakeBilling 同一個立場)。
type fakeWrites struct {
	writes []writeRecord
	audits []recordedAuditTx
	// commits 為成功提交的交易數(含 no-op:那也是一個交易)。
	commits int
	// failWrite 注入寫入失敗(在任何稽核之前);failAudit 注入稽核失敗。
	failWrite bool
	failAudit bool
	// revoked/lastOverrideID 供撤銷與新增的斷言。「撤銷由 overrideId 反查租戶」這條路徑
	// 的 company 由 store 回帶,故假 store 必須能給出不同的值(而不是永遠 42)。
	revokeCompany int64
	revokeFeature string
}

type recordedAuditTx struct {
	operatorID                           int64
	action, targetType, targetID, reason string
	before, after                        []byte
}

// defaultRevokeCompany 為撤銷回帶的預設租戶(沒特別指定時)。
const defaultRevokeCompany = 42

func (f *fakePlatformStore) WithTx(_ context.Context, fn func(*sql.Tx) error) error {
	mark, auditMark := len(f.writes.writes), len(f.writes.audits)
	if err := fn(nil); err != nil {
		// 回滾:交易內寫入的資料與稽核一律消失。
		f.writes.writes = f.writes.writes[:mark]
		f.writes.audits = f.writes.audits[:auditMark]
		return err
	}
	f.writes.commits++
	return nil
}

func (f *fakePlatformStore) RecordAuditTx(_ context.Context, _ *sql.Tx, operatorID int64,
	action, targetType, targetID, reason string, before, after []byte) error {
	if f.writes.failAudit {
		return errors.New("模擬稽核寫入失敗")
	}
	f.writes.audits = append(f.writes.audits, recordedAuditTx{
		operatorID: operatorID, action: action, targetType: targetType, targetID: targetID,
		reason: reason, before: before, after: after,
	})
	return nil
}

func (f *fakePlatformStore) SetTenantOverrideTx(_ context.Context, _ *sql.Tx,
	in platformstore.TenantOverrideInput) (int64, error) {
	if f.writes.failWrite {
		return 0, errors.New("模擬寫入失敗")
	}
	f.writes.writes = append(f.writes.writes, writeRecord{
		kind: "override.set", companyID: in.CompanyID, featureCode: in.FeatureCode,
		enabled: in.Enabled, limit: in.Limit, operatorID: in.CreatedBy,
	})
	return 900 + int64(len(f.writes.writes)), nil
}

func (f *fakePlatformStore) RevokeTenantOverrideTx(_ context.Context, _ *sql.Tx,
	overrideID int64) (platformstore.OverrideRef, error) {
	if f.writes.failWrite {
		return platformstore.OverrideRef{}, errors.New("模擬寫入失敗")
	}
	f.writes.writes = append(f.writes.writes, writeRecord{kind: "override.revoke", companyID: overrideID})
	company := f.writes.revokeCompany
	if company == 0 {
		company = defaultRevokeCompany
	}
	return platformstore.OverrideRef{CompanyID: company, FeatureCode: f.writes.revokeFeature}, nil
}

func (f *fakePlatformStore) UpsertPlanPriceTx(_ context.Context, _ *sql.Tx,
	in platformstore.PlanPriceInput) error {
	if f.writes.failWrite {
		return errors.New("模擬寫入失敗")
	}
	f.writes.writes = append(f.writes.writes, writeRecord{
		kind: "plan.price_upsert", planCode: in.PlanCode, cycle: in.BillingCycle,
		baseCents: in.BaseCents, seatCents: in.SeatCents,
	})
	return nil
}

func (f *fakePlatformStore) SetPlanEntitlementTx(_ context.Context, _ *sql.Tx,
	planCode, featureCode string, enabled bool, limit *int64) error {
	if f.writes.failWrite {
		return errors.New("模擬寫入失敗")
	}
	e := enabled
	f.writes.writes = append(f.writes.writes, writeRecord{
		kind: "plan.entitlement_set", planCode: planCode, featureCode: featureCode,
		enabled: &e, limit: limit,
	})
	return nil
}

func (f *fakePlatformStore) CreateOperatorTx(_ context.Context, _ *sql.Tx,
	email, name, role string) (int64, error) {
	if f.createOperatorConflict {
		return 0, platformstore.ErrConflict
	}
	if f.writes.failWrite {
		return 0, errors.New("模擬寫入失敗")
	}
	f.writes.writes = append(f.writes.writes, writeRecord{kind: "operator.create", email: email, name: name, role: role})
	return 77, nil
}

func (f *fakePlatformStore) DisableOperatorTx(_ context.Context, _ *sql.Tx, operatorID int64) error {
	if f.disableLastAdmin {
		return platformstore.ErrLastAdmin
	}
	if f.writes.failWrite {
		return errors.New("模擬寫入失敗")
	}
	f.writes.writes = append(f.writes.writes, writeRecord{kind: "operator.disable", operatorID: operatorID})
	return nil
}

func (f *fakePlatformStore) Settings(context.Context) (map[string]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	out := map[string]string{}
	for k, v := range f.settings {
		out[k] = v
	}
	return out, nil
}

func (f *fakePlatformStore) UpsertSettingTx(_ context.Context, _ *sql.Tx, key, value string) error {
	if f.writes.failWrite {
		return errors.New("模擬寫入失敗")
	}
	f.writes.writes = append(f.writes.writes, writeRecord{kind: "setting.upsert", key: key, value: value})
	return nil
}

func (f *fakePlatformStore) ListReceivables(context.Context, int32, int32) ([]ReceivableRow, int, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.receivables, len(f.receivables), nil
}

// withOperator 注入平台身分(operator id 固定為 42、角色 admin)。身分由 interceptor 在真實路徑
// 上注入,單元測試直接放進 ctx —— 服務層的第二層檢查看的就是這件事。
func withOperator(ctx context.Context) context.Context { return withOperatorRole(ctx, "admin") }

// withOperatorRole 注入指定角色的平台身分(operator id 固定為 42):角色守衛的測試用。
func withOperatorRole(ctx context.Context, role string) context.Context {
	return operatorauth.WithIdentity(ctx,
		operatorauth.Identity{OperatorID: 42, Email: "ops@example.com", Role: role})
}

// newWriteHarness 建構一組乾淨的服務:假 store ＋ 真帳務(假 BillingStore)＋ 記錄版快取
// ＋ 假席位計數器。每個 case 各自一份,避免互相污染。
func newWriteHarness(used int) (*PlatformAdminService, *fakePlatformStore, *recordingCache, *platformstore.FakeBilling) {
	st := &fakePlatformStore{
		writes:   &fakeWrites{},
		settings: map[string]string{"trial_days": "14", "grace_days": "7", "lead_days": "14"},
	}
	book := platformstore.NewFakeBilling()
	book.PutPlan("std", 1)
	book.PutPlan("pro", 2)
	book.PutSubscription(platformstore.Subscription{
		ID: 5, CompanyID: 42, PlanID: 1, SeatCount: 5, BillingCycle: "monthly", Status: "active"})
	book.PutPeriod(platformstore.Period{
		ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open", AmountCents: 150000,
		PeriodStart: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)})

	cache := &recordingCache{}
	svc := NewPlatformAdminService(st, billing.NewBilling(book), cache, &fakeSeatCounter{used: used})
	return svc, st, cache, book
}

// cacheScope 為一次寫入應失效的快取範圍:none(不影響權益)、tenant(單一租戶)、all(全量)。
type cacheScope string

const (
	cacheNone   cacheScope = ""
	cacheTenant cacheScope = "ent:42"
	cacheAll    cacheScope = "*"
)

// writeCall 為一個寫入 RPC 的呼叫器 ＋ 它的性質宣告。同一個簽章讓「三條共同契約」與「失效範圍」
// 可以表驅動逐條驗;性質寫在這裡的理由是**新增寫入 RPC 時漏掉宣告會少一個 case**,而不是漏一行
// 斷言(後者在評審眼裡看不出來)。
type writeCall struct {
	name string
	// call 以參數化的 reason 呼叫 RPC(reason 是唯一會變的欄位)。
	call func(svc *PlatformAdminService, ctx context.Context, reason string) error
	// billingBacked 為 true 表示資料由 billing 的狀態機寫入(服務層只帶參數過去),
	// 故稽核與資料要在 FakeBilling 上看,而不是在假 store 上。
	billingBacked bool
	// invalidates 為這次寫入應失效的快取範圍。
	invalidates cacheScope
}

// writeCalls 覆蓋**每一個**寫入 RPC。
func writeCalls() []writeCall {
	return []writeCall{
		{"RecordPayment", func(svc *PlatformAdminService, ctx context.Context, reason string) error {
			_, err := svc.RecordPayment(ctx, connect.NewRequest(&platformv1.RecordPaymentRequest{
				CompanyId: "42", PeriodNo: 1, Reason: reason}))
			return err
		}, true, cacheTenant},
		{"SetSeatCount", func(svc *PlatformAdminService, ctx context.Context, reason string) error {
			_, err := svc.SetSeatCount(ctx, connect.NewRequest(&platformv1.SetSeatCountRequest{
				CompanyId: "42", SeatCount: 10, Reason: reason}))
			return err
		}, true, cacheTenant},
		{"ChangePlan", func(svc *PlatformAdminService, ctx context.Context, reason string) error {
			_, err := svc.ChangePlan(ctx, connect.NewRequest(&platformv1.ChangePlanRequest{
				CompanyId: "42", PlanCode: "pro", Reason: reason}))
			return err
		}, true, cacheTenant},
		{"CancelSubscription", func(svc *PlatformAdminService, ctx context.Context, reason string) error {
			_, err := svc.CancelSubscription(ctx, connect.NewRequest(&platformv1.CancelSubscriptionRequest{
				CompanyId: "42", AtPeriodEnd: true, Reason: reason}))
			return err
		}, true, cacheTenant},
		{"UpdateBillingSettings", func(svc *PlatformAdminService, ctx context.Context, reason string) error {
			_, err := svc.UpdateBillingSettings(ctx, connect.NewRequest(&platformv1.UpdateBillingSettingsRequest{
				Settings: []*platformv1.BillingSetting{{Key: "grace_days", Value: "3"}}, Reason: reason}))
			return err
		}, false, cacheNone},
		{"SetTenantOverride", func(svc *PlatformAdminService, ctx context.Context, reason string) error {
			_, err := svc.SetTenantOverride(ctx, connect.NewRequest(&platformv1.SetTenantOverrideRequest{
				CompanyId: "42", FeatureCode: entitlements.LimitSeats, LimitSet: true, LimitValue: 50,
				Owner: "sales@example.com", Reason: reason}))
			return err
		}, false, cacheTenant},
		{"RevokeTenantOverride", func(svc *PlatformAdminService, ctx context.Context, reason string) error {
			_, err := svc.RevokeTenantOverride(ctx, connect.NewRequest(&platformv1.RevokeTenantOverrideRequest{
				OverrideId: "1", Reason: reason}))
			return err
		}, false, cacheTenant},
		{"UpsertPlanPrice", func(svc *PlatformAdminService, ctx context.Context, reason string) error {
			_, err := svc.UpsertPlanPrice(ctx, connect.NewRequest(&platformv1.UpsertPlanPriceRequest{
				PlanCode: "std", BillingCycle: "monthly", BasePrice: "1000.00", SeatPrice: "200.00",
				Reason: reason}))
			return err
		}, false, cacheAll},
		{"SetPlanEntitlement", func(svc *PlatformAdminService, ctx context.Context, reason string) error {
			_, err := svc.SetPlanEntitlement(ctx, connect.NewRequest(&platformv1.SetPlanEntitlementRequest{
				PlanCode: "std", FeatureCode: entitlements.LimitSeats, Enabled: true, LimitSet: true,
				LimitValue: 20, Reason: reason}))
			return err
		}, false, cacheAll},
		{"CreateOperator", func(svc *PlatformAdminService, ctx context.Context, reason string) error {
			_, err := svc.CreateOperator(ctx, connect.NewRequest(&platformv1.CreateOperatorRequest{
				Email: "new@example.com", Name: "新同事", Role: "operator", Reason: reason}))
			return err
		}, false, cacheNone},
		{"DisableOperator", func(svc *PlatformAdminService, ctx context.Context, reason string) error {
			_, err := svc.DisableOperator(ctx, connect.NewRequest(&platformv1.DisableOperatorRequest{
				OperatorId: "9", Reason: reason}))
			return err
		}, false, cacheNone},
	}
}

// TestWriteRPCsRequireOperatorAndReason 驗每個寫入 RPC 的兩條共同前置契約。
//
// 「擋下來了」不能只看回應:擋下時**不得寫任何資料、不得寫任何稽核** —— 一筆沒有 operator 的
// 稽核在 schema 上就寫不進去(operator_id NOT NULL),而一筆無人的寫入是無法回溯的變更。
func TestWriteRPCsRequireOperatorAndReason(t *testing.T) {
	for _, c := range writeCalls() {
		t.Run(c.name+"/無身分", func(t *testing.T) {
			svc, st, _, _ := newWriteHarness(0)
			err := c.call(svc, context.Background(), "匯款入帳")
			if connect.CodeOf(err) != connect.CodeUnauthenticated {
				t.Fatalf("無 operator 身分應 Unauthenticated,got %v", err)
			}
			if got := errorInfoOf(t, err).GetCode(); got != "AUTH-4001" {
				t.Fatalf("必須是註冊碼 AUTH-4001,got %q", got)
			}
			assertNoWrites(t, st)
		})
		t.Run(c.name+"/reason 全空白", func(t *testing.T) {
			svc, st, _, _ := newWriteHarness(0)
			err := c.call(svc, withOperator(context.Background()), "   ")
			if connect.CodeOf(err) != connect.CodeInvalidArgument {
				t.Fatalf("reason 全空白應 InvalidArgument,got %v", err)
			}
			if got := errorInfoOf(t, err).GetCode(); got != "SYS-1001" {
				t.Fatalf("必須是註冊碼 SYS-1001,got %q", got)
			}
			assertNoWrites(t, st)
		})
	}
}

// TestWriteRPCsWriteExactlyOneAudit 驗每個寫入 RPC 成功時**恰好一筆**稽核,且 actor 是真的
// operator id(42,來自 ctx 的身分,不是租戶 user id);稽核與資料落在同一個交易(billing 寫入的
// 狀態轉移由 billing 自己寫稽核,服務層不再寫第二筆 —— 兩處各寫一筆就是兩份「誰改的」)。
func TestWriteRPCsWriteExactlyOneAudit(t *testing.T) {
	for _, c := range writeCalls() {
		t.Run(c.name, func(t *testing.T) {
			svc, st, cache, book := newWriteHarness(0)
			if c.invalidates == cacheAll {
				cache.scans = []string{"ent:1", "ent:2"}
			}
			if err := c.call(svc, withOperator(context.Background()), "匯款入帳"); err != nil {
				t.Fatalf("合法呼叫應成功: %v", err)
			}

			// 資料:寫入 RPC 必須真的寫了東西(否則「恰一筆稽核」是在稽核一件沒發生的事)。
			if c.billingBacked {
				if len(book.Audits()) == 0 {
					t.Fatal("帳務寫入必須留下 billing 的稽核(audits 空＝狀態機沒被走到)")
				}
			} else if len(st.writes.writes) == 0 {
				t.Fatal("寫入 RPC 必須真的寫資料")
			}

			audits := st.writes.audits
			if c.billingBacked {
				for _, a := range book.Audits() {
					if a.OperatorID != 42 {
						t.Fatalf("actor 必須是 ctx 上的 operator id(42),got %d", a.OperatorID)
					}
					if a.Reason != "匯款入帳" {
						t.Fatalf("稽核必須留下真實原因,got %q", a.Reason)
					}
				}
				if len(audits) != 0 {
					t.Fatalf("billing 路徑不得再寫第二筆平台稽核,got %+v", audits)
				}
			} else {
				if len(audits) != 1 {
					t.Fatalf("一次寫入必須恰寫一筆稽核,got %d", len(audits))
				}
				if audits[0].operatorID != 42 || audits[0].reason != "匯款入帳" {
					t.Fatalf("稽核的 actor／reason 錯誤: %+v", audits[0])
				}
			}

			assertInvalidated(t, c.invalidates, cache)
		})
	}
}

// TestWriteRPCsInvalidateExpectedScope 驗失效的範圍:方案／價目 → **全量**(改一格方案等於改到
// 每一個用它的租戶),其餘 → **該租戶**;而設定與 operator 白名單**不動**它(它們不進權益快照,
// 全刪只是白打一趟 Valkey)。
func TestWriteRPCsInvalidateExpectedScope(t *testing.T) {
	for _, c := range writeCalls() {
		t.Run(c.name, func(t *testing.T) {
			svc, _, cache, _ := newWriteHarness(0)
			if c.invalidates == cacheAll {
				cache.scans = []string{"ent:1", "ent:2"}
				// 同一個租戶的鍵也在掃描結果裡(全量失效的定義就是「連它一起刪」)。
				cache.scans = append(cache.scans, string(cacheTenant))
			}
			if err := c.call(svc, withOperator(context.Background()), "合法原因"); err != nil {
				t.Fatalf("呼叫失敗: %v", err)
			}
			assertInvalidated(t, c.invalidates, cache)
		})
	}
}

// assertInvalidated 依範圍斷言記錄版快取上的刪除。
func assertInvalidated(t *testing.T, scope cacheScope, cache *recordingCache) {
	t.Helper()
	switch scope {
	case cacheAll:
		// 全量失效的實作是 SCAN ent:* 後逐鍵刪;假快取回的掃描結果即應全部被刪(含單租戶鍵)。
		if len(cache.deleted) != len(cache.scans) || len(cache.deleted) == 0 {
			t.Fatalf("方案／價目異動必須失效**所有**租戶,got %v（掃到 %v）", cache.deleted, cache.scans)
		}
	case cacheNone:
		if len(cache.deleted) != 0 {
			t.Fatalf("此寫入不影響權益,不該失效快取,got %v", cache.deleted)
		}
	default:
		if len(cache.deleted) != 1 || cache.deleted[0] != string(scope) {
			t.Fatalf("應只失效 %s,got %v", scope, cache.deleted)
		}
	}
}

// TestRevokeTenantOverrideInvalidatesItsCompany 驗撤銷路徑的失效對象來自 **store 回帶的公司**。
//
// 撤銷的入參只有 override id,租戶是把它反查出來的 —— 服務層若偷懶用 0 或別的值失效,快取
// 上該租戶的舊權益會留到 TTL 到期(而失效看起來「有做」,只是刪錯鍵)。
func TestRevokeTenantOverrideInvalidatesItsCompany(t *testing.T) {
	svc, st, cache, _ := newWriteHarness(0)
	st.writes.revokeCompany = 77
	st.writes.revokeFeature = "limit.seats"

	resp, err := svc.RevokeTenantOverride(withOperator(context.Background()),
		connect.NewRequest(&platformv1.RevokeTenantOverrideRequest{OverrideId: "31", Reason: "合約到期"}))
	if err != nil {
		t.Fatalf("RevokeTenantOverride: %v", err)
	}
	if resp.Msg.GetCompanyId() != "77" || resp.Msg.GetFeatureCode() != "limit.seats" {
		t.Fatalf("回帶的租戶／功能來自 store: %v", resp.Msg)
	}
	if len(cache.deleted) != 1 || cache.deleted[0] != "ent:77" {
		t.Fatalf("必須失效 store 回帶的那個租戶(ent:77),got %v", cache.deleted)
	}
	// 稽核的目標是公司(operator 在 console 上是依 target_id 篩選的):override 的內部 id 對營運
	// 沒有意義,而 company_id 是查得到的東西。
	audit := st.writes.audits[0]
	if audit.targetType != "company" || audit.targetID != "77" {
		t.Fatalf("撤銷的稽核目標應為該公司,got %s/%s", audit.targetType, audit.targetID)
	}
}

// TestWriteFailureLeavesNoHalfWrite 驗「失敗不落半成品」的兩個方向:
//   - 資料寫入失敗 → 沒有稽核;
//   - 稽核寫入失敗 → 資料**回滾**(否則會留下一次沒有人知道是誰做的變更)。
//
// 以及一個關鍵的邊界:交易失敗時**不得失效快取** —— 資料沒變而快取被刪,只是白打一趟 Valkey;
// 更糟的是它讓「失效成功」與「寫入成功」在 log 上看不出區別。
func TestWriteFailureLeavesNoHalfWrite(t *testing.T) {
	for _, tc := range []struct {
		name      string
		failWrite bool
		failAudit bool
	}{
		{"資料寫入失敗", true, false},
		{"稽核寫入失敗", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, st, cache, _ := newWriteHarness(0)
			st.writes.failWrite, st.writes.failAudit = tc.failWrite, tc.failAudit

			_, err := svc.SetTenantOverride(withOperator(context.Background()),
				connect.NewRequest(&platformv1.SetTenantOverrideRequest{
					CompanyId: "42", FeatureCode: entitlements.LimitSeats, LimitSet: true, LimitValue: 50,
					Owner: "sales@example.com", Reason: "簽約承諾"}))
			if err == nil {
				t.Fatal("寫入失敗時必須回錯誤")
			}
			if connect.CodeOf(err) != connect.CodeInternal {
				t.Fatalf("非預期的失敗應收斂為 SYS-9000,got %v", err)
			}
			assertNoWrites(t, st)
			if len(cache.deleted) != 0 {
				t.Fatalf("交易沒有提交就不該失效快取,got %v", cache.deleted)
			}
		})
	}
}

// TestRecordPaymentGoesThroughBilling 驗收款 RPC **只**經 billing 的狀態機(沒有第二個改變訂閱
// 狀態的入口):期別被標為 paid、訂閱被帶回 active、事件與稽核都在 billing 的交易內。
func TestRecordPaymentGoesThroughBilling(t *testing.T) {
	svc, st, _, book := newWriteHarness(0)

	resp, err := svc.RecordPayment(withOperator(context.Background()),
		connect.NewRequest(&platformv1.RecordPaymentRequest{
			CompanyId: "42", PeriodNo: 1, Amount: "1500.00", ExternalRef: "BANK-1",
			InvoiceNo: "AB12345678", InvoiceStatus: "issued", BuyerTaxId: "12345678",
			Carrier: "/ABC1234", Reason: "匯款入帳"}))
	if err != nil {
		t.Fatalf("RecordPayment: %v", err)
	}
	if resp.Msg.GetPeriodNo() != 1 || resp.Msg.GetStatus() != "paid" {
		t.Fatalf("回應應帶期別與新狀態: %v", resp.Msg)
	}
	periods, err := book.PeriodsByStatus(context.Background(), "paid")
	if err != nil || len(periods) != 1 {
		t.Fatalf("期別必須被標為已付款: %v (%d 筆)", err, len(periods))
	}
	// 稽核由 billing 寫(action 為 record_payment),actor 是 RPC 帶入的 operator;
	// 服務層**不得**再寫第二筆(兩份「誰改的」只會有一份寫對)。
	audits := book.Audits()
	if len(audits) != 1 || audits[0].Action != "record_payment" || audits[0].OperatorID != 42 {
		t.Fatalf("收款必須留下 billing 的稽核(actor=operator): %+v", audits)
	}
	if len(st.writes.audits) != 0 {
		t.Fatalf("收款路徑不得另寫平台稽核(billing 已寫): %+v", st.writes.audits)
	}
}

// TestRecordPaymentSurfacesBillingRegisteredCode 驗收款衝突以 billing 的註冊碼原樣回給 console。
//
// 這裡刻意**不改寫**那個碼的語意(也不在服務層把它當成「不可重試」):PLAT-3002 由
// billing.RecordPayment 產生,涵蓋的失敗比「重複收款」更廣(金額不符、交易號衝突、唯一鍵)。
// 服務層只負責讓它原樣上桌,operator 才能看到可讀的原因並決定要不要重試。
func TestRecordPaymentSurfacesBillingRegisteredCode(t *testing.T) {
	svc, book := newPaidPeriodHarness()
	_ = book

	_, err := svc.RecordPayment(withOperator(context.Background()),
		connect.NewRequest(&platformv1.RecordPaymentRequest{
			CompanyId: "42", PeriodNo: 1, ExternalRef: "BANK-2", Reason: "第二筆匯入"}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("收款衝突應為 FailedPrecondition,got %v", err)
	}
	if got := errorInfoOf(t, err).GetCode(); got != "PLAT-3002" {
		t.Fatalf("必須是 billing 的註冊碼 PLAT-3002,got %q", got)
	}
}

// newPaidPeriodHarness 種一期「已付款(BANK-1)」的期別:同一期再收一筆**不同交易號**是帳務衝突。
func newPaidPeriodHarness() (*PlatformAdminService, *platformstore.FakeBilling) {
	st := &fakePlatformStore{writes: &fakeWrites{}, settings: map[string]string{}}
	book := platformstore.NewFakeBilling()
	book.PutSubscription(platformstore.Subscription{
		ID: 5, CompanyID: 42, PlanID: 1, SeatCount: 5, BillingCycle: "monthly", Status: "active"})
	book.PutPeriod(platformstore.Period{
		ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "paid", AmountCents: 150000,
		ExternalRef: "BANK-1", PaymentProvider: "manual"})
	svc := NewPlatformAdminService(st, billing.NewBilling(book), &recordingCache{}, &fakeSeatCounter{})
	return svc, book
}

// TestCancelSubscriptionRejectsImmediateTermination 驗 v1 只支援期末終止:at_period_end=false
// 回 PLAT-3001(訂閱狀態不允許此操作),且**什麼都沒寫** —— 立即終止需要按日比例計費與退款,
// 那是另一條流程(見 billing.CancelSubscription 的說明)。
func TestCancelSubscriptionRejectsImmediateTermination(t *testing.T) {
	svc, st, cache, _ := newWriteHarness(0)

	_, err := svc.CancelSubscription(withOperator(context.Background()),
		connect.NewRequest(&platformv1.CancelSubscriptionRequest{
			CompanyId: "42", AtPeriodEnd: false, Reason: "立即停用"}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("立即終止應 FailedPrecondition,got %v", err)
	}
	if got := errorInfoOf(t, err).GetCode(); got != "PLAT-3001" {
		t.Fatalf("必須是註冊碼 PLAT-3001,got %q", got)
	}
	assertNoWrites(t, st)
	if len(cache.deleted) != 0 {
		t.Fatalf("失敗不得失效快取,got %v", cache.deleted)
	}
}

// TestChangePlanSamePlanIsNoOpWithoutAudit 驗「與現行方案相同 → no-op,不寫稽核」:
// before 與 after 一模一樣的稽核會讓「誰真的改了方案」需要逐筆比對才看得出來。
func TestChangePlanSamePlanIsNoOpWithoutAudit(t *testing.T) {
	svc, st, cache, _ := newWriteHarness(0)

	resp, err := svc.ChangePlan(withOperator(context.Background()),
		connect.NewRequest(&platformv1.ChangePlanRequest{
			CompanyId: "42", PlanCode: "std", Reason: "確認方案"}))
	if err != nil {
		t.Fatalf("同方案的變更應成功(no-op): %v", err)
	}
	if len(st.writes.audits) != 0 {
		t.Fatalf("no-op 不得寫稽核,got %+v", st.writes.audits)
	}
	// 生效日仍要回(console 顯示「自 X 起維持原方案」比空白好):期別 1 的期末。
	if resp.Msg.GetEffectiveFrom() != "2026-10-01T00:00:00Z" {
		t.Fatalf("生效日應為當期期末,got %q", resp.Msg.GetEffectiveFrom())
	}
	// 快取失效照做:no-op 也要清一次(分辨「到底改了什麼」需要在交易內多帶旗標出來,
	// 而 DEL 是冪等的 —— 與 billing 的重送路徑同一個取捨)。
	if len(cache.deleted) == 0 {
		t.Fatal("成功路徑必須失效快取(no-op 亦同)")
	}
}

// TestSetSeatCountRejectsBelowUsage 驗「不得降到使用中席次以下」→ PLAT-5001,且
// details 帶 feature／used／limit(前端據以顯示用量),而失敗時沒有任何寫入與稽核。
func TestSetSeatCountRejectsBelowUsage(t *testing.T) {
	svc, st, _, _ := newWriteHarness(6) // 目前用 6 席

	_, err := svc.SetSeatCount(withOperator(context.Background()),
		connect.NewRequest(&platformv1.SetSeatCountRequest{
			CompanyId: "42", SeatCount: 5, Reason: "降席位"}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("席次低於使用量應 FailedPrecondition,got %v", err)
	}
	info := errorInfoOf(t, err)
	if info.GetCode() != "PLAT-5001" {
		t.Fatalf("必須是註冊碼 PLAT-5001,got %q", info.GetCode())
	}
	if got := info.GetDetails()["used"]; got != "6" {
		t.Fatalf("details 必須帶 used=6,got %q", got)
	}
	if got := info.GetDetails()["limit"]; got != "5" {
		t.Fatalf("details 必須帶 limit=5,got %q", got)
	}
	if got := info.GetDetails()["feature"]; got != entitlements.LimitSeats {
		t.Fatalf("details 必須帶 feature=%s,got %q", entitlements.LimitSeats, got)
	}
	assertNoWrites(t, st)
}

// TestSetSeatCountWithoutCounterFailsClosed 驗沒有計數器時**拒絕**而不是放行:
// 「不知道用了几席」不得被當成「0 席」(那等於關掉這個守衛)。
func TestSetSeatCountWithoutCounterFailsClosed(t *testing.T) {
	svc, st, _, _ := newWriteHarness(0)
	svc.counters = nil

	_, err := svc.SetSeatCount(withOperator(context.Background()),
		connect.NewRequest(&platformv1.SetSeatCountRequest{
			CompanyId: "42", SeatCount: 5, Reason: "降席位"}))
	if connect.CodeOf(err) != connect.CodeInternal {
		t.Fatalf("缺計數器應失敗,got %v", err)
	}
	assertNoWrites(t, st)
}

// TestInvalidateAllUnsupportedCacheDoesNotFailWrite 驗降級部署(Valkey 缺席 → MemoryCache,
// 不支援 SCAN)下,方案／價目的寫入**仍然成功**:
//   - 「不支援全量失效」是已知的部署形態,不是故障 → log 一行(最長 TTL 收斂);
//   - 讓它回錯誤會出現最壞的組合:資料改了、console 說失敗,operator 再按一次。
func TestInvalidateAllUnsupportedCacheDoesNotFailWrite(t *testing.T) {
	svc, st, _, _ := newWriteHarness(0)
	svc.cache = entitlements.NewMemoryCache() // 沒有 Keys → InvalidateAll 回 ErrScanUnsupported

	if _, err := svc.UpsertPlanPrice(withOperator(context.Background()),
		connect.NewRequest(&platformv1.UpsertPlanPriceRequest{
			PlanCode: "std", BillingCycle: "monthly", BasePrice: "1000.00", SeatPrice: "200.00",
			Reason: "調價"})); err != nil {
		t.Fatalf("不支援全量失效不得讓寫入失敗: %v", err)
	}
	if len(st.writes.writes) != 1 || len(st.writes.audits) != 1 {
		t.Fatalf("寫入與稽核都必須成立: %+v", st.writes)
	}
}

// TestCacheFailureDoesNotFailCommittedWrite 驗單租戶失效失敗(例:Valkey 掛掉)不讓**已提交**的
// 寫入回錯誤 —— 錢收了卻回 500,呼叫端會重試,而重試撞上的是冪等路徑。
func TestCacheFailureDoesNotFailCommittedWrite(t *testing.T) {
	svc, _, cache, book := newWriteHarness(0)
	cache.failNext = errors.New("模擬 Valkey 故障")

	if _, err := svc.SetSeatCount(withOperator(context.Background()),
		connect.NewRequest(&platformv1.SetSeatCountRequest{
			CompanyId: "42", SeatCount: 10, Reason: "擴編"})); err != nil {
		t.Fatalf("快取故障不得讓已提交的寫入失敗: %v", err)
	}
	subs, err := book.ActiveOrTrialingSubscriptions(context.Background())
	if err != nil || len(subs) != 1 || subs[0].SeatCount != 10 {
		t.Fatalf("席位必須真的落地(快取故障不影響交易): %v (%+v)", err, subs)
	}
	if len(book.Audits()) != 1 {
		t.Fatalf("稽核必須成立: %+v", book.Audits())
	}
}

// TestCreateOperatorConflictIsAlreadyExists 驗重複的 email 回 SYS-2001(輸入問題),
// 而不是讓唯一鍵的 23505 冒成 5xx。
func TestCreateOperatorConflictIsAlreadyExists(t *testing.T) {
	svc, st, _, _ := newWriteHarness(0)
	st.createOperatorConflict = true

	_, err := svc.CreateOperator(withOperator(context.Background()),
		connect.NewRequest(&platformv1.CreateOperatorRequest{
			Email: "dup@example.com", Role: "operator", Reason: "新增同事"}))
	if connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Fatalf("email 重複應 AlreadyExists,got %v", err)
	}
	if got := errorInfoOf(t, err).GetCode(); got != "SYS-2001" {
		t.Fatalf("必須是註冊碼 SYS-2001,got %q", got)
	}
}

// TestOperatorManagementRequiresAdmin 驗**操作者管理**只能由 admin 執行(裁決:operator 可
// 自我提權 / 停用所有 admin 的漏洞修補)。
//
// 為什麼 role 一定要檢查:此 patch 之前唯一建立 operator 的途徑是 seed,而後果是不可回復的
// 治理破壞 —— 任何一個 operator 都能 `CreateOperator(role="admin")` 憑空新增 admin,或停用
// 種子 admin 乃至最後一位 admin(console 全鎖死,只能直接改資料庫救)。
func TestOperatorManagementRequiresAdmin(t *testing.T) {
	calls := map[string]func(svc *PlatformAdminService, ctx context.Context) error{
		"CreateOperator": func(svc *PlatformAdminService, ctx context.Context) error {
			_, err := svc.CreateOperator(ctx, connect.NewRequest(&platformv1.CreateOperatorRequest{
				Email: "new@example.com", Role: "admin", Reason: "新增同事"}))
			return err
		},
		"DisableOperator": func(svc *PlatformAdminService, ctx context.Context) error {
			_, err := svc.DisableOperator(ctx, connect.NewRequest(&platformv1.DisableOperatorRequest{
				OperatorId: "9", Reason: "離職"}))
			return err
		},
	}
	for name, call := range calls {
		t.Run(name+"/operator 角色被拒", func(t *testing.T) {
			svc, st, _, _ := newWriteHarness(0)
			err := call(svc, withOperatorRole(context.Background(), "operator"))
			if connect.CodeOf(err) != connect.CodePermissionDenied {
				t.Fatalf("operator 角色應 PermissionDenied,got %v", err)
			}
			if got := errorInfoOf(t, err).GetCode(); got != "SYS-4001" {
				t.Fatalf("必須是註冊碼 SYS-4001,got %q", got)
			}
			assertNoWrites(t, st)
		})
		t.Run(name+"/admin 角色可執行", func(t *testing.T) {
			svc, st, _, _ := newWriteHarness(0)
			if err := call(svc, withOperator(context.Background())); err != nil {
				t.Fatalf("admin 角色應可執行: %v", err)
			}
			if len(st.writes.writes) != 1 || len(st.writes.audits) != 1 {
				t.Fatalf("admin 路徑必須寫入並留下稽核: %+v", st.writes)
			}
		})
	}
}

// TestDisableOperatorRejectsSelfAndLastAdmin 驗治理不變式:不得停用自己(當場把自己鎖在門外)、
// 不得停用最後一位 admin(console 全鎖死)。兩者都回 PLAT-3003 且訊息可行動。
func TestDisableOperatorRejectsSelfAndLastAdmin(t *testing.T) {
	t.Run("不得停用自己", func(t *testing.T) {
		svc, st, _, _ := newWriteHarness(0)
		_, err := svc.DisableOperator(withOperator(context.Background()),
			connect.NewRequest(&platformv1.DisableOperatorRequest{
				OperatorId: "42", Reason: "離職"})) // 42 = 呼叫者自己
		if connect.CodeOf(err) != connect.CodeFailedPrecondition {
			t.Fatalf("停用自己應 FailedPrecondition,got %v", err)
		}
		info := errorInfoOf(t, err)
		if info.GetCode() != "PLAT-3003" {
			t.Fatalf("必須是註冊碼 PLAT-3003,got %q", info.GetCode())
		}
		if !strings.Contains(info.GetMessage(), "不得停用自己") {
			t.Fatalf("訊息必須可行動(說出為什麼),got %q", info.GetMessage())
		}
		assertNoWrites(t, st)
	})

	t.Run("不得停用最後一位 admin", func(t *testing.T) {
		svc, st, _, _ := newWriteHarness(0)
		st.disableLastAdmin = true

		_, err := svc.DisableOperator(withOperator(context.Background()),
			connect.NewRequest(&platformv1.DisableOperatorRequest{
				OperatorId: "7", Reason: "離職"}))
		if connect.CodeOf(err) != connect.CodeFailedPrecondition {
			t.Fatalf("停用最後一位 admin 應 FailedPrecondition,got %v", err)
		}
		info := errorInfoOf(t, err)
		if info.GetCode() != "PLAT-3003" {
			t.Fatalf("必須是註冊碼 PLAT-3003,got %q", info.GetCode())
		}
		if !strings.Contains(info.GetMessage(), "最後一位 admin") ||
			!strings.Contains(info.GetMessage(), "新增另一位") {
			t.Fatalf("訊息必須可行動(說出怎麼做),got %q", info.GetMessage())
		}
		assertNoWrites(t, st)
	})
}

// TestWriteRPCPayloadGuards 驗兩個「缺少守衛就會造成破壞性營運狀態」的輸入:
//   - 負的 limit_value 在判定層等於「任何用量都超額」→ 該租戶/該方案的功能被永久關掉,
//     而 console 顯示「已設定限額」;
//   - 兩處都要在寫入**之前**拒絕(否則會留下一筆已改壞的權益)。
func TestWriteRPCsRejectNegativeLimit(t *testing.T) {
	calls := map[string]func(svc *PlatformAdminService) error{
		"SetTenantOverride": func(svc *PlatformAdminService) error {
			_, err := svc.SetTenantOverride(withOperator(context.Background()),
				connect.NewRequest(&platformv1.SetTenantOverrideRequest{
					CompanyId: "42", FeatureCode: entitlements.LimitSeats, LimitSet: true, LimitValue: -1,
					Owner: "業務", Reason: "誤填"}))
			return err
		},
		"SetPlanEntitlement": func(svc *PlatformAdminService) error {
			_, err := svc.SetPlanEntitlement(withOperator(context.Background()),
				connect.NewRequest(&platformv1.SetPlanEntitlementRequest{
					PlanCode: "std", FeatureCode: entitlements.LimitSeats, Enabled: true,
					LimitSet: true, LimitValue: -1, Reason: "誤填"}))
			return err
		},
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			svc, st, _, _ := newWriteHarness(0)
			err := call(svc)
			if connect.CodeOf(err) != connect.CodeInvalidArgument {
				t.Fatalf("負的限額應 InvalidArgument,got %v", err)
			}
			info := errorInfoOf(t, err)
			if info.GetCode() != "SYS-1001" || info.GetDetails()["field"] != "limit_value" {
				t.Fatalf("必須是 SYS-1001 且指出 limit_value,got %q %v", info.GetCode(), info.GetDetails())
			}
			assertNoWrites(t, st)
		})
	}
}

// TestPlatformWriteRejectsInvalidArguments 驗各寫入 RPC 的參數驗證:SYS-1001 且不寫任何東西。
func TestPlatformWriteRejectsInvalidArguments(t *testing.T) {
	ctx := withOperator(context.Background())
	cases := map[string]func(svc *PlatformAdminService) error{
		"company_id 非數字": func(svc *PlatformAdminService) error {
			_, err := svc.RecordPayment(ctx, connect.NewRequest(&platformv1.RecordPaymentRequest{
				CompanyId: "abc", Reason: "收款"}))
			return err
		},
		"金額非數字": func(svc *PlatformAdminService) error {
			_, err := svc.RecordPayment(ctx, connect.NewRequest(&platformv1.RecordPaymentRequest{
				CompanyId: "42", Amount: "一千", Reason: "收款"}))
			return err
		},
		"席位為 0": func(svc *PlatformAdminService) error {
			_, err := svc.SetSeatCount(ctx, connect.NewRequest(&platformv1.SetSeatCountRequest{
				CompanyId: "42", SeatCount: 0, Reason: "降席位"}))
			return err
		},
		"plan_code 為空": func(svc *PlatformAdminService) error {
			_, err := svc.ChangePlan(ctx, connect.NewRequest(&platformv1.ChangePlanRequest{
				CompanyId: "42", Reason: "換方案"}))
			return err
		},
		"override 兩個維度都沒給": func(svc *PlatformAdminService) error {
			_, err := svc.SetTenantOverride(ctx, connect.NewRequest(&platformv1.SetTenantOverrideRequest{
				CompanyId: "42", FeatureCode: entitlements.LimitSeats, Owner: "業務", Reason: "承諾"}))
			return err
		},
		"override 沒有承諾者": func(svc *PlatformAdminService) error {
			_, err := svc.SetTenantOverride(ctx, connect.NewRequest(&platformv1.SetTenantOverrideRequest{
				CompanyId: "42", FeatureCode: entitlements.LimitSeats, LimitSet: true, LimitValue: 5,
				Reason: "承諾"}))
			return err
		},
		"價目週期未知": func(svc *PlatformAdminService) error {
			_, err := svc.UpsertPlanPrice(ctx, connect.NewRequest(&platformv1.UpsertPlanPriceRequest{
				PlanCode: "std", BillingCycle: "weekly", BasePrice: "1.00", SeatPrice: "1.00",
				Reason: "調價"}))
			return err
		},
		"價目金額為空": func(svc *PlatformAdminService) error {
			_, err := svc.UpsertPlanPrice(ctx, connect.NewRequest(&platformv1.UpsertPlanPriceRequest{
				PlanCode: "std", BillingCycle: "monthly", SeatPrice: "1.00", Reason: "調價"}))
			return err
		},
		"設定值非數字": func(svc *PlatformAdminService) error {
			_, err := svc.UpdateBillingSettings(ctx, connect.NewRequest(&platformv1.UpdateBillingSettingsRequest{
				Settings: []*platformv1.BillingSetting{{Key: "grace_days", Value: "abc"}}, Reason: "調整"}))
			return err
		},
		"設定鍵不在允許清單": func(svc *PlatformAdminService) error {
			_, err := svc.UpdateBillingSettings(ctx, connect.NewRequest(&platformv1.UpdateBillingSettingsRequest{
				Settings: []*platformv1.BillingSetting{
					{Key: "system_actor_user_id", Value: "1"}}, Reason: "調整"}))
			return err
		},
		"operator email 不合法": func(svc *PlatformAdminService) error {
			_, err := svc.CreateOperator(ctx, connect.NewRequest(&platformv1.CreateOperatorRequest{
				Email: "not-an-email", Reason: "新增"}))
			return err
		},
		"operator role 未知": func(svc *PlatformAdminService) error {
			_, err := svc.CreateOperator(ctx, connect.NewRequest(&platformv1.CreateOperatorRequest{
				Email: "a@b.com", Role: "root", Reason: "新增"}))
			return err
		},
	}
	for name, call := range cases {
		t.Run(name, func(t *testing.T) {
			svc, st, _, _ := newWriteHarness(0)
			err := call(svc)
			if connect.CodeOf(err) != connect.CodeInvalidArgument {
				t.Fatalf("應回 InvalidArgument,got %v", err)
			}
			if got := errorInfoOf(t, err).GetCode(); got != "SYS-1001" {
				t.Fatalf("必須是註冊碼 SYS-1001,got %q", got)
			}
			assertNoWrites(t, st)
		})
	}
}

// TestUpdateBillingSettingsRejectsUnknownKeyBeforeWrite 驗鍵的驗證發生在**寫入之前**:
// 打錯的鍵會被 upsert 成一個永遠讀不到的新列,而 console 顯示「已儲存」。
func TestUpdateBillingSettingsRejectsUnknownKeyBeforeWrite(t *testing.T) {
	svc, st, _, _ := newWriteHarness(0)
	_, err := svc.UpdateBillingSettings(withOperator(context.Background()),
		connect.NewRequest(&platformv1.UpdateBillingSettingsRequest{
			Settings: []*platformv1.BillingSetting{
				{Key: "grace_days", Value: "3"}, {Key: "grace_dayz", Value: "5"}}, Reason: "調整"}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("未知鍵應 InvalidArgument,got %v", err)
	}
	if len(st.writes.writes) != 0 {
		t.Fatalf("整批拒絕:一個鍵不合法時不得寫入任何一鍵,got %+v", st.writes.writes)
	}
}

// TestGetBillingSettingsListsKnownKeys 驗設定頁的鍵清單固定(缺鍵照樣列出,值為空字串):
// 「這一項還沒設定」與「這一項不存在」在 console 上是兩件事。
func TestGetBillingSettingsListsKnownKeys(t *testing.T) {
	svc, _, _, _ := newWriteHarness(0)
	resp, err := svc.GetBillingSettings(withOperator(context.Background()),
		connect.NewRequest(&platformv1.GetBillingSettingsRequest{}))
	if err != nil {
		t.Fatalf("GetBillingSettings: %v", err)
	}
	if len(resp.Msg.GetSettings()) != 3 {
		t.Fatalf("應列出三個營運參數,got %d", len(resp.Msg.GetSettings()))
	}
	got := map[string]string{}
	for _, s := range resp.Msg.GetSettings() {
		got[s.GetKey()] = s.GetValue()
	}
	if got["trial_days"] != "14" || got["grace_days"] != "7" || got["lead_days"] != "14" {
		t.Fatalf("設定值映射錯誤: %v", got)
	}
	if _, ok := got["system_actor_user_id"]; ok {
		t.Fatal("系統 actor 不得出現在可編輯的營運參數清單(開放編輯它等於權限提升的路徑)")
	}
}

// TestListReceivablesMapsRows 驗待收款清單的投影(金額已是兩位小數字串、到期日為 RFC3339、
// 分頁回音為正規化後的值)。
func TestListReceivablesMapsRows(t *testing.T) {
	svc, st, _, _ := newWriteHarness(0)
	end := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	st.receivables = []ReceivableRow{{
		CompanyID: "42", CompanyName: "甲公司", PlanCode: "std", PeriodNo: 2,
		Amount: "2600.00", PeriodEnd: end, Status: "open"}}

	resp, err := svc.ListReceivables(withOperator(context.Background()),
		connect.NewRequest(&platformv1.ListReceivablesRequest{}))
	if err != nil {
		t.Fatalf("ListReceivables: %v", err)
	}
	if len(resp.Msg.GetRows()) != 1 {
		t.Fatalf("應回一列,got %d", len(resp.Msg.GetRows()))
	}
	row := resp.Msg.GetRows()[0]
	if row.GetAmount() != "2600.00" || row.GetPeriodEnd() != "2026-09-30T00:00:00Z" ||
		row.GetCompanyName() != "甲公司" || row.GetStatus() != "open" {
		t.Fatalf("投影錯誤: %+v", row)
	}
	if resp.Msg.GetPagination().GetPage() != 1 || resp.Msg.GetPagination().GetPageSize() != defaultPageSize {
		t.Fatalf("分頁必須回正規化後的值: %+v", resp.Msg.GetPagination())
	}
}

// assertNoWrites 斷言假 store 上沒有任何資料寫入與稽核。
func assertNoWrites(t *testing.T, st *fakePlatformStore) {
	t.Helper()
	if len(st.writes.writes) != 0 {
		t.Fatalf("不得寫入任何資料,got %+v", st.writes.writes)
	}
	if len(st.writes.audits) != 0 {
		t.Fatalf("不得寫入任何稽核,got %+v", st.writes.audits)
	}
	if st.writes.commits != 0 {
		t.Fatalf("不得提交任何交易,got %d", st.writes.commits)
	}
}

package services

import (
	"context"
	"strconv"
	"testing"

	"connectrpc.com/connect"
	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent/fileasset"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// TestDeliveryTransitionTable 狀態機表:合法/非法轉移一次講清楚(10.6)。
// 這是本檔的核心契約 —— 表若被改動,這裡必須跟著紅。
func TestDeliveryTransitionTable(t *testing.T) {
	cases := []struct {
		from, to string
		want     bool
	}{
		{DeliveryStatusPending, DeliveryStatusInProgress, true},
		{DeliveryStatusPending, DeliveryStatusCancelled, true},
		{DeliveryStatusInProgress, DeliveryStatusCompleted, true},
		{DeliveryStatusInProgress, DeliveryStatusCancelled, true},
		{DeliveryStatusPending, DeliveryStatusCompleted, false},    // 不可跳過開始
		{DeliveryStatusCompleted, DeliveryStatusInProgress, false}, // 終態無出口
		{DeliveryStatusCompleted, DeliveryStatusCancelled, false},
		{DeliveryStatusCancelled, DeliveryStatusInProgress, false},
		{DeliveryStatusCancelled, DeliveryStatusCompleted, false},
	}
	for _, c := range cases {
		if got := CanDeliveryTransition(c.from, c.to); got != c.want {
			t.Errorf("CanDeliveryTransition(%s→%s) = %v, want %v", c.from, c.to, got, c.want)
		}
	}
}

// setupAssignedDelivery 建一位司機與一筆已指派配送,回傳(dept_admin client, 司機 client, 司機 driver id, 配送 id)。
func setupAssignedDelivery(t *testing.T) (*logisticsFixture, int, string, string) {
	t.Helper()
	ctx := context.Background()
	f := openLogisticsFixture(t)
	admin := f.newLogisticsServer(t, logisticsIdentity("dept_admin", f.managerID, f.coID, f.deptID))
	drv, err := admin.CreateDriver(ctx, connect.NewRequest(&v1.CreateDriverRequest{
		UserId: uItoa(f.driver1ID), Name: "司機一",
	}))
	if err != nil {
		t.Fatalf("CreateDriver: %v", err)
	}
	asg, err := admin.AssignDelivery(ctx, connect.NewRequest(&v1.AssignDeliveryRequest{
		RouteId: uItoa(f.routeID), DriverId: drv.Msg.GetDriver().GetId(), Version: "0",
	}))
	if err != nil {
		t.Fatalf("AssignDelivery: %v", err)
	}
	return f, f.driver1ID, drv.Msg.GetDriver().GetId(), asg.Msg.GetDelivery().GetId()
}

// TestDeliveryStateMachineHappyPath 開始 → 完成:狀態、時間戳、事件軌跡(10.6 驗收)。
func TestDeliveryStateMachineHappyPath(t *testing.T) {
	ctx := context.Background()
	f, driverUID, _, delID := setupAssignedDelivery(t)
	driver := f.newLogisticsServer(t, logisticsIdentity("staff", driverUID, f.coID, f.deptID))

	started, err := driver.StartDelivery(ctx, connect.NewRequest(&v1.StartDeliveryRequest{
		DeliveryId: delID, Version: "1",
	}))
	if err != nil {
		t.Fatalf("StartDelivery: %v", err)
	}
	if got := started.Msg.GetDelivery().GetStatus(); got != DeliveryStatusInProgress {
		t.Fatalf("開始後狀態應為 in_progress,got %s", got)
	}
	if started.Msg.GetDelivery().GetStartedAt() == "" {
		t.Fatal("開始後應寫 started_at")
	}

	done, err := driver.CompleteDelivery(ctx, connect.NewRequest(&v1.CompleteDeliveryRequest{
		DeliveryId: delID, Version: "2",
	}))
	if err != nil {
		t.Fatalf("CompleteDelivery: %v", err)
	}
	if got := done.Msg.GetDelivery().GetStatus(); got != DeliveryStatusCompleted {
		t.Fatalf("完成後狀態應為 completed,got %s", got)
	}
	if done.Msg.GetDelivery().GetCompletedAt() == "" {
		t.Fatal("完成後應寫 completed_at")
	}

	// 事件軌跡:started + completed 皆在(append-only 軌跡,10.6)。
	events := f.db.LogisticsDeliveryEvent.Query().AllX(ctx)
	kinds := map[string]int{}
	for _, e := range events {
		kinds[e.EventType]++
	}
	if kinds[DeliveryEventStarted] != 1 || kinds[DeliveryEventCompleted] != 1 {
		t.Fatalf("事件軌跡應各有 1 筆 started/completed,got %v", kinds)
	}
}

// TestDeliverySelfGate 非本人不得操作(10.6 驗收:permission_denied);
// 且非司機身分(無 logistics_drivers 列)亦被拒。
func TestDeliverySelfGate(t *testing.T) {
	ctx := context.Background()
	f, _, _, delID := setupAssignedDelivery(t)

	// 另一位司機(有 drivers 列,但非本配送的指派對象)。
	admin := f.newLogisticsServer(t, logisticsIdentity("dept_admin", f.managerID, f.coID, f.deptID))
	if _, err := admin.CreateDriver(ctx, connect.NewRequest(&v1.CreateDriverRequest{
		UserId: uItoa(f.driver2ID), Name: "司機二",
	})); err != nil {
		t.Fatalf("CreateDriver(2): %v", err)
	}
	other := f.newLogisticsServer(t, logisticsIdentity("staff", f.driver2ID, f.coID, f.deptID))
	_, err := other.StartDelivery(ctx, connect.NewRequest(&v1.StartDeliveryRequest{
		DeliveryId: delID, Version: "1",
	}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("非被指派司機應 permission_denied,got %v", err)
	}

	// 無 drivers 列的一般員工(10.12 閘門①)。
	plainUID := f.db.User.Create().SetEmail("plain@test").SetName("plain").SetRole("staff").
		SetStatus("active").SetPasswordHash("x").SetCompanyID(f.coID).SetDepartmentID(f.deptID).
		SaveX(ctx).ID
	plain := f.newLogisticsServer(t, logisticsIdentity("staff", plainUID, f.coID, f.deptID))
	_, err = plain.StartDelivery(ctx, connect.NewRequest(&v1.StartDeliveryRequest{
		DeliveryId: delID, Version: "1",
	}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("非司機身分應 permission_denied,got %v", err)
	}
}

// TestDeliveryIllegalTransitionAndVersion 非法轉移與樂觀鎖:跳過開始直接完成、
// 舊 version、重複開始,皆 invalid_argument(前端重查後重送)。
func TestDeliveryIllegalTransitionAndVersion(t *testing.T) {
	ctx := context.Background()
	f, driverUID, _, delID := setupAssignedDelivery(t)
	driver := f.newLogisticsServer(t, logisticsIdentity("staff", driverUID, f.coID, f.deptID))

	// 跳過開始直接完成(pending → completed 不在表上)。
	if _, err := driver.CompleteDelivery(ctx, connect.NewRequest(&v1.CompleteDeliveryRequest{
		DeliveryId: delID, Version: "1",
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("pending→completed 應 invalid_argument,got %v", err)
	}
	// 舊 version。
	if _, err := driver.StartDelivery(ctx, connect.NewRequest(&v1.StartDeliveryRequest{
		DeliveryId: delID, Version: "99",
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("版本不符應 invalid_argument,got %v", err)
	}
	// 正常開始後重複開始(in_progress → in_progress 不在表上)。
	if _, err := driver.StartDelivery(ctx, connect.NewRequest(&v1.StartDeliveryRequest{
		DeliveryId: delID, Version: "1",
	})); err != nil {
		t.Fatalf("首次開始: %v", err)
	}
	if _, err := driver.StartDelivery(ctx, connect.NewRequest(&v1.StartDeliveryRequest{
		DeliveryId: delID, Version: "2",
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("重複開始應 invalid_argument,got %v", err)
	}
	// 取消必填原因。
	if _, err := driver.CancelDelivery(ctx, connect.NewRequest(&v1.CancelDeliveryRequest{
		DeliveryId: delID, Version: "2",
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("缺原因應 invalid_argument,got %v", err)
	}
}

// TestDeliveryCompleteWithPOD 完成附 POD:檔案歸屬驗證、逐筆寫入、
// 非本配送的檔案被拒(10.6 驗收:完成 + 上傳 photo POD 成功)。
func TestDeliveryCompleteWithPOD(t *testing.T) {
	ctx := context.Background()
	f, driverUID, _, delID := setupAssignedDelivery(t)
	driver := f.newLogisticsServer(t, logisticsIdentity("staff", driverUID, f.coID, f.deptID))

	mkAsset := func(ownerType string, ownerID int) int {
		return f.db.FileAsset.Create().
			SetCompanyID(f.coID).SetDepartmentID(f.deptID).
			SetOwnerType(ownerType).SetOwnerID(ownerID).
			SetFilename("f.jpg").SetOriginalFilename("photo.jpg").
			SetMimeType("image/jpeg").SetSizeBytes(10).
			SetStoragePath("x/f.jpg").SetURL("/api/v1/files/1/download").
			SetCreatedBy(driverUID).SaveX(ctx).ID
	}
	mine := mkAsset(logisticsDeliveryOwnerType, atoiT(t, delID))

	if _, err := driver.StartDelivery(ctx, connect.NewRequest(&v1.StartDeliveryRequest{
		DeliveryId: delID, Version: "1",
	})); err != nil {
		t.Fatalf("StartDelivery: %v", err)
	}

	// 他人檔案(owner 指向別的配送)不得冒充簽收。
	foreign := mkAsset(logisticsDeliveryOwnerType, 99999)
	_, err := driver.CompleteDelivery(ctx, connect.NewRequest(&v1.CompleteDeliveryRequest{
		DeliveryId: delID, Version: "2",
		Proofs: []*v1.CompleteProof{{ProofType: "photo", FileAssetId: uItoa(foreign)}},
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("非本配送檔案應 invalid_argument,got %v", err)
	}

	// 型別白名單。
	_, err = driver.CompleteDelivery(ctx, connect.NewRequest(&v1.CompleteDeliveryRequest{
		DeliveryId: delID, Version: "2",
		Proofs: []*v1.CompleteProof{{ProofType: "video", FileAssetId: uItoa(mine)}},
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("非白名單型別應 invalid_argument,got %v", err)
	}

	// 正常完成 + photo POD。
	resp, err := driver.CompleteDelivery(ctx, connect.NewRequest(&v1.CompleteDeliveryRequest{
		DeliveryId: delID, Version: "2",
		Proofs: []*v1.CompleteProof{{ProofType: "photo", FileAssetId: uItoa(mine), Remarks: "已簽收"}},
	}))
	if err != nil {
		t.Fatalf("CompleteDelivery: %v", err)
	}
	if len(resp.Msg.GetProofs()) != 1 {
		t.Fatalf("應回 1 筆 POD,got %d", len(resp.Msg.GetProofs()))
	}
	p := resp.Msg.GetProofs()[0]
	if p.GetProofType() != "photo" || p.GetFileAssetId() != uItoa(mine) {
		t.Fatalf("POD 內容不符: %+v", p)
	}
	// 失敗路徑不留半套:先前兩次被拒的完成嘗試不得留下任何 POD 列。
	// (兩次皆在寫入前被擋,故資料庫仍只有成功那筆。)
	if n := f.db.LogisticsProof.Query().CountX(ctx); n != 1 {
		t.Fatalf("POD 應恰有 1 筆(被拒的嘗試不留痕),got %d", n)
	}
	// 檔案確實屬於本配送(避免測到別的 owner)。
	if fa, err := f.db.FileAsset.Query().Where(fileasset.ID(mine)).Only(ctx); err == nil {
		if fa.OwnerType != logisticsDeliveryOwnerType || fa.OwnerID != atoiT(t, delID) {
			t.Fatalf("測試前置錯誤:檔案 owner 非本配送")
		}
	}
	_ = auth.RolesFor // 保持 auth import 於本檔語意(身分建構走 logisticsIdentity)
}

// atoiT 測試用:字串 id → int(id 皆為測試內產生的合法整數字串)。
func atoiT(t *testing.T, s string) int {
	t.Helper()
	n, err := strconv.Atoi(s)
	if err != nil {
		t.Fatalf("atoiT(%q): %v", s, err)
	}
	return n
}

// TestAssignDeliveryGuardsTerminalStatus 終態配送不得重指派(10.6 狀態機補上的守衛):
// 司機/車輛欄位與 FGA 判決不該在配送完成後再被改動。
func TestAssignDeliveryGuardsTerminalStatus(t *testing.T) {
	ctx := context.Background()
	f, driverUID, _, _ := setupAssignedDelivery(t)
	admin := f.newLogisticsServer(t, logisticsIdentity("dept_admin", f.managerID, f.coID, f.deptID))
	driver := f.newLogisticsServer(t, logisticsIdentity("staff", driverUID, f.coID, f.deptID))

	if _, err := driver.StartDelivery(ctx, connect.NewRequest(&v1.StartDeliveryRequest{
		DeliveryId: "1", Version: "1",
	})); err != nil {
		t.Fatalf("開始: %v", err)
	}
	if _, err := driver.CompleteDelivery(ctx, connect.NewRequest(&v1.CompleteDeliveryRequest{
		DeliveryId: "1", Version: "2",
	})); err != nil {
		t.Fatalf("完成: %v", err)
	}
	// 已完成 → 重指派應被拒(即使 version 正確)。
	if _, err := admin.AssignDelivery(ctx, connect.NewRequest(&v1.AssignDeliveryRequest{
		RouteId: uItoa(f.routeID), DriverId: "1", Version: "3",
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("已完成配送重指派應 invalid_argument,got %v", err)
	}
	// 取消後同理。
	f2, driverUID2, _, _ := setupAssignedDelivery(t)
	admin2 := f2.newLogisticsServer(t, logisticsIdentity("dept_admin", f2.managerID, f2.coID, f2.deptID))
	driver2 := f2.newLogisticsServer(t, logisticsIdentity("staff", driverUID2, f2.coID, f2.deptID))
	if _, err := driver2.CancelDelivery(ctx, connect.NewRequest(&v1.CancelDeliveryRequest{
		DeliveryId: "1", Version: "1", Reason: "客戶臨時取消",
	})); err != nil {
		t.Fatalf("取消: %v", err)
	}
	if _, err := admin2.AssignDelivery(ctx, connect.NewRequest(&v1.AssignDeliveryRequest{
		RouteId: uItoa(f2.routeID), DriverId: "1", Version: "2",
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("已取消配送重指派應 invalid_argument,got %v", err)
	}
}

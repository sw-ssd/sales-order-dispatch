// Package services 的 fleet tuple 同步(D32/10.8)。
//
// OpenFGA tuple 寫入是外部副作用 → 一律 AfterCommit(慣例:副作用不得在業務交易內)。
// 本檔的「outbox 語意」取簡化形:desired 集合以 DB 值為準**全量對帳**(同 Provision
// 慣例,見 internal/authz/provision.go),同步失敗只落日誌不回滾業務 —— 指派事實留在
// DB,重指派/重建即重新對帳(不做佇列重放:tuple 可由單一權威來源重算,重放價值低)。
// managed relation 白名單防止誤刪非本模組寫入的 tuple。
package services

import (
	"context"
	"log"
	"strconv"

	authzopenfga "github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
)

// managedFleetRelations 為本模組管理的 relation 集合(reconcile 只動這些)。
var managedFleetRelations = map[string]bool{
	"company": true, "department": true, "manager": true,
	"driver": true, "assigned_by": true, "assignee": true,
}

// fleetObj / driverObj 為 tuple 物件定位(D32 model 型別名)。
func fleetObj(id int) string   { return "fleet_delivery:" + strconv.Itoa(id) }
func driverObj(id int) string  { return "driver:" + strconv.Itoa(id) }
func vehicleObj(id int) string { return "vehicle:" + strconv.Itoa(id) }

// companyMemberSubj / departmentMemberSubj 為租戶 parent 的 userset 主體形狀
// (model: company: [company#member] 等 —— subject 必須帶 #member/#admin)。
func companyMemberSubj(cid int) string    { return "company:" + strconv.Itoa(cid) + "#member" }
func departmentMemberSubj(did int) string { return "department:" + strconv.Itoa(did) + "#member" }
func departmentAdminSubj(did int) string  { return "department:" + strconv.Itoa(did) + "#admin" }
func driverAssigneeSubj(did int) string   { return "driver:" + strconv.Itoa(did) + "#assignee" }

// deliveryDesiredTuples 計算一筆配送的期望 tuple 集合(租戶 parent + 管理邊 + 指派 + 指派人)。
func deliveryDesiredTuples(deliveryID, cid int, did *int, driverID *int, actor int) [][3]string {
	obj := fleetObj(deliveryID)
	desired := [][3]string{{companyMemberSubj(cid), "company", obj}}
	if did != nil {
		desired = append(desired,
			[3]string{departmentMemberSubj(*did), "department", obj},
			[3]string{departmentAdminSubj(*did), "manager", obj},
		)
	}
	if driverID != nil {
		desired = append(desired, [3]string{driverAssigneeSubj(*driverID), "driver", obj})
	}
	desired = append(desired, [3]string{"user:" + strconv.Itoa(actor), "assigned_by", obj})
	return desired
}

// driverDesiredTuples 司機物件的期望集合(租戶 parent + 被指派人)。
func driverDesiredTuples(driverID, cid int, did *int, userID int) [][3]string {
	obj := driverObj(driverID)
	desired := [][3]string{
		{companyMemberSubj(cid), "company", obj},
		{"user:" + strconv.Itoa(userID), "assignee", obj},
	}
	if did != nil {
		desired = append(desired, [3]string{departmentMemberSubj(*did), "department", obj})
	}
	return desired
}

// vehicleDesiredTuples 車輛物件的期望集合(僅租戶 parent)。
func vehicleDesiredTuples(vehicleID, cid int, did *int) [][3]string {
	obj := vehicleObj(vehicleID)
	desired := [][3]string{{companyMemberSubj(cid), "company", obj}}
	if did != nil {
		desired = append(desired, [3]string{departmentMemberSubj(*did), "department", obj})
	}
	return desired
}

// reconcileFleetTuples 把 obj 的現存 managed tuples 對帳到 desired(刪 stale、補缺漏)。
func reconcileFleetTuples(ctx context.Context, e *authzopenfga.Engine, obj string, desired [][3]string) error {
	existing, err := e.ListTuples(ctx)
	if err != nil {
		return err
	}
	want := make(map[string][3]string, len(desired))
	for _, d := range desired {
		want[d[0]+"\x00"+d[1]+"\x00"+d[2]] = d
	}
	have := map[string]bool{}
	for _, t := range existing {
		if t[2] != obj || !managedFleetRelations[t[1]] {
			continue
		}
		have[t[0]+"\x00"+t[1]+"\x00"+t[2]] = true
		if _, ok := want[t[0]+"\x00"+t[1]+"\x00"+t[2]]; !ok {
			if err := e.DeleteTuple(ctx, t[0], t[1], t[2]); err != nil {
				return err
			}
		}
	}
	for key, d := range want {
		if have[key] {
			continue
		}
		if err := e.WriteTuple(ctx, d[0], d[1], d[2]); err != nil {
			return err
		}
	}
	return nil
}

// afterCommitFleetTuples 註冊提交後的 tuple 對帳(outbox 簡化語意:失敗僅記日誌)。
// e/obj/desired 皆在提交前取定,closure 只讀快照。
func afterCommitFleetTuples(ctx context.Context, e *authzopenfga.Engine, obj string, desired [][3]string) error {
	return dbtenant.AfterCommit(ctx, func(ctx context.Context) error {
		if err := reconcileFleetTuples(ctx, e, obj, desired); err != nil {
			log.Printf("fleet: tuple 對帳失敗(%s,重指派/重建可修復): %v", obj, err)
		}
		return nil
	})
}

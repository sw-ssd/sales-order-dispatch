package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// FleetDelivery 配送執行單元(D32/10.1):承接 route_id 的指派(粒度 = 車次,
// 10.4:非車次↔車 1:1,可重指派)。version 樂觀鎖(重指派遞增,10.4);
// driver_id/vehicle_id 可空(未指派/部分指派);軟刪除(D10)。
//
// OpenFGA instance 級授權(D32/10.8):建立/更新後經 AfterCommit 同步租戶 parent
// 邊與 driver userset tuple(見 internal/services/fleet_events.go)。
type FleetDelivery struct {
	ent.Schema
}

// Fields of the FleetDelivery.
func (FleetDelivery) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("department_id").
			Optional().
			Nillable(),
		field.Int("route_id"), // 承接車次(routes.id)
		field.Int("driver_id").
			Optional().
			Nillable(), // 指派的司機(drivers.id)
		field.Int("vehicle_id").
			Optional().
			Nillable(), // 指派的車輛(vehicles.id)
		field.Int("assigned_by"), // 操作者(users.id)
		field.String("status").
			Default("pending"), // pending / in_progress / completed / cancelled(10.6 狀態機另案)
		field.Int("version").
			Default(1), // 樂觀鎖(10.4 重指派遞增)
		field.Time("deleted_at").
			Optional().
			Nillable(), // 軟刪除(D10)
		field.Time("created_at").
			Default(mutableNow),
		field.Time("updated_at").
			Default(mutableNow).
			UpdateDefault(mutableNow),
	}
}

// Indexes of the FleetDelivery.
func (FleetDelivery) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("department_id", "route_id"),
		index.Fields("driver_id"),
	}
}

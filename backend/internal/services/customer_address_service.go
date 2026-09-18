// CustomerService 客戶地址簿(04 計畫 3.2.1, D10/D18):List/Add/Update/Delete。
// 地址隸屬客戶並繼承其 company_id / department_id(供 RLS);同類型至多一筆預設(部分唯一索引兜底)。
// 設定預設採「先清同類型其餘預設、再設目標」同一交易;首筆同類型自動為預設;刪除預設不自動遞補。
package services

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/ent/customeraddress"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
)

// validAddressType 判斷地址類型是否為合法列舉值。
func validAddressType(t string) bool {
	switch customeraddress.Type(t) {
	case customeraddress.TypeShipping, customeraddress.TypeBilling, customeraddress.TypeOther:
		return true
	}
	return false
}

// requireCustomer 於交易外以範圍 + 未刪除載入客戶;不存在/跨部門 → not_found。
// 目的:地址/聯絡人複寫客戶的 company_id / department_id,並確保寫入僅限可見範圍。
func (s *CustomerService) requireCustomer(ctx context.Context, cid int, did *int, custID int) (*ent.Customer, error) {
	c, err := customerScopeQuery(s.db.Customer.Query(), cid, did).
		Where(customer.ID(custID), customer.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	return c, nil
}

// addressScopeQuery 依範圍對地址查詢加入 company/department where。
func addressScopeQuery(q *ent.CustomerAddressQuery, cid int, did *int) *ent.CustomerAddressQuery {
	if did != nil {
		return q.Where(customeraddress.CompanyIDEQ(cid), customeraddress.DepartmentIDEQ(*did))
	}
	return q.Where(customeraddress.CompanyIDEQ(cid))
}

// addressToProto 將 ent.CustomerAddress 轉為 proto CustomerAddress。
func addressToProto(a *ent.CustomerAddress) *customersv1.CustomerAddress {
	p := &customersv1.CustomerAddress{
		Id:            strconv.FormatInt(int64(a.ID), 10),
		CustomerId:    strconv.FormatInt(int64(a.CustomerID), 10),
		Type:          string(a.Type),
		RecipientName: a.RecipientName,
		Phone:         a.Phone,
		AddressLine:   a.AddressLine,
		City:          a.City,
		PostalCode:    a.PostalCode,
		IsDefault:     a.IsDefault,
	}
	if !a.CreatedAt.IsZero() {
		p.CreatedAt = a.CreatedAt.Format(time.RFC3339)
	}
	if !a.UpdatedAt.IsZero() {
		p.UpdatedAt = a.UpdatedAt.Format(time.RFC3339)
	}
	if a.DeletedAt != nil {
		p.DeletedAt = a.DeletedAt.Format(time.RFC3339)
	}
	return p
}

// ListAddresses 列客戶地址(限可見範圍;預設排除已刪除)。
func (s *CustomerService) ListAddresses(ctx context.Context, req *connect.Request[customersv1.ListAddressesRequest]) (*connect.Response[customersv1.ListAddressesResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := customerScope(id)
	if err != nil {
		return nil, err
	}
	custID, err := parseID(req.Msg.GetCustomerId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("customer_id 格式錯誤"))
	}
	// 客戶須在範圍內(未刪除)。
	if _, err := s.requireCustomer(ctx, cid, did, custID); err != nil {
		return nil, err
	}
	q := addressScopeQuery(s.db.CustomerAddress.Query(), cid, did).Where(customeraddress.CustomerIDEQ(custID))
	if !req.Msg.GetIncludeDeleted() {
		q = q.Where(customeraddress.DeletedAtIsNil())
	}
	items, err := q.Order(ent.Asc(customeraddress.FieldID)).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	out := make([]*customersv1.CustomerAddress, 0, len(items))
	for _, a := range items {
		out = append(out, addressToProto(a))
	}
	return connect.NewResponse(&customersv1.ListAddressesResponse{Addresses: out}), nil
}

// AddAddress 新增地址:複寫客戶租戶;同類型首筆自動為預設;設為預設時先清同類型其餘預設(同一交易)。
func (s *CustomerService) AddAddress(ctx context.Context, req *connect.Request[customersv1.AddAddressRequest]) (*connect.Response[customersv1.AddAddressResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := customerScope(id)
	if err != nil {
		return nil, err
	}
	custID, err := parseID(req.Msg.GetCustomerId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("customer_id 格式錯誤"))
	}
	c, err := s.requireCustomer(ctx, cid, did, custID)
	if err != nil {
		return nil, err
	}
	typ := strings.TrimSpace(req.Msg.GetType())
	if typ == "" {
		typ = string(customeraddress.TypeOther)
	}
	if !validAddressType(typ) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("type 僅接受 shipping / billing / other"))
	}
	recipient := strings.TrimSpace(req.Msg.GetRecipientName())
	addrLine := strings.TrimSpace(req.Msg.GetAddressLine())
	if recipient == "" || addrLine == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("recipient_name 與 address_line 必填"))
	}
	companyID := c.CompanyID

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	n, err := tx.CustomerAddress.Query().
		Where(customeraddress.CustomerIDEQ(custID), customeraddress.TypeEQ(customeraddress.Type(typ)), customeraddress.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	wantDefault := req.Msg.GetIsDefault() || n == 0 // 首筆同類型自動為預設(3.2.1 步驟 4)
	if wantDefault {
		// 先清同類型其餘預設(3.2.1 步驟 3)。
		if err := tx.CustomerAddress.Update().
			Where(
				customeraddress.CustomerIDEQ(custID),
				customeraddress.TypeEQ(customeraddress.Type(typ)),
				customeraddress.DeletedAtIsNil(),
				customeraddress.IsDefaultEQ(true),
			).SetIsDefault(false).Exec(ctx); err != nil {
			return nil, toConnectError(err)
		}
	}
	actor, _ := parseID(id.UserID)
	build := tx.CustomerAddress.Create().
		SetCompanyID(companyID).
		SetCustomerID(custID).
		SetType(customeraddress.Type(typ)).
		SetRecipientName(recipient).
		SetAddressLine(addrLine).
		SetIsDefault(wantDefault).
		SetCreatedBy(actor).
		SetUpdatedBy(actor)
	if c.DepartmentID != nil {
		build = build.SetDepartmentID(*c.DepartmentID)
	}
	if p := strings.TrimSpace(req.Msg.GetPhone()); p != "" {
		build = build.SetPhone(p)
	}
	if city := strings.TrimSpace(req.Msg.GetCity()); city != "" {
		build = build.SetCity(city)
	}
	if pc := strings.TrimSpace(req.Msg.GetPostalCode()); pc != "" {
		build = build.SetPostalCode(pc)
	}
	created, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action:       "create",
		ResourceType: "customer_address",
		ResourceID:   strconv.FormatInt(int64(created.ID), 10),
		CompanyID:    companyID,
		DepartmentID: c.DepartmentID,
		UserID:       actor,
		After:        map[string]any{"customer_id": custID, "type": typ},
		IPAddress:    audit.MetaFrom(ctx).IP,
		UserAgent:    audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&customersv1.AddAddressResponse{Address: addressToProto(created)}), nil
}

// UpdateAddress 欄位式更新地址;設為預設時同一交易先清同類型其餘預設。
func (s *CustomerService) UpdateAddress(ctx context.Context, req *connect.Request[customersv1.UpdateAddressRequest]) (*connect.Response[customersv1.UpdateAddressResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := customerScope(id)
	if err != nil {
		return nil, err
	}
	addrID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("address id 格式錯誤"))
	}
	cur, err := addressScopeQuery(s.db.CustomerAddress.Query(), cid, did).
		Where(customeraddress.ID(addrID), customeraddress.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	upd := tx.CustomerAddress.UpdateOneID(addrID)
	if req.Msg.Type != nil {
		typ := strings.TrimSpace(*req.Msg.Type)
		if !validAddressType(typ) {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("type 僅接受 shipping / billing / other"))
		}
		upd = upd.SetType(customeraddress.Type(typ))
	}
	if req.Msg.RecipientName != nil {
		r := strings.TrimSpace(*req.Msg.RecipientName)
		if r == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("recipient_name 不可為空"))
		}
		upd = upd.SetRecipientName(r)
	}
	if req.Msg.Phone != nil {
		upd = upd.SetPhone(*req.Msg.Phone)
	}
	if req.Msg.AddressLine != nil {
		l := strings.TrimSpace(*req.Msg.AddressLine)
		if l == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("address_line 不可為空"))
		}
		upd = upd.SetAddressLine(l)
	}
	if req.Msg.City != nil {
		upd = upd.SetCity(*req.Msg.City)
	}
	if req.Msg.PostalCode != nil {
		upd = upd.SetPostalCode(*req.Msg.PostalCode)
	}
	if req.Msg.IsDefault != nil && *req.Msg.IsDefault {
		// 目標類型(型別可能一併更新):以最終有效類型清其餘預設。
		effType := cur.Type
		if req.Msg.Type != nil {
			effType = customeraddress.Type(strings.TrimSpace(*req.Msg.Type))
		}
		if err := tx.CustomerAddress.Update().
			Where(
				customeraddress.CustomerIDEQ(cur.CustomerID),
				customeraddress.TypeEQ(effType),
				customeraddress.DeletedAtIsNil(),
				customeraddress.IsDefaultEQ(true),
				customeraddress.IDNEQ(addrID),
			).SetIsDefault(false).Exec(ctx); err != nil {
			return nil, toConnectError(err)
		}
		upd = upd.SetIsDefault(true)
	} else if req.Msg.IsDefault != nil {
		upd = upd.SetIsDefault(false)
	}
	actor, _ := parseID(id.UserID)
	upd = upd.SetUpdatedBy(actor)

	updated, err := upd.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action:       "update",
		ResourceType: "customer_address",
		ResourceID:   strconv.FormatInt(int64(addrID), 10),
		CompanyID:    cur.CompanyID,
		DepartmentID: cur.DepartmentID,
		UserID:       actor,
		After:        map[string]any{"customer_id": cur.CustomerID},
		IPAddress:    audit.MetaFrom(ctx).IP,
		UserAgent:    audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&customersv1.UpdateAddressResponse{Address: addressToProto(updated)}), nil
}

// DeleteAddress 軟刪除地址 + 稽核(同一交易);刪除預設不自動遞補。
func (s *CustomerService) DeleteAddress(ctx context.Context, req *connect.Request[customersv1.DeleteAddressRequest]) (*connect.Response[customersv1.DeleteAddressResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := customerScope(id)
	if err != nil {
		return nil, err
	}
	addrID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("address id 格式錯誤"))
	}
	cur, err := addressScopeQuery(s.db.CustomerAddress.Query(), cid, did).
		Where(customeraddress.ID(addrID), customeraddress.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	actor, _ := parseID(id.UserID)
	if err := tx.CustomerAddress.UpdateOneID(addrID).SetDeletedAt(time.Now().UTC()).SetUpdatedBy(actor).Exec(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action:       "delete",
		ResourceType: "customer_address",
		ResourceID:   strconv.FormatInt(int64(addrID), 10),
		CompanyID:    cur.CompanyID,
		DepartmentID: cur.DepartmentID,
		UserID:       actor,
		Before:       map[string]any{"customer_id": cur.CustomerID, "type": string(cur.Type)},
		IPAddress:    audit.MetaFrom(ctx).IP,
		UserAgent:    audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&customersv1.DeleteAddressResponse{}), nil
}

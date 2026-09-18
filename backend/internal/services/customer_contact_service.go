// CustomerService 客戶聯絡人(04 計畫 3.2.2, D10/D18):List/Add/Update/Delete。
// 聯絡人隸屬客戶並繼承其 company_id / department_id;每客戶至多一筆預設(不分類型)。
// 設定預設採「先清其餘預設、再設目標」同一交易;首筆自動為預設;刪除預設不自動遞補。
package services

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/customercontact"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
)

// contactEmailRe 為輕量 email 格式驗證(3.2.2 步驟 4:僅格式,不做投遞驗證)。
var contactEmailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// contactScopeQuery 依範圍對聯絡人查詢加入 company/department where。
func contactScopeQuery(q *ent.CustomerContactQuery, cid int, did *int) *ent.CustomerContactQuery {
	if did != nil {
		return q.Where(customercontact.CompanyIDEQ(cid), customercontact.DepartmentIDEQ(*did))
	}
	return q.Where(customercontact.CompanyIDEQ(cid))
}

// contactToProto 將 ent.CustomerContact 轉為 proto CustomerContact。
func contactToProto(c *ent.CustomerContact) *customersv1.CustomerContact {
	p := &customersv1.CustomerContact{
		Id:         strconv.FormatInt(int64(c.ID), 10),
		CustomerId: strconv.FormatInt(int64(c.CustomerID), 10),
		Name:       c.Name,
		Title:      c.Title,
		Email:      c.Email,
		Phone:      c.Phone,
		IsDefault:  c.IsDefault,
	}
	if !c.CreatedAt.IsZero() {
		p.CreatedAt = c.CreatedAt.Format(time.RFC3339)
	}
	if !c.UpdatedAt.IsZero() {
		p.UpdatedAt = c.UpdatedAt.Format(time.RFC3339)
	}
	if c.DeletedAt != nil {
		p.DeletedAt = c.DeletedAt.Format(time.RFC3339)
	}
	return p
}

// ListContacts 列客戶聯絡人(限可見範圍;預設排除已刪除)。
func (s *CustomerService) ListContacts(ctx context.Context, req *connect.Request[customersv1.ListContactsRequest]) (*connect.Response[customersv1.ListContactsResponse], error) {
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
	if _, err := s.requireCustomer(ctx, cid, did, custID); err != nil {
		return nil, err
	}
	q := contactScopeQuery(s.db.CustomerContact.Query(), cid, did).Where(customercontact.CustomerIDEQ(custID))
	if !req.Msg.GetIncludeDeleted() {
		q = q.Where(customercontact.DeletedAtIsNil())
	}
	items, err := q.Order(ent.Asc(customercontact.FieldID)).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	out := make([]*customersv1.CustomerContact, 0, len(items))
	for _, c := range items {
		out = append(out, contactToProto(c))
	}
	return connect.NewResponse(&customersv1.ListContactsResponse{Contacts: out}), nil
}

// AddContact 新增聯絡人:複寫客戶租戶;首筆自動為預設;設為預設時先清其餘預設(同一交易)。
func (s *CustomerService) AddContact(ctx context.Context, req *connect.Request[customersv1.AddContactRequest]) (*connect.Response[customersv1.AddContactResponse], error) {
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
	name := strings.TrimSpace(req.Msg.GetName())
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("name 必填"))
	}
	if e := strings.TrimSpace(req.Msg.GetEmail()); e != "" && !contactEmailRe.MatchString(e) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("email 格式非法"))
	}
	companyID := c.CompanyID

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	n, err := tx.CustomerContact.Query().
		Where(customercontact.CustomerIDEQ(custID), customercontact.DeletedAtIsNil()).Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	wantDefault := req.Msg.GetIsDefault() || n == 0 // 首筆聯絡人自動為預設(3.2.2 步驟 3)
	if wantDefault {
		if err := tx.CustomerContact.Update().
			Where(
				customercontact.CustomerIDEQ(custID),
				customercontact.DeletedAtIsNil(),
				customercontact.IsDefaultEQ(true),
			).SetIsDefault(false).Exec(ctx); err != nil {
			return nil, toConnectError(err)
		}
	}
	actor, _ := parseID(id.UserID)
	build := tx.CustomerContact.Create().
		SetCompanyID(companyID).
		SetCustomerID(custID).
		SetName(name).
		SetIsDefault(wantDefault).
		SetCreatedBy(actor).
		SetUpdatedBy(actor)
	if c.DepartmentID != nil {
		build = build.SetDepartmentID(*c.DepartmentID)
	}
	if t := strings.TrimSpace(req.Msg.GetTitle()); t != "" {
		build = build.SetTitle(t)
	}
	if e := strings.TrimSpace(req.Msg.GetEmail()); e != "" {
		build = build.SetEmail(e)
	}
	if p := strings.TrimSpace(req.Msg.GetPhone()); p != "" {
		build = build.SetPhone(p)
	}
	created, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action:       "create",
		ResourceType: "customer_contact",
		ResourceID:   strconv.FormatInt(int64(created.ID), 10),
		CompanyID:    companyID,
		DepartmentID: c.DepartmentID,
		UserID:       actor,
		After:        map[string]any{"customer_id": custID, "name": created.Name},
		IPAddress:    audit.MetaFrom(ctx).IP,
		UserAgent:    audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&customersv1.AddContactResponse{Contact: contactToProto(created)}), nil
}

// UpdateContact 欄位式更新聯絡人;設為預設時同一交易先清其餘預設。
func (s *CustomerService) UpdateContact(ctx context.Context, req *connect.Request[customersv1.UpdateContactRequest]) (*connect.Response[customersv1.UpdateContactResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := customerScope(id)
	if err != nil {
		return nil, err
	}
	contactID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("contact id 格式錯誤"))
	}
	cur, err := contactScopeQuery(s.db.CustomerContact.Query(), cid, did).
		Where(customercontact.ID(contactID), customercontact.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	upd := tx.CustomerContact.UpdateOneID(contactID)
	if req.Msg.Name != nil {
		n := strings.TrimSpace(*req.Msg.Name)
		if n == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("name 不可為空"))
		}
		upd = upd.SetName(n)
	}
	if req.Msg.Title != nil {
		upd = upd.SetTitle(*req.Msg.Title)
	}
	if req.Msg.Email != nil {
		e := strings.TrimSpace(*req.Msg.Email)
		if e != "" && !contactEmailRe.MatchString(e) {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("email 格式非法"))
		}
		upd = upd.SetEmail(e)
	}
	if req.Msg.Phone != nil {
		upd = upd.SetPhone(*req.Msg.Phone)
	}
	if req.Msg.IsDefault != nil && *req.Msg.IsDefault {
		if err := tx.CustomerContact.Update().
			Where(
				customercontact.CustomerIDEQ(cur.CustomerID),
				customercontact.DeletedAtIsNil(),
				customercontact.IsDefaultEQ(true),
				customercontact.IDNEQ(contactID),
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
		ResourceType: "customer_contact",
		ResourceID:   strconv.FormatInt(int64(contactID), 10),
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
	return connect.NewResponse(&customersv1.UpdateContactResponse{Contact: contactToProto(updated)}), nil
}

// DeleteContact 軟刪除聯絡人 + 稽核(同一交易);刪除預設不自動遞補。
func (s *CustomerService) DeleteContact(ctx context.Context, req *connect.Request[customersv1.DeleteContactRequest]) (*connect.Response[customersv1.DeleteContactResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := customerScope(id)
	if err != nil {
		return nil, err
	}
	contactID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("contact id 格式錯誤"))
	}
	cur, err := contactScopeQuery(s.db.CustomerContact.Query(), cid, did).
		Where(customercontact.ID(contactID), customercontact.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	actor, _ := parseID(id.UserID)
	if err := tx.CustomerContact.UpdateOneID(contactID).SetDeletedAt(time.Now().UTC()).SetUpdatedBy(actor).Exec(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action:       "delete",
		ResourceType: "customer_contact",
		ResourceID:   strconv.FormatInt(int64(contactID), 10),
		CompanyID:    cur.CompanyID,
		DepartmentID: cur.DepartmentID,
		UserID:       actor,
		Before:       map[string]any{"customer_id": cur.CustomerID, "name": cur.Name},
		IPAddress:    audit.MetaFrom(ctx).IP,
		UserAgent:    audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&customersv1.DeleteContactResponse{}), nil
}

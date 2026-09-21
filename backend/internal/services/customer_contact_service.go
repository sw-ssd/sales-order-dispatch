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
	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/ent/customercontact"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/customers/qrcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
)

// qrSecret 取 QR 簽章密鑰(JWT_SECRET 複用)。預設 dev 值;server 組裝時以
// config.Auth.JWTSecret 呼叫 SetQRSecret 覆寫(正式環境由 Server.Init 防護)。
// 包級變數而非構造子參數:CustomerService 已有 10+ 測試呼叫點,改簽名全數陪葬。
var qrSecretValue = "dev-only-jwt-secret-change-me"

// SetQRSecret 設定 QR 簽章密鑰(僅 server 組裝鏈呼叫)。
func SetQRSecret(s string) {
	if s != "" {
		qrSecretValue = s
	}
}

func qrSecret() string { return qrSecretValue }

// qrDeepLink 組深層連結(App 未裝導商店、已裝直開)。
func qrDeepLink(base, token string) string {
	return strings.TrimRight(base, "/") + "/customer_account_qrcode/" + token
}

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
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
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
	q := contactScopeQuery(dbtenant.Client(ctx, s.db).CustomerContact.Query(), cid, did).Where(customercontact.CustomerIDEQ(custID))
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
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
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

	// 由 ctx 取請求交易:稽核寫入需要 *ent.Tx(audit.Record 的簽章),且 D18「業務寫入與稽核
	// 同一交易」正是靠它維持。無請求交易(CLI/seed/未掛 interceptor 的路徑)即回明確錯誤,
	// 不得默默退回 fallback client 寫入 —— ENABLE+FORCE 後那會靜默漏掉租戶範圍。
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, errors.New("缺少租戶交易(context)"))
	}
	db := tx.Client() // 查詢與寫入都用它;同一個交易

	n, err := db.CustomerContact.Query().
		Where(customercontact.CustomerIDEQ(custID), customercontact.DeletedAtIsNil()).Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	wantDefault := req.Msg.GetIsDefault() || n == 0 // 首筆聯絡人自動為預設(3.2.2 步驟 3)
	if wantDefault {
		if err := db.CustomerContact.Update().
			Where(
				customercontact.CustomerIDEQ(custID),
				customercontact.DeletedAtIsNil(),
				customercontact.IsDefaultEQ(true),
			).SetIsDefault(false).Exec(ctx); err != nil {
			return nil, toConnectError(err)
		}
	}
	actor, _ := parseID(id.UserID)
	build := db.CustomerContact.Create().
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
	if err := recordAudit(ctx, tx, "customer_contact", "create", created.ID, companyID, c.DepartmentID, actor, map[string]any{"customer_id": custID, "name": created.Name}); err != nil {
		return nil, toConnectError(err)
	}
	// 交易由 interceptor 擁有(成功即 commit),服務層不再 commit/rollback。
	return connect.NewResponse(&customersv1.AddContactResponse{Contact: contactToProto(created)}), nil
}

// UpdateContact 欄位式更新聯絡人;設為預設時同一交易先清其餘預設。
func (s *CustomerService) UpdateContact(ctx context.Context, req *connect.Request[customersv1.UpdateContactRequest]) (*connect.Response[customersv1.UpdateContactResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	contactID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("contact id 格式錯誤"))
	}
	cur, err := contactScopeQuery(dbtenant.Client(ctx, s.db).CustomerContact.Query(), cid, did).
		Where(customercontact.ID(contactID), customercontact.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}

	// 取請求交易:查詢/寫入用 db,稽核續用 tx(同一交易,D18);理由見本檔 AddContact。
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, errors.New("缺少租戶交易(context)"))
	}
	db := tx.Client()

	upd := db.CustomerContact.UpdateOneID(contactID)
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
		if err := db.CustomerContact.Update().
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
	if err := recordAudit(ctx, tx, "customer_contact", "update", contactID, cur.CompanyID, cur.DepartmentID, actor, map[string]any{"customer_id": cur.CustomerID}); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&customersv1.UpdateContactResponse{Contact: contactToProto(updated)}), nil
}

// DeleteContact 軟刪除聯絡人 + 稽核(同一交易);刪除預設不自動遞補。
func (s *CustomerService) DeleteContact(ctx context.Context, req *connect.Request[customersv1.DeleteContactRequest]) (*connect.Response[customersv1.DeleteContactResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	contactID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("contact id 格式錯誤"))
	}
	cur, err := contactScopeQuery(dbtenant.Client(ctx, s.db).CustomerContact.Query(), cid, did).
		Where(customercontact.ID(contactID), customercontact.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	// 取請求交易:查詢/寫入用 db,稽核續用 tx(同一交易,D18);理由見本檔 AddContact。
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, errors.New("缺少租戶交易(context)"))
	}
	db := tx.Client()
	actor, _ := parseID(id.UserID)
	if err := db.CustomerContact.UpdateOneID(contactID).SetDeletedAt(time.Now().UTC()).SetUpdatedBy(actor).Exec(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := recordAudit(ctx, tx, "customer_contact", "delete", contactID, cur.CompanyID, cur.DepartmentID, actor, map[string]any{"customer_id": cur.CustomerID, "name": cur.Name}); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&customersv1.DeleteContactResponse{}), nil
}

// GetCustomerQRCode 為本部門客戶產生登入 QR(3.8.2 產生端):驗客戶可見 → 產生簽章 token →
// 組深層連結 → 寫稽核。每呼叫產生新 token(舊 token 各自有效,不互作廢)。
// 簽章密鑰複用 JWT_SECRET(環境注入,不進版控);token 本身不回傳(前端由 qr_url 取 token)。
func (s *CustomerService) GetCustomerQRCode(ctx context.Context, req *connect.Request[customersv1.GetCustomerQRCodeRequest]) (*connect.Response[customersv1.GetCustomerQRCodeResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	custID, err := parseID(req.Msg.GetCustomerId())
	if err != nil {
		return nil, err
	}
	cust, err := customerScopeQuery(db.Customer.Query(), cid, did).
		Where(customer.ID(custID), customer.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	token, _, err := qrcode.Generate(qrSecret(), cust.CompanyID, cust.CustomerCode, 0)
	if err != nil {
		return nil, err
	}
	actor, _ := parseID(id.UserID)
	if err := recordAudit(ctx, tx, "customer", "qr_produce", cust.ID, cid, cust.DepartmentID, actor,
		map[string]any{"customer_code": cust.CustomerCode}); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&customersv1.GetCustomerQRCodeResponse{
		QrUrl: qrDeepLink(s.accountManageBaseURL, token),
	}), nil
}

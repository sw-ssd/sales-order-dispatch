// Package fileassets 的儲存與 HTTP 端點(04 計畫 Task 3.6.2–3.6.3)。
// 儲存檔名系統產生(uuid + 正規化副檔名);路徑 <root>/<company>/<yyyy>/<mm>/<filename>;
// 先落盤 fsync → 再建 DB 記錄(+稽核)同交易;DB 失敗刪孤兒檔。
package fileassets

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
)

// Store 為檔案儲存依賴(根目錄由 config.Storage.StorageRoot 注入)。
type Store struct {
	db   *ent.Client
	root string
}

// NewStore 建立 Store。
func NewStore(db *ent.Client, root string) *Store {
	return &Store{db: db, root: root}
}

// Saved 為一次成功上傳的結果(回應 JSON 用,不含 storage_path)。
type Saved struct {
	ID               int
	URL              string
	Filename         string
	OriginalFilename string
	MIME             string
	Size             int64
}

// SaveUpload 驗證 → 落盤 → 建記錄(+稽核),同一流程。owner 由呼叫端傳入(已驗租戶範圍)。
// ctx 須帶請求交易(TxFrom):DB 寫入與稽核同交易;檔案落盤在交易外先行,DB 失敗刪孤兒檔。
func (s *Store) SaveUpload(ctx context.Context, id authz.Identity, cid int, did *int,
	ownerType string, ownerID int, r io.Reader, declaredMIME, filename string) (*Saved, error) {
	v, err := Validate(r, declaredMIME, filename)
	if err != nil {
		return nil, err
	}
	name, err := randName(v.ext)
	if err != nil {
		return nil, errcode.SysInternal.Error(nil)
	}
	now := time.Now()
	rel := filepath.Join(strconv.Itoa(cid), fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", int(now.Month())), name)
	abs := filepath.Join(s.root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return nil, errcode.SysInternal.Error(nil)
	}
	f, err := os.OpenFile(abs, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return nil, errcode.SysInternal.Error(nil)
	}
	if _, err := io.Copy(f, v.reader(r)); err != nil {
		_ = f.Close()
		_ = os.Remove(abs)
		return nil, errcode.SysInternal.Error(nil)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(abs)
		return nil, errcode.SysInternal.Error(nil)
	}
	_ = f.Close()

	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		_ = os.Remove(abs) // 無交易不建記錄,亦不留孤兒檔
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	build := db.FileAsset.Create().
		SetCompanyID(cid).SetOwnerType(ownerType).SetOwnerID(ownerID).
		SetFilename(name).SetOriginalFilename(filename).
		SetMimeType(v.mime).SetSizeBytes(int(v.size)).
		SetStoragePath(rel).SetURL("/api/v1/files/" + name + "/download")
	if did != nil {
		build = build.SetDepartmentID(*did)
	}
	if n, err := parseActor(id.UserID); err == nil {
		build = build.SetCreatedBy(n)
	}
	saved, err := build.Save(ctx)
	if err != nil {
		_ = os.Remove(abs) // DB 失敗刪孤兒檔
		return nil, toConnectErr(err)
	}
	actor, _ := parseActor(id.UserID)
	if err := recordAudit(ctx, tx, saved.ID, cid, did, actor); err != nil {
		_ = os.Remove(abs)
		return nil, err
	}
	return &Saved{
		ID: saved.ID, URL: saved.URL, Filename: name,
		OriginalFilename: filename, MIME: v.mime, Size: v.size,
	}, nil
}

// randName 產生系統檔名(uuid hex + 正規化副檔名)。
func randName(ext string) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]) + ext, nil
}

// parseActor 由身分取數字 actor id(解析失敗回註冊碼錯誤)。
func parseActor(s string) (int, error) {
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil || n <= 0 {
		return 0, errcode.SysInvalidArgument.Error(map[string]string{"field": "id"})
	}
	return int(n), nil
}

// toConnectErr 轉 ent 錯誤為 connect 錯誤(本層不依賴 services 的單一出口)。
func toConnectErr(err error) error {
	if ent.IsNotFound(err) {
		return errcode.SysNotFound.Error(nil)
	}
	return errcode.SysInternal.Wrap(err)
}

// recordAudit 寫上傳稽核(與建記錄同交易,D18)。
func recordAudit(ctx context.Context, tx *ent.Tx, id, cid int, did *int, actor int) error {
	return audit.Record(ctx, tx, audit.Entry{
		Action: "create", ResourceType: "file_asset", ResourceID: strconv.Itoa(id),
		CompanyID: cid, DepartmentID: did, UserID: actor,
		After:     map[string]any{"filename": id},
		IPAddress: audit.MetaFrom(ctx).IP,
		UserAgent: audit.MetaFrom(ctx).UserAgent,
	})
}

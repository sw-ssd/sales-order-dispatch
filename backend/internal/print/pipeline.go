// Package print 的 PDF 產線(09 計畫 Task 5.4.3):渲染 → Gotenberg → file_assets 落檔。
// 空表不產生(回 failed_precondition,不呼叫轉換、不寫記錄);PDF ≤ 10 MB 白名單上限;
// 每次產生皆新 file_asset(全留存);落檔在 DB 交易外,元資料由呼叫端(5.5)在同交易寫入。
package print

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
)

// Pipeline 為產線依賴(轉換器 + 儲存根目錄)。
type Pipeline struct {
	conv Converter
	root string
}

// NewPipeline 建立 Pipeline。
func NewPipeline(conv Converter, root string) *Pipeline {
	return &Pipeline{conv: conv, root: root}
}

// Produced 為一次產線結果(落檔實體路徑 + PDF 位元組;元資料由呼叫端寫入)。
type Produced struct {
	RelPath  string
	Filename string
	Size     int64
	PDF      []byte
}

// Produce 執行產線:空表守門 → 渲染 → 轉換 → 大小檢查 → 落檔(fsyc)。
// 落檔路徑與 fileassets 同慣例 <root>/<company>/<yyyy>/<mm>/<uuid>.pdf。
func (p *Pipeline) Produce(model any, typ DocumentType, empty bool, cid int) (*Produced, error) {
	if empty {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"reason": "無可列印資料"})
	}
	html, err := Render(typ, model)
	if err != nil {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "document_type"})
	}
	pdf, err := p.conv.Convert(html)
	if err != nil {
		return nil, err
	}
	if int64(len(pdf)) > MaxPDFBytes {
		return nil, errcode.SysInternal.Error(nil)
	}
	name := uuid.NewString() + ".pdf"
	now := time.Now()
	rel := filepath.Join(strconv.Itoa(cid),
		now.Format("2006"), now.Format("01"), name)
	abs := filepath.Join(p.root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return nil, errcode.SysInternal.Error(nil)
	}
	f, err := os.OpenFile(abs, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return nil, errcode.SysInternal.Error(nil)
	}
	if _, err := io.Copy(f, bytes.NewReader(pdf)); err != nil {
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
	return &Produced{RelPath: rel, Filename: name, Size: int64(len(pdf)), PDF: pdf}, nil
}

// Discard 補償刪除(交易失敗時清孤兒檔)。
func (p *Pipeline) Discard(rel string) {
	_ = os.Remove(filepath.Join(p.root, rel))
}

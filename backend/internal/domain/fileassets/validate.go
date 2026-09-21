// Package fileassets 檔案資產(04 計畫 Task 3.6):上傳驗證(3.6.1) + 本地儲存(3.6.2)。
// 白名單:宣告 MIME ↔ 副檔名 ↔ magic bytes 三者一致才接受;1.0 不做病毒掃描(規格明定)。
package fileassets

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
)

// 大小上限(3.6.1):圖片 5 MB,PDF 10 MB。
const (
	MaxImageBytes = 5 << 20
	MaxPDFBytes   = 10 << 20
)

// validatedFile 為驗證通過的檔案(正規化 MIME + 副檔名 + 大小 + 檔頭)。
type validatedFile struct {
	mime string
	ext  string
	size int64
	head []byte // magic bytes 檢查讀到的檔頭(寫檔時先回填,不重讀上傳串流)
}

// fileSpec 為白名單一列(宣告 MIME ↔ 副檔名 ↔ magic bytes ↔ 上限)。
type fileSpec struct {
	mime  string
	exts  []string
	magic []byte
	// webp 以雙段簽章判定(RIFF....WEBP),magic 為前綴,webb 為第 8–11 位元組。
	webb  []byte
	limit int64
}

var allowlist = []fileSpec{
	{mime: "image/jpeg", exts: []string{".jpg", ".jpeg"}, magic: []byte{0xFF, 0xD8, 0xFF}, limit: MaxImageBytes},
	{mime: "image/png", exts: []string{".png"}, magic: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, limit: MaxImageBytes},
	{mime: "image/webp", exts: []string{".webp"}, magic: []byte{0x52, 0x49, 0x46, 0x46}, webb: []byte{0x57, 0x45, 0x42, 0x50}, limit: MaxImageBytes},
	{mime: "application/pdf", exts: []string{".pdf"}, magic: []byte{0x25, 0x50, 0x44, 0x46}, limit: MaxPDFBytes},
}

// Validate 驗證上傳(3.6.1):大小 → 宣告 MIME 白名單 → 副檔名匹配 → magic bytes。
// r 為檔案串流(只讀一次:上限截斷 + 檔頭保留);declaredMIME 為呼叫端宣告;filename 為原始檔名。
// 成功回正規化結果;任一不符即 invalid_argument(不回傳檔案內容細節)。
func Validate(r io.Reader, declaredMIME, filename string) (*validatedFile, error) {
	mime := strings.ToLower(strings.TrimSpace(declaredMIME))
	var spec *fileSpec
	for i := range allowlist {
		if allowlist[i].mime == mime {
			spec = &allowlist[i]
			break
		}
	}
	if spec == nil {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "mime_type"})
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "filename"})
	}
	okExt := false
	for _, e := range spec.exts {
		if e == ext {
			okExt = true
			break
		}
	}
	if !okExt {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "filename"})
	}
	// 上限截斷讀取:超限即拒,不完整落盤(多讀 1 位元組判定超限)。
	limited := io.LimitReader(r, spec.limit+1)
	buf, err := io.ReadAll(limited)
	if err != nil {
		return nil, errcode.SysInternal.Error(nil)
	}
	if int64(len(buf)) > spec.limit {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{
			"field": "size_bytes", "limit": fmt.Sprintf("%d", spec.limit),
		})
	}
	if len(buf) < len(spec.magic) || !bytes.Equal(buf[:len(spec.magic)], spec.magic) {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "magic_bytes"})
	}
	if spec.webb != nil {
		if len(buf) < 12 || !bytes.Equal(buf[8:12], spec.webb) {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "magic_bytes"})
		}
	}
	return &validatedFile{mime: spec.mime, ext: ext, size: int64(len(buf)), head: buf}, nil
}

// headerReader 回放檔頭 + 剩餘串流(儲存端先寫 head,再拷貝原串流剩餘部分)。
func (v *validatedFile) reader(r io.Reader) io.Reader {
	return io.MultiReader(bytes.NewReader(v.head), r)
}

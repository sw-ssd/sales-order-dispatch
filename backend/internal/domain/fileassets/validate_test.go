package fileassets_test

import (
	"strings"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/domain/fileassets"
)

// TestValidateTable 白名單四型各一正一反:副檔名/MIME/magic 三者一致才過;
// 改副檔名的偽裝檔被 magic 擋;超限被擋。
func TestValidateTable(t *testing.T) {
	pngHead := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00}
	jpgHead := []byte{0xFF, 0xD8, 0xFF, 0x00}
	webpHead := []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50}
	pdfHead := []byte{0x25, 0x50, 0x44, 0x46, 0x2D}

	cases := []struct {
		name     string
		head     []byte
		mime     string
		filename string
		wantErr  bool
	}{
		{"png 通過", pngHead, "image/png", "a.png", false},
		{"jpg 通過(大寫副檔名)", jpgHead, "image/jpeg", "a.JPG", false},
		{"webp 通過", webpHead, "image/webp", "a.webp", false},
		{"pdf 通過", pdfHead, "application/pdf", "a.pdf", false},
		{"exe 改名 jpg 被 magic 擋", []byte("MZ......."), "image/jpeg", "evil.jpg", true},
		{"副檔名與 MIME 不符", pngHead, "image/jpeg", "a.png", true},
		{"MIME 非白名單", pngHead, "image/gif", "a.gif", true},
		{"無副檔名", pngHead, "image/png", "a", true},
		{"webp 缺 WEBP 段", []byte{0x52, 0x49, 0x46, 0x46, 0, 0, 0, 0, 0, 0, 0, 0}, "image/webp", "a.webp", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fileassets.Validate(bytesReader(tc.head), tc.mime, tc.filename)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate(%q,%q) err=%v;wantErr=%v", tc.mime, tc.filename, err, tc.wantErr)
			}
		})
	}
}

// TestValidateSizeLimit 超限即拒(含上限邊界通過)。
func TestValidateSizeLimit(t *testing.T) {
	head := append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, make([]byte, 16)...)
	if _, err := fileassets.Validate(bytesReader(head), "image/png", "a.png"); err != nil {
		t.Fatalf("小檔應通過: %v", err)
	}
	big := append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
		make([]byte, fileassets.MaxImageBytes)...)
	if _, err := fileassets.Validate(bytesReader(big), "image/png", "a.png"); err == nil {
		t.Fatal("超 5MB 應拒絕")
	}
}

func bytesReader(b []byte) *strings.Reader {
	return strings.NewReader(string(b))
}

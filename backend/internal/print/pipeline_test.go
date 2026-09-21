package print_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/print"
)

// fakeConv 為假轉換器(計次 + 回固定 PDF)。
type fakeConv struct {
	calls int
	pdf   []byte
	err   error
}

func (f *fakeConv) Convert(html string) ([]byte, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.pdf, nil
}

// TestPipelineEmptyGuardsConverter 空表不呼叫轉換(不產生 PDF、不寫檔)。
func TestPipelineEmptyGuardsConverter(t *testing.T) {
	fc := &fakeConv{pdf: []byte("%PDF-1.4 fake")}
	p := print.NewPipeline(fc, t.TempDir())
	if _, err := p.Produce(nil, print.DispatchSummary, true, 7); err == nil {
		t.Fatal("空表應拒絕")
	}
	if fc.calls != 0 {
		t.Fatalf("空表不得呼叫轉換,got %d 次", fc.calls)
	}
}

// TestPipelineWritesFile 成功落檔(路徑含公司/年/月;實體存在;Discard 清除)。
func TestPipelineWritesFile(t *testing.T) {
	root := t.TempDir()
	fc := &fakeConv{pdf: []byte("%PDF-1.4 fake")}
	p := print.NewPipeline(fc, root)
	model := print.DispatchSummaryModel{CompanyName: "C",
		Stores: []print.StoreBlock{{Name: "S", Items: []print.StoreItem{{Name: "N"}}}}}
	out, err := p.Produce(model, print.DispatchSummary, false, 7)
	if err != nil {
		t.Fatalf("Produce: %v", err)
	}
	if fc.calls != 1 {
		t.Fatalf("應轉換 1 次,got %d", fc.calls)
	}
	if _, err := os.Stat(filepath.Join(root, out.RelPath)); err != nil {
		t.Fatalf("實體檔應存在: %v", err)
	}
	p.Discard(out.RelPath)
	if _, err := os.Stat(filepath.Join(root, out.RelPath)); !os.IsNotExist(err) {
		t.Fatal("Discard 應刪除實體檔")
	}
}

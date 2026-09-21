package print_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/print"
)

// TestClientConvertSuccess 合法 HTML 可轉出 PDF(httptest 假 Gotenberg 回 %PDF)。
func TestClientConvertSuccess(t *testing.T) {
	var gotCT string
	var gotPaper string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		_ = r.ParseMultipartForm(1 << 20)
		gotPaper = r.FormValue("paperWidth")
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("%PDF-1.4 fake"))
	}))
	defer srv.Close()
	pdf, err := print.NewClient(srv.URL).Convert("<html><body>單</body></html>")
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if !strings.HasPrefix(string(pdf), "%PDF") {
		t.Fatalf("PDF 魔術位元組缺失: %q", pdf)
	}
	if !strings.HasPrefix(gotCT, "multipart/form-data") {
		t.Fatalf("應為 multipart,got %q", gotCT)
	}
	if gotPaper != "8.27in" {
		t.Fatalf("紙張應為 A4 寬,got %q", gotPaper)
	}
}

// TestClientRetryOn5xx 上游 5xx 重試後成功(首 500、次 200 → 共 2 次請求)。
func TestClientRetryOn5xx(t *testing.T) {
	n := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		if n == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = w.Write([]byte("%PDF-1.4 fake"))
	}))
	defer srv.Close()
	if _, err := print.NewClient(srv.URL).Convert("<html></html>"); err != nil {
		t.Fatalf("重試後應成功: %v", err)
	}
	if n != 2 {
		t.Fatalf("應請求 2 次,got %d", n)
	}
}

// TestClientNoRetryOn4xx 4xx 不重試(恰 1 次請求即失敗)。
func TestClientNoRetryOn4xx(t *testing.T) {
	n := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		w.WriteHeader(http.StatusUnprocessableEntity)
	}))
	defer srv.Close()
	if _, err := print.NewClient(srv.URL).Convert("<html></html>"); err == nil {
		t.Fatal("4xx 應失敗")
	}
	if n != 1 {
		t.Fatalf("4xx 不得重試,got %d 次", n)
	}
}

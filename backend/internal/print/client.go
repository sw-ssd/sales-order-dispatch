// Package print 的 Gotenberg client(09 計畫 Task 5.4.1):HTML → A4 PDF。
// multipart 主文件 index.html 送 Chromium HTML 轉換路由;紙張 A4、背景圖形開、
// 等待網路靜止;連線錯誤與上游 5xx 至多重試 2 次(指數退避),4xx 不重試;整體 deadline;
// PDF 全程記憶體傳遞,不落暫存檔。
package print

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
)

// 轉換參數(固定):A4、合理邊距、背景圖形開、等待網路靜止。
const (
	gotenbergPath  = "/forms/chromium/convert/html"
	requestTimeout = 60 * time.Second
	maxAttempts    = 3 // 首次 + 至多 2 次重試
	retryBaseDelay = time.Second
	paperWidth     = "8.27in"
	paperHeight    = "11.69in"
	marginTop      = "0.4in"
	marginBottom   = "0.4in"
	marginLeft     = "0.4in"
	marginRight    = "0.4in"
)

// Client 為 Gotenberg client(Converter 介面正式實作)。
type Client struct {
	baseURL string
	http    *http.Client
}

// Converter 為 HTML→PDF 轉換介面(fake / httptest 測試用)。
type Converter interface {
	Convert(html string) ([]byte, error)
}

// NewClient 建立 Client。
func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: requestTimeout}}
}

// Convert 轉 HTML 為 PDF(重試規則:連線錯誤與 5xx 重試,4xx 不重試)。
func (c *Client) Convert(html string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(retryBaseDelay << (attempt - 1))
		}
		pdf, retryable, err := c.once(html)
		if err == nil {
			return pdf, nil
		}
		lastErr = err
		if !retryable {
			return nil, err
		}
	}
	return nil, lastErr
}

// once 執行單次轉換;回 (pdf, 可重試否, err)。
func (c *Client) once(html string) ([]byte, bool, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, false, errcode.SysInternal.Wrap(err)
	}
	if _, err := io.WriteString(fw, html); err != nil {
		return nil, false, errcode.SysInternal.Wrap(err)
	}
	// 固定轉換參數。
	for k, v := range map[string]string{
		"paperWidth": paperWidth, "paperHeight": paperHeight,
		"marginTop": marginTop, "marginBottom": marginBottom,
		"marginLeft": marginLeft, "marginRight": marginRight,
		"printBackground": "true", "waitForNetworkIdle": "true",
	} {
		_ = w.WriteField(k, v)
	}
	if err := w.Close(); err != nil {
		return nil, false, errcode.SysInternal.Wrap(err)
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+gotenbergPath, &b)
	if err != nil {
		return nil, false, errcode.SysInternal.Wrap(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, errcode.SysInternal.Wrap(err) // 連線錯誤可重試
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxPDFBytes+1))
	if err != nil {
		return nil, true, errcode.SysInternal.Wrap(err)
	}
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		if int64(len(body)) > MaxPDFBytes {
			return nil, false, errcode.SysInternal.Error(nil) // 超白名單上限屬異常
		}
		return body, false, nil
	case resp.StatusCode >= 500:
		return nil, true, errcode.SysInternal.Wrap(fmt.Errorf("gotenberg %d: %s", resp.StatusCode, truncate(body)))
	default: // 4xx:請求錯誤(模板/資料問題),不重試
		return nil, false, errcode.SysInvalidArgument.Error(map[string]string{"field": "html"})
	}
}

// truncate 截短上游錯誤內容(記 trace 用,不回傳檔案細節)。
func truncate(b []byte) string {
	if len(b) > 200 {
		return string(b[:200])
	}
	return string(b)
}

// MaxPDFBytes 為 PDF 上限(pdf ≤ 10 MB,D17;與 fileassets 白名單同值,避免重複定義漂移由測試釘住)。
const MaxPDFBytes = 10 << 20

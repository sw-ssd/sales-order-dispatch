/**
 * vitest（jsdom）專用的環境補丁。
 *
 * jsdom 沒有 `ResizeObserver`/`IntersectionObserver`，而 Ark 的 scroll-area 會在掛載時
 * 建立這兩個 observer（任何渲染到 `ScrollArea` 的元件測試都會跟著炸，例如 dialog）。
 * 這裡只把平台缺口補上：不觸發回呼、不改任何行為；需要量測的測試自行以真實事件驅動。
 */
class NoopObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}

Object.assign(window, { ResizeObserver: NoopObserver, IntersectionObserver: NoopObserver });

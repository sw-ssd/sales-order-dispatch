/**
 * vitest（jsdom）專用的環境補丁。
 *
 * jsdom 沒有 `ResizeObserver`/`IntersectionObserver`，而共用的 Ark UI 元件（scroll-area 等）
 * 會在掛載時建立這兩個 observer。這裡只把平台缺口補上：不觸發回呼、不改任何行為。
 * （與租戶 SPA 的 `frontend/src/test-setup.ts` 相同；那是 jsdom 的缺口，不是兩份實作。）
 */
class NoopObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}

Object.assign(window, { ResizeObserver: NoopObserver, IntersectionObserver: NoopObserver });

// jsdom 沒有實作捲動：TanStack Router 的 scroll restoration 會呼叫它並噴
// 「Not implemented: window.scrollTo」到 stderr。測試不驗捲動，直接給 no-op。
window.scrollTo = () => {};

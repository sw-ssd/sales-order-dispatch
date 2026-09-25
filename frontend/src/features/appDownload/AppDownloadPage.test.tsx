import { render, screen } from "@solidjs/testing-library";
import { describe, expect, it } from "vitest";

import { ManageDownloadPage, QRDownloadPage } from "./AppDownloadPage";

/**
 * App Link 的未安裝落頁（規格 §4.2：未安裝導向商店）。
 *
 * 這頁唯一的職責是「把人送去商店」，所以斷言分兩層：
 * 1. 兩種流程各自說對自己的話（掃碼 vs 帳號管理）—— 文案互換會讓使用者做錯事。
 * 2. 商店網址未設定時（目前實況：App 未上架）**不得**出現指向假網址的按鈕。
 *    這是本頁最可能的退化方式：為了「看起來完整」而寫死兩條死連結。
 */
describe("App 下載落頁", () => {
  it("QR 流程說明掃碼登入，而不是帳號管理", () => {
    render(() => <QRDownloadPage />);
    expect(screen.getByText(/掃碼登入/)).toBeTruthy();
    expect(screen.queryByText(/帳號管理在 App 內操作/)).toBeNull();
  });

  it("帳號管理流程說明以主帳號登入", () => {
    render(() => <ManageDownloadPage />);
    expect(screen.getByText(/帳號管理在 App 內操作/)).toBeTruthy();
    expect(screen.queryByText(/掃碼登入/)).toBeNull();
  });

  it("商店網址未設定時給聯絡業務的退路，而非死連結", () => {
    render(() => <ManageDownloadPage />);
    expect(screen.getByText(/請聯絡您的業務人員/)).toBeTruthy();
    expect(screen.queryByRole("link", { name: /App Store/ })).toBeNull();
    expect(screen.queryByRole("link", { name: /Google Play/ })).toBeNull();
  });
});

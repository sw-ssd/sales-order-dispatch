import { Code, ConnectError } from "@connectrpc/connect";
import { describe, expect, it } from "vitest";
import { errorMessage } from "./error-message";
import { apiError, bareConnectError } from "@/test-api-error";

describe("errorMessage", () => {
  it("後端 details.reason 的中文原因優先於通用樣板（這是使用者真正需要的線索）", () => {
    // 真實案例：print.Assemble 的「無可列印資料」只放在 details.reason，
    // 訊息本體是樣板原文「參數驗證失敗」。只讀 message 會丟掉唯一的線索。
    const err = apiError("SYS-1001", "參數驗證失敗", { reason: "無可列印資料" });
    expect(errorMessage(err)).toBe("無可列印資料");
  });

  it("details.reason 若是機器可讀 token 則不採用（不外洩 already_inactive 這類內部值）", () => {
    const err = apiError("SYS-1001", "參數驗證失敗", { reason: "already_inactive" });
    expect(errorMessage(err)).toBe("輸入資料有誤，請檢查後再試");
  });

  it("後端帶實際數值的訊息勝過通用文案（樣板沒有的補充資訊要保留）", () => {
    const rendered = "已達方案上限（9/10），請升級方案";
    const err = apiError("PLAT-5001", rendered, {}, Code.FailedPrecondition);
    expect(errorMessage(err)).toBe(rendered);
  });

  it("後端訊息與碼表樣板逐字相同時改用本檔文案（registry 的字串是給 log 與平台端用的）", () => {
    // SYS-1001 的 registry 訊息就是「參數驗證失敗」，對租戶端使用者不具行動指引。
    const err = apiError("SYS-1001", "參數驗證失敗");
    expect(errorMessage(err)).toBe("輸入資料有誤，請檢查後再試");
  });

  it("碼表文案優先於 connect 碼分類文案：SYS-3001 不可說成「狀態不允許」", () => {
    // SYS-3001 與 SYS-3002 同為 FailedPrecondition，但語意是「範圍」與「約束」，
    // 用分類文案會讓使用者以為是狀態問題而重試同一個操作。
    const err = apiError(
      "SYS-3001",
      "資料超出目前的存取範圍，無法完成此操作",
      {},
      Code.FailedPrecondition
    );
    expect(errorMessage(err)).toBe("資料超出目前的存取範圍，無法完成此操作");
  });

  it("碼表樣板代入 details；填不滿時退回通用文案而不是把 {used} 印給使用者", () => {
    const filled = apiError(
      "PLAT-5001",
      "已達方案上限（9/10），請升級方案",
      { used: "9", limit: "10" },
      Code.FailedPrecondition
    );
    expect(errorMessage(filled)).toBe("已達方案上限（9/10），請升級方案");

    // message 刻意與樣板相同 → 走碼表＋代入路徑；details 缺參數 → 不可回含 { 的字串。
    const unfilled = apiError("PLAT-5001", "已達方案上限（{used}/{limit}），請升級方案", {}, Code.FailedPrecondition);
    const msg = errorMessage(unfilled);
    expect(msg).not.toContain("{");
  });

  it("不顯示錯誤碼（租戶端使用者用不到），且標點一律全形", () => {
    const err = apiError("SYS-1001", "參數驗證失敗");
    const msg = errorMessage(err);
    expect(msg).not.toContain("SYS-");
    // 全形標點才是正確形式；半形逗號／括號是排版缺陷。
    expect(msg).not.toContain(",");
    expect(msg).not.toContain("(");
    expect(msg).toMatch(/，/);
  });

  it("同一組 connect 碼在各頁面得到同一句話（不再有『連線至／連線到』兩種寫法）", () => {
    const a = errorMessage(new ConnectError("x", Code.Unavailable));
    const b = errorMessage(new ConnectError("y", Code.Unavailable));
    expect(a).toBe(b);
    expect(a).toBe("無法連線至伺服器，請確認後端服務已啟動");
  });

  it("非 ConnectError（網路層直接失敗）也給可行動的訊息，不外洩原始錯誤", () => {
    expect(errorMessage(new TypeError("Failed to fetch"))).toBe(
      "無法連線至伺服器，請確認後端服務已啟動"
    );
    expect(errorMessage(undefined)).toBe("無法連線至伺服器，請確認後端服務已啟動");
  });

  it("未知碼退回 rawMessage，仍不顯示碼本身", () => {
    const err = apiError("SYS-9999", "後端自訂說明", {}, Code.Unknown);
    expect(errorMessage(err)).toBe("後端自訂說明");
  });

  it("permissionHint 只覆寫權限拒絕，其他錯誤不受影響（公告頁的範圍說明靠它保留）", () => {
    const denied = apiError("SYS-4001", "缺少權限", {}, Code.PermissionDenied);
    expect(errorMessage(denied, { permissionHint: "沒有權限執行此操作（公告範圍須在你的管理範圍內）" })).toBe(
      "沒有權限執行此操作（公告範圍須在你的管理範圍內）"
    );
    const other = apiError("SYS-1001", "參數驗證失敗");
    expect(errorMessage(other, { permissionHint: "不會用到" })).toBe("輸入資料有誤，請檢查後再試");
  });

  it("沒有 ErrorInfo 的裸中文訊息直接採用（後端服務層手寫的那些句子是唯一來源）", () => {
    // 後端 create/delete 有大量 `connect.NewError(Code, errors.New("中文"))`；
    // 若一律改寫成 connect 碼的分類文案，「部門仍有使用者」這種可修復的線索就沒了。
    expect(errorMessage(bareConnectError("部門仍有使用者,無法刪除", Code.FailedPrecondition))).toBe(
      "部門仍有使用者，無法刪除"
    );
    // 半形逗號（後端手寫字串的實況）在出口統一成全形，但兩側非中文的不動。
    expect(errorMessage(bareConnectError("公司仍設定客戶編號前綴", Code.FailedPrecondition))).toBe(
      "公司仍設定客戶編號前綴"
    );
  });

  it("Code.Internal 的裸訊息是內部字串，一律改寫（不得外洩『缺少租戶交易(context)』）", () => {
    const msg = errorMessage(bareConnectError("缺少租戶交易(context)", Code.Internal));
    expect(msg).not.toContain("租戶交易");
    expect(msg).toBe("伺服器暫時無法使用，請稍後再試");
  });
});

describe("errorMessage 回歸：先前各自為政的文案", () => {
  it("FailedPrecondition 曾有 4 種寫法，現在只有一種", () => {
    const err = apiError("SYS-3001", "資料超出目前的存取範圍，無法完成此操作", {}, Code.FailedPrecondition);
    expect(errorMessage(err)).toBe("資料超出目前的存取範圍，無法完成此操作");
  });
});

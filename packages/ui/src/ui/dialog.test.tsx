import { render, screen, fireEvent, waitFor } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";
import { createSignal } from "solid-js";
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogTitle } from "./dialog";

const Harness = (props: {
  onOpenChange?: (open: boolean) => void;
  closeOnOutsideClick?: boolean;
  showCloseButton?: boolean;
}) => {
  const [open, setOpen] = createSignal(true);
  return (
    <Dialog
      open={open()}
      closeOnOutsideClick={props.closeOnOutsideClick}
      onOpenChange={(value) => {
        setOpen(value);
        props.onOpenChange?.(value);
      }}
    >
      <DialogContent showCloseButton={props.showCloseButton}>
        <DialogTitle>新增公司</DialogTitle>
        <DialogDescription>建立新的公司主檔</DialogDescription>
        <DialogClose>取消</DialogClose>
      </DialogContent>
    </Dialog>
  );
};

describe("Dialog", () => {
  it("open 時顯示標題與說明", () => {
    render(() => <Harness />);
    // getByRole 預設排除 hidden 節點：關閉時 Ark 仍把內容留在 DOM（帶 hidden），
    // 少了這條正向斷言，即使 open 沒接上本案例也會全綠。
    expect(screen.getByRole("dialog")).toBeTruthy();
    expect(screen.getByText("新增公司")).toBeTruthy();
    expect(screen.getByText("建立新的公司主檔")).toBeTruthy();
    expect(screen.getByRole("button", { name: "關閉" })).toBeTruthy();
  });

  it("點關閉鈕回報 onOpenChange(false)", async () => {
    const onOpenChange = vi.fn();
    render(() => <Harness onOpenChange={onOpenChange} />);
    fireEvent.click(screen.getByText("取消"));
    // Ark 的狀態機在 microtask 後才回報，用真實等待而非同步斷言。
    await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
  });

  it("showCloseButton=false 時沒有關閉鈕", () => {
    render(() => <Harness showCloseButton={false} />);
    // 先確認對話框確實開著，否則「沒有關閉鈕」也可能只是因為整個對話框沒渲染。
    expect(screen.getByRole("dialog")).toBeTruthy();
    expect(screen.queryByRole("button", { name: "關閉" })).toBeNull();
  });

  it("按 Escape 回報 onOpenChange(false)", async () => {
    const onOpenChange = vi.fn();
    render(() => <Harness onOpenChange={onOpenChange} />);
    fireEvent.keyDown(document.body, { key: "Escape", code: "Escape" });
    await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
  });
});

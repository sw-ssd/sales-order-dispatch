import { createSignal, Show, type JSX } from "solid-js";
import { Button, buttonVariants } from "@ui/button";
import { DialogClose, DialogFooter } from "@ui/dialog";
import { Field, FieldDescription, FieldError, FieldLabel } from "@ui/field";
import { Input } from "@ui/input";

/**
 * 平台寫入的稽核必填：後端 `platformReason` 對每個寫入 RPC 都擋空白 reason。
 * 前端先擋不是授權，是**少跑一趟**白工的 RPC（並讓 operator 當場知道缺什麼）。
 */
export const REASON_REQUIRED = "原因必填：平台寫入會寫進稽核，日後查帳只看得到這裡的文字。";

/** 每個表單自己的欄位 id：同一頁同時存在多個對話框時不得撞號（label 的 for 要對得上）。 */
let reasonSeq = 0;

/**
 * 平台寫入表單的共同骨架：**原因欄 ＋ 可讀的錯誤 ＋ 送出／取消**。
 *
 * 四個寫入流程（例外設定／例外撤銷／價目 upsert／方案權益）的必填欄位不同，
 * 但「reason 必填、失敗訊息一律經 describeError、送出中的按鈕狀態」完全一樣，
 * 所以在這裡寫一次；各自的欄位由呼叫端以 children 傳入。
 *
 * 呼叫端負責兩件事：
 * - `validate`：after reason 通過後檢查自己的欄位（回傳訊息＝不送出並顯示）；
 * - `onSubmit(reason)`：真的打 RPC，並在成功時讓相關查詢失效（見各頁的 mutation onSuccess）。
 */
export function WriteForm(props: {
  /** 送出（reason 已通過必填檢查、validate 已通過）。 */
  onSubmit: (reason: string) => void;
  submitLabel: string;
  /** 送出前的欄位檢查：回傳訊息即擋下（例：金額格式、負的上限）。 */
  validate?: () => string | undefined;
  /** 後端失敗的訊息（一律經 `describeError`；沒失敗時不要傳值）。 */
  error?: string;
  pending?: boolean;
  children?: JSX.Element;
}): JSX.Element {
  const [reason, setReason] = createSignal("");
  const [reasonError, setReasonError] = createSignal("");
  const [fieldError, setFieldError] = createSignal("");
  const reasonId = `write-reason-${++reasonSeq}`;

  const submit = (e: Event) => {
    e.preventDefault();

    // 每次送出都重算：上一次的錯誤不得殘留成下一次的假訊息。
    const missingReason = reason().trim() === "";
    setReasonError(missingReason ? REASON_REQUIRED : "");
    const invalid = missingReason ? undefined : props.validate?.();
    setFieldError(invalid ?? "");
    if (missingReason || invalid) return;

    props.onSubmit(reason());
  };

  return (
    <form class="space-y-4" novalidate onSubmit={submit}>
      {props.children}

      <Field invalid={reasonError() !== ""}>
        <FieldLabel for={reasonId}>原因 *</FieldLabel>
        <Input
          id={reasonId}
          value={reason()}
          placeholder="例：客戶來信、合約增補、年度調價"
          onInput={(e) => {
            setReason(e.currentTarget.value);
            setReasonError("");
          }}
        />
        <FieldDescription>這行文字會寫進平台稽核（含你的 operator 身分）。</FieldDescription>
        <FieldError>{reasonError()}</FieldError>
      </Field>

      <Show when={fieldError() || props.error}>
        <p role="alert" class="text-sm font-medium text-destructive">
          {fieldError() || props.error}
        </p>
      </Show>

      <DialogFooter>
        <DialogClose class={buttonVariants({ variant: "outline", size: "sm" })}>取消</DialogClose>
        <Button type="submit" size="sm" disabled={props.pending}>
          {props.submitLabel}
        </Button>
      </DialogFooter>
    </form>
  );
}

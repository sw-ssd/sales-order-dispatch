import { createForm } from "@tanstack/solid-form";
import { createQuery, useQueryClient } from "@tanstack/solid-query";
import { batch, createSignal, For, Show, type JSX } from "solid-js";
import { Code, ConnectError } from "@connectrpc/connect";
import type { Customer } from "~/lib/proto/customers/v1/customer_pb";
import type { CustomerProduct } from "~/lib/proto/products/v1/product_pb";
import {
  Button,
  buttonVariants,
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Field,
  FieldError,
  FieldLabel,
  Input,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui";
import { appFormOptions, fieldValidators, firstMessage } from "../../form-helpers";
import {
  customerProductsQueryOptions,
  customerProductClient,
  pickerProductsQueryOptions,
} from "../queries";
import { customerProductSchema } from "../schemas";
import { queryData } from "~/lib/query-data";

/**
 * 錯誤訊息對照；樣板 = `AddressBookDialog` 的同一份 switch（各元件各自一份是本 repo 既有慣例）。
 */
function errorMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    switch (err.code) {
      case Code.NotFound:
        return "資料不存在或已被刪除";
      case Code.AlreadyExists:
        return err.rawMessage || "資料已存在";
      case Code.InvalidArgument:
        return err.rawMessage || "輸入資料有誤,請檢查後再試";
      case Code.FailedPrecondition:
        return err.rawMessage || "資料狀態不允許此操作";
      case Code.PermissionDenied:
        return "沒有權限執行此操作";
      case Code.Unauthenticated:
        return "請先登入";
      case Code.Unavailable:
        return "無法連線至伺服器,請確認後端服務已啟動";
      default:
        return err.rawMessage || "操作失敗,請稍後再試";
    }
  }
  return "無法連線至伺服器,請確認後端服務已啟動";
}

const EMPTY_VALUES = {
  productId: "",
  aliasName: "",
  defaultQty: "",
  cutNote: "",
};

export interface CustomerProductsDialogProps {
  /** 目前選取的客戶（null＝關閉）。 */
  customer: Customer | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/**
 * 客戶專屬商品清單對話框（04 Task 3.5 的 Web 面）。
 *
 * 與 `AddressBookDialog` 同形（從客戶列上開、一次只開一個對話框），三處契約：
 * - **清單不帶 `forOrder`**：那支旗標是「下單用途，排除 default_qty=0」（訂單入口語意），
 *   管理頁要看全部列 —— 否則管理員看不到自己剛設的 0 數量預設列。
 * - **新增要有產品來源**：產品下拉來自 `listProducts`（啟用中、上限 100、關鍵字可重查）；
 *   手打別名走 `aliasName`（空＝用後端的預設商品名，`AddCustomerProduct` 如此歸一）。
 * - **編輯不可改產品**：後端 `UpdateCustomerProduct` 只動 alias/default_qty/cut_note
 *   （不可改 customer/product），所以編輯態根本不呈現產品欄 —— 想換商品＝刪一列再加一列，
 *   保留原列的稽核軌跡。
 *
 * 不用 `EnsureCustomerProduct`：那是「下單手打確認儲存後」的冪等入口（重複呼叫回既有列
 * `created=false`），管理頁新增重複列本來就該被後端 `already_exists` 擋下並顯示訊息，
 * 悄悄回既有列會讓「我明明沒加成功」變成疑案。
 */
export default function CustomerProductsDialog(props: CustomerProductsDialogProps) {
  const client = useQueryClient();
  const [listError, setListError] = createSignal<string | null>(null);
  const [serverError, setServerError] = createSignal<string | undefined>();
  const [formOpen, setFormOpen] = createSignal(false);
  const [editing, setEditing] = createSignal<CustomerProduct | null>(null);
  // 產品挑選器：輸入草稿與「生效關鍵字」分離——key 只吃生效值，按查詢才打 API
  //（草稿直接進 key 會每敲一個字打一次；同清單頁草稿慣例）。
  const [pickerKeyword, setPickerKeyword] = createSignal("");
  const [appliedKeyword, setAppliedKeyword] = createSignal("");

  const customerId = () => props.customer?.id ?? "";

  const listQuery = createQuery(() => ({
    ...customerProductsQueryOptions(customerId()),
    enabled: props.open && customerId() !== "",
  }));

  const pickerQuery = createQuery(() => ({
    ...pickerProductsQueryOptions(appliedKeyword()),
    enabled: props.open,
  }));

  const refresh = async () => {
    await client.invalidateQueries({ queryKey: ["customers"] });
  };

  const productIdValidators = fieldValidators(customerProductSchema.entries.productId);
  const defaultQtyValidators = fieldValidators(customerProductSchema.entries.defaultQty);

  const form = createForm(() => ({
    ...appFormOptions,
    defaultValues: { ...EMPTY_VALUES },
    onSubmit: async ({ value }) => {
      const current = editing();
      try {
        if (current) {
          // 產品不可改（見檔頭註解）→ 編輯只送後端接受的三個欄位。
          await customerProductClient.updateCustomerProduct({
            id: current.id,
            aliasName: value.aliasName.trim(),
            defaultQty: value.defaultQty.trim(),
            cutNote: value.cutNote.trim(),
          });
        } else {
          await customerProductClient.addCustomerProduct({
            customerId: customerId(),
            productId: value.productId,
            aliasName: value.aliasName.trim(),
            defaultQty: value.defaultQty.trim(),
            cutNote: value.cutNote.trim(),
          });
        }
        setFormOpen(false);
        setEditing(null);
        setServerError(undefined);
        await refresh();
      } catch (err) {
        setServerError(errorMessage(err));
      }
    },
  }));

  const isSubmitting = form.useSelector((state) => state.isSubmitting);

  /** 開啟表單（編輯傳該筆；新增傳 null）。`reset` 整份取代（同 `CustomersPage` 慣例）。 */
  const openForm = (row: CustomerProduct | null) => {
    setEditing(row);
    setServerError(undefined);
    form.reset(
      row
        ? {
            productId: row.productId,
            aliasName: row.aliasName,
            defaultQty: row.defaultQty,
            cutNote: row.cutNote,
          }
        : { ...EMPTY_VALUES }
    );
    setFormOpen(true);
  };

  const remove = async (row: CustomerProduct) => {
    if (!window.confirm(`確定移除專屬商品「${row.aliasName}」?`)) return;
    setListError(null);
    try {
      await customerProductClient.deleteCustomerProduct({ id: row.id });
      await refresh();
    } catch (err) {
      setListError(errorMessage(err));
    }
  };

  const submit: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    if (isSubmitting()) return;
    setServerError(undefined);
    void form.handleSubmit();
  };

  const products = () => queryData(pickerQuery, (d) => d?.products) ?? [];

  return (
    <Dialog
      open={props.open && props.customer !== null}
      onOpenChange={(open) => {
        if (!open) {
          props.onOpenChange(false);
          batch(() => {
            setFormOpen(false);
            setEditing(null);
            setListError(null);
          });
        }
      }}
    >
      <DialogContent class="max-w-3xl">
        <DialogHeader>
          <DialogTitle>客戶專屬商品</DialogTitle>
          <DialogDescription>
            {props.customer ? `${props.customer.name}（${props.customer.customerCode}）` : ""}
            ：列進清單的商品會出現在該客戶的訂單商品下拉（04 Task 3.5）
          </DialogDescription>
        </DialogHeader>

        <Show when={listError()}>
          {(message) => (
            <p class="rounded-lg bg-destructive/15 px-3 py-2 text-sm text-destructive" role="alert">
              {message()}
            </p>
          )}
        </Show>

        <section class="space-y-2">
          <div class="flex items-center justify-between">
            <h3 class="font-semibold text-foreground">專屬清單</h3>
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={() => (formOpen() ? setFormOpen(false) : openForm(null))}
            >
              {formOpen() ? "取消新增" : "新增專屬商品"}
            </Button>
          </div>

          <Show when={formOpen()}>
            <form onSubmit={submit} class="space-y-2 rounded-lg border border-border bg-muted p-3">
              <Show when={serverError()}>
                {(message) => (
                  <p class="text-sm text-destructive" role="alert">
                    {message()}
                  </p>
                )}
              </Show>

              {/* 產品挑選器：關鍵字草稿 → 按查詢才換 key（不加即查，避免每敲一個字就打一次）。 */}
              <div class="flex items-end gap-3">
                <Field class="flex-1">
                  <FieldLabel for="cp-product-keyword">產品關鍵字</FieldLabel>
                  <Input
                    id="cp-product-keyword"
                    value={pickerKeyword()}
                    placeholder="代號 / 名稱"
                    onInput={(e) => setPickerKeyword(e.currentTarget.value)}
                  />
                </Field>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    // 換鍵＝自動查；同鍵再按＝顯式重查（不然按了沒反應像壞掉）。
                    if (pickerKeyword() === appliedKeyword()) void pickerQuery.refetch();
                    else setAppliedKeyword(pickerKeyword());
                  }}
                >
                  查詢產品
                </Button>
              </div>

              <form.Field name="productId" validators={productIdValidators}>
                {(field) => (
                  <Field invalid={!field().state.meta.isValid}>
                    <FieldLabel for="cp-product">產品 *</FieldLabel>
                    <Show
                      when={!editing()}
                      fallback={
                        // 編輯態不呈現產品欄：後端不可改 product（見檔頭），呈現只會誤導。
                        <p class="text-sm text-muted-foreground">
                          產品：不可修改（想換商品請移除本列後重新新增）
                        </p>
                      }
                    >
                      <select
                        id="cp-product"
                        value={field().state.value}
                        onBlur={field().handleBlur}
                        onChange={(e) => field().handleChange(e.currentTarget.value)}
                      >
                        <option value="">（請選擇產品）</option>
                        <For each={products()}>
                          {(p) => (
                            <option value={p.id}>
                              {p.code}｜{p.name}
                            </option>
                          )}
                        </For>
                      </select>
                    </Show>
                    <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                  </Field>
                )}
              </form.Field>

              <form.Field name="aliasName">
                {(field) => (
                  <Field>
                    <FieldLabel for="cp-alias">別名（留空＝用商品名）</FieldLabel>
                    <Input
                      id="cp-alias"
                      value={field().state.value}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                    />
                  </Field>
                )}
              </form.Field>

              <form.Field name="defaultQty" validators={defaultQtyValidators}>
                {(field) => (
                  <Field invalid={!field().state.meta.isValid}>
                    <FieldLabel for="cp-qty">預設數量（留空＝不預設）</FieldLabel>
                    <Input
                      id="cp-qty"
                      value={field().state.value}
                      onBlur={field().handleBlur}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                      placeholder="例：2.5"
                    />
                    <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                  </Field>
                )}
              </form.Field>

              <form.Field name="cutNote">
                {(field) => (
                  <Field>
                    <FieldLabel for="cp-cut-note">分切備註</FieldLabel>
                    <Input
                      id="cp-cut-note"
                      value={field().state.value}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                    />
                  </Field>
                )}
              </form.Field>

              <div class="flex justify-end gap-3">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    setFormOpen(false);
                    setEditing(null);
                  }}
                >
                  取消
                </Button>
                <Button type="submit" loading={isSubmitting()}>
                  {editing() ? "儲存" : "新增"}
                </Button>
              </div>
            </form>
          </Show>

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>商品</TableHead>
                <TableHead>別名</TableHead>
                <TableHead>預設數量</TableHead>
                <TableHead>分切備註</TableHead>
                <TableHead class="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <Show
                when={queryData(listQuery, (d) => d?.products?.length)}
                fallback={
                  <TableRow>
                    <TableCell colSpan={5}>
                      {listQuery.isFetching ? "載入中…" : "尚無專屬商品"}
                    </TableCell>
                  </TableRow>
                }
              >
                <For each={queryData(listQuery, (d) => d?.products)}>
                  {(row) => (
                    <TableRow>
                      <TableCell class="font-medium text-foreground">{`#${row.productId}`}</TableCell>
                      <TableCell class="text-foreground">{row.aliasName}</TableCell>
                      <TableCell class="text-muted-foreground">
                        {row.defaultQty === "0" || row.defaultQty === ""
                          ? "—"
                          : row.defaultQty}
                      </TableCell>
                      <TableCell class="text-muted-foreground">{row.cutNote || "—"}</TableCell>
                      <TableCell class="text-right">
                        <button
                          type="button"
                          class="font-medium text-primary hover:underline"
                          onClick={() => openForm(row)}
                        >
                          編輯
                        </button>
                        <button
                          type="button"
                          class="ml-3 font-medium text-destructive hover:underline"
                          onClick={() => void remove(row)}
                        >
                          移除
                        </button>
                      </TableCell>
                    </TableRow>
                  )}
                </For>
              </Show>
            </TableBody>
          </Table>
        </section>

        <DialogFooter>
          <DialogClose class={buttonVariants({ variant: "outline" })}>關閉</DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

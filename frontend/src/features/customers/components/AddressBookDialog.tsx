import { createForm } from "@tanstack/solid-form";
import { createQuery, useQueryClient } from "@tanstack/solid-query";
import { batch, createSignal, For, Show, type JSX } from "solid-js";
import { Code, ConnectError } from "@connectrpc/connect";
import type {
  Customer,
  CustomerAddress,
  CustomerContact,
} from "~/lib/proto/customers/v1/customer_pb";
import {
  Badge,
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
import { addressesQueryOptions, contactsQueryOptions, customerClient } from "../queries";
import { addressSchema, contactSchema } from "../schemas";
import { queryData } from "~/lib/query-data";

/** 地址類型標籤（後端 `validAddressType` 只收這三個值）。 */
const ADDRESS_TYPE_LABELS: Record<string, string> = {
  shipping: "出貨",
  billing: "請款",
  other: "其他",
};

/**
 * 錯誤訊息對照；樣板 = `CustomersPage` 的同一份 switch（各頁各自一份是本 repo 既有慣例）。
 * `InvalidArgument` 原樣透傳：後端的「recipient_name 與 address_line 必填」「email 格式非法」
 * 都是給人看的中文說明，改寫會蓋掉唯一有辨識度的訊息。
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

export interface AddressBookDialogProps {
  /** 目前選取的客戶（null＝關閉）。 */
  customer: Customer | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/** 地址類型（後端 `validAddressType` 的三個合法值）。 */
type AddressType = "shipping" | "billing" | "other";

const EMPTY_ADDRESS_VALUES: {
  type: AddressType;
  recipientName: string;
  phone: string;
  addressLine: string;
  city: string;
  postalCode: string;
  isDefault: boolean;
} = {
  type: "shipping",
  recipientName: "",
  phone: "",
  addressLine: "",
  city: "",
  postalCode: "",
  isDefault: false,
};

const EMPTY_CONTACT_VALUES = {
  name: "",
  title: "",
  email: "",
  phone: "",
  isDefault: false,
};

/**
 * 客戶地址簿與聯絡人對話框（04 計畫 3.2.1／3.2.2 的 Web 面）。
 *
 * 兩段共用一個對話框而不是各開一個：Ark 對「同一 tick 內 A 關、B 開」會對新開的 B
 * 回呼一次 `onOpenChange(false)`（群組焦點交接，見 `CustomersPage` 的 `deliveryPayload`
 * 註解），把兩個清單疊在同一個框裡就沒有這個交接要處理。
 *
 * 契約：
 * - 列表呼叫 `listAddresses`/`listContacts`（**不帶 `includeDeleted`**：軟刪除的地址/
 *   聯絡人在 3.2 是不可見資料，本框不做「含已刪除」檢視）。
 * - 欄位必填只鏡射後端（收件人、地址、姓名；email 選填但填了就要合格式），格式權威在後端。
 * - **預設地址/聯絡人**由後端判定（同類型首筆自動預設；`is_default` 可由表單主動指定，
 *   同類型其餘預設由後端清除），前端只顯示 `isDefault` 欄位、不自行推導。
 * - 每次寫入後 `invalidateQueries({ queryKey: ["customers"] })`：兩個清單的 key 都掛在
 *   `["customers"]` 前綴下，一次失效就同時刷新。
 */
export default function AddressBookDialog(props: AddressBookDialogProps) {
  const client = useQueryClient();
  // 表單層錯誤（訊息貼在該表單上方）；清單層錯誤（刪除失敗等）另放，避免互相覆蓋。
  const [addressFormError, setAddressFormError] = createSignal<string | undefined>();
  const [contactFormError, setContactFormError] = createSignal<string | undefined>();
  const [listError, setListError] = createSignal<string | null>(null);
  const [addressFormOpen, setAddressFormOpen] = createSignal(false);
  const [addressEditing, setAddressEditing] = createSignal<CustomerAddress | null>(null);
  const [contactFormOpen, setContactFormOpen] = createSignal(false);
  const [contactEditing, setContactEditing] = createSignal<CustomerContact | null>(null);

  const customerId = () => props.customer?.id ?? "";

  const addressQuery = createQuery(() => ({
    ...addressesQueryOptions(customerId()),
    enabled: props.open && customerId() !== "",
  }));

  const contactQuery = createQuery(() => ({
    ...contactsQueryOptions(customerId()),
    enabled: props.open && customerId() !== "",
  }));

  const refresh = async () => {
    await client.invalidateQueries({ queryKey: ["customers"] });
  };

  const recipientValidators = fieldValidators(addressSchema.entries.recipientName);
  const addressLineValidators = fieldValidators(addressSchema.entries.addressLine);
  const nameValidators = fieldValidators(contactSchema.entries.name);
  const emailValidators = fieldValidators(contactSchema.entries.email);

  const addressForm = createForm(() => ({
    ...appFormOptions,
    defaultValues: { ...EMPTY_ADDRESS_VALUES },
    onSubmit: async ({ value }) => {
      const current = addressEditing();
      try {
        if (current) {
          await customerClient.updateAddress({
            id: current.id,
            type: value.type,
            recipientName: value.recipientName.trim(),
            phone: value.phone,
            addressLine: value.addressLine.trim(),
            city: value.city,
            postalCode: value.postalCode,
            isDefault: value.isDefault,
          });
        } else {
          await customerClient.addAddress({
            customerId: customerId(),
            type: value.type,
            recipientName: value.recipientName.trim(),
            phone: value.phone,
            addressLine: value.addressLine.trim(),
            city: value.city,
            postalCode: value.postalCode,
            isDefault: value.isDefault,
          });
        }
        setAddressFormOpen(false);
        setAddressEditing(null);
        setAddressFormError(undefined);
        await refresh();
      } catch (err) {
        setAddressFormError(errorMessage(err));
      }
    },
  }));

  const contactForm = createForm(() => ({
    ...appFormOptions,
    defaultValues: { ...EMPTY_CONTACT_VALUES },
    onSubmit: async ({ value }) => {
      const current = contactEditing();
      try {
        if (current) {
          await customerClient.updateContact({
            id: current.id,
            name: value.name.trim(),
            title: value.title,
            email: value.email.trim(),
            phone: value.phone,
            isDefault: value.isDefault,
          });
        } else {
          await customerClient.addContact({
            customerId: customerId(),
            name: value.name.trim(),
            title: value.title,
            email: value.email.trim(),
            phone: value.phone,
            isDefault: value.isDefault,
          });
        }
        setContactFormOpen(false);
        setContactEditing(null);
        setContactFormError(undefined);
        await refresh();
      } catch (err) {
        setContactFormError(errorMessage(err));
      }
    },
  }));

  const addressSubmitting = addressForm.useSelector((state) => state.isSubmitting);
  const contactSubmitting = contactForm.useSelector((state) => state.isSubmitting);

  /** 開啟地址表單（編輯傳該筆；新增傳 null）。`reset` 整份取代（同 CustomersPage 慣例）。 */
  const openAddressForm = (address: CustomerAddress | null) => {
    setAddressEditing(address);
    setAddressFormError(undefined);
    addressForm.reset(
      address
        ? {
            type: address.type as AddressType,
            recipientName: address.recipientName,
            phone: address.phone,
            addressLine: address.addressLine,
            city: address.city,
            postalCode: address.postalCode,
            isDefault: address.isDefault,
          }
        : { ...EMPTY_ADDRESS_VALUES }
    );
    setAddressFormOpen(true);
  };

  const openContactForm = (contact: CustomerContact | null) => {
    setContactEditing(contact);
    setContactFormError(undefined);
    contactForm.reset(
      contact
        ? {
            name: contact.name,
            title: contact.title,
            email: contact.email,
            phone: contact.phone,
            isDefault: contact.isDefault,
          }
        : { ...EMPTY_CONTACT_VALUES }
    );
    setContactFormOpen(true);
  };

  const removeAddress = async (address: CustomerAddress) => {
    if (!window.confirm(`確定刪除地址「${address.recipientName}」?`)) return;
    setListError(null);
    try {
      await customerClient.deleteAddress({ id: address.id });
      await refresh();
    } catch (err) {
      setListError(errorMessage(err));
    }
  };

  const removeContact = async (contact: CustomerContact) => {
    if (!window.confirm(`確定刪除聯絡人「${contact.name}」?`)) return;
    setListError(null);
    try {
      await customerClient.deleteContact({ id: contact.id });
      await refresh();
    } catch (err) {
      setListError(errorMessage(err));
    }
  };

  const submitAddress: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    if (addressSubmitting()) return;
    setAddressFormError(undefined);
    void addressForm.handleSubmit();
  };

  const submitContact: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    if (contactSubmitting()) return;
    setContactFormError(undefined);
    void contactForm.handleSubmit();
  };

  /** 列表載入中/失敗/無資料的統一文案（兩段共用，避免兩份判斷漂移）。 */
  const listStateText = (loading: boolean, emptyText: string) =>
    loading ? "載入中…" : emptyText;

  return (
    <Dialog
      open={props.open && props.customer !== null}
      onOpenChange={(open) => {
        if (!open) {
          props.onOpenChange(false);
          // 關閉時收起表單：下次開別的客戶不該看到上一個人的表單狀態。
          batch(() => {
            setAddressFormOpen(false);
            setContactFormOpen(false);
            setAddressEditing(null);
            setContactEditing(null);
            setListError(null);
          });
        }
      }}
    >
      <DialogContent class="max-w-3xl">
        <DialogHeader>
          <DialogTitle>地址簿與聯絡人</DialogTitle>
          <DialogDescription>
            {props.customer ? `${props.customer.name}（${props.customer.customerCode}）` : ""}
          </DialogDescription>
        </DialogHeader>

        <Show when={listError()}>
          {(message) => (
            <p class="rounded-lg bg-destructive/15 px-3 py-2 text-sm text-destructive" role="alert">
              {message()}
            </p>
          )}
        </Show>

        {/* 地址簿 */}
        <section class="space-y-2">
          <div class="flex items-center justify-between">
            <h3 class="font-semibold text-foreground">地址簿</h3>
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={() => (addressFormOpen() ? setAddressFormOpen(false) : openAddressForm(null))}
            >
              {addressFormOpen() ? "取消新增" : "新增地址"}
            </Button>
          </div>

          <Show when={addressFormOpen()}>
            <form
              onSubmit={submitAddress}
              class="space-y-2 rounded-lg border border-border bg-muted p-3"
            >
                <Show when={addressFormError()}>
                  {(message) => (
                    <p class="text-sm text-destructive" role="alert">
                      {message()}
                    </p>
                  )}
                </Show>
                <div class="grid gap-3 sm:grid-cols-2">
                  <Field>
                    <FieldLabel for="address-type">類型</FieldLabel>
                    <select
                      id="address-type"
                      value={addressForm.state.values.type}
                      onChange={(e) =>
                        addressForm.setFieldValue(
                          "type",
                          e.currentTarget.value as AddressType
                        )
                      }
                    >
                      <option value="shipping">出貨</option>
                      <option value="billing">請款</option>
                      <option value="other">其他</option>
                    </select>
                  </Field>
                  <addressForm.Field name="recipientName" validators={recipientValidators}>
                    {(field) => (
                      <Field invalid={!field().state.meta.isValid}>
                        <FieldLabel for="address-recipient">收件人 *</FieldLabel>
                        <Input
                          id="address-recipient"
                          required
                          value={field().state.value}
                          onBlur={field().handleBlur}
                          onInput={(e) => field().handleChange(e.currentTarget.value)}
                        />
                        <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                      </Field>
                    )}
                  </addressForm.Field>
                  <addressForm.Field name="addressLine" validators={addressLineValidators}>
                    {(field) => (
                      <Field invalid={!field().state.meta.isValid}>
                        <FieldLabel for="address-line">地址 *</FieldLabel>
                        <Input
                          id="address-line"
                          required
                          value={field().state.value}
                          onBlur={field().handleBlur}
                          onInput={(e) => field().handleChange(e.currentTarget.value)}
                          placeholder="台北市信義路一段 1 號"
                        />
                        <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                      </Field>
                    )}
                  </addressForm.Field>
                  <addressForm.Field name="phone">
                    {(field) => (
                      <Field>
                        <FieldLabel for="address-phone">電話</FieldLabel>
                        <Input
                          id="address-phone"
                          value={field().state.value}
                          onInput={(e) => field().handleChange(e.currentTarget.value)}
                        />
                      </Field>
                    )}
                  </addressForm.Field>
                  <addressForm.Field name="city">
                    {(field) => (
                      <Field>
                        <FieldLabel for="address-city">縣市</FieldLabel>
                        <Input
                          id="address-city"
                          value={field().state.value}
                          onInput={(e) => field().handleChange(e.currentTarget.value)}
                        />
                      </Field>
                    )}
                  </addressForm.Field>
                  <addressForm.Field name="postalCode">
                    {(field) => (
                      <Field>
                        <FieldLabel for="address-postal">郵遞區號</FieldLabel>
                        <Input
                          id="address-postal"
                          value={field().state.value}
                          onInput={(e) => field().handleChange(e.currentTarget.value)}
                        />
                      </Field>
                    )}
                  </addressForm.Field>
                </div>
                <label class="flex cursor-pointer items-center gap-2 text-sm text-foreground">
                  <addressForm.Field name="isDefault">
                    {(field) => (
                      <input
                        type="checkbox"
                        checked={field().state.value}
                        onChange={(e) => field().handleChange(e.currentTarget.checked)}
                      />
                    )}
                  </addressForm.Field>
                  設為預設（同類型其餘預設由後端清除）
                </label>
                <div class="flex justify-end gap-3">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setAddressFormOpen(false);
                      setAddressEditing(null);
                    }}
                  >
                    取消
                  </Button>
                  <Button type="submit" loading={addressSubmitting()}>
                    {addressEditing() ? "儲存" : "新增"}
                  </Button>
                </div>
              </form>
          </Show>

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>類型</TableHead>
                <TableHead>收件人</TableHead>
                <TableHead>地址</TableHead>
                <TableHead>電話</TableHead>
                <TableHead class="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <Show
                when={queryData(addressQuery, (d) => d?.addresses?.length)}
                fallback={
                  <TableRow>
                    <TableCell colSpan={5}>
                      {listStateText(addressQuery.isFetching, "尚無地址")}
                    </TableCell>
                  </TableRow>
                }
              >
                <For each={queryData(addressQuery, (d) => d?.addresses)}>
                  {(address) => (
                    <TableRow>
                      <TableCell>
                        <Badge variant="secondary">
                          {ADDRESS_TYPE_LABELS[address.type] ?? address.type}
                        </Badge>
                        <Show when={address.isDefault}>
                          <Badge variant="info" class="ml-2">
                            預設
                          </Badge>
                        </Show>
                      </TableCell>
                      <TableCell class="font-medium text-foreground">
                        {address.recipientName}
                      </TableCell>
                      <TableCell class="text-muted-foreground">{address.addressLine}</TableCell>
                      <TableCell class="text-muted-foreground">{address.phone || "—"}</TableCell>
                      <TableCell class="text-right">
                        <button
                          type="button"
                          class="font-medium text-primary hover:underline"
                          onClick={() => openAddressForm(address)}
                        >
                          編輯
                        </button>
                        <button
                          type="button"
                          class="ml-3 font-medium text-destructive hover:underline"
                          onClick={() => void removeAddress(address)}
                        >
                          刪除
                        </button>
                      </TableCell>
                    </TableRow>
                  )}
                </For>
              </Show>
            </TableBody>
          </Table>
        </section>

        {/* 聯絡人 */}
        <section class="space-y-2">
          <div class="flex items-center justify-between">
            <h3 class="font-semibold text-foreground">聯絡人</h3>
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={() => (contactFormOpen() ? setContactFormOpen(false) : openContactForm(null))}
            >
              {contactFormOpen() ? "取消新增" : "新增聯絡人"}
            </Button>
          </div>

          <Show when={contactFormOpen()}>
            <form
              onSubmit={submitContact}
              class="space-y-2 rounded-lg border border-border bg-muted p-3"
            >
              <Show when={contactFormError()}>
                {(message) => (
                  <p class="text-sm text-destructive" role="alert">
                    {message()}
                  </p>
                )}
              </Show>
              <div class="grid gap-3 sm:grid-cols-2">
                <contactForm.Field name="name" validators={nameValidators}>
                  {(field) => (
                    <Field invalid={!field().state.meta.isValid}>
                      <FieldLabel for="contact-name">姓名 *</FieldLabel>
                      <Input
                        id="contact-name"
                        required
                        value={field().state.value}
                        onBlur={field().handleBlur}
                        onInput={(e) => field().handleChange(e.currentTarget.value)}
                      />
                      <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                    </Field>
                  )}
                </contactForm.Field>
                <contactForm.Field name="title">
                  {(field) => (
                    <Field>
                      <FieldLabel for="contact-title">職稱</FieldLabel>
                      <Input
                        id="contact-title"
                        value={field().state.value}
                        onInput={(e) => field().handleChange(e.currentTarget.value)}
                      />
                    </Field>
                  )}
                </contactForm.Field>
                <contactForm.Field name="email" validators={emailValidators}>
                  {(field) => (
                    <Field invalid={!field().state.meta.isValid}>
                      <FieldLabel for="contact-email">Email</FieldLabel>
                      <Input
                        id="contact-email"
                        type="email"
                        value={field().state.value}
                        onBlur={field().handleBlur}
                        onInput={(e) => field().handleChange(e.currentTarget.value)}
                      />
                      <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                    </Field>
                  )}
                </contactForm.Field>
                <contactForm.Field name="phone">
                  {(field) => (
                    <Field>
                      <FieldLabel for="contact-phone">電話</FieldLabel>
                      <Input
                        id="contact-phone"
                        value={field().state.value}
                        onInput={(e) => field().handleChange(e.currentTarget.value)}
                      />
                    </Field>
                  )}
                </contactForm.Field>
              </div>
              <label class="flex cursor-pointer items-center gap-2 text-sm text-foreground">
                <contactForm.Field name="isDefault">
                  {(field) => (
                    <input
                      type="checkbox"
                      checked={field().state.value}
                      onChange={(e) => field().handleChange(e.currentTarget.checked)}
                    />
                  )}
                </contactForm.Field>
                設為預設聯絡人（其餘預設由後端清除）
              </label>
              <div class="flex justify-end gap-3">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    setContactFormOpen(false);
                    setContactEditing(null);
                  }}
                >
                  取消
                </Button>
                <Button type="submit" loading={contactSubmitting()}>
                  {contactEditing() ? "儲存" : "新增"}
                </Button>
              </div>
            </form>
          </Show>

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>姓名</TableHead>
                <TableHead>職稱</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>電話</TableHead>
                <TableHead class="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <Show
                when={queryData(contactQuery, (d) => d?.contacts?.length)}
                fallback={
                  <TableRow>
                    <TableCell colSpan={5}>
                      {listStateText(contactQuery.isFetching, "尚無聯絡人")}
                    </TableCell>
                  </TableRow>
                }
              >
                <For each={queryData(contactQuery, (d) => d?.contacts)}>
                  {(contact) => (
                    <TableRow>
                      <TableCell class="font-medium text-foreground">
                        {contact.name}
                        <Show when={contact.isDefault}>
                          <Badge variant="info" class="ml-2">
                            預設
                          </Badge>
                        </Show>
                      </TableCell>
                      <TableCell class="text-muted-foreground">{contact.title || "—"}</TableCell>
                      <TableCell class="text-muted-foreground">{contact.email || "—"}</TableCell>
                      <TableCell class="text-muted-foreground">{contact.phone || "—"}</TableCell>
                      <TableCell class="text-right">
                        <button
                          type="button"
                          class="font-medium text-primary hover:underline"
                          onClick={() => openContactForm(contact)}
                        >
                          編輯
                        </button>
                        <button
                          type="button"
                          class="ml-3 font-medium text-destructive hover:underline"
                          onClick={() => void removeContact(contact)}
                        >
                          刪除
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

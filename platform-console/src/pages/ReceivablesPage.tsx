import { createMutation, createQuery, useQueryClient } from "@tanstack/solid-query";
import { createSignal, For, Show } from "solid-js";
import { Badge } from "@ui/badge";
import { Button } from "@ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@ui/dialog";
import { Field, FieldDescription, FieldLabel } from "@ui/field";
import { Input } from "@ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@ui/table";
import { Pagination } from "../components/pagination";
import { EmptyState, PageShell, queryBoundary } from "../components/page";
import { Select } from "../components/select";
import { WriteForm } from "../components/write";
import { platform } from "../lib/api";
import { describeError } from "../lib/errors";
import type { Receivable } from "../lib/proto/platform/v1/platform_pb";
import { downloadCsv, receivableStatus, receivablesCsv } from "../lib/receivables";

/**
 * 待收款（未付期別，含逾期）與人工收款。
 *
 * 清單與分頁都交給後端（`ListReceivables`）；「逾期」是期末已過的事實，由前端標記
 * （服務層不改語意，見其註解）。
 *
 * **記收款**走 T9 的 `RecordPayment`（不另寫一條入帳路徑）：狀態機、期別金額驗證、事件與稽核
 * 都在後端 billing 的同一個交易裡。三個欄位語意必須讓營運看得懂：
 * 1. **金額留空＝採用期別快照金額**（proto：空字串＝快照；v1 不支援部分付款）——
 *    留空時線路上根本不送 `amount`，所以畫面直接把「會入帳多少」寫出來。
 * 2. **交易號**是同一 provider 的重送比對依據（相同交易號重送＝no-op，不重複入帳）。
 * 3. **原因必填**（後端每個寫入都擋空白 reason），且會寫進平台稽核。
 */
const PAGE_SIZE = 50;

/** 與後端 money.ParseCents 相同的形狀（整數位 ＋ 最多兩位小數）；留空代表採用期別快照。 */
const MONEY_PATTERN = /^\d+(\.\d{1,2})?$/;

const PROVIDERS = [
  { value: "manual", label: "人工匯款" },
  { value: "ecpay", label: "綠界" },
  { value: "newebpay", label: "藍新" },
  { value: "tappay", label: "台新" },
  { value: "stripe", label: "Stripe" },
];

/**
 * 收款表單。欄位狀態放在對話框內容的子元件裡：`DialogContent` 關閉時整棵子樹卸載，
 * 欄位自然清空（signal 放外層包裝元件就會殘留上一個租戶的金額與交易號）。
 */
function PaymentForm(props: { row: Receivable; onDone: () => void }) {
  const queryClient = useQueryClient();
  const [amount, setAmount] = createSignal("");
  const [provider, setProvider] = createSignal("manual");
  const [externalRef, setExternalRef] = createSignal("");
  const [invoiceNo, setInvoiceNo] = createSignal("");
  const [note, setNote] = createSignal("");

  const mutation = createMutation(() => ({
    mutationFn: (reason: string) =>
      platform.recordPayment({
        companyId: props.row.companyId,
        // 明確送出期別：0 代表「當前（最新）一期」，但這一列本來就是某一期。
        periodNo: props.row.periodNo,
        amount: amount().trim(),
        provider: provider(),
        externalRef: externalRef().trim(),
        invoiceNo: invoiceNo().trim(),
        note: note().trim(),
        reason,
      }),
    onSuccess: () => {
      // 收款會寫入平台稽核 → 兩份清單都要重查（不靠手動重整）。
      void queryClient.invalidateQueries({ queryKey: ["receivables"] });
      void queryClient.invalidateQueries({ queryKey: ["platform-audit"] });
      props.onDone();
    },
  }));

  return (
    <WriteForm
      submitLabel="確認收款"
      pending={mutation.isPending}
      error={mutation.isError ? describeError(mutation.error) : undefined}
      validate={() => {
        const value = amount().trim();
        if (value === "") return undefined;
        if (!MONEY_PATTERN.test(value)) {
          return "金額格式錯誤：請輸入數字，最多兩位小數（例：1500.00）；留空則採用期別快照金額。";
        }
        // `AmountCents=0` 在後端的語意是「採用期別快照」，**不是**「本期不收」：
        // 填 0 會把整期記成已收（全額入帳）。要表達不收費得走別的路徑，不能靠這個欄位。
        if (Number(value) === 0) {
          return "金額必須大於 0：後端把 0 當成「採用期別快照」，填 0 會把整期記成已收。";
        }
        return undefined;
      }}
      onSubmit={(reason) => mutation.mutate(reason)}
    >
      <p class="text-sm text-muted-foreground">
        {props.row.companyName}（方案 {props.row.planCode}）第 {props.row.periodNo} 期，
        期末 {props.row.periodEnd}，期別金額 {props.row.amount}。
      </p>

      <Field>
        <FieldLabel for="pay-amount">金額</FieldLabel>
        <Input
          id="pay-amount"
          value={amount()}
          placeholder={props.row.amount}
          onInput={(e) => setAmount(e.currentTarget.value)}
        />
        <FieldDescription>
          留空＝採用期別快照金額 {props.row.amount}；填了就必須大於 0（不支援部分付款；短收／溢收請記於備註）。
        </FieldDescription>
      </Field>

      <Field>
        <FieldLabel for="pay-provider">收款方式</FieldLabel>
        <Select
          id="pay-provider"
          value={provider()}
          onChange={(e) => setProvider(e.currentTarget.value)}
        >
          <For each={PROVIDERS}>
            {(it) => <option value={it.value}>{it.label}</option>}
          </For>
        </Select>
      </Field>

      <Field>
        <FieldLabel for="pay-ref">交易號</FieldLabel>
        <Input
          id="pay-ref"
          value={externalRef()}
          placeholder="例：匯款帳號末五碼、金流交易序號"
          onInput={(e) => setExternalRef(e.currentTarget.value)}
        />
        <FieldDescription>
          同一個收款方式＋交易號重送視為重送（不重複入帳）；換了交易號再送同一期會被視為重複收款。
        </FieldDescription>
      </Field>

      <Field>
        <FieldLabel for="pay-invoice">發票號</FieldLabel>
        <Input
          id="pay-invoice"
          value={invoiceNo()}
          onInput={(e) => setInvoiceNo(e.currentTarget.value)}
        />
      </Field>

      <Field>
        <FieldLabel for="pay-note">備註</FieldLabel>
        <Input id="pay-note" value={note()} onInput={(e) => setNote(e.currentTarget.value)} />
        <FieldDescription>短收／溢收的差異記在這裡，不會改變期別金額。</FieldDescription>
      </Field>
    </WriteForm>
  );
}

function PaymentDialog(props: {
  row: Receivable | undefined;
  onOpenChange: (open: boolean) => void;
  onDone: () => void;
}) {
  return (
    <Dialog
      open={props.row !== undefined}
      onOpenChange={props.onOpenChange}
      lazyMount
      unmountOnExit
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>記收款：{props.row?.companyName}</DialogTitle>
          <DialogDescription>
            確認收款後期別轉為已付款、訂閱回到訂閱中，並在平台稽核留下一筆紀錄（含你的 operator 身分與原因）。
          </DialogDescription>
        </DialogHeader>
        <Show when={props.row}>
          {(row) => <PaymentForm row={row()} onDone={props.onDone} />}
        </Show>
      </DialogContent>
    </Dialog>
  );
}

export default function ReceivablesPage() {
  const [page, setPage] = createSignal(1);
  const [paying, setPaying] = createSignal<Receivable | undefined>();

  const receivables = createQuery(() => ({
    queryKey: ["receivables", page()],
    queryFn: () => platform.listReceivables({ page: page(), pageSize: PAGE_SIZE }),
  }));

  return (
    <PageShell
      title="待收款"
      description="所有未付期別（平台自營公司不算租戶，不列入）；期末已過者標記為逾期。匯出 CSV 只含目前這一頁。"
    >
      {queryBoundary(receivables, (data) => (
        <div class="space-y-4">
          <Show
            when={data.rows.length > 0}
            fallback={
              <EmptyState>
                {page() > 1
                  ? `第 ${page()} 頁沒有未付期別（未付的期別可能在前面的頁次）。`
                  : "沒有未付期別。"}
              </EmptyState>
            }
          >
            <div class="flex items-center gap-3">
              <Button
                variant="outline"
                size="sm"
                onClick={() => downloadCsv("receivables.csv", receivablesCsv(data.rows))}
              >
                匯出 CSV（本頁）
              </Button>
              <span class="text-sm text-muted-foreground">欄位：公司、方案、期別、金額、到期日、狀態。</span>
            </div>

            <Table aria-label="待收款清單">
              <TableHeader>
                <TableRow>
                  <TableHead scope="col">公司</TableHead>
                  <TableHead scope="col">方案</TableHead>
                  <TableHead scope="col">期別</TableHead>
                  <TableHead scope="col">金額</TableHead>
                  <TableHead scope="col">期末</TableHead>
                  <TableHead scope="col">狀態</TableHead>
                  <TableHead scope="col">動作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <For each={data.rows}>
                  {(row) => (
                    <TableRow>
                      <TableCell>{row.companyName}</TableCell>
                      <TableCell>{row.planCode}</TableCell>
                      <TableCell>第 {row.periodNo} 期</TableCell>
                      <TableCell>{row.amount}</TableCell>
                      <TableCell>{row.periodEnd}</TableCell>
                      <TableCell>
                        <Badge
                          variant={receivableStatus(row) === "逾期" ? "warning" : "secondary"}
                        >
                          {receivableStatus(row)}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Button variant="outline" size="sm" onClick={() => setPaying(row)}>
                          記收款
                        </Button>
                      </TableCell>
                    </TableRow>
                  )}
                </For>
              </TableBody>
            </Table>

          </Show>

          {/* 分頁控制**永遠**在（含這一頁被清空時）：關在「有資料」分支裡就會變成死路 ——
              畫面只說「沒有未付期別」，卻沒有上一頁可按。 */}
          <Pagination
            page={page()}
            pageSize={data.pagination?.pageSize ?? PAGE_SIZE}
            total={data.pagination?.total ?? data.rows.length}
            onPage={setPage}
          />
        </div>
      ))}

      <PaymentDialog
        row={paying()}
        onOpenChange={(open) => {
          if (!open) setPaying(undefined);
        }}
        onDone={() => setPaying(undefined)}
      />
    </PageShell>
  );
}

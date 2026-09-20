import { createMutation, createQuery, useQueryClient } from "@tanstack/solid-query";
import { createSignal, For, Show } from "solid-js";
import { Badge } from "@ui/badge";
import { Button } from "@ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@ui/card";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@ui/dialog";
import { Field, FieldDescription, FieldLabel } from "@ui/field";
import { Input } from "@ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@ui/table";
import { LabelledCheckbox } from "../components/labelled-checkbox";
import { EmptyState, PageShell, queryBoundary } from "../components/page";
import { Select } from "../components/select";
import { WriteForm } from "../components/write";
import { platform } from "../lib/api";
import { describeError } from "../lib/errors";

/**
 * 方案與價目。
 *
 * 兩個**生效語意**必須讓營運看見，而且不能混成一句（講錯會害人以為配額也要等下個月）：
 * 1. **金額／期別＝下一期生效**：調價是新增一列價格史（後端 `UpsertPlanPriceTx`），現行價取
 *    `effective_from` 最新者；既有期別的金額在開立時已快照 → 只影響之後新開的期別，不回溯改帳。
 * 2. **權益（配額與功能開關）＝立即生效**：`SetPlanEntitlement` 提交後會 `invalidateAll`，
 *    該方案所有租戶的判定立刻改變；租戶層級的例外（override）仍優先於方案。
 *
 * 金額前端只做**格式檢查**（後端 `money.ParseCents` 才是真偽的決定者：負數、超過兩位小數、
 * 非數字都回 SYS-1001）。`reason` 必填同樣由後端擋，這裡先擋只是少跑一趟。
 */

/** 單列價格的輸入檢查：與後端 money.ParseCents 相同的形狀（整數位 ＋ 最多兩位小數）。 */
const MONEY_PATTERN = /^\d+(\.\d{1,2})?$/;

/**
 * 調價表單。欄位狀態放在**對話框內容的子元件**裡：`DialogContent` 關閉時整棵子樹卸載，
 * 欄位自然清空——signal 放在外層包裝元件就會殘留（對甲方案填到一半取消、再開乙方案會看到甲的金額）。
 */
function PriceForm(props: { planCode: string; onClose: () => void; onDone: () => void }) {
  const [cycle, setCycle] = createSignal("monthly");
  const [basePrice, setBasePrice] = createSignal("");
  const [seatPrice, setSeatPrice] = createSignal("");
  const [currency, setCurrency] = createSignal("");

  const mutation = createMutation(() => ({
    mutationFn: (input: {
      billingCycle: string;
      basePrice: string;
      seatPrice: string;
      currency: string;
      reason: string;
    }) => platform.upsertPlanPrice({ planCode: props.planCode, ...input }),
    onSuccess: () => {
      props.onClose();
      props.onDone();
    },
  }));

  return (
    <WriteForm
      submitLabel="儲存價目"
      pending={mutation.isPending}
      error={mutation.isError ? describeError(mutation.error) : undefined}
      validate={() => {
        if (!MONEY_PATTERN.test(basePrice().trim())) {
          return "基本價金額格式錯誤：請輸入數字，最多兩位小數（例：1500.00）；不接受負金額。";
        }
        if (!MONEY_PATTERN.test(seatPrice().trim())) {
          return "單席價金額格式錯誤：請輸入數字，最多兩位小數（例：200.00）。";
        }
        return undefined;
      }}
      onSubmit={(reason) =>
        mutation.mutate({
          billingCycle: cycle(),
          basePrice: basePrice().trim(),
          seatPrice: seatPrice().trim(),
          currency: currency().trim(),
          reason,
        })
      }
    >
      <Field>
        <FieldLabel for="price-cycle">計費週期 *</FieldLabel>
        <Select id="price-cycle" value={cycle()} onChange={(e) => setCycle(e.currentTarget.value)}>
          <option value="monthly">月繳</option>
          <option value="yearly">年繳</option>
        </Select>
        <FieldDescription>週期決定期別的加期方式（加一個月或加一年）。</FieldDescription>
      </Field>

      <Field>
        <FieldLabel for="price-base">基本價 *</FieldLabel>
        <Input
          id="price-base"
          value={basePrice()}
          placeholder="1500.00"
          onInput={(e) => setBasePrice(e.currentTarget.value)}
        />
      </Field>

      <Field>
        <FieldLabel for="price-seat">單席價 *</FieldLabel>
        <Input
          id="price-seat"
          value={seatPrice()}
          placeholder="200.00"
          onInput={(e) => setSeatPrice(e.currentTarget.value)}
        />
      </Field>

      <Field>
        <FieldLabel for="price-currency">幣別</FieldLabel>
        <Input
          id="price-currency"
          value={currency()}
          placeholder="TWD"
          onInput={(e) => setCurrency(e.currentTarget.value)}
        />
        <FieldDescription>留空＝TWD。</FieldDescription>
      </Field>
    </WriteForm>
  );
}

function PriceDialog(props: {
  planCode: string | undefined;
  planName: string;
  onOpenChange: (open: boolean) => void;
  onDone: () => void;
}) {
  return (
    <Dialog
      open={props.planCode !== undefined}
      onOpenChange={props.onOpenChange}
      lazyMount
      unmountOnExit
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>調整價目：{props.planName}</DialogTitle>
          <DialogDescription>
            金額下一期生效：調價是新增一列價格史，從現在起成為該計費週期的現行價；
            已開立的期別金額已有快照，不受影響（不會回溯改帳）。
          </DialogDescription>
        </DialogHeader>

        <Show when={props.planCode}>
          {(planCode) => (
            <PriceForm
              planCode={planCode()}
              onClose={() => props.onOpenChange(false)}
              onDone={props.onDone}
            />
          )}
        </Show>
      </DialogContent>
    </Dialog>
  );
}

/**
 * 方案權益表單。配額與功能開關**立即**改變判定（後端 `SetPlanEntitlement` 會 `invalidateAll`）；
 * 只有金額／期別是下一期生效——兩件事不要混在一個句子裡講，否則營運會以為改了配額還要等下個月。
 */
function EntitlementForm(props: { planCode: string; onClose: () => void; onDone: () => void }) {
  const [featureCode, setFeatureCode] = createSignal("");
  const [enabled, setEnabled] = createSignal(false);
  const [limitSet, setLimitSet] = createSignal(false);
  const [limitValue, setLimitValue] = createSignal("");

  const matrix = createQuery(() => ({
    queryKey: ["entitlements", props.planCode],
    queryFn: () => platform.getPlanEntitlements({ planCode: props.planCode }),
  }));

  const mutation = createMutation(() => ({
    mutationFn: (input: {
      featureCode: string;
      enabled: boolean;
      limitSet: boolean;
      limitValue: bigint;
      reason: string;
    }) => platform.setPlanEntitlement({ planCode: props.planCode, ...input }),
    onSuccess: () => {
      props.onClose();
      props.onDone();
    },
  }));

  // 選到的功能的現行設定（唯讀顯示）：不預填進表單，免得覆蓋 operator 的輸入，
  // 但一定要看得見「現在是什麼」，否則改權益是閉著眼睛改。
  const current = () => matrix.data?.entitlements.find((e) => e.featureCode === featureCode());

  return (
    <div>
      {queryBoundary(matrix, (data) => (
        <WriteForm
          submitLabel="儲存權益"
          pending={mutation.isPending}
          error={mutation.isError ? describeError(mutation.error) : undefined}
          validate={() => {
            if (featureCode() === "") return "請先選擇功能。";
            if (limitSet() && !/^\d+$/.test(limitValue().trim())) {
              return "上限不得為負：請填不小於 0 的整數（負的上限等於把功能對所有用該方案的租戶關掉；不限額請不要勾「指定上限」）。";
            }
            return undefined;
          }}
          onSubmit={(reason) =>
            mutation.mutate({
              featureCode: featureCode(),
              enabled: enabled(),
              limitSet: limitSet(),
              limitValue: limitSet() ? BigInt(limitValue().trim()) : 0n,
              reason,
            })
          }
        >
          <Field>
            <FieldLabel for="entitlement-feature">功能 *</FieldLabel>
            <Select
              id="entitlement-feature"
              value={featureCode()}
              onChange={(e) => setFeatureCode(e.currentTarget.value)}
            >
              <option value="">請選擇</option>
              <For each={data.features}>
                {(f) => (
                  <option value={f.code}>
                    {f.code}（{f.type === "integer" ? "數量" : "開關"}）
                  </option>
                )}
              </For>
            </Select>
            <Show when={featureCode() !== ""}>
              <FieldDescription>
                {current()
                  ? `目前設定：${current()?.enabled ? "啟用" : "停用"}・${
                      current()?.limitSet ? `上限 ${current()?.limitValue}` : "不限"
                    }`
                  : "目前設定：未設定（未開通）"}
              </FieldDescription>
            </Show>
          </Field>

          <div class="flex flex-wrap gap-4">
            <LabelledCheckbox label="啟用" checked={enabled()} onCheckedChange={setEnabled} />
            <LabelledCheckbox label="指定上限" checked={limitSet()} onCheckedChange={setLimitSet} />
          </div>

          <Show when={limitSet()}>
            <Field>
              <FieldLabel for="entitlement-limit">上限值 *</FieldLabel>
              <Input
                id="entitlement-limit"
                inputmode="numeric"
                value={limitValue()}
                placeholder="10"
                onInput={(e) => setLimitValue(e.currentTarget.value)}
              />
              <FieldDescription>0 是有效上限（等於不能用）；不限額請不要勾「指定上限」。</FieldDescription>
            </Field>
          </Show>
        </WriteForm>
      ))}
    </div>
  );
}

function EntitlementDialog(props: {
  planCode: string | undefined;
  planName: string;
  onOpenChange: (open: boolean) => void;
  onDone: () => void;
}) {
  return (
    <Dialog
      open={props.planCode !== undefined}
      onOpenChange={props.onOpenChange}
      lazyMount
      unmountOnExit
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>方案權益：{props.planName}</DialogTitle>
          <DialogDescription>
            配額與功能開關立即生效（改變該方案所有租戶的判定）；期別金額不受影響。
            上限的「不指定」＝不限額，與上限 0（等於不能用）不同。
          </DialogDescription>
        </DialogHeader>

        <Show when={props.planCode}>
          {(planCode) => (
            <EntitlementForm
              planCode={planCode()}
              onClose={() => props.onOpenChange(false)}
              onDone={props.onDone}
            />
          )}
        </Show>
      </DialogContent>
    </Dialog>
  );
}

export default function PlansPage() {
  const queryClient = useQueryClient();
  const plans = createQuery(() => ({ queryKey: ["plans"], queryFn: () => platform.listPlans({}) }));

  const [priceFor, setPriceFor] = createSignal<{ code: string; name: string } | undefined>(undefined);
  const [entitlementFor, setEntitlementFor] = createSignal<
    { code: string; name: string } | undefined
  >(undefined);

  // 價目與權益都影響判定（價目決定新期別金額、權益決定方案上限）→ 兩者都讓查詢失效。
  const refresh = () => {
    void queryClient.invalidateQueries({ queryKey: ["plans"] });
    void queryClient.invalidateQueries({ queryKey: ["entitlements"] });
  };

  return (
    <PageShell
      title="方案與價目"
      description="方案定義與各計費週期價格。金額為後端 money.ParseCents 的十進位字串（最多兩位小數）。"
    >
      {queryBoundary(plans, (data) =>
        data.plans.length === 0 ? (
          <EmptyState>尚未定義任何方案。</EmptyState>
        ) : (
          <div class="space-y-4">
            <p class="text-sm text-muted-foreground">
              金額：下一期生效——調價新增一列價格史，已開立的期別金額已有快照，不會被回溯改帳。
              權益：立即生效——配額與功能開關立刻改變該方案所有租戶的判定（租戶層級的例外優先於方案）。
            </p>

            <For each={data.plans}>
              {(plan) => (
                <Card>
                  <CardHeader class="flex flex-row flex-wrap items-center justify-between gap-2">
                    <CardTitle class="flex items-center gap-2">
                      {plan.name}
                      <Badge variant={plan.status === "active" ? "success" : "secondary"}>
                        {plan.status === "active" ? "上架中" : plan.status}
                      </Badge>
                      <span class="text-sm font-normal text-muted-foreground">{plan.code}</span>
                    </CardTitle>
                    <div class="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        aria-label={`調價：${plan.name}`}
                        onClick={() => setPriceFor({ code: plan.code, name: plan.name })}
                      >
                        調價
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        aria-label={`權益：${plan.name}`}
                        onClick={() => setEntitlementFor({ code: plan.code, name: plan.name })}
                      >
                        權益
                      </Button>
                    </div>
                  </CardHeader>
                  <CardContent>
                    {plan.prices.length === 0 ? (
                      <EmptyState>此方案尚未設定價目：新開的期別會取不到金額。</EmptyState>
                    ) : (
                      <Table aria-label={`${plan.name} 價目`}>
                        <TableHeader>
                          <TableRow>
                            <TableHead scope="col">計費週期</TableHead>
                            <TableHead scope="col">基本價</TableHead>
                            <TableHead scope="col">單席價</TableHead>
                            <TableHead scope="col">幣別</TableHead>
                            <TableHead scope="col">現行價生效日</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          <For each={plan.prices}>
                            {(price) => (
                              <TableRow>
                                <TableCell>
                                  {price.billingCycle === "yearly" ? "年繳" : "月繳"}
                                </TableCell>
                                <TableCell>{price.basePrice}</TableCell>
                                <TableCell>{price.seatPrice}</TableCell>
                                <TableCell>{price.currency || "TWD"}</TableCell>
                                <TableCell>{price.effectiveFrom}</TableCell>
                              </TableRow>
                            )}
                          </For>
                        </TableBody>
                      </Table>
                    )}
                  </CardContent>
                </Card>
              )}
            </For>
          </div>
        ),
      )}

      <PriceDialog
        planCode={priceFor()?.code}
        planName={priceFor()?.name ?? ""}
        onOpenChange={(open) => {
          if (!open) setPriceFor(undefined);
        }}
        onDone={refresh}
      />
      <EntitlementDialog
        planCode={entitlementFor()?.code}
        planName={entitlementFor()?.name ?? ""}
        onOpenChange={(open) => {
          if (!open) setEntitlementFor(undefined);
        }}
        onDone={refresh}
      />
    </PageShell>
  );
}

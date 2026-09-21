import { createMutation, createQuery, useQueryClient } from "@tanstack/solid-query";
import { useParams } from "@tanstack/solid-router";
import { createEffect, createSignal, For, Show } from "solid-js";
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
import { SubscriptionBadge } from "../components/status";
import { WriteForm } from "../components/write";
import { platform } from "../lib/api";
import { describeError } from "../lib/errors";
import { isExpired, projectTenantEntitlements, type ProjectedEntitlement } from "../lib/entitlements";
// 列上帶的是 RPC 回來的例外（有 id 才能撤銷）；投影只用到其中幾個欄位。
import type { TenantOverride } from "../lib/proto/platform/v1/platform_pb";

/**
 * 租戶詳情：訂閱概況 ＋ 平台對該租戶的功能例外（override）＋ 權益現況。
 *
 * **前端守衛／disable 不構成授權**：這裡的檢查只是為了少跑一趟白工的 RPC，
 * 真正的決策者是後端（`platformReason`、`limit_value < 0` 的守衛、operator 身分）。
 * 因此按鈕一律可見——console 拿不到 operator 的角色（v1 沒有「查自己」的 RPC），
 * 依角色猜著隱藏只會變成假的安全感；被拒絕時由 `describeError` 說清楚要做什麼。
 */

/** 上限／啟用的顯示：`limit_set` 是「有沒有指定上限」，0 是有效值，不得混為一談。 */
function describeValue(enabled: boolean, limitSet: boolean, limitValue: bigint): string {
  if (!enabled) return "停用";
  return limitSet ? `啟用（上限 ${limitValue}）` : "啟用（不限）";
}

/**
 * 例外的新增（SetTenantOverride）：兩個維度都用 `*_set` 表達，未勾＝不覆寫該維度。
 *
 * 欄位狀態放在**對話框內容的子元件**裡：`DialogContent` 關閉時整棵子樹卸載（`lazyMount` ＋
 * `unmountOnExit`），欄位自然清空。signal 若放在外層包裝元件就會殘留——那個元件不隨內容卸載，
 * 於是「對甲租戶填到一半按取消、再對乙租戶開」會看到甲的承諾。
 */
function SetOverrideForm(props: { companyId: string; onClose: () => void; onDone: () => void }) {
  const [featureCode, setFeatureCode] = createSignal("");
  const [enabledSet, setEnabledSet] = createSignal(false);
  const [enabled, setEnabled] = createSignal(true);
  const [limitSet, setLimitSet] = createSignal(false);
  const [limitValue, setLimitValue] = createSignal("");
  const [owner, setOwner] = createSignal("");
  const [expiresAt, setExpiresAt] = createSignal("");

  const mutation = createMutation(() => ({
    mutationFn: (input: {
      featureCode: string;
      enabledSet: boolean;
      enabled: boolean;
      limitSet: boolean;
      limitValue: bigint;
      owner: string;
      expiresAt: string;
      reason: string;
    }) => platform.setTenantOverride({ companyId: props.companyId, ...input }),
    onSuccess: () => {
      props.onClose();
      props.onDone();
    },
  }));

  const validate = (): string | undefined => {
    if (featureCode().trim() === "") return "功能代碼必填（例：limit.seats、feature.printing）。";
    if (!enabledSet() && !limitSet()) {
      return "至少要指定一個維度：勾「指定啟用」或「指定上限」（兩個都沒有的例外沒有任何效果）。";
    }
    if (owner().trim() === "") {
      return "負責人必填：例外是有人承諾的，沒有承諾者就沒有可追溯的責任。";
    }
    if (limitSet() && !/^\d+$/.test(limitValue().trim())) {
      // 負的上限在判定層等於「任何用量都超額」＝把功能永久關掉；不限額請用「不指定上限」。
      return "上限不得為負：請填不小於 0 的整數（負的上限在判定上等於任何用量都超額；不限額請不要勾「指定上限」）。";
    }
    if (expiresAt().trim() !== "" && Number.isNaN(new Date(expiresAt().trim()).getTime())) {
      return "到期日請填 RFC3339（例：2027-01-01T00:00:00Z），或留空表示不過期（只填日期後端會擋）。";
    }
    return undefined;
  };

  return (
    <WriteForm
      submitLabel="建立例外"
      validate={validate}
      pending={mutation.isPending}
      error={mutation.isError ? describeError(mutation.error) : undefined}
      onSubmit={(reason) =>
        mutation.mutate({
          featureCode: featureCode().trim(),
          enabledSet: enabledSet(),
          // 沒指定啟用維度時送 proto 的零值：不要讓線路上出現一個「沒被指定」的真值。
          enabled: enabledSet() ? enabled() : false,
          limitSet: limitSet(),
          limitValue: limitSet() ? BigInt(limitValue().trim()) : 0n,
          owner: owner().trim(),
          expiresAt: expiresAt().trim(),
          reason,
        })
      }
    >
      <Field>
        <FieldLabel for="override-feature">功能代碼 *</FieldLabel>
        <Input
          id="override-feature"
          value={featureCode()}
          placeholder="limit.seats"
          onInput={(e) => setFeatureCode(e.currentTarget.value)}
        />
      </Field>

      <div class="flex flex-wrap gap-4">
        <LabelledCheckbox label="指定啟用" checked={enabledSet()} onCheckedChange={setEnabledSet} />
        <Show when={enabledSet()}>
          <LabelledCheckbox label="啟用" checked={enabled()} onCheckedChange={setEnabled} />
        </Show>
      </div>

      <LabelledCheckbox label="指定上限" checked={limitSet()} onCheckedChange={setLimitSet} />

      <Show when={limitSet()}>
        <Field>
          <FieldLabel for="override-limit">上限值 *</FieldLabel>
          <Input
            id="override-limit"
            inputmode="numeric"
            value={limitValue()}
            placeholder="30"
            onInput={(e) => setLimitValue(e.currentTarget.value)}
          />
          <FieldDescription>0 是有效上限（等於不能用）；不限額請不要勾「指定上限」。</FieldDescription>
        </Field>
      </Show>

      <Field>
        <FieldLabel for="override-owner">負責人（平台側承諾者）*</FieldLabel>
        <Input
          id="override-owner"
          value={owner()}
          placeholder="ops@example.com"
          onInput={(e) => setOwner(e.currentTarget.value)}
        />
      </Field>

      <Field>
        <FieldLabel for="override-expires">到期日（選填）</FieldLabel>
        <Input
          id="override-expires"
          value={expiresAt()}
          placeholder="2027-01-01T00:00:00Z"
          onInput={(e) => setExpiresAt(e.currentTarget.value)}
        />
        <FieldDescription>留空＝不過期；已過期的例外不列入生效值。</FieldDescription>
      </Field>
    </WriteForm>
  );
}

function SetOverrideDialog(props: {
  companyId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onDone: () => void;
}) {
  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange} lazyMount unmountOnExit>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>新增例外</DialogTitle>
          <DialogDescription>
            例外疊在方案權益之上（方案 ⊕ 例外），寫入後立即改變該租戶的判定；
            撤銷是單向的，事後更正請另立一筆。
          </DialogDescription>
        </DialogHeader>

        <SetOverrideForm
          companyId={props.companyId}
          onClose={() => props.onOpenChange(false)}
          onDone={props.onDone}
        />
      </DialogContent>
    </Dialog>
  );
}

/** 例外的撤銷（RevokeTenantOverride）：只帶 override_id；原因是稽核的必填欄位。 */
function RevokeOverrideDialog(props: {
  override: TenantOverride | undefined;
  onOpenChange: (open: boolean) => void;
  onDone: () => void;
}) {
  const mutation = createMutation(() => ({
    mutationFn: (input: { overrideId: string; reason: string }) => platform.revokeTenantOverride(input),
    onSuccess: () => {
      props.onOpenChange(false);
      props.onDone();
    },
  }));

  return (
    <Dialog
      open={props.override !== undefined}
      onOpenChange={props.onOpenChange}
      lazyMount
      unmountOnExit
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>撤銷例外</DialogTitle>
          <DialogDescription>
            撤銷後該功能回到方案權益（{props.override?.featureCode}）。撤銷是單向的，
            要恢復請另立一筆例外。
          </DialogDescription>
        </DialogHeader>

        <WriteForm
          submitLabel="確認撤銷"
          pending={mutation.isPending}
          error={mutation.isError ? describeError(mutation.error) : undefined}
          onSubmit={(reason) =>
            mutation.mutate({ overrideId: props.override?.id ?? "", reason })
          }
        />
      </DialogContent>
    </Dialog>
  );
}

/**
 * 試用天數的上限：**與後端的 `billing.maxTrialDays` 同值**（365）。開通是不收錢地放行一個方案的
 * 全部權益，無上界的試用等於送出無限期免費 —— 前端先講清楚，後端（權威）也會擋。
 */
const MAX_TRIAL_DAYS = 365;

/** inDaysRFC3339 回「n 天後」的 RFC3339（UTC）；後端的 trial_ends_at 就是這個形狀。 */
function inDaysRFC3339(days: number): string {
  return new Date(Date.now() + days * 24 * 60 * 60 * 1000).toISOString().replace(/\.\d{3}Z$/, "Z");
}

/**
 * canOfferSubscription 判斷是否提供「開通訂閱」按鈕：只在「沒有可服務的合約」時提供。
 * 已經有生效中的合約時後端會回 SYS-2001（00029 的部分唯一索引：
 * 同一家公司只能有一份未取消的合約），按了只是白跑一趟。
 *
 * 未結項 #39：後端若放寬成「同公司多份合約」，這個函式要同步放寬
 * （否則介面比後端保守而不自知）。改這裡時，測試的「生效中合約不提供開通」也要同步改。
 */
export function canOfferSubscription(status: string | undefined): boolean {
  return ["", "none", "cancelled"].includes(status ?? "");
}

/**
 * 開通（CreateSubscription）：建立訂閱與第一期（後端同一個交易）。
 * 沒有期別即 no-op、`SetSeatCount`／`ChangePlan` 都先要一份可服務的合約 —— 少了開通，收款、
 * 期別、催收、凍結全部停擺（只能靠手工 SQL 開合約）。
 *
 * 方案清單用既有的 `ListPlans`（與方案頁同一個 `queryKey`，共用快取、不多打一趟）。
 * 前端驗證不是授權，是**少跑一趟白工**：方案是否 active、該週期有沒有生效價目、金額一律以後端為準。
 */
function CreateSubscriptionForm(props: {
  companyId: string;
  onClose: () => void;
  onDone: () => void;
}) {
  const plans = createQuery(() => ({ queryKey: ["plans"], queryFn: () => platform.listPlans({}) }));
  // 試用的預設天數沿用營運參數 trial_days（platform.settings 的既有 RPC，不新增欄位）。
  const settings = createQuery(() => ({
    queryKey: ["billing-settings"],
    queryFn: () => platform.getBillingSettings({}),
  }));
  const [planCode, setPlanCode] = createSignal("");
  const [billingCycle, setBillingCycle] = createSignal("monthly");
  const [seatCount, setSeatCount] = createSignal("");
  const [trialEndsAt, setTrialEndsAt] = createSignal("");

  // 開通預設帶試用（D2：試用是常態），天數由營運參數決定 —— 但**只在 operator 還沒動過欄位時**填一次
  // （把 operator 清空的欄位又填回去，等於跟他搶方向鍵）。拿不到設定就留空：猜一個天數等於
  // 無聲地送出免費期，而 operator 不會知道那個數字是從哪來的。
  let prefilledTrial = false;
  let trialTouchedByOperator = false;
  createEffect(() => {
    const days = Number(
      settings.data?.settings.find((s) => s.key === "trial_days")?.value ?? Number.NaN,
    );
    if (prefilledTrial || trialTouchedByOperator || !Number.isInteger(days) || days <= 0) return;
    prefilledTrial = true;
    // 夾住上限：trial_days 是可以被改大的營運參數，照抄會預填一個**這個表單自己會拒**的值
    // （operator 一開啟對話框就看到紅字，卻不知道紅字是自己填的）。
    setTrialEndsAt(inDaysRFC3339(Math.min(days, MAX_TRIAL_DAYS)));
  });

  const mutation = createMutation(() => ({
    mutationFn: (input: {
      planCode: string;
      billingCycle: string;
      seatCount: number;
      trialEndsAt: string;
      reason: string;
    }) => platform.createSubscription({ companyId: props.companyId, ...input }),
    onSuccess: () => {
      props.onClose();
      props.onDone();
    },
  }));

  // 只填日期（2027-01-01）在 JS 的 Date 是合法的，但後端用 time.Parse(time.RFC3339) 會擋 ——
  // 前端若用 `new Date()` 判形狀，就會放行一個註定失敗的輸入，所以要照 RFC3339 的形狀判。
  const RFC3339 = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$/;

  const validate = (): string | undefined => {
    if (planCode() === "") {
      return "請選擇方案：合約必須掛在一個賣得動（上架且有該週期價目）的方案上。";
    }
    if (!/^\d+$/.test(seatCount().trim()) || Number(seatCount().trim()) < 1) {
      return "席位數必須是不小於 1 的整數：0 席的訂閱等於停用，那該走取消。";
    }
    const trial = trialEndsAt().trim();
    if (trial === "") return undefined;
    if (!RFC3339.test(trial)) {
      return "試用到期請填 RFC3339（例：2027-01-01T00:00:00Z），或留空表示不試用（只填日期後端會擋）。";
    }
    // 未來的試用才有意義（過去的試用等於一開通就過期），而**超過 MAX_TRIAL_DAYS 天**後端一律拒
    // （試用是不收錢地放行全部權益）。兩者都在這裡先講清楚，不讓 operator 白跑一趟。
    const at = new Date(trial).getTime();
    if (at <= Date.now()) {
      return "試用到期必須是未來時間：過去的試用等於一開通就過期。";
    }
    if (at > Date.now() + MAX_TRIAL_DAYS * 24 * 60 * 60 * 1000) {
      return `試用到期不得超過 ${MAX_TRIAL_DAYS} 天：試用是不收錢地放行全部權益，後端會拒（SYS-1001）。`;
    }
    return undefined;
  };

  return (
    <WriteForm
      submitLabel="建立訂閱"
      validate={validate}
      pending={mutation.isPending}
      error={mutation.isError ? describeError(mutation.error) : undefined}
      onSubmit={(reason) =>
        mutation.mutate({
          planCode: planCode(),
          billingCycle: billingCycle(),
          seatCount: Number(seatCount().trim()),
          trialEndsAt: trialEndsAt().trim(),
          reason,
        })
      }
    >
      <Field>
        <FieldLabel for="create-plan">方案 *</FieldLabel>
        <Select
          id="create-plan"
          value={planCode()}
          onChange={(e) => setPlanCode(e.currentTarget.value)}
        >
          <option value="">請選擇…</option>
          <For each={plans.data?.plans ?? []}>
            {(plan) => (
              <option value={plan.code}>
                {plan.name || plan.code}
                {plan.status === "active" ? "" : `（${plan.status}）`}
              </option>
            )}
          </For>
        </Select>
        <FieldDescription>
          已歸檔的方案不得指派；該週期沒有生效價目時後端會拒絕（不會開出 0 元期別）。
        </FieldDescription>
        {/* 清單載不到時說清楚：空的下拉選單會被讀成「沒有方案可選」，那不是事實。 */}
        <Show when={plans.isError}>
          <p role="alert" class="text-sm font-medium text-destructive">
            {describeError(plans.error)}（沒有方案清單就無法選擇方案）
          </p>
        </Show>
      </Field>

      <Field>
        <FieldLabel for="create-cycle">計費週期 *</FieldLabel>
        <Select
          id="create-cycle"
          value={billingCycle()}
          onChange={(e) => setBillingCycle(e.currentTarget.value)}
        >
          <option value="monthly">月繳</option>
          <option value="yearly">年繳</option>
        </Select>
        <FieldDescription>決定第一期的長度（月 ＋1 月、年 ＋1 年）與取用的價目。</FieldDescription>
      </Field>

      <Field>
        <FieldLabel for="create-seats">席位數 *</FieldLabel>
        <Input
          id="create-seats"
          inputmode="numeric"
          value={seatCount()}
          placeholder="3"
          onInput={(e) => setSeatCount(e.currentTarget.value)}
        />
        <FieldDescription>
          第一期金額＝方案基本價 ＋ 席位數 × 單席價（當期生效價的快照）。日後可用「調整席位」變更。
        </FieldDescription>
      </Field>

      <Field>
        <FieldLabel for="create-trial">試用到期（選填）</FieldLabel>
        <Input
          id="create-trial"
          value={trialEndsAt()}
          placeholder="2026-10-05T00:00:00Z"
          onInput={(e) => {
            trialTouchedByOperator = true;
            setTrialEndsAt(e.currentTarget.value);
          }}
        />
        <FieldDescription>
          預設帶入營運參數的試用天數（trial_days）；留空＝直接生效（active）。填了（必須是未來且不超過
          {MAX_TRIAL_DAYS} 天）→ 狀態為試用中（trialing），到期後排程轉為逾期、由收款帶回 active。
          無論試用與否都會開出第一期。
        </FieldDescription>
      </Field>
    </WriteForm>
  );
}

function CreateSubscriptionDialog(props: {
  companyId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onDone: () => void;
}) {
  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange} lazyMount unmountOnExit>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>開通訂閱</DialogTitle>
          <DialogDescription>
            建立合約與第一期（訂閱、期別、事件與平台稽核在同一個交易）。同一家公司同時只能有一份
            未取消的合約；已取消的合約要再服務時，是再開一筆新合約，不是把舊的復活。
          </DialogDescription>
        </DialogHeader>

        <CreateSubscriptionForm
          companyId={props.companyId}
          onClose={() => props.onOpenChange(false)}
          onDone={props.onDone}
        />
      </DialogContent>
    </Dialog>
  );
}

export default function TenantDetailPage() {
  const params = useParams({ from: "/tenants/$tenantId" });
  const companyId = () => params().tenantId;
  const queryClient = useQueryClient();

  const tenant = createQuery(() => ({
    queryKey: ["tenant", companyId()],
    queryFn: () => platform.getTenant({ companyId: companyId() }),
  }));

  const planCode = () => tenant.data?.tenant?.planCode ?? "";
  const matrix = createQuery(() => ({
    queryKey: ["entitlements", planCode()],
    queryFn: () => platform.getPlanEntitlements({ planCode: planCode() }),
    enabled: planCode() !== "",
  }));

  const [setOpen, setSetOpen] = createSignal(false);
  const [revokeTarget, setRevokeTarget] = createSignal<TenantOverride | undefined>(undefined);
  const [createOpen, setCreateOpen] = createSignal(false);

  // 寫入後讓相關查詢失效：例外改變判定，方案權益決定投影，兩者都要重取，
  // 不靠 operator 自己按重新整理（那樣很容易看著舊值做下一個決定）。
  const refresh = () => {
    void queryClient.invalidateQueries({ queryKey: ["tenant", companyId()] });
    void queryClient.invalidateQueries({ queryKey: ["entitlements", planCode()] });
  };

  // 開通改變的比例外更多：租戶投影（方案／狀態／席位）、整個權益家族（方案權益與價目）與
  // 租戶清單都跟著變 —— 只失效「這一個租戶」會讓下一頁還顯示舊的方案與狀態。
  const refreshAfterCreate = () => {
    refresh();
    void queryClient.invalidateQueries({ queryKey: ["tenants"] });
    void queryClient.invalidateQueries({ queryKey: ["entitlements"] });
    void queryClient.invalidateQueries({ queryKey: ["plans"] });
  };

  const now = new Date();

  return (
    <PageShell title="租戶詳情" description={`公司識別碼 ${companyId()}`}>
      {queryBoundary(tenant, (data) => {
        const summary = data.tenant;

        return (
          <div class="space-y-6">
            <Card>
              <CardHeader class="flex flex-row flex-wrap items-center justify-between gap-2">
                <CardTitle>{summary?.companyName || companyId()}</CardTitle>
                {/* 只在「沒有可服務的合約」時提供開通：已經有生效中的合約時後端會回 SYS-2001
                    （同一家公司只能有一份未取消的合約），按了只是白跑一趟。
                    判定邏輯見 canOfferSubscription（與 00029 的部分唯一索引同義；
                    後端若放寬成「同公司多份合約」，那裡要同步放寬 —— 未結項 #39）。 */}
                <Show when={canOfferSubscription(summary?.subscriptionStatus)}>
                  <Button size="sm" onClick={() => setCreateOpen(true)}>
                    開通訂閱
                  </Button>
                </Show>
              </CardHeader>
              <CardContent class="grid gap-2 text-sm sm:grid-cols-2">
                <p>方案：{summary?.planName || summary?.planCode || "—"}</p>
                <p>
                  訂閱狀態：
                  <SubscriptionBadge status={summary?.subscriptionStatus ?? "none"} />
                </p>
                <p>席位數：{summary?.seatCount ?? 0}</p>
                <p>當期到期日：{summary?.currentPeriodEnd || "—"}</p>
                <p>
                  待收款：{summary?.overdue ? "有未付期別" : "無"}
                </p>
              </CardContent>
            </Card>

            <section class="space-y-3">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <h2 class="text-lg font-semibold">功能例外（override）</h2>
                <Button size="sm" onClick={() => setSetOpen(true)}>
                  新增例外
                </Button>
              </div>

              <Show
                when={data.overrides.length > 0}
                fallback={<EmptyState>沒有例外：此租戶一律依方案權益。</EmptyState>}
              >
                <Table aria-label="功能例外（override）">
                  <TableHeader>
                    <TableRow>
                      <TableHead scope="col">功能</TableHead>
                      <TableHead scope="col">啟用</TableHead>
                      <TableHead scope="col">上限</TableHead>
                      <TableHead scope="col">負責人</TableHead>
                      <TableHead scope="col">到期</TableHead>
                      <TableHead scope="col">原因</TableHead>
                      <TableHead scope="col">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    <For each={data.overrides}>
                      {(o) => (
                        <TableRow>
                          <TableCell>{o.featureCode}</TableCell>
                          <TableCell>{o.enabledSet ? (o.enabled ? "是" : "否") : "未指定"}</TableCell>
                          <TableCell>{o.limitSet ? String(o.limitValue) : "未指定"}</TableCell>
                          <TableCell>{o.owner}</TableCell>
                          <TableCell>
                            {o.expiresAt ? (
                              <Show when={isExpired(o.expiresAt, now)} fallback={o.expiresAt}>
                                <Badge variant="secondary">已過期</Badge>
                              </Show>
                            ) : (
                              "不過期"
                            )}
                          </TableCell>
                          <TableCell>{o.reason}</TableCell>
                          <TableCell>
                            <Button variant="outline" size="sm" onClick={() => setRevokeTarget(o)}>
                              撤銷
                            </Button>
                          </TableCell>
                        </TableRow>
                      )}
                    </For>
                  </TableBody>
                </Table>
              </Show>
            </section>

            <section class="space-y-3">
              <h2 class="text-lg font-semibold">權益現況（投影）</h2>
              <p class="text-sm text-muted-foreground">
                方案權益 ⊕ 例外（已過期的例外不列入生效值）。這是設定值的投影，不是判定結果：
                判定另受訂閱狀態與用量影響（試用中一律開放、非可用狀態一律拒絕、未定義的功能一律拒絕），
                最終以後端為準。
              </p>

              <Show
                when={planCode() !== ""}
                fallback={<EmptyState>此租戶目前沒有方案，沒有可投影的權益。</EmptyState>}
              >
                {queryBoundary(matrix, (m) => {
                  const rows: ProjectedEntitlement[] = projectTenantEntitlements({
                    features: m.features,
                    entitlements: m.entitlements,
                    overrides: data.overrides,
                    now,
                  });

                  return (
                    <div class="space-y-3">
                      <Show when={rows.some((row) => row.ambiguous)}>
                        <p class="text-sm font-medium text-foreground">
                          同一功能有多筆未逾期的例外：例外的新舊順序（created_at）不在 RPC 的回應裡，
                          前端重建不出後端的判定順序 → 生效值僅供參考，最終以後端判定為準。
                          要確認實際結果，請看該租戶端顯示的權益或請平台端查判定。
                        </p>
                      </Show>
                      <Table aria-label="權益現況（投影）">
                      <TableHeader>
                        <TableRow>
                          <TableHead scope="col">功能</TableHead>
                          <TableHead scope="col">方案</TableHead>
                          <TableHead scope="col">例外</TableHead>
                          <TableHead scope="col">生效</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        <For each={rows}>
                          {(row) => (
                            <TableRow>
                              <TableCell>{row.featureCode}</TableCell>
                              <TableCell>
                                {row.planConfigured
                                  ? describeValue(
                                      row.planEnabled,
                                      row.planLimitSet,
                                      row.planLimitValue,
                                    )
                                  : "未設定"}
                              </TableCell>
                              <TableCell>
                                {(row.override
                                  ? [
                                      row.override.enabledSet
                                        ? row.override.enabled
                                          ? "啟用"
                                          : "停用"
                                        : undefined,
                                      row.override.limitSet
                                        ? `上限 ${row.override.limitValue}`
                                        : undefined,
                                    ]
                                      .filter(Boolean)
                                      .join("・")
                                  : "—") +
                                  (row.ambiguous ? `（多筆例外，共 ${row.overrideCount} 筆）` : "")}
                              </TableCell>
                              <TableCell>
                                {row.ambiguous
                                  ? "多筆例外：以後端判定為準"
                                  : row.planConfigured || row.override
                                    ? describeValue(row.enabled, row.limitSet, row.limitValue)
                                    : "未開通"}
                              </TableCell>
                            </TableRow>
                          )}
                        </For>
                        </TableBody>
                        </Table>
                    </div>
                  );
                })}
              </Show>
            </section>
          </div>
        );
      })}

      <CreateSubscriptionDialog
        companyId={companyId()}
        open={createOpen()}
        onOpenChange={setCreateOpen}
        onDone={refreshAfterCreate}
      />
      <SetOverrideDialog
        companyId={companyId()}
        open={setOpen()}
        onOpenChange={setSetOpen}
        onDone={refresh}
      />
      <RevokeOverrideDialog
        override={revokeTarget()}
        onOpenChange={(open) => {
          if (!open) setRevokeTarget(undefined);
        }}
        onDone={refresh}
      />
    </PageShell>
  );
}

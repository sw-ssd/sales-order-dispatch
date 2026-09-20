import { createMutation, createQuery, useQueryClient } from "@tanstack/solid-query";
import { useParams } from "@tanstack/solid-router";
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

/** 例外的新增（SetTenantOverride）：兩個維度都用 `*_set` 表達，未勾＝不覆寫該維度。 */
function SetOverrideDialog(props: {
  companyId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onDone: () => void;
}) {
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
      props.onOpenChange(false);
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
      return "到期日格式須為 RFC3339（例：2027-01-01T00:00:00Z），或留空表示不過期。";
    }
    return undefined;
  };

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
            <LabelledCheckbox
              label="指定啟用"
              checked={enabledSet()}
              onCheckedChange={setEnabledSet}
            />
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

  // 寫入後讓相關查詢失效：例外改變判定，方案權益決定投影，兩者都要重取，
  // 不靠 operator 自己按重新整理（那樣很容易看著舊值做下一個決定）。
  const refresh = () => {
    void queryClient.invalidateQueries({ queryKey: ["tenant", companyId()] });
    void queryClient.invalidateQueries({ queryKey: ["entitlements", planCode()] });
  };

  const now = new Date();

  return (
    <PageShell title="租戶詳情" description={`公司識別碼 ${companyId()}`}>
      {queryBoundary(tenant, (data) => {
        const summary = data.tenant;

        return (
          <div class="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>{summary?.companyName || companyId()}</CardTitle>
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
                                {row.override
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
                                  : "—"}
                              </TableCell>
                              <TableCell>
                                {row.planConfigured || row.override
                                  ? describeValue(row.enabled, row.limitSet, row.limitValue)
                                  : "未開通"}
                              </TableCell>
                            </TableRow>
                          )}
                        </For>
                      </TableBody>
                    </Table>
                  );
                })}
              </Show>
            </section>
          </div>
        );
      })}

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

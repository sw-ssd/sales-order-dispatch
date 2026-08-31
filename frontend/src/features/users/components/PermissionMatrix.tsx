import { Checkbox } from "@ark-ui/solid/checkbox";
import { create } from "@bufbuild/protobuf";
import { For, Show } from "solid-js";
import {
  PermissionSchema,
  type Permission,
} from "~/lib/proto/salesorder/v1/role_pb";
import { cn } from "~/lib/cn";

// 權限矩陣目錄:資源/動作慣例對齊 rbac_policy.csv 與設計書 §3.3。
const RESOURCES = [
  "sales_order",
  "customer",
  "product",
  "user",
  "company",
  "department",
  "role",
  "print",
  "dispatch",
  "accounting",
] as const;

const ACTIONS = [
  "create",
  "read",
  "update",
  "delete",
  "manage",
  "print",
  "dispatch",
  "cancel",
] as const;

const RESOURCE_LABELS: Record<string, string> = {
  sales_order: "訂單",
  customer: "客戶",
  product: "商品",
  user: "使用者",
  company: "公司",
  department: "部門",
  role: "角色",
  print: "列印",
  dispatch: "派車",
  accounting: "會計",
};

const ACTION_LABELS: Record<string, string> = {
  create: "新增",
  read: "檢視",
  update: "編輯",
  delete: "刪除",
  manage: "管理",
  print: "列印",
  dispatch: "派車",
  cancel: "作廢",
};

interface PermissionMatrixProps {
  /** 目前角色全部權限規則(含條件規則;保存時全量取代) */
  permissions: Permission[];
  /** 勾選變更(父層以受控方式持有規則清單) */
  onChange: (permissions: Permission[]) => void;
  /** 是否內建角色(顯示提示;實際可否修改由後端依身分決定) */
  isSystem: boolean;
}

/**
 * 權限矩陣(resource × action 勾選;T19)。
 * 勾選 = 新增/保留該組合規則;取消 = 移除該組合全部規則(含條件規則)。
 * 未觸及的條件規則(如 sales_order×cancel {status:pending})原樣保留。
 */
export function PermissionMatrix(props: PermissionMatrixProps) {
  const rulesFor = (resource: string, action: string) =>
    props.permissions.filter(
      (p) => p.resource === resource && p.action === action,
    );

  const toggle = (resource: string, action: string, checked: boolean) => {
    const others = props.permissions.filter(
      (p) => !(p.resource === resource && p.action === action),
    );
    if (!checked) {
      props.onChange(others);
      return;
    }
    if (rulesFor(resource, action).length > 0) {
      props.onChange(props.permissions); // 已存在(可能帶條件):原樣保留
      return;
    }
    props.onChange([
      ...others,
      create(PermissionSchema, {
        resource,
        action,
        sortOrder: props.permissions.length,
      }),
    ]);
  };

  return (
    <div class="overflow-x-auto rounded-lg border border-gray-200 bg-white shadow-sm">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="sticky left-0 z-10 bg-gray-50 px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
              資源
            </th>
            <For each={ACTIONS}>
              {(action) => (
                <th class="px-3 py-3 text-center text-xs font-medium uppercase tracking-wide text-gray-500">
                  {ACTION_LABELS[action] ?? action}
                </th>
              )}
            </For>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <For each={RESOURCES}>
            {(resource) => (
              <tr class="hover:bg-gray-50">
                <td class="sticky left-0 z-10 bg-white px-4 py-3 text-sm font-medium text-gray-900">
                  {RESOURCE_LABELS[resource] ?? resource}
                </td>
                <For each={ACTIONS}>
                  {(action) => {
                    const checked = rulesFor(resource, action).length > 0;
                    return (
                      <td class="px-3 py-3 text-center">
                        <Checkbox.Root
                          checked={checked}
                          onCheckedChange={(e) =>
                            toggle(resource, action, e.checked === true)
                          }
                          aria-label={`${RESOURCE_LABELS[resource] ?? resource} ${ACTION_LABELS[action] ?? action}`}
                        >
                          <Checkbox.Control
                            class={cn(
                              "inline-flex h-4 w-4 items-center justify-center rounded border",
                              checked
                                ? "border-blue-600 bg-blue-600"
                                : "border-gray-300 bg-white",
                            )}
                          >
                            <Checkbox.Indicator class="text-white">
                              <svg
                                viewBox="0 0 12 12"
                                class="h-3 w-3"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                              >
                                <path d="M2 6l3 3 5-6" />
                              </svg>
                            </Checkbox.Indicator>
                          </Checkbox.Control>
                          <Checkbox.HiddenInput />
                        </Checkbox.Root>
                      </td>
                    );
                  }}
                </For>
              </tr>
            )}
          </For>
        </tbody>
      </table>
      <Show when={props.isSystem}>
        <p class="border-t border-gray-200 px-4 py-2 text-xs text-gray-500">
          內建角色權限為系統預設值;僅 super / developer 可修改。
        </p>
      </Show>
    </div>
  );
}

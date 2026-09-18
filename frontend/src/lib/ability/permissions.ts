// 權限集合模組:以輕量 Set 取代 CASL ability(D32,前端 OpenFGA proxy 驅動)。
// 後端 GetAbility 已改為 OpenFGA 驅動,回傳可用 (resource, action) 集合;
// 前端以 resource:action key 查詢,取代 @casl/ability 的 can().
// 物件狀態條件由 domain 狀態機處理,不屬前端權限集合。
export type Permission = { resource: string; action: string };

export const permKey = (resource: string, action: string) => `${resource}:${action}`;

export function createPermissions(rules: readonly Permission[]): Set<string> {
  return new Set(rules.map((r) => permKey(r.resource, r.action)));
}

// hasPermission 查詢權限集合;未含該 (resource, action) → false(fail-closed)。
export function hasPermission(
  perms: ReadonlySet<string>,
  resource: string,
  action: string,
): boolean {
  return perms.has(permKey(resource, action));
}

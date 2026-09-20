import PlanCard from "./PlanCard";

/**
 * 帳號／方案頁（`/account`）：租戶後台的訂閱資訊**唯讀且不吵**（spec §2.4 規則 2）——
 * 提醒用的 banner 由 shell 負責，這裡放完整、需要時才來看的那一份。
 */
export default function AccountPage() {
  return (
    <main>
      <header class="mb-6 border-b-2 border-border pb-4">
        <h1 class="text-2xl font-bold text-foreground">帳號／方案</h1>
        <p class="mt-1 text-sm text-muted-foreground">
          目前訂閱的方案與用量（唯讀）
        </p>
      </header>

      <PlanCard />
    </main>
  );
}

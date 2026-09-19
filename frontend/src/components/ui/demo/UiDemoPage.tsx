import { createSignal, For, Show, type JSX } from "solid-js";
import {
  Badge,
  Button,
  buttonVariants,
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
  Checkbox,
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Field,
  FieldDescription,
  FieldError,
  FieldLabel,
  Input,
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationFirst,
  PaginationItem,
  PaginationLast,
  PaginationNext,
  PaginationPrevious,
  PaginationSummary,
  ScrollArea,
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarTrigger,
  Spinner,
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
  ThemeSwitcher,
  type SidebarCollapsible,
  type SidebarSide,
  type SidebarVariant,
} from "~/components/ui";
import { Building2, House, Plus, ShieldCheck, Trash2, Users } from "lucide-solid";

/**
 * `/ui` 元件庫展示頁（dev-only）。
 *
 * 路由只在開發環境註冊（見 `~/router`），這裡再以 `import.meta.env.DEV` 把關一次：
 * 正式站即使用手動輸入網址也拿不到展示內容。
 *
 * **深色預覽用 `ThemeSwitcher`**（Task 6 的 `ThemeProvider` 已由 AppShell 提供），
 * 這裡不自行切換 `documentElement` 的 class——那會變成第二套主題機制。
 */

const BUTTON_VARIANTS = [
  "default",
  "destructive",
  "outline",
  "secondary",
  "ghost",
  "link",
  "success",
  "warning",
  "info",
] as const;

const BUTTON_SIZES = ["xs", "sm", "default", "lg", "icon"] as const;

const BADGE_VARIANTS = [
  "default",
  "secondary",
  "destructive",
  "outline",
  "success",
  "warning",
  "info",
] as const;

const PAGINATION_VARIANTS = ["default", "outline", "ghost", "flat"] as const;

const PAGINATION_SIZES = ["sm", "default", "lg", "icon"] as const;

const SIDEBAR_VARIANTS: readonly SidebarVariant[] = ["sidebar", "floating", "inset"];

const SIDEBAR_SIDES: readonly SidebarSide[] = ["left", "right"];

const SIDEBAR_COLLAPSIBLES: readonly SidebarCollapsible[] = ["icon", "offcanvas", "none"];

/** 展示區塊外框。 */
function Section(props: { title: string; hint?: string; children: JSX.Element }) {
  return (
    <section class="rounded-lg border border-border bg-card p-5 text-card-foreground">
      <h2 class="text-lg font-semibold text-foreground">{props.title}</h2>
      <Show when={props.hint}>
        {(hint) => <p class="mt-1 text-sm text-muted-foreground">{hint()}</p>}
      </Show>
      <div class="mt-4 flex flex-col gap-4">{props.children}</div>
    </section>
  );
}

/** 單列展示：左側固定寬度的標籤 + 右側內容。 */
function Row(props: { label: string; children: JSX.Element }) {
  return (
    <div class="flex flex-wrap items-center gap-x-3 gap-y-2">
      <span class="w-20 shrink-0 font-mono text-xs text-muted-foreground">{props.label}</span>
      <div class="flex flex-wrap items-center gap-3">{props.children}</div>
    </div>
  );
}

/** 單選的展示控制：用 `Button` 的分段控制外觀，只為了切換 demo 的 props。 */
function Segmented<T extends string>(props: {
  label: string;
  value: T;
  options: readonly T[];
  onChange: (value: T) => void;
}) {
  return (
    <Row label={props.label}>
      <div class="inline-flex items-center gap-1 rounded-lg border border-border bg-muted p-1">
        <For each={props.options}>
          {(option) => (
            <Button
              size="xs"
              variant={props.value === option ? "default" : "ghost"}
              aria-pressed={props.value === option}
              onClick={() => props.onChange(option)}
            >
              {option}
            </Button>
          )}
        </For>
      </div>
    </Row>
  );
}

/** Checkbox 目前不接受 `children`（標籤文字不落地），可辨識名稱一律走 `aria-label`；這裡補上可見文字。 */
function LabeledCheckbox(props: {
  label: string;
  checked?: boolean;
  defaultChecked?: boolean;
  disabled?: boolean;
  onCheckedChange?: (checked: boolean) => void;
}) {
  return (
    <span class="inline-flex items-center gap-2">
      <Checkbox
        aria-label={props.label}
        checked={props.checked}
        defaultChecked={props.defaultChecked}
        disabled={props.disabled}
        onCheckedChange={props.onCheckedChange}
      />
      <span class="text-sm text-muted-foreground">{props.label}</span>
    </span>
  );
}

/** 對話框展示：`blur`／`showCloseButton` 兩種開關都要能看到差異，因此自帶開關狀態。 */
function DialogDemo(props: { label: string; blur?: boolean; showCloseButton?: boolean }) {
  const [open, setOpen] = createSignal(false);

  return (
    <Show when={import.meta.env.DEV}>
      <Button variant="outline" size="sm" onClick={() => setOpen(true)}>
        {props.label}
      </Button>
      <Dialog open={open()} onOpenChange={setOpen}>
        <DialogContent blur={props.blur} showCloseButton={props.showCloseButton}>
          <DialogHeader>
            <DialogTitle>對話框範例</DialogTitle>
            <DialogDescription>
              Esc、遮罩或關閉鈕都能關；焦點鎖在對話框內，關閉後歸還原處。
            </DialogDescription>
          </DialogHeader>
          <div class="text-sm text-muted-foreground">
            {`blur=${props.blur !== false}、showCloseButton=${props.showCloseButton !== false}`}
          </div>
          <DialogFooter>
            <DialogClose class={buttonVariants({ variant: "outline", size: "sm" })}>
              取消
            </DialogClose>
            <DialogClose class={buttonVariants({ size: "sm" })}>確認</DialogClose>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Show>
  );
}

/** 側邊欄展示：variant／side／collapsible 三個維度都用控制列切換，全部組合都可達到。 */
function SidebarDemo() {
  const [variant, setVariant] = createSignal<SidebarVariant>("sidebar");
  const [side, setSide] = createSignal<SidebarSide>("left");
  const [collapsible, setCollapsible] = createSignal<SidebarCollapsible>("icon");

  return (
    <div class="flex flex-col gap-3">
      <Segmented
        label="variant"
        value={variant()}
        options={SIDEBAR_VARIANTS}
        onChange={setVariant}
      />
      <Segmented label="side" value={side()} options={SIDEBAR_SIDES} onChange={setSide} />
      <Segmented
        label="collapsible"
        value={collapsible()}
        options={SIDEBAR_COLLAPSIBLES}
        onChange={setCollapsible}
      />

      {/*
        demo 自帶 `SidebarProvider`，與 AppShell 的 provider 各自持有開關狀態；但兩者共用
        localStorage `ui:sidebar`（Provider 只認這一個 key），所以在這裡切換會記住。
        `mod+B` 也由兩個 provider 各自註冊，因此會同時切換 shell 與 demo 的側欄。
      */}
      <SidebarProvider>
        <div
          class="flex h-80 overflow-hidden rounded-lg border border-border"
          classList={{ "flex-row-reverse": side() === "right" }}
        >
          <Sidebar
            side={side()}
            variant={variant()}
            collapsible={collapsible()}
            class="h-full"
          >
            <SidebarContent>
              <SidebarGroup>
                <SidebarGroupLabel>營運</SidebarGroupLabel>
                <SidebarGroupContent>
                  <SidebarMenu>
                    <SidebarMenuItem>
                      <SidebarMenuButton isActive tooltip="首頁">
                        <House />
                        <span>首頁</span>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                    <SidebarMenuItem>
                      <SidebarMenuButton tooltip="客戶總表">
                        <Building2 />
                        <span>客戶總表</span>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                    <SidebarMenuItem>
                      <SidebarMenuButton tooltip="部門">
                        <Users />
                        <span>部門</span>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                    <SidebarMenuItem>
                      <SidebarMenuButton tooltip="角色">
                        <ShieldCheck />
                        <span>角色</span>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  </SidebarMenu>
                </SidebarGroupContent>
              </SidebarGroup>
            </SidebarContent>
          </Sidebar>

          <SidebarInset class="h-full">
            <div class="flex items-center gap-3 border-b border-border p-2">
              <SidebarTrigger />
              <span class="font-mono text-xs text-muted-foreground">
                {`variant=${variant()} side=${side()} collapsible=${collapsible()}`}
              </span>
            </div>
            <div class="grow p-4 text-sm text-muted-foreground">
              `SidebarInset`：內容區。`icon` 收合成 3rem 圖示列（標籤轉 tooltip）、`offcanvas` 收合成
              0 寬並 `inert`、`none` 不隨開關改變寬度。
            </div>
          </SidebarInset>
        </div>
      </SidebarProvider>
    </div>
  );
}

export default function UiDemoPage() {
  const [name, setName] = createSignal("");
  const [code, setCode] = createSignal("１２３");
  const [agree, setAgree] = createSignal(true);
  const [reviewed, setReviewed] = createSignal(false);
  const [defaultPage, setDefaultPage] = createSignal(7);
  const [variantPage, setVariantPage] = createSignal(2);
  const [customPage, setCustomPage] = createSignal(12);

  const codeInvalid = () => code().trim().length < 8;

  return (
    <Show
      when={import.meta.env.DEV}
      fallback={
        <main class="p-8">
          <h1 class="text-xl font-semibold text-foreground">元件庫展示</h1>
          <p class="mt-2 text-muted-foreground">此頁僅在開發環境提供。</p>
        </main>
      }
    >
      <main class="min-h-dvh bg-background p-6 text-foreground lg:p-8">
        <div class="mx-auto flex w-full max-w-5xl flex-col gap-6">
          <header class="flex flex-wrap items-center justify-between gap-4">
            <div>
              <h1 class="text-2xl font-bold">元件庫展示</h1>
              <p class="mt-1 text-sm text-muted-foreground">
                僅開發環境（<code>/ui</code>）提供。深色預覽請用右側切換器——
                它與 app 共用同一套 `ThemeProvider` 與 localStorage `ui:theme`。
              </p>
            </div>
            <ThemeSwitcher />
          </header>

          {/* --- Button --- */}
          <Section title="Button" hint="variant × size，另含 loading 與 disabled。">
            <Row label="variant">
              <For each={BUTTON_VARIANTS}>
                {(variant) => <Button variant={variant}>{variant}</Button>}
              </For>
            </Row>
            <Row label="size">
              <For each={BUTTON_SIZES}>
                {(size) => (
                  <Button size={size} aria-label={size === "icon" ? "新增" : undefined}>
                    {size === "icon" ? <Plus /> : size}
                  </Button>
                )}
              </For>
            </Row>
            <Row label="state">
              <Button loading>載入中</Button>
              <Button disabled>停用</Button>
              <Button variant="destructive" size="icon" aria-label="刪除">
                <Trash2 />
              </Button>
            </Row>
          </Section>

          {/* --- Badge --- */}
          <Section title="Badge" hint="七種語意色，皆為軟底或外框，無互動。">
            <Row label="variant">
              <For each={BADGE_VARIANTS}>
                {(variant) => <Badge variant={variant}>{variant}</Badge>}
              </For>
            </Row>
          </Section>

          {/* --- Card --- */}
          <Section title="Card" hint="六個部件：Card / Header / Title / Description / Content / Footer。">
            <Card class="max-w-lg">
              <CardHeader>
                <CardTitle>客戶資料</CardTitle>
                <CardDescription>共 42 筆，最近更新 3 分鐘前</CardDescription>
              </CardHeader>
              <CardContent class="text-sm text-muted-foreground">
                標題與頁尾是 `bg-muted` 色帶；內容區 `grow p-5`。
              </CardContent>
              <CardFooter class="justify-between">
                <span>頁尾色帶</span>
                <Button size="sm">匯出</Button>
              </CardFooter>
            </Card>
          </Section>

          {/* --- Field / Label / Input --- */}
          <Section
            title="Field / FieldLabel / FieldError / FieldDescription / Input"
            hint="錯誤態由呼叫端傳 invalid；Input 會自行接手 Ark Field context 的 aria-invalid／aria-describedby。"
          >
            <div class="grid gap-6 sm:grid-cols-2">
              <Field>
                <FieldLabel for="demo-name">公司名稱</FieldLabel>
                <Input
                  id="demo-name"
                  placeholder="範例股份有限公司"
                  value={name()}
                  onInput={(event) => setName(event.currentTarget.value)}
                />
                <FieldDescription>發票上的正式名稱（一般狀態）。</FieldDescription>
              </Field>

              <Field invalid={codeInvalid()}>
                <FieldLabel for="demo-code">統一編號</FieldLabel>
                <Input
                  id="demo-code"
                  value={code()}
                  onInput={(event) => setCode(event.currentTarget.value)}
                />
                <Show
                  when={codeInvalid()}
                  fallback={<FieldDescription>八碼數字。</FieldDescription>}
                >
                  <FieldError>統一編號需為 8 碼數字。</FieldError>
                </Show>
              </Field>

              <Field>
                <FieldLabel for="demo-required">必填欄位 *</FieldLabel>
                <Input id="demo-required" required />
                <FieldDescription>
                  `required` 下在控件上（Field 包裝層只額外開放 `invalid`）。
                </FieldDescription>
              </Field>

              <Field>
                <FieldLabel for="demo-disabled">停用欄位</FieldLabel>
                <Input id="demo-disabled" disabled value="不可編輯" />
                <FieldDescription>原生 `disabled`。</FieldDescription>
              </Field>

              <Field>
                <FieldLabel for="demo-file">檔案</FieldLabel>
                <Input id="demo-file" type="file" />
                <FieldDescription>`file:` 變體已統一按鈕樣式。</FieldDescription>
              </Field>

              <Field>
                <FieldLabel for="demo-password">密碼</FieldLabel>
                <Input id="demo-password" type="password" placeholder="••••••••" />
                <FieldDescription>`type="password"`。</FieldDescription>
              </Field>
            </div>
          </Section>

          {/* --- Table --- */}
          <Section title="Table" hint="表頭／主體／頁尾／說明；外框與圓角交給 Card。">
            <Card>
              <Table>
                <TableCaption>客戶清單（示範資料）</TableCaption>
                <TableHeader>
                  <TableRow>
                    <TableHead scope="col">名稱</TableHead>
                    <TableHead scope="col">統一編號</TableHead>
                    <TableHead scope="col">狀態</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow>
                    <TableCell>範例股份有限公司</TableCell>
                    <TableCell>12345678</TableCell>
                    <TableCell>
                      <Badge variant="success">啟用</Badge>
                    </TableCell>
                  </TableRow>
                  <TableRow data-state="selected">
                    <TableCell>選取中的列</TableCell>
                    <TableCell>23456789</TableCell>
                    <TableCell>
                      <Badge variant="outline">草稿</Badge>
                    </TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>已停用公司</TableCell>
                    <TableCell>34567890</TableCell>
                    <TableCell>
                      <Badge variant="destructive">停用</Badge>
                    </TableCell>
                  </TableRow>
                </TableBody>
                <TableFooter>
                  <TableRow>
                    <TableCell colspan={3}>共 3 筆</TableCell>
                  </TableRow>
                </TableFooter>
              </Table>
            </Card>
          </Section>

          {/* --- Pagination --- */}
          <Section
            title="Pagination"
            hint="預設版型（含省略號）＋ 直接收 count／page／pageSize／onPageChange；下方為 variant、size 與自組版型。"
          >
            <Row label="預設">
              <Pagination
                class="mx-0 w-auto justify-start"
                count={500}
                page={defaultPage()}
                pageSize={10}
                onPageChange={setDefaultPage}
              />
            </Row>

            <For each={PAGINATION_VARIANTS}>
              {(variant) => (
                <Row label={variant}>
                  <Pagination
                    class="mx-0 w-auto justify-start"
                    count={100}
                    page={variantPage()}
                    pageSize={10}
                    onPageChange={setVariantPage}
                  >
                    <PaginationContent>
                      <li>
                        <PaginationPrevious variant={variant} hideText />
                      </li>
                      <li>
                        <PaginationItem variant={variant} value={1}>
                          1
                        </PaginationItem>
                      </li>
                      <li>
                        <PaginationItem variant={variant} value={2}>
                          2
                        </PaginationItem>
                      </li>
                      <li>
                        <PaginationItem variant={variant} value={3}>
                          3
                        </PaginationItem>
                      </li>
                      <li>
                        <PaginationNext variant={variant} hideText />
                      </li>
                    </PaginationContent>
                  </Pagination>
                </Row>
              )}
            </For>

            <For each={PAGINATION_SIZES}>
              {(size) => (
                <Row label={size}>
                  <Pagination
                    class="mx-0 w-auto justify-start"
                    count={100}
                    page={variantPage()}
                    pageSize={10}
                    onPageChange={setVariantPage}
                  >
                    <PaginationContent>
                      <li>
                        <PaginationPrevious size={size} hideText />
                      </li>
                      <li>
                        <PaginationItem size={size} value={2}>
                          2
                        </PaginationItem>
                      </li>
                      <li>
                        <PaginationItem size={size} value={3}>
                          3
                        </PaginationItem>
                      </li>
                      <li>
                        <PaginationNext size={size} hideText />
                      </li>
                    </PaginationContent>
                  </Pagination>
                </Row>
              )}
            </For>

            <Row label="自組">
              <div class="flex flex-wrap items-center gap-3">
                <PaginationSummary>第 {(customPage() - 1) * 10 + 1}–{customPage() * 10} 筆，共 500 筆</PaginationSummary>
                <Pagination
                  class="mx-0 w-auto justify-start"
                  count={500}
                  page={customPage()}
                  pageSize={10}
                  onPageChange={setCustomPage}
                >
                  <PaginationContent>
                    <li>
                      <PaginationFirst />
                    </li>
                    <li>
                      <PaginationPrevious />
                    </li>
                    <li>
                      <PaginationItem value={1}>1</PaginationItem>
                    </li>
                    <li>
                      <PaginationEllipsis index={1} />
                    </li>
                    <li>
                      <PaginationItem value={customPage()}>{customPage()}</PaginationItem>
                    </li>
                    <li>
                      <PaginationEllipsis index={2} />
                    </li>
                    <li>
                      <PaginationItem value={50}>50</PaginationItem>
                    </li>
                    <li>
                      <PaginationNext />
                    </li>
                    <li>
                      <PaginationLast />
                    </li>
                  </PaginationContent>
                </Pagination>
              </div>
            </Row>
          </Section>

          {/* --- Spinner --- */}
          <Section title="Spinner" hint="role=status；預設 aria-label 為「載入中」，可用 label 覆寫。">
            <Row label="size">
              <Spinner size="sm" label="載入中" />
              <Spinner label="載入中" />
              <Spinner size="lg" label="載入中" />
            </Row>
            <Row label="跟隨顏色">
              <Spinner label="載入中" class="text-current" />
              <Button loading>載入中</Button>
            </Row>
          </Section>

          {/* --- Dialog --- */}
          <Section title="Dialog" hint="三個開關的差異：預設、blur 關閉、右上角關閉鈕關閉。">
            <Row label="open">
              <DialogDemo label="預設" />
              <DialogDemo label="blur=false" blur={false} />
              <DialogDemo label="showCloseButton=false" showCloseButton={false} />
            </Row>
          </Section>

          {/* --- Tabs --- */}
          <Section title="Tabs" hint="horizontal／vertical；未選中的 panel 不在 DOM（lazyMount + unmountOnExit）。">
            <Row label="horizontal">
              <Tabs defaultValue="employee" class="max-w-xl">
                <TabsList>
                  <TabsTrigger value="employee">員工</TabsTrigger>
                  <TabsTrigger value="store">店家</TabsTrigger>
                  <TabsTrigger value="guest">訪客</TabsTrigger>
                </TabsList>
                <TabsContent value="employee">
                  <p class="text-sm text-muted-foreground">員工分頁內容。</p>
                </TabsContent>
                <TabsContent value="store">
                  <p class="text-sm text-muted-foreground">店家分頁內容。</p>
                </TabsContent>
                <TabsContent value="guest">
                  <p class="text-sm text-muted-foreground">訪客分頁內容。</p>
                </TabsContent>
              </Tabs>
            </Row>
            <Row label="vertical">
              <Tabs defaultValue="one" orientation="vertical" class="max-w-xl">
                <TabsList>
                  <TabsTrigger value="one">第一項</TabsTrigger>
                  <TabsTrigger value="two">第二項</TabsTrigger>
                </TabsList>
                <TabsContent value="one">
                  <p class="text-sm text-muted-foreground">方向鍵可切換分頁。</p>
                </TabsContent>
                <TabsContent value="two">
                  <p class="text-sm text-muted-foreground">未選中的 panel 已卸載。</p>
                </TabsContent>
              </Tabs>
            </Row>
          </Section>

          {/* --- Checkbox --- */}
          <Section
            title="Checkbox"
            hint="未勾選／已勾選／停用／受控；可辨識名稱一律由 aria-label 提供（見 checkbox.md 的已知限制）。"
          >
            <Row label="state">
              <LabeledCheckbox label="未勾選" />
              <LabeledCheckbox label="已勾選" defaultChecked />
              <LabeledCheckbox label="停用" disabled />
              <LabeledCheckbox label="停用（已勾選）" defaultChecked disabled />
            </Row>
            <Row label="受控">
              <LabeledCheckbox
                label={`我已閱讀條款（${agree()}）`}
                checked={agree()}
                onCheckedChange={setAgree}
              />
              <LabeledCheckbox
                label={`已覆核（${reviewed()}）`}
                checked={reviewed()}
                onCheckedChange={setReviewed}
              />
            </Row>
          </Section>

          {/* --- ScrollArea --- */}
          <Section title="ScrollArea" hint="高度下在 Root；捲軸由 Ark 依 overflow 自動顯隱。">
            <Row label="vertical">
              <ScrollArea class="max-h-40 w-64 rounded-lg border border-border">
                <div class="space-y-2 p-3 text-sm">
                  <For each={Array.from({ length: 20 }, (_, index) => index + 1)}>
                    {(item) => <p class="text-muted-foreground">第 {item} 列</p>}
                  </For>
                </div>
              </ScrollArea>
            </Row>
            <Row label="horizontal">
              <ScrollArea orientation="horizontal" class="h-20 w-64 rounded-lg border border-border">
                <div class="w-[40rem] p-3 text-sm text-muted-foreground">
                  這一行的寬度是 40rem，因此會出現水平捲軸。
                </div>
              </ScrollArea>
            </Row>
            <Row label="both">
              <ScrollArea orientation="both" class="h-32 w-64 rounded-lg border border-border">
                <div class="w-[40rem] space-y-2 p-3 text-sm text-muted-foreground">
                  <For each={Array.from({ length: 12 }, (_, index) => index + 1)}>
                    {(item) => <p>兩向都能捲：第 {item} 列</p>}
                  </For>
                </div>
              </ScrollArea>
            </Row>
          </Section>

          {/* --- Sidebar --- */}
          <Section
            title="Sidebar"
            hint="variant × side × collapsible 由上方控制列切換；在範例容器內自帶 SidebarProvider，高度固定 20rem。"
          >
            <SidebarDemo />
          </Section>

          {/* --- Theme --- */}
          <Section
            title="Theme / ThemeSwitcher"
            hint="頁首的切換器就是它；深色只換 token，元件內沒有任何顏色相關的 dark: 變體。"
          >
            <Row label="切換器">
              <ThemeSwitcher />
              <span class="text-sm text-muted-foreground">
                三態：light｜system｜dark，持久化在 localStorage `ui:theme`。
              </span>
            </Row>
          </Section>
        </div>
      </main>
    </Show>
  );
}

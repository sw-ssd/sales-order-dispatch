import { Route } from "@solidjs/router";
import type { ParentProps } from "solid-js";

function Home() {
  return (
    <main class="p-8">
      <h1 class="text-2xl font-bold">多公司訂出貨系統</h1>
      <p class="mt-2 text-gray-600">首頁佔位（Wave 1 骨架）</p>
    </main>
  );
}

function Login() {
  return (
    <main class="p-8">
      <h1 class="text-2xl font-bold">登入</h1>
      <p class="mt-2 text-gray-600">登入頁佔位（01-auth 計畫實作）</p>
    </main>
  );
}

export default function App(props: ParentProps) {
  return (
    <>
      <Route path="/" component={Home} />
      <Route path="/login" component={Login} />
      {props.children}
    </>
  );
}

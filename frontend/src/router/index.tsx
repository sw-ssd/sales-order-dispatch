import { Route, Router } from "@solidjs/router";
import App from "~/App";
import LoginPage from "~/features/auth/pages/LoginPage";

function HomePage() {
  return (
    <main class="p-8">
      <h1 class="text-2xl font-bold">多公司訂出貨系統</h1>
      <p class="mt-2 text-gray-600">首頁佔位（Wave 1 骨架）</p>
    </main>
  );
}

export default function AppRouter() {
  return (
    <Router root={App}>
      <Route path="/" component={HomePage} />
      <Route path="/login" component={LoginPage} />
    </Router>
  );
}

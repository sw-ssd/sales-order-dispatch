import { QueryClientProvider } from "@tanstack/solid-query";
import { render } from "solid-js/web";
import "./index.css";
import { queryClient } from "./lib/query";
import { AppRouter } from "./router";

const root = document.getElementById("root");
if (!root) throw new Error("index.html 缺少 #root");

render(
  () => (
    <QueryClientProvider client={queryClient}>
      <AppRouter />
    </QueryClientProvider>
  ),
  root,
);

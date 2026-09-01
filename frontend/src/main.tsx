import { render } from "solid-js/web";
import AppRouter from "~/router";
import "~/index.css";

const root = document.getElementById("root");
if (!root) throw new Error("找不到 #root 掛載點");

render(() => <AppRouter />, root);

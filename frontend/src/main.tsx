import { render } from "solid-js/web";
import { Router } from "@solidjs/router";
import App from "~/App";
import "~/index.css";

const root = document.getElementById("root");
if (!root) throw new Error("找不到 #root 掛載點");

render(() => <Router root={App} />, root);

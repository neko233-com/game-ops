import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
// React Flow 基础样式先加载，本站自定义样式随后覆盖。
import "@xyflow/react/dist/style.css";
import "./styles.css";
import "./flow/flow.css";

const root = document.getElementById("root");

if (!root) {
  throw new Error("找不到 #root 挂载节点");
}

createRoot(root).render(
  <StrictMode>
    <App />
  </StrictMode>,
);

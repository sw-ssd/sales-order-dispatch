import tseslint from "typescript-eslint";
import solid from "eslint-plugin-solid/configs/typescript";

export default tseslint.config(
  // src/lib/proto 與 src/lib/errcode.ts 是產生檔（buf generate／go generate），不檢。
  { ignores: ["dist", "src/lib/proto", "src/lib/errcode.ts"] },
  ...tseslint.configs.recommended,
  {
    files: ["**/*.{ts,tsx}"],
    ...solid,
  },
);

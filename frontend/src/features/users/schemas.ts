import * as v from "valibot";

/**
 * 公司 modal 表單的欄位規則。
 *
 * 只鏡射改寫前的規則：`name` 與 `identifier` 必填——改寫前是 HTML `required`
 * 加上提交前對 trim 後字串的檢查，因此**純空白視為未填**（與登入表單不同，
 * 登入表單只鏡射 `required`，純空白算已填）。`taxId`、`status` 沒有規則
 * （`status` 是選單、永遠有值；`taxId` 選填）。
 * 刻意不新增統一編號／識別碼的格式驗證，格式規則只有後端知道。
 */
export const companySchema = v.object({
  name: v.pipe(v.string(), v.trim(), v.nonEmpty("請輸入公司名稱")),
  identifier: v.pipe(v.string(), v.trim(), v.nonEmpty("請輸入識別碼(identifier)")),
  taxId: v.string(),
  status: v.string(),
});

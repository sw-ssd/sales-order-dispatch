// 產生 Go casl 套件的 golden fixture。用法:pnpm --filter frontend exec node scripts/casl-golden-gen.mjs
import { createMongoAbility, subject as wrapSubject } from "@casl/ability";
import { rulesToQuery } from "@casl/ability/extra";
import { writeFileSync } from "node:fs";

const U = (id) => Object.fromEntries(
  Object.entries({ user_id: id.user, company_id: id.company, department_id: id.department, customer_id: id.customer })
    .filter(([, v]) => v !== undefined),
);

// 與 Go 端相同的佔位符展開(產生 fixture 時展開,讓 JS 端見到具體值)
function expand(rules, id) {
  const map = {
    "${user.id}": id.user_id, "${user.company_id}": id.company_id,
    "${user.department_id}": id.department_id, "${user.customer_id}": id.customer_id,
  };
  return rules.map((r) => {
    if (!r.conditions) return { action: r.action, subject: r.subject, ...(r.inverted ? { inverted: true } : {}) };
    const conds = {};
    for (const [f, v] of Object.entries(r.conditions)) {
      conds[f] = typeof v === "string" && map[v] !== undefined ? map[v] : v;
    }
    return { action: r.action, subject: r.subject, conditions: conds, ...(r.inverted ? { inverted: true } : {}) };
  });
}

const staff = { user: "u1", company: "c1", department: "d1" };

const cases = [
  {
    name: "無條件 + cannot 狀態排除",
    identity: U(staff),
    rules: [
      { action: "read", subject: "sales_order" },
      { action: "read", subject: "sales_order", inverted: true, conditions: { status: "voided" } },
    ],
    checks: [
      { action: "read", subject: "sales_order", instance: { status: "pending" }, expect: true },
      { action: "read", subject: "sales_order", instance: { status: "voided" }, expect: false },
    ],
  },
  {
    name: "佔位符 + $in",
    identity: U(staff),
    rules: [
      { action: "read", subject: "sales_order", conditions: { department_id: "${user.department_id}" } },
      { action: "cancel", subject: "sales_order", conditions: { status: { $in: ["pending"] } } },
    ],
    checks: [
      { action: "read", subject: "sales_order", instance: { department_id: "d1" }, expect: true },
      { action: "read", subject: "sales_order", instance: { department_id: "d2" }, expect: false },
      { action: "cancel", subject: "sales_order", instance: { status: "pending" }, expect: true },
      { action: "cancel", subject: "sales_order", instance: { status: "processing" }, expect: false },
      { action: "cancel", subject: "sales_order", expect: false }, // type-level:帶條件規則不命中
    ],
  },
  {
    name: "manage all 與無規則 deny",
    identity: U(staff),
    rules: [{ action: "manage", subject: "all" }],
    checks: [
      { action: "delete", subject: "customer", instance: { x: 1 }, expect: true },
    ],
  },
  {
    name: "多允許規則 OR 聚合(query 形狀)",
    identity: U(staff),
    rules: [
      { action: "read", subject: "sales_order", conditions: { department_id: "${user.department_id}" } },
      { action: "read", subject: "sales_order", conditions: { created_by: "${user.id}" } },
      { action: "read", subject: "sales_order", inverted: true, conditions: { status: "voided" } },
    ],
    checks: [
      { action: "read", subject: "sales_order", instance: { department_id: "d1" }, expect: true },
      { action: "read", subject: "sales_order", instance: { department_id: "d2", created_by: "u1" }, expect: true },
      { action: "read", subject: "sales_order", instance: { department_id: "d1", status: "voided" }, expect: false },
      { action: "read", subject: "customer", instance: {}, expect: false }, // 無規則 subject → query denied
    ],
  },
];

for (const c of cases) {
  const expanded = expand(c.rules, JSON.parse(JSON.stringify(c.identity)));
  const ability = createMongoAbility(expanded);
  for (const chk of c.checks) {
    const inst = chk.instance ? wrapSubject(chk.subject, { ...chk.instance }) : undefined;
    const got = ability.can(chk.action, inst ?? chk.subject);
    // type-level(無 instance)檢查:真 CASL 忽略 conditions 回 true,但 Go 端契約為
    // fail-closed(帶條件規則不命中,見 Task 3 規格) — 僅此類檢查容許差異,不 throw。
    if (chk.instance === undefined && got !== chk.expect) {
      chk.note = "type-level divergence: Go fail-closed vs CASL ignore-conditions";
    } else if (got !== chk.expect) {
      throw new Error(`${c.name}: can(${chk.action},${chk.subject}) = ${got}, expect ${chk.expect}`);
    }
    const q = rulesToQuery(ability, chk.action, chk.subject, (rule) =>
      rule.inverted ? { not: rule.conditions ?? {} } : (rule.conditions ?? {}),
    );
    chk.query = q === null
      ? { denied: true, or: 0, and: 0 }
      : { denied: false, or: (q.$or ?? []).length, and: (q.$and ?? []).length };
  }
}

const out = new URL("../../backend/internal/authz/casl/testdata/cases.json", import.meta.url);
writeFileSync(out, JSON.stringify({ cases }, null, 2) + "\n");
console.log(`wrote ${out.pathname}`);

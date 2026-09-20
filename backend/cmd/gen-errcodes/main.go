// Command gen-errcodes 把 internal/errcode 的 registry 投影成碼表與三端常數。
//
// 執行：`go generate ./internal/errcode`（見 internal/errcode/generate.go）；產物必須入 commit，
// CI 以 `git diff --exit-code` 驗證與 registry 同步（.github/workflows/ci.yml）。
//
// 碼的常數名取自碼宣告處的 Go 常數名（`var CustomerCodeExists = MustRegister(…, id: "CUST-2001")`）：
// 找不到對應宣告即失敗——寧可產生器紅，也不要前端拿到一個沒有名字的碼。
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
)

const notice = "由 `go generate ./internal/errcode` 產生（來源：backend/internal/errcode/codes_*.go），請勿手改。"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gen-errcodes:", err)
		os.Exit(1)
	}
}

func run() error {
	root, err := repoRoot() // monorepo 根（backend 的上一層）
	if err != nil {
		return err
	}
	names, err := constNames(filepath.Join(root, "backend", "internal", "errcode"))
	if err != nil {
		return err
	}

	codes := errcode.All() // 已依 ID 排序：產生檔的 git diff 穩定性靠它
	rows := make([]row, 0, len(codes))
	for _, c := range codes {
		name, ok := names[c.ID()]
		if !ok {
			return fmt.Errorf("碼 %s 找不到對應的 Go 常數宣告（產生器以 `var X = MustRegister(...id: %q...)` 的形狀定位）", c.ID(), c.ID())
		}
		rows = append(rows, row{code: c, name: name})
	}

	md := renderMarkdown(rows)
	ts := renderTS(rows)
	dart := renderDart(rows)
	outputs := []output{
		{rel: "docs/error-codes.md", body: md},
		{rel: "frontend/src/lib/errcode.ts", body: ts},
		{rel: "app/lib/gen/errcode.dart", body: dart},
		// Plan C 的產物：目錄還不存在就先跳過（不為了這步去建 platform-console 骨架）。
		{rel: "platform-console/src/lib/errcode.ts", body: ts, optional: true},
	}

	for _, out := range outputs {
		abs := filepath.Join(root, filepath.FromSlash(out.rel))
		if _, err := os.Stat(filepath.Dir(abs)); os.IsNotExist(err) {
			if !out.optional {
				return fmt.Errorf("輸出目錄不存在: %s", filepath.Dir(abs))
			}
			fmt.Printf("gen-errcodes: 跳過 %s（目錄不存在；Plan C 落地後即自動納入）\n", out.rel)
			continue
		}
		changed, err := writeIfChanged(abs, out.body)
		if err != nil {
			return err
		}
		fmt.Printf("gen-errcodes: %s %s\n", map[bool]string{true: "已寫入", false: "未變更"}[changed], out.rel)
	}
	return nil
}

type row struct {
	code errcode.Code
	name string
}

type output struct {
	rel      string // 相對 monorepo 根，以 `/` 分隔
	body     string
	optional bool // 目錄不存在時跳過而非失敗
}

// repoRoot 由工作目錄向上找 go.mod（= backend/），其上一層即 monorepo 根。
//
// 不假設 cwd 深度：`go generate` 在**套件目錄**（backend/internal/errcode）執行，
// 手動 `go run ./cmd/gen-errcodes` 則在 backend/——兩種 cwd 都得產生同一組檔案。
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for start := dir; ; {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Dir(dir), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("由 %s 向上找不到 go.mod（請在 backend/ 或其子目錄執行）", start)
		}
		dir = parent
	}
}

// constNames 回傳「碼 ID → Go 常數名」，掃描碼宣告處的 var 區塊。
func constNames(dir string) (map[string]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	names := map[string]string{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil, err
		}
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}
			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, val := range vs.Values {
					if i >= len(vs.Names) {
						continue
					}
					if id := registeredID(val); id != "" {
						names[id] = vs.Names[i].Name
					}
				}
			}
		}
	}
	return names, nil
}

// registeredID 由 `MustRegister(Code{id: "X", …})` 取出 id 字面值；形狀不符回 ""。
func registeredID(v ast.Expr) string {
	call, ok := v.(*ast.CallExpr)
	if !ok || len(call.Args) == 0 {
		return ""
	}
	lit, ok := call.Args[0].(*ast.CompositeLit)
	if !ok {
		return ""
	}
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "id" {
			continue
		}
		val, ok := kv.Value.(*ast.BasicLit)
		if !ok || val.Kind != token.STRING {
			continue
		}
		return strings.Trim(val.Value, `"`)
	}
	return ""
}

func renderMarkdown(rows []row) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "<!-- %s -->\n\n# 錯誤碼一覽\n\n", notice)
	fmt.Fprintf(&b, "**已遷移路徑**的對外錯誤一律帶 `ErrorInfo`（`code` 與 `trace_id`）；**尚未遷移者**列於\n")
	fmt.Fprintf(&b, "`backend/internal/services/errcode_baseline.txt`（基線只減不增，由守門測試強制）。\n")
	fmt.Fprintf(&b, "**唯一真相來源**是 `backend/internal/errcode`，本檔是它的投影——要改碼表請改 `codes_*.go` 後重跑 `go generate ./internal/errcode`。\n\n")
	fmt.Fprintf(&b, "- 碼的形態為 `域-4位數`；對外 connect 碼由區段決定（見 `sectionRules`）。\n")
	fmt.Fprintf(&b, "- **碼發佈後不得重用或改義**，廢止只標狀態；訊息中的 `{param}` 由 `ErrorInfo.details` 帶入（前端顯示前請自行填入）。\n")
	fmt.Fprintf(&b, "- 共 %d 碼（%s）。\n\n", len(rows), domainSummary(rows))
	fmt.Fprintf(&b, "| 碼 | 域 | Go 常數 | connect 碼 | 訊息 | 參數 | 狀態 |\n|---|---|---|---|---|---|---|\n")
	for _, r := range rows {
		params := "—"
		if p := placeholders(r.code.Message()); len(p) > 0 {
			params = "`" + strings.Join(p, "`, `") + "`"
		}
		state := "使用中"
		if r.code.IsDeprecated() {
			state = "已廢止"
		}
		fmt.Fprintf(&b, "| %s | %s | `%s` | %s | %s | %s | %s |\n",
			r.code.ID(), r.code.Domain(), r.name, r.code.ConnectCode().String(), r.code.Message(), params, state)
	}
	return b.String()
}

func renderTS(rows []row) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "// %s\n", notice)
	fmt.Fprintf(&b, "// 碼的唯一真相來源是後端 registry；本檔只是投影，CI 會驗證同步（.github/workflows/ci.yml）。\n\n")
	fmt.Fprintf(&b, "/** 對外錯誤碼（`ErrorInfo.code`；形如 \"域-4位數\"）。 */\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "export const %s = \"%s\";\n", errConstName(r.name), r.code.ID())
	}
	fmt.Fprintf(&b, "\n/** 碼 → 繁中訊息樣板（`{param}` 由 `ErrorInfo.details` 帶入）。 */\n")
	fmt.Fprintf(&b, "export const CODE_MESSAGES: Record<string, string> = {\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "  \"%s\": \"%s\",\n", r.code.ID(), r.code.Message())
	}
	fmt.Fprintf(&b, "};\n")
	return b.String()
}

func renderDart(rows []row) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "// %s\n", notice)
	fmt.Fprintf(&b, "// 碼的唯一真相來源是後端 registry；本檔只是投影，CI 會驗證同步（.github/workflows/ci.yml）。\n\n")
	fmt.Fprintf(&b, "/// 對外錯誤碼（`ErrorInfo.code`；形如 '域-4位數'）。\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "const String %s = '%s';\n", errDartName(r.name), r.code.ID())
	}
	fmt.Fprintf(&b, "\n/// 碼 → 繁中訊息樣板（`{param}` 由 `ErrorInfo.details` 帶入）。\n")
	fmt.Fprintf(&b, "const Map<String, String> errCodeMessages = {\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "  '%s': '%s',\n", r.code.ID(), r.code.Message())
	}
	fmt.Fprintf(&b, "};\n")
	return b.String()
}

var placeholderRe = regexp.MustCompile(`\{(\w+)\}`)

// placeholders 依出現順序回傳訊息樣板中的參數名（去重）。
func placeholders(message string) []string {
	var out []string
	for _, m := range placeholderRe.FindAllStringSubmatch(message, -1) {
		if !slices.Contains(out, m[1]) {
			out = append(out, m[1])
		}
	}
	return out
}

// domainSummary 例："AUTH 7／CUST 3／PLAT 4／SYS 7"。
func domainSummary(rows []row) string {
	counts := map[string]int{}
	for _, r := range rows {
		counts[string(r.code.Domain())]++
	}
	domains := make([]string, 0, len(counts))
	for d := range counts {
		domains = append(domains, d)
	}
	sort.Strings(domains)
	parts := make([]string, 0, len(domains))
	for _, d := range domains {
		parts = append(parts, fmt.Sprintf("%s %d", d, counts[d]))
	}
	return strings.Join(parts, "／")
}

// errConstName 把 Go 常數名轉成前端常數名：CustomerCodeExists → ERR_CUSTOMER_CODE_EXISTS。
//
// 縮寫也切開（SysRLSViolation → ERR_SYS_RLS_VIOLATION，而非 ERR_SYS_R_L_S_VIOLATION）：
// 本碼庫的詞彙有不少全大寫縮寫（RLS／FGA／OIDC）。
func errConstName(goName string) string {
	runes := []rune(goName)
	var b bytes.Buffer
	b.WriteString("ERR_")
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) &&
			(unicode.IsLower(runes[i-1]) || (i+1 < len(runes) && unicode.IsLower(runes[i+1]))) {
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToUpper(r))
	}
	return b.String()
}

// errDartName：CustomerCodeExists → errCustomerCodeExists（Go 常數名本身即 camelCase，
// 加 prefix 後仍是合法 lowerCamelCase）。
func errDartName(goName string) string {
	return "err" + goName
}

// writeIfChanged 只在內容不同時寫檔（不必每次 generate 都動 mtime／觸發 watcher）。
func writeIfChanged(path, body string) (bool, error) {
	if old, err := os.ReadFile(path); err == nil && string(old) == body {
		return false, nil
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return false, err
	}
	return true, nil
}

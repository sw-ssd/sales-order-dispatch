package services

// 錯誤碼守門：新增的 connect.NewError( 必須帶註冊碼。
//
// 背景：碼的唯一真相來源是 internal/errcode（常數即註冊）。但「registry 存在」不等於「新錯誤會用它」，
// 本測試把兩者綁起來：internal/** 內每個 connect.NewError( 呼叫點都必須落在 errcode_baseline.txt
// （T4／T5 之前的歷史遺留，逐批遷移中）或改用 errcode.<Code>.Error(…)／Wrap(…)。
//
// 基線行的形態為 "<相對模組根路徑>:<歸屬名>:<筆數>"，例：
//
//	internal/services/shared_service.go:trimNonEmpty:1
//	internal/services/company_service.go:CompanyService.DeleteCompany:2
//
// 三個設計要點：
//   - **不含行號**：行號會因無關編輯全數失效，讓守門測試每次都被迫更新基線，最後淪為無腦照抄。
//   - **含筆數**：只比對「這個函式有沒有出現過」會讓「同一函式新增第二個未註冊呼叫」靜默通過
//     （基線中大量函式本來就各有一筆）——帶筆數後，多一筆、少一筆都會紅。
//   - **歸屬名涵蓋套件層級宣告**：整個檔案的 AST 都掃（不只看函式 body），
//     `var x = connect.NewError(…)`、`init()`、別名 import（`c "connectrpc.com/connect"`）
//     的呼叫點同樣納管；套件層級且無具名宣告者歸到 `<package-level>`。
//
// 基線只能縮小（廣義）：筆數增加 → 紅；筆數減少或呼叫點消失 → 也紅（否則基線會腐化成永久豁免清單，
// 且看不出哪些歷史遺留已經遷移掉了）。有意縮小時跑：
//
//	go test ./internal/services/ -run TestNoUnregisteredErrorConstruction -update-errcode-baseline
//
// 不掃的檔案：
//   - `_test.go`：測試自造錯誤不是對外錯誤路徑。
//   - `internal/errcode/`：碼的定義處本來就得呼叫 connect.NewError。
//   - 產生檔（`// Code generated … DO NOT EDIT.`，如 internal/proto/**/*.connect.go）：
//     機器輸出，proto 重生成就會冒出新的 CodeUnimplemented 建構點，那不是開發者手寫的錯誤路徑。

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const (
	errcodeBaselineFile = "errcode_baseline.txt"
	connectPkgPath      = "connectrpc.com/connect"
	packageLevelScope   = "<package-level>"
)

var updateErrcodeBaseline = flag.Bool("update-errcode-baseline", false,
	"重寫 errcode_baseline.txt 為目前掃描結果（僅在有意縮小基線時使用，並在 PR 說明）")

func TestNoUnregisteredErrorConstruction(t *testing.T) {
	sites := scanConnectNewError(t)

	if *updateErrcodeBaseline {
		writeBaseline(t, sites)
		t.Logf("已更新 %s：%d 個呼叫點位置", errcodeBaselineFile, len(sites))
		return
	}

	raw, err := os.ReadFile(errcodeBaselineFile)
	if os.IsNotExist(err) {
		// 基線不存在時建檔後仍失敗：逼人檢視內容再入 commit，不靜默建檔放行。
		writeBaseline(t, sites)
		t.Fatalf("%s 不存在，已依目前樹產生 %d 個呼叫點位置；請檢視後入 commit（CI 出現此訊息表示基線被刪了）",
			errcodeBaselineFile, len(sites))
	}
	if err != nil {
		t.Fatalf("讀取 %s 失敗: %v", errcodeBaselineFile, err)
	}
	baseline, err := parseBaseline(string(raw))
	if err != nil {
		t.Fatalf("%s 格式錯誤（應為 path:歸屬名:筆數）: %v", errcodeBaselineFile, err)
	}

	total := 0
	for _, n := range sites {
		total += n
	}
	baselineTotal := 0
	for _, n := range baseline {
		baselineTotal += n
	}
	t.Logf("掃描到 %d 個呼叫點（%d 處），基線 %d 個（%d 處）", total, len(sites), baselineTotal, len(baseline))

	var added, increased, decreased, stale []string
	for key, n := range sites {
		old, ok := baseline[key]
		switch {
		case !ok:
			added = append(added, fmt.Sprintf("  %s（%d 筆）", key, n))
		case n > old:
			increased = append(increased, fmt.Sprintf("  %s（基線 %d 筆 → 現 %d 筆）", key, old, n))
		case n < old:
			decreased = append(decreased, fmt.Sprintf("  %s（基線 %d 筆 → 現 %d 筆）", key, old, n))
		}
	}
	for key, n := range baseline {
		if _, ok := sites[key]; !ok {
			stale = append(stale, fmt.Sprintf("  %s（基線 %d 筆）", key, n))
		}
	}
	for _, list := range [][]string{added, increased, decreased, stale} {
		sort.Strings(list)
	}

	var sections []string
	if len(added) > 0 {
		sections = append(sections, "新增的錯誤建構未使用註冊碼（請改用 errcode.<Code>.Error(…)）:\n"+
			strings.Join(added, "\n"))
	}
	if len(increased) > 0 {
		sections = append(sections, "既有位置新增了未註冊的錯誤建構（基線只能縮小）:\n"+
			strings.Join(increased, "\n"))
	}
	if len(decreased) > 0 {
		sections = append(sections, "未註冊的錯誤建構減少了（已遷移的請用 -update-errcode-baseline 縮小基線）:\n"+
			strings.Join(decreased, "\n"))
	}
	if len(stale) > 0 {
		sections = append(sections, "基線含已不存在的呼叫點（請刪除以下行，基線只能縮小）:\n"+
			strings.Join(stale, "\n"))
	}
	if len(sections) > 0 {
		t.Fatalf("錯誤碼基線守門失敗:\n\n%s", strings.Join(sections, "\n\n"))
	}
}

// writeBaseline 以 "path:歸屬名:筆數" 排序後覆寫基線。
func writeBaseline(t *testing.T, sites map[string]int) {
	t.Helper()
	lines := make([]string, 0, len(sites))
	for key, n := range sites {
		lines = append(lines, key+":"+strconv.Itoa(n))
	}
	sort.Strings(lines)
	content := ""
	if len(lines) > 0 {
		content = strings.Join(lines, "\n") + "\n"
	}
	if err := os.WriteFile(errcodeBaselineFile, []byte(content), 0o644); err != nil {
		t.Fatalf("寫入 %s 失敗: %v", errcodeBaselineFile, err)
	}
}

// parseBaseline 解析 "<path>:<歸屬名>:<筆數>"（以**最後一個冒號**切筆數：歸屬名不含冒號，路徑也不會）。
func parseBaseline(raw string) (map[string]int, error) {
	baseline := map[string]int{}
	for i, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		sep := strings.LastIndex(line, ":")
		if sep <= 0 {
			return nil, fmt.Errorf("第 %d 行缺少筆數: %q", i+1, line)
		}
		n, err := strconv.Atoi(line[sep+1:])
		if err != nil || n < 1 {
			return nil, fmt.Errorf("第 %d 行的筆數無效: %q", i+1, line)
		}
		baseline[line[:sep]] = n
	}
	return baseline, nil
}

// scanConnectNewError 回傳 internal/** 內所有 connect.NewError( 呼叫點的鍵與筆數。
func scanConnectNewError(t *testing.T) map[string]int {
	t.Helper()
	root := moduleRoot(t)
	fset := token.NewFileSet()
	sites := map[string]int{}

	err := filepath.WalkDir(filepath.Join(root, "internal"), func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, filePath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel == "internal/errcode" {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(filePath, ".go") || strings.HasSuffix(filePath, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		if ast.IsGenerated(f) {
			return nil
		}
		names, dotImport := connectImportNames(f)
		if len(names) == 0 && !dotImport {
			return nil
		}
		// 走**整個檔案**（不只是函式 body）：套件層級的 var 也能有呼叫點。
		collectSites(f, rel, "", names, dotImport, sites)
		return nil
	})
	if err != nil {
		t.Fatalf("掃描 internal/** 失敗: %v", err)
	}
	return sites
}

// collectSites 走訪節點並累計呼叫點；scope 為最近的具名宣告（函式名，或套件層級的 "var <名>"）。
//
// scope 為空字串＝還在套件層級：那裡的 var 宣告改用變數名當歸屬（否則 `var x = connect.NewError(…)`
// 會沒有鍵）；函式內的區域變數則維持歸屬該函式（鍵不因局部變數改名而漂移）。
func collectSites(node ast.Node, rel, scope string, names map[string]bool, dotImport bool, sites map[string]int) {
	ast.Inspect(node, func(n ast.Node) bool {
		switch d := n.(type) {
		case *ast.FuncDecl:
			if d.Body == nil {
				return false
			}
			collectSites(d.Body, rel, funcName(d), names, dotImport, sites)
			return false
		case *ast.ValueSpec:
			if scope != "" {
				return true
			}
			for i, v := range d.Values {
				collectSites(v, rel, "var "+nameAt(d.Names, i), names, dotImport, sites)
			}
			return false
		case *ast.CallExpr:
			if isConnectNewError(d, names, dotImport) {
				key := scope
				if key == "" {
					key = packageLevelScope
				}
				sites[rel+":"+key]++
			}
		}
		return true
	})
}

// connectImportNames 回傳檔案中指向 connectrpc.com/connect 的區域名稱（含別名）。
// dotImport＝`import . "connectrpc.com/connect"`（此時 NewError 是裸識別字）。
func connectImportNames(f *ast.File) (names map[string]bool, dotImport bool) {
	names = map[string]bool{}
	for _, imp := range f.Imports {
		p, err := strconv.Unquote(imp.Path.Value)
		if err != nil || p != connectPkgPath {
			continue
		}
		switch {
		case imp.Name == nil:
			names[path.Base(connectPkgPath)] = true
		case imp.Name.Name == ".":
			dotImport = true
		case imp.Name.Name == "_":
			// 只為副作用引入，不可能有呼叫點
		default:
			names[imp.Name.Name] = true
		}
	}
	return names, dotImport
}

func isConnectNewError(call *ast.CallExpr, names map[string]bool, dotImport bool) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		id, ok := call.Fun.(*ast.Ident)
		return ok && dotImport && id.Name == "NewError"
	}
	if sel.Sel.Name != "NewError" {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && names[pkg.Name]
}

// funcName 為 "<接收者型別>.<函式名>"（無接收者時只有函式名）；指標與值接收者視為同一鍵。
func funcName(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return fd.Name.Name
	}
	recv := fd.Recv.List[0].Type
	if star, ok := recv.(*ast.StarExpr); ok {
		recv = star.X
	}
	if idx, ok := recv.(*ast.IndexExpr); ok { // 泛型接收者
		recv = idx.X
	}
	if id, ok := recv.(*ast.Ident); ok {
		return id.Name + "." + fd.Name.Name
	}
	return fd.Name.Name
}

// nameAt 取 ValueSpec 第 i 個值對應的變數名（`var a, b = x, y`）；數量對不上或無名時用最後一個／<unnamed>。
func nameAt(names []*ast.Ident, i int) string {
	if len(names) == 0 {
		return "<unnamed>"
	}
	if i < len(names) {
		return names[i].Name
	}
	return names[len(names)-1].Name
}

// moduleRoot 由測試的工作目錄向上找 go.mod。
//
// 不硬寫 "../.."：硬寫的上層相對路徑換個 cwd 就指向別的目錄，而掃錯目錄的失敗長相是「全綠」——最壞的
// 那種失敗（守門測試永遠通過）。
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("取工作目錄失敗: %v", err)
	}
	for start := dir; ; {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("由 %s 向上找不到 go.mod", start)
		}
		dir = parent
	}
}

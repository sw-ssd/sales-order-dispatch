package services

// 錯誤碼守門：新增的 connect.NewError( 必須帶註冊碼。
//
// 背景：碼的唯一真相來源是 internal/errcode（常數即註冊）。但「registry 存在」不等於「新錯誤會用它」，
// 本測試把兩者綁起來：internal/** 內每個 connect.NewError( 呼叫點都必須落在 errcode_baseline.txt
// （T4／T5 之前的歷史遺留，逐批遷移中）或改用 errcode.<Code>.Error(…)／Wrap(…)。
//
// 鍵為 "<相對模組根路徑>:<函式名>"（方法含接收者型別）——**刻意不含行號**：行號會因無關編輯而全數失效，
// 讓守門測試每次都逼人更新基線，最後淪為無腦照抄（守門失效）。
//
// 基線只能縮小：掃描結果中不在基線者 → 紅；基線中已不存在的呼叫點 → 也紅
// （否則基線會腐化成永久豁免清單）。
//
// 不掃的檔案：
//   - `_test.go`：測試自造錯誤不是對外錯誤路徑。
//   - `internal/errcode/`：碼的定義處本來就得呼叫 connect.NewError。
//   - 產生檔（`// Code generated … DO NOT EDIT.`，如 internal/proto/**/*.connect.go）：
//     機器輸出，proto 重生成就會冒出新的 CodeUnimplemented 建構點，那不是開發者手寫的錯誤路徑。
//
// 更新基線（僅在有意縮小基線、且 PR 說明時使用）：
//
//	go test ./internal/services/ -run TestNoUnregisteredErrorConstruction -update-errcode-baseline

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
)

const errcodeBaselineFile = "errcode_baseline.txt"

var updateErrcodeBaseline = flag.Bool("update-errcode-baseline", false,
	"重寫 errcode_baseline.txt 為目前掃描結果（僅在有意縮小基線時使用，並在 PR 說明）")

func TestNoUnregisteredErrorConstruction(t *testing.T) {
	occurrences := scanConnectNewError(t)

	if *updateErrcodeBaseline {
		writeBaseline(t, occurrences)
		t.Logf("已更新 %s：%d 筆", errcodeBaselineFile, len(occurrences))
		return
	}

	raw, err := os.ReadFile(errcodeBaselineFile)
	if os.IsNotExist(err) {
		// 基線不存在時建檔後仍失敗：逼人檢視內容再入 commit，不靜默建檔放行。
		writeBaseline(t, occurrences)
		t.Fatalf("%s 不存在，已依目前樹產生 %d 筆；請檢視後入 commit（CI 出現此訊息表示基線被刪了）",
			errcodeBaselineFile, len(occurrences))
	}
	if err != nil {
		t.Fatalf("讀取 %s 失敗: %v", errcodeBaselineFile, err)
	}

	var baseline []string
	for _, line := range strings.Split(string(raw), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			baseline = append(baseline, line)
		}
	}
	t.Logf("掃描到 %d 個 connect.NewError( 呼叫點，基線 %d 筆", len(occurrences), len(baseline))

	inBaseline := make(map[string]bool, len(baseline))
	for _, b := range baseline {
		inBaseline[b] = true
	}
	var added, stale []string
	for _, occ := range occurrences {
		if !inBaseline[occ] {
			added = append(added, occ)
		}
	}
	for _, b := range baseline {
		if !slices.Contains(occurrences, b) {
			stale = append(stale, b)
		}
	}
	if len(added) > 0 {
		t.Fatalf("新增的錯誤建構未使用註冊碼（請改用 errcode.<Code>.Error(…)）:\n%s",
			strings.Join(added, "\n"))
	}
	if len(stale) > 0 {
		t.Fatalf("基線含已不存在的呼叫點（請刪除以下行，基線只能縮小）:\n%s", strings.Join(stale, "\n"))
	}
}

func writeBaseline(t *testing.T, occurrences []string) {
	t.Helper()
	content := ""
	if len(occurrences) > 0 {
		content = strings.Join(occurrences, "\n") + "\n"
	}
	if err := os.WriteFile(errcodeBaselineFile, []byte(content), 0o644); err != nil {
		t.Fatalf("寫入 %s 失敗: %v", errcodeBaselineFile, err)
	}
}

// scanConnectNewError 回傳 internal/** 內所有 connect.NewError( 呼叫點的鍵（排序、去重）。
func scanConnectNewError(t *testing.T) []string {
	t.Helper()
	root := moduleRoot(t)
	fset := token.NewFileSet()
	sites := map[string]bool{}

	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
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
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		if ast.IsGenerated(f) {
			return nil
		}
		for _, site := range connectNewErrorSitesIn(f, rel) {
			sites[site] = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("掃描 internal/** 失敗: %v", err)
	}

	out := make([]string, 0, len(sites))
	for site := range sites {
		out = append(out, site)
	}
	sort.Strings(out)
	return out
}

// connectNewErrorSitesIn 回傳單一檔案內的呼叫點鍵（同一函式多處呼叫只算一筆）。
func connectNewErrorSitesIn(f *ast.File, rel string) []string {
	var sites []string
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}
		key := rel + ":" + funcName(fd)
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			if isConnectNewError(n) {
				sites = append(sites, key)
			}
			return true
		})
	}
	return sites
}

func isConnectNewError(n ast.Node) bool {
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "NewError" {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "connect"
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

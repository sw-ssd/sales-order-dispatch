package errcode

// 產生器在 cmd/gen-errcodes：讀 registry（All() 與存取子）輸出
//   - ../docs/error-codes.md：客服與文件用的碼表（域／碼／常數／connect 碼／訊息／參數）
//   - ../frontend/src/lib/errcode.ts、../app/lib/gen/errcode.dart：三端常數
//   - ../platform-console/src/lib/errcode.ts：目錄存在（Plan C 落地）才寫
//
// 產生檔的檔頭都標明「由 go generate 產生，請勿手改」與來源；產物必須入 commit，
// CI 以 `git diff --exit-code` 驗證與 registry 同步（.github/workflows/ci.yml）。
//
// 為何不手寫這三份：碼的真相只有 registry 一份，手抄就會漂移，而漂移的症狀是
// 「前端顯示的訊息與後端不同」——沒有測試會抓到。
//
//go:generate go run ../../cmd/gen-errcodes

# sharescan v0.1.0 リリースチェックリスト

## ビルド確認

- [ ] `go build -o sharescan.exe ./cmd/sharescan` がエラーなく完了する
- [ ] `go test ./...` が全てパスする
- [ ] 生成された `sharescan.exe` のサイズが適切である（数MB以内）
- [ ] Linux/Mac からのクロスコンパイル（`GOOS=windows GOARCH=amd64`）が成功する

## CLI基本動作確認

- [ ] `sharescan.exe --version` で `sharescan v0.1.0` が表示される
- [ ] `sharescan.exe --help` で使用方法が日本語で表示される
- [ ] 引数なし実行でエラーメッセージと使用方法が表示される（exit code 1）
- [ ] 存在しないフォルダを指定した場合にわかりやすいエラーメッセージが表示される

## 設定ファイル確認

- [ ] `config.sample.json` が同梱されている
- [ ] `LICENSE`（自作部分のMIT License）が配布物（exe/ZIP）に同梱されている
- [ ] ビルドに使用したGoのライセンスが、自作の `LICENSE` とは別ファイルとして配布物（exe/ZIP）に同梱されている
  - Goの標準ライブラリ・ランタイムはBSD系ライセンス（Copyright The Go Authors）で、バイナリ再配布時も著作権表示・条件・免責文の添付が必要
  - 同梱するファイルは、ビルドに使用したGoのインストール先（`go env GOROOT` で確認）にある `LICENSE`。例: `LICENSE`（自作MIT）と `LICENSE-Go`（Goのライセンス）として並べて同梱し、自作MITに混ぜない
  - 参考: <https://go.dev/LICENSE>
- [ ] `sharescan.exe --config config.sample.json` が実行できる（targets修正後）
- [ ] 設定ファイルが存在しない場合にエラーメッセージが表示される
- [ ] 不正なJSONを指定した場合にエラーメッセージが表示される

## 小規模フォルダでの実行確認

- [ ] テスト用小規模フォルダを走査できる
- [ ] `output/` ディレクトリが生成される
- [ ] `scan_log.txt` が生成される
- [ ] 走査完了後にサマリーが標準出力に表示される

## 出力ファイル確認

- [ ] `report.html` が生成される
- [ ] `summary.csv` が生成される
- [ ] `large_files.csv` が生成される
- [ ] `duplicate_files.csv` が生成される
- [ ] `long_paths.csv` が生成される
- [ ] `old_files.csv` が生成される
- [ ] `extension_summary.csv` が生成される
- [ ] `scan_errors.csv` が生成される

## HTMLレポート確認

- [ ] `report.html` をブラウザで開ける
- [ ] 9つのセクション（概要〜注意事項）が表示される
- [ ] 数値がカンマ区切りで表示される
- [ ] 容量がKB/MB/GBで表示される
- [ ] 注意事項が記載されている（削除・移動不実施の旨など）
- [ ] 外部CSS・CDN・JavaScriptを読み込んでいない（オフラインで表示できる）
- [ ] 日本語フォントで表示される

## CSV確認（Excel）

- [ ] 各CSVをExcelで開いたとき文字化けしない（UTF-8 BOM）
- [ ] ヘッダー行がある
- [ ] カンマや改行を含むパスが正しくエスケープされている
- [ ] サイズが読みやすい単位で表示されている

## 各検出機能確認

- [ ] 大容量ファイルが large_files.csv に出力される
- [ ] 同一内容ファイルが duplicate_files.csv に重複候補として出力される
- [ ] 長パスが long_paths.csv に出力される
- [ ] 古いファイルが old_files.csv に出力される
- [ ] 拡張子別集計が extension_summary.csv に出力される

## エラー処理確認

- [ ] アクセス拒否があるフォルダが含まれても処理が継続する
- [ ] アクセス拒否のエラーが scan_errors.csv に記録される
- [ ] アクセス拒否のエラーが scan_log.txt に記録される
- [ ] エラーがあっても report.html が生成される

## 安全性確認

- [ ] ソースコードにファイル削除（`os.Remove`、`os.RemoveAll`）の呼び出しがない
- [ ] ソースコードにファイル移動（`os.Rename` による移動）の呼び出しがない（出力ファイル作成除く）
- [ ] ソースコードにアクセス権変更（`os.Chmod` 等）の呼び出しがない
- [ ] 走査対象フォルダへの書き込みが発生しない

## 再実行確認

- [ ] `output/` が既に存在する状態で再実行しても正常に完了する
- [ ] 再実行でCSVとHTMLが上書きされる

## ドキュメント確認

- [ ] `README.md` が存在し、実行方法・注意事項が記載されている
- [ ] `docs/tool_design.md` が存在する
- [ ] `docs/test_scenarios.md` が存在する
- [ ] `docs/operation_handover.md` が存在する
- [ ] `config.sample.json` が存在する
- [ ] `LICENSE`（MIT License）が存在し、`README.md` のライセンス節（ソースとビルド済みバイナリの区別を含む）と整合している

# sharescan

共有フォルダ棚卸ツール v0.1.0

## ツール概要

`sharescan` は、自治体の閉域ネットワーク内にある共有フォルダの棚卸・可視化を目的とした **読み取り専用** のCLIツールです。

指定したフォルダ配下を再帰的に走査し、以下を集計・出力します。

- 大容量ファイルの一覧
- 重複ファイル候補の一覧
- 長パス候補の一覧
- 古いファイル候補の一覧
- 拡張子別集計
- 走査エラーの記録

結果は **HTMLレポート**（ブラウザで閲覧可能）と **CSV**（Excel対応）として出力します。

> **重要**: 本ツールはファイルの削除・移動・リネーム・権限変更を一切行いません。

---

## 想定利用者

- 自治体の情報システム担当者
- ファイルサーバの管理・移行・整理を担当する職員
- 共有フォルダの利用状況を把握したい管理者

## 利用場面

- ファイルサーバの移行前調査
- 定期的な共有フォルダ棚卸
- 大容量・重複ファイルの把握
- 古いファイルの整理検討（削除判断は人間が行う）

---

## 実行方法

### 対象フォルダを直接指定

```bash
sharescan.exe "\\server\share"
```

### 設定ファイルを使用

```bash
sharescan.exe --config config.json
```

### ヘルプ表示

```bash
sharescan.exe --help
```

### バージョン表示

```bash
sharescan.exe --version
```

---

## config.json の説明

`config.sample.json` をコピーして編集してください。

| 項目 | 型 | 既定値 | 説明 |
|------|-----|--------|------|
| `targets` | 文字列配列 | （必須） | 走査対象フォルダのパス（UNCパス可） |
| `outputDir` | 文字列 | `"output"` | 出力先ディレクトリ |
| `largeFileThresholdMB` | 数値 | `100` | 大容量ファイルとみなすサイズ（MB） |
| `longPathThreshold` | 数値 | `240` | 長パスとみなす文字数 |
| `oldFileYears` | 数値 | `5` | 古いファイルとみなす経過年数 |
| `enableHashDuplicateCheck` | 真偽値 | `true` | SHA-256ハッシュで重複を判定する |
| `maxRowsInHtml` | 数値 | `100` | HTMLレポートの各表の最大表示件数 |
| `excludeDirectories` | 文字列配列 | `[]` | 走査から除外するディレクトリ名 |
| `excludeExtensions` | 文字列配列 | `[]` | 走査から除外する拡張子（例: `.tmp`） |

設定例：

```json
{
  "targets": ["\\\\server\\share1", "\\\\server\\share2"],
  "outputDir": "output",
  "largeFileThresholdMB": 100,
  "longPathThreshold": 240,
  "oldFileYears": 5,
  "enableHashDuplicateCheck": true,
  "maxRowsInHtml": 100,
  "excludeDirectories": [".git", "tmp"],
  "excludeExtensions": [".tmp", ".bak"]
}
```

---

## 出力ファイル一覧

実行後、`outputDir`（既定: `output/`）に以下のファイルが生成されます。

| ファイル名 | 内容 |
|-----------|------|
| `report.html` | 総合HTMLレポート（ブラウザで閲覧） |
| `summary.csv` | 概要サマリー |
| `large_files.csv` | 大容量ファイル一覧 |
| `duplicate_files.csv` | 重複ファイル候補一覧 |
| `long_paths.csv` | 長パス候補一覧 |
| `old_files.csv` | 古いファイル候補一覧 |
| `extension_summary.csv` | 拡張子別集計 |
| `scan_errors.csv` | 走査エラー一覧 |
| `scan_log.txt` | 実行ログ |

---

## 注意事項

1. **読み取り専用**: 本ツールはファイルの削除・移動・リネーム・権限変更を行いません。
2. **重複候補**: ハッシュ一致はあくまで候補です。削除判断は業務所管部署が行ってください。
3. **情報管理**: レポートにはファイル名・パスが含まれます。個人情報・機密情報が含まれる可能性があるため、取扱いに注意してください。
4. **エラー継続**: アクセス拒否等のエラーが発生しても処理を継続し、エラーを記録します。
5. **ハッシュ計算**: `enableHashDuplicateCheck: true` の場合、同一サイズのファイルのみハッシュ計算します。全ファイルに対してハッシュ計算は行いません。

---

## ビルド方法

Go 1.21 以上が必要です。外部ライブラリは使用していません。

### Windows向けビルド（Windows上で実行）

```bash
go build -o sharescan.exe ./cmd/sharescan
```

### クロスコンパイル（Linux/Mac から Windows向けビルド）

```bash
GOOS=windows GOARCH=amd64 go build -o sharescan.exe ./cmd/sharescan
```

### テスト実行

```bash
go test ./...
```

---

## ライセンス

MIT License（または各組織のポリシーに従って適用してください）

---

*sharescan v0.1.0 — 自治体向け共有フォルダ棚卸ツール*

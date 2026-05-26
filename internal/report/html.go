package report

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/yantkys/sharescan/internal/config"
	"github.com/yantkys/sharescan/internal/model"
	"github.com/yantkys/sharescan/internal/util"
)

const htmlTmpl = `<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>共有フォルダ棚卸レポート</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: "Meiryo", "MS PGothic", sans-serif; font-size: 13px; color: #222; background: #f5f5f5; }
.container { max-width: 1200px; margin: 0 auto; padding: 20px; }
h1 { font-size: 20px; font-weight: bold; margin-bottom: 4px; color: #1a3a5c; }
.subtitle { font-size: 12px; color: #555; margin-bottom: 20px; }
h2 { font-size: 15px; font-weight: bold; color: #1a3a5c; background: #e0eaf4; padding: 6px 10px; margin: 24px 0 8px 0; border-left: 4px solid #2a6099; }
table { width: 100%; border-collapse: collapse; background: #fff; margin-bottom: 8px; }
th { background: #2a6099; color: #fff; padding: 6px 8px; text-align: left; font-size: 12px; white-space: nowrap; }
td { padding: 5px 8px; border-bottom: 1px solid #ddd; vertical-align: top; word-break: break-all; }
tr:nth-child(even) td { background: #f0f5fb; }
tr:hover td { background: #ddeeff; }
.num { text-align: right; white-space: nowrap; }
.truncated { color: #c00; font-size: 11px; margin-top: 4px; }
.summary-table td:first-child { font-weight: bold; width: 200px; white-space: nowrap; }
.note-box { background: #fff8e1; border: 1px solid #f0c040; padding: 12px 16px; margin-top: 8px; }
.note-box p { margin-bottom: 6px; line-height: 1.7; }
.note-box p:last-child { margin-bottom: 0; }
.footer { margin-top: 30px; font-size: 11px; color: #888; text-align: right; }
.tag-warn { color: #c00; font-weight: bold; }
</style>
</head>
<body>
<div class="container">
<h1>共有フォルダ棚卸レポート</h1>
<div class="subtitle">生成日時: {{.ScannedAtStr}} &nbsp;|&nbsp; sharescan v0.1.0</div>

<!-- 1. 概要 -->
<h2>1. 概要 (Summary)</h2>
<table class="summary-table">
<tr><td>走査日時</td><td>{{.ScannedAtStr}}</td></tr>
<tr><td>対象フォルダ数</td><td>{{.TargetCount}}</td></tr>
<tr><td>総ファイル数</td><td class="num">{{.TotalFilesStr}}</td></tr>
<tr><td>総フォルダ数</td><td class="num">{{.TotalDirsStr}}</td></tr>
<tr><td>総容量</td><td class="num">{{.TotalSizeStr}}</td></tr>
<tr><td>大容量ファイル数</td><td class="num">{{.LargeFilesCountStr}}</td></tr>
<tr><td>重複候補グループ数</td><td class="num">{{.DupGroupsCountStr}}</td></tr>
<tr><td>長パス候補数</td><td class="num">{{.LongPathsCountStr}}</td></tr>
<tr><td>古いファイル候補数</td><td class="num">{{.OldFilesCountStr}}</td></tr>
<tr><td>走査エラー数</td><td class="num {{if gt .Result.ErrorCount 0}}tag-warn{{end}}">{{.ErrorCountStr}}</td></tr>
</table>

<!-- 2. 対象フォルダ -->
<h2>2. 対象フォルダ (Target Folders)</h2>
<table>
<tr><th>#</th><th>フォルダパス</th></tr>
{{range $i, $t := .Result.Targets}}<tr><td class="num">{{add $i 1}}</td><td>{{$t}}</td></tr>
{{end}}
</table>

<!-- 3. 大容量ファイル -->
<h2>3. 容量上位ファイル・大容量ファイル (Large Files ≥ {{.LargeThresholdStr}})</h2>
{{if .LargeFiles}}
<table>
<tr><th>#</th><th>ファイル名</th><th>フルパス</th><th class="num">サイズ</th><th>最終更新日時</th></tr>
{{range $i, $f := .LargeFiles}}<tr>
<td class="num">{{add $i 1}}</td>
<td>{{basename $f.Path}}</td>
<td>{{$f.Path}}</td>
<td class="num">{{fmtsize $f.Size}}</td>
<td>{{fmttime $f.ModTime}}</td>
</tr>{{end}}
</table>
{{if .LargeFilesTruncated}}<div class="truncated">※ 表示を {{.MaxRows}} 件に制限しています。全件はCSVファイルを参照してください。</div>{{end}}
{{else}}<p>該当なし</p>{{end}}

<!-- 4. 重複ファイル候補 -->
<h2>4. 重複ファイル候補 (Duplicate Candidates)</h2>
{{if .DuplicateGroups}}
<table>
<tr><th>グループ</th><th>ファイル数</th><th class="num">ファイルサイズ</th><th class="num">グループ合計</th><th>ハッシュ値</th><th>ファイルパス</th></tr>
{{range .DuplicateGroups}}{{$g := .}}<tr>
<td class="num">{{$g.GroupNo}}</td>
<td class="num">{{$g.FileCount}}</td>
<td class="num">{{fmtsize $g.FileSize}}</td>
<td class="num">{{fmtsize $g.TotalSize}}</td>
<td style="font-size:11px;word-break:break-all;">{{$g.Hash}}</td>
<td>{{joinpaths $g.Paths}}</td>
</tr>{{end}}
</table>
{{if .DupGroupsTruncated}}<div class="truncated">※ 表示を {{.MaxRows}} 件に制限しています。全件はCSVファイルを参照してください。</div>{{end}}
{{else}}<p>該当なし</p>{{end}}

<!-- 5. 拡張子別集計 -->
<h2>5. 拡張子別集計 (Extension Summary)</h2>
{{if .ExtSummary}}
<table>
<tr><th>拡張子</th><th class="num">ファイル数</th><th class="num">合計容量</th></tr>
{{range .ExtSummary}}<tr>
<td>{{.Ext}}</td>
<td class="num">{{fmtnum .FileCount}}</td>
<td class="num">{{fmtsize .TotalSize}}</td>
</tr>{{end}}
</table>
{{if .ExtSummaryTruncated}}<div class="truncated">※ 表示を {{.MaxRows}} 件に制限しています。全件はCSVファイルを参照してください。</div>{{end}}
{{else}}<p>該当なし</p>{{end}}

<!-- 6. 長パス候補 -->
<h2>6. 長パス候補 (Long Path Candidates ≥ {{.LongPathThreshold}} 文字)</h2>
{{if .LongPaths}}
<table>
<tr><th>#</th><th class="num">文字数</th><th>フルパス</th><th class="num">サイズ</th><th>最終更新日時</th></tr>
{{range $i, $f := .LongPaths}}<tr>
<td class="num">{{add $i 1}}</td>
<td class="num">{{pathlen $f.Path}}</td>
<td>{{$f.Path}}</td>
<td class="num">{{fmtsize $f.Size}}</td>
<td>{{fmttime $f.ModTime}}</td>
</tr>{{end}}
</table>
{{if .LongPathsTruncated}}<div class="truncated">※ 表示を {{.MaxRows}} 件に制限しています。全件はCSVファイルを参照してください。</div>{{end}}
{{else}}<p>該当なし</p>{{end}}

<!-- 7. 古いファイル候補 -->
<h2>7. 古いファイル候補 (Old Files — 最終更新 {{.OldFileYears}} 年以上前)</h2>
{{if .OldFiles}}
<table>
<tr><th>#</th><th>ファイル名</th><th>フルパス</th><th class="num">サイズ</th><th>最終更新日時</th></tr>
{{range $i, $f := .OldFiles}}<tr>
<td class="num">{{add $i 1}}</td>
<td>{{basename $f.Path}}</td>
<td>{{$f.Path}}</td>
<td class="num">{{fmtsize $f.Size}}</td>
<td>{{fmttime $f.ModTime}}</td>
</tr>{{end}}
</table>
{{if .OldFilesTruncated}}<div class="truncated">※ 表示を {{.MaxRows}} 件に制限しています。全件はCSVファイルを参照してください。</div>{{end}}
{{else}}<p>該当なし</p>{{end}}

<!-- 8. 走査エラー -->
<h2>8. 走査エラー (Scan Errors)</h2>
{{if .ScanErrors}}
<table>
<tr><th>#</th><th>パス</th><th>エラー内容</th></tr>
{{range $i, $e := .ScanErrors}}<tr>
<td class="num">{{add $i 1}}</td>
<td>{{$e.Path}}</td>
<td>{{$e.Error}}</td>
</tr>{{end}}
</table>
{{if .ScanErrorsTruncated}}<div class="truncated">※ 表示を {{.MaxRows}} 件に制限しています。全件はCSVファイルを参照してください。</div>{{end}}
{{else}}<p>走査エラーなし</p>{{end}}

<!-- 9. 注意事項 -->
<h2>9. 注意事項 (Notes)</h2>
<div class="note-box">
<p>このレポートは共有フォルダの棚卸を目的としたものです。</p>
<p>本ツールはファイルの削除・移動・リネーム・権限変更を行いません。</p>
<p>重複ファイルは機械的に抽出した候補であり、削除可否は業務所管部署で確認してください。</p>
<p>ファイル名やパスに個人情報・機密情報が含まれる可能性があるため、レポートの取扱いに注意してください。</p>
</div>

<div class="footer">sharescan v0.1.0 &mdash; 生成: {{.ScannedAtStr}}</div>
</div>
</body>
</html>
`

// templateData holds all values passed to the HTML template.
type templateData struct {
	Result              model.ScanResult
	ScannedAtStr        string
	TargetCount         int
	TotalFilesStr       string
	TotalDirsStr        string
	TotalSizeStr        string
	LargeFilesCountStr  string
	DupGroupsCountStr   string
	LongPathsCountStr   string
	OldFilesCountStr    string
	ErrorCountStr       string
	LargeThresholdStr   string
	LongPathThreshold   int
	OldFileYears        int
	MaxRows             int
	LargeFiles          []model.FileInfo
	DuplicateGroups     []model.DuplicateGroup
	ExtSummary          []model.ExtStats
	LongPaths           []model.FileInfo
	OldFiles            []model.FileInfo
	ScanErrors          []model.ScanError
	LargeFilesTruncated bool
	DupGroupsTruncated  bool
	ExtSummaryTruncated bool
	LongPathsTruncated  bool
	OldFilesTruncated   bool
	ScanErrorsTruncated bool
}

// GenerateHTML writes an HTML report for result to outputPath.
func GenerateHTML(result model.ScanResult, cfg config.Config, outputPath string) error {
	funcs := template.FuncMap{
		"add":       func(a, b int) int { return a + b },
		"basename":  filepath.Base,
		"fmtsize":   util.FormatSize,
		"fmtnum":    func(n int64) string { return util.FormatNumber(n) },
		"fmttime":   func(t interface{ Format(string) string }) string { return t.Format("2006-01-02 15:04:05") },
		"pathlen":   func(s string) int { return len(s) },
		"joinpaths": func(paths []string) template.HTML { return template.HTML(strings.Join(paths, "<br>")) },
	}

	tmpl, err := template.New("report").Funcs(funcs).Parse(htmlTmpl)
	if err != nil {
		return fmt.Errorf("テンプレート解析エラー: %w", err)
	}

	maxRows := cfg.MaxRowsInHtml
	if maxRows <= 0 {
		maxRows = 100
	}

	largeFiles, largeTrunc := limitFileInfos(result.LargeFiles, maxRows)
	dupGroups, dupTrunc := limitDupGroups(result.DuplicateGroups, maxRows)
	extSummary, extTrunc := limitExtStats(result.ExtSummary, maxRows)
	longPaths, longTrunc := limitFileInfos(result.LongPaths, maxRows)
	oldFiles, oldTrunc := limitFileInfos(result.OldFiles, maxRows)
	scanErrors, errTrunc := limitScanErrors(result.ScanErrors, maxRows)

	data := templateData{
		Result:              result,
		ScannedAtStr:        result.ScannedAt.Format("2006-01-02 15:04:05"),
		TargetCount:         len(result.Targets),
		TotalFilesStr:       util.FormatNumber(result.TotalFiles),
		TotalDirsStr:        util.FormatNumber(result.TotalDirs),
		TotalSizeStr:        util.FormatSize(result.TotalSize),
		LargeFilesCountStr:  util.FormatNumber(int64(len(result.LargeFiles))),
		DupGroupsCountStr:   util.FormatNumber(int64(len(result.DuplicateGroups))),
		LongPathsCountStr:   util.FormatNumber(int64(len(result.LongPaths))),
		OldFilesCountStr:    util.FormatNumber(int64(len(result.OldFiles))),
		ErrorCountStr:       util.FormatNumber(result.ErrorCount),
		LargeThresholdStr:   fmt.Sprintf("%d MB", cfg.LargeFileThresholdMB),
		LongPathThreshold:   cfg.LongPathThreshold,
		OldFileYears:        cfg.OldFileYears,
		MaxRows:             maxRows,
		LargeFiles:          largeFiles,
		DuplicateGroups:     dupGroups,
		ExtSummary:          extSummary,
		LongPaths:           longPaths,
		OldFiles:            oldFiles,
		ScanErrors:          scanErrors,
		LargeFilesTruncated: largeTrunc,
		DupGroupsTruncated:  dupTrunc,
		ExtSummaryTruncated: extTrunc,
		LongPathsTruncated:  longTrunc,
		OldFilesTruncated:   oldTrunc,
		ScanErrorsTruncated: errTrunc,
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("レポートファイルの作成に失敗しました: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("テンプレート出力エラー: %w", err)
	}
	return nil
}

func limitFileInfos(items []model.FileInfo, max int) ([]model.FileInfo, bool) {
	if len(items) > max {
		return items[:max], true
	}
	return items, false
}

func limitDupGroups(items []model.DuplicateGroup, max int) ([]model.DuplicateGroup, bool) {
	if len(items) > max {
		return items[:max], true
	}
	return items, false
}

func limitExtStats(items []model.ExtStats, max int) ([]model.ExtStats, bool) {
	if len(items) > max {
		return items[:max], true
	}
	return items, false
}

func limitScanErrors(items []model.ScanError, max int) ([]model.ScanError, bool) {
	if len(items) > max {
		return items[:max], true
	}
	return items, false
}

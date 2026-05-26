package csvout

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yantkys/sharescan/internal/model"
	"github.com/yantkys/sharescan/internal/util"
)

const bom = "\xEF\xBB\xBF"

// WriteAll generates all CSV output files into outputDir.
func WriteAll(result model.ScanResult, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("出力ディレクトリの作成に失敗: %w", err)
	}

	type task struct {
		name string
		fn   func(string) error
	}
	tasks := []task{
		{"summary.csv", func(p string) error { return writeSummary(p, result) }},
		{"large_files.csv", func(p string) error { return writeLargeFiles(p, result) }},
		{"duplicate_files.csv", func(p string) error { return writeDuplicates(p, result) }},
		{"long_paths.csv", func(p string) error { return writeLongPaths(p, result) }},
		{"old_files.csv", func(p string) error { return writeOldFiles(p, result) }},
		{"extension_summary.csv", func(p string) error { return writeExtSummary(p, result) }},
		{"scan_errors.csv", func(p string) error { return writeScanErrors(p, result) }},
	}

	var errs []string
	for _, t := range tasks {
		if err := t.fn(filepath.Join(outputDir, t.name)); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", t.name, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("CSV出力エラー: %s", strings.Join(errs, "; "))
	}
	return nil
}

func createCSV(path string) (*os.File, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	_, err = f.WriteString(bom)
	if err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

func writeRow(f *os.File, fields []string) {
	escaped := make([]string, len(fields))
	for i, v := range fields {
		escaped[i] = util.EscapeCSV(v)
	}
	f.WriteString(strings.Join(escaped, ",") + "\r\n")
}

func writeSummary(path string, r model.ScanResult) error {
	f, err := createCSV(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writeRow(f, []string{"項目", "値"})
	writeRow(f, []string{"スキャン日時", r.ScannedAt.Format("2006-01-02 15:04:05")})
	writeRow(f, []string{"対象フォルダ", strings.Join(r.Targets, " / ")})
	writeRow(f, []string{"総ファイル数", util.FormatNumber(r.TotalFiles)})
	writeRow(f, []string{"総フォルダ数", util.FormatNumber(r.TotalDirs)})
	writeRow(f, []string{"総容量", util.FormatSize(r.TotalSize)})
	writeRow(f, []string{"総容量(バイト)", fmt.Sprintf("%d", r.TotalSize)})
	writeRow(f, []string{"走査エラー件数", util.FormatNumber(r.ErrorCount)})
	return nil
}

func writeLargeFiles(path string, r model.ScanResult) error {
	f, err := createCSV(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writeRow(f, []string{"ファイル名", "フルパス", "サイズ(バイト)", "サイズ", "最終更新日時"})
	for _, fi := range r.LargeFiles {
		writeRow(f, []string{
			filepath.Base(fi.Path),
			fi.Path,
			fmt.Sprintf("%d", fi.Size),
			util.FormatSize(fi.Size),
			fi.ModTime.Format("2006-01-02 15:04:05"),
		})
	}
	return nil
}

func writeDuplicates(path string, r model.ScanResult) error {
	f, err := createCSV(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writeRow(f, []string{"グループ番号", "ファイル数", "グループ内合計サイズ", "ファイルサイズ", "ハッシュ値", "ファイルパス"})
	for _, g := range r.DuplicateGroups {
		for _, p := range g.Paths {
			writeRow(f, []string{
				fmt.Sprintf("%d", g.GroupNo),
				fmt.Sprintf("%d", g.FileCount),
				util.FormatSize(g.TotalSize),
				util.FormatSize(g.FileSize),
				g.Hash,
				p,
			})
		}
	}
	return nil
}

func writeLongPaths(path string, r model.ScanResult) error {
	f, err := createCSV(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writeRow(f, []string{"パス文字数", "フルパス", "サイズ", "最終更新日時"})
	for _, fi := range r.LongPaths {
		writeRow(f, []string{
			fmt.Sprintf("%d", len(fi.Path)),
			fi.Path,
			util.FormatSize(fi.Size),
			fi.ModTime.Format("2006-01-02 15:04:05"),
		})
	}
	return nil
}

func writeOldFiles(path string, r model.ScanResult) error {
	f, err := createCSV(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writeRow(f, []string{"ファイル名", "フルパス", "サイズ", "最終更新日時"})
	for _, fi := range r.OldFiles {
		writeRow(f, []string{
			filepath.Base(fi.Path),
			fi.Path,
			util.FormatSize(fi.Size),
			fi.ModTime.Format("2006-01-02 15:04:05"),
		})
	}
	return nil
}

func writeExtSummary(path string, r model.ScanResult) error {
	f, err := createCSV(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writeRow(f, []string{"拡張子", "ファイル数", "合計容量(バイト)", "合計容量"})
	for _, e := range r.ExtSummary {
		ext := e.Ext
		if ext == "(なし)" {
			ext = "(拡張子なし)"
		}
		writeRow(f, []string{
			ext,
			fmt.Sprintf("%d", e.FileCount),
			fmt.Sprintf("%d", e.TotalSize),
			util.FormatSize(e.TotalSize),
		})
	}
	return nil
}

func writeScanErrors(path string, r model.ScanResult) error {
	f, err := createCSV(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writeRow(f, []string{"パス", "エラー内容"})
	for _, e := range r.ScanErrors {
		writeRow(f, []string{e.Path, e.Error})
	}
	return nil
}

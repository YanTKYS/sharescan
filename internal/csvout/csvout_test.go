package csvout

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yantkys/sharescan/internal/model"
)

func makeTestResult() model.ScanResult {
	now := time.Date(2024, 4, 1, 10, 0, 0, 0, time.UTC)
	return model.ScanResult{
		Targets:    []string{"/tmp/share"},
		ScannedAt:  now,
		TotalFiles: 10,
		TotalDirs:  3,
		TotalSize:  1024 * 1024,
		ErrorCount: 1,
		LargeFiles: []model.FileInfo{
			{Path: "/tmp/share/big.dat", Size: 200 * 1024 * 1024, ModTime: now},
		},
		DuplicateGroups: []model.DuplicateGroup{
			{GroupNo: 1, FileCount: 2, FileSize: 100, TotalSize: 200, Hash: "abc123", Paths: []string{"/a/f1.txt", "/b/f2.txt"}},
		},
		LongPaths: []model.FileInfo{
			{Path: strings.Repeat("x", 250), Size: 100, ModTime: now},
		},
		OldFiles: []model.FileInfo{
			{Path: "/tmp/share/old.txt", Size: 50, ModTime: now.AddDate(-10, 0, 0)},
		},
		ExtSummary: []model.ExtStats{
			{Ext: ".txt", FileCount: 5, TotalSize: 500},
			{Ext: "(なし)", FileCount: 2, TotalSize: 200},
		},
		ScanErrors: []model.ScanError{
			{Path: "/tmp/share/denied", Error: "permission denied"},
		},
	}
}

func TestWriteAll_CreatesFiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "sharescan_csvtest_*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	result := makeTestResult()
	if err := WriteAll(result, dir); err != nil {
		t.Fatalf("WriteAll error: %v", err)
	}

	expected := []string{
		"summary.csv",
		"large_files.csv",
		"duplicate_files.csv",
		"long_paths.csv",
		"old_files.csv",
		"extension_summary.csv",
		"scan_errors.csv",
	}
	for _, name := range expected {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", name)
		}
	}
}

func TestWriteAll_BOM(t *testing.T) {
	dir, err := os.MkdirTemp("", "sharescan_bom_*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := WriteAll(makeTestResult(), dir); err != nil {
		t.Fatalf("WriteAll: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "summary.csv"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if len(data) < 3 || string(data[:3]) != bom {
		t.Errorf("summary.csv does not start with UTF-8 BOM")
	}
}

func TestWriteAll_ScanErrors(t *testing.T) {
	dir, err := os.MkdirTemp("", "sharescan_err_*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := WriteAll(makeTestResult(), dir); err != nil {
		t.Fatalf("WriteAll: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "scan_errors.csv"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "permission denied") {
		t.Errorf("scan_errors.csv should contain 'permission denied', got: %s", content)
	}
}

func TestWriteAll_ExtSummary_RenamesNone(t *testing.T) {
	dir, err := os.MkdirTemp("", "sharescan_ext_*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := WriteAll(makeTestResult(), dir); err != nil {
		t.Fatalf("WriteAll: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "extension_summary.csv"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "(拡張子なし)") {
		t.Errorf("extension_summary.csv should contain '(拡張子なし)', got: %s", content)
	}
}

func TestWriteAll_DuplicateFiles_MultipleRows(t *testing.T) {
	dir, err := os.MkdirTemp("", "sharescan_dup_*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := WriteAll(makeTestResult(), dir); err != nil {
		t.Fatalf("WriteAll: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "duplicate_files.csv"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	// Should have header + 2 rows (one per path in the group)
	lines := strings.Split(strings.TrimRight(string(data[3:]), "\r\n"), "\r\n") // skip BOM
	if len(lines) < 3 { // header + 2 data rows
		t.Errorf("duplicate_files.csv: expected at least 3 lines (header + 2 paths), got %d", len(lines))
	}
}

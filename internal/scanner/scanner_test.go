package scanner

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yantkys/sharescan/internal/config"
)

// createTempDir builds a temporary directory tree for testing and returns its path.
func createTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "sharescan_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func writeFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
}

func TestScan_BasicFileCount(t *testing.T) {
	root := createTempDir(t)

	writeFile(t, filepath.Join(root, "a.txt"), []byte("hello"))
	writeFile(t, filepath.Join(root, "b.docx"), []byte("world"))
	writeFile(t, filepath.Join(root, "sub", "c.pdf"), []byte("pdf content"))

	cfg := config.DefaultConfig()
	cfg.Targets = []string{root}

	s := New(cfg, nil)
	result := s.Scan()

	if result.TotalFiles != 3 {
		t.Errorf("TotalFiles: got %d, want 3", result.TotalFiles)
	}
	if result.TotalDirs != 1 {
		t.Errorf("TotalDirs: got %d, want 1 (sub)", result.TotalDirs)
	}
}

func TestScan_ExtensionExclusion(t *testing.T) {
	root := createTempDir(t)

	writeFile(t, filepath.Join(root, "keep.txt"), []byte("keep"))
	writeFile(t, filepath.Join(root, "skip.tmp"), []byte("skip"))
	writeFile(t, filepath.Join(root, "skip.bak"), []byte("skip"))

	cfg := config.DefaultConfig()
	cfg.Targets = []string{root}
	cfg.ExcludeExtensions = []string{".tmp", ".bak"}

	s := New(cfg, nil)
	result := s.Scan()

	if result.TotalFiles != 1 {
		t.Errorf("TotalFiles: got %d, want 1 (only keep.txt)", result.TotalFiles)
	}
}

func TestScan_DirectoryExclusion(t *testing.T) {
	root := createTempDir(t)

	writeFile(t, filepath.Join(root, "top.txt"), []byte("top"))
	writeFile(t, filepath.Join(root, ".git", "config"), []byte("git"))
	writeFile(t, filepath.Join(root, "src", "main.go"), []byte("go"))

	cfg := config.DefaultConfig()
	cfg.Targets = []string{root}
	cfg.ExcludeDirectories = []string{".git"}

	s := New(cfg, nil)
	result := s.Scan()

	// Should see top.txt and src/main.go but NOT .git/config
	if result.TotalFiles != 2 {
		t.Errorf("TotalFiles: got %d, want 2", result.TotalFiles)
	}
}

func TestScan_ExtensionSummary(t *testing.T) {
	root := createTempDir(t)

	writeFile(t, filepath.Join(root, "a.txt"), []byte("aaaa"))  // 4 B
	writeFile(t, filepath.Join(root, "b.txt"), []byte("bb"))    // 2 B
	writeFile(t, filepath.Join(root, "c.docx"), []byte("cccc")) // 4 B

	cfg := config.DefaultConfig()
	cfg.Targets = []string{root}

	s := New(cfg, nil)
	result := s.Scan()

	if len(result.ExtSummary) != 2 {
		t.Errorf("ExtSummary length: got %d, want 2", len(result.ExtSummary))
	}

	extMap := make(map[string]int64)
	for _, es := range result.ExtSummary {
		extMap[es.Ext] = es.TotalSize
	}
	if extMap[".txt"] != 6 {
		t.Errorf(".txt total size: got %d, want 6", extMap[".txt"])
	}
	if extMap[".docx"] != 4 {
		t.Errorf(".docx total size: got %d, want 4", extMap[".docx"])
	}
}

func TestScan_LargeFileDetection(t *testing.T) {
	root := createTempDir(t)

	// Create a file larger than the default 100 MB threshold using sparse file approach.
	// Instead, lower the threshold in config.
	writeFile(t, filepath.Join(root, "big.dat"), make([]byte, 200))  // 200 B
	writeFile(t, filepath.Join(root, "small.dat"), make([]byte, 50)) // 50 B

	cfg := config.DefaultConfig()
	cfg.Targets = []string{root}
	cfg.LargeFileThresholdMB = 0 // 0 MB threshold → 0 bytes → everything qualifies
	// Actually use 1 byte threshold via direct field.
	// Override: set threshold to something that catches only big.dat
	// We set threshold bytes directly by setting LargeFileThresholdMB such that 200 >= threshold > 50.
	// 1 MB = 1048576 B; instead set to a small value by adjusting the config.
	// Since the threshold is in MB, the smallest non-zero value is 1 MB.
	// Let's test differently: set threshold to 0 MB. The code does LargeFileThresholdMB * 1024 * 1024.
	// 0 * 1024 * 1024 = 0. So every file qualifies.

	s := New(cfg, nil)
	result := s.Scan()

	if len(result.LargeFiles) != 2 {
		t.Errorf("LargeFiles (0 MB threshold): got %d, want 2", len(result.LargeFiles))
	}
}

func TestScan_LargeFileThreshold(t *testing.T) {
	root := createTempDir(t)

	// Create files using actual size bytes.
	big := make([]byte, 2*1024*1024)   // 2 MB
	small := make([]byte, 512*1024)    // 0.5 MB

	writeFile(t, filepath.Join(root, "big.dat"), big)
	writeFile(t, filepath.Join(root, "small.dat"), small)

	cfg := config.DefaultConfig()
	cfg.Targets = []string{root}
	cfg.LargeFileThresholdMB = 1 // 1 MB threshold

	s := New(cfg, nil)
	result := s.Scan()

	if len(result.LargeFiles) != 1 {
		t.Errorf("LargeFiles (1 MB threshold): got %d, want 1", len(result.LargeFiles))
	}
	if len(result.LargeFiles) > 0 && filepath.Base(result.LargeFiles[0].Path) != "big.dat" {
		t.Errorf("expected big.dat in LargeFiles, got %s", result.LargeFiles[0].Path)
	}
}

func TestScan_DuplicateDetectionWithHash(t *testing.T) {
	root := createTempDir(t)

	content := []byte("duplicate content")
	writeFile(t, filepath.Join(root, "dup1.txt"), content)
	writeFile(t, filepath.Join(root, "dup2.txt"), content)
	writeFile(t, filepath.Join(root, "unique.txt"), []byte("different content here"))

	cfg := config.DefaultConfig()
	cfg.Targets = []string{root}
	cfg.EnableHashDuplicateCheck = true

	s := New(cfg, nil)
	result := s.Scan()

	if len(result.DuplicateGroups) != 1 {
		t.Errorf("DuplicateGroups: got %d, want 1", len(result.DuplicateGroups))
	}
	if len(result.DuplicateGroups) > 0 {
		g := result.DuplicateGroups[0]
		if g.FileCount != 2 {
			t.Errorf("DuplicateGroups[0].FileCount: got %d, want 2", g.FileCount)
		}
		if g.Hash == "" {
			t.Error("expected non-empty hash in duplicate group")
		}
	}
}

func TestScan_SameSizeDifferentContent_NoFalsePositive(t *testing.T) {
	root := createTempDir(t)

	// Same size, different content — should NOT be duplicates when hash check is on.
	writeFile(t, filepath.Join(root, "file1.txt"), []byte("aaa"))
	writeFile(t, filepath.Join(root, "file2.txt"), []byte("bbb"))

	cfg := config.DefaultConfig()
	cfg.Targets = []string{root}
	cfg.EnableHashDuplicateCheck = true

	s := New(cfg, nil)
	result := s.Scan()

	if len(result.DuplicateGroups) != 0 {
		t.Errorf("expected no duplicate groups for same-size different-content files, got %d", len(result.DuplicateGroups))
	}
}

func TestScan_DuplicateWithoutHash(t *testing.T) {
	root := createTempDir(t)

	// Same size, different content — with hash disabled they ARE grouped.
	writeFile(t, filepath.Join(root, "file1.txt"), []byte("aaa"))
	writeFile(t, filepath.Join(root, "file2.txt"), []byte("bbb"))

	cfg := config.DefaultConfig()
	cfg.Targets = []string{root}
	cfg.EnableHashDuplicateCheck = false

	s := New(cfg, nil)
	result := s.Scan()

	if len(result.DuplicateGroups) != 1 {
		t.Errorf("expected 1 duplicate group (size-only), got %d", len(result.DuplicateGroups))
	}
	if len(result.DuplicateGroups) > 0 && result.DuplicateGroups[0].Hash != "" {
		t.Error("expected empty hash when hash check is disabled")
	}
}

func TestScan_LongPathDetection(t *testing.T) {
	root := createTempDir(t)

	// Build a filename that makes the full path exceed the threshold.
	cfg := config.DefaultConfig()
	cfg.LongPathThreshold = 10 // very low for testing

	// Create a file whose path is long enough.
	writeFile(t, filepath.Join(root, "this_is_a_long_filename.txt"), []byte("x"))

	cfg.Targets = []string{root}
	s := New(cfg, nil)
	result := s.Scan()

	// The full path should exceed 10 characters.
	if len(result.LongPaths) == 0 {
		t.Error("expected at least one long path entry")
	}
}

func TestScan_OldFileDetection(t *testing.T) {
	root := createTempDir(t)
	filePath := filepath.Join(root, "old.txt")
	writeFile(t, filePath, []byte("old content"))

	// Set the file's modification time to 10 years ago.
	oldTime := time.Now().AddDate(-10, 0, 0)
	if err := os.Chtimes(filePath, oldTime, oldTime); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.Targets = []string{root}
	cfg.OldFileYears = 5

	s := New(cfg, nil)
	result := s.Scan()

	if len(result.OldFiles) != 1 {
		t.Errorf("OldFiles: got %d, want 1", len(result.OldFiles))
	}
}

func TestScan_ErrorHandling_NonexistentTarget(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Targets = []string{"/nonexistent/path/that/does/not/exist"}

	s := New(cfg, nil)
	result := s.Scan()

	// Should not panic. Should record an error.
	if result.ErrorCount == 0 {
		t.Error("expected at least one error for nonexistent target, got 0")
	}
}

package scanner

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yantkys/sharescan/internal/config"
	"github.com/yantkys/sharescan/internal/model"
)

// Scanner performs read-only recursive directory scanning.
type Scanner struct {
	Config config.Config
	Logger *log.Logger
}

// New returns a Scanner with the given config and logger.
// If logger is nil, a discard logger is used.
func New(cfg config.Config, logger *log.Logger) *Scanner {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}
	return &Scanner{Config: cfg, Logger: logger}
}

// Scan walks all target directories and returns aggregated results.
func (s *Scanner) Scan() model.ScanResult {
	result := model.ScanResult{
		Targets:   s.Config.Targets,
		ScannedAt: time.Now(),
	}

	// Build lookup sets for exclusions (lower-cased for case-insensitive match).
	excludeDirs := toLowerSet(s.Config.ExcludeDirectories)
	excludeExts := toLowerSet(s.Config.ExcludeExtensions)

	largeThreshold := s.Config.LargeFileThresholdMB * 1024 * 1024
	oldCutoff := time.Now().AddDate(-s.Config.OldFileYears, 0, 0)

	// allFiles collects every successfully scanned file.
	var allFiles []model.FileInfo

	for _, target := range s.Config.Targets {
		s.Logger.Printf("走査開始: %s", target)
		err := filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				s.Logger.Printf("アクセスエラー: %s: %v", path, err)
				result.ScanErrors = append(result.ScanErrors, model.ScanError{
					Path:  path,
					Error: err.Error(),
				})
				result.ErrorCount++
				// If the root itself is inaccessible, stop walking this target.
				if path == target {
					return fs.SkipAll
				}
				return nil
			}

			name := d.Name()

			if d.IsDir() {
				if _, excluded := excludeDirs[strings.ToLower(name)]; excluded {
					s.Logger.Printf("ディレクトリをスキップ: %s", path)
					return filepath.SkipDir
				}
				if path != target {
					result.TotalDirs++
				}
				return nil
			}

			// It's a file.
			ext := strings.ToLower(filepath.Ext(name))
			if _, excluded := excludeExts[ext]; excluded {
				return nil
			}

			info, err := d.Info()
			if err != nil {
				s.Logger.Printf("ファイル情報取得エラー: %s: %v", path, err)
				result.ScanErrors = append(result.ScanErrors, model.ScanError{
					Path:  path,
					Error: err.Error(),
				})
				result.ErrorCount++
				return nil
			}

			fi := model.FileInfo{
				Path:    path,
				Size:    info.Size(),
				ModTime: info.ModTime(),
				Ext:     ext,
			}

			allFiles = append(allFiles, fi)
			result.TotalFiles++
			result.TotalSize += fi.Size

			if fi.Size >= largeThreshold {
				result.LargeFiles = append(result.LargeFiles, fi)
			}

			if len(path) >= s.Config.LongPathThreshold {
				result.LongPaths = append(result.LongPaths, fi)
			}

			if fi.ModTime.Before(oldCutoff) {
				result.OldFiles = append(result.OldFiles, fi)
			}

			return nil
		})
		if err != nil {
			s.Logger.Printf("WalkDirエラー: %s: %v", target, err)
		}
	}

	// Sort large files by size descending.
	sort.Slice(result.LargeFiles, func(i, j int) bool {
		return result.LargeFiles[i].Size > result.LargeFiles[j].Size
	})

	// Extension summary.
	result.ExtSummary = buildExtSummary(allFiles)

	// Duplicate detection.
	result.DuplicateGroups = s.findDuplicates(allFiles)

	return result
}

// buildExtSummary aggregates file count and total size per extension.
func buildExtSummary(files []model.FileInfo) []model.ExtStats {
	m := make(map[string]*model.ExtStats)
	for _, f := range files {
		ext := f.Ext
		if ext == "" {
			ext = "(なし)"
		}
		if s, ok := m[ext]; ok {
			s.FileCount++
			s.TotalSize += f.Size
		} else {
			m[ext] = &model.ExtStats{Ext: ext, FileCount: 1, TotalSize: f.Size}
		}
	}
	stats := make([]model.ExtStats, 0, len(m))
	for _, v := range m {
		stats = append(stats, *v)
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].TotalSize != stats[j].TotalSize {
			return stats[i].TotalSize > stats[j].TotalSize
		}
		return stats[i].FileCount > stats[j].FileCount
	})
	return stats
}

// findDuplicates groups files by size, then optionally by hash.
func (s *Scanner) findDuplicates(files []model.FileInfo) []model.DuplicateGroup {
	// Group by size.
	sizeGroups := make(map[int64][]model.FileInfo)
	for _, f := range files {
		if f.Size > 0 { // skip empty files
			sizeGroups[f.Size] = append(sizeGroups[f.Size], f)
		}
	}

	var groups []model.DuplicateGroup
	groupNo := 1

	for size, sameSize := range sizeGroups {
		if len(sameSize) < 2 {
			continue
		}

		if !s.Config.EnableHashDuplicateCheck {
			// Without hash: treat all same-size files as one group.
			paths := make([]string, 0, len(sameSize))
			for _, f := range sameSize {
				paths = append(paths, f.Path)
			}
			sort.Strings(paths)
			groups = append(groups, model.DuplicateGroup{
				GroupNo:   groupNo,
				FileCount: len(paths),
				FileSize:  size,
				TotalSize: size * int64(len(paths)),
				Hash:      "",
				Paths:     paths,
			})
			groupNo++
			continue
		}

		// Hash each file and group by hash.
		hashGroups := make(map[string][]string)
		for _, f := range sameSize {
			h, err := hashFile(f.Path)
			if err != nil {
				s.Logger.Printf("ハッシュ計算エラー: %s: %v", f.Path, err)
				continue
			}
			hashGroups[h] = append(hashGroups[h], f.Path)
		}

		for hash, paths := range hashGroups {
			if len(paths) < 2 {
				continue
			}
			sort.Strings(paths)
			groups = append(groups, model.DuplicateGroup{
				GroupNo:   groupNo,
				FileCount: len(paths),
				FileSize:  size,
				TotalSize: size * int64(len(paths)),
				Hash:      hash,
				Paths:     paths,
			})
			groupNo++
		}
	}

	// Sort groups by TotalSize descending for consistent output.
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].TotalSize > groups[j].TotalSize
	})
	// Re-number after sort.
	for i := range groups {
		groups[i].GroupNo = i + 1
	}

	return groups
}

// hashFile computes the SHA-256 hex digest of the file at path.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("read: %w", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// toLowerSet converts a string slice to a map keyed by lower-cased values.
func toLowerSet(ss []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[strings.ToLower(s)] = struct{}{}
	}
	return m
}

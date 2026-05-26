package model

import "time"

// FileInfo represents metadata for a single scanned file.
type FileInfo struct {
	Path    string
	Size    int64
	ModTime time.Time
	Ext     string
	Hash    string
}

// DuplicateGroup holds a group of files that are considered duplicates.
type DuplicateGroup struct {
	GroupNo   int
	FileCount int
	TotalSize int64
	FileSize  int64
	Hash      string
	Paths     []string
}

// ExtStats holds aggregated statistics for a file extension.
type ExtStats struct {
	Ext       string
	FileCount int64
	TotalSize int64
}

// ScanError records a path and the error encountered while scanning it.
type ScanError struct {
	Path  string
	Error string
}

// ScanResult holds all aggregated results from a scan run.
type ScanResult struct {
	Targets         []string
	ScannedAt       time.Time
	TotalFiles      int64
	TotalDirs       int64
	TotalSize       int64
	ErrorCount      int64
	LargeFiles      []FileInfo
	DuplicateGroups []DuplicateGroup
	LongPaths       []FileInfo
	OldFiles        []FileInfo
	ExtSummary      []ExtStats
	ScanErrors      []ScanError
}

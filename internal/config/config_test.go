package config

import (
	"encoding/json"
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.OutputDir != "output" {
		t.Errorf("OutputDir: got %q, want %q", cfg.OutputDir, "output")
	}
	if cfg.LargeFileThresholdMB != 100 {
		t.Errorf("LargeFileThresholdMB: got %d, want 100", cfg.LargeFileThresholdMB)
	}
	if cfg.LongPathThreshold != 240 {
		t.Errorf("LongPathThreshold: got %d, want 240", cfg.LongPathThreshold)
	}
	if cfg.OldFileYears != 5 {
		t.Errorf("OldFileYears: got %d, want 5", cfg.OldFileYears)
	}
	if cfg.MaxRowsInHtml != 100 {
		t.Errorf("MaxRowsInHtml: got %d, want 100", cfg.MaxRowsInHtml)
	}
}

func TestApplyDefaults_FillsZeroValues(t *testing.T) {
	cfg := Config{} // all zero values
	ApplyDefaults(&cfg)

	if cfg.OutputDir != "output" {
		t.Errorf("OutputDir: got %q", cfg.OutputDir)
	}
	if cfg.LargeFileThresholdMB != 100 {
		t.Errorf("LargeFileThresholdMB: got %d", cfg.LargeFileThresholdMB)
	}
	if cfg.LongPathThreshold != 240 {
		t.Errorf("LongPathThreshold: got %d", cfg.LongPathThreshold)
	}
	if cfg.OldFileYears != 5 {
		t.Errorf("OldFileYears: got %d", cfg.OldFileYears)
	}
	if cfg.MaxRowsInHtml != 100 {
		t.Errorf("MaxRowsInHtml: got %d", cfg.MaxRowsInHtml)
	}
}

func TestApplyDefaults_PreservesExistingValues(t *testing.T) {
	cfg := Config{
		OutputDir:            "mydir",
		LargeFileThresholdMB: 50,
		LongPathThreshold:    200,
		OldFileYears:         10,
		MaxRowsInHtml:        500,
	}
	ApplyDefaults(&cfg)

	if cfg.OutputDir != "mydir" {
		t.Errorf("OutputDir should not be overwritten: got %q", cfg.OutputDir)
	}
	if cfg.LargeFileThresholdMB != 50 {
		t.Errorf("LargeFileThresholdMB should not be overwritten: got %d", cfg.LargeFileThresholdMB)
	}
	if cfg.LongPathThreshold != 200 {
		t.Errorf("LongPathThreshold should not be overwritten: got %d", cfg.LongPathThreshold)
	}
	if cfg.OldFileYears != 10 {
		t.Errorf("OldFileYears should not be overwritten: got %d", cfg.OldFileYears)
	}
	if cfg.MaxRowsInHtml != 500 {
		t.Errorf("MaxRowsInHtml should not be overwritten: got %d", cfg.MaxRowsInHtml)
	}
}

func TestLoadConfig(t *testing.T) {
	// Write a temporary JSON config file.
	raw := Config{
		Targets:              []string{"/tmp/share1", "/tmp/share2"},
		OutputDir:            "out",
		LargeFileThresholdMB: 200,
		LongPathThreshold:    260,
		OldFileYears:         3,
		MaxRowsInHtml:        50,
		ExcludeDirectories:   []string{".git"},
		ExcludeExtensions:    []string{".tmp"},
	}
	data, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("failed to marshal test config: %v", err)
	}

	f, err := os.CreateTemp("", "sharescan_config_*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(data); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	f.Close()

	cfg, err := LoadConfig(f.Name())
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}

	if len(cfg.Targets) != 2 {
		t.Errorf("Targets length: got %d, want 2", len(cfg.Targets))
	}
	if cfg.Targets[0] != "/tmp/share1" {
		t.Errorf("Targets[0]: got %q", cfg.Targets[0])
	}
	if cfg.OutputDir != "out" {
		t.Errorf("OutputDir: got %q", cfg.OutputDir)
	}
	if cfg.LargeFileThresholdMB != 200 {
		t.Errorf("LargeFileThresholdMB: got %d", cfg.LargeFileThresholdMB)
	}
	if cfg.LongPathThreshold != 260 {
		t.Errorf("LongPathThreshold: got %d", cfg.LongPathThreshold)
	}
	if cfg.OldFileYears != 3 {
		t.Errorf("OldFileYears: got %d", cfg.OldFileYears)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	f, err := os.CreateTemp("", "sharescan_bad_*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString("this is not json {{{{")
	f.Close()

	_, err = LoadConfig(f.Name())
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

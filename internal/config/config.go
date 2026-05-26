package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds all configuration fields for the scanner.
type Config struct {
	Targets                  []string `json:"targets"`
	OutputDir                string   `json:"outputDir"`
	LargeFileThresholdMB     int64    `json:"largeFileThresholdMB"`
	LongPathThreshold        int      `json:"longPathThreshold"`
	OldFileYears             int      `json:"oldFileYears"`
	EnableHashDuplicateCheck bool     `json:"enableHashDuplicateCheck"`
	MaxRowsInHtml            int      `json:"maxRowsInHtml"`
	ExcludeDirectories       []string `json:"excludeDirectories"`
	ExcludeExtensions        []string `json:"excludeExtensions"`
}

// DefaultConfig returns a Config with default values applied.
func DefaultConfig() Config {
	cfg := Config{}
	ApplyDefaults(&cfg)
	return cfg
}

// ApplyDefaults fills zero/empty values in cfg with sensible defaults.
func ApplyDefaults(cfg *Config) {
	if cfg.OutputDir == "" {
		cfg.OutputDir = "output"
	}
	if cfg.LargeFileThresholdMB == 0 {
		cfg.LargeFileThresholdMB = 100
	}
	if cfg.LongPathThreshold == 0 {
		cfg.LongPathThreshold = 240
	}
	if cfg.OldFileYears == 0 {
		cfg.OldFileYears = 5
	}
	if cfg.MaxRowsInHtml == 0 {
		cfg.MaxRowsInHtml = 100
	}
}

// LoadConfig reads and parses a JSON config file from path.
// Defaults are applied after loading.
func LoadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("設定ファイルを開けません: %w", err)
	}
	defer f.Close()

	var cfg Config
	dec := json.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("設定ファイルの解析に失敗しました: %w", err)
	}

	ApplyDefaults(&cfg)
	return cfg, nil
}

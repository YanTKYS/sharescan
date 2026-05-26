package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/yantkys/sharescan/internal/config"
	"github.com/yantkys/sharescan/internal/csvout"
	"github.com/yantkys/sharescan/internal/report"
	"github.com/yantkys/sharescan/internal/scanner"
	"github.com/yantkys/sharescan/internal/util"
)

const version = "sharescan v0.1.0"

func printUsage() {
	fmt.Fprintf(os.Stderr, `使用方法:
  sharescan <対象フォルダパス>
  sharescan --config <設定ファイル.json>
  sharescan --help
  sharescan --version

例:
  sharescan "\\server\share"
  sharescan --config config.json

`)
}

func main() {
	var (
		configPath  = flag.String("config", "", "設定ファイル（JSON）のパス")
		showVersion = flag.Bool("version", false, "バージョンを表示して終了")
		showHelp    = flag.Bool("help", false, "使用方法を表示して終了")
	)

	flag.Usage = printUsage
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if *showHelp {
		printUsage()
		os.Exit(0)
	}

	var cfg config.Config

	switch {
	case *configPath != "":
		// Load from config file.
		loaded, err := config.LoadConfig(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: 設定ファイルの読み込みに失敗しました: %v\n", err)
			os.Exit(1)
		}
		cfg = loaded
		if len(cfg.Targets) == 0 {
			fmt.Fprintf(os.Stderr, "エラー: 設定ファイルに targets が指定されていません。\n")
			os.Exit(1)
		}

	case flag.NArg() >= 1:
		// Target path from positional argument.
		targetPath := flag.Arg(0)
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "エラー: 対象フォルダが存在しません: %s\n", targetPath)
			os.Exit(1)
		}
		cfg = config.DefaultConfig()
		cfg.Targets = []string{targetPath}

	default:
		fmt.Fprintf(os.Stderr, "エラー: 対象フォルダまたは設定ファイルを指定してください。\n\n")
		printUsage()
		os.Exit(1)
	}

	// Create output directory.
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 出力ディレクトリの作成に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// Set up logger: writes to both stdout and scan_log.txt.
	logPath := filepath.Join(cfg.OutputDir, "scan_log.txt")
	logFile, err := os.Create(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "警告: ログファイルを作成できません (%v)。標準出力のみに記録します。\n", err)
		logFile = nil
	}

	var logWriter io.Writer
	if logFile != nil {
		defer logFile.Close()
		logWriter = io.MultiWriter(os.Stdout, logFile)
	} else {
		logWriter = os.Stdout
	}

	logger := log.New(logWriter, "[sharescan] ", log.LstdFlags)
	logger.Printf("%s 走査を開始します", version)
	logger.Printf("出力先: %s", cfg.OutputDir)

	// Run scan.
	s := scanner.New(cfg, logger)
	result := s.Scan()

	logger.Printf("走査完了: ファイル %s 件, フォルダ %s 件, 総容量 %s",
		util.FormatNumber(result.TotalFiles),
		util.FormatNumber(result.TotalDirs),
		util.FormatSize(result.TotalSize),
	)
	if result.ErrorCount > 0 {
		logger.Printf("走査エラー: %s 件", util.FormatNumber(result.ErrorCount))
	}

	// Write CSV files.
	logger.Printf("CSVファイルを出力中...")
	if err := csvout.WriteAll(result, cfg.OutputDir); err != nil {
		logger.Printf("警告: CSV出力エラー: %v", err)
	}

	// Generate HTML report.
	htmlPath := filepath.Join(cfg.OutputDir, "report.html")
	logger.Printf("HTMLレポートを生成中...")
	if err := report.GenerateHTML(result, cfg, htmlPath); err != nil {
		logger.Printf("警告: HTMLレポート生成エラー: %v", err)
	}

	// Print summary to stdout.
	fmt.Println()
	fmt.Println("=== 走査結果サマリー ===")
	fmt.Printf("走査日時      : %s\n", result.ScannedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("総ファイル数  : %s 件\n", util.FormatNumber(result.TotalFiles))
	fmt.Printf("総フォルダ数  : %s 件\n", util.FormatNumber(result.TotalDirs))
	fmt.Printf("総容量        : %s\n", util.FormatSize(result.TotalSize))
	fmt.Printf("大容量ファイル: %s 件\n", util.FormatNumber(int64(len(result.LargeFiles))))
	fmt.Printf("重複候補      : %s グループ\n", util.FormatNumber(int64(len(result.DuplicateGroups))))
	fmt.Printf("長パス候補    : %s 件\n", util.FormatNumber(int64(len(result.LongPaths))))
	fmt.Printf("古いファイル  : %s 件\n", util.FormatNumber(int64(len(result.OldFiles))))
	if result.ErrorCount > 0 {
		fmt.Printf("走査エラー    : %s 件\n", util.FormatNumber(result.ErrorCount))
	}
	fmt.Println()
	fmt.Printf("レポート出力先: %s\n", cfg.OutputDir)
}

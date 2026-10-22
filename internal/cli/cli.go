package cli

import (
	"errors"
	"flag"
	"fmt"
	"ghostops/internal/analyzer"
	"ghostops/internal/parser"
	"ghostops/internal/report"
	"io"
	"os"
)

type Config struct {
	LogFilePath    string
	OutputJSONPath string
}

func ParseFlags(args []string, errOutput io.Writer) (*Config, error) {
	fs := flag.NewFlagSet("ghostops", flag.ContinueOnError)
	fs.SetOutput(errOutput)

	var outputJSON string
	fs.StringVar(&outputJSON, "o", "", "Path to export report as JSON (optional)")

	fs.Usage = func() {
		fmt.Fprintf(errOutput, "Usage: ghostops [options] <log-file-path>\n\n")
		fmt.Fprintf(errOutput, "Options:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	positional := fs.Args()
	if len(positional) < 1 {
		fs.Usage()
		return nil, errors.New("missing required log file path argument")
	}

	return &Config{
		LogFilePath:    positional[0],
		OutputJSONPath: outputJSON,
	}, nil
}

func Run(cfg *Config) error {
	file, err := os.Open(cfg.LogFilePath)
	if err != nil {
		return fmt.Errorf("failed to open log file %q: %w", cfg.LogFilePath, err)
	}
	defer file.Close()

	sc := parser.NewScanner(file)
	az := analyzer.New()

	for sc.Scan() {
		if sc.IsMalformed() {
			az.RecordMalformed()
		} else {
			az.Process(sc.Entry())
		}
	}

	if err := sc.Err(); err != nil {
		return fmt.Errorf("error reading log file: %w", err)
	}

	rep := az.Report(cfg.LogFilePath)

	terminalReporter := report.NewTerminalReporter()
	if err := terminalReporter.Render(rep, os.Stdout); err != nil {
		return fmt.Errorf("failed to render terminal report: %w", err)
	}

	if cfg.OutputJSONPath != "" {
		if err := report.ExportJSON(cfg.OutputJSONPath, rep); err != nil {
			return fmt.Errorf("failed to export JSON report: %w", err)
		}
		fmt.Printf("Report successfully exported to: %s\n\n", cfg.OutputJSONPath)
	}

	return nil
}

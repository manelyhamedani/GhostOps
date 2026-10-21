package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/manelyhamedani/GhostOps/internal/config"
	logprocessor "github.com/manelyhamedani/GhostOps/internal/logProcessor"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var configPath string
	var jsonFilePath string

	flag.StringVar(&configPath, "c", "config.json", "path to config file")
	flag.StringVar(&jsonFilePath, "o", "", "path to output json report")

	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage: %s [flags] [logfile]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(out, "Reads logfile and reports summary and incidents.")
		fmt.Fprintf(out, "Flags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() != 1 {
		return fmt.Errorf("Expected at one logfile, got %d", flag.NArg())
	}

	logPath := flag.Arg(0)

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		return fmt.Errorf("Load config %q: %w", configPath, err)
	}

	analyzer, err := logprocessor.NewAnalyzer(*cfg)
	if err != nil {
		return fmt.Errorf("Init analyzer: %w", err)
	}

	if err = logprocessor.ProcessLogs(logPath, analyzer); err != nil {
		return fmt.Errorf("Process logs: %w", err)
	}

	if jsonFilePath != "" {
		jsonFile, err := os.Create(jsonFilePath)
		if err != nil {
			return fmt.Errorf("Create json file: %w", err)
		}

		if err = analyzer.ReportJSON(jsonFile); err != nil {
			return fmt.Errorf("Report json: %w", err)
		}
	}

	pager := exec.Command("less", "-R", "-F", "-X")
	pager.Stdout = os.Stdout
	pager.Stderr = os.Stderr

	pipe, err := pager.StdinPipe()
	if err != nil {
		return fmt.Errorf("Pager stding: %w", err)
	}
	if err = pager.Start(); err != nil {
		return fmt.Errorf("Start pager: %w", err)
	}

	reportErr := analyzer.ReportText(pipe)
	pipe.Close()
	waitErr := pager.Wait()

	if errors.Is(reportErr, syscall.EPIPE) || errors.Is(reportErr, io.ErrClosedPipe) {
		reportErr = nil
	}

	if reportErr != nil {
		return fmt.Errorf("Report text: %w", reportErr)
	}

	return waitErr
}

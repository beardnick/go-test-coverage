package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/beardnick/go-test-coverage/internal/render"
	"github.com/beardnick/go-test-coverage/internal/report"
)

func main() {
	profilePath := flag.String("profile", "coverage.out", "path to coverprofile file")
	outputPath := flag.String("out", "coverage.html", "output HTML file")
	root := flag.String("root", "", "root directory for resolving source files (defaults to profile directory)")
	title := flag.String("title", "Go Coverage Report", "report title")
	flag.Parse()

	if *profilePath == "" {
		fmt.Fprintln(os.Stderr, "-profile cannot be empty")
		os.Exit(2)
	}

	rootPath := *root
	if rootPath == "" {
		rootPath = filepath.Dir(*profilePath)
	}

	reportData, err := report.Generate(*profilePath, rootPath, *title)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	outputFile, err := os.Create(*outputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer outputFile.Close()

	if err := render.HTML(outputFile, reportData); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

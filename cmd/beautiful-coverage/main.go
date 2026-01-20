package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"go-test-coverage/internal/render"
	"go-test-coverage/internal/report"
)

func main() {
	profilePath := flag.String("profile", "", "path to coverprofile file")
	outputPath := flag.String("out", "coverage.html", "output HTML file")
	root := flag.String("root", ".", "root directory for resolving source files")
	title := flag.String("title", "Go Coverage Report", "report title")
	assetsPath := flag.String("assets", "assets", "assets directory for styles and scripts")
	flag.Parse()

	if *profilePath == "" {
		fmt.Fprintln(os.Stderr, "-profile is required")
		os.Exit(2)
	}

	reportData, err := report.Generate(*profilePath, *root, *title)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	assetsDir := filepath.Clean(*assetsPath)
	assetsOutputDir := assetsDir
	if !filepath.IsAbs(assetsOutputDir) {
		assetsOutputDir = filepath.Join(filepath.Dir(*outputPath), assetsOutputDir)
	}
	if err := render.CopyAssets(assetsOutputDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	reportData.AssetsPath = filepath.ToSlash(assetsDir)

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

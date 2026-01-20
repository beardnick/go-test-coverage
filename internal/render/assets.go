package render

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const assetsRoot = "assets"

//go:embed assets/highlight/*
var embeddedAssets embed.FS

func CopyAssets(destDir string) error {
	if destDir == "" {
		return fmt.Errorf("assets destination is empty")
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("create assets dir: %w", err)
	}

	return fs.WalkDir(embeddedAssets, assetsRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		relative := strings.TrimPrefix(path, assetsRoot+"/")
		if relative == path {
			return fmt.Errorf("unexpected asset path: %s", path)
		}

		outputPath := filepath.Join(destDir, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return fmt.Errorf("create asset dir: %w", err)
		}

		source, err := embeddedAssets.Open(path)
		if err != nil {
			return fmt.Errorf("open asset: %w", err)
		}
		defer source.Close()

		destination, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("create asset: %w", err)
		}
		defer destination.Close()

		if _, err := io.Copy(destination, source); err != nil {
			return fmt.Errorf("write asset: %w", err)
		}

		return nil
	})
}

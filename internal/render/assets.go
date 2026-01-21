package render

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const assetsRoot = "assets"

//go:embed assets/highlight/*
var embeddedAssets embed.FS

type InlineAssets struct {
	HighlightDarkCSS  string
	HighlightLightCSS string
	HighlightJS       string
	HighlightGoJS     string
}

func LoadInlineAssets() (InlineAssets, error) {
	dark, err := readAsset(path.Join(assetsRoot, "highlight", "github-dark.min.css"))
	if err != nil {
		return InlineAssets{}, err
	}

	light, err := readAsset(path.Join(assetsRoot, "highlight", "github.min.css"))
	if err != nil {
		return InlineAssets{}, err
	}

	highlightJS, err := readAsset(path.Join(assetsRoot, "highlight", "highlight.min.js"))
	if err != nil {
		return InlineAssets{}, err
	}

	goJS, err := readAsset(path.Join(assetsRoot, "highlight", "go.min.js"))
	if err != nil {
		return InlineAssets{}, err
	}

	return InlineAssets{
		HighlightDarkCSS:  dark,
		HighlightLightCSS: light,
		HighlightJS:       highlightJS,
		HighlightGoJS:     goJS,
	}, nil
}

func readAsset(assetPath string) (string, error) {
	content, err := embeddedAssets.ReadFile(assetPath)
	if err != nil {
		return "", fmt.Errorf("read asset %s: %w", assetPath, err)
	}

	return string(content), nil
}

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

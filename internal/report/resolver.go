package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/tools/cover"
)

type fileResolver struct {
	root string
	pkgs map[string]*goPackage
}

type goPackage struct {
	ImportPath string
	Dir        string
	Error      *struct {
		Err string
	}
}

func newFileResolver(root string, profiles []*cover.Profile) (*fileResolver, error) {
	resolvedRoot := root
	if resolvedRoot == "" {
		resolvedRoot = "."
	}
	if absRoot, err := filepath.Abs(resolvedRoot); err == nil {
		resolvedRoot = absRoot
	}

	pkgs, err := findPackages(resolvedRoot, profiles)
	if err != nil {
		return nil, err
	}

	return &fileResolver{
		root: resolvedRoot,
		pkgs: pkgs,
	}, nil
}

func (resolver *fileResolver) resolve(fileName string) (string, string) {
	if filepath.IsAbs(fileName) {
		relative := fileName
		if resolver.root != "" {
			if rel, err := filepath.Rel(resolver.root, fileName); err == nil && !strings.HasPrefix(rel, "..") {
				relative = rel
			}
		}
		return fileName, relative
	}

	if strings.HasPrefix(fileName, ".") {
		relative := filepath.FromSlash(fileName)
		candidate := filepath.Join(resolver.root, relative)
		if fileExists(candidate) {
			return candidate, relative
		}
	}

	if resolved, relative := resolver.resolveFromPackages(fileName); resolved != "" {
		return resolved, relative
	}

	relative := filepath.FromSlash(fileName)
	return filepath.Join(resolver.root, relative), relative
}

func (resolver *fileResolver) resolveFromPackages(fileName string) (string, string) {
	pkg := resolver.pkgs[path.Dir(fileName)]
	if pkg == nil || pkg.Dir == "" || pkg.Error != nil {
		return "", ""
	}

	candidate := filepath.Join(pkg.Dir, path.Base(fileName))
	if !fileExists(candidate) {
		return "", ""
	}

	relative := candidate
	if relativePath, err := filepath.Rel(resolver.root, candidate); err == nil {
		relative = relativePath
	}

	return candidate, relative
}

func findPackages(root string, profiles []*cover.Profile) (map[string]*goPackage, error) {
	pkgs := make(map[string]*goPackage)
	list := make([]string, 0)

	for _, profile := range profiles {
		fileName := profile.FileName
		if strings.HasPrefix(fileName, ".") || filepath.IsAbs(fileName) {
			continue
		}

		pkg := path.Dir(fileName)
		if _, ok := pkgs[pkg]; ok {
			continue
		}
		pkgs[pkg] = nil
		list = append(list, pkg)
	}

	if len(list) == 0 {
		return pkgs, nil
	}

	cmd := exec.Command("go", append([]string{"list", "-e", "-json"}, list...)...)
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message != "" {
			return nil, fmt.Errorf("cannot run go list: %w: %s", err, message)
		}
		return nil, fmt.Errorf("cannot run go list: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(stdout))
	for {
		var pkg goPackage
		if err := decoder.Decode(&pkg); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("decoding go list json: %w", err)
		}
		pkgs[pkg.ImportPath] = &pkg
	}

	return pkgs, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

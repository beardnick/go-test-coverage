package report

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type fileResolver struct {
	root   string
	module moduleInfo
	pkgs   map[string]*goPackage
}

type goPackage struct {
	ImportPath string
	Dir        string
	Error      *struct {
		Err string
	}
}

type moduleInfo struct {
	path string
	base string
}

func newFileResolver(root string, profiles []Profile) (*fileResolver, error) {
	pkgs, err := findPackages(root, profiles)
	if err != nil {
		return nil, err
	}

	return &fileResolver{
		root:   root,
		module: loadModuleInfo(root),
		pkgs:   pkgs,
	}, nil
}

func (resolver *fileResolver) resolve(fileName string) (string, string) {
	if filepath.IsAbs(fileName) {
		return fileName, fileName
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

	return resolveByModuleFallback(fileName, resolver.root, resolver.module)
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

func findPackages(root string, profiles []Profile) (map[string]*goPackage, error) {
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

func loadModuleInfo(root string) moduleInfo {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return moduleInfo{}
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				modulePath := parts[1]
				return moduleInfo{
					path: modulePath,
					base: path.Base(modulePath),
				}
			}
		}
	}

	return moduleInfo{}
}

func resolveByModuleFallback(fileName, root string, module moduleInfo) (string, string) {
	relative := filepath.FromSlash(fileName)
	candidate := filepath.Join(root, relative)
	if fileExists(candidate) {
		return candidate, relative
	}

	if module.path != "" {
		prefix := module.path + "/"
		if strings.HasPrefix(fileName, prefix) {
			trimmed := strings.TrimPrefix(fileName, prefix)
			relativeTrimmed := filepath.FromSlash(trimmed)
			candidate = filepath.Join(root, relativeTrimmed)
			if fileExists(candidate) {
				return candidate, relativeTrimmed
			}
		}
	}

	if module.base != "" {
		prefix := module.base + "/"
		if strings.HasPrefix(fileName, prefix) {
			trimmed := strings.TrimPrefix(fileName, prefix)
			relativeTrimmed := filepath.FromSlash(trimmed)
			candidate = filepath.Join(root, relativeTrimmed)
			if fileExists(candidate) {
				return candidate, relativeTrimmed
			}
		}
	}

	if resolved, relativeResolved := resolveBySuffix(root, fileName); resolved != "" {
		return resolved, relativeResolved
	}

	return candidate, relative
}

func resolveBySuffix(root, fileName string) (string, string) {
	parts := strings.Split(fileName, "/")
	for index := 1; index < len(parts); index++ {
		trimmed := strings.Join(parts[index:], "/")
		if trimmed == "" {
			continue
		}
		relative := filepath.FromSlash(trimmed)
		candidate := filepath.Join(root, relative)
		if fileExists(candidate) {
			return candidate, relative
		}
	}
	return "", ""
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

package report

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Report struct {
	Title                string
	GeneratedAt          string
	TotalCoveragePercent string
	TotalCoverageClass   string
	CoveredStmts         int
	TotalStmts           int
	TotalFiles           int
	MissingFiles         int
	AssetsPath           string
	Files                []FileReport
}

type FileReport struct {
	Name               string
	CoveragePercent    string
	CoverageClass      string
	CoveredStmts       int
	TotalStmts         int
	Anchor             string
	Lines              []LineCoverage
	Missing            bool
	MissingDescription string
	RelativeSourcePath string
}

type LineCoverage struct {
	Number int
	Code   string
	Class  string
}

func Generate(profilePath, root, title string) (Report, error) {
	profiles, err := ParseProfiles(profilePath)
	if err != nil {
		return Report{}, err
	}

	module := loadModuleInfo(root)
	report := Report{
		Title:       title,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		AssetsPath:  "assets",
	}

	totalCovered := 0
	totalStmts := 0

	for _, profile := range profiles {
		fileReport, err := buildFileReport(profile, root, module)
		if err != nil {
			return Report{}, err
		}

		totalCovered += fileReport.CoveredStmts
		totalStmts += fileReport.TotalStmts
		if fileReport.Missing {
			report.MissingFiles++
		}
		report.Files = append(report.Files, fileReport)
	}

	report.CoveredStmts = totalCovered
	report.TotalStmts = totalStmts
	report.TotalFiles = len(report.Files)
	totalPercent := percent(totalCovered, totalStmts)
	report.TotalCoveragePercent = formatPercent(totalPercent)
	report.TotalCoverageClass = coverageClass(totalPercent)

	return report, nil
}

func buildFileReport(profile Profile, root string, module moduleInfo) (FileReport, error) {
	fileName := profile.FileName
	totalStmts := 0
	coveredStmts := 0

	for _, block := range profile.Blocks {
		totalStmts += block.NumStmt
		if block.Count > 0 {
			coveredStmts += block.NumStmt
		}
	}

	coveragePercent := percent(coveredStmts, totalStmts)
	report := FileReport{
		Name:            fileName,
		CoveredStmts:    coveredStmts,
		TotalStmts:      totalStmts,
		CoveragePercent: formatPercent(coveragePercent),
		CoverageClass:   coverageClass(coveragePercent),
		Anchor:          sanitizeAnchor(fileName),
	}

	sourcePath, relativePath := resolveSourcePath(fileName, root, module)
	report.RelativeSourcePath = relativePath

	content, err := os.ReadFile(sourcePath)
	if err != nil {
		report.Missing = true
		report.MissingDescription = fmt.Sprintf("source not found at %s", sourcePath)
		return report, nil
	}

	lines := strings.Split(string(content), "\n")
	lineStates := make([]lineState, len(lines))

	for _, block := range profile.Blocks {
		start := block.StartLine
		end := block.EndLine
		if start < 1 {
			start = 1
		}
		if end > len(lines) {
			end = len(lines)
		}
		for line := start; line <= end; line++ {
			state := &lineStates[line-1]
			state.hasStmt = true
			if block.Count > 0 {
				state.covered = true
			} else {
				state.missed = true
			}
		}
	}

	report.Lines = make([]LineCoverage, 0, len(lines))
	for index, raw := range lines {
		cleaned := strings.ReplaceAll(raw, "\t", "    ")
		state := lineStates[index]
		className := "not-tracked"
		if state.hasStmt {
			if state.covered && state.missed {
				className = "partial"
			} else if state.covered {
				className = "covered"
			} else if state.missed {
				className = "missed"
			}
		}

		report.Lines = append(report.Lines, LineCoverage{
			Number: index + 1,
			Code:   cleaned,
			Class:  className,
		})
	}

	return report, nil
}

type moduleInfo struct {
	path string
	base string
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

func resolveSourcePath(fileName, root string, module moduleInfo) (string, string) {
	if filepath.IsAbs(fileName) {
		return fileName, fileName
	}

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

	return candidate, relative
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

type lineState struct {
	hasStmt bool
	covered bool
	missed  bool
}

func percent(covered, total int) float64 {
	if total == 0 {
		return 100
	}
	return float64(covered) / float64(total) * 100
}

func formatPercent(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

func coverageClass(value float64) string {
	switch {
	case value >= 90:
		return "high"
	case value >= 75:
		return "medium"
	case value > 0:
		return "low"
	default:
		return "none"
	}
}

var anchorPattern = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

func sanitizeAnchor(value string) string {
	sanitized := anchorPattern.ReplaceAllString(value, "-")
	sanitized = strings.Trim(sanitized, "-")
	if sanitized == "" {
		return "file"
	}
	return sanitized
}

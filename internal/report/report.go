package report

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
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
	Tree                 []TreeNode
	Files                []FileReport
}

type TreeNode struct {
	Name            string
	Path            string
	RelativePath    string
	Anchor          string
	CoveragePercent string
	CoverageClass   string
	CoveredStmts    int
	TotalStmts      int
	IsDir           bool
	Children        []TreeNode
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
	Ranges string
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
	report.Tree = buildTree(report.Files)

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

			lineText := lines[line-1]
			maxCol := len(lineText) + 1
			startCol := 1
			endCol := maxCol
			if line == block.StartLine {
				startCol = block.StartCol
			}
			if line == block.EndLine {
				endCol = block.EndCol
			}
			if startCol < 1 {
				startCol = 1
			}
			if startCol > maxCol {
				startCol = maxCol
			}
			if endCol < startCol {
				endCol = startCol
			}
			if endCol > maxCol {
				endCol = maxCol
			}

			if block.Count > 0 {
				state.covered = true
			} else {
				state.missed = true
				if endCol > startCol {
					state.missedRanges = append(state.missedRanges, lineRange{
						start: startCol,
						end:   endCol,
					})
				}
			}
		}
	}

	report.Lines = make([]LineCoverage, 0, len(lines))
	for index, raw := range lines {
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

		partialRanges := ""
		if state.covered && state.missed {
			mergedRanges := mergeRanges(state.missedRanges)
			partialRanges = formatRanges(mergedRanges)
		}

		report.Lines = append(report.Lines, LineCoverage{
			Number: index + 1,
			Code:   raw,
			Class:  className,
			Ranges: partialRanges,
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

type treeEntry struct {
	name         string
	path         string
	children     map[string]*treeEntry
	file         *FileReport
	coveredStmts int
	totalStmts   int
}

func buildTree(files []FileReport) []TreeNode {
	root := &treeEntry{children: map[string]*treeEntry{}}

	for index := range files {
		file := &files[index]
		relative := file.RelativeSourcePath
		if relative == "" {
			relative = file.Name
		}
		relative = filepath.ToSlash(relative)
		if relative == "" {
			continue
		}

		parts := strings.Split(relative, "/")
		current := root
		currentPath := ""
		for partIndex, part := range parts {
			if part == "" {
				continue
			}
			if currentPath == "" {
				currentPath = part
			} else {
				currentPath = currentPath + "/" + part
			}

			next := current.children[part]
			if next == nil {
				next = &treeEntry{name: part, path: currentPath, children: map[string]*treeEntry{}}
				current.children[part] = next
			}
			if partIndex == len(parts)-1 {
				next.file = file
			}
			current = next
		}
	}

	computeTreeCoverage(root)
	return buildTreeNodes(root)
}

func computeTreeCoverage(entry *treeEntry) (int, int) {
	if entry == nil {
		return 0, 0
	}
	if entry.file != nil {
		entry.coveredStmts = entry.file.CoveredStmts
		entry.totalStmts = entry.file.TotalStmts
		return entry.coveredStmts, entry.totalStmts
	}
	covered := 0
	total := 0
	for _, child := range entry.children {
		childCovered, childTotal := computeTreeCoverage(child)
		covered += childCovered
		total += childTotal
	}
	entry.coveredStmts = covered
	entry.totalStmts = total
	return covered, total
}

func buildTreeNodes(entry *treeEntry) []TreeNode {
	directories := make([]TreeNode, 0)
	files := make([]TreeNode, 0)
	keys := make([]string, 0, len(entry.children))
	for key := range entry.children {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		child := entry.children[key]
		if child.file != nil {
			relativePath := filepath.ToSlash(child.file.RelativeSourcePath)
			if relativePath == "" {
				relativePath = child.path
			}
			files = append(files, TreeNode{
				Name:            child.name,
				Path:            child.path,
				RelativePath:    relativePath,
				Anchor:          child.file.Anchor,
				CoveragePercent: child.file.CoveragePercent,
				CoverageClass:   child.file.CoverageClass,
				IsDir:           false,
			})
			continue
		}

		coveragePercent := percent(child.coveredStmts, child.totalStmts)
		directories = append(directories, TreeNode{
			Name:            child.name,
			Path:            child.path,
			CoveredStmts:    child.coveredStmts,
			TotalStmts:      child.totalStmts,
			CoveragePercent: formatPercent(coveragePercent),
			CoverageClass:   coverageClass(coveragePercent),
			IsDir:           true,
			Children:        buildTreeNodes(child),
		})
	}

	return append(directories, files...)
}

type lineState struct {
	hasStmt      bool
	covered      bool
	missed       bool
	missedRanges []lineRange
}

type lineRange struct {
	start int
	end   int
}

func mergeRanges(ranges []lineRange) []lineRange {
	if len(ranges) == 0 {
		return nil
	}
	sort.Slice(ranges, func(i, j int) bool {
		if ranges[i].start == ranges[j].start {
			return ranges[i].end < ranges[j].end
		}
		return ranges[i].start < ranges[j].start
	})

	merged := make([]lineRange, 0, len(ranges))
	current := ranges[0]
	for _, item := range ranges[1:] {
		if item.start <= current.end {
			if item.end > current.end {
				current.end = item.end
			}
			continue
		}
		merged = append(merged, current)
		current = item
	}
	merged = append(merged, current)
	return merged
}

func formatRanges(ranges []lineRange) string {
	if len(ranges) == 0 {
		return ""
	}
	parts := make([]string, 0, len(ranges))
	for _, item := range ranges {
		parts = append(parts, fmt.Sprintf("%d-%d", item.start, item.end))
	}
	return strings.Join(parts, ",")
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

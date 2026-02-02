package report

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Profile struct {
	FileName string
	Mode     string
	Blocks   []ProfileBlock
}

type ProfileBlock struct {
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
	NumStmt   int
	Count     int
}

const modeSet = "set"

func ParseProfiles(path string) ([]Profile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open profile: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	profiles := make(map[string]*Profile)
	order := make([]string, 0)
	mode := ""
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "mode:") {
			mode = strings.TrimSpace(strings.TrimPrefix(line, "mode:"))
			continue
		}

		if mode == "" {
			return nil, fmt.Errorf("profile missing mode line")
		}

		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, fmt.Errorf("invalid profile line %d", lineNumber)
		}

		fileName, block, err := parseProfileBlock(fields[0])
		if err != nil {
			return nil, fmt.Errorf("parse block line %d: %w", lineNumber, err)
		}

		numStmt, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("parse statements line %d: %w", lineNumber, err)
		}
		count, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, fmt.Errorf("parse count line %d: %w", lineNumber, err)
		}

		block.NumStmt = numStmt
		block.Count = count

		profile, ok := profiles[fileName]
		if !ok {
			profile = &Profile{FileName: fileName, Mode: mode}
			profiles[fileName] = profile
			order = append(order, fileName)
		}
		profile.Blocks = append(profile.Blocks, block)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan profile: %w", err)
	}

	if mode == "" {
		return nil, fmt.Errorf("profile missing mode line")
	}

	result := make([]Profile, 0, len(order))
	for _, name := range order {
		profile := profiles[name]
		mergedBlocks, err := mergeProfileBlocks(profile.Blocks, mode)
		if err != nil {
			return nil, err
		}
		profile.Blocks = mergedBlocks
		profile.Mode = mode
		result = append(result, *profile)
	}

	return result, nil
}

func parseProfileBlock(field string) (string, ProfileBlock, error) {
	index := strings.LastIndex(field, ":")
	if index == -1 {
		return "", ProfileBlock{}, fmt.Errorf("missing ':'")
	}

	fileName := field[:index]
	ranges := field[index+1:]
	parts := strings.Split(ranges, ",")
	if len(parts) != 2 {
		return "", ProfileBlock{}, fmt.Errorf("invalid range")
	}

	startLine, startCol, err := parseLineColumn(parts[0])
	if err != nil {
		return "", ProfileBlock{}, fmt.Errorf("start range: %w", err)
	}
	endLine, endCol, err := parseLineColumn(parts[1])
	if err != nil {
		return "", ProfileBlock{}, fmt.Errorf("end range: %w", err)
	}

	return fileName, ProfileBlock{
		StartLine: startLine,
		StartCol:  startCol,
		EndLine:   endLine,
		EndCol:    endCol,
	}, nil
}

func parseLineColumn(value string) (int, int, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid line column")
	}
	line, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	col, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return line, col, nil
}

type blocksByStart []ProfileBlock

func (blocks blocksByStart) Len() int { return len(blocks) }

func (blocks blocksByStart) Less(index, other int) bool {
	left := blocks[index]
	right := blocks[other]
	if left.StartLine != right.StartLine {
		return left.StartLine < right.StartLine
	}
	return left.StartCol < right.StartCol
}

func (blocks blocksByStart) Swap(index, other int) {
	blocks[index], blocks[other] = blocks[other], blocks[index]
}

func mergeProfileBlocks(blocks []ProfileBlock, mode string) ([]ProfileBlock, error) {
	if len(blocks) == 0 {
		return blocks, nil
	}

	sort.Sort(blocksByStart(blocks))
	merged := make([]ProfileBlock, 0, len(blocks))
	merged = append(merged, blocks[0])

	for index := 1; index < len(blocks); index++ {
		block := blocks[index]
		lastIndex := len(merged) - 1
		last := merged[lastIndex]
		if sameBlockRange(block, last) {
			if block.NumStmt != last.NumStmt {
				return nil, fmt.Errorf("inconsistent NumStmt: changed from %d to %d", last.NumStmt, block.NumStmt)
			}
			if mode == modeSet {
				merged[lastIndex].Count |= block.Count
			} else {
				merged[lastIndex].Count += block.Count
			}
			continue
		}
		merged = append(merged, block)
	}

	return merged, nil
}

func sameBlockRange(left, right ProfileBlock) bool {
	return left.StartLine == right.StartLine &&
		left.StartCol == right.StartCol &&
		left.EndLine == right.EndLine &&
		left.EndCol == right.EndCol
}

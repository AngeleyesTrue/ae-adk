package platform

import (
	"os"
	"strings"
)

// VerifyPaths는 PATH 문자열의 각 경로가 실제로 존재하는지 검증한다.
func VerifyPaths(sys SystemInfo, pathStr string) []PathVerifyResult {
	sep := string(os.PathListSeparator)
	entries := strings.Split(pathStr, sep)

	results := make([]PathVerifyResult, 0, len(entries))
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		results = append(results, PathVerifyResult{
			Path:   entry,
			Exists: sys.DirExists(entry),
		})
	}
	return results
}

// FilterExistingPaths는 실제 존재하는 경로만 필터링하여 새 PATH 문자열을 반환한다.
func FilterExistingPaths(sys SystemInfo, pathStr string) string {
	sep := string(os.PathListSeparator)
	entries := strings.Split(pathStr, sep)

	var existing []string
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if sys.DirExists(entry) {
			existing = append(existing, entry)
		}
	}
	return strings.Join(existing, sep)
}

// CountMissing은 존재하지 않는 경로 수를 반환한다.
func CountMissing(results []PathVerifyResult) int {
	count := 0
	for _, r := range results {
		if !r.Exists {
			count++
		}
	}
	return count
}

// MissingPaths는 존재하지 않는 경로 목록을 반환한다.
func MissingPaths(results []PathVerifyResult) []string {
	var paths []string
	for _, r := range results {
		if !r.Exists {
			paths = append(paths, r.Path)
		}
	}
	return paths
}

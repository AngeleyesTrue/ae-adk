package platform

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const profileFileName = "platform-profile.json"

// ProfilePath는 프로필 저장 경로를 반환한다.
func ProfilePath(sys SystemInfo) string {
	return filepath.Join(sys.HomeDir(), ".ae", profileFileName)
}

// SaveProfile은 플랫폼 프로필을 JSON 파일로 저장한다.
func SaveProfile(sys SystemInfo, profile *PlatformProfile) error {
	path := ProfilePath(sys)
	dir := filepath.Dir(path)

	if !sys.DirExists(dir) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create profile directory: %w", err)
		}
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}

	if err := sys.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write profile: %w", err)
	}
	return nil
}

// LoadProfile은 기존 플랫폼 프로필을 로드한다.
// 파일이 없으면 nil, nil을 반환한다.
func LoadProfile(sys SystemInfo) (*PlatformProfile, error) {
	path := ProfilePath(sys)
	data, err := sys.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read profile: %w", err)
	}

	var profile PlatformProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("unmarshal profile: %w", err)
	}
	return &profile, nil
}

// CompareProfiles는 두 프로필 간의 차이점을 분석한다.
func CompareProfiles(old, new *PlatformProfile) *ProfileDiff {
	if old == nil {
		return nil
	}

	diff := &ProfileDiff{
		ChangedTools: make(map[string][2]string),
	}

	// PATH 비교
	oldPaths := make(map[string]bool, len(old.PATH))
	for _, p := range old.PATH {
		oldPaths[p] = true
	}
	newPaths := make(map[string]bool, len(new.PATH))
	for _, p := range new.PATH {
		newPaths[p] = true
	}
	for _, p := range new.PATH {
		if !oldPaths[p] {
			diff.AddedPaths = append(diff.AddedPaths, p)
		}
	}
	for _, p := range old.PATH {
		if !newPaths[p] {
			diff.RemovedPaths = append(diff.RemovedPaths, p)
		}
	}

	// 도구 버전 비교
	for tool, newVer := range new.ToolVersions {
		if oldVer, ok := old.ToolVersions[tool]; ok && oldVer != newVer {
			diff.ChangedTools[tool] = [2]string{oldVer, newVer}
		}
	}

	// 진단 상태 비교
	oldChecks := make(map[string]CheckStatus, len(old.Checks))
	for _, c := range old.Checks {
		oldChecks[c.Name] = c.Status
	}
	for _, c := range new.Checks {
		if oldStatus, ok := oldChecks[c.Name]; ok && oldStatus != c.Status {
			diff.StatusDiffs = append(diff.StatusDiffs, CheckStatusDiff{
				Name:      c.Name,
				OldStatus: oldStatus,
				NewStatus: c.Status,
			})
		}
	}

	return diff
}

// HasChanges는 프로필 차이가 있는지 확인한다.
func (d *ProfileDiff) HasChanges() bool {
	if d == nil {
		return false
	}
	return len(d.AddedPaths) > 0 || len(d.RemovedPaths) > 0 ||
		len(d.ChangedTools) > 0 || len(d.StatusDiffs) > 0
}

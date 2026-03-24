package platform

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// RunDiagnostics는 플랫폼 진단을 실행하고 PlatformProfile을 생성한다.
// smartPATH는 외부에서 BuildSmartPATH()로 생성한 값을 전달받아 중복 호출을 방지한다.
// @MX:ANCHOR: [AUTO] 플랫폼 진단의 핵심 오케스트레이션 함수
// @MX:REASON: [AUTO] fan_in>=3, CLI platform_common/win/mac에서 호출
func RunDiagnostics(sys SystemInfo, targetPlatform string, smartPATH string, pathResults []PathVerifyResult) (*PlatformProfile, error) {
	validator := NewValidator(sys, targetPlatform)

	checks := validator.RunChecks()

	// PATH 검증 결과로 경고 추가
	for _, r := range pathResults {
		if !r.Exists {
			checks = append(checks, PlatformCheck{
				Name:    "PATH: " + r.Path,
				Status:  StatusWarn,
				Message: "경로가 존재하지 않습니다",
			})
		}
	}

	// 도구 버전 수집 (병렬)
	toolVersions := collectToolVersions(sys)

	// PATH 엔트리 분리
	pathEntries := strings.Split(smartPATH, pathSep)

	profile := &PlatformProfile{
		Platform:     targetPlatform,
		Timestamp:    time.Now(),
		Checks:       checks,
		PATH:         pathEntries,
		ToolVersions: toolVersions,
	}

	return profile, nil
}

// collectToolVersions는 주요 도구들의 버전을 병렬로 수집한다.
func collectToolVersions(sys SystemInfo) map[string]string {
	type toolCmd struct {
		name string
		cmd  string
		args []string
	}
	tools := []toolCmd{
		{"go", "go", []string{"version"}},
		{"node", "node", []string{"--version"}},
		{"git", "git", []string{"--version"}},
		{"ae", "ae", []string{"--version"}},
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	versions := make(map[string]string, len(tools))

	for _, t := range tools {
		wg.Add(1)
		go func(t toolCmd) {
			defer wg.Done()
			out, err := sys.ExecCommand(t.cmd, t.args...)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				versions[t.name] = "not found"
			} else {
				versions[t.name] = strings.TrimSpace(out)
			}
		}(t)
	}
	wg.Wait()
	return versions
}

// FormatDiagnostics는 진단 결과를 텍스트로 포맷한다.
func FormatDiagnostics(profile *PlatformProfile, verbose bool) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Platform: %s\n", profile.Platform)
	fmt.Fprintf(&sb, "Timestamp: %s\n\n", profile.Timestamp.Format(time.RFC3339))

	// 진단 결과
	sb.WriteString("=== Diagnostics ===\n")
	maxLabel := 0
	for _, c := range profile.Checks {
		if len(c.Name) > maxLabel {
			maxLabel = len(c.Name)
		}
	}

	okCount, warnCount, failCount := 0, 0, 0
	for _, c := range profile.Checks {
		icon := statusIconText(c.Status)
		fmt.Fprintf(&sb, "%s %-*s  %s\n", icon, maxLabel, c.Name, c.Message)
		if verbose && c.Detail != "" {
			fmt.Fprintf(&sb, "    %s\n", c.Detail)
		}
		switch c.Status {
		case StatusOK:
			okCount++
		case StatusWarn:
			warnCount++
		case StatusFail:
			failCount++
		}
	}

	fmt.Fprintf(&sb, "\n%d passed / %d warnings / %d failed\n", okCount, warnCount, failCount)

	// 도구 버전
	sb.WriteString("\n=== Tool Versions ===\n")
	for tool, ver := range profile.ToolVersions {
		fmt.Fprintf(&sb, "  %-8s %s\n", tool+":", ver)
	}

	return sb.String()
}

// FormatDiff는 프로필 차이를 텍스트로 포맷한다.
func FormatDiff(diff *ProfileDiff) string {
	if diff == nil || !diff.HasChanges() {
		return "이전 프로필과 동일합니다.\n"
	}

	var sb strings.Builder
	sb.WriteString("=== Changes from Previous Profile ===\n")

	if len(diff.AddedPaths) > 0 {
		sb.WriteString("\nAdded paths:\n")
		for _, p := range diff.AddedPaths {
			fmt.Fprintf(&sb, "  + %s\n", p)
		}
	}
	if len(diff.RemovedPaths) > 0 {
		sb.WriteString("\nRemoved paths:\n")
		for _, p := range diff.RemovedPaths {
			fmt.Fprintf(&sb, "  - %s\n", p)
		}
	}
	if len(diff.ChangedTools) > 0 {
		sb.WriteString("\nChanged tools:\n")
		for tool, versions := range diff.ChangedTools {
			fmt.Fprintf(&sb, "  %s: %s -> %s\n", tool, versions[0], versions[1])
		}
	}
	if len(diff.StatusDiffs) > 0 {
		sb.WriteString("\nStatus changes:\n")
		for _, d := range diff.StatusDiffs {
			fmt.Fprintf(&sb, "  %s: %s -> %s\n", d.Name, d.OldStatus, d.NewStatus)
		}
	}

	return sb.String()
}

func statusIconText(status CheckStatus) string {
	switch status {
	case StatusOK:
		return "[OK]"
	case StatusWarn:
		return "[!!]"
	case StatusFail:
		return "[XX]"
	default:
		return "[??]"
	}
}

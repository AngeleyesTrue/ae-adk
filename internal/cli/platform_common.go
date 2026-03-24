package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AngeleyesTrue/ae-adk/internal/platform"
	"github.com/AngeleyesTrue/ae-adk/internal/template"
)

// addPlatformFlags는 ae win/mac 명령어에 공통 플래그를 추가한다.
func addPlatformFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("force", false, "플랫폼 불일치 경고 무시")
	cmd.Flags().BoolP("verbose", "v", false, "상세 진단 출력")
	cmd.Flags().Bool("json", false, "JSON 형식 출력")
	cmd.Flags().Bool("auto", false, "존재하지 않는 경로 자동 제외")
	cmd.Flags().Bool("dry-run", false, "실제 변경 없이 미리보기")
	cmd.Flags().Bool("skip-backup", false, "settings.json 백업 건너뛰기")
}

// parsePlatformFlags는 명령어에서 공통 플래그를 파싱한다.
func parsePlatformFlags(cmd *cobra.Command) platform.PlatformFlags {
	return platform.PlatformFlags{
		Force:      getBoolFlag(cmd, "force"),
		Verbose:    getBoolFlag(cmd, "verbose"),
		JSON:       getBoolFlag(cmd, "json"),
		Auto:       getBoolFlag(cmd, "auto"),
		DryRun:     getBoolFlag(cmd, "dry-run"),
		SkipBackup: getBoolFlag(cmd, "skip-backup"),
	}
}

// @MX:ANCHOR: [AUTO] 플랫폼 전환 명령어의 핵심 실행 흐름
// @MX:REASON: [AUTO] fan_in=2, platform_win.go/platform_mac.go에서 호출
func runPlatformCommand(cmd *cobra.Command, targetPlatform string) error {
	flags := parsePlatformFlags(cmd)
	out := cmd.OutOrStdout()
	sys := &platform.DefaultSystemInfo{}

	// 1. 플랫폼 감지 (REQ-001)
	currentPlatform := runtime.GOOS
	if currentPlatform != targetPlatform && !flags.Force {
		_, _ = fmt.Fprintf(out, "%s 현재 플랫폼(%s)과 대상 플랫폼(%s)이 다릅니다.\n",
			symWarning(), currentPlatform, targetPlatform)
		_, _ = fmt.Fprintln(out, "  --force 플래그로 강제 실행할 수 있습니다.")
		return nil
	}

	// 2. settings.json 백업 (REQ-002)
	settingsPath := findSettingsJSON()
	if settingsPath != "" && !flags.SkipBackup && !flags.DryRun {
		backupPath, err := platform.BackupSettings(sys, settingsPath)
		if err != nil {
			_, _ = fmt.Fprintf(out, "%s 백업 실패: %v\n", symError(), err)
		} else {
			_, _ = fmt.Fprintf(out, "%s settings.json 백업 완료: %s\n", symSuccess(), filepath.Base(backupPath))
			if err := platform.CleanupOldBackups(settingsPath); err != nil {
				_, _ = fmt.Fprintf(out, "%s 오래된 백업 정리 실패: %v\n", symWarning(), err)
			}
		}
	}

	// 3. PATH 재구성 (REQ-003)
	smartPATH := template.BuildSmartPATH()

	if flags.Auto {
		smartPATH = platform.FilterExistingPaths(sys, smartPATH)
	}

	if flags.DryRun {
		_, _ = fmt.Fprintln(out, "\n=== Dry Run: PATH Preview ===")
		sep := string(os.PathListSeparator)
		for _, entry := range strings.Split(smartPATH, sep) {
			exists := sys.DirExists(entry)
			icon := symSuccess()
			if !exists {
				icon = symWarning()
			}
			_, _ = fmt.Fprintf(out, "  %s %s\n", icon, entry)
		}
	} else if settingsPath != "" {
		if err := updateSettingsPATH(sys, settingsPath, smartPATH); err != nil {
			_, _ = fmt.Fprintf(out, "%s PATH 업데이트 실패: %v\n", symError(), err)
		} else {
			_, _ = fmt.Fprintf(out, "%s settings.json PATH 업데이트 완료\n", symSuccess())
		}
	}

	// 4-5. 플랫폼별 검증 (REQ-004, REQ-005)
	profile, err := platform.RunDiagnostics(sys, targetPlatform, flags)
	if err != nil {
		return fmt.Errorf("run diagnostics: %w", err)
	}

	// 6. 진단 결과 출력 (REQ-006)
	if flags.JSON {
		data, _ := json.MarshalIndent(profile, "", "  ")
		_, _ = fmt.Fprintln(out, string(data))
	} else {
		_, _ = fmt.Fprintln(out)
		formatted := platform.FormatDiagnostics(profile, flags.Verbose)
		_, _ = fmt.Fprint(out, formatted)
	}

	// 7. 프로필 저장 및 비교 (REQ-007)
	if !flags.DryRun {
		oldProfile, _ := platform.LoadProfile(sys)
		if err := platform.SaveProfile(sys, profile); err != nil {
			_, _ = fmt.Fprintf(out, "\n%s 프로필 저장 실패: %v\n", symWarning(), err)
		} else {
			_, _ = fmt.Fprintf(out, "\n%s 프로필 저장 완료: %s\n", symSuccess(), platform.ProfilePath(sys))
		}

		// 이전 프로필과 비교
		if oldProfile != nil {
			diff := platform.CompareProfiles(oldProfile, profile)
			if diff.HasChanges() {
				_, _ = fmt.Fprintln(out)
				_, _ = fmt.Fprint(out, platform.FormatDiff(diff))
			}
		}
	}

	return nil
}

// findSettingsJSON는 프로젝트의 settings.json 경로를 찾는다.
func findSettingsJSON() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	path := filepath.Join(cwd, ".claude", "settings.json")
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

// updateSettingsPATH는 settings.json의 PATH 필드만 업데이트한다.
func updateSettingsPATH(sys platform.SystemInfo, settingsPath, newPATH string) error {
	data, err := sys.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("read settings: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("parse settings: %w", err)
	}

	// env.PATH만 업데이트
	env, ok := settings["env"].(map[string]interface{})
	if !ok {
		env = make(map[string]interface{})
		settings["env"] = env
	}

	// Windows 형식으로 PATH 저장 (세미콜론 -> 백슬래시)
	if runtime.GOOS == "windows" {
		newPATH = strings.ReplaceAll(newPATH, "/", "\\")
	}
	env["PATH"] = newPATH

	updated, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	return sys.WriteFile(settingsPath, updated, 0o644)
}

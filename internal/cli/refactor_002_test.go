package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================================
// SPEC-REFACTOR-002 검증 테스트
// cc/glm/cg 런치 커맨드 제거 리팩토링이 올바르게 완료되었는지 포괄적으로 검증한다.
// ============================================================================

// findProjectRoot는 go.mod 파일이 위치한 프로젝트 루트 디렉토리를 찾는다.
func findProjectRoot(t *testing.T) string {
	t.Helper()

	// 현재 작업 디렉토리에서 시작하여 상위로 올라가며 go.mod를 찾는다
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("작업 디렉토리를 가져올 수 없음: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod를 찾을 수 없음: 프로젝트 루트를 탐지할 수 없다")
		}
		dir = parent
	}
}

// ============================================================================
// 1. 삭제된 파일 존재 여부 검증 테스트
// 삭제된 12개 파일이 디스크에 존재하지 않음을 확인한다.
// ============================================================================

func TestRefactor002_DeletedFilesDoNotExist(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	// 삭제 대상 소스 파일 (3개)
	deletedSourceFiles := []struct {
		name string
		path string
	}{
		{"cc.go (cc 커맨드 소스)", filepath.Join("internal", "cli", "cc.go")},
		{"glm.go (glm 커맨드 소스)", filepath.Join("internal", "cli", "glm.go")},
		{"cg.go (cg 커맨드 소스)", filepath.Join("internal", "cli", "cg.go")},
	}

	// 삭제 대상 테스트 파일 (7개)
	deletedTestFiles := []struct {
		name string
		path string
	}{
		{"cc_test.go", filepath.Join("internal", "cli", "cc_test.go")},
		{"glm_test.go", filepath.Join("internal", "cli", "glm_test.go")},
		{"glm_compat_test.go", filepath.Join("internal", "cli", "glm_compat_test.go")},
		{"glm_model_override_test.go", filepath.Join("internal", "cli", "glm_model_override_test.go")},
		{"glm_new_test.go", filepath.Join("internal", "cli", "glm_new_test.go")},
		{"glm_team_test.go", filepath.Join("internal", "cli", "glm_team_test.go")},
		{"oauth_token_preservation_test.go", filepath.Join("internal", "cli", "oauth_token_preservation_test.go")},
	}

	// 삭제 대상 런처 파일 (2개)
	deletedLauncherFiles := []struct {
		name string
		path string
	}{
		{"launcher.go (런처 로직)", filepath.Join("internal", "cli", "launcher.go")},
		{"launcher_test.go (런처 테스트)", filepath.Join("internal", "cli", "launcher_test.go")},
	}

	t.Run("소스파일_삭제확인", func(t *testing.T) {
		t.Parallel()
		for _, tc := range deletedSourceFiles {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				fullPath := filepath.Join(root, tc.path)
				if _, err := os.Stat(fullPath); err == nil {
					t.Errorf("삭제되었어야 하는 파일이 존재함: %s", tc.path)
				} else if !os.IsNotExist(err) {
					t.Errorf("파일 존재 여부 확인 중 예기치 않은 오류: %s: %v", tc.path, err)
				}
			})
		}
	})

	t.Run("테스트파일_삭제확인", func(t *testing.T) {
		t.Parallel()
		for _, tc := range deletedTestFiles {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				fullPath := filepath.Join(root, tc.path)
				if _, err := os.Stat(fullPath); err == nil {
					t.Errorf("삭제되었어야 하는 테스트 파일이 존재함: %s", tc.path)
				} else if !os.IsNotExist(err) {
					t.Errorf("파일 존재 여부 확인 중 예기치 않은 오류: %s: %v", tc.path, err)
				}
			})
		}
	})

	t.Run("런처파일_삭제확인", func(t *testing.T) {
		t.Parallel()
		for _, tc := range deletedLauncherFiles {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				fullPath := filepath.Join(root, tc.path)
				if _, err := os.Stat(fullPath); err == nil {
					t.Errorf("삭제되었어야 하는 런처 파일이 존재함: %s", tc.path)
				} else if !os.IsNotExist(err) {
					t.Errorf("파일 존재 여부 확인 중 예기치 않은 오류: %s: %v", tc.path, err)
				}
			})
		}
	})
}

// ============================================================================
// 2. Root 커맨드 헬프 텍스트 검증 테스트
// 제거된 커맨드가 헬프에 표시되지 않고, 기존 커맨드 그룹은 유지됨을 확인한다.
// ============================================================================

func TestRefactor002_RootHelpText_NoLaunchCommands(t *testing.T) {
	t.Parallel()

	// UsageString()은 전역 상태를 변경하지 않으므로 병렬 안전하다
	output := rootCmd.UsageString()

	// "ae cc", "ae glm", "ae cg" 문자열이 usage에 없어야 한다
	removedRefs := []struct {
		name string
		text string
	}{
		{"ae cc 참조", "ae cc"},
		{"ae glm 참조", "ae glm"},
		{"ae cg 참조", "ae cg"},
	}

	for _, tc := range removedRefs {
		if strings.Contains(output, tc.text) {
			t.Errorf("root usage에 제거된 커맨드 참조가 포함됨: %s", tc.name)
		}
	}
}

func TestRefactor002_RootHelpText_NoLaunchCommandsGroup(t *testing.T) {
	t.Parallel()

	output := rootCmd.UsageString()

	// "Launch Commands" 그룹이 더 이상 존재하지 않아야 한다
	if strings.Contains(output, "Launch Commands") {
		t.Error("root usage에 'Launch Commands' 그룹이 남아있음: 이 그룹은 제거되었어야 한다")
	}
}

func TestRefactor002_RootHelpText_RetainsExistingGroups(t *testing.T) {
	t.Parallel()

	output := rootCmd.UsageString()

	// 기존 커맨드 그룹이 여전히 존재해야 한다
	requiredGroups := []struct {
		name  string
		label string
	}{
		{"Project Commands 그룹", "Project Commands"},
		{"Tools 그룹", "Tools"},
	}

	for _, tc := range requiredGroups {
		if !strings.Contains(output, tc.label) {
			t.Errorf("root usage에 필수 그룹이 없음: %s (검색: %q)", tc.name, tc.label)
		}
	}
}

func TestRefactor002_RootHelpText_NoGLMSpecificTerms(t *testing.T) {
	t.Parallel()

	output := strings.ToLower(rootCmd.UsageString())

	// GLM/CG 관련 용어가 usage 텍스트에 없어야 한다
	forbiddenTerms := []string{
		"glm mode",
		"claude mode",
		"cg mode",
		"launch claude",
		"switch to glm",
	}

	for _, term := range forbiddenTerms {
		if strings.Contains(output, term) {
			t.Errorf("root usage에 GLM/CG 관련 용어가 남아있음: %q", term)
		}
	}
}

// ============================================================================
// 3. 커맨드 등록 상태 검증 테스트
// cc, glm, cg가 하위 커맨드로 등록되지 않고, 기존 커맨드는 정상 등록되어 있는지 확인한다.
// ============================================================================

func TestRefactor002_RemovedCommandsNotRegistered(t *testing.T) {
	t.Parallel()

	// 제거된 커맨드 이름 목록
	removedCommands := []string{"cc", "glm", "cg"}

	registered := make(map[string]bool)
	for _, cmd := range rootCmd.Commands() {
		registered[cmd.Name()] = true
	}

	for _, name := range removedCommands {
		t.Run(name+"_미등록_확인", func(t *testing.T) {
			t.Parallel()
			if registered[name] {
				t.Errorf("제거된 커맨드 %q가 여전히 rootCmd에 등록되어 있음", name)
			}
		})
	}
}

func TestRefactor002_ExistingCommandsStillRegistered(t *testing.T) {
	t.Parallel()

	// 유지되어야 하는 핵심 커맨드 목록
	expectedCommands := []struct {
		name string
		desc string
	}{
		{"init", "프로젝트 초기화 커맨드"},
		{"doctor", "환경 진단 커맨드"},
		{"status", "상태 확인 커맨드"},
		{"version", "버전 정보 커맨드"},
		{"update", "업데이트 커맨드"},
		{"hook", "훅 관리 커맨드"},
		{"worktree", "워크트리 관리 커맨드"},
		{"statusline", "상태줄 커맨드"},
	}

	registered := make(map[string]bool)
	for _, cmd := range rootCmd.Commands() {
		registered[cmd.Name()] = true
	}

	for _, tc := range expectedCommands {
		t.Run(tc.name+"_등록_확인", func(t *testing.T) {
			t.Parallel()
			if !registered[tc.name] {
				t.Errorf("%s (%s)가 rootCmd에 등록되어 있지 않음", tc.name, tc.desc)
			}
		})
	}
}

func TestRefactor002_CommandCount_Reasonable(t *testing.T) {
	t.Parallel()

	count := len(rootCmd.Commands())

	// cc, glm, cg 3개가 제거되었으므로 최소 8개 이상이어야 한다
	// (init, doctor, status, version, update, hook, worktree, statusline + 기타)
	if count < 8 {
		t.Errorf("rootCmd에 등록된 커맨드 수가 너무 적음: %d (최소 8개 이상 예상)", count)
	}

	// cc, glm, cg가 제거되었으므로 이전보다 줄어들어야 한다
	// 기존에 launch 그룹이 3개 커맨드를 가지고 있었으므로, 지나치게 많으면 제거가 안 된 것
	if count > 30 {
		t.Errorf("rootCmd에 등록된 커맨드 수가 비정상적으로 많음: %d (제거가 누락되었을 수 있음)", count)
	}
}

// ============================================================================
// 4. 잔여 참조 검증 테스트
// internal/ 디렉토리에 제거된 함수의 잔여 참조가 없음을 확인한다.
// ============================================================================

func TestRefactor002_NoResidualFunctionReferences(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)
	internalDir := filepath.Join(root, "internal")

	// 제거된 함수명 목록 - 이 함수들은 비테스트, 비스펙 Go 파일에 존재하면 안 된다
	removedFunctions := []struct {
		name string
		desc string
	}{
		{"applyCCMode", "CC 모드 적용 함수"},
		{"applyGLMMode", "GLM 모드 적용 함수"},
		{"applyCGMode", "CG 모드 적용 함수"},
		{"unifiedLaunch", "통합 런처 함수"},
		{"resolveMode", "모드 해석 함수"},
		{"GLMConfigFromYAML", "GLM YAML 설정 파서"},
	}

	for _, fn := range removedFunctions {
		t.Run(fn.name+"_잔여참조_없음", func(t *testing.T) {
			t.Parallel()

			// grep으로 internal/ 디렉토리에서 함수명을 검색한다
			// _test.go 파일과 _spec 디렉토리는 제외한다
			cmd := exec.Command("grep", "-r", "--include=*.go",
				"--exclude=*_test.go",
				"--exclude-dir=.moai",
				fn.name, internalDir)
			output, err := cmd.CombinedOutput()

			// grep 종료 코드 1은 "매칭 없음" — 이것이 기대 결과
			if err == nil && len(output) > 0 {
				t.Errorf("제거된 함수 %q (%s)의 잔여 참조가 발견됨:\n%s",
					fn.name, fn.desc, string(output))
			}
		})
	}
}

func TestRefactor002_NoResidualGLMHookReferences(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)
	internalDir := filepath.Join(root, "internal")

	// hook/session_end.go에서 제거된 GLM 관련 함수/변수
	removedHookSymbols := []struct {
		name string
		desc string
	}{
		{"clearTmuxSessionEnv", "tmux 세션 GLM 환경변수 정리 함수"},
		{"cleanupGLMSettingsLocal", "GLM settings.local.json 정리 함수"},
		{"glmEnvVarsToClean", "GLM 환경변수 정리 목록"},
	}

	for _, sym := range removedHookSymbols {
		t.Run(sym.name+"_잔여참조_없음", func(t *testing.T) {
			t.Parallel()

			cmd := exec.Command("grep", "-r", "--include=*.go",
				"--exclude=*_test.go",
				sym.name, internalDir)
			output, err := cmd.CombinedOutput()

			if err == nil && len(output) > 0 {
				t.Errorf("제거된 hook 심볼 %q (%s)의 잔여 참조가 발견됨:\n%s",
					sym.name, sym.desc, string(output))
			}
		})
	}
}

func TestRefactor002_NoResidualLauncherReferences(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)
	internalDir := filepath.Join(root, "internal")

	// launcher.go에서 제거된 핵심 타입/함수
	launcherSymbols := []string{
		"launchConfig",
		"launcherCmd",
		"buildLaunchCmd",
	}

	for _, sym := range launcherSymbols {
		t.Run(sym+"_잔여참조_없음", func(t *testing.T) {
			t.Parallel()

			cmd := exec.Command("grep", "-r", "--include=*.go",
				"--exclude=*_test.go",
				sym, internalDir)
			output, err := cmd.CombinedOutput()

			if err == nil && len(output) > 0 {
				t.Errorf("제거된 런처 심볼 %q의 잔여 참조가 발견됨:\n%s",
					sym, string(output))
			}
		})
	}
}

// ============================================================================
// 5. moai-adk 경계 테스트
// .claude/, .moai/, CLAUDE.md가 이 리팩토링으로 수정되지 않았는지 확인한다.
// ============================================================================

func TestRefactor002_MoaiADKBoundary_NoCLAUDEMDModified(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	// git diff로 CLAUDE.md의 변경 여부를 확인한다
	cmd := exec.Command("git", "-C", root, "diff", "HEAD", "--name-only", "--", "CLAUDE.md")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("git diff 실행 실패 (CI 환경일 수 있음): %v", err)
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed != "" {
		t.Errorf("CLAUDE.md가 수정됨 — moai-adk 파일은 변경하면 안 됨:\n%s", trimmed)
	}
}

func TestRefactor002_MoaiADKBoundary_NoClaudeDirModified(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	// .claude/ 디렉토리 내 파일 변경 여부를 확인한다
	cmd := exec.Command("git", "-C", root, "diff", "HEAD", "--name-only", "--", ".claude/")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("git diff 실행 실패 (CI 환경일 수 있음): %v", err)
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed != "" {
		t.Errorf(".claude/ 디렉토리 내 파일이 수정됨 — moai-adk 파일은 변경하면 안 됨:\n%s", trimmed)
	}
}

func TestRefactor002_MoaiADKBoundary_NoMoaiConfigModified(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	// .moai/config/ 디렉토리 변경 여부를 확인한다 (config 섹션은 moai-adk 소유)
	cmd := exec.Command("git", "-C", root, "diff", "HEAD", "--name-only", "--", ".moai/config/")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("git diff 실행 실패 (CI 환경일 수 있음): %v", err)
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed != "" {
		t.Errorf(".moai/config/ 디렉토리 내 파일이 수정됨 — moai-adk 설정 파일은 변경하면 안 됨:\n%s", trimmed)
	}
}

// ============================================================================
// 6. 추가 안전성 검증 테스트
// 리팩토링이 기존 기능을 손상시키지 않았음을 추가로 확인한다.
// ============================================================================

func TestRefactor002_RootCmd_StillFunctional(t *testing.T) {
	t.Parallel()

	// rootCmd가 nil이 아니고 기본 속성이 설정되어 있어야 한다
	if rootCmd == nil {
		t.Fatal("rootCmd가 nil — 리팩토링으로 인해 루트 커맨드가 손상됨")
	}

	if rootCmd.Use != "ae" {
		t.Errorf("rootCmd.Use = %q, 기대값 %q", rootCmd.Use, "ae")
	}

	if rootCmd.Version == "" {
		t.Error("rootCmd.Version이 비어있음 — 버전 정보가 누락됨")
	}
}

func TestRefactor002_NoLaunchGroupInRootGroups(t *testing.T) {
	t.Parallel()

	groups := rootCmd.Groups()
	for _, g := range groups {
		if strings.Contains(strings.ToLower(g.Title), "launch") {
			t.Errorf("rootCmd에 launch 관련 그룹이 남아있음: ID=%q, Title=%q", g.ID, g.Title)
		}
		if g.ID == "launch" {
			t.Errorf("rootCmd에 launch ID를 가진 그룹이 남아있음: Title=%q", g.Title)
		}
	}
}

func TestRefactor002_RequiredGroupsExist(t *testing.T) {
	t.Parallel()

	groups := rootCmd.Groups()
	groupIDs := make(map[string]bool)
	for _, g := range groups {
		groupIDs[g.ID] = true
	}

	// project와 tools 그룹은 반드시 존재해야 한다
	requiredGroupIDs := []string{"project", "tools"}
	for _, id := range requiredGroupIDs {
		if !groupIDs[id] {
			t.Errorf("필수 커맨드 그룹 %q가 rootCmd.Groups()에 없음", id)
		}
	}
}

func TestRefactor002_GroupCount_NoLaunchGroup(t *testing.T) {
	t.Parallel()

	groups := rootCmd.Groups()

	// project + tools = 2개 그룹만 있어야 한다 (launch 그룹 제거됨)
	if len(groups) < 2 {
		t.Errorf("rootCmd 그룹 수가 2 미만: %d (project, tools 그룹이 필요)", len(groups))
	}

	// launch 그룹이 제거되었으므로 과도하게 많으면 안 된다
	for _, g := range groups {
		if g.ID == "launch" {
			t.Error("launch 그룹이 제거되지 않았음")
		}
	}
}

func TestRefactor002_RemovedCommandsNotResolvable(t *testing.T) {
	t.Parallel()

	// 제거된 커맨드 이름이 rootCmd에서 찾을 수 없어야 한다
	// rootCmd.Execute()는 전역 상태를 변경하므로 병렬 안전한 방법을 사용한다
	removedNames := []string{"cc", "glm", "cg"}

	for _, name := range removedNames {
		t.Run(name+"_해석불가_확인", func(t *testing.T) {
			t.Parallel()

			for _, cmd := range rootCmd.Commands() {
				if cmd.Name() == name || cmd.HasAlias(name) {
					t.Errorf("제거된 커맨드 %q가 rootCmd.Commands()에서 발견됨 (Name=%q, Aliases=%v)",
						name, cmd.Name(), cmd.Aliases)
				}
			}
		})
	}
}

func TestRefactor002_StatuslineCmd_UsesProjectFindProjectRoot(t *testing.T) {
	t.Parallel()

	// statusline.go에서 findProjectRootFn이 project.FindProjectRoot로 변경되었는지
	// 간접적으로 확인: StatuslineCmd가 정상 등록되어 실행 가능해야 한다
	if StatuslineCmd == nil {
		t.Fatal("StatuslineCmd가 nil — statusline 커맨드가 손상됨")
	}

	if StatuslineCmd.RunE == nil {
		t.Error("StatuslineCmd.RunE가 nil — 실행 로직이 누락됨")
	}
}

func TestRefactor002_HelpOutput_DoesNotMentionLauncher(t *testing.T) {
	t.Parallel()

	// UsageString()은 rootCmd 상태를 변경하지 않으므로 병렬 안전하다
	output := strings.ToLower(rootCmd.UsageString())

	// "launcher" 관련 용어가 헬프에 없어야 한다
	if strings.Contains(output, "launcher") {
		t.Error("root usage에 'launcher' 관련 텍스트가 남아있음")
	}
}

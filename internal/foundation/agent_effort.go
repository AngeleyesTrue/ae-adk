package foundation

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/AngeleyesTrue/ae-adk/internal/manifest"
	"gopkg.in/yaml.v3"
)

// normalizeEffortValue normalizes a user-provided effort string per A-09.5:
//   - trims surrounding whitespace
//   - strips surrounding double or single quotes
//   - lowercases
//
// Returns the canonical lowercase value if it matches one of the standard
// 5 levels (low/medium/high/xhigh/max), or "" for any non-standard input.
// "max" is included as a forward-compatible alias even when not yet in the
// active EffortLevel set.
func normalizeEffortValue(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"'`)
	s = strings.ToLower(s)
	switch s {
	case "low", "medium", "high", "xhigh", "max":
		return s
	}
	return ""
}

// agentEffortMap은 추론 집약 에이전트와 해당 effort 레벨을 매핑한다.
// v1.2.0: plan-auditor 제외 (templates에 파일 미존재, REQ-24 BLOCKER 해소)
// @MX:ANCHOR: [AUTO] GetAgentEffort, ApplyEffortPolicy 등 3개 함수에서 참조
// @MX:REASON: 에이전트 effort 정책의 단일 진실 출처 — 변경 시 모든 에이전트 frontmatter 주입 결과에 영향
var agentEffortMap = map[string]EffortLevel{
	"manager-spec":      EffortXHigh,
	"manager-strategy":  EffortXHigh,
	"evaluator-active":  EffortHigh,
	"expert-security":   EffortHigh,
	"expert-refactoring": EffortHigh,
	"builder-agent":     EffortHigh,
}

// GetAgentEffort는 에이전트 이름에 대한 effort 레벨을 반환한다.
// 매핑에 없는 에이전트는 (기본값, false)를 반환한다.
func GetAgentEffort(agentName string) (EffortLevel, bool) {
	level, ok := agentEffortMap[agentName]
	return level, ok
}

// agentFrontmatter는 YAML frontmatter 파싱에 사용하는 구조체이다.
type agentFrontmatter struct {
	Effort string `yaml:"effort"`
}

// ApplyEffortPolicy는 projectRoot 아래 .claude/agents/ae/ 디렉토리에서
// agentEffortMap에 등록된 에이전트 .md 파일을 찾아 effort: 필드를 주입한다.
//
// 동작 규칙:
//  - frontmatter에 이미 effort: 필드가 있으면 보존 (사용자 커스텀 값 존중, A-01)
//  - effort: 필드가 없으면 agentEffortMap 값을 주입
//  - 파일 내용이 변경된 경우에만 디스크에 쓴다
//  - 변경이 발생하면 manifestMgr 해시를 갱신한다 (idempotency 보장)
//
// @MX:TODO: [AUTO] manifestMgr.Track 호출 시 templateHash 인자 개선 필요 (현재 빈 문자열)
func ApplyEffortPolicy(projectRoot string, manifestMgr manifest.Manager) error {
	agentsDir := filepath.Join(projectRoot, ".claude", "agents", "ae")

	info, err := os.Stat(agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			// 에이전트 디렉토리가 없으면 조용히 종료 (초기 init 전 상황)
			return nil
		}
		return fmt.Errorf("effort policy: stat agents dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("effort policy: %s is not a directory", agentsDir)
	}

	modifiedCount := 0

	for agentName, targetLevel := range agentEffortMap {
		agentFile := filepath.Join(agentsDir, agentName+".md")

		data, err := os.ReadFile(agentFile)
		if err != nil {
			if os.IsNotExist(err) {
				// 에이전트 파일 미존재는 건너뜀 (REQ-24: plan-auditor 등)
				continue
			}
			return fmt.Errorf("effort policy: read %s: %w", agentName, err)
		}

		original := string(data)
		updated, changed, err := injectEffortIntoFrontmatter(original, targetLevel)
		if err != nil {
			return fmt.Errorf("effort policy: inject %s: %w", agentName, err)
		}

		if !changed {
			continue
		}

		if err := os.WriteFile(agentFile, []byte(updated), 0o644); err != nil {
			return fmt.Errorf("effort policy: write %s: %w", agentName, err)
		}

		// 매니페스트 해시 갱신
		relPath := filepath.Join(".claude", "agents", "ae", agentName+".md")
		if err := manifestMgr.Track(relPath, manifest.TemplateManaged, ""); err != nil {
			// 매니페스트 갱신 실패는 경고로만 처리 (비필수)
			_ = err
		}

		modifiedCount++
	}

	_ = modifiedCount
	return nil
}

// injectEffortIntoFrontmatter는 마크다운 frontmatter에서 effort: 필드를 처리한다.
//
// - effort: 필드가 이미 존재하면 값을 보존하고 changed=false 반환
// - effort: 필드가 없으면 주입하고 changed=true 반환
// - frontmatter가 없는 파일(--- 없음)은 변경 없이 반환
func injectEffortIntoFrontmatter(content string, level EffortLevel) (updated string, changed bool, err error) {
	// frontmatter 경계 감지 (--- 로 시작하는 파일)
	if !strings.HasPrefix(strings.TrimSpace(content), "---") {
		return content, false, nil
	}

	// 첫 번째 --- 이후 두 번째 --- 위치 찾기
	rest := content
	firstDelim := strings.Index(rest, "---")
	if firstDelim < 0 {
		return content, false, nil
	}
	afterFirst := rest[firstDelim+3:]

	secondDelim := strings.Index(afterFirst, "\n---")
	if secondDelim < 0 {
		return content, false, nil
	}

	frontmatterContent := afterFirst[:secondDelim]
	afterFrontmatter := afterFirst[secondDelim+4:] // "\n---" 이후

	// 이미 effort: 필드가 있는지 확인
	var fm agentFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatterContent), &fm); err != nil {
		// 파싱 실패 시 원본 유지
		return content, false, nil
	}

	if fm.Effort != "" {
		// 사용자 커스텀 값 보존 (A-01) + A-09.5 정규화 검사
		// 표준 5개 값(low/medium/high/xhigh/max) 또는 정규화 가능(따옴표/대문자)이면
		// 사용자 명시 의도로 간주하여 그대로 보존한다. 비표준 입력은 경고만 발생시키고
		// 기존 값을 유지한다 — 정책값 강제 덮어쓰기는 사용자 의도를 침해할 수 있어 보수적
		// 처리.
		if normalizeEffortValue(fm.Effort) == "" {
			slog.Warn("non-standard effort value detected in agent frontmatter; preserved as-is",
				"effort", fm.Effort,
				"policy_default", level.String(),
			)
		}
		return content, false, nil
	}

	// effort: 필드 주입
	injectedLine := fmt.Sprintf("\neffort: %s", level.String())
	newFrontmatter := frontmatterContent + injectedLine

	newContent := "---" + newFrontmatter + "\n---" + afterFrontmatter
	return newContent, true, nil
}

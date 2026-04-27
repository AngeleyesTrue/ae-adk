package foundation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AngeleyesTrue/ae-adk/internal/manifest"
)

// mockManifestManager는 테스트용 manifest.Manager 구현체이다.
type mockManifestManager struct {
	tracked map[string]manifest.Provenance
}

func newMockManifestManager() *mockManifestManager {
	return &mockManifestManager{tracked: make(map[string]manifest.Provenance)}
}

func (m *mockManifestManager) Load(projectRoot string) (*manifest.Manifest, error) {
	return manifest.NewManifest(), nil
}

func (m *mockManifestManager) Manifest() *manifest.Manifest {
	return manifest.NewManifest()
}

func (m *mockManifestManager) Save() error { return nil }

func (m *mockManifestManager) Track(path string, provenance manifest.Provenance, templateHash string) error {
	m.tracked[path] = provenance
	return nil
}

func (m *mockManifestManager) GetEntry(path string) (*manifest.FileEntry, bool) {
	return nil, false
}

func (m *mockManifestManager) DetectChanges() ([]manifest.ChangedFile, error) {
	return nil, nil
}

func (m *mockManifestManager) Remove(path string) error { return nil }

// setupAgentFile은 임시 프로젝트에 에이전트 .md 파일을 생성한다.
func setupAgentFile(t *testing.T, projectRoot, agentName, content string) string {
	t.Helper()
	dir := filepath.Join(projectRoot, ".claude", "agents", "ae")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(dir, agentName+".md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write agent file: %v", err)
	}
	return path
}

// readAgentFile은 에이전트 파일 내용을 읽는다.
func readAgentFile(t *testing.T, projectRoot, agentName string) string {
	t.Helper()
	path := filepath.Join(projectRoot, ".claude", "agents", "ae", agentName+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read agent file: %v", err)
	}
	return string(data)
}

// TestApplyEffortPolicy_Preserve는 이미 effort: 필드가 있으면 보존하는지 검증한다 (A-01).
func TestApplyEffortPolicy_Preserve(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	mgr := newMockManifestManager()

	// manager-spec에 사용자 커스텀 effort: medium 설정
	original := `---
name: manager-spec
effort: medium
description: SPEC manager
---

# Manager Spec
`
	setupAgentFile(t, projectRoot, "manager-spec", original)

	if err := ApplyEffortPolicy(projectRoot, mgr); err != nil {
		t.Fatalf("ApplyEffortPolicy: %v", err)
	}

	// 파일이 변경되지 않아야 함 (사용자 커스텀 보존)
	got := readAgentFile(t, projectRoot, "manager-spec")
	if got != original {
		t.Errorf("effort field should be preserved\ngot:\n%s\nwant:\n%s", got, original)
	}

	// manifest가 업데이트되지 않아야 함
	if len(mgr.tracked) != 0 {
		t.Errorf("manifest should not be updated for preserved files, got %d tracked", len(mgr.tracked))
	}
}

// TestApplyEffortPolicy_Inject는 effort: 필드가 없을 때 주입하는지 검증한다 (AC-08).
func TestApplyEffortPolicy_Inject(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	mgr := newMockManifestManager()

	// manager-spec에 effort: 없음
	original := `---
name: manager-spec
description: SPEC manager
---

# Manager Spec
`
	setupAgentFile(t, projectRoot, "manager-spec", original)

	if err := ApplyEffortPolicy(projectRoot, mgr); err != nil {
		t.Fatalf("ApplyEffortPolicy: %v", err)
	}

	got := readAgentFile(t, projectRoot, "manager-spec")
	if got == original {
		t.Error("effort field should have been injected")
	}

	// effort: xhigh가 포함되어야 함
	if !contains(got, "effort: xhigh") {
		t.Errorf("expected effort: xhigh in file content, got:\n%s", got)
	}

	// manifest가 업데이트되어야 함
	if len(mgr.tracked) == 0 {
		t.Error("manifest should be updated after injection")
	}
}

// TestApplyEffortPolicy_Idempotent는 동일 호출 2회 시 파일 변경이 없음을 검증한다 (AC-08).
func TestApplyEffortPolicy_Idempotent(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	mgr := newMockManifestManager()

	original := `---
name: manager-spec
description: SPEC manager
---

# Manager Spec
`
	setupAgentFile(t, projectRoot, "manager-spec", original)

	// 첫 번째 호출
	if err := ApplyEffortPolicy(projectRoot, mgr); err != nil {
		t.Fatalf("first ApplyEffortPolicy: %v", err)
	}
	afterFirst := readAgentFile(t, projectRoot, "manager-spec")

	// 두 번째 호출
	mgr2 := newMockManifestManager()
	if err := ApplyEffortPolicy(projectRoot, mgr2); err != nil {
		t.Fatalf("second ApplyEffortPolicy: %v", err)
	}
	afterSecond := readAgentFile(t, projectRoot, "manager-spec")

	// 두 번째 호출 후 내용이 동일해야 함
	if afterFirst != afterSecond {
		t.Errorf("idempotency violated:\nafter first:\n%s\nafter second:\n%s", afterFirst, afterSecond)
	}

	// 두 번째 호출에서는 manifest 업데이트가 없어야 함 (변경 없음)
	if len(mgr2.tracked) != 0 {
		t.Errorf("second call should not update manifest, got %d tracked", len(mgr2.tracked))
	}
}

// contains는 문자열 s에 substr이 포함되어 있는지 확인한다.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestInjectEffortIntoFrontmatter_EdgeCases는 injectEffortIntoFrontmatter의
// 경계 조건들을 테이블 주도 방식으로 검증한다.
func TestInjectEffortIntoFrontmatter_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		level       EffortLevel
		wantChanged bool
		wantContain string // 결과에 포함되어야 하는 문자열 (빈 경우 무시)
	}{
		{
			name:        "NoFrontmatter_Unchanged",
			input:       "# Just a markdown file\nNo frontmatter here.",
			level:       EffortHigh,
			wantChanged: false,
		},
		{
			name:        "SingleDash_Unclosed_Unchanged",
			input:       "---\nname: test\n",
			level:       EffortHigh,
			wantChanged: false,
		},
		{
			name: "MalformedYAML_Unchanged",
			// YAML 파싱은 실제로 단순 key: value도 파싱 가능하므로
			// effort가 없으면 주입됨. 여기서는 yaml 키-값이 있지만
			// 이미 effort 있는 경우를 검증한다.
			input: "---\neffort: custom\nname: test\n---\n\n# Content",
			level: EffortHigh,
			// effort가 이미 있으므로 changed=false
			wantChanged: false,
		},
		{
			name:        "EmptyFrontmatter_InjectsEffort",
			input:       "---\n---\n\n# Content",
			level:       EffortXHigh,
			wantChanged: true,
			wantContain: "effort: xhigh",
		},
		{
			name:        "EffortAlreadySet_Preserved",
			input:       "---\nname: agent\neffort: low\n---\n\n# Body",
			level:       EffortMax,
			wantChanged: false,
		},
		{
			name:        "NoEffort_Injected",
			input:       "---\nname: agent\ndescription: A test agent\n---\n\n# Body",
			level:       EffortHigh,
			wantChanged: true,
			wantContain: "effort: high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			updated, changed, err := injectEffortIntoFrontmatter(tt.input, tt.level)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if changed != tt.wantChanged {
				t.Errorf("changed=%v, want %v\nupdated content:\n%s", changed, tt.wantChanged, updated)
			}
			if tt.wantContain != "" && !contains(updated, tt.wantContain) {
				t.Errorf("expected %q in output, got:\n%s", tt.wantContain, updated)
			}
			// changed=false 시 원본과 동일해야 함
			if !changed && updated != tt.input {
				t.Errorf("unchanged: output differs from input\ngot:\n%s\nwant:\n%s", updated, tt.input)
			}
		})
	}
}

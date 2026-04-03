package convention

import (
	"testing"
)

// TestIntegration_InitToValidation은 전체 파이프라인을 관통하는 통합 테스트:
// bracket-scope 컨벤션 로드 -> 메시지 검증 체인 (T-9)
func TestIntegration_InitToValidation(t *testing.T) {
	// Step 1: Manager를 통한 bracket-scope 로드
	m := NewManager(t.TempDir())
	if err := m.LoadConvention("bracket-scope"); err != nil {
		t.Fatalf("LoadConvention(bracket-scope): %v", err)
	}

	conv := m.Convention()
	if conv == nil {
		t.Fatal("Convention() is nil after loading bracket-scope")
	}
	if conv.Name != "bracket-scope" {
		t.Errorf("Convention().Name = %q, want %q", conv.Name, "bracket-scope")
	}
	if conv.ScopeDelimiter != "[]" {
		t.Errorf("ScopeDelimiter = %q, want %q", conv.ScopeDelimiter, "[]")
	}

	// Step 2: bracket-scope 형식 메시지 Valid 확인
	validMessages := []struct {
		name string
		msg  string
	}{
		{"single scope", "feat: [Web] add new feature"},
		{"multi scope", "fix: [Auth/Session] resolve token issue"},
		{"3-level scope", "refactor: [Core/DB/Query] optimize performance"},
		{"breaking change", "feat!: [API] change endpoint signature"},
		{"no scope", "chore: update dependencies"},
		{"docs with scope", "docs: [Build] update CI guide"},
	}

	for _, tt := range validMessages {
		t.Run("valid_"+tt.name, func(t *testing.T) {
			result := m.ValidateMessage(tt.msg)
			if !result.Valid {
				t.Errorf("ValidateMessage(%q) should be valid, violations: %v", tt.msg, result.Violations)
			}
		})
	}

	// Step 3: 기존 format 메시지 Invalid 확인 (bracket-scope에서 parenthesis는 무효)
	invalidMessages := []struct {
		name string
		msg  string
	}{
		{"old paren format", "feat(auth): old format description"},
		{"missing space after bracket", "feat: [Web]no space"},
		{"empty brackets", "feat: [] empty scope"},
		{"numeric scope start", "feat: [123] numeric"},
	}

	for _, tt := range invalidMessages {
		t.Run("invalid_"+tt.name, func(t *testing.T) {
			result := m.ValidateMessage(tt.msg)
			if result.Valid {
				t.Errorf("ValidateMessage(%q) should be invalid for bracket-scope", tt.msg)
			}
		})
	}

	// Step 4: Auto fallback이 bracket-scope인지 확인
	autoMgr := NewManager(t.TempDir())
	if err := autoMgr.LoadConvention("auto"); err != nil {
		t.Fatalf("LoadConvention(auto): %v", err)
	}
	if autoMgr.Convention().Name != "bracket-scope" {
		t.Errorf("Auto fallback convention = %q, want %q", autoMgr.Convention().Name, "bracket-scope")
	}
}

// TestIntegration_ConventionCoexistence는 bracket-scope와 기존 컨벤션이 공존하는지 검증한다.
func TestIntegration_ConventionCoexistence(t *testing.T) {
	conventions := []string{"bracket-scope", "conventional-commits", "angular", "karma"}

	for _, name := range conventions {
		t.Run(name, func(t *testing.T) {
			m := NewManager("/unused")
			if err := m.LoadConvention(name); err != nil {
				t.Fatalf("LoadConvention(%q): %v", name, err)
			}
			conv := m.Convention()
			if conv == nil {
				t.Fatal("Convention() is nil")
			}
			if conv.Name != name {
				t.Errorf("Name = %q, want %q", conv.Name, name)
			}

			// 각 컨벤션의 예제가 자신의 패턴에 매치되는지 확인
			for _, example := range conv.Examples {
				result := Validate(example, conv)
				if !result.Valid {
					t.Errorf("Example %q should be valid for %s, violations: %v",
						example, name, result.Violations)
				}
			}
		})
	}
}

// TestIntegration_ScopeExtractionByConvention은 컨벤션별 scope 추출이 올바른지 검증한다.
func TestIntegration_ScopeExtractionByConvention(t *testing.T) {
	tests := []struct {
		convention string
		message    string
		wantScope  string
	}{
		{"bracket-scope", "feat: [Web] add feature", "Web"},
		{"bracket-scope", "feat: [Web/Auth] multi scope", "Web/Auth"},
		{"bracket-scope", "feat: no scope msg", ""},
		{"conventional-commits", "feat(auth): add feature", "auth"},
		{"conventional-commits", "feat: no scope", ""},
	}

	for _, tt := range tests {
		t.Run(tt.convention+"_"+tt.message, func(t *testing.T) {
			conv, err := ParseBuiltin(tt.convention)
			if err != nil {
				t.Fatalf("ParseBuiltin(%q): %v", tt.convention, err)
			}
			scope := extractScope(tt.message, conv.ScopeDelimiter)
			if scope != tt.wantScope {
				t.Errorf("extractScope(%q, %q) = %q, want %q",
					tt.message, conv.ScopeDelimiter, scope, tt.wantScope)
			}
		})
	}
}

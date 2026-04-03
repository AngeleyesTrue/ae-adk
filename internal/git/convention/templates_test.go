package convention

import (
	"regexp"
	"sort"
	"testing"
)

func TestBuiltinNames(t *testing.T) {
	names := BuiltinNames()

	if len(names) != 4 {
		t.Fatalf("BuiltinNames() length = %d, want 4", len(names))
	}

	expected := []string{"angular", "bracket-scope", "conventional-commits", "karma"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("BuiltinNames()[%d] = %q, want %q", i, name, expected[i])
		}
	}

	// T-5b: 정렬 순서 검증
	if !sort.StringsAreSorted(names) {
		t.Error("BuiltinNames() should return sorted slice")
	}
}

func TestGetBuiltin(t *testing.T) {
	tests := []struct {
		name     string
		wantName string
		wantNil  bool
	}{
		{name: "conventional-commits", wantName: "conventional-commits"},
		{name: "angular", wantName: "angular"},
		{name: "karma", wantName: "karma"},
		{name: "bracket-scope", wantName: "bracket-scope"},
		{name: "nonexistent", wantNil: true},
		{name: "", wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GetBuiltin(tt.name)
			if tt.wantNil {
				if cfg != nil {
					t.Errorf("GetBuiltin(%q) = %v, want nil", tt.name, cfg)
				}
				return
			}
			if cfg == nil {
				t.Fatalf("GetBuiltin(%q) = nil, want non-nil", tt.name)
			}
			if cfg.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", cfg.Name, tt.wantName)
			}
		})
	}
}

func TestBuiltinPatternsCompile(t *testing.T) {
	for name, cfg := range builtinConventions {
		t.Run(name, func(t *testing.T) {
			if cfg.Pattern == "" {
				t.Error("Pattern is empty")
			}
			_, err := regexp.Compile(cfg.Pattern)
			if err != nil {
				t.Errorf("Pattern %q failed to compile: %v", cfg.Pattern, err)
			}
		})
	}
}

func TestBuiltinConventionsHaveExamples(t *testing.T) {
	for name, cfg := range builtinConventions {
		t.Run(name, func(t *testing.T) {
			if len(cfg.Examples) == 0 {
				t.Error("convention has no examples")
			}
		})
	}
}

func TestBuiltinConventionsHaveTypes(t *testing.T) {
	for name, cfg := range builtinConventions {
		t.Run(name, func(t *testing.T) {
			if len(cfg.Types) == 0 {
				t.Error("convention has no types")
			}
		})
	}
}

func TestBuiltinExamplesMatchPatterns(t *testing.T) {
	for name, cfg := range builtinConventions {
		t.Run(name, func(t *testing.T) {
			re, err := regexp.Compile(cfg.Pattern)
			if err != nil {
				t.Fatalf("compile: %v", err)
			}
			for _, example := range cfg.Examples {
				if !re.MatchString(example) {
					t.Errorf("example %q does not match pattern %q", example, cfg.Pattern)
				}
			}
		})
	}
}

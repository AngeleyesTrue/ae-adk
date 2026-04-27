package foundation

import "testing"

// TestEffortConstants는 5단계 EffortLevel 상수와 ModelIDOpus47 값을 검증한다 (AC-06).
func TestEffortConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level EffortLevel
		want  string
	}{
		{"low", EffortLow, "low"},
		{"medium", EffortMedium, "medium"},
		{"high", EffortHigh, "high"},
		{"xhigh", EffortXHigh, "xhigh"},
		{"max", EffortMax, "max"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.level.String(); got != tt.want {
				t.Errorf("EffortLevel.String() = %q, want %q", got, tt.want)
			}
			if string(tt.level) != tt.want {
				t.Errorf("string(EffortLevel) = %q, want %q", string(tt.level), tt.want)
			}
		})
	}

	if ModelIDOpus47 != "claude-opus-4-7" {
		t.Errorf("ModelIDOpus47 = %q, want %q", ModelIDOpus47, "claude-opus-4-7")
	}
}

// TestGetAgentEffort는 agentEffortMap의 6개 항목과 미지 에이전트 케이스를 검증한다 (AC-07).
func TestGetAgentEffort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		agentName string
		wantLevel EffortLevel
		wantOK    bool
	}{
		{"manager-spec", EffortXHigh, true},
		{"manager-strategy", EffortXHigh, true},
		{"evaluator-active", EffortHigh, true},
		{"expert-security", EffortHigh, true},
		{"expert-refactoring", EffortHigh, true},
		{"builder-agent", EffortHigh, true},
		// 미등록 에이전트는 false 반환
		{"unknown-agent", "", false},
		{"plan-auditor", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.agentName, func(t *testing.T) {
			t.Parallel()
			got, ok := GetAgentEffort(tt.agentName)
			if ok != tt.wantOK {
				t.Errorf("GetAgentEffort(%q) ok = %v, want %v", tt.agentName, ok, tt.wantOK)
			}
			if tt.wantOK && got != tt.wantLevel {
				t.Errorf("GetAgentEffort(%q) level = %q, want %q", tt.agentName, got, tt.wantLevel)
			}
		})
	}
}

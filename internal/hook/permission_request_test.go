package hook

import (
	"context"
	"encoding/json"
	"testing"
)

// mustMarshal은 테스트 헬퍼로 json.Marshal 실패 시 패닉한다.
func mustMarshal(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func TestPermissionRequestHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewPermissionRequestHandler()

	if got := h.EventType(); got != EventPermissionRequest {
		t.Errorf("EventType() = %q, want %q", got, EventPermissionRequest)
	}
}

func TestPermissionRequestHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "permission for Bash tool",
			input: &HookInput{
				SessionID:     "sess-perm-1",
				ToolName:      "Bash",
				HookEventName: "PermissionRequest",
			},
		},
		{
			name: "permission for Write tool",
			input: &HookInput{
				SessionID:     "sess-perm-2",
				ToolName:      "Write",
				HookEventName: "PermissionRequest",
			},
		},
		{
			name: "permission without tool name",
			input: &HookInput{
				SessionID: "sess-perm-3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewPermissionRequestHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			if got.HookSpecificOutput == nil {
				t.Fatal("PermissionRequest hook should set hookSpecificOutput")
			}
			if got.HookSpecificOutput.PermissionDecision != DecisionAsk {
				t.Errorf("PermissionDecision = %q, want %q", got.HookSpecificOutput.PermissionDecision, DecisionAsk)
			}
			if got.HookSpecificOutput.HookEventName != "PermissionRequest" {
				t.Errorf("HookEventName = %q, want %q", got.HookSpecificOutput.HookEventName, "PermissionRequest")
			}
		})
	}
}

// TestPermissionRequestHandler_MarkerSentinel verifies REQ-16:
// the handler rejects requests that contain the "__updated_input_marker__" sentinel key.
func TestPermissionRequestHandler_MarkerSentinel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		toolInput      json.RawMessage
		wantDecision   string
		wantSentinel   bool // true → deny expected
	}{
		{
			name: "positive: sentinel key present → deny",
			toolInput: mustMarshal(map[string]any{
				"__updated_input_marker__": true,
				"command":                  "ls",
			}),
			wantDecision: DecisionDeny,
			wantSentinel: true,
		},
		{
			name: "positive: sentinel key with string value → deny",
			toolInput: mustMarshal(map[string]any{
				"__updated_input_marker__": "1",
			}),
			wantDecision: DecisionDeny,
			wantSentinel: true,
		},
		{
			name: "negative: normal tool input without sentinel → ask",
			toolInput: mustMarshal(map[string]any{
				"command": "ls -la",
			}),
			wantDecision: DecisionAsk,
			wantSentinel: false,
		},
		{
			name:         "negative: nil tool input → ask",
			toolInput:    nil,
			wantDecision: DecisionAsk,
			wantSentinel: false,
		},
		{
			name:         "negative: empty tool input → ask",
			toolInput:    json.RawMessage{},
			wantDecision: DecisionAsk,
			wantSentinel: false,
		},
		{
			name:         "negative: non-object JSON → ask (safe fallback)",
			toolInput:    json.RawMessage(`"just a string"`),
			wantDecision: DecisionAsk,
			wantSentinel: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewPermissionRequestHandler()
			input := &HookInput{
				SessionID:     "sess-marker-test",
				ToolName:      "Bash",
				HookEventName: "PermissionRequest",
				ToolInput:     tt.toolInput,
			}

			got, err := h.Handle(context.Background(), input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			if got.HookSpecificOutput == nil {
				t.Fatal("HookSpecificOutput must not be nil")
			}
			if got.HookSpecificOutput.PermissionDecision != tt.wantDecision {
				t.Errorf("PermissionDecision = %q, want %q",
					got.HookSpecificOutput.PermissionDecision, tt.wantDecision)
			}
			if got.HookSpecificOutput.HookEventName != "PermissionRequest" {
				t.Errorf("HookEventName = %q, want \"PermissionRequest\"",
					got.HookSpecificOutput.HookEventName)
			}
		})
	}
}

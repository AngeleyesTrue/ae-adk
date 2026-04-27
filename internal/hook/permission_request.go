package hook

import (
	"context"
	"encoding/json"
	"log/slog"
)

// updatedInputMarkerSentinel은 Claude Code가 툴 입력을 갱신할 때 사용하는
// 내부 마커 키이다 (REQ-16). 이 키가 updatedInput에 존재하면 재검증 후 거부한다.
const updatedInputMarkerSentinel = "__updated_input_marker__"

// permissionRequestHandler processes PermissionRequest events.
// It logs permission requests and defers to the default decision.
// If the hook input's UpdatedInput JSON contains the sentinel key
// "__updated_input_marker__", the handler re-validates and returns a rejection (REQ-16).
type permissionRequestHandler struct{}

// NewPermissionRequestHandler creates a new PermissionRequest event handler.
func NewPermissionRequestHandler() Handler {
	return &permissionRequestHandler{}
}

// EventType returns EventPermissionRequest.
func (h *permissionRequestHandler) EventType() EventType {
	return EventPermissionRequest
}

// Handle processes a PermissionRequest event. It logs the permission request.
// When updatedInput contains the "__updated_input_marker__" sentinel key,
// the handler re-validates the input and returns a deny decision (REQ-16).
// Otherwise it returns "ask" (defer to user/default settings).
func (h *permissionRequestHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("permission requested",
		"session_id", input.SessionID,
		"tool_name", input.ToolName,
	)

	// REQ-16: Detect and reject the updated-input marker sentinel.
	if h.hasMarkerSentinel(input) {
		slog.Warn("permission request contains updated-input marker sentinel — rejecting",
			"session_id", input.SessionID,
			"tool_name", input.ToolName,
		)
		return NewPermissionRequestOutput(DecisionDeny, "input marker sentinel detected; request rejected"), nil
	}

	// Default to "ask" - defer decision to user/settings.
	// Per Claude Code protocol (v2.1.59+), hookSpecificOutput.hookEventName must be
	// "PermissionRequest" for PermissionRequest events.
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:      "PermissionRequest",
			PermissionDecision: DecisionAsk,
		},
	}, nil
}

// hasMarkerSentinel returns true when input.HookSpecificOutput.UpdatedInput JSON
// contains the "__updated_input_marker__" key at the top level (REQ-16).
func (h *permissionRequestHandler) hasMarkerSentinel(input *HookInput) bool {
	// updatedInput arrives via HookInput.ToolInput for PermissionRequest events.
	raw := input.ToolInput
	if len(raw) == 0 {
		return false
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return false
	}
	_, found := m[updatedInputMarkerSentinel]
	return found
}

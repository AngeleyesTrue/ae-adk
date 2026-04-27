package hook

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/AngeleyesTrue/ae-adk/internal/config"
)

// sessionStartHandler processes SessionStart events.
// It initializes the session, loads project configuration, and validates
// the execution environment (REQ-HOOK-030).
type sessionStartHandler struct {
	cfg ConfigProvider
}

// NewSessionStartHandler creates a new SessionStart event handler.
func NewSessionStartHandler(cfg ConfigProvider) Handler {
	return &sessionStartHandler{cfg: cfg}
}

// EventType returns EventSessionStart.
func (h *sessionStartHandler) EventType() EventType {
	return EventSessionStart
}

// Handle processes a SessionStart event. It logs the session ID, loads
// project configuration, and returns project information in the Data field.
// Errors are non-blocking: the handler logs warnings and returns allow.
func (h *sessionStartHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("session started",
		"session_id", input.SessionID,
		"cwd", input.CWD,
		"project_dir", input.ProjectDir,
	)

	// REQ-17: Windows 전용 — 프로젝트 루트에 .env가 있으면
	// .claude/settings.local.json의 env 맵에 CLAUDE_ENV_FILE을 주입한다.
	// non-Windows 빌드에서는 build tag로 분리된 no-op 스텁이 호출된다.
	// 비차단 정책: 주입 실패는 로깅만 하고 세션 시작을 계속 진행한다.
	if input.ProjectDir != "" {
		if err := injectCLAUDEEnvFile(input.ProjectDir); err != nil {
			slog.Warn("CLAUDE_ENV_FILE injection failed (non-fatal)",
				"session_id", input.SessionID,
				"project_dir", input.ProjectDir,
				"error", err.Error(),
			)
		}
	}

	data := map[string]any{
		"session_id": input.SessionID,
		"status":     "initialized",
	}

	// Load project information from config if available
	cfg := h.getConfig()
	if cfg != nil {
		if cfg.Project.Name != "" {
			data["project_name"] = cfg.Project.Name
		}
		if string(cfg.Project.Type) != "" {
			data["project_type"] = string(cfg.Project.Type)
		}
		if cfg.Project.Language != "" {
			data["project_language"] = cfg.Project.Language
		}
	} else {
		slog.Warn("configuration not available, proceeding with defaults",
			"session_id", input.SessionID,
		)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal session data",
			"error", err.Error(),
		)
		return &HookOutput{}, nil
	}

	return &HookOutput{Data: jsonData}, nil
}

// getConfig safely retrieves the configuration, returning nil if unavailable.
func (h *sessionStartHandler) getConfig() *config.Config {
	if h.cfg == nil {
		return nil
	}
	return h.cfg.Get()
}

//go:build windows

package hook

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// injectCLAUDEEnvFile는 Windows 전용 구현이다 (REQ-17).
//
// 프로젝트 루트에 .env 파일이 존재하면 .claude/settings.local.json의
// env 맵에 CLAUDE_ENV_FILE=<절대경로> 를 병합한다.
//
// 동작 규칙:
//   - macOS/Linux: 이 파일은 빌드에서 제외되므로 no-op
//   - 경로는 filepath.ToSlash + JSON-safe 인코딩 처리
//   - 쓰기는 원자적으로 수행 (임시 파일 생성 후 os.Rename)
//   - 이미 동일 값이 존재하면 파일을 수정하지 않음 (idempotent)
func injectCLAUDEEnvFile(projectRoot string) error {
	if projectRoot == "" {
		return nil
	}

	envFilePath := filepath.Join(projectRoot, ".env")
	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		return nil // .env가 없으면 no-op
	}

	// JSON-safe 슬래시 경로로 변환한다.
	absEnvPath := filepath.ToSlash(envFilePath)

	settingsPath := filepath.Join(projectRoot, ".claude", "settings.local.json")

	// 기존 settings.local.json을 읽거나 빈 구조체로 초기화한다.
	raw, err := os.ReadFile(settingsPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read settings.local.json: %w", err)
	}

	var settings map[string]json.RawMessage
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &settings); err != nil {
			return fmt.Errorf("parse settings.local.json: %w", err)
		}
	}
	if settings == nil {
		settings = make(map[string]json.RawMessage)
	}

	// env 맵을 읽거나 새로 생성한다.
	var envMap map[string]string
	if envRaw, ok := settings["env"]; ok {
		if err := json.Unmarshal(envRaw, &envMap); err != nil {
			return fmt.Errorf("parse env section: %w", err)
		}
	}
	if envMap == nil {
		envMap = make(map[string]string)
	}

	// 이미 동일 값이 있으면 파일을 건드리지 않는다 (idempotent).
	if envMap["CLAUDE_ENV_FILE"] == absEnvPath {
		return nil
	}

	envMap["CLAUDE_ENV_FILE"] = absEnvPath

	// env 맵을 다시 직렬화한다.
	envBytes, err := json.Marshal(envMap)
	if err != nil {
		return fmt.Errorf("marshal env map: %w", err)
	}
	settings["env"] = envBytes

	// 전체 settings 파일을 직렬화한다 (들여쓰기 포함).
	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	// 원자적 쓰기: os.CreateTemp (O_EXCL + crypto-rand 이름) → os.Rename
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	tmpFile, err := os.CreateTemp(dir, "settings-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmp := tmpFile.Name()
	defer func() {
		if err != nil {
			_ = os.Remove(tmp)
		}
	}()
	if _, err = tmpFile.Write(out); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err = tmpFile.Chmod(0o600); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err = os.Rename(tmp, settingsPath); err != nil {
		return fmt.Errorf("rename temp to settings.local.json: %w", err)
	}

	slog.Info("injected CLAUDE_ENV_FILE into settings.local.json",
		"path", absEnvPath,
	)
	return nil
}

// claudeEnvFilePath는 현재 settings.local.json에서 CLAUDE_ENV_FILE 값을 반환한다.
// 파일이 없거나 키가 없으면 빈 문자열을 반환한다. 테스트용 헬퍼.
func claudeEnvFilePath(projectRoot string) string {
	settingsPath := filepath.Join(projectRoot, ".claude", "settings.local.json")
	raw, err := os.ReadFile(settingsPath)
	if err != nil {
		return ""
	}
	var settings map[string]json.RawMessage
	if err := json.Unmarshal(raw, &settings); err != nil {
		return ""
	}
	envRaw, ok := settings["env"]
	if !ok {
		return ""
	}
	var envMap map[string]string
	if err := json.Unmarshal(envRaw, &envMap); err != nil {
		return ""
	}
	return strings.TrimSpace(envMap["CLAUDE_ENV_FILE"])
}

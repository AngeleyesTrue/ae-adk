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
	// json.RawMessage로 디코딩하여 다른 키의 비-문자열 값(숫자/객체 등)도
	// 원형 그대로 보존한다. map[string]string으로 디코딩하면 외부 도구가 추가한
	// 임의 타입 값에서 unmarshal 실패가 발생하여 전체 주입이 silent하게 중단된다.
	//
	// env가 객체가 아닌 형태(문자열/배열 등)로 잘못 작성되어 있으면 비차단
	// 경고 후 새 맵으로 시작한다. 외부 도구가 settings.local.json을 손상시켰을
	// 때 세션 시작 자체가 막히지 않도록 reliability를 우선한다.
	envMap := make(map[string]json.RawMessage)
	if envRaw, ok := settings["env"]; ok {
		if len(envRaw) > 0 && string(envRaw) != "null" {
			if err := json.Unmarshal(envRaw, &envMap); err != nil {
				slog.Warn("settings.local.json env section is malformed (not an object); resetting env to empty",
					"path", settingsPath,
					"error", err.Error(),
				)
				envMap = make(map[string]json.RawMessage)
			}
		}
	}

	// CLAUDE_ENV_FILE의 기대 직렬화 형태를 미리 계산하고 기존 값과 비교한다.
	expectedRaw, err := json.Marshal(absEnvPath)
	if err != nil {
		return fmt.Errorf("marshal env value: %w", err)
	}
	if existing, ok := envMap["CLAUDE_ENV_FILE"]; ok {
		// 이미 동일 값이 있으면 파일을 건드리지 않는다 (idempotent).
		// json.RawMessage 비교는 직렬화 표현이 동일해야 idempotent로 간주.
		if string(existing) == string(expectedRaw) {
			return nil
		}
	}

	envMap["CLAUDE_ENV_FILE"] = json.RawMessage(expectedRaw)

	// env 맵을 다시 직렬화한다 — 다른 키의 원본 값은 RawMessage 형태로 보존됨.
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
//
// CLAUDE_ENV_FILE 슬롯만 string으로 디코딩하고, 다른 env 키는 무시한다.
// (env 전체를 map[string]string으로 디코딩하면 외부 도구가 추가한 비-문자열
// 값에 대해 fail하므로 슬롯 단위 디코딩이 필요하다.)
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
	envMap := make(map[string]json.RawMessage)
	if err := json.Unmarshal(envRaw, &envMap); err != nil {
		return ""
	}
	target, ok := envMap["CLAUDE_ENV_FILE"]
	if !ok {
		return ""
	}
	var value string
	if err := json.Unmarshal(target, &value); err != nil {
		// 비-문자열 값이면 빈 문자열을 반환한다.
		return ""
	}
	return strings.TrimSpace(value)
}

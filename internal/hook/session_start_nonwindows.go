//go:build !windows

package hook

// injectCLAUDEEnvFile은 Windows 전용 기능이다 (REQ-17).
// macOS/Linux에서는 아무것도 하지 않는다 (no-op).
func injectCLAUDEEnvFile(_ string) error {
	return nil
}

// claudeEnvFilePath는 Windows 전용 헬퍼의 non-Windows stub이다.
func claudeEnvFilePath(_ string) string {
	return ""
}

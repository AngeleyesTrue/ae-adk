package hook

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// sha256File returns the SHA256 hex digest of a file's content.
// Used to verify file content equality in idempotency tests without depending
// on filesystem-specific modtime precision (Windows NTFS 100ns vs Linux ext4 1ns)
// or OS-cached metadata.
func sha256File(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file for hash: %v", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// TestInjectCLAUDEEnvFile_Windows verifies REQ-17 behavior on Windows.
// On non-Windows platforms, the function is a no-op and the test validates that.
func TestInjectCLAUDEEnvFile_NoEnvFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// No .env file present.
	if err := injectCLAUDEEnvFile(dir); err != nil {
		t.Fatalf("injectCLAUDEEnvFile with no .env should not error: %v", err)
	}
	// settings.local.json should NOT be created.
	settingsPath := filepath.Join(dir, ".claude", "settings.local.json")
	if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
		t.Errorf("settings.local.json should not be created when .env is absent")
	}
}

func TestInjectCLAUDEEnvFile_EmptyProjectRoot(t *testing.T) {
	t.Parallel()

	if err := injectCLAUDEEnvFile(""); err != nil {
		t.Fatalf("empty projectRoot should not error: %v", err)
	}
}

func TestInjectCLAUDEEnvFile_WithEnvFile(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific injection: skipping on non-Windows (function is no-op)")
	}

	dir := t.TempDir()

	// Create .env file.
	envFilePath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFilePath, []byte("API_KEY=secret"), 0o600); err != nil {
		t.Fatalf("create .env: %v", err)
	}

	if err := injectCLAUDEEnvFile(dir); err != nil {
		t.Fatalf("injectCLAUDEEnvFile: %v", err)
	}

	got := claudeEnvFilePath(dir)
	if got == "" {
		t.Fatal("CLAUDE_ENV_FILE not written to settings.local.json")
	}
	// Path should use forward slashes.
	for _, c := range got {
		if c == '\\' {
			t.Errorf("path should use forward slashes, got: %s", got)
			break
		}
	}
}

func TestInjectCLAUDEEnvFile_Idempotent(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific injection: skipping on non-Windows (function is no-op)")
	}

	dir := t.TempDir()
	envFilePath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFilePath, []byte("X=1"), 0o600); err != nil {
		t.Fatalf("create .env: %v", err)
	}

	// First call writes the file.
	if err := injectCLAUDEEnvFile(dir); err != nil {
		t.Fatalf("first call: %v", err)
	}
	settingsPath := filepath.Join(dir, ".claude", "settings.local.json")
	hash1 := sha256File(t, settingsPath)

	// Second call should be idempotent (same value → no write).
	if err := injectCLAUDEEnvFile(dir); err != nil {
		t.Fatalf("second call: %v", err)
	}
	hash2 := sha256File(t, settingsPath)

	// Compare file content via SHA256 instead of modtime — content equality is the
	// true idempotency invariant and is independent of filesystem timestamp precision.
	if hash1 != hash2 {
		t.Errorf("idempotency violated: settings.local.json content changed on second call\nfirst SHA256:  %s\nsecond SHA256: %s", hash1, hash2)
	}
}

func TestInjectCLAUDEEnvFile_NonWindowsNoOp(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("this test verifies the non-Windows no-op path")
	}

	dir := t.TempDir()
	// Create .env file — even with it present, function should be no-op on non-Windows.
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("X=1"), 0o600); err != nil {
		t.Fatalf("create .env: %v", err)
	}

	if err := injectCLAUDEEnvFile(dir); err != nil {
		t.Fatalf("non-Windows injectCLAUDEEnvFile should not error: %v", err)
	}

	// settings.local.json must NOT be created on non-Windows.
	settingsPath := filepath.Join(dir, ".claude", "settings.local.json")
	if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
		t.Errorf("settings.local.json must not be created on non-Windows by injectCLAUDEEnvFile")
	}
}

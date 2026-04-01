package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBackupSettings는 설정 파일 백업이 올바르게 동작하는지 확인한다.
func TestBackupSettings(t *testing.T) {
	t.Parallel()

	t.Run("성공: 정상 백업", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		settingsPath := "/home/testuser/.ae/settings.json"
		originalData := []byte(`{"key": "value"}`)
		mock.Files[settingsPath] = originalData

		backupPath, err := BackupSettings(mock, settingsPath)
		if err != nil {
			t.Fatalf("BackupSettings() error = %v", err)
		}

		// 백업 경로 형식 확인
		if !strings.HasPrefix(backupPath, settingsPath+".") {
			t.Errorf("backup path %q should start with %q", backupPath, settingsPath+".")
		}
		if !strings.HasSuffix(backupPath, ".bak") {
			t.Errorf("backup path %q should end with .bak", backupPath)
		}

		// 백업 파일 내용 확인
		data, ok := mock.WrittenFiles[backupPath]
		if !ok {
			t.Fatalf("백업 파일이 기록되지 않았음: %s", backupPath)
		}
		if string(data) != string(originalData) {
			t.Errorf("백업 내용 = %q, want %q", string(data), string(originalData))
		}
	})

	t.Run("실패: 설정 파일이 존재하지 않음", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		settingsPath := "/home/testuser/.ae/settings.json"
		// Files에 등록하지 않음 -> FileExists가 false 반환

		_, err := BackupSettings(mock, settingsPath)
		if err == nil {
			t.Fatal("expected error for missing settings file, got nil")
		}
		if !strings.Contains(err.Error(), "settings file not found") {
			t.Errorf("error message = %q, want containing 'settings file not found'", err.Error())
		}
	})

	t.Run("실패: 파일 읽기 에러", func(t *testing.T) {
		t.Parallel()
		// ReadFile이 실패하도록 Files에는 등록하되(FileExists는 true),
		// 실제 ReadFile에서 에러를 발생시키는 상황을 시뮬레이션
		// MockSystemInfo에서는 Files에 있으면 ReadFile이 성공하므로
		// 별도의 에러 시뮬레이션이 필요 -> 커스텀 mock 사용
		mock := &readErrorMock{
			MockSystemInfo: *NewMockSystemInfo(),
			readError:      fmt.Errorf("disk read error"),
		}
		settingsPath := "/home/testuser/.ae/settings.json"
		mock.Files[settingsPath] = []byte("content") // FileExists가 true 반환

		_, err := BackupSettings(mock, settingsPath)
		if err == nil {
			t.Fatal("expected error for read failure, got nil")
		}
		if !strings.Contains(err.Error(), "read settings") {
			t.Errorf("error message = %q, want containing 'read settings'", err.Error())
		}
	})

	t.Run("타임스탬프 형식 검증", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		settingsPath := "/test/settings.json"
		mock.Files[settingsPath] = []byte("{}")

		backupPath, err := BackupSettings(mock, settingsPath)
		if err != nil {
			t.Fatalf("BackupSettings() error = %v", err)
		}

		// 경로에서 타임스탬프 부분 추출
		// 형식: settings.json.20260324-153000.bak
		parts := strings.TrimPrefix(backupPath, settingsPath+".")
		parts = strings.TrimSuffix(parts, ".bak")
		if len(parts) != 15 { // YYYYMMDD-HHMMSS = 15자
			t.Errorf("타임스탬프 길이 = %d, want 15 (format: YYYYMMDD-HHMMSS), got %q", len(parts), parts)
		}
	})
}

// readErrorMock은 ReadFile에서 에러를 반환하는 테스트용 mock이다.
type readErrorMock struct {
	MockSystemInfo
	readError error
}

func (m *readErrorMock) ReadFile(path string) ([]byte, error) {
	return nil, m.readError
}

// TestCleanupOldBackups는 오래된 백업 파일 정리가 올바르게 동작하는지 확인한다.
func TestCleanupOldBackups(t *testing.T) {
	t.Parallel()

	t.Run("백업이 maxBackups 이하이면 삭제하지 않음", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		settingsPath := filepath.Join(tmpDir, "settings.json")

		// 5개 이하의 백업 생성 (maxBackups = 5)
		for i := 0; i < 3; i++ {
			name := fmt.Sprintf("settings.json.2026010%d-120000.bak", i+1)
			path := filepath.Join(tmpDir, name)
			if err := os.WriteFile(path, []byte("backup"), 0o644); err != nil {
				t.Fatal(err)
			}
		}

		err := CleanupOldBackups(settingsPath)
		if err != nil {
			t.Fatalf("CleanupOldBackups() error = %v", err)
		}

		// 모든 파일이 남아있는지 확인
		entries, _ := os.ReadDir(tmpDir)
		count := 0
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".bak") {
				count++
			}
		}
		if count != 3 {
			t.Errorf("남은 백업 수 = %d, want 3", count)
		}
	})

	t.Run("백업이 maxBackups 초과이면 오래된 것 삭제", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		settingsPath := filepath.Join(tmpDir, "settings.json")

		// 8개 백업 생성
		timestamps := []string{
			"20260101-120000",
			"20260102-120000",
			"20260103-120000",
			"20260104-120000",
			"20260105-120000",
			"20260106-120000",
			"20260107-120000",
			"20260108-120000",
		}
		for _, ts := range timestamps {
			name := fmt.Sprintf("settings.json.%s.bak", ts)
			path := filepath.Join(tmpDir, name)
			if err := os.WriteFile(path, []byte("backup"), 0o644); err != nil {
				t.Fatal(err)
			}
		}

		err := CleanupOldBackups(settingsPath)
		if err != nil {
			t.Fatalf("CleanupOldBackups() error = %v", err)
		}

		// maxBackups(5)개만 남아있는지 확인
		entries, _ := os.ReadDir(tmpDir)
		var remaining []string
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".bak") {
				remaining = append(remaining, e.Name())
			}
		}
		if len(remaining) != 5 {
			t.Errorf("남은 백업 수 = %d, want 5, files: %v", len(remaining), remaining)
		}

		// 가장 최근 5개가 남아있는지 확인
		for _, ts := range timestamps[3:] { // 20260104 ~ 20260108
			expected := fmt.Sprintf("settings.json.%s.bak", ts)
			found := false
			for _, r := range remaining {
				if r == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("최근 백업 %s 가 삭제됨", expected)
			}
		}
	})

	t.Run("정확히 maxBackups 개이면 삭제하지 않음", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		settingsPath := filepath.Join(tmpDir, "settings.json")

		for i := 0; i < 5; i++ {
			name := fmt.Sprintf("settings.json.2026010%d-120000.bak", i+1)
			path := filepath.Join(tmpDir, name)
			if err := os.WriteFile(path, []byte("backup"), 0o644); err != nil {
				t.Fatal(err)
			}
		}

		err := CleanupOldBackups(settingsPath)
		if err != nil {
			t.Fatalf("CleanupOldBackups() error = %v", err)
		}

		entries, _ := os.ReadDir(tmpDir)
		count := 0
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".bak") {
				count++
			}
		}
		if count != 5 {
			t.Errorf("남은 백업 수 = %d, want 5", count)
		}
	})

	t.Run("관련 없는 파일은 무시", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		settingsPath := filepath.Join(tmpDir, "settings.json")

		// 7개 백업 + 관련 없는 파일
		for i := 0; i < 7; i++ {
			name := fmt.Sprintf("settings.json.2026010%d-120000.bak", i+1)
			path := filepath.Join(tmpDir, name)
			if err := os.WriteFile(path, []byte("backup"), 0o644); err != nil {
				t.Fatal(err)
			}
		}

		// 관련 없는 파일들
		unrelated := []string{"other.txt", "settings.json", "readme.md"}
		for _, name := range unrelated {
			path := filepath.Join(tmpDir, name)
			if err := os.WriteFile(path, []byte("not a backup"), 0o644); err != nil {
				t.Fatal(err)
			}
		}

		err := CleanupOldBackups(settingsPath)
		if err != nil {
			t.Fatalf("CleanupOldBackups() error = %v", err)
		}

		// 관련 없는 파일이 남아있는지 확인
		for _, name := range unrelated {
			path := filepath.Join(tmpDir, name)
			if _, statErr := os.Stat(path); statErr != nil {
				t.Errorf("관련 없는 파일 %s 이 삭제됨", name)
			}
		}
	})

	t.Run("빈 디렉토리", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		settingsPath := filepath.Join(tmpDir, "settings.json")

		err := CleanupOldBackups(settingsPath)
		if err != nil {
			t.Fatalf("CleanupOldBackups() error = %v", err)
		}
	})

	t.Run("존재하지 않는 디렉토리", func(t *testing.T) {
		t.Parallel()
		settingsPath := filepath.Join(t.TempDir(), "nonexistent", "settings.json")

		err := CleanupOldBackups(settingsPath)
		if err == nil {
			t.Fatal("expected error for nonexistent directory, got nil")
		}
	})
}

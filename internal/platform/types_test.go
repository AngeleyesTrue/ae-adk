package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestDefaultSystemInfo_GOOS는 실제 OS 값을 반환하는지 확인한다.
func TestDefaultSystemInfo_GOOS(t *testing.T) {
	t.Parallel()
	d := &DefaultSystemInfo{}
	got := d.GOOS()
	if got != runtime.GOOS {
		t.Errorf("GOOS() = %q, want %q", got, runtime.GOOS)
	}
}

// TestDefaultSystemInfo_GOARCH는 실제 아키텍처 값을 반환하는지 확인한다.
func TestDefaultSystemInfo_GOARCH(t *testing.T) {
	t.Parallel()
	d := &DefaultSystemInfo{}
	got := d.GOARCH()
	if got != runtime.GOARCH {
		t.Errorf("GOARCH() = %q, want %q", got, runtime.GOARCH)
	}
}

// TestDefaultSystemInfo_HomeDir는 비어있지 않은 홈 디렉토리를 반환하는지 확인한다.
func TestDefaultSystemInfo_HomeDir(t *testing.T) {
	t.Parallel()
	d := &DefaultSystemInfo{}
	got := d.HomeDir()
	if got == "" {
		t.Error("HomeDir() returned empty string")
	}
}

// TestDefaultSystemInfo_GetEnv는 환경 변수를 올바르게 읽는지 확인한다.
func TestDefaultSystemInfo_GetEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
		want string // 빈 문자열이면 존재하지 않는 변수
	}{
		{
			name: "존재하는 환경 변수 (PATH)",
			key:  "PATH",
			want: os.Getenv("PATH"),
		},
		{
			name: "존재하지 않는 환경 변수",
			key:  "AE_ADK_NONEXISTENT_VAR_FOR_TEST_12345",
			want: "",
		},
	}

	d := &DefaultSystemInfo{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := d.GetEnv(tt.key)
			if got != tt.want {
				t.Errorf("GetEnv(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

// TestDefaultSystemInfo_FileExists는 파일 존재 여부를 정확히 판별하는지 확인한다.
func TestDefaultSystemInfo_FileExists(t *testing.T) {
	t.Parallel()

	// 임시 파일 생성
	tmpFile := filepath.Join(t.TempDir(), "test_exists.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "존재하는 파일",
			path: tmpFile,
			want: true,
		},
		{
			name: "존재하지 않는 파일",
			path: filepath.Join(t.TempDir(), "nonexistent.txt"),
			want: false,
		},
		{
			name: "디렉토리는 false",
			path: t.TempDir(),
			want: false,
		},
	}

	d := &DefaultSystemInfo{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := d.FileExists(tt.path)
			if got != tt.want {
				t.Errorf("FileExists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestDefaultSystemInfo_DirExists는 디렉토리 존재 여부를 정확히 판별하는지 확인한다.
func TestDefaultSystemInfo_DirExists(t *testing.T) {
	t.Parallel()

	// 임시 파일 생성 (디렉토리 아닌 것 확인용)
	tmpFile := filepath.Join(t.TempDir(), "not_a_dir.txt")
	if err := os.WriteFile(tmpFile, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "존재하는 디렉토리",
			path: t.TempDir(),
			want: true,
		},
		{
			name: "존재하지 않는 경로",
			path: filepath.Join(t.TempDir(), "no_such_dir"),
			want: false,
		},
		{
			name: "파일은 false",
			path: tmpFile,
			want: false,
		},
	}

	d := &DefaultSystemInfo{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := d.DirExists(tt.path)
			if got != tt.want {
				t.Errorf("DirExists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestDefaultSystemInfo_ExecCommand는 명령어 실행과 에러 처리를 확인한다.
func TestDefaultSystemInfo_ExecCommand(t *testing.T) {
	t.Parallel()

	d := &DefaultSystemInfo{}

	t.Run("성공하는 명령어", func(t *testing.T) {
		t.Parallel()
		out, err := d.ExecCommand("go", "version")
		if err != nil {
			t.Skipf("go가 설치되어 있지 않음: %v", err)
		}
		if out == "" {
			t.Error("ExecCommand(go version) returned empty output")
		}
	})

	t.Run("존재하지 않는 명령어", func(t *testing.T) {
		t.Parallel()
		_, err := d.ExecCommand("ae_adk_nonexistent_cmd_12345")
		if err == nil {
			t.Error("expected error for nonexistent command, got nil")
		}
	})
}

// TestDefaultSystemInfo_ReadWriteFile는 파일 읽기/쓰기를 확인한다.
func TestDefaultSystemInfo_ReadWriteFile(t *testing.T) {
	t.Parallel()

	d := &DefaultSystemInfo{}
	tmpPath := filepath.Join(t.TempDir(), "rw_test.txt")
	content := []byte("read-write test content")

	// 쓰기
	if err := d.WriteFile(tmpPath, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// 읽기
	got, err := d.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("ReadFile() = %q, want %q", string(got), string(content))
	}
}

// TestDefaultSystemInfo_ReadFile_NotFound는 존재하지 않는 파일 읽기 에러를 확인한다.
func TestDefaultSystemInfo_ReadFile_NotFound(t *testing.T) {
	t.Parallel()

	d := &DefaultSystemInfo{}
	_, err := d.ReadFile(filepath.Join(t.TempDir(), "no_such_file.txt"))
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

// TestMockSystemInfo_ImplementsInterface는 MockSystemInfo가 SystemInfo 인터페이스를 충족하는지 확인한다.
func TestMockSystemInfo_ImplementsInterface(t *testing.T) {
	t.Parallel()
	var _ SystemInfo = (*MockSystemInfo)(nil)
	var _ SystemInfo = (*DefaultSystemInfo)(nil)
}

// TestCheckStatus_Constants는 상태 상수 값을 확인한다.
func TestCheckStatus_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status CheckStatus
		want   string
	}{
		{"StatusOK", StatusOK, "ok"},
		{"StatusWarn", StatusWarn, "warn"},
		{"StatusFail", StatusFail, "fail"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.status) != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, string(tt.status), tt.want)
			}
		})
	}
}

// TestPlatformProfile_ZeroValue는 PlatformProfile 제로값이 안전한지 확인한다.
func TestPlatformProfile_ZeroValue(t *testing.T) {
	t.Parallel()
	var p PlatformProfile
	if p.Platform != "" {
		t.Errorf("zero value Platform = %q, want empty", p.Platform)
	}
	if p.Checks != nil {
		t.Errorf("zero value Checks should be nil")
	}
	if p.ToolVersions != nil {
		t.Errorf("zero value ToolVersions should be nil")
	}
}

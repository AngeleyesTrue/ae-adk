package statusline

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestVersionCollector_CheckUpdate(t *testing.T) {
	tests := []struct {
		name                string
		setupConfig         func(t *testing.T) string
		binaryVersion       string
		wantVersion         string
		wantAvailable       bool
		wantUpdate          bool
		wantLatest          string
		wantSyncNeeded      bool
		wantTemplateVersion string
		wantErr             bool
	}{
		{
			name: "valid config with version, same as binary",
			setupConfig: func(t *testing.T) string {
				dir := t.TempDir()
				configDir := filepath.Join(dir, ".ae", "config", "sections")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}
				configPath := filepath.Join(configDir, "system.yaml")
				content := []byte("ae:\n  version: 1.14.0\n")
				if err := os.WriteFile(configPath, content, 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			binaryVersion:  "v1.14.0",
			wantVersion:    "1.14.0",
			wantAvailable:  true,
			wantUpdate:     false,
			wantSyncNeeded: false,
		},
		{
			// REQ-FIX-002-001: Current = 바이너리 버전, 템플릿과 다르면 SyncNeeded = true
			name: "binary newer than template",
			setupConfig: func(t *testing.T) string {
				dir := t.TempDir()
				configDir := filepath.Join(dir, ".ae", "config", "sections")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}
				configPath := filepath.Join(configDir, "system.yaml")
				content := []byte("ae:\n  version: v2.0.0\n")
				if err := os.WriteFile(configPath, content, 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			binaryVersion:       "v2.0.1",
			wantVersion:         "2.0.1", // 바이너리 버전이 Current
			wantAvailable:       true,
			wantUpdate:          false, // UpdateAvailable은 원격 버전 체크용으로 예약
			wantLatest:          "",
			wantSyncNeeded:      true,
			wantTemplateVersion: "2.0.0",
		},
		{
			name: "valid config with v prefix",
			setupConfig: func(t *testing.T) string {
				dir := t.TempDir()
				configDir := filepath.Join(dir, ".ae", "config", "sections")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}
				configPath := filepath.Join(configDir, "system.yaml")
				content := []byte("ae:\n  version: v2.0.0\n")
				if err := os.WriteFile(configPath, content, 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			binaryVersion:  "v2.0.0",
			wantVersion:    "2.0.0",
			wantAvailable:  true,
			wantUpdate:     false,
			wantSyncNeeded: false,
		},
		{
			name: "no config file falls back to binary version",
			setupConfig: func(t *testing.T) string {
				return t.TempDir()
			},
			binaryVersion:  "v2.0.0",
			wantVersion:    "2.0.0",
			wantAvailable:  true,
			wantUpdate:     false,
			wantSyncNeeded: false,
		},
		{
			name: "empty version falls back to binary version",
			setupConfig: func(t *testing.T) string {
				dir := t.TempDir()
				configDir := filepath.Join(dir, ".ae", "config", "sections")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}
				configPath := filepath.Join(configDir, "system.yaml")
				content := []byte("ae:\n  version: ''\n")
				if err := os.WriteFile(configPath, content, 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			binaryVersion:  "v2.0.0",
			wantVersion:    "2.0.0",
			wantAvailable:  true,
			wantUpdate:     false,
			wantSyncNeeded: false,
		},
		{
			// AC-007: 바이너리 "" + 설정 없음 → Available = false
			name: "no config and no binary version",
			setupConfig: func(t *testing.T) string {
				return t.TempDir()
			},
			binaryVersion: "",
			wantAvailable: false,
		},
		{
			// 바이너리 버전이 없으면 템플릿 버전을 Current로 사용 (우아한 폴백)
			name: "no binary version provided",
			setupConfig: func(t *testing.T) string {
				dir := t.TempDir()
				configDir := filepath.Join(dir, ".ae", "config", "sections")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}
				configPath := filepath.Join(configDir, "system.yaml")
				content := []byte("ae:\n  version: v2.0.0\n")
				if err := os.WriteFile(configPath, content, 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			binaryVersion:  "",
			wantVersion:    "2.0.0",
			wantAvailable:  true,
			wantUpdate:     false,
			wantSyncNeeded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 테스트 디렉토리로 이동
			testDir := tt.setupConfig(t)
			originalDir, _ := os.Getwd()
			defer func() { _ = os.Chdir(originalDir) }()
			if err := os.Chdir(testDir); err != nil {
				t.Fatal(err)
			}

			// 캐시 초기화를 위해 새 collector 생성
			v := NewVersionCollector(tt.binaryVersion)
			ctx := context.Background()

			got, err := v.CheckUpdate(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.Available != tt.wantAvailable {
				t.Errorf("CheckUpdate() Available = %v, want %v", got.Available, tt.wantAvailable)
			}

			if tt.wantVersion != "" && got.Current != tt.wantVersion {
				t.Errorf("CheckUpdate() Current = %v, want %v", got.Current, tt.wantVersion)
			}

			if got.UpdateAvailable != tt.wantUpdate {
				t.Errorf("CheckUpdate() UpdateAvailable = %v, want %v", got.UpdateAvailable, tt.wantUpdate)
			}

			if tt.wantLatest != "" && got.Latest != tt.wantLatest {
				t.Errorf("CheckUpdate() Latest = %v, want %v", got.Latest, tt.wantLatest)
			}

			if got.SyncNeeded != tt.wantSyncNeeded {
				t.Errorf("CheckUpdate() SyncNeeded = %v, want %v", got.SyncNeeded, tt.wantSyncNeeded)
			}

			if got.TemplateVersion != tt.wantTemplateVersion {
				t.Errorf("CheckUpdate() TemplateVersion = %v, want %v", got.TemplateVersion, tt.wantTemplateVersion)
			}
		})
	}
}

// TestBuildVersionData_BinaryAsCurrent는 바이너리 버전이 Current가 되고
// 템플릿 버전과 다를 때 SyncNeeded가 true로 설정되는지 검증한다.
// AC-001: Binary "v1.2.4", template "1.2.3" → Current = "1.2.4", SyncNeeded = true
func TestBuildVersionData_BinaryAsCurrent(t *testing.T) {
	v := NewVersionCollector("v1.2.4")
	got := v.buildVersionData("1.2.3")

	if got.Current != "1.2.4" {
		t.Errorf("buildVersionData() Current = %q, want %q", got.Current, "1.2.4")
	}
	if !got.SyncNeeded {
		t.Errorf("buildVersionData() SyncNeeded = false, want true")
	}
	if got.TemplateVersion != "1.2.3" {
		t.Errorf("buildVersionData() TemplateVersion = %q, want %q", got.TemplateVersion, "1.2.3")
	}
	if got.UpdateAvailable {
		t.Errorf("buildVersionData() UpdateAvailable = true, want false (원격 버전 체크용으로 예약)")
	}
	if got.Latest != "" {
		t.Errorf("buildVersionData() Latest = %q, want empty (원격 버전 체크용으로 예약)", got.Latest)
	}
	if !got.Available {
		t.Errorf("buildVersionData() Available = false, want true")
	}
}

// TestBuildVersionData_VersionsMatch는 바이너리와 템플릿 버전이 같을 때
// SyncNeeded가 false로 유지되는지 검증한다.
// AC-002: Binary "v1.2.4", template "1.2.4" → Current = "1.2.4", SyncNeeded = false
func TestBuildVersionData_VersionsMatch(t *testing.T) {
	v := NewVersionCollector("v1.2.4")
	got := v.buildVersionData("1.2.4")

	if got.Current != "1.2.4" {
		t.Errorf("buildVersionData() Current = %q, want %q", got.Current, "1.2.4")
	}
	if got.SyncNeeded {
		t.Errorf("buildVersionData() SyncNeeded = true, want false")
	}
	if got.TemplateVersion != "" {
		t.Errorf("buildVersionData() TemplateVersion = %q, want empty", got.TemplateVersion)
	}
}

// TestBuildVersionData_NoSyncWhenConfigMissing는 설정 파일이 없어서
// CheckUpdate 폴백 경로에서 SyncNeeded가 false로 유지되는지 검증한다.
// AC-006: No config + binary "v1.2.4" → Current = "1.2.4", SyncNeeded = false
func TestBuildVersionData_NoSyncWhenConfigMissing(t *testing.T) {
	dir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	v := NewVersionCollector("v1.2.4")
	got, err := v.CheckUpdate(context.Background())
	if err != nil {
		t.Fatalf("CheckUpdate() error = %v", err)
	}

	if got.Current != "1.2.4" {
		t.Errorf("CheckUpdate() Current = %q, want %q", got.Current, "1.2.4")
	}
	if got.SyncNeeded {
		t.Errorf("CheckUpdate() SyncNeeded = true, want false")
	}
	if !got.Available {
		t.Errorf("CheckUpdate() Available = false, want true")
	}
}

func TestVersionCollector_PrefersTemplateVersion(t *testing.T) {
	// 재현 테스트: ae.template_version이 ae.version과 다를 때,
	// collector는 ae.template_version (ae update로 갱신됨)을 사용해야 한다.
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".ae", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "system.yaml")
	// v0.40.1에서 초기화 후 v2.2.1로 업데이트된 프로젝트 시뮬레이션
	content := []byte("ae:\n  version: 0.40.1\n  template_version: 2.2.1\n")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	v := NewVersionCollector("v2.2.1")
	got, err := v.CheckUpdate(context.Background())
	if err != nil {
		t.Fatalf("CheckUpdate() error = %v", err)
	}

	// 바이너리 버전(2.2.1)이 Current여야 하며, template_version과 동일
	if got.Current != "2.2.1" {
		t.Errorf("CheckUpdate() Current = %q, want %q (바이너리 버전이 Current)", got.Current, "2.2.1")
	}
	// 바이너리와 template_version이 동일하므로 SyncNeeded = false
	if got.SyncNeeded {
		t.Errorf("CheckUpdate() SyncNeeded = true, want false (바이너리가 template_version과 동일)")
	}
	if got.UpdateAvailable {
		t.Errorf("CheckUpdate() UpdateAvailable = true, want false")
	}
}

func TestVersionCollector_FallbackToAEVersion(t *testing.T) {
	// ae.template_version이 없으면 ae.version으로 폴백
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".ae", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "system.yaml")
	content := []byte("ae:\n  version: 1.5.0\n")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	v := NewVersionCollector("v1.5.0")
	got, err := v.CheckUpdate(context.Background())
	if err != nil {
		t.Fatalf("CheckUpdate() error = %v", err)
	}

	// 바이너리 버전(1.5.0)이 Current
	if got.Current != "1.5.0" {
		t.Errorf("CheckUpdate() Current = %q, want %q", got.Current, "1.5.0")
	}
	if got.UpdateAvailable {
		t.Errorf("CheckUpdate() UpdateAvailable = true, want false")
	}
}

func TestFormatVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"v1.14.0", "1.14.0"},
		{"1.14.0", "1.14.0"},
		{"v2.0.0", "2.0.0"},
		{"2.0.0", "2.0.0"},
		{"v", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := formatVersion(tt.input); got != tt.want {
				t.Errorf("formatVersion(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

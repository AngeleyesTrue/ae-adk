package platform

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestProfilePath는 프로필 경로가 올바르게 구성되는지 확인한다.
func TestProfilePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		homeDir string
		want    string
	}{
		{
			name:    "일반 홈 디렉토리",
			homeDir: "/home/user",
			want:    filepath.Join("/home/user", ".ae", "platform-profile.json"),
		},
		{
			name:    "Windows 스타일 경로",
			homeDir: `C:\Users\testuser`,
			want:    filepath.Join(`C:\Users\testuser`, ".ae", "platform-profile.json"),
		},
		{
			name:    "빈 홈 디렉토리",
			homeDir: "",
			want:    filepath.Join("", ".ae", "platform-profile.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := NewMockSystemInfo()
			mock.HomeDirVal = tt.homeDir

			got := ProfilePath(mock)
			if got != tt.want {
				t.Errorf("ProfilePath() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSaveProfile은 프로필 저장이 올바르게 동작하는지 확인한다.
func TestSaveProfile(t *testing.T) {
	t.Parallel()

	t.Run("성공: 디렉토리가 이미 존재하는 경우", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.HomeDirVal = "/home/testuser"
		expectedDir := filepath.Dir(ProfilePath(mock))
		mock.Dirs[expectedDir] = true

		profile := &PlatformProfile{
			Platform:  "windows",
			Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Checks: []PlatformCheck{
				{Name: "Go", Status: StatusOK, Message: "go1.26"},
			},
			PATH:         []string{"/usr/bin", "/usr/local/bin"},
			ToolVersions: map[string]string{"go": "go1.26"},
		}

		err := SaveProfile(mock, profile)
		if err != nil {
			t.Fatalf("SaveProfile() error = %v", err)
		}

		// 파일이 기록되었는지 확인
		expectedPath := ProfilePath(mock)
		data, ok := mock.WrittenFiles[expectedPath]
		if !ok {
			t.Fatalf("프로필 파일이 기록되지 않았음: %s", expectedPath)
		}

		// JSON 역직렬화 검증
		var loaded PlatformProfile
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Fatalf("기록된 데이터 JSON 파싱 실패: %v", err)
		}
		if loaded.Platform != profile.Platform {
			t.Errorf("Platform = %q, want %q", loaded.Platform, profile.Platform)
		}
		if len(loaded.Checks) != len(profile.Checks) {
			t.Errorf("Checks count = %d, want %d", len(loaded.Checks), len(profile.Checks))
		}
	})

	t.Run("성공: 디렉토리가 존재하지 않는 경우 (실제 파일시스템)", func(t *testing.T) {
		t.Parallel()

		// os.MkdirAll이 실제로 호출되므로 실제 임시 디렉토리 사용
		tmpDir := t.TempDir()
		mock := NewMockSystemInfo()
		mock.HomeDirVal = tmpDir
		// Dirs에 해당 경로가 없으면 os.MkdirAll이 호출됨

		profile := &PlatformProfile{
			Platform:     "linux",
			Timestamp:    time.Now(),
			ToolVersions: map[string]string{},
		}

		err := SaveProfile(mock, profile)
		if err != nil {
			t.Fatalf("SaveProfile() error = %v", err)
		}

		// .ae 디렉토리가 생성되었는지 확인
		aeDir := filepath.Join(tmpDir, ".ae")
		if info, statErr := os.Stat(aeDir); statErr != nil || !info.IsDir() {
			t.Errorf(".ae 디렉토리가 생성되지 않았음: %s", aeDir)
		}
	})

	t.Run("실패: nil 프로필 직렬화", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.Dirs[filepath.Dir(ProfilePath(mock))] = true

		// nil 프로필은 패닉이 아닌 에러를 발생시키지 않음 (json.Marshal(nil) -> "null")
		// 하지만 nil 포인터 역참조가 발생할 수 있으므로 방어적 테스트
		defer func() {
			if r := recover(); r != nil {
				// nil 프로필에 대한 패닉은 예상된 동작
				t.Logf("nil 프로필에 대해 패닉 발생 (예상됨): %v", r)
			}
		}()
		_ = SaveProfile(mock, nil)
	})
}

// TestLoadProfile은 프로필 로드가 올바르게 동작하는지 확인한다.
func TestLoadProfile(t *testing.T) {
	t.Parallel()

	t.Run("성공: 정상 프로필 로드", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.HomeDirVal = "/home/testuser"

		profile := &PlatformProfile{
			Platform:  "windows",
			Timestamp: time.Date(2026, 3, 24, 12, 0, 0, 0, time.UTC),
			Checks: []PlatformCheck{
				{Name: "Go", Status: StatusOK, Message: "go1.26"},
				{Name: "Node", Status: StatusWarn, Message: "not found"},
			},
			PATH:         []string{"/usr/bin"},
			ToolVersions: map[string]string{"go": "go1.26", "node": "not found"},
		}

		data, _ := json.MarshalIndent(profile, "", "  ")
		profilePath := ProfilePath(mock)
		mock.Files[profilePath] = data

		loaded, err := LoadProfile(mock)
		if err != nil {
			t.Fatalf("LoadProfile() error = %v", err)
		}
		if loaded == nil {
			t.Fatal("LoadProfile() returned nil")
		}
		if loaded.Platform != "windows" {
			t.Errorf("Platform = %q, want %q", loaded.Platform, "windows")
		}
		if len(loaded.Checks) != 2 {
			t.Errorf("Checks count = %d, want 2", len(loaded.Checks))
		}
		if loaded.ToolVersions["go"] != "go1.26" {
			t.Errorf("ToolVersions[go] = %q, want %q", loaded.ToolVersions["go"], "go1.26")
		}
	})

	t.Run("파일이 존재하지 않으면 nil, nil 반환", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.HomeDirVal = "/home/testuser"

		loaded, err := LoadProfile(mock)
		if err != nil {
			t.Fatalf("LoadProfile() error = %v, want nil", err)
		}
		if loaded != nil {
			t.Errorf("LoadProfile() = %v, want nil", loaded)
		}
	})

	t.Run("잘못된 JSON 데이터", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.HomeDirVal = "/home/testuser"
		profilePath := ProfilePath(mock)
		mock.Files[profilePath] = []byte("{invalid json}")

		loaded, err := LoadProfile(mock)
		if err == nil {
			t.Fatal("LoadProfile() expected error for invalid JSON, got nil")
		}
		if loaded != nil {
			t.Errorf("LoadProfile() profile should be nil on error, got %v", loaded)
		}
	})

	t.Run("빈 JSON 객체", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.HomeDirVal = "/home/testuser"
		profilePath := ProfilePath(mock)
		mock.Files[profilePath] = []byte("{}")

		loaded, err := LoadProfile(mock)
		if err != nil {
			t.Fatalf("LoadProfile() error = %v", err)
		}
		if loaded == nil {
			t.Fatal("LoadProfile() returned nil for empty JSON object")
		}
		if loaded.Platform != "" {
			t.Errorf("Platform = %q, want empty", loaded.Platform)
		}
	})
}

// TestCompareProfiles는 두 프로필 간의 차이를 올바르게 감지하는지 확인한다.
func TestCompareProfiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		old            *PlatformProfile
		new            *PlatformProfile
		wantNil        bool
		wantAdded      int
		wantRemoved    int
		wantChanged    int
		wantStatusDiff int
	}{
		{
			name:    "old가 nil이면 nil 반환",
			old:     nil,
			new:     &PlatformProfile{},
			wantNil: true,
		},
		{
			name: "동일한 프로필",
			old: &PlatformProfile{
				PATH:         []string{"/usr/bin", "/usr/local/bin"},
				ToolVersions: map[string]string{"go": "1.26"},
				Checks:       []PlatformCheck{{Name: "Go", Status: StatusOK}},
			},
			new: &PlatformProfile{
				PATH:         []string{"/usr/bin", "/usr/local/bin"},
				ToolVersions: map[string]string{"go": "1.26"},
				Checks:       []PlatformCheck{{Name: "Go", Status: StatusOK}},
			},
			wantAdded:      0,
			wantRemoved:    0,
			wantChanged:    0,
			wantStatusDiff: 0,
		},
		{
			name: "PATH 추가",
			old: &PlatformProfile{
				PATH:         []string{"/usr/bin"},
				ToolVersions: map[string]string{},
			},
			new: &PlatformProfile{
				PATH:         []string{"/usr/bin", "/opt/new/bin"},
				ToolVersions: map[string]string{},
			},
			wantAdded:   1,
			wantRemoved: 0,
		},
		{
			name: "PATH 삭제",
			old: &PlatformProfile{
				PATH:         []string{"/usr/bin", "/old/path"},
				ToolVersions: map[string]string{},
			},
			new: &PlatformProfile{
				PATH:         []string{"/usr/bin"},
				ToolVersions: map[string]string{},
			},
			wantAdded:   0,
			wantRemoved: 1,
		},
		{
			name: "PATH 추가 및 삭제 동시",
			old: &PlatformProfile{
				PATH:         []string{"/a", "/b"},
				ToolVersions: map[string]string{},
			},
			new: &PlatformProfile{
				PATH:         []string{"/b", "/c"},
				ToolVersions: map[string]string{},
			},
			wantAdded:   1, // /c
			wantRemoved: 1, // /a
		},
		{
			name: "도구 버전 변경",
			old: &PlatformProfile{
				ToolVersions: map[string]string{"go": "1.25", "node": "20.0"},
			},
			new: &PlatformProfile{
				ToolVersions: map[string]string{"go": "1.26", "node": "20.0"},
			},
			wantChanged: 1, // go만 변경
		},
		{
			name: "진단 상태 변경",
			old: &PlatformProfile{
				ToolVersions: map[string]string{},
				Checks: []PlatformCheck{
					{Name: "Go", Status: StatusOK},
					{Name: "Node", Status: StatusOK},
				},
			},
			new: &PlatformProfile{
				ToolVersions: map[string]string{},
				Checks: []PlatformCheck{
					{Name: "Go", Status: StatusOK},
					{Name: "Node", Status: StatusFail},
				},
			},
			wantStatusDiff: 1, // Node 상태 변경
		},
		{
			name: "복합 변경 (PATH + 도구 + 상태)",
			old: &PlatformProfile{
				PATH:         []string{"/a"},
				ToolVersions: map[string]string{"go": "1.25"},
				Checks:       []PlatformCheck{{Name: "Go", Status: StatusWarn}},
			},
			new: &PlatformProfile{
				PATH:         []string{"/b"},
				ToolVersions: map[string]string{"go": "1.26"},
				Checks:       []PlatformCheck{{Name: "Go", Status: StatusOK}},
			},
			wantAdded:      1,
			wantRemoved:    1,
			wantChanged:    1,
			wantStatusDiff: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			diff := CompareProfiles(tt.old, tt.new)

			if tt.wantNil {
				if diff != nil {
					t.Fatalf("CompareProfiles() = %v, want nil", diff)
				}
				return
			}

			if diff == nil {
				t.Fatal("CompareProfiles() returned nil, want non-nil")
			}

			if len(diff.AddedPaths) != tt.wantAdded {
				t.Errorf("AddedPaths count = %d, want %d (paths: %v)",
					len(diff.AddedPaths), tt.wantAdded, diff.AddedPaths)
			}
			if len(diff.RemovedPaths) != tt.wantRemoved {
				t.Errorf("RemovedPaths count = %d, want %d (paths: %v)",
					len(diff.RemovedPaths), tt.wantRemoved, diff.RemovedPaths)
			}
			if len(diff.ChangedTools) != tt.wantChanged {
				t.Errorf("ChangedTools count = %d, want %d", len(diff.ChangedTools), tt.wantChanged)
			}
			if len(diff.StatusDiffs) != tt.wantStatusDiff {
				t.Errorf("StatusDiffs count = %d, want %d", len(diff.StatusDiffs), tt.wantStatusDiff)
			}
		})
	}
}

// TestProfileDiff_HasChanges는 변경사항 존재 여부를 올바르게 판별하는지 확인한다.
func TestProfileDiff_HasChanges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		diff *ProfileDiff
		want bool
	}{
		{
			name: "nil diff는 false",
			diff: nil,
			want: false,
		},
		{
			name: "빈 diff는 false",
			diff: &ProfileDiff{
				ChangedTools: make(map[string][2]string),
			},
			want: false,
		},
		{
			name: "AddedPaths가 있으면 true",
			diff: &ProfileDiff{
				AddedPaths:   []string{"/new/path"},
				ChangedTools: make(map[string][2]string),
			},
			want: true,
		},
		{
			name: "RemovedPaths가 있으면 true",
			diff: &ProfileDiff{
				RemovedPaths: []string{"/old/path"},
				ChangedTools: make(map[string][2]string),
			},
			want: true,
		},
		{
			name: "ChangedTools가 있으면 true",
			diff: &ProfileDiff{
				ChangedTools: map[string][2]string{
					"go": {"1.25", "1.26"},
				},
			},
			want: true,
		},
		{
			name: "StatusDiffs가 있으면 true",
			diff: &ProfileDiff{
				ChangedTools: make(map[string][2]string),
				StatusDiffs: []CheckStatusDiff{
					{Name: "Go", OldStatus: StatusWarn, NewStatus: StatusOK},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.diff.HasChanges()
			if got != tt.want {
				t.Errorf("HasChanges() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSaveAndLoadProfile_RoundTrip은 저장 후 로드했을 때 데이터가 보존되는지 확인한다.
func TestSaveAndLoadProfile_RoundTrip(t *testing.T) {
	t.Parallel()

	mock := NewMockSystemInfo()
	mock.HomeDirVal = "/home/testuser"
	mock.Dirs[filepath.Dir(ProfilePath(mock))] = true

	original := &PlatformProfile{
		Platform:  "darwin",
		Timestamp: time.Date(2026, 3, 24, 15, 30, 0, 0, time.UTC),
		Checks: []PlatformCheck{
			{Name: "Homebrew", Status: StatusOK, Message: "4.0.0"},
			{Name: "Git", Status: StatusOK, Message: "2.44.0", Detail: "최신 버전"},
			{Name: "Node", Status: StatusWarn, Message: "미설치"},
		},
		PATH:         []string{"/opt/homebrew/bin", "/usr/local/bin", "/usr/bin"},
		ToolVersions: map[string]string{"go": "go1.26", "git": "2.44.0", "node": "not found"},
	}

	// 저장
	if err := SaveProfile(mock, original); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}

	// 로드
	loaded, err := LoadProfile(mock)
	if err != nil {
		t.Fatalf("LoadProfile() error = %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadProfile() returned nil")
	}

	// 필드별 검증
	if loaded.Platform != original.Platform {
		t.Errorf("Platform = %q, want %q", loaded.Platform, original.Platform)
	}
	if len(loaded.Checks) != len(original.Checks) {
		t.Errorf("Checks count = %d, want %d", len(loaded.Checks), len(original.Checks))
	}
	if len(loaded.PATH) != len(original.PATH) {
		t.Errorf("PATH count = %d, want %d", len(loaded.PATH), len(original.PATH))
	}
	for k, v := range original.ToolVersions {
		if loaded.ToolVersions[k] != v {
			t.Errorf("ToolVersions[%s] = %q, want %q", k, loaded.ToolVersions[k], v)
		}
	}
}

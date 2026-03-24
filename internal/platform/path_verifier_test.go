package platform

import (
	"os"
	"strings"
	"testing"
)

// TestVerifyPaths는 PATH 문자열의 각 경로 존재 여부를 올바르게 검증하는지 확인한다.
func TestVerifyPaths(t *testing.T) {
	t.Parallel()

	sep := string(os.PathListSeparator)

	tests := []struct {
		name      string
		pathStr   string
		dirs      map[string]bool
		wantCount int        // 결과 항목 수
		wantExist []bool     // 각 항목의 Exists 값
		wantPaths []string   // 각 항목의 Path 값
	}{
		{
			name:      "모든 경로 존재",
			pathStr:   strings.Join([]string{"/usr/bin", "/usr/local/bin"}, sep),
			dirs:      map[string]bool{"/usr/bin": true, "/usr/local/bin": true},
			wantCount: 2,
			wantExist: []bool{true, true},
			wantPaths: []string{"/usr/bin", "/usr/local/bin"},
		},
		{
			name:      "일부 경로만 존재",
			pathStr:   strings.Join([]string{"/usr/bin", "/nonexistent"}, sep),
			dirs:      map[string]bool{"/usr/bin": true},
			wantCount: 2,
			wantExist: []bool{true, false},
			wantPaths: []string{"/usr/bin", "/nonexistent"},
		},
		{
			name:      "모든 경로 미존재",
			pathStr:   strings.Join([]string{"/no1", "/no2"}, sep),
			dirs:      map[string]bool{},
			wantCount: 2,
			wantExist: []bool{false, false},
			wantPaths: []string{"/no1", "/no2"},
		},
		{
			name:      "빈 문자열",
			pathStr:   "",
			dirs:      map[string]bool{},
			wantCount: 0,
		},
		{
			name:      "공백만 있는 경로는 무시",
			pathStr:   strings.Join([]string{"  ", "/usr/bin", "  "}, sep),
			dirs:      map[string]bool{"/usr/bin": true},
			wantCount: 1,
			wantExist: []bool{true},
			wantPaths: []string{"/usr/bin"},
		},
		{
			name:      "단일 경로",
			pathStr:   "/usr/bin",
			dirs:      map[string]bool{"/usr/bin": true},
			wantCount: 1,
			wantExist: []bool{true},
			wantPaths: []string{"/usr/bin"},
		},
		{
			name:      "경로 앞뒤 공백 트림",
			pathStr:   strings.Join([]string{" /usr/bin ", " /opt/bin "}, sep),
			dirs:      map[string]bool{"/usr/bin": true, "/opt/bin": true},
			wantCount: 2,
			wantExist: []bool{true, true},
			wantPaths: []string{"/usr/bin", "/opt/bin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := NewMockSystemInfo()
			mock.Dirs = tt.dirs

			results := VerifyPaths(mock, tt.pathStr)

			if len(results) != tt.wantCount {
				t.Fatalf("VerifyPaths() returned %d results, want %d", len(results), tt.wantCount)
			}

			for i, r := range results {
				if i < len(tt.wantExist) && r.Exists != tt.wantExist[i] {
					t.Errorf("results[%d].Exists = %v, want %v (path: %q)", i, r.Exists, tt.wantExist[i], r.Path)
				}
				if i < len(tt.wantPaths) && r.Path != tt.wantPaths[i] {
					t.Errorf("results[%d].Path = %q, want %q", i, r.Path, tt.wantPaths[i])
				}
			}
		})
	}
}

// TestFilterExistingPaths는 존재하는 경로만 필터링하는지 확인한다.
func TestFilterExistingPaths(t *testing.T) {
	t.Parallel()

	sep := string(os.PathListSeparator)

	tests := []struct {
		name    string
		pathStr string
		dirs    map[string]bool
		want    string
	}{
		{
			name:    "모든 경로 존재 시 전체 반환",
			pathStr: strings.Join([]string{"/a", "/b"}, sep),
			dirs:    map[string]bool{"/a": true, "/b": true},
			want:    strings.Join([]string{"/a", "/b"}, sep),
		},
		{
			name:    "일부만 존재 시 존재하는 것만 반환",
			pathStr: strings.Join([]string{"/a", "/b", "/c"}, sep),
			dirs:    map[string]bool{"/a": true, "/c": true},
			want:    strings.Join([]string{"/a", "/c"}, sep),
		},
		{
			name:    "모두 미존재 시 빈 문자열",
			pathStr: strings.Join([]string{"/x", "/y"}, sep),
			dirs:    map[string]bool{},
			want:    "",
		},
		{
			name:    "빈 입력",
			pathStr: "",
			dirs:    map[string]bool{},
			want:    "",
		},
		{
			name:    "공백 항목은 무시",
			pathStr: strings.Join([]string{"", "/a", "  ", "/b"}, sep),
			dirs:    map[string]bool{"/a": true, "/b": true},
			want:    strings.Join([]string{"/a", "/b"}, sep),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := NewMockSystemInfo()
			mock.Dirs = tt.dirs

			got := FilterExistingPaths(mock, tt.pathStr)
			if got != tt.want {
				t.Errorf("FilterExistingPaths() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCountMissing은 존재하지 않는 경로 수를 올바르게 세는지 확인한다.
func TestCountMissing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		results []PathVerifyResult
		want    int
	}{
		{
			name:    "nil 입력",
			results: nil,
			want:    0,
		},
		{
			name:    "빈 슬라이스",
			results: []PathVerifyResult{},
			want:    0,
		},
		{
			name: "모두 존재",
			results: []PathVerifyResult{
				{Path: "/a", Exists: true},
				{Path: "/b", Exists: true},
			},
			want: 0,
		},
		{
			name: "모두 미존재",
			results: []PathVerifyResult{
				{Path: "/a", Exists: false},
				{Path: "/b", Exists: false},
			},
			want: 2,
		},
		{
			name: "혼합",
			results: []PathVerifyResult{
				{Path: "/a", Exists: true},
				{Path: "/b", Exists: false},
				{Path: "/c", Exists: true},
				{Path: "/d", Exists: false},
				{Path: "/e", Exists: false},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CountMissing(tt.results)
			if got != tt.want {
				t.Errorf("CountMissing() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestMissingPaths는 존재하지 않는 경로 목록을 올바르게 반환하는지 확인한다.
func TestMissingPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		results []PathVerifyResult
		want    []string
	}{
		{
			name:    "nil 입력",
			results: nil,
			want:    nil,
		},
		{
			name:    "빈 슬라이스",
			results: []PathVerifyResult{},
			want:    nil,
		},
		{
			name: "모두 존재 시 nil 반환",
			results: []PathVerifyResult{
				{Path: "/a", Exists: true},
				{Path: "/b", Exists: true},
			},
			want: nil,
		},
		{
			name: "미존재 경로만 반환",
			results: []PathVerifyResult{
				{Path: "/a", Exists: true},
				{Path: "/b", Exists: false},
				{Path: "/c", Exists: true},
				{Path: "/d", Exists: false},
			},
			want: []string{"/b", "/d"},
		},
		{
			name: "모두 미존재",
			results: []PathVerifyResult{
				{Path: "/x", Exists: false},
				{Path: "/y", Exists: false},
			},
			want: []string{"/x", "/y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MissingPaths(tt.results)

			if tt.want == nil {
				if got != nil {
					t.Errorf("MissingPaths() = %v, want nil", got)
				}
				return
			}

			if len(got) != len(tt.want) {
				t.Fatalf("MissingPaths() returned %d paths, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("MissingPaths()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

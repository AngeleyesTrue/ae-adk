package platform

import (
	"fmt"
	"os"
	"sync"
)

// MockSystemInfo는 테스트용 SystemInfo 인터페이스 구현체이다.
// 모든 시스템 호출을 인메모리로 대체하여 격리된 테스트를 가능하게 한다.
type MockSystemInfo struct {
	GOOSVal       string
	GOARCHVal     string
	HomeDirVal    string
	EnvVars       map[string]string
	Files         map[string][]byte // 경로 -> 파일 내용
	Dirs          map[string]bool   // 경로 -> 존재 여부
	Commands      map[string]string // "cmd args" -> 출력
	CommandErrors map[string]error  // "cmd args" -> 에러
	WrittenFiles  map[string][]byte // 기록된 파일 추적

	mu sync.Mutex // WrittenFiles 동시 접근 보호
}

// NewMockSystemInfo는 기본값으로 초기화된 MockSystemInfo를 생성한다.
func NewMockSystemInfo() *MockSystemInfo {
	return &MockSystemInfo{
		GOOSVal:       "windows",
		GOARCHVal:     "amd64",
		HomeDirVal:    "/home/testuser",
		EnvVars:       make(map[string]string),
		Files:         make(map[string][]byte),
		Dirs:          make(map[string]bool),
		Commands:      make(map[string]string),
		CommandErrors: make(map[string]error),
		WrittenFiles:  make(map[string][]byte),
	}
}

func (m *MockSystemInfo) GOOS() string   { return m.GOOSVal }
func (m *MockSystemInfo) GOARCH() string { return m.GOARCHVal }
func (m *MockSystemInfo) HomeDir() string { return m.HomeDirVal }

func (m *MockSystemInfo) GetEnv(key string) string {
	if m.EnvVars == nil {
		return ""
	}
	return m.EnvVars[key]
}

func (m *MockSystemInfo) FileExists(path string) bool {
	if m.Files == nil {
		return false
	}
	_, ok := m.Files[path]
	return ok
}

func (m *MockSystemInfo) DirExists(path string) bool {
	if m.Dirs == nil {
		return false
	}
	return m.Dirs[path]
}

func (m *MockSystemInfo) ExecCommand(name string, args ...string) (string, error) {
	// 명령어 키 생성: "name arg1 arg2 ..."
	key := name
	for _, a := range args {
		key += " " + a
	}

	// 에러 맵 먼저 확인
	if m.CommandErrors != nil {
		if err, ok := m.CommandErrors[key]; ok {
			return "", err
		}
	}

	// 출력 맵 확인
	if m.Commands != nil {
		if out, ok := m.Commands[key]; ok {
			return out, nil
		}
	}

	return "", fmt.Errorf("command not found: %s", key)
}

func (m *MockSystemInfo) ReadFile(path string) ([]byte, error) {
	// 먼저 기록된 파일 확인
	m.mu.Lock()
	if m.WrittenFiles != nil {
		if data, ok := m.WrittenFiles[path]; ok {
			m.mu.Unlock()
			return data, nil
		}
	}
	m.mu.Unlock()

	// 그 다음 초기 파일 확인
	if m.Files != nil {
		if data, ok := m.Files[path]; ok {
			return data, nil
		}
	}
	return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
}

func (m *MockSystemInfo) WriteFile(path string, data []byte, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.WrittenFiles == nil {
		m.WrittenFiles = make(map[string][]byte)
	}
	m.WrittenFiles[path] = append([]byte(nil), data...) // 방어적 복사
	return nil
}

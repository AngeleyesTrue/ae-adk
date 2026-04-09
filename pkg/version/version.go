package version

import (
	"fmt"
	"runtime/debug"
	"sync"
)

// -ldflags로 빌드 시 주입되는 변수.
// 프로덕션 빌드에서는 ldflags로 덮어쓰며, RC/테스트 빌드에서는 기본값 사용.
var (
	Version = "v1.0.0"
	Commit  = "none"
	Date    = "unknown"
)

// defaultVersion은 ldflags 미주입 시 사용되는 기본 버전.
// GetVersion() 호출 시 이 값이면 runtime/debug.ReadBuildInfo로 폴백 시도.
const defaultVersion = "v1.0.0"

// once는 resolveFromBuildInfo를 한 번만 실행하기 위한 sync.Once 포인터.
// 테스트에서 새 인스턴스로 교체 가능하도록 포인터로 선언.
var once = &sync.Once{}

// readBuildInfo는 runtime/debug.ReadBuildInfo의 래퍼.
// 테스트에서 모킹 가능하도록 함수 변수로 선언.
var readBuildInfo = debug.ReadBuildInfo

// resolveFromBuildInfo는 ldflags 미주입 시 runtime/debug.ReadBuildInfo에서
// 모듈 버전과 VCS 정보를 읽어 패키지 변수를 갱신하는 내부 함수.
// 각 필드(Version/Commit/Date)는 독립적으로 폴백됨.
func resolveFromBuildInfo() {
	bi, ok := readBuildInfo()
	if !ok || bi == nil {
		return
	}

	// Version 폴백: ldflags 기본값이고, 모듈 버전이 유효한 경우에만
	if Version == defaultVersion {
		mv := bi.Main.Version
		if mv != "" && mv != "(devel)" {
			Version = mv
		}
	}

	// Commit/Date 폴백: 기본값인 경우 vcs.revision/vcs.time에서 독립적으로 갱신
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			if Commit == "none" && s.Value != "" {
				Commit = s.Value
			}
		case "vcs.time":
			if Date == "unknown" && s.Value != "" {
				Date = s.Value
			}
		}
	}
}

// @MX:ANCHOR: [AUTO] 13개 파일에서 참조하는 버전 조회 핵심 함수
// @MX:REASON: 변경 시 CLI, hook, statusline 등 전체 버전 표시에 영향
// GetVersion은 현재 버전 문자열을 반환.
// 첫 호출 시 sync.Once로 BuildInfo 폴백을 시도함.
func GetVersion() string {
	once.Do(resolveFromBuildInfo)
	return Version
}

// GetCommit은 빌드 커밋 해시를 반환.
func GetCommit() string {
	once.Do(resolveFromBuildInfo)
	return Commit
}

// GetDate는 빌드 일시를 반환.
func GetDate() string {
	once.Do(resolveFromBuildInfo)
	return Date
}

// GetFullVersion은 포맷된 전체 버전 문자열을 반환.
func GetFullVersion() string {
	once.Do(resolveFromBuildInfo)
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, Date)
}

package foundation

// constants.go — 하드코딩 상수 집중화 (REQ-04, v2.10.2).
//
// 업스트림 커밋 d29771e 패턴 기반으로 ae-adk 코드베이스 내 산재한
// 상수들을 단일 출처로 통합한다.
//
// ── 실제 사용 현황 대 Appendix A 예상 목록 차이 ───────────────────────
//
// Appendix A에 열거된 예상 상수 중 ae-adk 코드베이스에 이미 별도 파일로
// 존재하는 항목은 이 파일에서 중복 선언하지 않는다:
//
//   - DefaultMaxIterations   → internal/config/defaults.go (DefaultMaxIterations = 5)
//   - DefaultLSPTimeout      → internal/foundation/timeouts.go (DefaultLSPTimeout = 3s)
//   - EnvVarEffortLevel      → internal/foundation/envvars.go (EnvClaudeCodeEffortLevel)
//
// 아래 상수들은 ae-adk 코드베이스에 분산되어 있거나 아직 선언이 없는
// 나머지 항목들이다.
//
// ─────────────────────────────────────────────────────────────────────

// GAN Loop / Quality Gate 상수 ─────────────────────────────────────────

const (
	// DefaultPassThreshold는 GAN 루프 품질 게이트의 기본 통과 임계값이다.
	// 0.75 이상이면 PASS, 미만이면 FAIL로 처리하여 피드백 루프를 반복한다.
	DefaultPassThreshold = 0.75

	// DefaultImprovementThreshold는 GAN 루프 반복 간 최소 점수 향상 폭이다.
	// 향상 폭이 이 값 미만이면 루프 정체(stagnation)로 판단한다.
	DefaultImprovementThreshold = 0.05

	// DefaultEscalationAfter는 사용자 개입 없이 자동으로 반복할 수 있는
	// GAN 루프 최대 횟수이다. 이 횟수 초과 시 진행 여부를 사용자에게 묻는다.
	DefaultEscalationAfter = 3
)

// Coverage 상수 ────────────────────────────────────────────────────────

const (
	// DefaultCoverageThreshold는 teammate_idle 훅이 적용하는
	// 기본 테스트 커버리지 임계값(%)이다.
	// quality.yaml에 test_coverage_target이 없을 때 이 값을 사용한다.
	DefaultCoverageThreshold = 85.0
)

// File / IO 상수 ──────────────────────────────────────────────────────

const (
	// DefaultDebounceSeconds는 파일 변경 이벤트 디바운싱에 사용하는
	// 기본 대기 시간(초)이다. 빠른 연속 변경에서 중복 처리를 방지한다.
	DefaultDebounceSeconds = 1

	// MaxFileSizeBytes는 품질 게이트와 스캐너가 처리할 수 있는
	// 단일 파일의 최대 크기이다 (1 MiB).
	// 이보다 큰 파일은 대용량 생성 파일로 간주하여 스킵한다.
	MaxFileSizeBytes = 1024 * 1024 // 1 MiB

	// MaxGoroutineFactor는 runtime.NumCPU() 대비 워커 고루틴 수
	// 배수 상한이다. CPU 코어 수 × 2 이상의 고루틴을 생성하지 않는다.
	MaxGoroutineFactor = 2
)

// Env var 상수 (ast-grep / LSP) ────────────────────────────────────────

const (
	// EnvVarAstGrepBinary는 ast-grep CLI 바이너리 경로를 재정의하는
	// 환경변수명이다. 기본값은 PATH의 "sg" 또는 "ast-grep"이다.
	EnvVarAstGrepBinary = "AE_ASTGREP_BINARY"

	// EnvVarLSPTimeout는 LSP 작업 타임아웃(밀리초)을 재정의하는
	// 환경변수명이다. 미설정 시 foundation.DefaultLSPTimeout이 적용된다.
	EnvVarLSPTimeout = "AE_LSP_TIMEOUT_MS"
)

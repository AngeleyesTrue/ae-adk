package foundation

// 환경변수명 상수 — 하드코딩된 문자열을 단일 출처로 집중화한다 (REQ-04, v2.10.2).
// 이 상수들은 ae-adk 바이너리, 훅 스크립트, 설정 렌더러에서 공통으로 참조한다.

const (
	// EnvAESkipBinaryUpdate는 재실행 시 바이너리 업데이트 단계를 건너뛰도록
	// ae update가 설정하는 환경변수이다.
	EnvAESkipBinaryUpdate = "AE_SKIP_BINARY_UPDATE"

	// EnvEnableToolSearch는 Claude Code 도구 검색 기능 활성화 플래그이다.
	// settings.json env 섹션에서 기본값 "1"로 관리된다.
	EnvEnableToolSearch = "ENABLE_TOOL_SEARCH"

	// EnvClaudeEnvFile은 세션 시작 시 환경 파일 경로를 Claude Code에 전달하는
	// 환경변수이다. Windows에서 injectCLAUDEEnvFile 함수가 주입한다.
	EnvClaudeEnvFile = "CLAUDE_ENV_FILE"

	// EnvClaudeCodeExperimentalAgentTeams는 Agent Teams 실험 기능을 활성화하는
	// 환경변수이다. settings.json env 섹션에서 "1"로 설정된다.
	EnvClaudeCodeExperimentalAgentTeams = "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"

	// EnvClaudeCodeEffortLevel은 Opus 4.7에서 effort 레벨을 지정하는 환경변수이다.
	// ProfilePreferences.EffortLevel 값이 ae 런처 경로에서 이 환경변수로 주입된다.
	EnvClaudeCodeEffortLevel = "CLAUDE_CODE_EFFORT_LEVEL"

	// EnvClaudeCodeFileReadMaxOutputTokens는 파일 읽기 최대 토큰 수를 지정한다.
	// settings.json env 섹션의 "CLAUDE_CODE_FILE_READ_MAX_OUTPUT_TOKENS" 키에 해당한다.
	EnvClaudeCodeFileReadMaxOutputTokens = "CLAUDE_CODE_FILE_READ_MAX_OUTPUT_TOKENS"

	// EnvAEConfigSource는 ae 설정 로딩 방식을 지정한다.
	// "sections" 값이면 개별 섹션 파일에서 설정을 로드한다.
	EnvAEConfigSource = "AE_CONFIG_SOURCE"
)

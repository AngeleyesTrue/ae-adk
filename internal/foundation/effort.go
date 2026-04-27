package foundation

// EffortLevel은 Claude Code의 추론 집약도를 나타내는 타입이다.
// CLAUDE_CODE_EFFORT_LEVEL 환경변수와 settings.json의 effortLevel 필드에 사용된다.
type EffortLevel string

const (
	// EffortLow는 낮은 집약도 — 빠른 탐색, 단순 작업에 적합하다.
	EffortLow EffortLevel = "low"
	// EffortMedium은 중간 집약도 — 대부분의 일반 작업에 적합하다.
	EffortMedium EffortLevel = "medium"
	// EffortHigh는 높은 집약도 — 복잡한 구현 작업에 적합하다.
	EffortHigh EffortLevel = "high"
	// EffortXHigh는 매우 높은 집약도 — 아키텍처 결정, 보안 분석에 적합하다.
	EffortXHigh EffortLevel = "xhigh"
	// EffortMax는 최대 집약도 — Opus 4.7 Adaptive Thinking 최대 활성화에 사용한다.
	EffortMax EffortLevel = "max"
)

// ModelIDOpus47은 Claude claude-opus-4-7 모델의 공식 ID이다.
// v2.12.0에서 도입된 Adaptive Thinking을 지원하는 최신 Opus 모델이다.
const ModelIDOpus47 = "claude-opus-4-7"

// String은 EffortLevel을 YAML/JSON 직렬화에 사용되는 문자열로 반환한다.
func (e EffortLevel) String() string {
	return string(e)
}

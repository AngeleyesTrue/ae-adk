# SPEC-LSP-CORE-002 제안서: powernap 기반 다중 언어 LSP 클라이언트 도입

> Status: Proposal (draft 없음)
> Trigger: SPEC-UPDATE-004 REQ-23 hold (v1.2.0 BLOCKER 해소 결과)
> Priority: Medium
> Estimated Scope: ~30~40 REQ (LSP 클라이언트 추상화, 16개 언어 통합 테스트, 마이그레이션 가이드)
> Dependencies: SPEC-UPDATE-006 (v2.14.0 LSP hygiene 패턴 흡수 후 진행 권장)

---

## 1. 배경 (Why now?)

### 1.1 발생 경위

SPEC-UPDATE-004 v1.0.0 초안은 moai-adk v2.13.0에서 도입된 `charmbracelet/x/powernap` v0.1.4 업그레이드(REQ-23)를 ae-adk에도 동일하게 반영하려 했다. v1.2.0 BLOCKER 진단에서 다음 사실이 드러났다.

- `D:/Sources/ae-adk/go.mod`에 `charmbracelet/x/powernap` 직접 의존성 **부재** (`grep "powernap" go.mod` → 0건)
- `internal/template/templates/.claude/rules/` 하위에 `lsp-client.md` 부재 (프로젝트 루트의 moai 영역 참고문서만 존재)
- 즉, ae-adk는 자체 `internal/lsp/` 패키지로 LSP 처리 중이며 powernap을 사용하지 않는다

`AC-23` 검증 명령 `grep "charmbracelet/x/powernap v0.1.4" go.mod`은 의존성 부재로 영구 실패할 운명이었다. 따라서 SPEC-UPDATE-004 v1.2.0에서 REQ-23은 **hold** 처리되고 별도 SPEC(`SPEC-LSP-CORE-002`)으로 분리하기로 결정됐다.

### 1.2 현재 ae-adk LSP 구현의 한계

`internal/lsp/`를 빠르게 살펴본 결과:

| 항목 | 현재 상태 | 한계 |
|---|---|---|
| 지원 언어 | Go 우선 (gopls 직접 통합), 일부 ad-hoc 처리 | 16개 언어 균질 지원 미확보 |
| JSON-RPC 처리 | 자체 구현 또는 sourcegraph/jsonrpc2 부분 사용 | LSP 라이프사이클(initialize→ready→shutdown) 일관성 부재 |
| 서브프로세스 관리 | os/exec 기반 직접 spawn | stderr drain, singleflight 등 hygiene 패턴 미반영 |
| 16개 언어 호환성 테스트 | 부분 커버 | compliance test 없음 |

이 상태에서는 SPEC-UPDATE-004로 신설한 `lsp.yaml.tmpl`(16개 언어 schema)이 런타임에서 모든 언어를 동등하게 처리한다는 보장이 없다.

### 1.3 powernap 채택 근거

- **검증된 프로덕션 사용**: `charmbracelet/x/powernap`은 charmbracelet/crush(23k+ GitHub stars, 2026-04-12 기준)에서 실제 LSP 서브프로세스 관리에 사용됨
- **안정된 추상화**: `transport.Connection`, `transport.Router`, `lsp.Client`, `lsp.ClientConfig` 등 LSP 라이프사이클 전반을 일관되게 다룸
- **다언어 중립성**: `ClientConfig`는 임의 command + args + initOptions를 받으므로 16개 언어 모든 LSP 서버에 적용 가능
- **moai-adk 정렬**: moai-adk v2.13.2가 동일 라이브러리를 사용 중이어서 향후 업스트림 동기화 비용이 낮음

---

## 2. 필요성 (Rationale)

### 2.1 다언어 균질 지원 보장

ae-adk는 16개 언어(go/python/typescript/javascript/rust/java/kotlin/swift/ruby/php/cpp/csharp/scala/elixir/r/dart)를 SPEC-UPDATE-004 REQ-19로 lsp.yaml.tmpl에 정의했지만, 이 schema가 실제 런타임에서 작동한다는 통합 테스트는 없다. powernap 같은 검증된 추상화 없이는 언어별 미묘한 차이(예: Java jdtls의 workspace folder 요구, Dart language-server의 client-id 인자)를 일일이 재구현해야 한다.

### 2.2 v2.14.0 LSP hygiene 패턴 정렬

moai-adk v2.14.0 Utility Hardening은 LSP subprocess hygiene(stderr drain, singleflight.Group)을 도입했다. SPEC-UPDATE-006이 이를 ae-adk에 이식할 예정이지만, ae-adk가 자체 LSP 클라이언트를 유지하면 매 업스트림 변경마다 재이식 비용이 발생한다. powernap 기반으로 통합하면 이런 hygiene 개선이 라이브러리 업데이트만으로 흡수된다.

### 2.3 Go-only 모놀리식 구현 회피

`golang.org/x/tools/gopls`를 라이브러리로 직접 사용하면 Go 언어만 지원되어 16-language 중립성이 깨진다. ae-adk는 다국어 프로젝트(특히 C# 콘솔 'codedaum')도 타깃이므로 Go-only 결정은 비합리적.

### 2.4 보안 표면 축소

자체 JSON-RPC 처리 코드는 라이브러리 대비 더 많은 attack surface(메모리 안전성, 파싱 오류 처리, 서브프로세스 권한)를 가진다. powernap은 sourcegraph/jsonrpc2 위에 얇은 LSP 추상화만 추가하여 이 표면을 줄인다.

---

## 3. 목표 (Goals)

### 3.1 기능 목표

1. ae-adk가 powernap을 기반으로 16개 언어 LSP 서버를 통합 관리한다
2. 기존 `internal/lsp/` 공용 인터페이스(예: `Diagnostics()`, `Hover()`, `Definition()`)를 보존하여 호출자(quality gate, mx scanner, run phase) 코드 변경을 최소화한다
3. v2.14.0 LSP hygiene 개선(stderr drain, singleflight)을 powernap 채택 동시에 흡수한다
4. lsp.yaml.tmpl의 16개 언어 schema를 powernap `ClientConfig`로 직접 매핑한다

### 3.2 비기능 목표

- **Go 1.26 호환**: 현재 ae-adk Go 버전과 동일
- **테스트 커버리지 ≥85%**: 16개 언어 통합 테스트 + transport 단위 테스트
- **마이그레이션 무중단**: 기존 SPEC-GOPLS-BRIDGE-001 사용처가 깨지지 않도록 dual-path 지원 (선택적 deprecation)
- **바이너리 크기 영향 ≤10 MiB**: powernap 자체는 가볍지만 의존성 jsonrpc2 포함 영향 측정

### 3.3 비목표 (Non-Goals)

- 새로운 언어(17번째 이상) 추가
- LSP 클라이언트 외 다른 영역(예: TreeSitter, Tree-grepper) 변경
- gopls 직접 임포트 제거 (Go 한정 최적화 경로는 유지 가능)

---

## 4. 방향 (Direction)

### 4.1 단계별 접근

**Phase 0: 사전 검증 (Pre-flight)**
- powernap v0.1.4의 public API 안정성 확인 (CHANGELOG, breaking change 검사)
- charmbracelet/x 모노레포 다른 의존성과의 호환성 확인 (`go mod tidy` 시 충돌 여부)
- moai-adk v2.13.2의 powernap 통합 코드를 reference architecture로 분석

**Phase 1: 의존성 추가**
- `go get github.com/charmbracelet/x/powernap@v0.1.4`
- `go.sum` 갱신 + `go vet` 통과 확인
- vendor or proxy 정책 확인

**Phase 2: 추상화 레이어 신설**
- `internal/lsp/transport/` 신규 패키지 — powernap의 `transport.Connection`을 ae 인터페이스로 wrap
- `internal/lsp/client/` 신규 패키지 — `lsp.Client` 라이프사이클을 ae가 관리하는 매니저로 통합
- 기존 `internal/lsp/` 공용 함수는 본 패키지로 위임

**Phase 3: 16개 언어 통합 테스트**
- `internal/lsp/core_test.go` 또는 `_integration_test.go` (build tag 분리)
- 각 언어 LSP 서버 기동 → initialize → diagnostics 1건 → shutdown 라이프사이클 검증
- LSP 서버 미설치 시 skip (CI는 사전 설치된 환경에서만 통과 강제)

**Phase 4: 사용처 마이그레이션**
- `internal/hook/quality/` 의 LSP 호출을 신규 추상화로 전환
- `internal/lsp/hook/` 의 fallback 경로 정리
- `cmd/ae/` 내 직접 LSP 호출 (예: `ae doctor` LSP 진단) 통합

**Phase 5: 템플릿 룰 작성**
- `internal/template/templates/.claude/rules/ae/core/lsp-client.md` 신규 (moai의 lsp-client.md 패턴을 ae 네임스페이스로 변환)
- 업그레이드 정책(REQ-LC-001a 패턴) 명시 — pin된 powernap 버전 변경 시 통합 테스트 통과 의무

**Phase 6: 회귀 테스트 + 문서화**
- 기존 LSP 사용처 모두 정상 동작 확인
- README/CHANGELOG에 마이그레이션 노트 추가
- `ae-delta.md`에 powernap 의존성 신규 등록

### 4.2 위험 완화

| 위험 | 완화 |
|---|---|
| **R-1**: powernap v0.1.4 → 향후 v0.2.x breaking change | pin된 버전 + 업그레이드 정책 문서(`lsp-client.md`)에 통합 테스트 통과 의무 명시 |
| **R-2**: 기존 `internal/lsp/` 호출자 회귀 | dual-path 일정 기간 유지 (legacy + powernap), feature flag로 런타임 전환 |
| **R-3**: Windows 환경 LSP 서버 spawn 차이 | Phase 3 통합 테스트에 Windows runner 포함 (현재 ae-adk Windows-first 정책과 정렬) |
| **R-4**: powernap이 sourcegraph/jsonrpc2를 내부 사용하므로 동일 라이브러리 직접 의존 시 충돌 | `internal/lsp/` 코드에서 sourcegraph/jsonrpc2 직접 import 금지 (powernap 노출 API만 사용) — `lsp-client.md`에 명시 |

### 4.3 의사결정 포인트

| 결정 | 옵션 | 권장 |
|---|---|---|
| LSP 클라이언트 라이브러리 | (A) powernap, (B) 자체 유지, (C) MCP bridge | (A) powernap — 검증·일관성·유지보수 비용 측면 우위 |
| 마이그레이션 전략 | (a) Big-bang 교체, (b) Strangler Fig 점진 교체 | (b) Strangler — Phase 4에서 모듈별 전환 |
| 업스트림 정렬 | (i) moai-adk와 동일 버전 pin, (ii) 독립 버전 관리 | (i) 동일 버전 pin — 향후 SPEC-UPDATE-NNN 동기화 비용 최소화 |

---

## 5. 영향 파일 (예상)

```
go.mod                              (powernap v0.1.4 신규 의존성)
go.sum                              (해시)
internal/lsp/transport/            (신규 패키지, 5~8 파일)
internal/lsp/client/                (신규 패키지, 5~8 파일)
internal/lsp/                       (기존 함수 위임 변경, 10~15 파일)
internal/lsp/core_test.go           (16-language 통합 테스트)
internal/hook/quality/              (LSP 호출 사용처 마이그레이션)
internal/cli/doctor.go              (LSP 진단 통합)
internal/template/templates/.claude/rules/ae/core/lsp-client.md  (신규 룰 문서)
.moai/upstream/ae-delta.md          (powernap 의존성 신규 등록)
```

---

## 6. 의존성 / 트리거

- **선행 머지**: SPEC-UPDATE-006 권장 (v2.14.0 LSP hygiene 패턴 통합 → 본 SPEC에서 powernap 채택과 동시 흡수 가능)
- **선행 정보**: powernap v0.1.4 안정성 외부 검증 (charmbracelet/x 릴리즈 노트, breaking change 이력)
- **트리거**: ae-adk가 16개 언어 모두를 일관되게 지원해야 하는 다음 SPEC(예: 다국어 quality gate 강화) 진입 시 본 SPEC 선행 권장

---

## 7. 다음 단계 (When to start)

본 SPEC은 SPEC-UPDATE-005 머지 + SPEC-UPDATE-006 머지 후 본격 착수 권장.

- **착수 전 체크**: ae-adk 코드베이스에 powernap을 import해도 되는지 라이선스 확인 (Apache-2.0 호환 가정)
- **초안 작성**: `/ae plan SPEC-LSP-CORE-002 "powernap 기반 다중 언어 LSP 클라이언트 도입"` 또는 동등 명령
- **작성자**: expert-backend (Go 라이브러리 통합) + manager-strategy (마이그레이션 전략)

---

## 8. 참고 자료

- moai-adk SPEC-LSP-CORE-002 결정 문서 (`.claude/rules/moai/core/lsp-client.md`)
- charmbracelet/crush의 powernap 사용 사례 (`go.mod` 참조)
- LSP 3.17 spec
- SPEC-UPDATE-004 v1.2.0 REQ-23 hold 처리 근거 (`.moai/specs/SPEC-UPDATE-004/spec.md`)

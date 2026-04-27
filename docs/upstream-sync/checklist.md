# ae-adk 업스트림 동기화 체크리스트

> 신규 SPEC-UPDATE-NNN 착수 시 이 파일을 복사하여 SPEC 디렉토리에 두고 채워간다.
> 사본 위치 예: `.moai/specs/SPEC-UPDATE-NNN/progress-checklist.md`

---

## SPEC 메타데이터

- SPEC ID: SPEC-UPDATE-___
- 스코프: v___ ~ v___
- 담당: ___
- 착수일: ____-__-__
- 목표 머지일: ____-__-__
- PR: #___

---

## Phase 0: 사전 점검

- [ ] `.moai/upstream/ae-delta.md` 읽음
- [ ] auto-memory 관련 항목 확인
  - [ ] `project_moai_update_todo.md`
  - [ ] `feedback_upstream_sync_checklist.md`
  - [ ] `project_ae_vs_moai.md`
- [ ] 현재 `template_version` 확인: `_____` (내부 표기) / `_____` (실제 반영)
- [ ] 버전 표기 vs 실제 반영 불일치 있는가? [ ] 있음 [ ] 없음
  - 있으면 본 SPEC에서 보정 REQ 포함
- [ ] 워킹 트리 깨끗함 (`git status` 변경 없음)
- [ ] 기존 SPEC-UPDATE 머지 상태 확인

---

## Phase 1: 업스트림 변경 감지

- [ ] 최신 릴리즈 조회 실행
  ```bash
  gh api repos/modu-ai/moai-adk/releases --jq '.[0:5] | .[] | "\(.tag_name) | \(.published_at)"'
  ```
- [ ] 감지된 신규 릴리즈 목록 (버전 / 날짜 / 핵심 변경):

  | 버전 | 날짜 | 핵심 변경 |
  |---|---|---|
  | | | |
  | | | |

- [ ] Delta 범위 결정: v___ ~ v___
- [ ] 스코프 분할 필요 여부
  - [ ] 단일 SPEC으로 처리
  - [ ] 품질·성능과 구조 변경 분리 (SPEC-A / SPEC-B)
  - [ ] 보안 패치 단독 SPEC 분리
- [ ] 업스트림 소스 확보
  - [ ] git shallow clone (`git clone --depth 30 --branch v<X.Y.Z>`)
  - [ ] tarball fetch
  - [ ] raw file fetch (단일 파일)
- [ ] 변경 파일 목록 생성 (`git diff --name-only`)
- [ ] 변경 파일 수: _____ 건

---

## Phase 2: 변경 분석

### 2.1 변경 분류

| 버킷 | 건수 | 메모 |
|---|---|---|
| A. 이식 (Port) | | |
| B. 갱신 (Update) | | |
| C. 삭제 (Remove) | | |
| D. 무관 (Skip) | | |

### 2.2 ae 고유 자산 충돌 체크

- [ ] `ae-design-impeccable` 스킬 영향 없음
- [ ] `ae-lang-csharp` 스킬 영향 없음
- [ ] `ae win`, `ae mac` 플랫폼 명령 영향 없음
- [ ] `bracket-scope` 커밋 컨벤션 영향 없음
- [ ] SPEC-PIPELINE-001/002 cascade 방지 영향 없음
- [ ] 제거된 기능 (`ae cc`, `ae glm`, `ae cg`) 재도입 요소 없음

### 2.3 리스크 맵

| 리스크 ID | 설명 | 심각도 | 완화책 |
|---|---|---|---|
| R-01 | | | |
| R-02 | | | |
| R-03 | | | |

### 2.4 외부 검증 필요 항목

- [ ] 새 모델 ID 있음? → Anthropic Models 문서 확인
- [ ] 새 환경변수 있음? → Claude Code 공식 문서 확인
- [ ] 새 frontmatter 필드 있음? → Claude Code Skills 공식 문서 확인
- [ ] 새 의존성 버전 있음? → GitHub Releases 확인
- [ ] 언급된 업스트림 커밋 해시 실존 확인

---

## Phase 3: SPEC 초안 작성

### 3.1 파일 생성

- [ ] `.moai/specs/SPEC-UPDATE-NNN/spec.md` 작성
- [ ] `.moai/specs/SPEC-UPDATE-NNN/plan.md` 작성
- [ ] `.moai/specs/SPEC-UPDATE-NNN/acceptance.md` 작성
- [ ] `.moai/specs/SPEC-UPDATE-NNN/research.md` 작성 (옵션, 복잡한 분석 시)

### 3.2 EARS 형식 검증

- [ ] 모든 REQ가 WHEN/THEN 또는 WHILE/IF/THEN 구조
- [ ] 각 REQ는 원자적 (하나의 측정 가능한 결과)
- [ ] 모호한 표현 없음 ("적절히", "필요에 따라" 등)

### 3.3 파일 경로 실존 검증

```bash
grep -oE '(internal/template/templates|\.claude|\.ae|\.moai)/[^\s`]+' \
  .moai/specs/SPEC-UPDATE-NNN/spec.md | sort -u \
  | while read f; do [ -e "$f" ] && echo "✓ $f" || echo "✗ MISSING $f"; done
```

- [ ] 모든 경로 `✓` 통과
- [ ] `✗ MISSING` 있으면 spec 수정 완료

### 3.4 구체 목록 Appendix

- [ ] "N개 상수" → 상수명 리스트 Appendix
- [ ] "N개 언어" → 언어 × 확장자 매트릭스 Appendix
- [ ] "N개 스킬" → 디렉토리명 리스트 Appendix
- [ ] 카테고리 표현만 있는 항목 → 구체화

### 3.5 네임스페이스 변환 규칙

- [ ] spec.md 환경 섹션에 변환 규칙표 포함
- [ ] 예외 경로 5종 명시
- [ ] 변환 대상 지점 체크리스트 (frontmatter, 본문, 링크, 참조, 코드블록, 주석)

### 3.6 사용자 승인

- [ ] 사용자가 spec 초안을 읽고 승인
- [ ] 승인 시 코멘트/변경 요청 반영 완료
- [ ] 최종 버전 커밋

---

## Phase 4: 실행

### 4.1 브랜치 생성

- [ ] `feature/SPEC-UPDATE-NNN` 브랜치 생성

### 4.2 Phase별 진행 (spec별로 커스텀)

- [ ] Phase 1 완료 + 커밋
- [ ] Phase 2 완료 + 커밋
- [ ] Phase 3 완료 + 커밋
- [ ] Phase 4 완료 + 커밋
- [ ] Phase 5 완료 + 커밋
- [ ] Phase 6 완료 + 커밋
- [ ] Phase 7+ (필요 시)

### 4.3 이식(Port) 작업 3단계 체크

각 이식 대상 파일에 대해:

- [ ] (a) 원본 파일 moai 참조 전수 기록
- [ ] (b) 변환 규칙 매핑 완료
- [ ] (c) 변환 후 grep 재검증 (허용 예외 외 0건)

### 4.4 테스트 추가

- [ ] 신규 Go 코드에 대한 단위 테스트
- [ ] 네임스페이스 변환 회귀 테스트 (`namespace_integrity_test.go`)
- [ ] AC별 자동화 검증 스크립트

---

## Phase 5: 검증 및 머지

### 5.1 Go 빌드 / 테스트

- [ ] `go build ./...` 통과
- [ ] `go vet ./...` 통과
- [ ] `go test ./...` 통과
- [ ] `go test -race ./...` 통과
- [ ] 커버리지 확인:
  - [ ] 신규 패키지 80%+
  - [ ] 전체 평균 85%+

### 5.2 템플릿 통합 테스트

- [ ] `go test ./internal/template/...` 통과
- [ ] 네임스페이스 무결성 테스트 모두 PASS
- [ ] 회귀 테스트 모두 PASS

### 5.3 수동 grep 검증

- [ ] 네임스페이스 잔존 0건
  ```bash
  grep -rn "moai" internal/template/templates/.claude/rules/ae/ \
    | grep -v "modu-ai/moai-adk\|upstream"
  ```
- [ ] 의도한 변경 사항 spot check (3-5개 파일 육안 확인)

### 5.4 AC 기반 수락 검증

| AC ID | 상태 | 비고 |
|---|---|---|
| AC-01 | ☐ PASS ☐ FAIL | |
| AC-02 | ☐ PASS ☐ FAIL | |
| ... | | |

### 5.5 문서 갱신

- [ ] `ae-delta.md` 변경 이력 추가
- [ ] 고유 자산 추가 시 ae-delta.md 체크리스트 갱신
- [ ] CHANGELOG 초안 작성 (선택, 릴리즈 시 반영)

### 5.6 PR 생성

- [ ] PR 제목 bracket-scope 형식 (`[spec-update-NNN] ...`)
- [ ] PR 본문 구조:
  - [ ] 요약 (변경 버킷별 카운트)
  - [ ] 영향 파일 리스트
  - [ ] 테스트 결과
  - [ ] 체크리스트 완료 여부
  - [ ] Breaking change 있으면 명시
- [ ] Reviewer 지정 (Angeleyes 자신 or 외부)

### 5.7 머지 후 후속 작업

- [ ] `project_moai_update_todo.md` 메모리 갱신
- [ ] `template_version` 반영 확인 (`grep template_version ...`)
- [ ] 사용자 프로젝트에서 `ae update` 테스트
- [ ] 다음 SPEC-UPDATE-NNN+1 스코프 사전 검토

---

## 실패 복구 로그

실행 중 막혔을 때 기록:

| 시점 | 문제 | 대응 | 결과 |
|---|---|---|---|
| | | | |

---

## 최종 승인

- [ ] 모든 체크박스 완료
- [ ] 사용자 최종 승인
- [ ] Squash merge 또는 merge commit 결정
- [ ] 머지 완료 (PR #___)
- [ ] 브랜치 삭제

---

**다음 SPEC-UPDATE로 이월된 항목**:

- [ ] ___
- [ ] ___

---

Version: 1.0.0
Last Updated: 2026-04-24

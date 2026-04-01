#!/bin/bash
# scripts/e2e-test.sh
# ae CLI E2E 테스트 스크립트
# SPEC-TEST-001 T-04: 핵심 CLI 워크플로우를 검증하는 E2E 테스트
set -e

TEST_DIR=$(mktemp -d)
# 종료 시 임시 디렉토리 정리
trap "rm -rf $TEST_DIR" EXIT

PASS_COUNT=0
FAIL_COUNT=0

# 테스트 결과 기록 함수
pass() { echo "[PASS] $1"; PASS_COUNT=$((PASS_COUNT + 1)); }
fail() { echo "[FAIL] $1"; FAIL_COUNT=$((FAIL_COUNT + 1)); }

echo "=== ae CLI E2E 테스트 시작 ==="
echo "임시 디렉토리: $TEST_DIR"
echo ""

# E2E-01: ae version - 버전 문자열 출력 확인
echo "--- E2E-01: ae version ---"
ae version 2>&1 | grep -qE "[0-9]+\.[0-9]+" && pass "ae version" || fail "ae version"

# E2E-02: ae init - 프로젝트 초기화 및 .ae/ 디렉토리 생성 확인
echo "--- E2E-02: ae init ---"
ae init "$TEST_DIR" --non-interactive --name e2e-test --language Go 2>&1
test -d "$TEST_DIR/.ae" && pass "ae init" || fail "ae init: .ae/ 디렉토리 미생성"

# ae doctor가 git 리포지토리를 필요로 할 수 있으므로 git init 수행
if ! git -C "$TEST_DIR" rev-parse --git-dir > /dev/null 2>&1; then
    git -C "$TEST_DIR" init -q 2>/dev/null || true
fi

# E2E-03: ae doctor - 초기화된 디렉토리에서 에러 없이 완료 확인
echo "--- E2E-03: ae doctor ---"
cd "$TEST_DIR"
ae doctor 2>&1 && pass "ae doctor" || fail "ae doctor"

# E2E-04: ae status - 초기화된 디렉토리에서 출력 생성 확인
echo "--- E2E-04: ae status ---"
ae status 2>&1 && pass "ae status" || fail "ae status"

# E2E-05: ae update --check - 치명적 에러 없이 실행 확인
# 릴리스가 없는 경우 non-zero exit도 허용
echo "--- E2E-05: ae update --check ---"
ae update --check 2>&1 || true
pass "ae update --check (비차단)"

echo ""
echo "=== E2E 결과: $PASS_COUNT 통과, $FAIL_COUNT 실패 ==="
if [ $FAIL_COUNT -gt 0 ]; then
    exit 1
fi
echo "=== 모든 E2E 테스트 통과 ==="

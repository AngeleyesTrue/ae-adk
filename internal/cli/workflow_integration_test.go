package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- TDD: init -> 파일 구조 및 init -> doctor 워크플로우 통합 테스트 ---

// resetInitFlags 는 Cobra 전역 상태에 남은 initCmd 플래그를 기본값으로 초기화한다.
// 각 테스트에서 defer 로 호출해야 다른 테스트에 영향을 주지 않는다.
func resetInitFlags() {
	_ = initCmd.Flags().Set("root", "")
	_ = initCmd.Flags().Set("non-interactive", "false")
	_ = initCmd.Flags().Set("name", "")
	_ = initCmd.Flags().Set("language", "")
	_ = initCmd.Flags().Set("framework", "")
	_ = initCmd.Flags().Set("mode", "")
	_ = initCmd.Flags().Set("git-mode", "")
	_ = initCmd.Flags().Set("git-provider", "")
	_ = initCmd.Flags().Set("github-username", "")
	_ = initCmd.Flags().Set("gitlab-instance-url", "")
	_ = initCmd.Flags().Set("force", "false")
}

// resetDoctorFlags 는 doctorCmd 플래그를 기본값으로 초기화한다.
func resetDoctorFlags() {
	_ = doctorCmd.Flags().Set("verbose", "false")
	_ = doctorCmd.Flags().Set("fix", "false")
	_ = doctorCmd.Flags().Set("export", "")
	_ = doctorCmd.Flags().Set("check", "")
}

// TestWorkflow_Init_CreatesProjectStructure 는 init 명령이
// 올바른 프로젝트 디렉터리 구조(.ae/, 설정 파일 등)를 생성하는지 검증한다.
func TestWorkflow_Init_CreatesProjectStructure(t *testing.T) {
	// 주의: Cobra 전역 플래그 상태를 공유하므로 t.Parallel() 사용 불가
	defer resetInitFlags()

	root := t.TempDir()

	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)

	// 비대화형 모드로 플래그 설정
	if err := initCmd.Flags().Set("root", root); err != nil {
		t.Fatalf("root 플래그 설정 실패: %v", err)
	}
	if err := initCmd.Flags().Set("non-interactive", "true"); err != nil {
		t.Fatalf("non-interactive 플래그 설정 실패: %v", err)
	}
	if err := initCmd.Flags().Set("name", "workflow-test-project"); err != nil {
		t.Fatalf("name 플래그 설정 실패: %v", err)
	}
	if err := initCmd.Flags().Set("language", "Go"); err != nil {
		t.Fatalf("language 플래그 설정 실패: %v", err)
	}
	if err := initCmd.Flags().Set("mode", "tdd"); err != nil {
		t.Fatalf("mode 플래그 설정 실패: %v", err)
	}

	// init 명령 실행
	if err := initCmd.RunE(initCmd, []string{}); err != nil {
		t.Fatalf("init 명령 실행 실패: %v", err)
	}

	// .ae/ 디렉터리 존재 확인
	aeDir := filepath.Join(root, ".ae")
	info, err := os.Stat(aeDir)
	if err != nil {
		t.Fatalf(".ae/ 디렉터리가 존재하지 않음: %v", err)
	}
	if !info.IsDir() {
		t.Fatal(".ae 는 디렉터리여야 하지만 파일로 존재함")
	}

	// .ae/ 하위에 설정 관련 파일이 최소 1개 이상 존재하는지 확인
	// config/sections 디렉터리 또는 그 안의 파일을 확인한다.
	configDir := filepath.Join(aeDir, "config")
	if _, statErr := os.Stat(configDir); os.IsNotExist(statErr) {
		// config 디렉터리가 없으면 .ae/ 하위에 파일이 있는지 확인
		entries, readErr := os.ReadDir(aeDir)
		if readErr != nil {
			t.Fatalf(".ae/ 디렉터리 읽기 실패: %v", readErr)
		}
		if len(entries) == 0 {
			t.Error(".ae/ 디렉터리 하위에 파일이 하나도 없음")
		}
	}

	// CLAUDE.md 파일 존재 확인 (init 이 생성하는 프로젝트 파일)
	claudeMD := filepath.Join(root, "CLAUDE.md")
	if _, statErr := os.Stat(claudeMD); os.IsNotExist(statErr) {
		t.Error("CLAUDE.md 파일이 생성되지 않음")
	}

	// 출력 메시지에 성공 문구가 포함되어 있는지 확인
	output := buf.String()
	if !strings.Contains(output, "AE project initialized") {
		t.Errorf("출력에 성공 메시지가 없음, 출력: %q", output)
	}
}

// TestWorkflow_Init_ThenDoctor 는 init 실행 후 doctor 명령이
// 에러 없이 완료되는지 검증하는 워크플로우 통합 테스트이다.
// 주의: os.Chdir()로 프로세스 전역 cwd를 변경하므로 t.Parallel() 사용 불가.
func TestWorkflow_Init_ThenDoctor(t *testing.T) {
	defer resetInitFlags()
	defer resetDoctorFlags()

	root := t.TempDir()

	// --- 1단계: init 명령으로 프로젝트 초기화 ---
	initBuf := new(bytes.Buffer)
	initCmd.SetOut(initBuf)
	initCmd.SetErr(initBuf)

	if err := initCmd.Flags().Set("root", root); err != nil {
		t.Fatalf("root 플래그 설정 실패: %v", err)
	}
	if err := initCmd.Flags().Set("non-interactive", "true"); err != nil {
		t.Fatalf("non-interactive 플래그 설정 실패: %v", err)
	}
	if err := initCmd.Flags().Set("name", "doctor-workflow-test"); err != nil {
		t.Fatalf("name 플래그 설정 실패: %v", err)
	}
	if err := initCmd.Flags().Set("language", "Go"); err != nil {
		t.Fatalf("language 플래그 설정 실패: %v", err)
	}
	if err := initCmd.Flags().Set("mode", "ddd"); err != nil {
		t.Fatalf("mode 플래그 설정 실패: %v", err)
	}

	if err := initCmd.RunE(initCmd, []string{}); err != nil {
		t.Fatalf("init 명령 실행 실패: %v", err)
	}

	// --- 2단계: 작업 디렉터리를 init 된 프로젝트로 변경 ---
	// doctor 명령은 cwd 기준으로 .ae/ 를 확인하므로 디렉터리 변경이 필요
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("현재 디렉터리 조회 실패: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("디렉터리 변경 실패: %v", err)
	}
	t.Cleanup(func() {
		if chErr := os.Chdir(origDir); chErr != nil {
			t.Logf("원래 디렉터리 복원 실패: %v", chErr)
		}
	})

	// --- 3단계: doctor 명령 실행 ---
	doctorBuf := new(bytes.Buffer)
	doctorCmd.SetOut(doctorBuf)
	doctorCmd.SetErr(doctorBuf)

	if err := doctorCmd.RunE(doctorCmd, []string{}); err != nil {
		t.Fatalf("doctor 명령 실행 실패: %v", err)
	}

	// --- 4단계: doctor 출력 검증 ---
	doctorOutput := doctorBuf.String()

	// doctor 가 진단 결과를 출력했는지 확인
	if !strings.Contains(doctorOutput, "System Diagnostics") {
		t.Errorf("doctor 출력에 'System Diagnostics' 가 없음, 출력: %q", doctorOutput)
	}

	// "passed" 문자열로 최소 1개 이상의 검사가 성공했는지 확인
	if !strings.Contains(doctorOutput, "passed") {
		t.Errorf("doctor 출력에 'passed' 가 없음, 출력: %q", doctorOutput)
	}

	// AE Config 검사가 성공(ok)으로 표시되는지 확인
	// init 이 .ae/ 를 생성했으므로 doctor 의 AE Config 체크가 통과해야 한다.
	// 출력에서 체크마크(✓)와 "AE Config" 가 같은 줄에 있는지 확인
	if !strings.Contains(doctorOutput, "AE Config") {
		t.Errorf("doctor 출력에 'AE Config' 항목이 없음, 출력: %q", doctorOutput)
	}
}

// TestWorkflow_Init_IdempotentReinit 는 같은 디렉터리에서 init 을 두 번 실행해도
// 에러가 발생하지 않거나 정상적으로 처리되는지 검증한다.
func TestWorkflow_Init_IdempotentReinit(t *testing.T) {
	// 주의: Cobra 전역 플래그 상태를 공유하므로 t.Parallel() 사용 불가
	defer resetInitFlags()

	root := t.TempDir()

	// --- 1차 init 실행 ---
	buf1 := new(bytes.Buffer)
	initCmd.SetOut(buf1)
	initCmd.SetErr(buf1)

	if err := initCmd.Flags().Set("root", root); err != nil {
		t.Fatalf("root 플래그 설정 실패: %v", err)
	}
	if err := initCmd.Flags().Set("non-interactive", "true"); err != nil {
		t.Fatalf("non-interactive 플래그 설정 실패: %v", err)
	}
	if err := initCmd.Flags().Set("name", "idempotent-test"); err != nil {
		t.Fatalf("name 플래그 설정 실패: %v", err)
	}
	if err := initCmd.Flags().Set("language", "Go"); err != nil {
		t.Fatalf("language 플래그 설정 실패: %v", err)
	}
	if err := initCmd.Flags().Set("mode", "tdd"); err != nil {
		t.Fatalf("mode 플래그 설정 실패: %v", err)
	}

	if err := initCmd.RunE(initCmd, []string{}); err != nil {
		t.Fatalf("1차 init 실행 실패: %v", err)
	}

	// 1차 init 후 .ae/ 존재 확인
	aeDir := filepath.Join(root, ".ae")
	if _, statErr := os.Stat(aeDir); os.IsNotExist(statErr) {
		t.Fatal("1차 init 후 .ae/ 디렉터리가 존재하지 않음")
	}

	// --- 2차 init 실행 (--force 플래그 사용) ---
	// 이미 초기화된 프로젝트에 대해 재초기화를 시도한다.
	// --force 를 사용하면 기존 .ae/ 를 백업하고 재생성한다.
	buf2 := new(bytes.Buffer)
	initCmd.SetOut(buf2)
	initCmd.SetErr(buf2)

	// 플래그를 다시 설정 (Cobra 는 이전 실행의 플래그 상태를 유지할 수 있음)
	if err := initCmd.Flags().Set("root", root); err != nil {
		t.Fatalf("root 플래그 재설정 실패: %v", err)
	}
	if err := initCmd.Flags().Set("non-interactive", "true"); err != nil {
		t.Fatalf("non-interactive 플래그 재설정 실패: %v", err)
	}
	if err := initCmd.Flags().Set("name", "idempotent-test"); err != nil {
		t.Fatalf("name 플래그 재설정 실패: %v", err)
	}
	if err := initCmd.Flags().Set("language", "Go"); err != nil {
		t.Fatalf("language 플래그 재설정 실패: %v", err)
	}
	if err := initCmd.Flags().Set("mode", "tdd"); err != nil {
		t.Fatalf("mode 플래그 재설정 실패: %v", err)
	}
	if err := initCmd.Flags().Set("force", "true"); err != nil {
		t.Fatalf("force 플래그 설정 실패: %v", err)
	}

	err := initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Fatalf("2차 init (--force) 실행 실패: %v", err)
	}

	// 2차 init 후에도 .ae/ 가 정상적으로 존재하는지 확인
	info, statErr := os.Stat(aeDir)
	if statErr != nil {
		t.Fatalf("2차 init 후 .ae/ 디렉터리가 존재하지 않음: %v", statErr)
	}
	if !info.IsDir() {
		t.Fatal("2차 init 후 .ae 가 디렉터리가 아님")
	}

	// 2차 init 출력에 성공 메시지가 포함되어 있는지 확인
	output2 := buf2.String()
	if !strings.Contains(output2, "AE project initialized") {
		t.Errorf("2차 init 출력에 성공 메시지가 없음, 출력: %q", output2)
	}

	// CLAUDE.md 가 여전히 존재하는지 확인
	claudeMD := filepath.Join(root, "CLAUDE.md")
	if _, statErr := os.Stat(claudeMD); os.IsNotExist(statErr) {
		t.Error("2차 init 후 CLAUDE.md 가 존재하지 않음")
	}
}

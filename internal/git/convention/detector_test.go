package convention

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestScore_ValidConventionalCommits(t *testing.T) {
	conv, err := ParseBuiltin("conventional-commits")
	if err != nil {
		t.Fatalf("ParseBuiltin: %v", err)
	}

	messages := []string{
		"feat(auth): add JWT token validation",
		"fix: resolve null pointer in user service",
		"docs(readme): update installation guide",
		"chore: update dependencies",
		"refactor: simplify error handling",
	}

	score := Score(messages, conv)
	if score < 0.9 {
		t.Errorf("Score = %.2f, want >= 0.9 for all-matching messages", score)
	}
}

func TestScore_MixedMessages(t *testing.T) {
	conv, err := ParseBuiltin("conventional-commits")
	if err != nil {
		t.Fatalf("ParseBuiltin: %v", err)
	}

	messages := []string{
		"feat(auth): add JWT token validation",
		"random commit message",
		"fix: resolve null pointer",
		"WIP save progress",
	}

	score := Score(messages, conv)
	if score < 0.4 || score > 0.6 {
		t.Errorf("Score = %.2f, want ~0.5 for half-matching messages", score)
	}
}

func TestScore_NoMatches(t *testing.T) {
	conv, err := ParseBuiltin("conventional-commits")
	if err != nil {
		t.Fatalf("ParseBuiltin: %v", err)
	}

	messages := []string{
		"random commit message",
		"another random message",
		"WIP save progress",
	}

	score := Score(messages, conv)
	if score != 0 {
		t.Errorf("Score = %.2f, want 0 for non-matching messages", score)
	}
}

func TestScore_NilConvention(t *testing.T) {
	messages := []string{"feat: add feature"}
	score := Score(messages, nil)
	if score != 0 {
		t.Errorf("Score(nil conv) = %.2f, want 0", score)
	}
}

func TestScore_EmptyMessages(t *testing.T) {
	conv, err := ParseBuiltin("conventional-commits")
	if err != nil {
		t.Fatalf("ParseBuiltin: %v", err)
	}

	score := Score([]string{}, conv)
	if score != 0 {
		t.Errorf("Score(empty) = %.2f, want 0", score)
	}

	score = Score(nil, conv)
	if score != 0 {
		t.Errorf("Score(nil) = %.2f, want 0", score)
	}
}

func TestGetRecentCommitMessages_CurrentRepo(t *testing.T) {
	// Find the repository root by walking up from the working directory.
	repoRoot := findGitRoot(t)

	messages, err := getRecentCommitMessages(repoRoot, 5)
	if err != nil {
		t.Fatalf("getRecentCommitMessages: %v", err)
	}

	if len(messages) == 0 {
		t.Fatal("expected at least one commit message from current repo")
	}
	if len(messages) > 5 {
		t.Errorf("expected at most 5 messages, got %d", len(messages))
	}
}

func TestGetRecentCommitMessages_InvalidRepo(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := getRecentCommitMessages(tmpDir, 10)
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestDetect_CurrentRepo(t *testing.T) {
	repoRoot := findGitRoot(t)

	result, err := Detect(repoRoot, 50)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if result == nil {
		t.Fatal("Detect returned nil for current repo")
	}
	if result.Convention == nil {
		t.Error("DetectionResult.Convention is nil")
	}
	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Confidence = %.2f, want [0, 1]", result.Confidence)
	}
	if result.SampleSize <= 0 {
		t.Errorf("SampleSize = %d, want > 0", result.SampleSize)
	}
}

func TestDetect_InvalidRepo(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := Detect(tmpDir, 10)
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestDetect_DefaultSampleSize(t *testing.T) {
	repoRoot := findGitRoot(t)

	// sampleSize <= 0 should default to 100.
	result, err := Detect(repoRoot, 0)
	if err != nil {
		t.Fatalf("Detect with sampleSize=0: %v", err)
	}
	if result == nil {
		t.Fatal("Detect returned nil")
	}
}

// setupTempGitRepo는 임시 git 저장소를 생성하고 커밋 메시지들을 추가한다.
func setupTempGitRepo(t *testing.T, messages []string) string {
	t.Helper()
	dir := t.TempDir()

	// git init
	cmd := exec.Command("git", "-C", dir, "init")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}

	// git config (CI 호환)
	for _, c := range [][]string{
		{"config", "user.name", "test"},
		{"config", "user.email", "test@test.com"},
	} {
		cmd := exec.Command("git", append([]string{"-C", dir}, c...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", c, err, out)
		}
	}

	// git commit --allow-empty
	for _, msg := range messages {
		cmd := exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", msg)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git commit %q: %v\n%s", msg, err, out)
		}
	}

	return dir
}

func TestDetect_TieBreaking_Deterministic(t *testing.T) {
	// scopeless 커밋만 생성 — bracket-scope와 conventional-commits 모두 매치
	repo := setupTempGitRepo(t, []string{
		"feat: add new feature",
		"fix: resolve bug",
		"docs: update readme",
		"chore: update deps",
		"refactor: simplify code",
	})

	// 10회 반복하여 결정적 결과 검증
	for i := 0; i < 10; i++ {
		result, err := Detect(repo, 10)
		if err != nil {
			t.Fatalf("iteration %d: Detect: %v", i, err)
		}
		if result.Convention.Name != "bracket-scope" {
			t.Errorf("iteration %d: Detect() = %q, want %q",
				i, result.Convention.Name, "bracket-scope")
		}
	}
}

func TestDetect_BracketScopeRepo(t *testing.T) {
	repo := setupTempGitRepo(t, []string{
		"feat: [Web] add restore button",
		"fix: [Auth] resolve token issue",
		"refactor: [Core/DB] optimize query",
		"feat!: [API] change endpoint",
		"docs: [Build] update CI guide",
		"test: [Web] add unit tests",
		"chore: [Build] update deps",
		"feat: [Web/Startup] init sequence",
		"fix: [Auth] session handling",
		"perf: [Core] cache optimization",
	})

	result, err := Detect(repo, 20)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if result.Convention.Name != "bracket-scope" {
		t.Errorf("Detect() = %q, want bracket-scope", result.Convention.Name)
	}
	if result.Confidence < 0.8 {
		t.Errorf("Confidence = %.2f, want >= 0.8", result.Confidence)
	}
}

// findGitRoot walks up from the current working directory to find the
// nearest .git directory.
func findGitRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find git root")
		}
		dir = parent
	}
}

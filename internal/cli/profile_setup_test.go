package cli

import (
	"testing"
)

// TestGetProfileText_ThemeFields verifies that all supported languages
// include translations for the new statusline theme selector fields.
func TestGetProfileText_ThemeFields(t *testing.T) {
	langs := []string{"en", "ko", "ja", "zh"}
	for _, lang := range langs {
		t.Run(lang, func(t *testing.T) {
			text := getProfileText(lang)
			if text.StatuslineThemeTitle == "" {
				t.Errorf("lang %q: StatuslineThemeTitle is empty", lang)
			}
			if text.StatuslineThemeDesc == "" {
				t.Errorf("lang %q: StatuslineThemeDesc is empty", lang)
			}
			if text.ThemeAEDark == "" {
				t.Errorf("lang %q: ThemeAEDark is empty", lang)
			}
			if text.ThemeAELight == "" {
				t.Errorf("lang %q: ThemeAELight is empty", lang)
			}
		})
	}
}

// TestGetProfileText_ModeFields verifies that all supported languages
// include translations for the statusline mode selector fields.
// REQ-V3-MODE-003: Profile wizard must display compact/default/full mode names.
func TestGetProfileText_ModeFields(t *testing.T) {
	langs := []string{"en", "ko", "ja", "zh"}
	for _, lang := range langs {
		t.Run(lang, func(t *testing.T) {
			text := getProfileText(lang)
			if text.StatuslineModeTitle == "" {
				t.Errorf("lang %q: StatuslineModeTitle is empty", lang)
			}
			if text.StatuslineModeDesc == "" {
				t.Errorf("lang %q: StatuslineModeDesc is empty", lang)
			}
			// Validate v3 mode labels
			if text.ModeDefault == "" {
				t.Errorf("lang %q: ModeDefault is empty", lang)
			}
			if text.ModeCompact == "" {
				t.Errorf("lang %q: ModeCompact is empty", lang)
			}
			if text.ModeFull == "" {
				t.Errorf("lang %q: ModeFull is empty", lang)
			}
			// Also validate deprecated fields for backward compatibility
			if text.ModeVerbose == "" {
				t.Errorf("lang %q: ModeVerbose is empty", lang)
			}
			if text.ModeMinimal == "" {
				t.Errorf("lang %q: ModeMinimal is empty", lang)
			}
		})
	}
}

// TestNormalizeModel verifies 100% line coverage on normalizeModel (REQ-09).
func TestNormalizeModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		// 소문자 변환
		{"uppercase passthrough", "CLAUDE-SONNET-4-6", "claude-sonnet-4-6"},
		// 따옴표 제거
		{"double-quoted input", `"claude-opus-4-6"`, "claude-opus-4-6"},
		{"single-quoted input", "'claude-opus-4-6'", "claude-opus-4-6"},
		// 공백 제거
		{"whitespace trimming", "  claude-opus-4-6  ", "claude-opus-4-6"},
		// 구 ID 매핑 → 최신 ID
		{"alias claude-opus-4 → opus-4-7", "claude-opus-4", "claude-opus-4-7"},
		{"alias claude-opus-4-5 → opus-4-6", "claude-opus-4-5", "claude-opus-4-6"},
		{"alias claude-opus-4-6 → opus-4-6 (identity)", "claude-opus-4-6", "claude-opus-4-6"},
		// 알 수 없는 입력은 원본 반환 (silent drop 없음)
		{"unknown model returns as-is", "my-custom-model-v99", "my-custom-model-v99"},
		// 빈 문자열
		{"empty string returns empty", "", ""},
		// 정식 현재 ID는 변경 없음
		{"current opus-4-7 unchanged", "claude-opus-4-7", "claude-opus-4-7"},
		{"current sonnet unchanged", "claude-sonnet-4-6", "claude-sonnet-4-6"},
		{"current haiku unchanged", "claude-haiku-4-5-20251001", "claude-haiku-4-5-20251001"},
		{"opusplan unchanged", "opusplan", "opusplan"},
		// 제어 문자 제거 (CWE-20 defense-in-depth)
		{"null byte stripped", "claude-sonnet\x004-6", "claude-sonnet4-6"},
		{"CR/LF stripped", "claude-sonnet\r\n4-6", "claude-sonnet4-6"},
		// 128자 초과 → 빈 문자열
		{"over 128 chars returns empty", string(make([]byte, 129)), ""},
		// 별칭 매핑은 sanitization 후에도 동작함
		{"alias after whitespace sanitize", "  claude-opus-4  ", "claude-opus-4-7"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeModel(tt.input)
			if got != tt.want {
				t.Errorf("normalizeModel(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestGetProfileText_EffortAndPermissionFields verifies all 4 languages
// have translations for effort level and permission mode fields (REQ-09, REQ-22).
func TestGetProfileText_EffortAndPermissionFields(t *testing.T) {
	t.Parallel()

	langs := []string{"en", "ko", "ja", "zh"}
	for _, lang := range langs {
		t.Run(lang, func(t *testing.T) {
			t.Parallel()
			text := getProfileText(lang)

			// Effort level fields
			if text.EffortLevelTitle == "" {
				t.Errorf("lang %q: EffortLevelTitle is empty", lang)
			}
			if text.EffortLow == "" {
				t.Errorf("lang %q: EffortLow is empty", lang)
			}
			if text.EffortMedium == "" {
				t.Errorf("lang %q: EffortMedium is empty", lang)
			}
			if text.EffortHigh == "" {
				t.Errorf("lang %q: EffortHigh is empty", lang)
			}
			if text.EffortXHigh == "" {
				t.Errorf("lang %q: EffortXHigh is empty", lang)
			}
			if text.EffortMax == "" {
				t.Errorf("lang %q: EffortMax is empty", lang)
			}

			// Opus 4.7 model label
			if text.ModelOpus47 == "" {
				t.Errorf("lang %q: ModelOpus47 is empty", lang)
			}

			// Permission mode fields
			if text.PermissionModeTitle == "" {
				t.Errorf("lang %q: PermissionModeTitle is empty", lang)
			}
			if text.PermModeDefault == "" {
				t.Errorf("lang %q: PermModeDefault is empty", lang)
			}
			if text.PermModeAuto == "" {
				t.Errorf("lang %q: PermModeAuto is empty", lang)
			}
			if text.PermModeAcceptEdits == "" {
				t.Errorf("lang %q: PermModeAcceptEdits is empty", lang)
			}

			// StatuslineMigrationBanner (REQ-22)
			if text.StatuslineMigrationBanner == "" {
				t.Errorf("lang %q: StatuslineMigrationBanner is empty", lang)
			}
		})
	}
}

// TestStatuslineBannerFallback은 미등록 언어(예: "fr")에서
// English 배너로 폴백되는지 검증한다 (REQ-22).
func TestStatuslineBannerFallback(t *testing.T) {
	t.Parallel()

	// 등록되지 않은 언어를 요청하면 English 텍스트를 반환해야 한다.
	text := getProfileText("fr")
	enText := getProfileText("en")

	if text.StatuslineMigrationBanner == "" {
		t.Error("StatuslineMigrationBanner should not be empty for unknown language (fallback to en)")
	}
	if text.StatuslineMigrationBanner != enText.StatuslineMigrationBanner {
		t.Errorf("unknown language fallback: got %q, want English %q",
			text.StatuslineMigrationBanner, enText.StatuslineMigrationBanner)
	}
}

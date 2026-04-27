package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AngeleyesTrue/ae-adk/internal/foundation"
	"github.com/AngeleyesTrue/ae-adk/internal/profile"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// modelAliases는 구 모델 ID를 현재 정식 ID로 매핑한다 (REQ-09, normalizeModel).
// 알 수 없는 입력은 원본 그대로 반환한다 (silent drop 금지).
var modelAliases = map[string]string{
	"claude-opus-4":   foundation.ModelIDOpus47,
	"claude-opus-4-5": "claude-opus-4-6",
	"claude-opus-4-6": "claude-opus-4-6",
}

// normalizeModel은 모델 ID 문자열을 정규화한다 (REQ-09):
//   - 제어 문자 제거 (CWE-20 defense-in-depth)
//   - 128자 초과 시 빈 문자열 반환
//   - 소문자 변환
//   - 앞뒤 따옴표/공백 제거
//   - 구 ID를 최신 정식 ID로 매핑
//   - 알 수 없는 입력은 원본 반환 (silent drop 없음)
func normalizeModel(m string) string {
	// 제어 문자 제거 (U+0000-U+001F, U+007F)
	m = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, m)
	// 길이 제한 (모델 ID는 128자를 초과하지 않는다)
	if len(m) > 128 {
		return ""
	}
	m = strings.ToLower(strings.TrimSpace(m))
	m = strings.Trim(m, `"'`)
	if canonical, ok := modelAliases[m]; ok {
		return canonical
	}
	return m
}

var profileSetupCmd = &cobra.Command{
	Use:   "setup [name]",
	Short: "Interactive setup wizard for profile preferences",
	Long: `Configure per-profile preferences through an interactive wizard.

Settings are stored in:
  ~/.ae/claude-profiles/<name>/preferences.yaml  (identity, language, model, display)

Examples:
  ae profile setup          # Configure default profile
  ae profile setup work     # Configure 'work' profile`,
	Args: cobra.MaximumNArgs(1),
	RunE: runProfileSetup,
}

func init() {
	profileCmd.AddCommand(profileSetupCmd)
}

// runProfileSetup runs the interactive profile configuration wizard.
// The first question is language selection; all subsequent UI text
// is displayed in the selected language.
func runProfileSetup(cmd *cobra.Command, args []string) error {
	profileName := "default"
	if len(args) > 0 {
		profileName = args[0]
	}

	// Load existing config as defaults
	existingPrefs, err := profile.ReadPreferences(profileName)
	if err != nil {
		return fmt.Errorf("read existing preferences: %w", err)
	}

	// Form values pre-filled from existing config
	userName := existingPrefs.UserName

	convLang := existingPrefs.ConversationLang
	if convLang == "" {
		convLang = "en"
	}
	gitCommitLang := existingPrefs.GitCommitLang
	if gitCommitLang == "" {
		gitCommitLang = "en"
	}
	codeCommentLang := existingPrefs.CodeCommentLang
	if codeCommentLang == "" {
		codeCommentLang = "en"
	}
	docLang := existingPrefs.DocLang
	if docLang == "" {
		docLang = "en"
	}

	modelPolicy := existingPrefs.ModelPolicy
	if modelPolicy == "" {
		modelPolicy = "high"
	}
	// normalizeModel 적용: 기존 저장값이 구 ID이면 최신 ID로 보정한다 (REQ-09).
	model := normalizeModel(existingPrefs.Model)
	bypass := existingPrefs.Bypass

	// 권한 모드 (REQ-22: "auto" 옵션 추가)
	permissionMode := existingPrefs.TeammateDisplay // TeammateDisplay를 모드 필드로 재활용하지 않음
	_ = permissionMode                              // 별도 변수로 분리
	permMode := "default"

	// Effort level (REQ-09: 5단계 선택)
	effortLevel := existingPrefs.EffortLevel
	if effortLevel == "" {
		effortLevel = string(foundation.EffortHigh)
	}

	statuslineMode := existingPrefs.StatuslineMode
	if statuslineMode == "" {
		statuslineMode = "default"
	}
	statuslineTheme := existingPrefs.StatuslineTheme
	if statuslineTheme == "" {
		statuslineTheme = "default"
	}
	// ====== Step 1: Language Selection ======
	langOptions := []huh.Option[string]{
		huh.NewOption("English", "en"),
		huh.NewOption("Korean (한국어)", "ko"),
		huh.NewOption("Japanese (日本語)", "ja"),
		huh.NewOption("Chinese (中文)", "zh"),
	}

	langForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select your language").
				Description("Language for this wizard and Claude's responses.").
				Options(langOptions...).
				Value(&convLang),
		).Title("Language"),
	)

	if err := langForm.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Setup cancelled.")
			return nil
		}
		return fmt.Errorf("wizard error: %w", err)
	}

	// ====== Step 2: Remaining form in selected language ======
	t := getProfileText(convLang)

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), t.ConfiguringProfile+"\n\n", profileName)

	form := huh.NewForm(
		// Section 1: Identity
		huh.NewGroup(
			huh.NewInput().
				Title(t.UserNameTitle).
				Description(t.UserNameDesc).
				Value(&userName),
		).Title(t.IdentityTitle),

		// Section 2: Languages (remaining)
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.GitCommitLangTitle).
				Description(t.GitCommitLangDesc).
				Options(langOptions...).
				Value(&gitCommitLang),
			huh.NewSelect[string]().
				Title(t.CodeCommentLangTitle).
				Description(t.CodeCommentLangDesc).
				Options(langOptions...).
				Value(&codeCommentLang),
			huh.NewSelect[string]().
				Title(t.DocLangTitle).
				Description(t.DocLangDesc).
				Options(langOptions...).
				Value(&docLang),
		).Title(t.LanguagesTitle),

		// Section 3: Model Settings (policy + model override + effort level)
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.ModelPolicyTitle).
				Description(t.ModelPolicyDesc).
				Options(
					huh.NewOption(t.ModelPolicyHigh, "high"),
					huh.NewOption(t.ModelPolicyMedium, "medium"),
					huh.NewOption(t.ModelPolicyLow, "low"),
				).
				Value(&modelPolicy),
			huh.NewSelect[string]().
				Title(t.ModelOverrideTitle).
				Description(t.ModelOverrideDesc).
				Options(
					huh.NewOption(t.ModelDefault, ""),
					huh.NewOption(t.ModelOpus47, foundation.ModelIDOpus47),
					huh.NewOption(t.ModelOpus, "claude-opus-4-6"),
					huh.NewOption(t.ModelSonnet, "claude-sonnet-4-6"),
					huh.NewOption(t.ModelHaiku, "claude-haiku-4-5-20251001"),
					huh.NewOption(t.ModelOpusPlan, "opusplan"),
				).
				Value(&model),
			// Effort level selector (REQ-09)
			huh.NewSelect[string]().
				Title(t.EffortLevelTitle).
				Description(t.EffortLevelDesc).
				Options(
					huh.NewOption(t.EffortLow, string(foundation.EffortLow)),
					huh.NewOption(t.EffortMedium, string(foundation.EffortMedium)),
					huh.NewOption(t.EffortHigh, string(foundation.EffortHigh)),
					huh.NewOption(t.EffortXHigh, string(foundation.EffortXHigh)),
					huh.NewOption(t.EffortMax, string(foundation.EffortMax)),
				).
				Value(&effortLevel),
			huh.NewConfirm().
				Title(t.BypassTitle).
				Description(t.BypassDesc).
				Value(&bypass),
		).Title(t.ModelSettingsTitle),

		// Section 4: Permission Mode (REQ-22: "auto" 옵션 추가)
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.PermissionModeTitle).
				Description(t.PermissionModeDesc).
				Options(
					huh.NewOption(t.PermModeDefault, "default"),
					huh.NewOption(t.PermModeAuto, "auto"),
					huh.NewOption(t.PermModeAcceptEdits, "acceptEdits"),
				).
				Value(&permMode),
		).Title(t.PermissionModeTitle),

		// Section 5: Display — Mode, Theme, and Preset in one screen
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.StatuslineModeTitle).
				Description(t.StatuslineModeDesc).
				Options(
					huh.NewOption(t.ModeDefault, "default"),
					huh.NewOption(t.ModeCompact, "compact"),
					huh.NewOption(t.ModeFull, "full"),
				).
				Value(&statuslineMode),
			huh.NewSelect[string]().
				Title(t.StatuslineThemeTitle).
				Description(t.StatuslineThemeDesc).
				Options(
					huh.NewOption(t.ThemeAEDark, "catppuccin-mocha"),
					huh.NewOption(t.ThemeAELight, "catppuccin-latte"),
				).
				Value(&statuslineTheme),
		).Title(t.DisplayTitle),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), t.SetupCancelled)
			return nil
		}
		return fmt.Errorf("wizard error: %w", err)
	}

	// normalizeModel 재적용: 위저드 입력값도 정규화 (REQ-09)
	model = normalizeModel(model)

	// Build and save preferences
	prefs := profile.ProfilePreferences{
		UserName:         userName,
		ConversationLang: convLang,
		GitCommitLang:    gitCommitLang,
		CodeCommentLang:  codeCommentLang,
		DocLang:          docLang,
		ModelPolicy:      modelPolicy,
		Model:            model,
		Bypass:           bypass,
		EffortLevel:      effortLevel,
		StatuslineMode:   statuslineMode,
		StatuslineTheme:  statuslineTheme,
		TeammateDisplay:  permMode,
	}

	if err := profile.WritePreferences(profileName, prefs); err != nil {
		return fmt.Errorf("save preferences: %w", err)
	}

	// Sync preferences to project config if inside a AE project
	if cwd, err := os.Getwd(); err == nil {
		aeDir := filepath.Join(cwd, ".ae")
		if info, err := os.Stat(aeDir); err == nil && info.IsDir() {
			if err := profile.SyncToProjectConfig(cwd, prefs); err != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Warning: failed to sync profile to project config: %v\n", err)
			}
		}
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), t.SavedProfile,
		profileName,
		profile.GetPreferencesPath(profileName))
	return nil
}

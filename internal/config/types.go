package config

import (
	"fmt"
	"slices"
	"time"

	"github.com/AngeleyesTrue/ae-adk/pkg/models"
)

// Config is the root configuration aggregate containing all sections.
// It imports types from pkg/models for shared types (UserConfig, LanguageConfig,
// QualityConfig, ProjectConfig) and defines internal types for the rest.
type Config struct {
	User          models.UserConfig          `yaml:"user"`
	Language      models.LanguageConfig      `yaml:"language"`
	Quality       models.QualityConfig       `yaml:"quality"`
	Project       models.ProjectConfig       `yaml:"project"`
	GitStrategy   GitStrategyConfig          `yaml:"git_strategy"`
	GitConvention models.GitConventionConfig `yaml:"git_convention"`
	System        SystemConfig               `yaml:"system"`
	LLM           LLMConfig                  `yaml:"llm"`
	Pricing       PricingConfig              `yaml:"pricing"`
	Loop          LoopConfig                `yaml:"loop"`
	Workflow      WorkflowConfig             `yaml:"workflow"`
	State         StateConfig                `yaml:"state"`
	Statusline    models.StatuslineConfig    `yaml:"statusline"`
	Gate          GateConfig                 `yaml:"gate"`
	Sunset        SunsetConfig               `yaml:"sunset"`
	Research      ResearchConfig             `yaml:"research"`
	Auto          AutoConfig                 `yaml:"auto"`
}

// GitStrategyConfig represents the git strategy configuration section.
type GitStrategyConfig struct {
	AutoBranch        bool   `yaml:"auto_branch"`
	BranchPrefix      string `yaml:"branch_prefix"`
	CommitStyle       string `yaml:"commit_style"`
	WorktreeRoot      string `yaml:"worktree_root"`
	Provider          string `yaml:"provider"`            // "github", "gitlab"
	GitLabInstanceURL string `yaml:"gitlab_instance_url"` // GitLab instance URL
}

// SystemConfig represents the system configuration section.
type SystemConfig struct {
	Version        string `yaml:"version"`
	LogLevel       string `yaml:"log_level"`
	LogFormat      string `yaml:"log_format"`
	NoColor        bool   `yaml:"no_color"`
	NonInteractive bool   `yaml:"non_interactive"`
}

// LLMConfig represents the LLM configuration section.
type LLMConfig struct {
	// Performance tier: "high", "medium", "low"
	// Controls model selection for all sub-agents and team agents
	PerformanceTier string `yaml:"performance_tier"`
	// Claude model mapping by tier
	ClaudeModels ClaudeTierModels `yaml:"claude_models"`
	// Legacy fields (kept for backward compatibility, mapped from tiers)
	DefaultModel string `yaml:"default_model"`
	QualityModel string `yaml:"quality_model"`
	SpeedModel   string `yaml:"speed_model"`
}

// ClaudeTierModels represents Claude model mappings by performance tier.
type ClaudeTierModels struct {
	High   string `yaml:"high"`   // Complex reasoning, architecture, security
	Medium string `yaml:"medium"` // Balanced performance for most tasks
	Low    string `yaml:"low"`    // Fast exploration, simple tasks
}

// PricingConfig represents the pricing configuration section.
type PricingConfig struct {
	TokenBudget  int  `yaml:"token_budget"`
	CostTracking bool `yaml:"cost_tracking"`
}

// LoopConfig represents the loop controller configuration section.
type LoopConfig struct {
	MaxIterations int  `yaml:"max_iterations"`
	AutoConverge  bool `yaml:"auto_converge"`
	HumanReview   bool `yaml:"human_review"`

	// LintAsInstruction enables injecting LSP diagnostics as systemMessage
	// so the AI receives errors as its next prompt (REQ-LAI-003).
	// Default: true.
	LintAsInstruction bool `yaml:"lint_as_instruction"`

	// WarnAsInstruction includes warnings in the systemMessage when there are
	// no errors and this flag is true (REQ-LAI-006). Default: false.
	WarnAsInstruction bool `yaml:"warn_as_instruction"`
}

// WorkflowConfig represents the workflow configuration section.
type WorkflowConfig struct {
	AutoClear  bool `yaml:"auto_clear"`
	PlanTokens int  `yaml:"plan_tokens"`
	RunTokens  int  `yaml:"run_tokens"`
	SyncTokens int  `yaml:"sync_tokens"`
}

// StateConfig represents the project state storage configuration.
// It controls the directory where structured state data (checkpoints,
// coverage, diagnostics) is stored.
type StateConfig struct {
	StateDir string `yaml:"state_dir"`
}

// LSPQualityGates represents LSP quality gate configuration.
type LSPQualityGates struct {
	Enabled         bool     `yaml:"enabled"`
	Plan            PlanGate `yaml:"plan"`
	Run             RunGate  `yaml:"run"`
	Sync            SyncGate `yaml:"sync"`
	CacheTTLSeconds int      `yaml:"cache_ttl_seconds"`
	TimeoutSeconds  int      `yaml:"timeout_seconds"`
}

// PlanGate represents the plan phase quality gate.
type PlanGate struct {
	RequireBaseline bool `yaml:"require_baseline"`
}

// RunGate represents the run phase quality gate.
type RunGate struct {
	MaxErrors       int  `yaml:"max_errors"`
	MaxTypeErrors   int  `yaml:"max_type_errors"`
	MaxLintErrors   int  `yaml:"max_lint_errors"`
	AllowRegression bool `yaml:"allow_regression"`
}

// SyncGate represents the sync phase quality gate.
type SyncGate struct {
	MaxErrors       int  `yaml:"max_errors"`
	MaxWarnings     int  `yaml:"max_warnings"`
	RequireCleanLSP bool `yaml:"require_clean_lsp"`
}

// GateConfig represents configuration for the deterministic quality gate
// that runs before git commit (SPEC-GATE-001).
type GateConfig struct {
	// Enabled controls whether the quality gate runs.
	Enabled bool `yaml:"enabled"`
	// SkipTests skips the go test step when true.
	SkipTests bool `yaml:"skip_tests"`
	// Timeouts holds per-step timeout values in seconds.
	Timeouts GateTimeouts `yaml:"timeouts"`
}

// GateTimeouts holds per-step timeout configuration in seconds.
type GateTimeouts struct {
	Vet  int `yaml:"vet"`
	Lint int `yaml:"lint"`
	Test int `yaml:"test"`
}

// VetTimeoutDuration converts the Vet timeout to time.Duration.
// Returns 30s when the value is zero or negative.
func (g *GateConfig) VetTimeoutDuration() time.Duration {
	if g.Timeouts.Vet <= 0 {
		return 30 * time.Second
	}
	return time.Duration(g.Timeouts.Vet) * time.Second
}

// LintTimeoutDuration converts the Lint timeout to time.Duration.
// Returns 60s when the value is zero or negative.
func (g *GateConfig) LintTimeoutDuration() time.Duration {
	if g.Timeouts.Lint <= 0 {
		return 60 * time.Second
	}
	return time.Duration(g.Timeouts.Lint) * time.Second
}

// TestTimeoutDuration converts the Test timeout to time.Duration.
// Returns 120s when the value is zero or negative.
func (g *GateConfig) TestTimeoutDuration() time.Duration {
	if g.Timeouts.Test <= 0 {
		return 120 * time.Second
	}
	return time.Duration(g.Timeouts.Test) * time.Second
}

// SunsetConfig defines the Build-to-Delete framework configuration.
// Quality gates that consistently pass can be relaxed over time.
type SunsetConfig struct {
	// Enabled controls whether sunset tracking is active.
	Enabled    bool              `yaml:"enabled"`
	Conditions []SunsetCondition `yaml:"conditions"`
}

// SunsetCondition defines when a quality gate can be relaxed.
type SunsetCondition struct {
	Gate        string `yaml:"gate"`
	Metric      string `yaml:"metric"`
	Threshold   int    `yaml:"threshold"`
	Action      string `yaml:"action"`
	Description string `yaml:"description"`
}

// ResearchConfig represents the Self-Research System configuration section.
type ResearchConfig struct {
	Enabled   bool                    `yaml:"enabled"`
	Passive   ResearchPassiveConfig   `yaml:"passive"`
	Active    ResearchActiveConfig    `yaml:"active"`
	Safety    ResearchSafetyConfig    `yaml:"safety"`
	Dashboard ResearchDashboardConfig `yaml:"dashboard"`
}

// ResearchPassiveConfig represents passive observation settings.
type ResearchPassiveConfig struct {
	Enabled                 bool                      `yaml:"enabled"`
	CorrectionWindowSeconds int                       `yaml:"correction_window_seconds"`
	PatternThresholds       ResearchPatternThresholds `yaml:"pattern_thresholds"`
}

// ResearchPatternThresholds defines observation count thresholds for pattern classification.
type ResearchPatternThresholds struct {
	Heuristic      int `yaml:"heuristic"`
	Rule           int `yaml:"rule"`
	HighConfidence int `yaml:"high_confidence"`
}

// ResearchActiveConfig represents active experiment settings.
type ResearchActiveConfig struct {
	RunsPerExperiment int     `yaml:"runs_per_experiment"`
	MaxExperiments    int     `yaml:"max_experiments"`
	PassThreshold     float64 `yaml:"pass_threshold"`
	TargetScore       float64 `yaml:"target_score"`
	BudgetCapTokens   int     `yaml:"budget_cap_tokens"`
}

// ResearchSafetyConfig represents safety layer settings.
type ResearchSafetyConfig struct {
	WorktreeIsolation         bool                    `yaml:"worktree_isolation"`
	CanaryRegressionThreshold float64                 `yaml:"canary_regression_threshold"`
	RateLimits                ResearchRateLimitConfig `yaml:"rate_limits"`
}

// ResearchRateLimitConfig represents rate limiting settings.
type ResearchRateLimitConfig struct {
	MaxExperimentsPerSession int `yaml:"max_experiments_per_session"`
	MaxAcceptedPerSession    int `yaml:"max_accepted_per_session"`
	MaxAutoResearchPerWeek   int `yaml:"max_auto_research_per_week"`
}

// ResearchDashboardConfig represents dashboard display settings.
type ResearchDashboardConfig struct {
	DefaultMode    string `yaml:"default_mode"`
	HTMLOpenBrowser bool   `yaml:"html_open_browser"`
}

// AutoConfig represents the auto pipeline configuration section.
type AutoConfig struct {
	ContextIsolated ContextIsolatedConfig `yaml:"context_isolated"`
}

// ContextIsolatedConfig represents the context-isolated pipeline settings.
type ContextIsolatedConfig struct {
	Enabled              bool             `yaml:"enabled"`
	SyncReviewIterations int              `yaml:"sync_review_iterations"`
	Copilot              CopilotConfig    `yaml:"copilot"`
	Teammate             TeammateConfig   `yaml:"teammate"`
	FinalMerge           FinalMergeConfig `yaml:"final_merge"`
}

// CopilotConfig represents Copilot review integration settings.
type CopilotConfig struct {
	Enabled        bool   `yaml:"enabled"`
	CheckIteration int    `yaml:"check_iteration"`
	WaitMinutes    int    `yaml:"wait_minutes"`
	BotLogin       string `yaml:"bot_login"`
}

// TeammateConfig represents teammate spawn settings.
type TeammateConfig struct {
	Count int    `yaml:"count"`
	Mode  string `yaml:"mode"`
	Model string `yaml:"model"`
}

// FinalMergeConfig represents the final merge settings.
type FinalMergeConfig struct {
	Strategy      string `yaml:"strategy"`
	DeleteBranch  bool   `yaml:"delete_branch"`
	RequireCIPass bool   `yaml:"require_ci_pass"`
}

// MaxSyncReviewIterations is the upper bound for sync_review_iterations.
const MaxSyncReviewIterations = 10

// Validate checks that AutoConfig values are within acceptable ranges.
func (a *AutoConfig) Validate() error {
	ci := a.ContextIsolated
	if ci.SyncReviewIterations <= 0 {
		return fmt.Errorf("auto: sync_review_iterations must be positive, got %d", ci.SyncReviewIterations)
	}
	if ci.SyncReviewIterations > MaxSyncReviewIterations {
		return fmt.Errorf("auto: sync_review_iterations must be <= %d, got %d", MaxSyncReviewIterations, ci.SyncReviewIterations)
	}
	if ci.Copilot.WaitMinutes < 0 {
		return fmt.Errorf("auto: copilot.wait_minutes must be non-negative, got %d", ci.Copilot.WaitMinutes)
	}
	if ci.Copilot.CheckIteration < 0 {
		return fmt.Errorf("auto: copilot.check_iteration must be non-negative, got %d", ci.Copilot.CheckIteration)
	}
	if ci.Teammate.Count <= 0 {
		return fmt.Errorf("auto: teammate.count must be positive, got %d", ci.Teammate.Count)
	}
	return nil
}

// sectionNames lists all valid configuration section names.
var sectionNames = []string{
	"user", "language", "quality", "project",
	"git_strategy", "git_convention", "system", "llm",
	"pricing", "loop", "workflow", "state", "statusline", "gate", "sunset",
	"research", "auto",
}

// IsValidSectionName checks if the given name is a valid section name.
func IsValidSectionName(name string) bool {
	return slices.Contains(sectionNames, name)
}

// ValidSectionNames returns all valid section names.
func ValidSectionNames() []string {
	result := make([]string, len(sectionNames))
	copy(result, sectionNames)
	return result
}

// YAML file wrapper types for proper unmarshaling with top-level keys.
// Each section file wraps its content under a top-level key.

type userFileWrapper struct {
	User models.UserConfig `yaml:"user"`
}

type languageFileWrapper struct {
	Language models.LanguageConfig `yaml:"language"`
}

// qualityFileWrapper handles the quality.yaml file which uses "constitution:"
// as the top-level key (Python AE-ADK backward compatibility).
type qualityFileWrapper struct {
	Constitution models.QualityConfig `yaml:"constitution"`
}

// gitConventionFileWrapper handles the git-convention.yaml section file.
type gitConventionFileWrapper struct {
	GitConvention models.GitConventionConfig `yaml:"git_convention"`
}

// llmFileWrapper handles the llm.yaml section file.
type llmFileWrapper struct {
	LLM LLMConfig `yaml:"llm"`
}

// stateFileWrapper handles the state.yaml section file.
type stateFileWrapper struct {
	State StateConfig `yaml:"state"`
}

// statuslineFileWrapper handles the statusline.yaml section file.
type statuslineFileWrapper struct {
	Statusline models.StatuslineConfig `yaml:"statusline"`
}

// researchFileWrapper handles the research.yaml section file.
type researchFileWrapper struct {
	Research ResearchConfig `yaml:"research"`
}

// autoFileWrapper handles the auto.yaml section file.
type autoFileWrapper struct {
	Auto AutoConfig `yaml:"auto"`
}

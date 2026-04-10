package config

import (
	"github.com/AngeleyesTrue/ae-adk/pkg/models"
)

// Default value constants to avoid magic numbers and strings.
const (
	DefaultConversationLanguage     = "en"
	DefaultConversationLanguageName = "English"
	DefaultAgentPromptLanguage      = "en"
	DefaultGitCommitMessages        = "en"
	DefaultCodeComments             = "en"
	DefaultDocumentation            = "en"
	DefaultErrorMessages            = "en"

	DefaultTestCoverageTarget    = 85
	DefaultMaxTransformationSize = "small"
	DefaultMinCoveragePerCommit  = 80
	DefaultMaxExemptPercentage   = 5

	DefaultLogLevel  = "info"
	DefaultLogFormat = "text"

	DefaultModel      = "sonnet"
	DefaultQualModel  = "opus"
	DefaultSpeedModel = "haiku"

	DefaultTokenBudget = 250000

	DefaultMaxIterations = 5

	DefaultPlanTokens = 30000
	DefaultRunTokens  = 180000
	DefaultSyncTokens = 40000

	DefaultBranchPrefix = "ae/"
	DefaultCommitStyle  = "conventional"

	// Default performance tier
	DefaultPerformanceTier = "medium"

	DefaultCacheTTLSeconds = 5
	DefaultTimeoutSeconds  = 3
	DefaultMaxWarnings     = 10

	DefaultGitConvention                    = "auto"
	DefaultGitConventionSampleSize          = 100
	DefaultGitConventionConfidenceThreshold = 0.5
	DefaultGitConventionFallback            = "bracket-scope"
	DefaultGitConventionMaxLength           = 100

	DefaultStateDir = ".ae/state"
)

// NewDefaultConfig returns a Config with all fields set to compiled defaults.
func NewDefaultConfig() *Config {
	return &Config{
		User:          NewDefaultUserConfig(),
		Language:      NewDefaultLanguageConfig(),
		Quality:       NewDefaultQualityConfig(),
		Project:       NewDefaultProjectConfig(),
		GitStrategy:   NewDefaultGitStrategyConfig(),
		GitConvention: NewDefaultGitConventionConfig(),
		System:        NewDefaultSystemConfig(),
		LLM:           NewDefaultLLMConfig(),
		Pricing:       NewDefaultPricingConfig(),
		Loop:          NewDefaultLoopConfig(),
		Workflow:      NewDefaultWorkflowConfig(),
		State:         NewDefaultStateConfig(),
		Gate:          NewDefaultGateConfig(),
		Sunset:        NewDefaultSunsetConfig(),
		Research:      NewDefaultResearchConfig(),
	}
}

// NewDefaultUserConfig returns a UserConfig with default values.
// Note: Name is intentionally empty; it is populated from user.yaml.
func NewDefaultUserConfig() models.UserConfig {
	return models.UserConfig{}
}

// NewDefaultLanguageConfig returns a LanguageConfig with default values.
func NewDefaultLanguageConfig() models.LanguageConfig {
	return models.LanguageConfig{
		ConversationLanguage:     DefaultConversationLanguage,
		ConversationLanguageName: DefaultConversationLanguageName,
		AgentPromptLanguage:      DefaultAgentPromptLanguage,
		GitCommitMessages:        DefaultGitCommitMessages,
		CodeComments:             DefaultCodeComments,
		Documentation:            DefaultDocumentation,
		ErrorMessages:            DefaultErrorMessages,
	}
}

// NewDefaultQualityConfig returns a QualityConfig with default values.
func NewDefaultQualityConfig() models.QualityConfig {
	return models.QualityConfig{
		DevelopmentMode:    models.ModeTDD,
		EnforceQuality:     true,
		TestCoverageTarget: DefaultTestCoverageTarget,
		DDDSettings:        NewDefaultDDDSettings(),
		TDDSettings:        NewDefaultTDDSettings(),
		CoverageExemptions: NewDefaultCoverageExemptions(),
	}
}

// NewDefaultDDDSettings returns DDDSettings with default values.
func NewDefaultDDDSettings() models.DDDSettings {
	return models.DDDSettings{
		RequireExistingTests:  true,
		CharacterizationTests: true,
		BehaviorSnapshots:     true,
		MaxTransformationSize: DefaultMaxTransformationSize,
		PreserveBeforeImprove: true,
	}
}

// NewDefaultTDDSettings returns TDDSettings with default values.
func NewDefaultTDDSettings() models.TDDSettings {
	return models.TDDSettings{
		RedGreenRefactor:       true,
		TestFirstRequired:      true,
		MinCoveragePerCommit:   DefaultMinCoveragePerCommit,
		MutationTestingEnabled: false,
	}
}

// NewDefaultCoverageExemptions returns CoverageExemptions with default values.
func NewDefaultCoverageExemptions() models.CoverageExemptions {
	return models.CoverageExemptions{
		Enabled:              false,
		RequireJustification: true,
		MaxExemptPercentage:  DefaultMaxExemptPercentage,
	}
}

// NewDefaultProjectConfig returns a ProjectConfig with default values.
func NewDefaultProjectConfig() models.ProjectConfig {
	return models.ProjectConfig{}
}

// NewDefaultGitStrategyConfig returns a GitStrategyConfig with default values.
func NewDefaultGitStrategyConfig() GitStrategyConfig {
	return GitStrategyConfig{
		AutoBranch:   false,
		BranchPrefix: DefaultBranchPrefix,
		CommitStyle:  DefaultCommitStyle,
		Provider:     "github",
	}
}

// NewDefaultSystemConfig returns a SystemConfig with default values.
func NewDefaultSystemConfig() SystemConfig {
	return SystemConfig{
		LogLevel:  DefaultLogLevel,
		LogFormat: DefaultLogFormat,
	}
}

// NewDefaultLLMConfig returns a LLMConfig with default values.
func NewDefaultLLMConfig() LLMConfig {
	return LLMConfig{
		PerformanceTier: DefaultPerformanceTier,
		ClaudeModels: ClaudeTierModels{
			High:   "opus",
			Medium: "sonnet",
			Low:    "haiku",
		},
		DefaultModel: DefaultModel,
		QualityModel: DefaultQualModel,
		SpeedModel:   DefaultSpeedModel,
	}
}

// NewDefaultPricingConfig returns a PricingConfig with default values.
func NewDefaultPricingConfig() PricingConfig {
	return PricingConfig{
		TokenBudget: DefaultTokenBudget,
	}
}

// NewDefaultLoopConfig returns a LoopConfig with default values.
func NewDefaultLoopConfig() LoopConfig {
	return LoopConfig{
		MaxIterations:     DefaultMaxIterations,
		AutoConverge:      true,
		HumanReview:       true,
		LintAsInstruction: true,  // REQ-LAI-003: enabled by default
		WarnAsInstruction: false, // REQ-LAI-006: disabled by default
	}
}

// NewDefaultWorkflowConfig returns a WorkflowConfig with default values.
func NewDefaultWorkflowConfig() WorkflowConfig {
	return WorkflowConfig{
		AutoClear:  true,
		PlanTokens: DefaultPlanTokens,
		RunTokens:  DefaultRunTokens,
		SyncTokens: DefaultSyncTokens,
	}
}

// NewDefaultStateConfig returns a StateConfig with default values.
func NewDefaultStateConfig() StateConfig {
	return StateConfig{
		StateDir: DefaultStateDir,
	}
}

// NewDefaultGateConfig returns a GateConfig with production-safe defaults.
func NewDefaultGateConfig() GateConfig {
	return GateConfig{
		Enabled:   true,
		SkipTests: false,
		Timeouts: GateTimeouts{
			Vet:  30,
			Lint: 60,
			Test: 120,
		},
	}
}

// NewDefaultSunsetConfig returns a SunsetConfig with default values.
func NewDefaultSunsetConfig() SunsetConfig {
	return SunsetConfig{
		Enabled:    false,
		Conditions: nil,
	}
}

// NewDefaultResearchConfig returns a ResearchConfig with safe defaults.
func NewDefaultResearchConfig() ResearchConfig {
	return ResearchConfig{
		Enabled: false,
		Passive: ResearchPassiveConfig{
			Enabled:                 true,
			CorrectionWindowSeconds: 60,
			PatternThresholds: ResearchPatternThresholds{
				Heuristic:      3,
				Rule:           5,
				HighConfidence: 10,
			},
		},
		Active: ResearchActiveConfig{
			RunsPerExperiment: 3,
			MaxExperiments:    20,
			PassThreshold:     0.80,
			TargetScore:       0.95,
			BudgetCapTokens:   500000,
		},
		Safety: ResearchSafetyConfig{
			WorktreeIsolation:         true,
			CanaryRegressionThreshold: 0.10,
			RateLimits: ResearchRateLimitConfig{
				MaxExperimentsPerSession: 20,
				MaxAcceptedPerSession:    5,
				MaxAutoResearchPerWeek:   3,
			},
		},
		Dashboard: ResearchDashboardConfig{
			DefaultMode:    "terminal",
			HTMLOpenBrowser: true,
		},
	}
}

// NewDefaultGitConventionConfig returns a GitConventionConfig with default values.
func NewDefaultGitConventionConfig() models.GitConventionConfig {
	return models.GitConventionConfig{
		Convention: DefaultGitConvention,
		AutoDetection: models.AutoDetectionConfig{
			Enabled:             true,
			SampleSize:          DefaultGitConventionSampleSize,
			ConfidenceThreshold: DefaultGitConventionConfidenceThreshold,
			Fallback:            DefaultGitConventionFallback,
		},
		Validation: models.ConventionValidationConfig{
			Enabled:         true,
			EnforceOnCommit: false,
			EnforceOnPush:   false,
			MaxLength:       DefaultGitConventionMaxLength,
		},
		Formatting: models.FormattingConfig{
			ShowExamples:    true,
			ShowSuggestions: true,
			Verbose:         false,
		},
	}
}

// NewDefaultLSPQualityGates returns LSPQualityGates with default values.
func NewDefaultLSPQualityGates() LSPQualityGates {
	return LSPQualityGates{
		Enabled: true,
		Plan: PlanGate{
			RequireBaseline: true,
		},
		Run: RunGate{
			MaxErrors:       0,
			MaxTypeErrors:   0,
			MaxLintErrors:   0,
			AllowRegression: false,
		},
		Sync: SyncGate{
			MaxErrors:       0,
			MaxWarnings:     DefaultMaxWarnings,
			RequireCleanLSP: true,
		},
		CacheTTLSeconds: DefaultCacheTTLSeconds,
		TimeoutSeconds:  DefaultTimeoutSeconds,
	}
}

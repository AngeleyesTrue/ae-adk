package defs

// Top-level directory names used by AE-ADK projects.
const (
	// AEDir is the hidden directory that stores AE project state.
	AEDir = ".ae"

	// ClaudeDir is the hidden directory that stores Claude Code configuration.
	ClaudeDir = ".claude"

	// BackupsDir is the directory where project backups are stored.
	BackupsDir = ".moai-backups"
)

// MoAI subdirectory segments (relative to AEDir).
const (
	ConfigSubdir   = "config"
	SectionsSubdir = "config/sections"
	SpecsSubdir    = "specs"
	ReportsSubdir  = "reports"
	StateSubdir    = "state"
	LogsSubdir     = "logs"
	RankSubdir     = "rank"
)

// Claude subdirectory segments (relative to ClaudeDir).
const (
	AgentsMoaiSubdir   = "agents/ae"
	SkillsSubdir       = "skills"
	CommandsMoaiSubdir = "commands/ae"
	RulesMoaiSubdir    = "rules/ae"
	OutputStylesSubdir = "output-styles"
	HooksMoaiSubdir    = "hooks/ae"
)

package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/AngeleyesTrue/ae-adk/internal/cli/worktree"
	"github.com/AngeleyesTrue/ae-adk/pkg/version"
)

var rootCmd = &cobra.Command{
	Use:   "ae",
	Short: "AE-ADK: Agentic Development Kit for Claude Code",
	Long: `AE-ADK (Go Edition) is a high-performance development toolkit
that serves as the runtime backbone for the AE framework within Claude Code.

It provides CLI tooling, configuration management, LSP integration,
Git operations, quality gates, and autonomous development loop capabilities.`,
	Version: version.GetVersion(),
	Run: func(cmd *cobra.Command, args []string) {
		PrintBanner(version.GetVersion())
		_ = cmd.Help()
	},
}

// @MX:ANCHOR: [AUTO] Execute is the main entry point for the ae CLI
// @MX:REASON: [AUTO] fan_in=3, called from cmd/ae/main.go, root_test.go, integration_test.go
// Execute initializes dependencies and runs the root command.
func Execute() error {
	initConsole()
	InitDependencies()
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("ae-adk %s\n", version.GetVersion()))

	// Command groups
	rootCmd.AddGroup(
		&cobra.Group{ID: "project", Title: "Project Commands:"},
		&cobra.Group{ID: "tools", Title: "Tools:"},
	)

	// Wire worktree subcommand with lazy Git initialization
	worktree.WorktreeCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if deps == nil {
			return fmt.Errorf("dependencies not initialized")
		}
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		if err := deps.EnsureGit(cwd); err != nil {
			return fmt.Errorf("initialize git: %w", err)
		}
		worktree.WorktreeProvider = deps.GitWorktree
		return nil
	}

	// Register worktree subcommand tree
	rootCmd.AddCommand(worktree.WorktreeCmd)

	// Register statusline command
	rootCmd.AddCommand(StatuslineCmd)
}

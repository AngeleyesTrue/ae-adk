package cli

import (
	"github.com/spf13/cobra"
)

var macCmd = &cobra.Command{
	Use:     "mac",
	Short:   "macOS 플랫폼 전환 및 진단",
	GroupID: "tools",
	Long: `macOS 플랫폼에 맞게 settings.json PATH를 재구성하고
시스템 환경을 진단합니다.

검증 항목:
  - Homebrew 설치 경로 (Intel /usr/local vs Apple Silicon /opt/homebrew)
  - 심볼릭 링크 해석 (node, python3 등)
  - 셸 호환성 (zsh, bash)
  - 도구 버전 (ae, go, node, git)`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runPlatformCommand(cmd, "darwin")
	},
}

func init() {
	rootCmd.AddCommand(macCmd)
	addPlatformFlags(macCmd)
}

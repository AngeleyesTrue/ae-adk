package cli

import (
	"github.com/spf13/cobra"
)

var winCmd = &cobra.Command{
	Use:     "win",
	Short:   "Windows 플랫폼 전환 및 진단",
	GroupID: "tools",
	Long: `Windows 플랫폼에 맞게 settings.json PATH를 재구성하고
시스템 환경을 진단합니다.

검증 항목:
  - UTF-8 코드페이지 (chcp 65001)
  - MCP 서버 경로 (npx, pwsh 등)
  - Git Bash 환경 감지
  - WSL2 상태 확인
  - 260자 경로 길이 제한 (LongPathsEnabled)
  - Hook 실행을 위한 bash 존재 확인
  - 도구 버전 (ae, go, node, git)`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runPlatformCommand(cmd, "windows")
	},
}

func init() {
	rootCmd.AddCommand(winCmd)
	addPlatformFlags(winCmd)
}

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Version 应用程序的版本号
var Version = "dev"

// CommitHash Git 的提交哈希
var CommitHash = "unknown"

// BuildTime 构建时间
var BuildTime = "unknown"

func printVersion() {
	fmt.Printf("Doghole Version: %s\n", Version)
	fmt.Printf("Git Commit Hash: %s\n", CommitHash)
	fmt.Printf("Build Time: %s\n", BuildTime)
}

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "打印应用程序的版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion()
		},
	})
}

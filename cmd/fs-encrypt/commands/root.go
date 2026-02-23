package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fs-encrypt",
	Short: "A tool to encrypt and compress directories",
	Long:  `fs-encrypt is a CLI tool that encrypts a directory recursively with compression.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

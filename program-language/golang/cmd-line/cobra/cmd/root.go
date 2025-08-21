package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "git",
	Short:   "A simple git demo",
	Long:    "usage: git [-v | --version] [-h | --help]",
	Version: "1.0.0",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello, this is git demo")
	},
}

func Execute() error {
	return rootCmd.Execute()
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:     "merge",
	Short:   "git merge demo",
	Long:    "usage: git merge [-v] [-h] [-p] [-n]",
	Version: "1.0.1",
	Run:     mergeFunc,
}

func mergeFunc(cmd *cobra.Command, args []string) {
	noFF, _ := cmd.Flags().GetBool("no-ff")
	if noFF {
		fmt.Println("Merge with --no-ff")
	}
}

func init() {
	rootCmd.AddCommand(mergeCmd)

	mergeCmd.Flags().BoolP("no-ff", "n", false, "do not fast-forward")
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "git clone",
	Long:  "git clone <url>",
	Run:   cloneFunc,
}

func cloneFunc(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println("must provide github url")
	}

	fmt.Println("run git clone", args[0])
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}

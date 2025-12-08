package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func BindGetVersionCommand(root *cobra.Command, version, commit, date string) error {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "get version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Build version: %s\n", ifNan(version))
			fmt.Printf("Build date: %s\n", ifNan(date))
			fmt.Printf("Build commit: %s\n", ifNan(commit))
			return nil
		},
	}
	root.AddCommand(cmd)
	return nil
}

func ifNan(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}

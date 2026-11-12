package cli

import (
	"github.com/spf13/cobra"
)

func newLogsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logs <workload>",
		Short: "Tail the logs of a workload",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return notImplemented("Phase 5")
		},
	}
}

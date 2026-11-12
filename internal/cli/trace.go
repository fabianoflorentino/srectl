package cli

import (
	"github.com/spf13/cobra"
)

func newTraceCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "trace <pod>",
		Short: "Run stracectl against a pod",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return notImplemented("Phase 6")
		},
	}
}

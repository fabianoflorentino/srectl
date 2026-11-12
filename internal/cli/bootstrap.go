package cli

import (
	"github.com/spf13/cobra"
)

func newBootstrapCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "bootstrap",
		Short: "Install Argo CD and register the root app",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return notImplemented("Phase 2")
		},
	}
}

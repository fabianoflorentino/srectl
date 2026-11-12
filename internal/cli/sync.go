package cli

import (
	"github.com/spf13/cobra"
)

func newSyncCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sync [app]",
		Short: "Force Argo CD to sync (default: all apps)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return notImplemented("Phase 2")
		},
	}
}

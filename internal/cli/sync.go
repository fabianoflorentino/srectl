package cli

import (
	"github.com/spf13/cobra"
)

func newSyncCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sync [app]",
		Short: "Force Argo CD to refresh an app (default: all)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := ""
			if len(args) == 1 {
				app = args[0]
			}
			if err := newArgocd(cmd).Sync(cmd.Context(), app); err != nil {
				return err
			}
			if app == "" {
				return writef(cmd.OutOrStdout(), "refresh triggered for all apps\n")
			}
			return writef(cmd.OutOrStdout(), "refresh triggered for %s\n", app)
		},
	}
}

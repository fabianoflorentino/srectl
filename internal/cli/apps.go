package cli

import (
	"github.com/spf13/cobra"
)

func newAppsCommand() *cobra.Command {
	apps := &cobra.Command{
		Use:   "apps",
		Short: "Deploy, restart and scale the demo apps",
	}
	apps.AddCommand(
		stubCommand("deploy", "Deploy the demo apps", "Phase 5"),
		stubCommand("restart", "Restart the demo apps", "Phase 5"),
		stubCommand("scale", "Scale a demo app", "Phase 5"),
	)
	return apps
}

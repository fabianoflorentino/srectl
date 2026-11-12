package cli

import (
	"github.com/spf13/cobra"
)

// openServices are the services reachable via `srectl open`.
var openServices = []string{"grafana", "argocd", "api-a", "api-b", "fraudctl", "certificate-validate"}

func newOpenCommand() *cobra.Command {
	return &cobra.Command{
		Use:       "open <service>",
		Short:     "Port-forward a service (grafana, argocd, api-a, ...)",
		Args:      cobra.ExactArgs(1),
		ValidArgs: openServices,
		RunE: func(cmd *cobra.Command, args []string) error {
			return notImplemented("Phase 1")
		},
	}
}

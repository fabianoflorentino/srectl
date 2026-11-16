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
			if err := newArgocd(cmd).Bootstrap(cmd.Context()); err != nil {
				return err
			}
			return writef(cmd.OutOrStdout(), "Argo CD bootstrapped\n")
		},
	}
}

package cli

import (
	"github.com/spf13/cobra"

	"github.com/fabianoflorentino/srectl/internal/kind"
)

func newUpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Create the kind cluster and bootstrap the platform",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := configFrom(cmd)
			client := kind.New(kind.WithStdout(cmd.OutOrStdout()), kind.WithStderr(cmd.ErrOrStderr()))

			if err := client.Create(cmd.Context(), kind.Cluster{Name: cfg.Cluster.Name, Nodes: cfg.Cluster.Nodes}); err != nil {
				return err
			}
			return writef(cmd.OutOrStdout(), "cluster ready — bootstrap (Argo CD) lands in Phase 2\n")
		},
	}
}

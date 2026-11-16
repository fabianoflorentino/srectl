package cli

import (
	"github.com/spf13/cobra"
)

func newUpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Create the kind cluster and bootstrap the platform via Argo CD",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := configFrom(cmd)
			mgr := newManager(cmd)
			if err := mgr.Up(cmd.Context(), mgr.ClusterSpec(cfg.Cluster.Name, cfg.Cluster.Nodes)); err != nil {
				return err
			}
			return writef(cmd.OutOrStdout(), "platform ready — Argo CD is reconciling deploy/\n")
		},
	}
}

package cli

import (
	"github.com/spf13/cobra"

	"github.com/fabianoflorentino/srectl/internal/k8s"
	"github.com/fabianoflorentino/srectl/internal/kind"
	"github.com/fabianoflorentino/srectl/internal/platform"
)

func newUpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Create the kind cluster and bootstrap the platform",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := configFrom(cmd)
			mgr := platform.New(
				kind.New(kind.WithStdout(cmd.OutOrStdout()), kind.WithStderr(cmd.ErrOrStderr())),
				k8s.New(k8s.WithStdout(cmd.OutOrStdout()), k8s.WithStderr(cmd.ErrOrStderr())),
			)
			if err := mgr.Up(cmd.Context(), mgr.ClusterSpec(cfg.Cluster.Name, cfg.Cluster.Nodes)); err != nil {
				return err
			}
			return writef(cmd.OutOrStdout(), "platform ready — hello at https://hello.localhost\n")
		},
	}
}

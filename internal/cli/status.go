package cli

import (
	"github.com/spf13/cobra"

	"github.com/fabianoflorentino/srectl/internal/k8s"
	"github.com/fabianoflorentino/srectl/internal/kind"
	"github.com/fabianoflorentino/srectl/internal/platform"
)

func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the component health matrix",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := configFrom(cmd)
			mgr := platform.New(
				kind.New(kind.WithStdout(cmd.OutOrStdout()), kind.WithStderr(cmd.ErrOrStderr())),
				k8s.New(k8s.WithStdout(cmd.OutOrStdout()), k8s.WithStderr(cmd.ErrOrStderr())),
			)
			report, err := mgr.Status(cmd.Context(), cfg.Cluster.Name)
			if err != nil {
				return err
			}
			return report.Write(cmd.OutOrStdout())
		},
	}
}

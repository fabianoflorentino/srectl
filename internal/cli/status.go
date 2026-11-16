package cli

import (
	"github.com/spf13/cobra"
)

func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the component health matrix",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := configFrom(cmd)
			mgr := newManager(cmd)
			report, err := mgr.Status(cmd.Context(), cfg.Cluster.Name)
			if err != nil {
				return err
			}
			return report.Write(cmd.OutOrStdout())
		},
	}
}

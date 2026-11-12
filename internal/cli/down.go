package cli

import (
	"github.com/spf13/cobra"

	"github.com/fabianoflorentino/srectl/internal/kind"
)

func newDownCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "Tear down the kind cluster",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := configFrom(cmd)
			client := kind.New(kind.WithStdout(cmd.OutOrStdout()), kind.WithStderr(cmd.ErrOrStderr()))

			if err := client.Delete(cmd.Context(), cfg.Cluster.Name); err != nil {
				return err
			}
			return writef(cmd.OutOrStdout(), "cluster %s deleted\n", cfg.Cluster.Name)
		},
	}
}

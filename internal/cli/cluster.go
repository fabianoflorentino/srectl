package cli

import (
	"github.com/spf13/cobra"

	"github.com/fabianoflorentino/srectl/internal/kind"
)

func newClusterCommand() *cobra.Command {
	cluster := &cobra.Command{
		Use:   "cluster",
		Short: "Manage the kind cluster",
	}
	cluster.AddCommand(
		newClusterCreateCommand(),
		newClusterDeleteCommand(),
		newClusterStatusCommand(),
	)
	return cluster
}

func newClusterCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create the kind cluster",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := configFrom(cmd)
			client := kind.New(kind.WithStdout(cmd.OutOrStdout()), kind.WithStderr(cmd.ErrOrStderr()))
			return client.Create(cmd.Context(), kind.Cluster{Name: cfg.Cluster.Name, Nodes: cfg.Cluster.Nodes})
		},
	}
}

func newClusterDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete the kind cluster",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := configFrom(cmd)
			client := kind.New(kind.WithStdout(cmd.OutOrStdout()), kind.WithStderr(cmd.ErrOrStderr()))
			return client.Delete(cmd.Context(), cfg.Cluster.Name)
		},
	}
}

func newClusterStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Report whether the kind cluster is up",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := configFrom(cmd)
			client := kind.New(kind.WithStdout(cmd.OutOrStdout()), kind.WithStderr(cmd.ErrOrStderr()))

			exists, err := client.Exists(cmd.Context(), cfg.Cluster.Name)
			if err != nil {
				return err
			}
			if !exists {
				return writef(cmd.OutOrStdout(), "cluster %s: down\n", cfg.Cluster.Name)
			}

			nodes, err := client.Nodes(cmd.Context(), cfg.Cluster.Name)
			if err != nil {
				return err
			}
			return writef(cmd.OutOrStdout(), "cluster %s: up (%d nodes)\n", cfg.Cluster.Name, len(nodes))
		},
	}
}

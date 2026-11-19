package cli

import (
	"github.com/spf13/cobra"

	"github.com/fabianoflorentino/srectl/internal/k8s"
	"github.com/fabianoflorentino/srectl/internal/slo"
)

func newSLOCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slo",
		Short: "Compute SLOs and inspect the error-budget burn rate (slo-generator)",
	}
	cmd.AddCommand(newSLOComputeCommand(), newSLOStatusCommand())
	return cmd
}

func newSLOComputeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "compute",
		Short: "Run a one-off slo-generator Job and wait for it to complete",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := configFrom(cmd)
			client := slo.New(k8s.New(
				k8s.WithStdout(cmd.OutOrStdout()),
				k8s.WithStderr(cmd.ErrOrStderr()),
			))
			name, err := client.Compute(cmd.Context(), cfg.SLO.Namespace, cfg.SLO.CronJob)
			if err != nil {
				return err
			}
			return writef(cmd.OutOrStdout(),
				"slo: job %s completed — run `srectl slo status`\n", name)
		},
	}
}

func newSLOStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the exported burn rate and threshold for every SLO",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := configFrom(cmd)
			client := slo.New(k8s.New(
				k8s.WithStdout(cmd.OutOrStdout()),
				k8s.WithStderr(cmd.ErrOrStderr()),
			))
			rows, err := client.Status(cmd.Context(), cfg.SLO.Namespace, cfg.SLO.Service, cfg.SLO.Port)
			if err != nil {
				return err
			}
			return slo.Write(cmd.OutOrStdout(), rows)
		},
	}
}

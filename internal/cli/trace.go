package cli

import (
	"github.com/spf13/cobra"

	"github.com/fabianoflorentino/srectl/internal/k8s"
)

func newTraceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trace [target]",
		Short: "Open the stracectl live dashboard (port-forward)",
		Long: "Port-forward the stracectl demo pod's dashboard. stracectl runs as a\n" +
			"sidecar with CAP_SYS_PTRACE and exports per-syscall metrics plus a live\n" +
			"view at /api/stats and /metrics.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			daemon, err := cmd.Flags().GetBool("daemon")
			if err != nil {
				return err
			}
			down, err := cmd.Flags().GetBool("down")
			if err != nil {
				return err
			}
			if down {
				return stopDaemon("stracectl")
			}

			const namespace, resource, port, localPort = "projects", "svc/stracectl", 8080, 8080
			const url = "http://localhost:8080"

			client := k8s.New(k8s.WithStdout(cmd.OutOrStdout()), k8s.WithStderr(cmd.ErrOrStderr()))
			if daemon {
				pid, err := client.PortForwardDaemon(namespace, resource, localPort, port)
				if err != nil {
					return err
				}
				if err := writePIDFile("stracectl", pid); err != nil {
					return err
				}
				return writef(cmd.OutOrStdout(),
					"stracectl dashboard running in background (pid %d) → %s\n", pid, url)
			}

			if err := writef(cmd.OutOrStdout(),
				"stracectl live view → %s (Ctrl+C to stop)\n", url); err != nil {
				return err
			}
			return client.PortForward(cmd.Context(), namespace, resource, localPort, port)
		},
	}
	cmd.Flags().BoolP("daemon", "d", false, "run the port-forward in the background")
	cmd.Flags().Bool("down", false, "stop a background port-forward")
	return cmd
}

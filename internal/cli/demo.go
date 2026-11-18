package cli

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/fabianoflorentino/srectl/internal/chaos"
	"github.com/fabianoflorentino/srectl/internal/demo"
	"github.com/fabianoflorentino/srectl/internal/k8s"
	"github.com/fabianoflorentino/srectl/internal/verify"
)

func newDemoCommand() *cobra.Command {
	demoCmd := &cobra.Command{
		Use:   "demo",
		Short: "Run a demo scenario (load, chaos, verify, error, security, certs, projects)",
	}
	demoCmd.AddCommand(newDemoLoadCommand(), newDemoChaosCommand(), newDemoVerifyCommand())

	scenarios := []struct {
		use      string
		short    string
		scenario demo.Scenario
	}{
		{"break", "Break a component", demo.ScenarioBreak},
		{"error", "Force errors to fire the high-error-rate alert", demo.ScenarioError},
		{"security", "Show trivy report, kyverno violation and falco event", demo.ScenarioSecurity},
		{"certs", "Run the certificate-validate scan", demo.ScenarioCerts},
		{"projects", "Exercise fraudctl, stracectl and certificate-validate", demo.ScenarioProjects},
	}
	for _, s := range scenarios {
		scenario := s
		demoCmd.AddCommand(&cobra.Command{
			Use:   scenario.use,
			Short: scenario.short,
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return demo.Run(scenario.scenario)
			},
		})
	}

	return demoCmd
}

func newDemoLoadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load",
		Short: "Generate traffic on the demo apps",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			url, err := cmd.Flags().GetString("url")
			if err != nil {
				return err
			}
			duration, err := cmd.Flags().GetDuration("duration")
			if err != nil {
				return err
			}
			rps, err := cmd.Flags().GetInt("rps")
			if err != nil {
				return err
			}

			result, err := demo.Load(cmd.Context(), demo.LoadConfig{
				URL:      url,
				Duration: duration,
				RPS:      rps,
				Client:   insecureClient(),
			})
			if err != nil {
				return err
			}
			return writef(cmd.OutOrStdout(), "load: %d requests, %d errors in %s\n",
				result.Requests, result.Errors, duration)
		},
	}
	cmd.Flags().String("url", "https://api-a.localhost", "base URL of api-a")
	cmd.Flags().Duration("duration", 30*time.Second, "load duration")
	cmd.Flags().Int("rps", 20, "approximate requests per second")
	return cmd
}

func newDemoChaosCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chaos [latency|pod-kill]",
		Short: "Inject chaos (latency via CPU stress, or pod-kill) and clean up",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scenario := chaos.ScenarioLatency
			if len(args) > 0 {
				scenario = chaos.Scenario(args[0])
			}
			duration, err := cmd.Flags().GetDuration("duration")
			if err != nil {
				return err
			}

			client := k8s.New(k8s.WithStdout(cmd.OutOrStdout()), k8s.WithStderr(cmd.ErrOrStderr()))
			if err := writef(cmd.OutOrStdout(), "injecting chaos %q for %s ...\n", scenario, duration); err != nil {
				return err
			}
			if err := chaos.Run(cmd.Context(), client, scenario, duration); err != nil {
				return err
			}
			return writef(cmd.OutOrStdout(), "chaos %q recovered\n", scenario)
		},
	}
	cmd.Flags().Duration("duration", 60*time.Second, "chaos duration")
	return cmd
}

func newDemoVerifyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Assert the cluster, Argo CD and every app are healthy",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := configFrom(cmd)
			report, err := newManager(cmd).Status(cmd.Context(), cfg.Cluster.Name)
			if err != nil {
				return err
			}
			if err := verify.Assess(report); err != nil {
				_ = report.Write(cmd.OutOrStdout())
				return err
			}
			return writef(cmd.OutOrStdout(),
				"verify: OK — cluster up, argocd ready, %d apps Synced · Healthy\n", len(report.Apps))
		},
	}
}

// insecureClient returns an HTTP client that tolerates the local self-signed CA.
func insecureClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			//nolint:gosec // local training cluster uses a self-signed CA
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

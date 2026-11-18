package cli

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/fabianoflorentino/srectl/internal/demo"
)

func newDemoCommand() *cobra.Command {
	demoCmd := &cobra.Command{
		Use:   "demo",
		Short: "Run a demo scenario (load, chaos, break, verify, error, security, certs, projects)",
	}
	demoCmd.AddCommand(newDemoLoadCommand())

	scenarios := []struct {
		use      string
		short    string
		scenario demo.Scenario
	}{
		{"chaos", "Inject latency and kill pods", demo.ScenarioChaos},
		{"break", "Break a component", demo.ScenarioBreak},
		{"verify", "Assert the stack recovers and alerts fire", demo.ScenarioVerify},
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

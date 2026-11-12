package cli

import (
	"github.com/spf13/cobra"

	"github.com/fabianoflorentino/srectl/internal/demo"
)

func newDemoCommand() *cobra.Command {
	demoCmd := &cobra.Command{
		Use:   "demo",
		Short: "Run a demo scenario (load, chaos, break, verify, error, security, certs, projects)",
	}

	scenarios := []struct {
		use      string
		short    string
		scenario demo.Scenario
	}{
		{"load", "Generate traffic on the demo apps", demo.ScenarioLoad},
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

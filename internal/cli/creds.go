package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newCredsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "creds",
		Short: "Show platform admin credentials (Argo CD, Grafana)",
		Long: "Print the admin credentials of the components that authenticate.\n\n" +
			"Argo CD and Grafana require a login; Loki and Tempo run without\n" +
			"authentication in this local stack, so they have no credentials.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listCreds(cmd)
		},
	}
	cmd.AddCommand(newCredsRotateCommand())
	return cmd
}

// listCreds prints the admin credentials of every credential-bearing component.
func listCreds(cmd *cobra.Command) error {
	argocdPassword, err := newArgocd(cmd).Password(cmd.Context())
	if err != nil {
		return err
	}

	gf := newGrafana(cmd)
	grafanaUser, err := gf.Username(cmd.Context())
	if err != nil {
		return err
	}
	grafanaPassword, err := gf.Password(cmd.Context())
	if err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("component  username  password\n")
	b.WriteString("argocd     admin     " + argocdPassword + "\n")
	b.WriteString("grafana    " + grafanaUser + "     " + grafanaPassword + "\n")
	return writef(cmd.OutOrStdout(), "%s", b.String())
}

func newCredsRotateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "rotate <argocd|grafana>",
		Short:     "Rotate a component's admin password",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"argocd", "grafana"},
		RunE: func(cmd *cobra.Command, args []string) error {
			password, err := cmd.Flags().GetString("password")
			if err != nil {
				return err
			}

			switch args[0] {
			case "argocd":
				newPassword, err := newArgocd(cmd).RotatePassword(cmd.Context(), password)
				if err != nil {
					return err
				}
				return writef(cmd.OutOrStdout(), "Argo CD admin password: %s\n", newPassword)
			case "grafana":
				return fmt.Errorf("grafana's admin password is managed by GitOps (grafana.adminPassword " +
					"in deploy/bootstrap/apps/kube-prometheus-stack-app.yaml); change it there and sync")
			default:
				return fmt.Errorf("unknown component %q (want argocd or grafana)", args[0])
			}
		},
	}
	cmd.Flags().String("password", "", "new admin password (default: generate a random one)")
	return cmd
}

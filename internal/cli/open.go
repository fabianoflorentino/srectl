package cli

import (
	"github.com/spf13/cobra"

	"github.com/fabianoflorentino/srectl/internal/k8s"
)

// openTarget describes how to port-forward a service.
type openTarget struct {
	namespace string
	resource  string
	port      int
	localPort int
}

// openTargets are the non-Argo CD services reachable via `srectl open`.
var openTargets = map[string]openTarget{
	"hello": {namespace: "demo", resource: "svc/hello", port: 80, localPort: 8081},
}

func newOpenCommand() *cobra.Command {
	return &cobra.Command{
		Use:       "open <service>",
		Short:     "Port-forward a service (argocd, hello)",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"argocd", "hello"},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := k8s.New(k8s.WithStdout(cmd.OutOrStdout()), k8s.WithStderr(cmd.ErrOrStderr()))

			if args[0] == "argocd" {
				ac := newArgocd(cmd)
				return client.PortForward(cmd.Context(), ac.Namespace(), "svc/"+ac.ServerService(), 8080, 443)
			}

			target, ok := openTargets[args[0]]
			if !ok {
				return notImplemented("Phase 3")
			}
			return client.PortForward(cmd.Context(), target.namespace, target.resource, target.localPort, target.port)
		},
	}
}

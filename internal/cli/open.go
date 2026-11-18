package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/fabianoflorentino/srectl/internal/k8s"
)

// openTarget describes how to port-forward a service.
type openTarget struct {
	namespace string
	resource  string
	port      int
	localPort int
	url       string
}

// openTargets are the non-Argo CD services reachable via `srectl open`.
var openTargets = map[string]openTarget{
	"hello":        {namespace: "demo", resource: "svc/hello", port: 80, localPort: 8081, url: "http://localhost:8081"},
	"api-a":        {namespace: "demo", resource: "svc/api-a", port: 80, localPort: 8082, url: "http://localhost:8082"},
	"api-b":        {namespace: "demo", resource: "svc/api-b", port: 80, localPort: 8083, url: "http://localhost:8083"},
	"fraudctl":     {namespace: "projects", resource: "svc/fraudctl", port: 80, localPort: 8084, url: "http://localhost:8084"},
	"seraph":       {namespace: "projects", resource: "svc/seraph", port: 80, localPort: 8085, url: "http://localhost:8085"},
	"grafana":      {namespace: "monitoring", resource: "svc/kube-prometheus-stack-grafana", port: 80, localPort: 3000, url: "http://localhost:3000"},
	"prometheus":   {namespace: "monitoring", resource: "svc/kube-prometheus-stack-prometheus", port: 9090, localPort: 9090, url: "http://localhost:9090"},
	"alertmanager": {namespace: "monitoring", resource: "svc/kube-prometheus-stack-alertmanager", port: 9093, localPort: 9093, url: "http://localhost:9093"},
}

func newOpenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "open <service>",
		Short:     "Port-forward a service (argocd, grafana, prometheus, alertmanager, hello, api-a, api-b, fraudctl, seraph)",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"argocd", "grafana", "prometheus", "alertmanager", "hello", "api-a", "api-b", "fraudctl", "seraph"},
		RunE: func(cmd *cobra.Command, args []string) error {
			service := args[0]

			daemon, err := cmd.Flags().GetBool("daemon")
			if err != nil {
				return err
			}
			down, err := cmd.Flags().GetBool("down")
			if err != nil {
				return err
			}

			if down {
				return stopDaemon(service)
			}

			target, ok := resolveOpenTarget(cmd, service)
			if !ok {
				return notImplemented("Phase 5")
			}

			client := k8s.New(k8s.WithStdout(cmd.OutOrStdout()), k8s.WithStderr(cmd.ErrOrStderr()))

			if daemon {
				pid, err := client.PortForwardDaemon(target.namespace, target.resource, target.localPort, target.port)
				if err != nil {
					return err
				}
				if err := writePIDFile(service, pid); err != nil {
					return err
				}
				return writef(cmd.OutOrStdout(), "%s port-forward running in background (pid %d) → %s\n", service, pid, target.url)
			}

			return client.PortForward(cmd.Context(), target.namespace, target.resource, target.localPort, target.port)
		},
	}
	cmd.Flags().BoolP("daemon", "d", false, "run the port-forward in the background")
	cmd.Flags().Bool("down", false, "stop a background port-forward")
	return cmd
}

// resolveOpenTarget returns the port-forward target for a service. Argo CD's
// namespace and service name come from the Argo CD config.
func resolveOpenTarget(cmd *cobra.Command, service string) (openTarget, bool) {
	if service == "argocd" {
		ac := newArgocd(cmd)
		return openTarget{
			namespace: ac.Namespace(),
			resource:  "svc/" + ac.ServerService(),
			port:      443,
			localPort: 8080,
			url:       "https://localhost:8080",
		}, true
	}
	t, ok := openTargets[service]
	return t, ok
}

func pidFilePath(service string) string {
	return filepath.Join(os.TempDir(), "srectl-port-forward-"+service+".pid")
}

func writePIDFile(service string, pid int) error {
	return os.WriteFile(pidFilePath(service), []byte(strconv.Itoa(pid)), 0o600)
}

func readPIDFile(service string) (int, error) {
	data, err := os.ReadFile(pidFilePath(service))
	if err != nil {
		return 0, fmt.Errorf("no background port-forward for %q: %w", service, err)
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

func stopDaemon(service string) error {
	pid, err := readPIDFile(service)
	if err != nil {
		return err
	}
	if err := k8s.StopPortForward(pid); err != nil {
		return err
	}
	return os.Remove(pidFilePath(service))
}

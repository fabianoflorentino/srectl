package cli

import (
	"github.com/spf13/cobra"

	"github.com/fabianoflorentino/srectl/internal/argocd"
	"github.com/fabianoflorentino/srectl/internal/grafana"
	"github.com/fabianoflorentino/srectl/internal/k8s"
	"github.com/fabianoflorentino/srectl/internal/kind"
	"github.com/fabianoflorentino/srectl/internal/platform"
)

// newManager wires the platform manager to the command's output streams.
func newManager(cmd *cobra.Command) *platform.Manager {
	cfg := configFrom(cmd)
	kc := kind.New(kind.WithStdout(cmd.OutOrStdout()), kind.WithStderr(cmd.ErrOrStderr()))
	k8c := k8s.New(k8s.WithStdout(cmd.OutOrStdout()), k8s.WithStderr(cmd.ErrOrStderr()))
	ac := argocd.New(k8c, argocd.Config{
		Namespace: cfg.Argocd.Namespace,
		Version:   cfg.Argocd.Version,
		Repo:      cfg.Argocd.Repo,
		Revision:  cfg.Argocd.Revision,
	})
	return platform.New(kc, k8c, ac)
}

// newArgocd wires an Argo CD client to the command's output streams.
func newArgocd(cmd *cobra.Command) *argocd.Client {
	cfg := configFrom(cmd)
	k8c := k8s.New(k8s.WithStdout(cmd.OutOrStdout()), k8s.WithStderr(cmd.ErrOrStderr()))
	return argocd.New(k8c, argocd.Config{
		Namespace: cfg.Argocd.Namespace,
		Version:   cfg.Argocd.Version,
		Repo:      cfg.Argocd.Repo,
		Revision:  cfg.Argocd.Revision,
	})
}

// newGrafana wires a Grafana client to the command's output streams.
func newGrafana(cmd *cobra.Command) *grafana.Client {
	cfg := configFrom(cmd)
	k8c := k8s.New(k8s.WithStdout(cmd.OutOrStdout()), k8s.WithStderr(cmd.ErrOrStderr()))
	return grafana.New(k8c, grafana.Config{
		Namespace: cfg.Monitoring.Namespace,
		Secret:    cfg.Monitoring.Grafana.Secret,
	})
}

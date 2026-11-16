// Package platform orchestrates the whole srectl lifecycle: the kind cluster,
// Argo CD bootstrap, and the health matrix that reports on both.
package platform

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/fabianoflorentino/srectl/internal/argocd"
	"github.com/fabianoflorentino/srectl/internal/k8s"
	"github.com/fabianoflorentino/srectl/internal/kind"
)

// kindConfigFile is the repo-relative path to the kind cluster topology.
const kindConfigFile = "deploy/kind/kind-config.yaml"

// KindConfigFile returns the repo kind config path when it exists.
func KindConfigFile() string {
	if _, err := os.Stat(kindConfigFile); err == nil {
		return kindConfigFile
	}
	return ""
}

// Manager provisions the cluster and drives Argo CD.
type Manager struct {
	kind   *kind.Client
	k8s    *k8s.Client
	argocd *argocd.Client
}

// New returns a Manager wired to the given clients.
func New(kc *kind.Client, k8c *k8s.Client, ac *argocd.Client) *Manager {
	return &Manager{kind: kc, k8s: k8c, argocd: ac}
}

// ClusterSpec builds the kind cluster spec, using the repo config file when present.
func (m *Manager) ClusterSpec(name string, nodes int) kind.Cluster {
	return kind.Cluster{Name: name, Nodes: nodes, ConfigFile: KindConfigFile()}
}

// Up creates the cluster (if needed) and bootstraps Argo CD, which reconciles
// the whole platform from deploy/.
func (m *Manager) Up(ctx context.Context, cluster kind.Cluster) error {
	exists, err := m.kind.Exists(ctx, cluster.Name)
	if err != nil {
		return fmt.Errorf("checking kind cluster: %w", err)
	}
	if !exists {
		if err := m.kind.Create(ctx, cluster); err != nil {
			return fmt.Errorf("creating kind cluster: %w", err)
		}
	}
	if err := m.argocd.Bootstrap(ctx); err != nil {
		return err
	}
	return nil
}

// Report is the platform health matrix.
type Report struct {
	Cluster string
	Argocd  string
	Apps    []argocd.App
}

// Write renders the report as aligned rows.
func (r Report) Write(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%-13s %s\n", "cluster:", r.Cluster); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%-13s %s\n", "argocd:", r.Argocd); err != nil {
		return err
	}
	for _, app := range r.Apps {
		if _, err := fmt.Fprintf(w, "%-13s %s · %s\n", app.Name+":", app.Sync, app.Health); err != nil {
			return err
		}
	}
	return nil
}

// Status builds the health matrix: the kind cluster, Argo CD itself, and every
// application Argo CD manages.
func (m *Manager) Status(ctx context.Context, name string) (Report, error) {
	r := Report{Cluster: "down", Argocd: "NotReady"}

	exists, err := m.kind.Exists(ctx, name)
	if err != nil {
		return Report{}, fmt.Errorf("checking cluster: %w", err)
	}
	if !exists {
		return r, nil
	}

	nodes, err := m.kind.Nodes(ctx, name)
	if err != nil {
		return Report{}, fmt.Errorf("listing nodes: %w", err)
	}
	r.Cluster = fmt.Sprintf("up (%d nodes)", len(nodes))

	ok, err := m.k8s.DeploymentAvailable(ctx, m.argocd.Namespace(), m.argocd.ServerDeployment())
	if err != nil {
		return Report{}, fmt.Errorf("checking argocd: %w", err)
	}
	if ok {
		r.Argocd = "Ready"
	}

	apps, err := m.argocd.Apps(ctx)
	if err == nil {
		r.Apps = apps
	}
	return r, nil
}

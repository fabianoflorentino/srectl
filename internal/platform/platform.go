// Package platform orchestrates the Phase 1 surface: the kind cluster,
// ingress-nginx, cert-manager and a TLS-terminated hello app.
package platform

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fabianoflorentino/srectl/internal/k8s"
	"github.com/fabianoflorentino/srectl/internal/kind"
)

// Pinned chart versions. Phase 1 installs these imperatively; from Phase 2 on,
// they move to GitOps and the versions live in deploy/.
const (
	IngressNGINXVersion = "4.15.1"
	CertManagerVersion  = "1.21.3"
)

// helmRepos are the chart repositories the manager requires.
var helmRepos = []struct{ name, url string }{
	{"ingress-nginx", "https://kubernetes.github.io/ingress-nginx"},
	{"jetstack", "https://charts.jetstack.io"},
}

// Paths locates the deploy/ artifacts the manager applies.
type Paths struct {
	KindConfig    string
	IngressValues string
	Namespace     string
	Issuers       []string
	HelloDir      string
}

// DefaultPaths resolves against the repo root (srectl is a repo-root task runner).
var DefaultPaths = Paths{
	KindConfig:    "deploy/kind/kind-config.yaml",
	IngressValues: "deploy/ingress/ingress-nginx-values.yaml",
	Namespace:     "deploy/apps/hello/namespace.yaml",
	Issuers: []string{
		"deploy/ingress/cluster-issuer.yaml",
		"deploy/ingress/ca.yaml",
	},
	HelloDir: "deploy/apps/hello/",
}

// KindConfigFile returns the repo kind config path when it exists.
func KindConfigFile() string {
	if _, err := os.Stat(DefaultPaths.KindConfig); err == nil {
		return DefaultPaths.KindConfig
	}
	return ""
}

// Manager provisions and inspects the Phase 1 platform.
type Manager struct {
	kind  *kind.Client
	k8s   *k8s.Client
	paths Paths
}

// Option configures a Manager.
type Option func(*Manager)

// WithPaths overrides the deploy artifact locations (useful for tests).
func WithPaths(p Paths) Option {
	return func(m *Manager) { m.paths = p }
}

// New returns a Manager wired to the given clients and default paths.
func New(kc *kind.Client, k8c *k8s.Client, opts ...Option) *Manager {
	m := &Manager{kind: kc, k8s: k8c, paths: DefaultPaths}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// ClusterSpec builds the kind cluster spec, using the repo config file when present.
func (m *Manager) ClusterSpec(name string, nodes int) kind.Cluster {
	return kind.Cluster{Name: name, Nodes: nodes, ConfigFile: KindConfigFile()}
}

// Up provisions the whole Phase 1 platform, idempotently.
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
	if err := m.AddRepos(ctx); err != nil {
		return err
	}
	if err := m.InstallIngress(ctx); err != nil {
		return err
	}
	if err := m.InstallCertManager(ctx); err != nil {
		return err
	}
	if err := m.EnsureNamespace(ctx); err != nil {
		return err
	}
	if err := m.ApplyIssuers(ctx); err != nil {
		return err
	}
	if err := m.DeployHello(ctx); err != nil {
		return err
	}
	if err := m.WaitReady(ctx); err != nil {
		return err
	}
	if err := m.WaitCertificate(ctx); err != nil {
		return err
	}
	return m.SyncTrustBundle(ctx)
}

// AddRepos ensures the required Helm chart repositories are present.
func (m *Manager) AddRepos(ctx context.Context) error {
	for _, r := range helmRepos {
		if err := m.k8s.Helm(ctx, "repo", "add", r.name, r.url, "--force-update"); err != nil {
			return fmt.Errorf("adding helm repo %s: %w", r.name, err)
		}
	}
	if err := m.k8s.Helm(ctx, "repo", "update"); err != nil {
		return fmt.Errorf("updating helm repos: %w", err)
	}
	return nil
}

// InstallIngress installs or upgrades ingress-nginx via Helm.
func (m *Manager) InstallIngress(ctx context.Context) error {
	args := []string{
		"upgrade", "--install", "ingress-nginx", "ingress-nginx/ingress-nginx",
		"--namespace", "ingress-nginx", "--create-namespace",
		"--version", IngressNGINXVersion,
		"-f", m.paths.IngressValues,
	}
	if err := m.k8s.Helm(ctx, args...); err != nil {
		return fmt.Errorf("installing ingress-nginx: %w", err)
	}
	return nil
}

// InstallCertManager installs or upgrades cert-manager (with CRDs) via Helm.
func (m *Manager) InstallCertManager(ctx context.Context) error {
	args := []string{
		"upgrade", "--install", "cert-manager", "jetstack/cert-manager",
		"--namespace", "cert-manager", "--create-namespace",
		"--version", CertManagerVersion,
		"--set", "crds.enabled=true",
	}
	if err := m.k8s.Helm(ctx, args...); err != nil {
		return fmt.Errorf("installing cert-manager: %w", err)
	}
	return nil
}

// ApplyIssuers applies the ClusterIssuer, root CA and trust bundle.
func (m *Manager) ApplyIssuers(ctx context.Context) error {
	if err := m.k8s.Apply(ctx, m.paths.Issuers...); err != nil {
		return fmt.Errorf("applying cert-manager issuers: %w", err)
	}
	return nil
}

// EnsureNamespace creates the demo namespace if it does not exist.
func (m *Manager) EnsureNamespace(ctx context.Context) error {
	if err := m.k8s.Apply(ctx, m.paths.Namespace); err != nil {
		return fmt.Errorf("creating demo namespace: %w", err)
	}
	return nil
}

// DeployHello deploys the hello app (namespace, deployment, service, ingress).
func (m *Manager) DeployHello(ctx context.Context) error {
	if err := m.k8s.Apply(ctx, m.paths.HelloDir); err != nil {
		return fmt.Errorf("deploying hello app: %w", err)
	}
	return nil
}

// WaitReady blocks until the ingress controller, cert-manager and hello app are rolled out.
func (m *Manager) WaitReady(ctx context.Context) error {
	wait := []struct{ namespace, resource string }{
		{"ingress-nginx", "deployment/ingress-nginx-controller"},
		{"cert-manager", "deployment/cert-manager"},
		{"demo", "deployment/hello"},
	}
	for _, w := range wait {
		if err := m.k8s.RolloutStatus(ctx, w.namespace, w.resource, "300s"); err != nil {
			return fmt.Errorf("waiting for %s: %w", w.resource, err)
		}
	}
	return nil
}

// WaitCertificate waits for cert-manager to issue the hello TLS certificate.
func (m *Manager) WaitCertificate(ctx context.Context) error {
	if err := m.k8s.Kubectl(ctx, "wait", "--for=condition=Ready", "certificate/hello-tls", "--namespace", "demo", "--timeout=120s"); err != nil {
		return fmt.Errorf("waiting for hello certificate: %w", err)
	}
	return nil
}

// SyncTrustBundle populates the trust-bundle ConfigMap with the root CA so the
// cluster CA can be mounted or installed into the host trust store.
func (m *Manager) SyncTrustBundle(ctx context.Context) error {
	out, err := m.k8s.Output(ctx, "get", "secret", "srectl-root-ca", "-n", "demo", "-o", "jsonpath={.data.ca\\.crt}")
	if err != nil {
		return fmt.Errorf("reading root CA: %w", err)
	}
	pem, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(out)))
	if err != nil {
		return fmt.Errorf("decoding root CA: %w", err)
	}

	caFile, err := os.CreateTemp("", "srectl-ca-*.crt")
	if err != nil {
		return fmt.Errorf("creating CA file: %w", err)
	}
	defer func() { _ = os.Remove(caFile.Name()) }()
	if _, err := caFile.Write(pem); err != nil {
		_ = caFile.Close()
		return fmt.Errorf("writing CA file: %w", err)
	}
	if err := caFile.Close(); err != nil {
		return fmt.Errorf("closing CA file: %w", err)
	}

	manifest, err := m.k8s.Output(ctx, "create", "configmap", "trust-bundle", "-n", "demo", "--from-file=ca.crt="+caFile.Name(), "--dry-run=client", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("generating trust bundle manifest: %w", err)
	}

	mf, err := os.CreateTemp("", "srectl-trust-bundle-*.yaml")
	if err != nil {
		return fmt.Errorf("creating manifest file: %w", err)
	}
	defer func() { _ = os.Remove(mf.Name()) }()
	if _, err := mf.Write(manifest); err != nil {
		_ = mf.Close()
		return fmt.Errorf("writing manifest file: %w", err)
	}
	if err := mf.Close(); err != nil {
		return fmt.Errorf("closing manifest file: %w", err)
	}

	if err := m.k8s.Apply(ctx, mf.Name()); err != nil {
		return fmt.Errorf("applying trust bundle: %w", err)
	}
	return nil
}

// Report is the Phase 1 health matrix.
type Report struct {
	Cluster     string
	Ingress     string
	CertManager string
	Hello       string
}

// Write renders the report as aligned key/value rows.
func (r Report) Write(w io.Writer) error {
	rows := []struct{ name, value string }{
		{"cluster", r.Cluster},
		{"ingress", r.Ingress},
		{"cert-manager", r.CertManager},
		{"hello", r.Hello},
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(w, "%-13s %s\n", row.name+":", row.value); err != nil {
			return err
		}
	}
	return nil
}

// Status builds the health matrix for cluster, ingress, cert-manager and hello.
func (m *Manager) Status(ctx context.Context, name string) (Report, error) {
	r := Report{Cluster: "down", Ingress: "NotReady", CertManager: "NotReady", Hello: "NotReady"}

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

	checks := []struct {
		namespace string
		name      string
		assign    func(*Report)
	}{
		{"ingress-nginx", "ingress-nginx-controller", func(r *Report) { r.Ingress = "Ready" }},
		{"cert-manager", "cert-manager", func(r *Report) { r.CertManager = "Ready" }},
		{"demo", "hello", func(r *Report) { r.Hello = "Ready" }},
	}
	for _, c := range checks {
		ok, err := m.k8s.DeploymentAvailable(ctx, c.namespace, c.name)
		if err != nil {
			return Report{}, fmt.Errorf("checking %s: %w", c.name, err)
		}
		if ok {
			c.assign(&r)
		}
	}
	return r, nil
}

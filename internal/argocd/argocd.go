// Package argocd bootstraps and operates Argo CD, the GitOps engine that
// reconciles deploy/ into the cluster.
package argocd

import (
	"bytes"
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/fabianoflorentino/srectl/internal/k8s"
)

//go:embed root-app.yaml
var embeddedRootApp string

// Config holds the Argo CD install and source settings.
type Config struct {
	Namespace string
	Version   string
	Repo      string
	Revision  string
}

// Paths locates the bootstrap artifacts.
type Paths struct {
	RootApp string
}

// Client drives Argo CD via kubectl/helm.
type Client struct {
	k8s   *k8s.Client
	cfg   Config
	paths Paths
}

// Option configures a Client.
type Option func(*Client)

// WithPaths overrides the bootstrap artifact locations.
func WithPaths(p Paths) Option {
	return func(c *Client) { c.paths = p }
}

// New returns an Argo CD client wired to the given k8s client.
func New(k8c *k8s.Client, cfg Config, opts ...Option) *Client {
	c := &Client{
		k8s:   k8c,
		cfg:   cfg,
		paths: Paths{RootApp: "deploy/bootstrap/argocd/root-app.yaml"},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Namespace returns the namespace Argo CD is installed into.
func (c *Client) Namespace() string {
	return c.cfg.Namespace
}

// serverDeployment is the argocd-server Deployment name produced by the
// argo/argo-cd chart (release name "argo-cd", fullnameOverride "argocd").
const serverDeployment = "argo-cd-argocd-server"

// ServerDeployment returns the Argo CD server Deployment name.
func (c *Client) ServerDeployment() string {
	return serverDeployment
}

// Bootstrap installs Argo CD (Helm), waits for it to be ready, and applies the
// rendered root application. This is the single imperative step of the platform.
func (c *Client) Bootstrap(ctx context.Context) error {
	if err := c.install(ctx); err != nil {
		return err
	}
	if err := c.wait(ctx); err != nil {
		return err
	}
	return c.applyRootApp(ctx)
}

func (c *Client) install(ctx context.Context) error {
	args := []string{
		"upgrade", "--install", "argo-cd", "argo/argo-cd",
		"--namespace", c.cfg.Namespace, "--create-namespace",
		"--version", c.cfg.Version,
	}
	if err := c.k8s.Helm(ctx, args...); err != nil {
		return fmt.Errorf("installing argo-cd: %w", err)
	}
	return nil
}

func (c *Client) wait(ctx context.Context) error {
	if err := c.k8s.RolloutStatus(ctx, c.cfg.Namespace, "deployment/"+c.ServerDeployment(), "300s"); err != nil {
		return fmt.Errorf("waiting for argocd-server: %w", err)
	}
	return nil
}

// applyRootApp renders the root application template with the configured repo
// and revision, then applies it.
func (c *Client) applyRootApp(ctx context.Context) error {
	manifest, err := c.renderRootApp()
	if err != nil {
		return err
	}
	f, err := os.CreateTemp("", "srectl-root-app-*.yaml")
	if err != nil {
		return fmt.Errorf("creating root app file: %w", err)
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if _, err := f.Write(manifest); err != nil {
		_ = f.Close()
		return fmt.Errorf("writing root app file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("closing root app file: %w", err)
	}

	if err := c.k8s.Apply(ctx, f.Name()); err != nil {
		return fmt.Errorf("applying root app: %w", err)
	}
	return nil
}

// renderRootApp renders the root application template with the configured repo
// and revision. It prefers the repo template file and falls back to an embedded
// copy (so the CLI works from `go install` without the repo checkout).
func (c *Client) renderRootApp() ([]byte, error) {
	content := embeddedRootApp
	if c.paths.RootApp != "" {
		if b, err := os.ReadFile(c.paths.RootApp); err == nil {
			content = string(b)
		}
	}

	tmpl, err := template.New("root-app").Parse(content)
	if err != nil {
		return nil, fmt.Errorf("parsing root app template: %w", err)
	}
	var buf bytes.Buffer
	data := struct{ Repo, Revision string }{Repo: c.cfg.Repo, Revision: c.cfg.Revision}
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering root app template: %w", err)
	}
	return buf.Bytes(), nil
}

// Sync triggers a hard refresh of the given application (or all when empty),
// which with automated sync causes Argo CD to reconcile immediately.
func (c *Client) Sync(ctx context.Context, app string) error {
	target := "--all"
	if app != "" {
		target = app
	}
	if err := c.k8s.Kubectl(ctx, "annotate", "application", target, "-n", c.cfg.Namespace, "argocd.argoproj.io/refresh=hard"); err != nil {
		return fmt.Errorf("refreshing application: %w", err)
	}
	return nil
}

// App is the health snapshot of a single Argo CD application.
type App struct {
	Name   string `json:"name"`
	Sync   string `json:"sync"`
	Health string `json:"health"`
}

type applicationList struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Status struct {
			Sync struct {
				Status string `json:"status"`
			} `json:"sync"`
			Health struct {
				Status string `json:"status"`
			} `json:"health"`
		} `json:"status"`
	} `json:"items"`
}

// Apps lists Argo CD applications with their sync and health status.
func (c *Client) Apps(ctx context.Context) ([]App, error) {
	out, err := c.k8s.Output(ctx, "get", "applications", "-n", c.cfg.Namespace, "-o", "json")
	if err != nil {
		return nil, fmt.Errorf("listing applications: %w", err)
	}

	var list applicationList
	if err := json.Unmarshal(out, &list); err != nil {
		return nil, fmt.Errorf("decoding applications: %w", err)
	}

	apps := make([]App, 0, len(list.Items))
	for _, item := range list.Items {
		apps = append(apps, App{
			Name:   item.Metadata.Name,
			Sync:   item.Status.Sync.Status,
			Health: item.Status.Health.Status,
		})
	}
	return apps, nil
}

// Password returns the Argo CD admin password from the initial admin secret.
func (c *Client) Password(ctx context.Context) (string, error) {
	out, err := c.k8s.Output(ctx, "get", "secret", "argocd-initial-admin-secret", "-n", c.cfg.Namespace, "-o", "jsonpath={.data.password}")
	if err != nil {
		return "", fmt.Errorf("reading admin password: %w", err)
	}
	b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(out)))
	if err != nil {
		return "", fmt.Errorf("decoding admin password: %w", err)
	}
	return string(b), nil
}

// RotatePassword updates the Argo CD admin password and returns the new value.
func (c *Client) RotatePassword(ctx context.Context, password string) (string, error) {
	if password == "" {
		password = randomPassword()
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}

	patch := map[string]any{
		"stringData": map[string]string{
			"password":            password,
			"admin.password":      string(hash),
			"admin.passwordMtime": time.Now().UTC().Format(time.RFC3339),
		},
	}
	data, err := json.Marshal(patch)
	if err != nil {
		return "", fmt.Errorf("encoding password patch: %w", err)
	}

	if err := c.k8s.Kubectl(ctx, "patch", "secret", "argocd-initial-admin-secret", "-n", c.cfg.Namespace, "--type=merge", "-p", string(data)); err != nil {
		return "", fmt.Errorf("rotating admin password: %w", err)
	}
	return password, nil
}

// randomPassword returns a short cryptographically-random alphanumeric password.
func randomPassword() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const n = 16
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	var b strings.Builder
	for _, v := range buf {
		b.WriteByte(chars[int(v)%len(chars)])
	}
	return b.String()
}

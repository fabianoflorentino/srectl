// Package kind wraps the `kind` CLI for cluster lifecycle. srectl is a thin
// task runner: it shells out to kind rather than reimplementing its logic.
package kind

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Client drives the `kind` binary on PATH.
type Client struct {
	binary string
	stdout io.Writer
	stderr io.Writer
	stdin  io.Reader
}

// Option configures a Client.
type Option func(*Client)

// WithBinary overrides the `kind` binary path (useful for tests and non-PATH installs).
func WithBinary(path string) Option {
	return func(c *Client) { c.binary = path }
}

// WithStdout redirects standard output.
func WithStdout(w io.Writer) Option {
	return func(c *Client) { c.stdout = w }
}

// WithStderr redirects standard error.
func WithStderr(w io.Writer) Option {
	return func(c *Client) { c.stderr = w }
}

// WithStdin redirects standard input.
func WithStdin(r io.Reader) Option {
	return func(c *Client) { c.stdin = r }
}

// New returns a Client bound to the default streams and the `kind` binary.
func New(opts ...Option) *Client {
	c := &Client{binary: "kind", stdout: os.Stdout, stderr: os.Stderr, stdin: os.Stdin}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Cluster describes the kind cluster to create.
type Cluster struct {
	Name       string
	Nodes      int
	ConfigFile string // optional path to a kind config file; when empty a minimal config is generated
}

// configFile is the minimal kind node topology written before `kind create`.
// JSON is used because it is a valid subset of YAML.
type configFile struct {
	Kind       string       `json:"kind"`
	APIVersion string       `json:"apiVersion"`
	Nodes      []configNode `json:"nodes"`
}

type configNode struct {
	Role string `json:"role"`
}

// Create provisions a kind cluster with one control-plane and Nodes-1 workers.
// When ConfigFile is set it is passed to `kind create --config`; otherwise a
// minimal config is generated from Nodes.
func (c *Client) Create(ctx context.Context, cluster Cluster) error {
	configPath := cluster.ConfigFile
	if configPath == "" {
		path, err := writeConfig(cluster.Nodes)
		if err != nil {
			return err
		}
		configPath = path
		defer func() { _ = os.Remove(configPath) }()
	}

	return c.run(ctx, "create", "cluster", "--name", cluster.Name, "--config", configPath)
}

// Delete removes a kind cluster by name.
func (c *Client) Delete(ctx context.Context, name string) error {
	return c.run(ctx, "delete", "cluster", "--name", name)
}

// Exists reports whether a kind cluster with the given name exists.
func (c *Client) Exists(ctx context.Context, name string) (bool, error) {
	out, err := c.output(ctx, "get", "clusters")
	if err != nil {
		return false, fmt.Errorf("listing kind clusters: %w", err)
	}
	for _, cluster := range strings.Fields(string(out)) {
		if cluster == name {
			return true, nil
		}
	}
	return false, nil
}

// Nodes returns the node names of a running cluster.
func (c *Client) Nodes(ctx context.Context, name string) ([]string, error) {
	out, err := c.output(ctx, "get", "nodes", "--name", name)
	if err != nil {
		return nil, fmt.Errorf("listing kind nodes: %w", err)
	}
	return strings.Fields(string(out)), nil
}

func (c *Client) run(ctx context.Context, args ...string) error {
	//nolint:gosec // G204: srectl shells out to kind by design (fixed binary, local dev tool)
	cmd := exec.CommandContext(ctx, c.binary, args...)
	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr
	cmd.Stdin = c.stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kind %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

func (c *Client) output(ctx context.Context, args ...string) ([]byte, error) {
	//nolint:gosec // G204: srectl shells out to kind by design (fixed binary, local dev tool)
	cmd := exec.CommandContext(ctx, c.binary, args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("kind %s: %w", strings.Join(args, " "), err)
	}
	return out, nil
}

// writeConfig builds a temporary kind config file with the requested topology
// and returns its path. The caller owns cleanup.
func writeConfig(nodes int) (string, error) {
	if nodes < 1 {
		return "", fmt.Errorf("kind cluster requires at least 1 node, got %d", nodes)
	}

	cfg := configFile{
		Kind:       "Cluster",
		APIVersion: "kind.x-k8s.io/v1alpha4",
		Nodes:      []configNode{{Role: "control-plane"}},
	}
	for i := 1; i < nodes; i++ {
		cfg.Nodes = append(cfg.Nodes, configNode{Role: "worker"})
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("encoding kind config: %w", err)
	}

	f, err := os.CreateTemp("", "srectl-kind-*.yaml")
	if err != nil {
		return "", fmt.Errorf("creating temp kind config: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(data); err != nil {
		return "", fmt.Errorf("writing kind config: %w", err)
	}
	return f.Name(), nil
}

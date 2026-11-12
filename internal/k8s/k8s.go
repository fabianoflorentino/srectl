// Package k8s wraps kubectl and helm for the operations srectl needs: apply,
// helm install/upgrade, and port-forward. Like the kind package, it shells out
// rather than embedding a Kubernetes client.
package k8s

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Client drives the kubectl and helm binaries on PATH.
type Client struct {
	kubectl string
	helm    string
	stdout  io.Writer
	stderr  io.Writer
	stdin   io.Reader
}

// Option configures a Client.
type Option func(*Client)

// WithKubectl overrides the kubectl binary path.
func WithKubectl(path string) Option {
	return func(c *Client) { c.kubectl = path }
}

// WithHelm overrides the helm binary path.
func WithHelm(path string) Option {
	return func(c *Client) { c.helm = path }
}

// WithStdout redirects standard output.
func WithStdout(w io.Writer) Option {
	return func(c *Client) { c.stdout = w }
}

// WithStderr redirects standard error.
func WithStderr(w io.Writer) Option {
	return func(c *Client) { c.stderr = w }
}

// New returns a Client bound to the default streams and PATH binaries.
func New(opts ...Option) *Client {
	c := &Client{kubectl: "kubectl", helm: "helm", stdout: os.Stdout, stderr: os.Stderr, stdin: os.Stdin}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Kubectl runs `kubectl` with the given arguments, streaming output.
func (c *Client) Kubectl(ctx context.Context, args ...string) error {
	return c.run(ctx, c.kubectl, args...)
}

// Helm runs `helm` with the given arguments, streaming output.
func (c *Client) Helm(ctx context.Context, args ...string) error {
	return c.run(ctx, c.helm, args...)
}

// PortForward starts a blocking `kubectl port-forward` from localPort to
// remotePort for the given resource in namespace. It returns when the command
// exits (or is cancelled via ctx).
func (c *Client) PortForward(ctx context.Context, namespace, resource string, localPort, remotePort int) error {
	args := []string{
		"port-forward",
		"--namespace", namespace,
		resource,
		fmt.Sprintf("%d:%d", localPort, remotePort),
	}
	return c.run(ctx, c.kubectl, args...)
}

func (c *Client) run(ctx context.Context, binary string, args ...string) error {
	//nolint:gosec // G204: srectl shells out to kubectl/helm by design (fixed binaries)
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr
	cmd.Stdin = c.stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", binary, strings.Join(args, " "), err)
	}
	return nil
}

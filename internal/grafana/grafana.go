// Package grafana reads the admin credentials of the Grafana instance
// provisioned by the kube-prometheus-stack chart.
package grafana

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/fabianoflorentino/srectl/internal/k8s"
)

// Config identifies where Grafana stores its admin credentials.
type Config struct {
	Namespace string
	Secret    string
}

// Client reads Grafana's admin credentials.
type Client struct {
	k8s *k8s.Client
	cfg Config
}

// New returns a Grafana client backed by the given kubectl wrapper.
func New(k8c *k8s.Client, cfg Config) *Client {
	return &Client{k8s: k8c, cfg: cfg}
}

// Username returns the current Grafana admin username.
func (c *Client) Username(ctx context.Context) (string, error) {
	return c.secretValue(ctx, "admin-user", "user")
}

// Password returns the current Grafana admin password.
func (c *Client) Password(ctx context.Context) (string, error) {
	return c.secretValue(ctx, "admin-password", "password")
}

func (c *Client) secretValue(ctx context.Context, key, label string) (string, error) {
	out, err := c.k8s.Output(ctx, "get", "secret", c.cfg.Secret, "-n", c.cfg.Namespace,
		"-o", "jsonpath={.data."+key+"}")
	if err != nil {
		return "", fmt.Errorf("reading grafana admin %s: %w", label, err)
	}
	b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(out)))
	if err != nil {
		return "", fmt.Errorf("decoding grafana admin %s: %w", label, err)
	}
	return string(b), nil
}

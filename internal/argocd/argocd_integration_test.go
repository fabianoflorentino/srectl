//go:build integration

package argocd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fabianoflorentino/srectl/internal/k8s"
)

// localPort is the host port the test forwards Argo CD to.
const localPort = 8443

// TestRotatePasswordLogin proves the real contract that a unit test cannot:
// Argo CD authenticates against argocd-secret, so rotating the password and
// restarting argocd-server must make the new password valid for login.
//
// Run against a cluster with Argo CD installed:
//
//	go test -tags=integration ./internal/argocd/ -run TestRotatePasswordLogin
//
// Side effect: it rotates the Argo CD admin password (and leaves it changed).
func TestRotatePasswordLogin(t *testing.T) {
	ctx := context.Background()
	namespace := envOr("SRECTL_ARGOCD_NAMESPACE", "argocd")

	k8c := k8s.New()
	c := New(k8c, Config{Namespace: namespace})

	newPassword := fmt.Sprintf("it-%d", time.Now().UnixNano())
	if got, err := c.RotatePassword(ctx, newPassword); err != nil {
		t.Fatalf("RotatePassword: %v", err)
	} else if got != newPassword {
		t.Fatalf("RotatePassword = %q, want %q", got, newPassword)
	}

	if err := k8c.RolloutStatus(ctx, namespace, "deployment/"+c.ServerDeployment(), "120s"); err != nil {
		t.Fatalf("waiting for argocd-server: %v", err)
	}

	cancel := startPortForward(t, ctx, k8c, namespace, c.ServerService())
	defer cancel()

	client := newHTTPClient()
	waitForAPI(t, client)

	if code, body := login(t, client, "admin", newPassword); code != http.StatusOK {
		t.Fatalf("login with new password = %d, want 200 (body: %v)", code, body)
	}
	if code, _ := login(t, client, "admin", "wrong-password"); code == http.StatusOK {
		t.Fatal("login with wrong password = 200, want a failure")
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		Timeout:   10 * time.Second,
	}
}

func startPortForward(t *testing.T, ctx context.Context, k8c *k8s.Client, namespace, service string) context.CancelFunc {
	t.Helper()
	pfCtx, cancel := context.WithCancel(ctx)
	go func() {
		_ = k8c.PortForward(pfCtx, namespace, "svc/"+service, localPort, 443)
	}()
	return cancel
}

func waitForAPI(t *testing.T, client *http.Client) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get(fmt.Sprintf("https://localhost:%d/api/version", localPort))
		if err == nil {
			_ = resp.Body.Close()
			return
		}
		time.Sleep(time.Second)
	}
	t.Fatal("Argo CD API did not become reachable")
}

func login(t *testing.T, client *http.Client, username, password string) (int, map[string]any) {
	t.Helper()
	body := fmt.Sprintf(`{"username":%q,"password":%q}`, username, password)
	resp, err := client.Post(fmt.Sprintf("https://localhost:%d/api/v1/session", localPort), "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("login request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

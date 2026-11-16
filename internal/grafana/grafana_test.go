package grafana

import (
	"bytes"
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/fabianoflorentino/srectl/internal/k8s"
)

func fakeExec(t *testing.T, name, script string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatalf("writing fake %s: %v", name, err)
	}
	return path
}

func testClient(t *testing.T) *Client {
	t.Helper()
	var out bytes.Buffer

	kubectlBin := fakeExec(t, "kubectl", `if [ "$1" = "get" ] && [ "$2" = "secret" ]; then
  case "$*" in
    *admin-user*) echo "$FAKE_USER_B64" ;;
    *admin-password*) echo "$FAKE_PASSWORD_B64" ;;
  esac
  exit 0
fi
echo "$@"
exit 0`)

	k8c := k8s.New(k8s.WithKubectl(kubectlBin), k8s.WithStdout(&out), k8s.WithStderr(&out))

	return New(k8c, Config{
		Namespace: "monitoring",
		Secret:    "kube-prometheus-stack-grafana",
	})
}

func TestUsernameAndPassword(t *testing.T) {
	t.Setenv("FAKE_USER_B64", base64.StdEncoding.EncodeToString([]byte("admin")))
	t.Setenv("FAKE_PASSWORD_B64", base64.StdEncoding.EncodeToString([]byte("s3cret")))
	c := testClient(t)

	user, err := c.Username(context.Background())
	if err != nil {
		t.Fatalf("Username: %v", err)
	}
	if user != "admin" {
		t.Errorf("Username = %q, want admin", user)
	}

	pass, err := c.Password(context.Background())
	if err != nil {
		t.Fatalf("Password: %v", err)
	}
	if pass != "s3cret" {
		t.Errorf("Password = %q, want s3cret", pass)
	}
}

func TestDecodeError(t *testing.T) {
	t.Setenv("FAKE_USER_B64", "!!!not-base64!!!")
	t.Setenv("FAKE_PASSWORD_B64", "!!!not-base64!!!")
	c := testClient(t)

	if _, err := c.Password(context.Background()); err == nil {
		t.Error("Password: want decode error, got nil")
	}
	if _, err := c.Username(context.Background()); err == nil {
		t.Error("Username: want decode error, got nil")
	}
}

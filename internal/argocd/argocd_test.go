package argocd

import (
	"bytes"
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
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

const fakeAppsJSON = `{"items":[{"metadata":{"name":"root-app"},"status":{"sync":{"status":"Synced"},"health":{"status":"Healthy"}}},{"metadata":{"name":"apps"},"status":{"sync":{"status":"OutOfSync"},"health":{"status":"Progressing"}}}]}`

func testClient(t *testing.T) (*Client, *bytes.Buffer) {
	t.Helper()
	var out bytes.Buffer

	kubectlBin := fakeExec(t, "kubectl", `case "$1 $2" in
  "get applications")
    echo '`+fakeAppsJSON+`'
    ;;
  "get secret")
    echo "$FAKE_ADMIN_PASSWORD_B64"
    ;;
  *)
    echo "$@"
    ;;
esac
exit 0`)

	helmBin := fakeExec(t, "helm", `echo "$@"; exit 0`)

	k8c := k8s.New(k8s.WithKubectl(kubectlBin), k8s.WithHelm(helmBin), k8s.WithStdout(&out), k8s.WithStderr(&out))

	c := New(k8c, Config{
		Namespace: "argocd",
		Version:   "10.18.0",
		Repo:      "https://example.com/repo",
		Revision:  "main",
	}, WithPaths(Paths{RootApp: "testdata/root-app.yaml"}))

	return c, &out
}

func TestBootstrap(t *testing.T) {
	c, out := testClient(t)
	if err := c.Bootstrap(context.Background()); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	for _, want := range []string{
		"upgrade --install argo-cd argo/argo-cd",
		"--version 10.18.0",
		"rollout status deployment/argo-cd-argocd-server",
		"apply -f",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("Bootstrap output missing %q, got %q", want, out.String())
		}
	}
}

func TestRenderRootApp(t *testing.T) {
	c, _ := testClient(t)
	manifest, err := c.renderRootApp()
	if err != nil {
		t.Fatalf("renderRootApp: %v", err)
	}
	s := string(manifest)
	if !strings.Contains(s, "https://example.com/repo") {
		t.Errorf("rendered manifest missing repo URL, got %q", s)
	}
	if !strings.Contains(s, "targetRevision: main") {
		t.Errorf("rendered manifest missing revision, got %q", s)
	}
	if strings.Contains(s, "{{") {
		t.Errorf("rendered manifest still has template placeholders, got %q", s)
	}
}

func TestSync(t *testing.T) {
	c, out := testClient(t)
	if err := c.Sync(context.Background(), ""); err != nil {
		t.Fatalf("Sync(all): %v", err)
	}
	if !strings.Contains(out.String(), "annotate application --all -n argocd") {
		t.Errorf("Sync output = %q", out.String())
	}
}

func TestSyncSingleApp(t *testing.T) {
	c, out := testClient(t)
	if err := c.Sync(context.Background(), "apps"); err != nil {
		t.Fatalf("Sync(apps): %v", err)
	}
	if !strings.Contains(out.String(), "annotate application apps -n argocd") {
		t.Errorf("Sync output = %q", out.String())
	}
}

func TestApps(t *testing.T) {
	c, _ := testClient(t)
	apps, err := c.Apps(context.Background())
	if err != nil {
		t.Fatalf("Apps: %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("Apps returned %d apps, want 2", len(apps))
	}
	if apps[0].Name != "root-app" || apps[0].Sync != "Synced" || apps[0].Health != "Healthy" {
		t.Errorf("apps[0] = %+v", apps[0])
	}
	if apps[1].Name != "apps" || apps[1].Sync != "OutOfSync" || apps[1].Health != "Progressing" {
		t.Errorf("apps[1] = %+v", apps[1])
	}
}

func TestPassword(t *testing.T) {
	t.Setenv("FAKE_ADMIN_PASSWORD_B64", base64.StdEncoding.EncodeToString([]byte("s3cret")))
	c, _ := testClient(t)

	got, err := c.Password(context.Background())
	if err != nil {
		t.Fatalf("Password: %v", err)
	}
	if got != "s3cret" {
		t.Errorf("Password = %q, want s3cret", got)
	}
}

func TestRotatePassword(t *testing.T) {
	c, out := testClient(t)

	got, err := c.RotatePassword(context.Background(), "newpass123")
	if err != nil {
		t.Fatalf("RotatePassword: %v", err)
	}
	if got != "newpass123" {
		t.Errorf("RotatePassword = %q, want newpass123", got)
	}
	if !strings.Contains(out.String(), "patch secret argocd-initial-admin-secret") {
		t.Errorf("RotatePassword output = %q", out.String())
	}
}

func TestRotatePasswordGenerates(t *testing.T) {
	c, _ := testClient(t)
	got, err := c.RotatePassword(context.Background(), "")
	if err != nil {
		t.Fatalf("RotatePassword: %v", err)
	}
	if len(got) == 0 {
		t.Error("generated password is empty")
	}
}

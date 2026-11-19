package platform

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fabianoflorentino/srectl/internal/argocd"
	"github.com/fabianoflorentino/srectl/internal/k8s"
	"github.com/fabianoflorentino/srectl/internal/kind"
)

func fakeExec(t *testing.T, name, script string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatalf("writing fake %s: %v", name, err)
	}
	return path
}

const fakeAppsJSON = `{"items":[{"metadata":{"name":"root-app"},"status":{"sync":{"status":"Synced"},"health":{"status":"Healthy"}}}]}`

func testManager(t *testing.T) (*Manager, *bytes.Buffer) {
	t.Helper()
	var out bytes.Buffer

	kindBin := fakeExec(t, "kind", `case "$1" in
  get)
    if [ "$2" = "clusters" ]; then printf '%s' "$FAKE_CLUSTERS"; fi
    if [ "$2" = "nodes" ]; then printf '%s' "$FAKE_NODES"; fi
    ;;
  *) echo "$@";;
esac
exit 0`)

	kubectlBin := fakeExec(t, "kubectl", `case "$1 $2" in
  "get deployment")
    echo "1"
    ;;
  "get applications")
    echo '`+fakeAppsJSON+`'
    ;;
  *)
    echo "$@"
    ;;
esac
exit 0`)

	helmBin := fakeExec(t, "helm", `echo "$@"; exit 0`)

	kc := kind.New(kind.WithBinary(kindBin), kind.WithStdout(&out), kind.WithStderr(&out))
	k8c := k8s.New(k8s.WithKubectl(kubectlBin), k8s.WithHelm(helmBin), k8s.WithStdout(&out), k8s.WithStderr(&out))
	ac := argocd.New(k8c, argocd.Config{
		Namespace: "argocd",
		Version:   "10.18.0",
		Repo:      "https://example.com/repo",
		Revision:  "main",
	}, argocd.WithPaths(argocd.Paths{RootApp: "testdata/root-app.yaml"}))

	return New(kc, k8c, ac), &out
}

func TestUp(t *testing.T) {
	mgr, out := testManager(t)
	if err := mgr.Up(context.Background(), kind.Cluster{Name: "srectl", Nodes: 3}); err != nil {
		t.Fatalf("Up: %v", err)
	}
	for _, want := range []string{
		"create cluster --name srectl",
		"upgrade --install argo-cd argo-cd",
		"--repo https://argoproj.github.io/argo-helm",
		"rollout status deployment/argo-cd-argocd-server",
		"apply -f",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("Up output missing %q, got %q", want, out.String())
		}
	}
}

func TestStatusClusterDown(t *testing.T) {
	t.Setenv("FAKE_CLUSTERS", "")
	mgr, _ := testManager(t)

	rep, err := mgr.Status(context.Background(), "srectl")
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if rep.Cluster != "down" {
		t.Errorf("Cluster = %q, want down", rep.Cluster)
	}
	if rep.Argocd != "NotReady" {
		t.Errorf("Argocd = %q, want NotReady", rep.Argocd)
	}
	if len(rep.Apps) != 0 {
		t.Errorf("Apps = %d, want 0", len(rep.Apps))
	}
}

func TestStatusClusterUp(t *testing.T) {
	t.Setenv("FAKE_CLUSTERS", "srectl")
	t.Setenv("FAKE_NODES", "cp w1 w2")
	mgr, _ := testManager(t)

	rep, err := mgr.Status(context.Background(), "srectl")
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if rep.Cluster != "up (3 nodes)" {
		t.Errorf("Cluster = %q, want up (3 nodes)", rep.Cluster)
	}
	if rep.Argocd != "Ready" {
		t.Errorf("Argocd = %q, want Ready", rep.Argocd)
	}
	if len(rep.Apps) != 1 || rep.Apps[0].Name != "root-app" {
		t.Errorf("Apps = %+v", rep.Apps)
	}
}

func TestReportWrite(t *testing.T) {
	var buf bytes.Buffer
	r := Report{Cluster: "up (3 nodes)", Argocd: "Ready", Apps: []argocd.App{{Name: "root-app", Sync: "Synced", Health: "Healthy"}}}
	if err := r.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	for _, want := range []string{"cluster:", "up (3 nodes)", "argocd:", "Ready", "root-app:", "Synced", "Healthy"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("report missing %q, got %q", want, buf.String())
		}
	}
}

func TestClusterSpec(t *testing.T) {
	mgr, _ := testManager(t)
	spec := mgr.ClusterSpec("srectl", 3)
	if spec.Name != "srectl" || spec.Nodes != 3 {
		t.Errorf("spec = %+v", spec)
	}
	if spec.ConfigFile != KindConfigFile() {
		t.Errorf("spec.ConfigFile = %q, want %q", spec.ConfigFile, KindConfigFile())
	}
}

func TestKindConfigFileExists(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "deploy", "kind"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "deploy", "kind", "kind-config.yaml"), []byte("kind: Cluster\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	if got := KindConfigFile(); got != "deploy/kind/kind-config.yaml" {
		t.Errorf("KindConfigFile = %q, want deploy/kind/kind-config.yaml", got)
	}
}

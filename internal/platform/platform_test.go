package platform

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

	kubectlBin := fakeExec(t, "kubectl", `case "$1" in
  get)
    if [ "$2" = "secret" ]; then echo "Zm9v"; else echo "1"; fi
    ;;
  create)
    echo "apiVersion: v1"
    ;;
  *)
    echo "$@"
    ;;
esac
exit 0`)

	helmBin := fakeExec(t, "helm", `echo "$@"; exit 0`)

	kc := kind.New(kind.WithBinary(kindBin), kind.WithStdout(&out), kind.WithStderr(&out))
	k8c := k8s.New(k8s.WithKubectl(kubectlBin), k8s.WithHelm(helmBin), k8s.WithStdout(&out), k8s.WithStderr(&out))

	mgr := New(kc, k8c, WithPaths(Paths{
		KindConfig:    "testdata/kind-config.yaml",
		IngressValues: "testdata/ingress-values.yaml",
		Namespace:     "testdata/namespace.yaml",
		Issuers:       []string{"testdata/a.yaml", "testdata/b.yaml"},
		HelloDir:      "testdata/hello/",
	}))

	return mgr, &out
}

func TestInstallIngressArgs(t *testing.T) {
	mgr, out := testManager(t)
	if err := mgr.InstallIngress(context.Background()); err != nil {
		t.Fatalf("InstallIngress: %v", err)
	}
	for _, want := range []string{
		"upgrade --install ingress-nginx ingress-nginx/ingress-nginx",
		"--version " + IngressNGINXVersion,
		"testdata/ingress-values.yaml",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("InstallIngress output missing %q, got %q", want, out.String())
		}
	}
}

func TestInstallCertManagerArgs(t *testing.T) {
	mgr, out := testManager(t)
	if err := mgr.InstallCertManager(context.Background()); err != nil {
		t.Fatalf("InstallCertManager: %v", err)
	}
	for _, want := range []string{
		"upgrade --install cert-manager jetstack/cert-manager",
		"--version " + CertManagerVersion,
		"crds.enabled=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("InstallCertManager output missing %q, got %q", want, out.String())
		}
	}
}

func TestApplyIssuersArgs(t *testing.T) {
	mgr, out := testManager(t)
	if err := mgr.ApplyIssuers(context.Background()); err != nil {
		t.Fatalf("ApplyIssuers: %v", err)
	}
	if want := "apply -f testdata/a.yaml,testdata/b.yaml"; !strings.Contains(out.String(), want) {
		t.Errorf("ApplyIssuers output = %q, want %q", out.String(), want)
	}
}

func TestDeployHelloArgs(t *testing.T) {
	mgr, out := testManager(t)
	if err := mgr.DeployHello(context.Background()); err != nil {
		t.Fatalf("DeployHello: %v", err)
	}
	if want := "apply -f testdata/hello/"; !strings.Contains(out.String(), want) {
		t.Errorf("DeployHello output = %q, want %q", out.String(), want)
	}
}

func TestWaitReady(t *testing.T) {
	mgr, out := testManager(t)
	if err := mgr.WaitReady(context.Background()); err != nil {
		t.Fatalf("WaitReady: %v", err)
	}
	for _, want := range []string{
		"rollout status deployment/ingress-nginx-controller",
		"rollout status deployment/cert-manager",
		"rollout status deployment/hello",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("WaitReady output missing %q, got %q", want, out.String())
		}
	}
}

func TestUp(t *testing.T) {
	mgr, out := testManager(t)
	if err := mgr.Up(context.Background(), kind.Cluster{Name: "srectl", Nodes: 3}); err != nil {
		t.Fatalf("Up: %v", err)
	}
	for _, want := range []string{
		"create cluster --name srectl",
		"repo add ingress-nginx",
		"repo add jetstack",
		"upgrade --install ingress-nginx",
		"upgrade --install cert-manager",
		"apply -f testdata/namespace.yaml",
		"apply -f testdata/a.yaml,testdata/b.yaml",
		"apply -f testdata/hello/",
		"rollout status",
		"wait --for=condition=Ready certificate/hello-tls",
		"apply -f",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("Up output missing %q, got %q", want, out.String())
		}
	}
}

func TestSyncTrustBundle(t *testing.T) {
	mgr, out := testManager(t)
	if err := mgr.SyncTrustBundle(context.Background()); err != nil {
		t.Fatalf("SyncTrustBundle: %v", err)
	}
	if !strings.Contains(out.String(), "apply -f") {
		t.Errorf("SyncTrustBundle output = %q, want apply -f", out.String())
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
	if rep.Ingress != "NotReady" {
		t.Errorf("Ingress = %q, want NotReady", rep.Ingress)
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
	if rep.Ingress != "Ready" {
		t.Errorf("Ingress = %q, want Ready", rep.Ingress)
	}
	if rep.CertManager != "Ready" {
		t.Errorf("CertManager = %q, want Ready", rep.CertManager)
	}
	if rep.Hello != "Ready" {
		t.Errorf("Hello = %q, want Ready", rep.Hello)
	}
}

func TestReportWrite(t *testing.T) {
	var buf bytes.Buffer
	r := Report{Cluster: "up (3 nodes)", Ingress: "Ready", CertManager: "Ready", Hello: "Ready"}
	if err := r.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	for _, want := range []string{"cluster:", "up (3 nodes)", "ingress:", "Ready", "cert-manager:", "hello:"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("report missing %q, got %q", want, buf.String())
		}
	}
}

func TestClusterSpec(t *testing.T) {
	mgr, _ := testManager(t)
	spec := mgr.ClusterSpec("srectl", 3)
	if spec.Name != "srectl" {
		t.Errorf("spec.Name = %q, want srectl", spec.Name)
	}
	if spec.Nodes != 3 {
		t.Errorf("spec.Nodes = %d, want 3", spec.Nodes)
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

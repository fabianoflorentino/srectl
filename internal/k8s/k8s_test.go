package k8s

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeBinary writes an executable shell script and returns its path.
func fakeBinary(t *testing.T, name, script string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatalf("writing fake %s binary: %v", name, err)
	}
	return path
}

func TestKubectlArgs(t *testing.T) {
	var out bytes.Buffer
	c := New(WithKubectl(fakeBinary(t, "kubectl", `echo "$@"`)), WithStdout(&out))

	if err := c.Kubectl(context.Background(), "get", "pods"); err != nil {
		t.Fatalf("Kubectl: %v", err)
	}
	if !strings.Contains(out.String(), "get pods") {
		t.Errorf("Kubectl output = %q, want get pods", out.String())
	}
}

func TestHelmArgs(t *testing.T) {
	var out bytes.Buffer
	c := New(WithHelm(fakeBinary(t, "helm", `echo "$@"`)), WithStdout(&out))

	if err := c.Helm(context.Background(), "install", "foo", "chart"); err != nil {
		t.Fatalf("Helm: %v", err)
	}
	if !strings.Contains(out.String(), "install foo chart") {
		t.Errorf("Helm output = %q, want install foo chart", out.String())
	}
}

func TestPortForwardArgs(t *testing.T) {
	var out bytes.Buffer
	c := New(WithKubectl(fakeBinary(t, "kubectl", `echo "$@"`)), WithStdout(&out))

	if err := c.PortForward(context.Background(), "default", "svc/foo", 8080, 80); err != nil {
		t.Fatalf("PortForward: %v", err)
	}
	want := "port-forward --namespace default svc/foo 8080:80"
	if !strings.Contains(out.String(), want) {
		t.Errorf("PortForward output = %q, want %q", out.String(), want)
	}
}

func TestRunError(t *testing.T) {
	c := New(WithKubectl(fakeBinary(t, "kubectl", `exit 1`)))

	if err := c.Kubectl(context.Background(), "get", "pods"); err == nil {
		t.Fatal("Kubectl = nil error, want an error")
	}
}

func TestStderrOption(t *testing.T) {
	var errBuf bytes.Buffer
	c := New(WithKubectl(fakeBinary(t, "kubectl", `echo "to-stderr" >&2`)), WithStderr(&errBuf))

	_ = c.Kubectl(context.Background(), "get", "pods")
	if !strings.Contains(errBuf.String(), "to-stderr") {
		t.Errorf("stderr = %q, want to-stderr", errBuf.String())
	}
}

func TestApplyArgs(t *testing.T) {
	var out bytes.Buffer
	c := New(WithKubectl(fakeBinary(t, "kubectl", `echo "$@"`)), WithStdout(&out))

	if err := c.Apply(context.Background(), "a.yaml", "b.yaml"); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if want := "apply -f a.yaml,b.yaml"; !strings.Contains(out.String(), want) {
		t.Errorf("Apply output = %q, want %q", out.String(), want)
	}
}

func TestDeleteArgs(t *testing.T) {
	var out bytes.Buffer
	c := New(WithKubectl(fakeBinary(t, "kubectl", `echo "$@"`)), WithStdout(&out))

	if err := c.Delete(context.Background(), "a.yaml"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if want := "delete -f a.yaml"; !strings.Contains(out.String(), want) {
		t.Errorf("Delete output = %q, want %q", out.String(), want)
	}
}

func TestRolloutStatusArgs(t *testing.T) {
	var out bytes.Buffer
	c := New(WithKubectl(fakeBinary(t, "kubectl", `echo "$@"`)), WithStdout(&out))

	if err := c.RolloutStatus(context.Background(), "demo", "deployment/hello", "300s"); err != nil {
		t.Fatalf("RolloutStatus: %v", err)
	}
	if want := "rollout status deployment/hello --namespace demo --timeout=300s"; !strings.Contains(out.String(), want) {
		t.Errorf("RolloutStatus output = %q, want %q", out.String(), want)
	}
}

func TestDeploymentAvailable(t *testing.T) {
	c := New(WithKubectl(fakeBinary(t, "kubectl", `echo "1"`)))

	ok, err := c.DeploymentAvailable(context.Background(), "ns", "deploy")
	if err != nil {
		t.Fatalf("DeploymentAvailable: %v", err)
	}
	if !ok {
		t.Fatal("DeploymentAvailable = false, want true")
	}
}

func TestDeploymentAvailableFalse(t *testing.T) {
	c := New(WithKubectl(fakeBinary(t, "kubectl", `exit 0`)))

	ok, err := c.DeploymentAvailable(context.Background(), "ns", "deploy")
	if err != nil {
		t.Fatalf("DeploymentAvailable: %v", err)
	}
	if ok {
		t.Fatal("DeploymentAvailable = true, want false")
	}
}

func TestDeploymentAvailableError(t *testing.T) {
	c := New(WithKubectl(fakeBinary(t, "kubectl", `exit 1`)))

	if _, err := c.DeploymentAvailable(context.Background(), "ns", "deploy"); err == nil {
		t.Fatal("DeploymentAvailable = nil error, want an error")
	}
}

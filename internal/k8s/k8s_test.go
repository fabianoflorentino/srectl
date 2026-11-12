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

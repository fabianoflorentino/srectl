package kind

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeBinary writes an executable shell script and returns its path. It lets
// tests drive the Client without a real `kind` binary.
func fakeBinary(t *testing.T, script string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "kind")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatalf("writing fake kind binary: %v", err)
	}
	return path
}

func TestWriteConfigTopology(t *testing.T) {
	path, err := writeConfig(3)
	if err != nil {
		t.Fatalf("writeConfig: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(path) })

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}

	var cfg configFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("decoding config: %v", err)
	}

	if cfg.Kind != "Cluster" {
		t.Errorf("kind = %q, want %q", cfg.Kind, "Cluster")
	}
	if got, want := len(cfg.Nodes), 3; got != want {
		t.Fatalf("node count = %d, want %d", got, want)
	}
	if cfg.Nodes[0].Role != "control-plane" {
		t.Errorf("node[0].role = %q, want control-plane", cfg.Nodes[0].Role)
	}
	for _, n := range cfg.Nodes[1:] {
		if n.Role != "worker" {
			t.Errorf("worker role = %q, want worker", n.Role)
		}
	}
}

func TestWriteConfigRejectsZeroNodes(t *testing.T) {
	if _, err := writeConfig(0); err == nil {
		t.Fatal("writeConfig(0) = nil error, want an error")
	}
}

func TestWriteConfigSingleNode(t *testing.T) {
	path, err := writeConfig(1)
	if err != nil {
		t.Fatalf("writeConfig: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(path) })

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}

	var cfg configFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("decoding config: %v", err)
	}

	if got, want := len(cfg.Nodes), 1; got != want {
		t.Fatalf("node count = %d, want %d", got, want)
	}
	if !strings.Contains(cfg.Nodes[0].Role, "control-plane") {
		t.Errorf("single node role = %q, want control-plane", cfg.Nodes[0].Role)
	}
}

func TestExists(t *testing.T) {
	c := New(WithBinary(fakeBinary(t, `if [ "$1 $2" = "get clusters" ]; then echo "alpha srectl beta"; fi`)))

	exists, err := c.Exists(context.Background(), "srectl")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !exists {
		t.Fatal("Exists = false, want true")
	}
}

func TestExistsFalse(t *testing.T) {
	c := New(WithBinary(fakeBinary(t, `if [ "$1 $2" = "get clusters" ]; then echo "alpha beta"; fi`)))

	exists, err := c.Exists(context.Background(), "srectl")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if exists {
		t.Fatal("Exists = true, want false")
	}
}

func TestExistsError(t *testing.T) {
	c := New(WithBinary(fakeBinary(t, `exit 1`)))

	if _, err := c.Exists(context.Background(), "srectl"); err == nil {
		t.Fatal("Exists = nil error, want an error")
	}
}

func TestNodes(t *testing.T) {
	c := New(WithBinary(fakeBinary(t, `if [ "$1 $2" = "get nodes" ]; then echo "cp w1 w2"; fi`)))

	nodes, err := c.Nodes(context.Background(), "srectl")
	if err != nil {
		t.Fatalf("Nodes: %v", err)
	}
	if got, want := len(nodes), 3; got != want {
		t.Fatalf("node count = %d, want %d", got, want)
	}
}

func TestCreateArgs(t *testing.T) {
	var out bytes.Buffer
	c := New(WithBinary(fakeBinary(t, `echo "$@"`)), WithStdout(&out), WithStdin(strings.NewReader("")))

	if err := c.Create(context.Background(), Cluster{Name: "srectl", Nodes: 2}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !strings.Contains(out.String(), "--name srectl") {
		t.Errorf("Create args = %q, want --name srectl", out.String())
	}
	if !strings.Contains(out.String(), "--config") {
		t.Errorf("Create args = %q, want --config", out.String())
	}
}

func TestDeleteArgs(t *testing.T) {
	var out bytes.Buffer
	c := New(WithBinary(fakeBinary(t, `echo "$@"`)), WithStdout(&out))

	if err := c.Delete(context.Background(), "srectl"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !strings.Contains(out.String(), "delete cluster --name srectl") {
		t.Errorf("Delete args = %q, want delete cluster --name srectl", out.String())
	}
}

func TestRunError(t *testing.T) {
	c := New(WithBinary(fakeBinary(t, `exit 1`)))

	if err := c.Delete(context.Background(), "srectl"); err == nil {
		t.Fatal("Delete = nil error, want an error")
	}
}

func TestStderrOption(t *testing.T) {
	var errBuf bytes.Buffer
	c := New(WithBinary(fakeBinary(t, `echo "to-stderr" >&2`)), WithStderr(&errBuf))

	if err := c.Delete(context.Background(), "srectl"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !strings.Contains(errBuf.String(), "to-stderr") {
		t.Errorf("stderr = %q, want to-stderr", errBuf.String())
	}
}

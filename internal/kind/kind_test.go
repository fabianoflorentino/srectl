package kind

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

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

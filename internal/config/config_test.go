package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	v := NewViper()
	cfg, err := Load(v)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Cluster.Name != "srectl" {
		t.Errorf("cluster.name = %q, want %q", cfg.Cluster.Name, "srectl")
	}
	if cfg.Cluster.Nodes != 3 {
		t.Errorf("cluster.nodes = %d, want 3", cfg.Cluster.Nodes)
	}
	if cfg.Registry != "localhost:5000" {
		t.Errorf("registry = %q, want %q", cfg.Registry, "localhost:5000")
	}
	if !cfg.Components.Observability {
		t.Errorf("components.observability = false, want true")
	}
}

func TestLoadEnvOverride(t *testing.T) {
	t.Setenv("SRECTL_CLUSTER_NAME", "from-env")
	t.Setenv("SRECTL_CLUSTER_NODES", "5")
	t.Setenv("SRECTL_COMPONENTS_CHAOS", "false")

	v := NewViper()
	cfg, err := Load(v)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Cluster.Name != "from-env" {
		t.Errorf("cluster.name = %q, want %q", cfg.Cluster.Name, "from-env")
	}
	if cfg.Cluster.Nodes != 5 {
		t.Errorf("cluster.nodes = %d, want 5", cfg.Cluster.Nodes)
	}
	if cfg.Components.Chaos {
		t.Errorf("components.chaos = true, want false")
	}
}

func TestLoadFileOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "srectl.yaml")
	content := "cluster:\n  name: file-cluster\n  nodes: 7\nregistry: file.registry:6000\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing config file: %v", err)
	}

	v := NewViper()
	v.SetConfigFile(path)
	cfg, err := Load(v)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Cluster.Name != "file-cluster" {
		t.Errorf("cluster.name = %q, want %q", cfg.Cluster.Name, "file-cluster")
	}
	if cfg.Cluster.Nodes != 7 {
		t.Errorf("cluster.nodes = %d, want 7", cfg.Cluster.Nodes)
	}
	if cfg.Registry != "file.registry:6000" {
		t.Errorf("registry = %q, want %q", cfg.Registry, "file.registry:6000")
	}
}

func TestLoadMalformedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(path, []byte("cluster: [unclosed"), 0o600); err != nil {
		t.Fatalf("writing config file: %v", err)
	}

	v := NewViper()
	v.SetConfigFile(path)
	if _, err := Load(v); err == nil {
		t.Fatal("Load() = nil error, want an error for malformed config")
	}
}

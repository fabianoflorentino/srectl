package chaos

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fabianoflorentino/srectl/internal/k8s"
)

func fakeKubectl(t *testing.T) *k8s.Client {
	t.Helper()
	path := filepath.Join(t.TempDir(), "kubectl")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho \"$@\"\n"), 0o755); err != nil {
		t.Fatalf("writing fake kubectl: %v", err)
	}
	return k8s.New(k8s.WithKubectl(path))
}

func TestRunLatency(t *testing.T) {
	if err := Run(context.Background(), fakeKubectl(t), ScenarioLatency, 5*time.Millisecond); err != nil {
		t.Fatalf("Run(latency): %v", err)
	}
}

func TestRunPodKill(t *testing.T) {
	if err := Run(context.Background(), fakeKubectl(t), ScenarioPodKill, 5*time.Millisecond); err != nil {
		t.Fatalf("Run(pod-kill): %v", err)
	}
}

func TestRunUnsupported(t *testing.T) {
	if err := Run(context.Background(), fakeKubectl(t), "dns", time.Millisecond); err == nil {
		t.Error("Run(dns): want error, got nil")
	}
}

func TestManifestPath(t *testing.T) {
	for _, s := range []Scenario{ScenarioLatency, ScenarioPodKill} {
		if _, ok := manifestPath(s); !ok {
			t.Errorf("manifestPath(%q) not supported", s)
		}
	}
	if _, ok := manifestPath("nope"); ok {
		t.Error("manifestPath(nope) should be unsupported")
	}
}

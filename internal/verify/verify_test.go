package verify

import (
	"testing"

	"github.com/fabianoflorentino/srectl/internal/argocd"
	"github.com/fabianoflorentino/srectl/internal/platform"
)

func TestAssessHealthy(t *testing.T) {
	report := platform.Report{
		Cluster: "up (3 nodes)",
		Argocd:  "Ready",
		Apps:    []argocd.App{{Name: "apps", Sync: "Synced", Health: "Healthy"}},
	}
	if err := Assess(report); err != nil {
		t.Errorf("Assess(healthy) = %v, want nil", err)
	}
}

func TestAssessProblems(t *testing.T) {
	cases := []struct {
		name   string
		report platform.Report
	}{
		{"cluster down", platform.Report{Cluster: "down", Argocd: "Ready"}},
		{"argocd not ready", platform.Report{Cluster: "up (3 nodes)", Argocd: "NotReady"}},
		{"app degraded", platform.Report{
			Cluster: "up (3 nodes)", Argocd: "Ready",
			Apps: []argocd.App{{Name: "x", Sync: "OutOfSync", Health: "Degraded"}},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := Assess(tc.report); err == nil {
				t.Error("want error, got nil")
			}
		})
	}
}

// Package chaos runs transient SRE chaos drills against the demo workloads.
//
// The manifests are embedded and applied on demand (not via GitOps, so an
// experiment never gets continuously re-applied by Argo CD). Every experiment
// also carries a duration, so chaos-mesh recovers it even if the CLI is killed.
package chaos

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fabianoflorentino/srectl/internal/k8s"
)

//go:embed manifests/*.yaml
var manifests embed.FS

// Scenario identifies a chaos drill.
type Scenario string

// Supported scenarios.
const (
	// ScenarioLatency stresses api-a CPU so request latency rises.
	ScenarioLatency Scenario = "latency"
	// ScenarioPodKill kills a fraudctl pod to exercise self-healing.
	ScenarioPodKill Scenario = "pod-kill"
)

// ErrUnsupported is returned for scenarios that need NetworkChaos, whose
// iptables/ipset setup does not work inside kind's nested containers.
var ErrUnsupported = errors.New("scenario is not supported on kind")

func manifestPath(s Scenario) (string, bool) {
	switch s {
	case ScenarioLatency:
		return "manifests/latency.yaml", true
	case ScenarioPodKill:
		return "manifests/pod-kill.yaml", true
	default:
		return "", false
	}
}

// Run applies the scenario, waits for duration (or ctx cancellation), then
// removes it and returns. Cleanup uses a fresh context so it still runs when
// ctx is cancelled.
func Run(ctx context.Context, k8sc *k8s.Client, s Scenario, duration time.Duration) error {
	path, ok := manifestPath(s)
	if !ok {
		return fmt.Errorf("chaos %q: %w", s, ErrUnsupported)
	}
	raw, err := manifests.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading chaos manifest: %w", err)
	}
	body := strings.ReplaceAll(string(raw), "__DURATION__", fmt.Sprintf("%ds", int(duration.Seconds())))

	tmp, err := os.CreateTemp("", "srectl-chaos-*.yaml")
	if err != nil {
		return fmt.Errorf("creating chaos manifest: %w", err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.WriteString(body); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("writing chaos manifest: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing chaos manifest: %w", err)
	}

	if err := k8sc.Apply(ctx, tmp.Name()); err != nil {
		return fmt.Errorf("applying chaos: %w", err)
	}

	select {
	case <-ctx.Done():
	case <-time.After(duration):
	}

	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	if err := k8sc.Delete(cleanupCtx, tmp.Name()); err != nil {
		return fmt.Errorf("cleaning up chaos: %w", err)
	}
	return nil
}

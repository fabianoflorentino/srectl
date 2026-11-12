package demo

import (
	"errors"
	"testing"
)

func TestAllContainsEveryScenario(t *testing.T) {
	got := All()
	want := []Scenario{
		ScenarioLoad,
		ScenarioChaos,
		ScenarioBreak,
		ScenarioVerify,
		ScenarioError,
		ScenarioSecurity,
		ScenarioCerts,
		ScenarioProjects,
	}

	if len(got) != len(want) {
		t.Fatalf("All() returned %d scenarios, want %d", len(got), len(want))
	}

	seen := make(map[Scenario]bool, len(got))
	for _, s := range got {
		seen[s] = true
	}
	for _, s := range want {
		if !seen[s] {
			t.Errorf("All() missing scenario %q", s)
		}
	}
}

func TestRunKnownReturnsNotImplemented(t *testing.T) {
	for _, s := range All() {
		err := Run(s)
		if !errors.Is(err, ErrNotImplemented) {
			t.Errorf("Run(%q) = %v, want ErrNotImplemented", s, err)
		}
	}
}

func TestRunUnknownScenario(t *testing.T) {
	err := Run("bogus")
	if err == nil {
		t.Fatal("Run(bogus) = nil, want an error")
	}
	if errors.Is(err, ErrNotImplemented) {
		t.Fatalf("Run(bogus) = %v, want an unknown-scenario error", err)
	}
}

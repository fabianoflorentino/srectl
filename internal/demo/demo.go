// Package demo runs the SRE drill scenarios (load, chaos, verify, ...). In
// Phase 0 the scenarios are stubs; the real load generator, chaos injection and
// verification logic land in Phases 5-7.
package demo

import (
	"errors"
	"fmt"
)

// ErrNotImplemented is returned by every scenario until its phase lands.
var ErrNotImplemented = errors.New("not implemented yet")

// Scenario identifies a demo scenario.
type Scenario string

// Known scenarios.
const (
	ScenarioLoad     Scenario = "load"
	ScenarioChaos    Scenario = "chaos"
	ScenarioBreak    Scenario = "break"
	ScenarioVerify   Scenario = "verify"
	ScenarioError    Scenario = "error"
	ScenarioSecurity Scenario = "security"
	ScenarioCerts    Scenario = "certs"
	ScenarioProjects Scenario = "projects"
)

// All returns every known scenario in a stable order.
func All() []Scenario {
	return []Scenario{
		ScenarioLoad,
		ScenarioChaos,
		ScenarioBreak,
		ScenarioVerify,
		ScenarioError,
		ScenarioSecurity,
		ScenarioCerts,
		ScenarioProjects,
	}
}

// Run executes a demo scenario against the running platform.
func Run(s Scenario) error {
	for _, known := range All() {
		if s == known {
			return fmt.Errorf("demo %s: %w", s, ErrNotImplemented)
		}
	}
	return fmt.Errorf("unknown demo scenario %q", s)
}

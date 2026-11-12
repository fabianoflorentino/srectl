package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	var out, errOut bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute version: %v", err)
	}
	if !strings.Contains(out.String(), "srectl") {
		t.Fatalf("version output = %q, want it to contain %q", out.String(), "srectl")
	}
}

func TestHelpListsCommands(t *testing.T) {
	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute --help: %v", err)
	}

	for _, want := range []string{"up", "down", "status", "cluster", "demo", "open", "version"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("help output missing command %q", want)
		}
	}
}

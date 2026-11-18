package cli

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestDemoVerifyOK(t *testing.T) {
	fakeBinaries(t, "srectl", "cp w1 w2")

	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"demo", "verify"})

	if err := root.Execute(); err != nil {
		t.Fatalf("demo verify: %v", err)
	}
	if !strings.Contains(out.String(), "verify: OK") {
		t.Errorf("output = %q", out.String())
	}
}

func TestDemoVerifyFails(t *testing.T) {
	fakeBinaries(t, "", "")

	root := NewRootCommand()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"demo", "verify"})

	if err := root.Execute(); err == nil {
		t.Error("demo verify with cluster down: want error")
	}
}

func TestDemoChaosPodKill(t *testing.T) {
	fakeBinaries(t, "", "")

	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"demo", "chaos", "pod-kill", "--duration", "10ms"})

	if err := root.Execute(); err != nil {
		t.Fatalf("demo chaos: %v", err)
	}
	if !strings.Contains(out.String(), "recovered") {
		t.Errorf("output = %q", out.String())
	}
}

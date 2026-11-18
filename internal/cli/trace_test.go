package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTraceForeground(t *testing.T) {
	fakeBinaries(t, "", "")

	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"trace"})

	if err := root.Execute(); err != nil {
		t.Fatalf("trace: %v", err)
	}
	if !strings.Contains(out.String(), "stracectl live view") {
		t.Errorf("output = %q", out.String())
	}
}

func TestTraceDownNoPID(t *testing.T) {
	_ = os.Remove(pidFilePath("stracectl"))

	root := NewRootCommand()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"trace", "--down"})

	if err := root.Execute(); err == nil {
		t.Error("trace --down without a pid file: want error")
	}
}

func TestOpenAPIA(t *testing.T) {
	fakeBinaries(t, "", "")

	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"open", "api-a"})

	if err := root.Execute(); err != nil {
		t.Fatalf("open api-a: %v", err)
	}
	if !strings.Contains(out.String(), "svc/api-a") {
		t.Errorf("output = %q, want svc/api-a", out.String())
	}
}

func TestTraceDaemon(t *testing.T) {
	dir := t.TempDir()
	writeExec(t, filepath.Join(dir, "kubectl"), "#!/bin/sh\nsleep 30\n")
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	_ = os.Remove(pidFilePath("stracectl"))

	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"trace", "--daemon"})

	if err := root.Execute(); err != nil {
		t.Fatalf("trace --daemon: %v", err)
	}
	if !strings.Contains(out.String(), "stracectl dashboard running in background") {
		t.Errorf("output = %q", out.String())
	}
	if err := stopDaemon("stracectl"); err != nil {
		t.Errorf("stopDaemon: %v", err)
	}
}

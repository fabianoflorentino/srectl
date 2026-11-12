package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
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

func writeExec(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("writing fake binary %s: %v", path, err)
	}
}

// fakeBinaries installs fake kind, kubectl and helm binaries at the front of
// PATH. kind reads cluster/node names from env for `get` subcommands; kubectl
// reports deployments as available; everything else echoes its arguments.
func fakeBinaries(t *testing.T, clusters, nodes string) {
	t.Helper()
	dir := t.TempDir()

	writeExec(t, filepath.Join(dir, "kind"), `#!/bin/sh
case "$1" in
  get)
    if [ "$2" = "clusters" ]; then printf '%s' "$FAKE_CLUSTERS"; fi
    if [ "$2" = "nodes" ]; then printf '%s' "$FAKE_NODES"; fi
    ;;
  *) echo "$@";;
esac
exit 0`)

	writeExec(t, filepath.Join(dir, "kubectl"), `#!/bin/sh
case "$1" in
  get)
    if [ "$2" = "secret" ]; then echo "Zm9v"; else echo "1"; fi
    ;;
  create)
    echo "apiVersion: v1"
    ;;
  *)
    echo "$@"
    ;;
esac
exit 0`)

	writeExec(t, filepath.Join(dir, "helm"), `#!/bin/sh
echo "$@"
exit 0`)

	t.Setenv("FAKE_CLUSTERS", clusters)
	t.Setenv("FAKE_NODES", nodes)
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
}

func TestClusterStatusDown(t *testing.T) {
	fakeBinaries(t, "", "")

	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetArgs([]string{"cluster", "status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute cluster status: %v", err)
	}
	if !strings.Contains(out.String(), "down") {
		t.Errorf("cluster status output = %q, want down", out.String())
	}
}

func TestClusterStatusUp(t *testing.T) {
	fakeBinaries(t, "srectl", "cp w1 w2")

	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetArgs([]string{"cluster", "status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute cluster status: %v", err)
	}
	if !strings.Contains(out.String(), "up (3 nodes)") {
		t.Errorf("cluster status output = %q, want up (3 nodes)", out.String())
	}
}

func TestClusterCreate(t *testing.T) {
	fakeBinaries(t, "", "")

	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetArgs([]string{"cluster", "create"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute cluster create: %v", err)
	}
	if !strings.Contains(out.String(), "--name srectl") {
		t.Errorf("cluster create output = %q, want --name srectl", out.String())
	}
}

func TestClusterDelete(t *testing.T) {
	fakeBinaries(t, "", "")

	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetArgs([]string{"cluster", "delete"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute cluster delete: %v", err)
	}
	if !strings.Contains(out.String(), "delete cluster --name srectl") {
		t.Errorf("cluster delete output = %q, want delete cluster --name srectl", out.String())
	}
}

func TestUpCommand(t *testing.T) {
	fakeBinaries(t, "", "")

	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetArgs([]string{"up"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute up: %v", err)
	}
	if !strings.Contains(out.String(), "platform ready") {
		t.Errorf("up output = %q, want platform ready", out.String())
	}
}

func TestDownCommand(t *testing.T) {
	fakeBinaries(t, "", "")

	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetArgs([]string{"down"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute down: %v", err)
	}
	if !strings.Contains(out.String(), "deleted") {
		t.Errorf("down output = %q, want deleted", out.String())
	}
}

func TestStatusCommand(t *testing.T) {
	fakeBinaries(t, "srectl", "cp w1 w2")

	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetArgs([]string{"status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute status: %v", err)
	}
	for _, want := range []string{"cluster:", "up (3 nodes)", "ingress:", "Ready", "cert-manager:", "hello:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("status output missing %q, got %q", want, out.String())
		}
	}
}

func TestStubCommands(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"bootstrap", []string{"bootstrap"}, "not implemented yet (Phase 2)"},
		{"sync", []string{"sync"}, "not implemented yet (Phase 2)"},
		{"apps deploy", []string{"apps", "deploy"}, "not implemented yet (Phase 5)"},
		{"apps restart", []string{"apps", "restart"}, "not implemented yet (Phase 5)"},
		{"apps scale", []string{"apps", "scale"}, "not implemented yet (Phase 5)"},
		{"trace", []string{"trace", "api-a"}, "not implemented yet (Phase 6)"},
		{"certs", []string{"certs"}, "not implemented yet (Phase 6)"},
		{"open", []string{"open", "grafana"}, "not implemented yet (Phase 1)"},
		{"logs", []string{"logs", "api-a"}, "not implemented yet (Phase 5)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := NewRootCommand()
			root.SetOut(io.Discard)
			root.SetErr(io.Discard)
			root.SetArgs(tc.args)

			err := root.Execute()
			if err == nil {
				t.Fatalf("execute %v = nil, want error containing %q", tc.args, tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("execute %v error = %q, want containing %q", tc.args, err, tc.wantErr)
			}
		})
	}
}

func TestDemoScenarios(t *testing.T) {
	for _, sc := range []string{"load", "chaos", "break", "verify", "error", "security", "certs", "projects"} {
		t.Run(sc, func(t *testing.T) {
			root := NewRootCommand()
			root.SetOut(io.Discard)
			root.SetErr(io.Discard)
			root.SetArgs([]string{"demo", sc})

			err := root.Execute()
			if err == nil {
				t.Fatalf("demo %s = nil, want error", sc)
			}
			if !strings.Contains(err.Error(), "not implemented yet") {
				t.Errorf("demo %s error = %q, want not implemented", sc, err)
			}
		})
	}
}

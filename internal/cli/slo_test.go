package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeSLOBinaries installs a fake kubectl that echoes subcommands and returns
// response to `kubectl get --raw` (the Prometheus service proxy).
func fakeSLOBinaries(t *testing.T, response string) {
	t.Helper()
	dir := t.TempDir()

	writeExec(t, filepath.Join(dir, "kubectl"), `#!/bin/sh
case "$1" in
  get)
    if [ "$2" = "--raw" ]; then
      printf '%s' "$FAKE_PROM_RESPONSE"
    else
      printf '%s' '1'
    fi
    ;;
  *)
    echo "$@"
    ;;
esac
exit 0`)

	t.Setenv("FAKE_PROM_RESPONSE", response)
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
}

func TestSLOCompute(t *testing.T) {
	fakeSLOBinaries(t, "")

	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetArgs([]string{"slo", "compute"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute slo compute: %v", err)
	}
	for _, want := range []string{"slo: job", "completed", "srectl slo status"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output missing %q: %q", want, out.String())
		}
	}
}

func TestSLOStatus(t *testing.T) {
	fakeSLOBinaries(t, `{"status":"success","data":{"result":[`+
		`{"metric":{"service_name":"api-a","slo_name":"availability"},"value":[1,"1"]}]}}`)

	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetArgs([]string{"slo", "status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute slo status: %v", err)
	}
	for _, want := range []string{"SERVICE", "BURN RATE", "api-a", "availability", "ok"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output missing %q: %q", want, out.String())
		}
	}
}

func TestSLOStatusEmpty(t *testing.T) {
	fakeSLOBinaries(t, `{"status":"success","data":{"result":[]}}`)

	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetArgs([]string{"slo", "status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute slo status: %v", err)
	}
	if !strings.Contains(out.String(), "no exported metrics") {
		t.Errorf("output = %q", out.String())
	}
}

func TestSLOStatusPrometheusError(t *testing.T) {
	fakeSLOBinaries(t, `{"status":"error","data":{"result":[]}}`)

	root := NewRootCommand()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"slo", "status"})

	err := root.Execute()
	if err == nil {
		t.Fatal("execute slo status = nil, want an error")
	}
	if !strings.Contains(err.Error(), "prometheus query failed") {
		t.Errorf("error = %q", err)
	}
}

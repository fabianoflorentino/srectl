package slo

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeRunner struct {
	calls     [][]string
	kubectlFn func(args []string) error
	outputFn  func(args []string) ([]byte, error)
}

func (f *fakeRunner) Kubectl(_ context.Context, args ...string) error {
	f.calls = append(f.calls, args)
	if f.kubectlFn != nil {
		return f.kubectlFn(args)
	}
	return nil
}

func (f *fakeRunner) Output(_ context.Context, args ...string) ([]byte, error) {
	f.calls = append(f.calls, args)
	if f.outputFn != nil {
		return f.outputFn(args)
	}
	return nil, nil
}

func TestCompute(t *testing.T) {
	runner := &fakeRunner{}
	client := New(runner, WithClock(func() time.Time { return time.Unix(1700000000, 0) }))

	name, err := client.Compute(context.Background(), "monitoring", "slo-generator")
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	if name != "slo-generator-manual-1700000000" {
		t.Errorf("Compute name = %q", name)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("kubectl calls = %d, want 2", len(runner.calls))
	}
	create := strings.Join(runner.calls[0], " ")
	if !strings.Contains(create, "create job slo-generator-manual-1700000000") {
		t.Errorf("create args = %q", create)
	}
	if !strings.Contains(create, "--from cronjob/slo-generator") {
		t.Errorf("create args missing --from: %q", create)
	}
	wait := strings.Join(runner.calls[1], " ")
	if !strings.Contains(wait, "--for=condition=complete") {
		t.Errorf("wait args = %q", wait)
	}
}

func TestComputeCreateError(t *testing.T) {
	runner := &fakeRunner{kubectlFn: func([]string) error { return errors.New("boom") }}
	client := New(runner)

	name, err := client.Compute(context.Background(), "monitoring", "slo-generator")
	if err == nil {
		t.Fatal("Compute() = nil error, want an error")
	}
	if name != "" {
		t.Errorf("Compute name = %q, want empty", name)
	}
}

func TestComputeWaitError(t *testing.T) {
	runner := &fakeRunner{kubectlFn: func(args []string) error {
		if args[0] == "wait" {
			return errors.New("timeout")
		}
		return nil
	}}
	client := New(runner)

	name, err := client.Compute(context.Background(), "monitoring", "slo-generator")
	if err == nil {
		t.Fatal("Compute() = nil error, want an error")
	}
	if name == "" {
		t.Error("Compute name is empty, want the job name")
	}
}

func promJSON(status, service, slo, value string) []byte {
	return []byte(`{"status":"` + status + `","data":{"resultType":"vector","result":[` +
		`{"metric":{"service_name":"` + service + `","slo_name":"` + slo + `"},"value":[1700000000,"` + value + `"]}]}}`)
}

func TestStatus(t *testing.T) {
	runner := &fakeRunner{outputFn: func(args []string) ([]byte, error) {
		query := args[2]
		if strings.Contains(query, "alerting_burn_rate_threshold") {
			return promJSON("success", "api-a", "latency", "6"), nil
		}
		return promJSON("success", "api-a", "latency", "14.4"), nil
	}}
	client := New(runner)

	rows, err := client.Status(context.Background(), "monitoring", "kube-prometheus-stack-prometheus", 9090)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	got := rows[0]
	if got.Service != "api-a" || got.SLO != "latency" {
		t.Errorf("row = %+v", got)
	}
	if got.BurnRate != 14.4 || got.Threshold != 6 {
		t.Errorf("row = %+v", got)
	}
	if !got.Alerting() {
		t.Error("Alerting() = false, want true")
	}
}

func TestStatusSortsRows(t *testing.T) {
	runner := &fakeRunner{outputFn: func(args []string) ([]byte, error) {
		if strings.Contains(args[2], "alerting_burn_rate_threshold") {
			return []byte(`{"status":"success","data":{"result":[]}}`), nil
		}
		return []byte(`{"status":"success","data":{"result":[` +
			`{"metric":{"service_name":"fraudctl","slo_name":"availability"},"value":[1,"1"]},` +
			`{"metric":{"service_name":"api-a","slo_name":"latency"},"value":[1,"2"]},` +
			`{"metric":{"service_name":"api-a","slo_name":"availability"},"value":[1,"3"]}]}}`), nil
	}}
	client := New(runner)

	rows, err := client.Status(context.Background(), "monitoring", "prom", 9090)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	order := []string{rows[0].Service + "/" + rows[0].SLO, rows[1].Service + "/" + rows[1].SLO, rows[2].Service + "/" + rows[2].SLO}
	want := []string{"api-a/availability", "api-a/latency", "fraudctl/availability"}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("row %d = %q, want %q", i, order[i], want[i])
		}
	}
}

func TestStatusBurnQueryError(t *testing.T) {
	runner := &fakeRunner{outputFn: func([]string) ([]byte, error) { return nil, errors.New("boom") }}
	client := New(runner)

	if _, err := client.Status(context.Background(), "monitoring", "prom", 9090); err == nil {
		t.Fatal("Status() = nil error, want an error")
	}
}

func TestStatusThresholdQueryError(t *testing.T) {
	runner := &fakeRunner{outputFn: func(args []string) ([]byte, error) {
		if strings.Contains(args[2], "alerting_burn_rate_threshold") {
			return nil, errors.New("boom")
		}
		return promJSON("success", "api-a", "latency", "1"), nil
	}}
	client := New(runner)

	if _, err := client.Status(context.Background(), "monitoring", "prom", 9090); err == nil {
		t.Fatal("Status() = nil error, want an error")
	}
}

func TestStatusBadJSON(t *testing.T) {
	runner := &fakeRunner{outputFn: func([]string) ([]byte, error) { return []byte("nope"), nil }}
	client := New(runner)

	if _, err := client.Status(context.Background(), "monitoring", "prom", 9090); err == nil {
		t.Fatal("Status() = nil error, want an error")
	}
}

func TestStatusPrometheusError(t *testing.T) {
	runner := &fakeRunner{outputFn: func([]string) ([]byte, error) {
		return []byte(`{"status":"error","data":{"result":[]}}`), nil
	}}
	client := New(runner)

	if _, err := client.Status(context.Background(), "monitoring", "prom", 9090); err == nil {
		t.Fatal("Status() = nil error, want an error")
	}
}

func TestStatusBadValue(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"missing value", `{"status":"success","data":{"result":[{"metric":{},"value":[]}]}}`},
		{"not a string", `{"status":"success","data":{"result":[{"metric":{},"value":[1,2]}]}}`},
		{"not a number", `{"status":"success","data":{"result":[{"metric":{},"value":[1,"x"]}]}}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner := &fakeRunner{outputFn: func([]string) ([]byte, error) { return []byte(tc.body), nil }}
			client := New(runner)
			if _, err := client.Status(context.Background(), "monitoring", "prom", 9090); err == nil {
				t.Fatal("Status() = nil error, want an error")
			}
		})
	}
}

func TestWriteEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, nil); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !strings.Contains(buf.String(), "no exported metrics") {
		t.Errorf("output = %q", buf.String())
	}
}

func TestWriteRows(t *testing.T) {
	var buf bytes.Buffer
	rows := []Row{
		{Service: "api-a", SLO: "availability", BurnRate: 0.5, Threshold: 14.4},
		{Service: "api-a", SLO: "latency", BurnRate: 20, Threshold: 14.4},
	}
	if err := Write(&buf, rows); err != nil {
		t.Fatalf("Write: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"SERVICE", "BURN RATE", "THRESHOLD", "api-a", "availability", "latency", "0.50", "20.00", "ok", "BURNING"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestAlertingZeroThreshold(t *testing.T) {
	row := Row{BurnRate: 100, Threshold: 0}
	if row.Alerting() {
		t.Error("Alerting() = true for a zero threshold, want false")
	}
}

// Package slo drives the in-cluster slo-generator CronJob and reads the
// error-budget metrics it exports to Prometheus.
package slo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strconv"
	"time"
)

// Queries read the metrics written by the slo-generator Prometheus exporter.
const (
	// BurnRateQuery returns the current error-budget burn rate per SLO.
	BurnRateQuery = `max by (service_name, slo_name) (error_budget_burn_rate)`
	// ThresholdQuery returns the alerting burn-rate threshold per SLO.
	ThresholdQuery = `max by (service_name, slo_name) (alerting_burn_rate_threshold)`
)

// Runner is the subset of *k8s.Client the SLO workflows need.
type Runner interface {
	Kubectl(ctx context.Context, args ...string) error
	Output(ctx context.Context, args ...string) ([]byte, error)
}

// Row is one SLO's exported burn-rate state.
type Row struct {
	Service   string
	SLO       string
	BurnRate  float64
	Threshold float64
}

// Alerting reports whether the burn rate exceeds a positive threshold.
func (r Row) Alerting() bool {
	return r.Threshold > 0 && r.BurnRate > r.Threshold
}

// Client drives the slo-generator CronJob and Prometheus.
type Client struct {
	runner Runner
	now    func() time.Time
}

// Option configures a Client.
type Option func(*Client)

// WithClock overrides the clock used to name one-off jobs.
func WithClock(now func() time.Time) Option {
	return func(c *Client) { c.now = now }
}

// New returns a Client bound to the given runner.
func New(r Runner, opts ...Option) *Client {
	c := &Client{runner: r, now: time.Now}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Compute starts a one-off Job from the slo-generator CronJob and waits for it
// to complete. It returns the generated Job name.
func (c *Client) Compute(ctx context.Context, namespace, cronjob string) (string, error) {
	name := fmt.Sprintf("%s-manual-%d", cronjob, c.now().Unix())

	if err := c.runner.Kubectl(ctx, "create", "job", name,
		"--namespace", namespace,
		"--from", "cronjob/"+cronjob,
	); err != nil {
		return "", fmt.Errorf("creating slo-generator job: %w", err)
	}
	if err := c.runner.Kubectl(ctx, "wait",
		"--namespace", namespace,
		"--for=condition=complete", "job/"+name,
		"--timeout=120s",
	); err != nil {
		return name, fmt.Errorf("waiting for slo-generator job %q: %w", name, err)
	}
	return name, nil
}

// Status reads the exported burn rate and threshold for every SLO through the
// Kubernetes service proxy (no port-forward required).
func (c *Client) Status(ctx context.Context, namespace, service string, port int) ([]Row, error) {
	burn, err := c.query(ctx, namespace, service, port, BurnRateQuery)
	if err != nil {
		return nil, err
	}
	thresholds, err := c.query(ctx, namespace, service, port, ThresholdQuery)
	if err != nil {
		return nil, err
	}

	rows := make([]Row, 0, len(burn))
	for key, rate := range burn {
		rows = append(rows, Row{
			Service:   key.service,
			SLO:       key.slo,
			BurnRate:  rate,
			Threshold: thresholds[key],
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Service != rows[j].Service {
			return rows[i].Service < rows[j].Service
		}
		return rows[i].SLO < rows[j].SLO
	})
	return rows, nil
}

type seriesKey struct {
	service string
	slo     string
}

func (c *Client) query(ctx context.Context, namespace, service string, port int, expr string) (map[seriesKey]float64, error) {
	path := fmt.Sprintf("/api/v1/namespaces/%s/services/http:%s:%d/proxy/api/v1/query?%s",
		url.PathEscape(namespace), url.PathEscape(service), port,
		url.Values{"query": []string{expr}}.Encode(),
	)

	out, err := c.runner.Output(ctx, "get", "--raw", path)
	if err != nil {
		return nil, fmt.Errorf("querying prometheus: %w", err)
	}

	var resp promResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("decoding prometheus response: %w", err)
	}
	if resp.Status != "success" {
		return nil, fmt.Errorf("prometheus query failed with status %q", resp.Status)
	}

	result := make(map[seriesKey]float64, len(resp.Data.Result))
	for _, s := range resp.Data.Result {
		value, err := s.value()
		if err != nil {
			return nil, err
		}
		result[seriesKey{service: s.Metric["service_name"], slo: s.Metric["slo_name"]}] = value
	}
	return result, nil
}

// Write renders the rows as an aligned table.
func Write(w io.Writer, rows []Row) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "slo: no exported metrics yet — run `srectl slo compute` and try again")
		return err
	}
	if _, err := fmt.Fprintf(w, "%-10s %-14s %9s %10s  %s\n",
		"SERVICE", "SLO", "BURN RATE", "THRESHOLD", "STATE"); err != nil {
		return err
	}
	for _, row := range rows {
		state := "ok"
		if row.Alerting() {
			state = "BURNING"
		}
		if _, err := fmt.Fprintf(w, "%-10s %-14s %9.2f %10.2f  %s\n",
			row.Service, row.SLO, row.BurnRate, row.Threshold, state); err != nil {
			return err
		}
	}
	return nil
}

type promResponse struct {
	Status string `json:"status"`
	Data   struct {
		Result []promSeries `json:"result"`
	} `json:"data"`
}

type promSeries struct {
	Metric map[string]string `json:"metric"`
	Value  []json.RawMessage `json:"value"`
}

func (s promSeries) value() (float64, error) {
	if len(s.Value) < 2 {
		return 0, fmt.Errorf("prometheus series %v has no value", s.Metric)
	}
	var raw string
	if err := json.Unmarshal(s.Value[1], &raw); err != nil {
		return 0, fmt.Errorf("decoding prometheus value: %w", err)
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing prometheus value %q: %w", raw, err)
	}
	return v, nil
}

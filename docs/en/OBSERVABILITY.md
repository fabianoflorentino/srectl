# Observability

How metrics, logs and traces flow through srectl, and how we turn them into
actionable SLOs and alerts.

## The three signals

| Signal | Pipeline | Backend | Query |
| --- | --- | --- | --- |
| Metrics | app `/metrics` → Prometheus (scrape via ServiceMonitor) | Prometheus | PromQL |
| Logs | app `slog` stdout → Alloy → Loki | Loki | LogQL |
| Traces | OTel SDK → OTel Collector → Tempo | Tempo | TraceQL |

Grafana is provisioned with all three datasources so a single query can pivot across
signals (trace → logs → metrics).

## Instrumentation contract (demo APIs)

Every workload we own emits the **RED** method signals plus a few extras:

```text
http_requests_total{service, method, route, status}
http_request_duration_seconds{service, method, route, status}  # histogram
http_errors_total{service, route, status}
process_cpu_seconds_total
process_resident_memory_bytes
```

- **api-a** and **api-b** use the OpenTelemetry Go SDK + `prometheus/client_golang`.
- Structured logs use `log/slog` (JSON on stdout) with `trace_id`/`span_id` fields for
  log→trace correlation.
- Every inbound request is a span, and api-a's outbound call to api-b propagates the
  trace context (W3C `traceparent`).

## Our projects' metrics

- **certificate-validate** (`k8s monitor`, `:9102`):
  `certificate_days_left{namespace,name,kind}`, `certificate_expired`,
  `certificate_revoked`.
- **stracectl** (`--serve :8080`): per-syscall call counts, error rates, P50/P95/P99
  latency histograms.
- **fraudctl**: observed via exporter sidecar / telemetry logs (no native `/metrics`).

## Golden dashboards

1. **Cluster overview** — node CPU/mem/disk, pod count, restarts, control-plane health.
2. **App RED** — request rate, error rate, latency percentiles per service/route.
3. **Distributed trace explorer** — Tempo search + service graph.
4. **Syscall (stracectl)** — top syscalls, latency percentiles, anomalies.
5. **Certificates (certificate-validate)** — days-left by namespace, expiring list.
6. **Security posture** — Trivy vulnerabilities, Kyverno violations, Falco events.

## Alerting

Prometheus `PrometheusRule`s (via the Operator) route to Alertmanager, which fans out to
Grafana + webhook receivers. Falco also emits events to Alertmanager.

Baseline rules:

```text
HighErrorRate       http_errors_total / http_requests_total > 0.05  (5m)
HighLatencyP99      histogram_quantile(0.99, ...) > 500ms          (5m)
CertExpiring        certificate_days_left < 15
CertRevoked         certificate_revoked == 1
NodePressure        node_filesystem/node_memory pressure
PodCrashLooping     kube_pod_container_status_waiting reason=CrashLoopBackOff
SyscallAnomaly      syscall error rate spike (stracectl)
```

## SLO / SLI definitions

| Service | SLI | SLO |
| --- | --- | --- |
| api-a `/orders` | availability = good/(good+bad) over 5m | 99.5% (monthly) |
| api-a `/orders` | latency p99 | < 300ms |
| fraudctl `/fraud-score` | availability + latency p99 | 99.5% / < 20ms |
| TLS certs | days-left > 0 | 100% (no expiry) |

Recording rules pre-compute the SLI time series; Grafana SLO dashboards show burn-rate.

## Chaos → observability loop

chaos-mesh injects faults; the observability stack is how we *detect* them:

1. Inject 200ms latency on api-a (`NetworkChaos`).
2. Watch p99 rise past the SLO threshold on the RED dashboard.
3. Confirm `HighLatencyP99` fires in Alertmanager.
4. Kill fraudctl pods; confirm self-healing restores availability within budget.

## Local access

| Surface | Command |
| --- | --- |
| Grafana | `srectl open grafana` (default admin/cred from secret) |
| Prometheus | `srectl open prometheus` |
| Alertmanager | `srectl open alertmanager` |
| App metrics | `srectl open api-a` → `/metrics` |

All are also reachable via ingress (`grafana.localhost`, `prometheus.localhost`, ...)
once ingress + cert-manager are up.

---
title: "SRE"
description: "SLOs, alerts, chaos drills and incident response on a live cluster."
---

Reliability is a practice. This path turns observability into action.

## SLO / SLI

The platform defines SLIs and objectives as Prometheus recording rules
(`deploy/observability/slo-rules.yaml`):

- `slo:api_a_availability:ratio5m` — good `/orders` requests / total (objective 99.5%).
- `slo:api_a_latency_p95:5m` — p95 latency of `/orders` (objective < 20ms).
- `slo:fraudctl_availability:ratio5m` and `slo:fraudctl_latency_p95:5m` (via ingress-nginx).

Alerts `ApiAHighErrorRate` and `ApiAHighLatency` fire when the objectives are missed.

## Drills

```bash
srectl demo load                 # traffic → RED dashboards
srectl demo chaos latency --duration 90s   # CPU stress on api-a → latency rises
srectl demo chaos pod-kill --duration 30s  # kill fraudctl → self-healing
srectl demo verify               # cluster up + every app Synced · Healthy
```

Open Grafana (`srectl open grafana`) and watch `srectl · App RED`, `srectl · Golden
Signals` and `srectl · Cluster` react. Then follow the [runbooks](/runbooks/).

> Note: on kind, NetworkChaos (packet loss / DNS) needs iptables/ipset that do not work
> in nested containers — the latency drill uses CPU stress instead.

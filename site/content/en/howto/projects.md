---
title: "Our projects under load"
description: "Exercise fraudctl, stracectl and certificate-validate as real workloads."
weight: 8
---

**SRE concept:** **observing heterogeneous workloads.** Not every service speaks the same
metrics. Sometimes you observe from the outside (black box), sometimes you go deep (syscalls),
sometimes you watch a specific resource (certificates).

**Prerequisites:** `srectl up`, `srectl demo verify`.

## fraudctl — high-performance API (black box + ingress)

```bash
srectl open fraudctl     # http://localhost:8084
curl -s -X POST http://localhost:8084/fraud-score -H 'Content-Type: application/json' -d '{
  "id":"tx-1","transaction":{"amount":384.88,"installments":3,"requested_at":"2026-03-11T20:23:35Z"},
  "customer":{"avg_amount":769.76,"tx_count_24h":3,"known_merchants":["MERC-001"]},
  "merchant":{"id":"MERC-001","mcc":"5912","avg_amount":298.95},
  "terminal":{"is_online":false,"card_present":true,"km_from_home":13.7}}'
```

It has no native `/metrics`, so throughput/latency come from ingress-nginx — see
`srectl · Golden Signals` with `ingress=fraudctl`.

## stracectl — deep observability (syscalls)

```bash
srectl trace --daemon    # → http://localhost:8080 (/api/stats, /metrics)
```

Per-syscall latency, error rates and counts land in `srectl · Syscall` (Prometheus).

## certificate-validate — resource-specific SLI

`srectl · Certificates` shows the days-left per TLS secret (see the certificates how-to).

## Why it matters (SRE)

Real platforms mix instrumented services, opaque binaries and infrastructure-level checks. The
skill is choosing the right vantage point per workload — and correlating them in one place
(Grafana) so an incident is one investigation, not five.

## Expected outcome

fraudctl throughput/latency in Golden Signals, stracectl syscall metrics in Prometheus, and
certificate days-left in Grafana — all in the same UI.

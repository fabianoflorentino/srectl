---
title: "Tutorials"
description: "Hands-on, driven by srectl demo scenarios."
---

Prerequisites: `srectl up` and `srectl status` (all apps Synced · Healthy).

## 1. Observe traffic

```bash
srectl demo load --duration 30s --rps 20
srectl open grafana   # → Dashboards → srectl · App RED / Golden Signals
```

Watch request rate, 5xx ratio and p50/p95/p99 for `api-a` and `api-b`.

## 2. Follow a trace

Grafana → Explore → **Tempo** → service `api-a`. Open a `/orders` trace: it spans
`api-a → api-b` under one traceID.

## 3. Break it on purpose

```bash
srectl demo chaos latency --duration 90s   # CPU stress → /orders p95 rises
srectl demo chaos pod-kill --duration 30s  # fraudctl pod killed → self-heals
```

Watch `srectl · App RED` and `srectl · Cluster`, then `srectl demo verify`.

## 4. Hunt a threat

```bash
kubectl exec -n demo deploy/hello -- cat /etc/shadow   # Falco: "Read sensitive file untrusted"
```

Grafana → Explore → **Loki** → `{namespace="demo"} |= "shadow"`.

## 5. Validate certificates

Grafana → `srectl · Certificates`: days left per certificate, sourced from
certificate-validate's `certificate_days_left`.

> Prefer concept-first pages? See the **[how-tos](../howto/)** — one per simulation.

---
title: "Metrics, logs & traces correlation"
description: "Follow one request across the three pillars using a trace ID."
weight: 5
---

**SRE concept:** the **three pillars of observability** — metrics (how much), logs (what
happened) and traces (where the time went). They are far more useful **correlated** than in
isolation.

**Prerequisites:** `srectl up`, `srectl demo verify`, traffic running.

## Run it

```bash
srectl demo load --duration 30s --rps 20
srectl open grafana
```

## What you'll see

- **Traces** — Explore → **Tempo** → service `api-a`. Open a `/orders` trace: it spans
  **api-a → api-b** under a single traceID.
- **Metrics** — `srectl · App RED`: rate/errors/duration for the same endpoints.
- **Logs** — Explore → **Loki** → `{namespace="demo", app="api-a"}`. Every line carries a
  `trace_id` (the slog logger injects it), so a log points back to its trace.

## Why it matters (SRE)

During an incident you start from a metric (p99 spiked), jump to a trace (the slow span is the
api-b call), then to the logs of that exact trace. Correlation is what turns three dashboards
into one story. Keep **traceID + spanID on every log line** and bounded metric labels (route,
status) to stay within cardinality limits.

## Expected outcome

One traceID links the metric spike, the trace and the log lines for the same request.

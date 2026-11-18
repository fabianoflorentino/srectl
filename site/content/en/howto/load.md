---
title: "Load & the RED method"
description: "Generate traffic and read Rate, Errors and Duration for api-a/api-b."
weight: 1
---

**SRE concept:** the **RED method** — for every service track **R**ate (requests/s),
**E**rrors (failure ratio) and **D**uration (latency distribution). It is the fastest way
to know whether a service is healthy.

**Prerequisites:** `srectl up` and `srectl demo verify`.

## Run it

```bash
srectl demo load --duration 30s --rps 20
srectl open grafana            # → Dashboards → srectl · App RED
```

The generator hits `/orders` (which calls api-b), `/slow` and `/error`.

## What you'll see

- **Rate** — `sum by (route) (rate(api_a_http_requests_total[5m]))`.
- **Errors** — the 5xx ratio on `srectl · App RED` (api-a/api-b).
- **Duration** — p50/p95/p99 from the histogram
  `api_a_http_request_duration_seconds_bucket`.
- Per-route and per-status breakdowns, plus the traffic on `srectl · Golden Signals`.

## Why it matters (SRE)

RED is the minimum instrumentation every service should expose. It answers "is it up, is it
fast, is it failing?" without guessing. Because the metrics are histograms, percentiles are
computed across replicas — not per-process — so they stay meaningful as you scale.

## Expected outcome

Active traffic, a few 5xx (the `/error` route) and a non-zero p95/p99 — the raw signal every
other drill builds on.

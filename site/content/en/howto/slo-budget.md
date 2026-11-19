---
title: "Burn an error budget"
description: "Compute SLOs with slo-generator, then use chaos to spend a real error budget."
weight: 9
---

**SRE concept:** the **error budget** (`1 − SLO`) is the reliability you are
allowed to spend. The **burn rate** is how fast it is going. This drill makes both
numbers move on a live cluster.

**Prerequisites:** `srectl up`, `srectl demo verify`, the `slo` Argo app
`Synced · Healthy`.

## Run it

```bash
srectl slo compute
srectl slo status
srectl demo load --duration 120s --rps 30
srectl demo chaos latency --duration 120s
srectl slo status
```

## What you'll see

- The first `slo status` shows every SLO below its burn-rate threshold (`ok`).
- During the chaos drill, `api-a / latency` crosses the 1h threshold and shows
  `BURNING`; the Grafana dashboard **srectl · SLO / Error Budget** turns red.
- After the experiment expires, the burn rate returns below the threshold.

## Why it matters (SRE)

The policy encodes the decision the on-call engineer must make: page at 1h/6h,
freeze releases at 3d. You are practising the **error-budget policy**, not just
reading a graph.

## Expected outcome

You can explain why a fast burn (1h) is paged while a slow burn (3d) only freezes
releases — and you have seen the budget respond to a real fault.

Deep dive: [SLO-as-code with slo-generator](/sre/slo-generator/).

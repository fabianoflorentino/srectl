---
title: "Latency chaos & the error budget"
description: "Stress api-a CPU, watch latency rise and the SLO alert fire."
weight: 2
---

**SRE concept:** **SLOs, SLIs and the error budget.** An SLO is a target (e.g. 99.5%
availability, p95 < 20ms). The SLI is how you measure it; the **error budget** is how much
you may fail before the SLO is broken. Alerts should fire on *burn*, not on raw symptoms.

**Prerequisites:** `srectl up`, `srectl demo verify`, and a little load running.

## Run it

```bash
srectl demo load --duration 120s --rps 20 &
srectl demo chaos latency --duration 90s
```

On kind, latency is induced with **CPU stress** (NetworkChaos needs iptables/ipsets that do
not work in nested containers).

## What you'll see

- `srectl · App RED` — `/orders` p95 climbs from ~5ms to tens of ms.
- `srectl · Cluster` — api-a CPU saturation.
- The SLO recording rules: `slo:api_a_latency_p95:5m` and `slo:api_a_availability:ratio5m`.
- The **`ApiAHighLatency`** alert (`> 20ms` for 30s) fires in Alertmanager.

## Why it matters (SRE)

This is a controlled burn: you *spend* error budget on purpose, observe the SLI move, and
confirm the alert fires and then clears. That closes the loop between *measurement*, *policy*
(the SLO) and *paging* — the core of operating a service against objectives instead of guesswork.

## Expected outcome

p99 rises, `ApiAHighLatency` goes firing, then recovers as the stress ends (the chaos has a
built-in `duration`, so it self-recovers even if the CLI dies).

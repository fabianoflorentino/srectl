---
title: "SLO-as-code with slo-generator"
description: "Model SLIs, SLOs, error budgets and burn-rate policies as declarative resources, then burn a budget on purpose."
weight: 1
---

**SRE concept:** an SLO is a target for a service level indicator. The **error
budget** is `1 − SLO`; the **burn rate** is how fast you are spending it. Alerting
on burn rate is more stable than alerting on a raw threshold.

srectl ships two layers of SLOs:

- **Recording rules** in `deploy/observability/slo-rules.yaml` — fast SLI series
  for dashboards.
- **`slo-generator`** in `deploy/slo/` — declarative `ServiceLevelObjective`
  resources that compute the budget, the burn rate and the policy.

The second layer runs as a CronJob in the `monitoring` namespace. It queries
Prometheus, pushes `error_budget_burn_rate`, `sli_measurement` and `slo_target`
to a Pushgateway, and Prometheus scrapes them back.

## Run it

```bash
srectl up
srectl demo verify
srectl slo compute     # one-off slo-generator Job
srectl slo status      # service · SLO · burn rate · threshold · state
```

Then open Grafana and select **srectl · SLO / Error Budget**:

```bash
srectl open grafana
```

## What you'll see

- **Burn rate vs threshold** — one line per SLO, plus the alerting threshold.
- **Current burn rate table** — red cells mean the SLO is burning budget too fast.
- **SLI vs target** — whether the measured SLI is above the `goal`.

The bundled SLOs:

| SLO | Source metric | Goal |
| --- | --- | --- |
| `api-a` availability | `api_a_http_requests_total` (non-5xx) | 99.5% |
| `api-a` latency | `api_a_http_request_duration_seconds_bucket` ≤ 20ms | 99% |
| `fraudctl` availability | `nginx_ingress_controller_requests` (non-5xx) | 99.5% |

## Lab: burn the budget

1. **Baseline:** `srectl slo compute && srectl slo status` — every row `ok`.
2. **Load:** `srectl demo load --duration 120s --rps 30`.
3. **Break:** `srectl demo chaos latency --duration 120s` (CPU stress on api-a).
4. **Watch:** run `srectl slo status` while the drill runs; the `api-a / latency`
   burn rate crosses the 1h threshold and the dashboard turns red.
5. **Decide:** the policy pages at 1h/6h and freezes releases at 3d. What would
   you do as the on-call engineer?
6. **Recover:** let the experiment expire and confirm the burn rate drops.

## Why it matters (SRE)

A budget turns reliability into a budget you can spend: ship faster while the
budget is healthy, slow down when it is not. Multi-window burn-rate alerting
catches both fast outages (1h) and slow erosion (3d) without drowning the team
in pages. Writing the objective, the policy and the alert in one versioned YAML
is also how mature teams keep SLOs reviewable.

## Expected outcome

You can explain the difference between an SLI, an SLO, an error budget and a
burn rate — and you have watched a chaos drill consume a real error budget.

Full guide: [`docs/en/SLO.md`](https://github.com/fabianoflorentino/srectl/blob/main/docs/en/SLO.md).

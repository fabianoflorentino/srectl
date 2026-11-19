---
title: "SRE"
description: "SLOs, alertas, drills de caos e resposta a incidentes num cluster real."
---

Confiabilidade é prática. Esta trilha transforma observabilidade em ação.

## SLO / SLI

A plataforma define SLIs e objetivos em duas camadas:

**Recording rules** (`deploy/observability/slo-rules.yaml`) — séries de SLI rápidas:

- `slo:api_a_availability:ratio5m` — requisições boas em `/orders` / total (objetivo 99,5%).
- `slo:api_a_latency_p95:5m` — p95 de `/orders` (objetivo < 20ms).
- `slo:fraudctl_availability:ratio5m` e `slo:fraudctl_latency_p95:5m` (via ingress-nginx).

**SLO-as-code** (`deploy/slo/`) — o [slo-generator](/pt-br/sre/slo-generator/) calcula
o error budget, o burn rate e uma política 1h/6h/3d a partir de recursos
`ServiceLevelObjective` declarativos:

```bash
srectl slo compute     # Job avulso do slo-generator
srectl slo status      # burn rate · limiar · estado por SLO
```

Os alertas `ApiAHighErrorRate` e `ApiAHighLatency` disparam quando os objetivos falham;
`error_budget_burn_rate` mostra quão rápido o budget está sendo gasto.

## Drills

```bash
srectl demo load                 # tráfego → dashboards RED
srectl demo chaos latency --duration 90s   # stress de CPU na api-a → latência sobe
srectl demo chaos pod-kill --duration 30s  # mata o fraudctl → self-healing
srectl demo verify               # cluster up + todos os apps Synced · Healthy
srectl slo compute               # recalcula error budgets + burn rates
srectl slo status                # tabela de burn rate
```

Comece por [SLO-as-code com slo-generator](/pt-br/sre/slo-generator/) para queimar
um budget de propósito.

Abra o Grafana (`srectl open grafana`) e veja `srectl · App RED`, `srectl · Golden
Signals` e `srectl · Cluster` reagirem. Depois siga os [runbooks](/pt-br/runbooks/).

> Nota: no kind, NetworkChaos (perda de pacotes / DNS) precisa de iptables/ipset que não
> funcionam em containers aninhados — o drill de latência usa stress de CPU.

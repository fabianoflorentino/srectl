---
title: "SLO-as-code com slo-generator"
description: "Modele SLIs, SLOs, error budgets e políticas de burn rate como recursos declarativos e queime um budget de propósito."
weight: 1
---

**Conceito de SRE:** um SLO é uma meta para um service level indicator. O
**error budget** é `1 − SLO`; o **burn rate** é a velocidade com que você o gasta.
Alertar em burn rate é mais estável do que alertar num limiar cru.

O srectl entrega SLOs em duas camadas:

- **Recording rules** em `deploy/observability/slo-rules.yaml` — séries de SLI
  rápidas para dashboards.
- **`slo-generator`** em `deploy/slo/` — recursos `ServiceLevelObjective`
  declarativos que calculam o budget, o burn rate e a política.

A segunda camada roda como CronJob no namespace `monitoring`. Ela consulta o
Prometheus, empurra `error_budget_burn_rate`, `sli_measurement` e `slo_target`
para um Pushgateway, e o Prometheus os raspa de volta.

## Rode

```bash
srectl up
srectl demo verify
srectl slo compute     # Job avulso do slo-generator
srectl slo status      # serviço · SLO · burn rate · limiar · estado
```

Depois abra o Grafana e selecione **srectl · SLO / Error Budget**:

```bash
srectl open grafana
```

## O que você verá

- **Burn rate vs limiar** — uma linha por SLO, mais o limiar de alerta.
- **Tabela de burn rate atual** — células vermelhas significam budget queimando
  rápido demais.
- **SLI vs meta** — se a SLI medida está acima do `goal`.

Os SLOs incluídos:

| SLO | Métrica de origem | Meta |
| --- | --- | --- |
| disponibilidade `api-a` | `api_a_http_requests_total` (não-5xx) | 99,5% |
| latência `api-a` | `api_a_http_request_duration_seconds_bucket` ≤ 20ms | 99% |
| disponibilidade `fraudctl` | `nginx_ingress_controller_requests` (não-5xx) | 99,5% |

## Lab: queime o budget

1. **Baseline:** `srectl slo compute && srectl slo status` — todas as linhas `ok`.
2. **Carga:** `srectl demo load --duration 120s --rps 30`.
3. **Quebre:** `srectl demo chaos latency --duration 120s` (stress de CPU na api-a).
4. **Observe:** rode `srectl slo status` durante o drill; o burn rate de
   `api-a / latency` cruza o limiar de 1h e o dashboard fica vermelho.
5. **Decida:** a política pagina em 1h/6h e congela releases em 3d. O que você
   faria como plantonista?
6. **Recupere:** deixe o experimento expirar e confirme que o burn rate cai.

## Por que isso importa (SRE)

O budget transforma confiabilidade num orçamento que pode ser gasto: entregue
mais rápido enquanto o budget está saudável, desacelere quando não está. Alerta
multi-janela de burn rate pega tanto indisponibilidades rápidas (1h) quanto
erosão lenta (3d) sem afogar o time em pages. Escrever objetivo, política e
alerta num único YAML versionado também é como times maduros mantêm SLOs
revisáveis.

## Resultado esperado

Você sabe explicar a diferença entre SLI, SLO, error budget e burn rate — e viu
um drill de caos consumir um error budget real.

Guia completo: [`docs/pt-br/SLO.md`](https://github.com/fabianoflorentino/srectl/blob/main/docs/pt-br/SLO.md).

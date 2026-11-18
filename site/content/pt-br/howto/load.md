---
title: "Carga e o método RED"
description: "Gere tráfego e leia Rate, Errors e Duration da api-a/api-b."
weight: 1
---

**Conceito de SRE:** o **método RED** — para cada serviço acompanhe **R**ate
(requisições/s), **E**rrors (razão de falha) e **D**uration (distribuição de latência). É a
forma mais rápida de saber se um serviço está saudável.

**Pré-requisitos:** `srectl up` e `srectl demo verify`.

## Rodar

```bash
srectl demo load --duration 30s --rps 20
srectl open grafana            # → Dashboards → srectl · App RED
```

O gerador bate em `/orders` (que chama a api-b), `/slow` e `/error`.

## O que você vê

- **Rate** — `sum by (route) (rate(api_a_http_requests_total[5m]))`.
- **Errors** — a razão 5xx no `srectl · App RED` (api-a/api-b).
- **Duration** — p50/p95/p99 do histograma `api_a_http_request_duration_seconds_bucket`.
- Quebras por rota/status e o tráfego no `srectl · Golden Signals`.

## Por que importa (SRE)

RED é a instrumentação mínima que todo serviço deveria expor: está no ar? está rápido? está
falhando? Como as métricas são histogramas, os percentis são agregados entre réplicas — não por
processo — então continuam válidos ao escalar.

## Resultado esperado

Tráfego ativo, alguns 5xx (rota `/error`) e p95/p99 não-zero — o sinal bruto que os outros drills
reaproveitam.

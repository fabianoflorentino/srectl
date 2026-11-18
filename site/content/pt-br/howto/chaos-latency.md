---
title: "Caos de latência e o error budget"
description: "Estresse a CPU da api-a, veja a latência subir e o alerta de SLO disparar."
weight: 2
---

**Conceito de SRE:** **SLOs, SLIs e o error budget.** Um SLO é uma meta (ex.: 99,5% de
disponibilidade, p95 < 20ms). O SLI é como se mede; o **error budget** é quanto você pode falhar
antes de quebrar o SLO. Alertas devem disparar no *consumo*, não no sintoma bruto.

**Pré-requisitos:** `srectl up`, `srectl demo verify` e um pouco de carga rodando.

## Rodar

```bash
srectl demo load --duration 120s --rps 20 &
srectl demo chaos latency --duration 90s
```

No kind, a latência é induzida com **stress de CPU** (NetworkChaos precisa de iptables/ipsets
que não funcionam em containers aninhados).

## O que você vê

- `srectl · App RED` — p95 de `/orders` sobe de ~5ms para dezenas de ms.
- `srectl · Cluster` — saturação de CPU da api-a.
- As regras de gravação: `slo:api_a_latency_p95:5m` e `slo:api_a_availability:ratio5m`.
- O alerta **`ApiAHighLatency`** (`> 20ms` por 30s) dispara no Alertmanager.

## Por que importa (SRE)

É uma queima controlada: você *gasta* error budget de propósito, observa o SLI mexer e confirma
que o alerta dispara e depois limpa. Isso fecha o ciclo entre *medição*, *política* (o SLO) e
*paging* — o núcleo de operar por objetivos em vez de achismo.

## Resultado esperado

p99 sobe, o `ApiAHighLatency` vai a firing, e recupera quando o stress termina (o chaos tem
`duration` embutida, então se recupera mesmo se o CLI morrer).

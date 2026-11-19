---
title: "Queime um error budget"
description: "Calcule SLOs com o slo-generator e use caos para gastar um error budget real."
weight: 9
---

**Conceito de SRE:** o **error budget** (`1 − SLO`) é a confiabilidade que você
pode gastar. O **burn rate** é a velocidade com que isso acontece. Este drill faz
os dois números se moverem num cluster real.

**Pré-requisitos:** `srectl up`, `srectl demo verify`, o app Argo `slo`
`Synced · Healthy`.

## Rode

```bash
srectl slo compute
srectl slo status
srectl demo load --duration 120s --rps 30
srectl demo chaos latency --duration 120s
srectl slo status
```

## O que você verá

- O primeiro `slo status` mostra todos os SLOs abaixo do limiar de burn rate (`ok`).
- Durante o drill de caos, `api-a / latency` cruza o limiar de 1h e mostra
  `BURNING`; o dashboard **srectl · SLO / Error Budget** fica vermelho.
- Depois que o experimento expira, o burn rate volta para baixo do limiar.

## Por que isso importa (SRE)

A política codifica a decisão que o plantonista precisa tomar: pagina em 1h/6h,
congela releases em 3d. Você está praticando a **política de error budget**, não
só lendo um gráfico.

## Resultado esperado

Você sabe explicar por que um burn rápido (1h) vira page enquanto um burn lento
(3d) só congela releases — e viu o budget reagir a uma falha real.

Aprofunde: [SLO-as-code com slo-generator](/pt-br/sre/slo-generator/).

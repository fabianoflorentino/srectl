---
title: "Nossos projetos sob carga"
description: "Exercite fraudctl, stracectl e certificate-validate como workloads reais."
weight: 8
---

**Conceito de SRE:** **observar workloads heterogêneos.** Nem todo serviço fala a mesma
métrica. Às vezes você observa de fora (caixa-preta), às vezes vai fundo (syscalls), às vezes
acompanha um recurso específico (certificados).

**Pré-requisitos:** `srectl up`, `srectl demo verify`.

## fraudctl — API de alta performance (caixa-preta + ingress)

```bash
srectl open fraudctl     # http://localhost:8084
curl -s -X POST http://localhost:8084/fraud-score -H 'Content-Type: application/json' -d '{
  "id":"tx-1","transaction":{"amount":384.88,"installments":3,"requested_at":"2026-03-11T20:23:35Z"},
  "customer":{"avg_amount":769.76,"tx_count_24h":3,"known_merchants":["MERC-001"]},
  "merchant":{"id":"MERC-001","mcc":"5912","avg_amount":298.95},
  "terminal":{"is_online":false,"card_present":true,"km_from_home":13.7}}'
```

Ele não tem `/metrics` nativo, então throughput/latência vêm do ingress-nginx — veja
`srectl · Golden Signals` com `ingress=fraudctl`.

## stracectl — observabilidade profunda (syscalls)

```bash
srectl trace --daemon    # → http://localhost:8080 (/api/stats, /metrics)
```

Latência, taxa de erro e contagem por syscall chegam ao `srectl · Syscall` (Prometheus).

## certificate-validate — SLI de um recurso específico

`srectl · Certificates` mostra os dias-restantes por secret TLS (ver o how-to de certificados).

## Por que importa (SRE)

Plataformas reais misturam serviços instrumentados, binários opacos e checagens de
infraestrutura. A habilidade é escolher o ponto de observação certo por workload — e correlacionar
tudo num lugar (Grafana) para o incidente ser uma investigação, não cinco.

## Resultado esperado

Throughput/latência do fraudctl no Golden Signals, métricas de syscall do stracectl no Prometheus
e dias-restantes de certificados no Grafana — tudo na mesma UI.

## seraph — status board estático

```bash
srectl open seraph     # http://localhost:8085
```

Serviço nginx com imagem pinada que serve um snapshot de status (`/data/state.json`)
atualizado pelo `seraph sync`. Observado pelas métricas RED do ingress-nginx e Loki; sem `/metrics`.

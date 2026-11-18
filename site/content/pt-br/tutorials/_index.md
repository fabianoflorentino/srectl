---
title: "Tutoriais"
description: "Práticos, guiados pelos cenários do srectl demo."
---

Pré-requisitos: `srectl up` e `srectl status` (todos os apps Synced · Healthy).

## 1. Observar tráfego

```bash
srectl demo load --duration 30s --rps 20
srectl open grafana   # → Dashboards → srectl · App RED / Golden Signals
```

Veja taxa de requisições, razão 5xx e p50/p95/p99 de `api-a` e `api-b`.

## 2. Seguir um trace

Grafana → Explore → **Tempo** → serviço `api-a`. Abra um trace de `/orders`: ele cobre
`api-a → api-b` num único traceID.

## 3. Quebrar de propósito

```bash
srectl demo chaos latency --duration 90s   # stress de CPU → p95 de /orders sobe
srectl demo chaos pod-kill --duration 30s  # pod do fraudctl morre → self-heal
```

Veja `srectl · App RED` e `srectl · Cluster`, depois `srectl demo verify`.

## 4. Caçar uma ameaça

```bash
kubectl exec -n demo deploy/hello -- cat /etc/shadow   # Falco: "Read sensitive file untrusted"
```

Grafana → Explore → **Loki** → `{namespace="demo"} |= "shadow"`.

## 5. Validar certificados

Grafana → `srectl · Certificates`: dias restantes por certificado, vindos do
`certificate_days_left` do certificate-validate.

> Prefere páginas concept-first? Veja os **[guias](../howto/)** — um por simulação.

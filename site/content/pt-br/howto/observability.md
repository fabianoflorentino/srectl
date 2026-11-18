---
title: "Correlação de métricas, logs e traces"
description: "Siga uma requisição pelos três pilares usando o trace ID."
weight: 5
---

**Conceito de SRE:** os **três pilares da observabilidade** — métricas (quanto), logs (o que
aconteceu) e traces (onde o tempo foi gasto). São muito mais úteis **correlacionados** do que
isolados.

**Pré-requisitos:** `srectl up`, `srectl demo verify`, tráfego rodando.

## Rodar

```bash
srectl demo load --duration 30s --rps 20
srectl open grafana
```

## O que você vê

- **Traces** — Explore → **Tempo** → serviço `api-a`. Abra um trace de `/orders`: ele cobre
  **api-a → api-b** num único traceID.
- **Métricas** — `srectl · App RED`: rate/errors/duration dos mesmos endpoints.
- **Logs** — Explore → **Loki** → `{namespace="demo", app="api-a"}`. Cada linha carrega um
  `trace_id` (o logger slog injeta), então um log aponta de volta pro trace.

## Por que importa (SRE)

Num incidente você começa numa métrica (p99 subiu), salta pro trace (o span lento é a chamada
à api-b) e vai pros logs daquele trace exato. Correlação é o que transforma três dashboards numa
história só. Mantenha **traceID + spanID em toda linha de log** e labels de métrica limitados
(rota, status) para não estourar cardinalidade.

## Resultado esperado

Um traceID liga o pico da métrica, o trace e as linhas de log da mesma requisição.

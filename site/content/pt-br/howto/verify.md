---
title: "Verificação de saúde"
description: "Garanta que o cluster, o Argo CD e todos os apps estão saudáveis num comando."
weight: 4
---

**Conceito de SRE:** **health checks / smoke tests.** Antes e depois de cada drill você precisa
de um veredito rápido e determinístico: a plataforma está no estado esperado?

**Pré-requisitos:** `srectl up`.

## Rodar

```bash
srectl demo verify
srectl status
```

`demo verify` checa: o cluster kind está no ar, o Argo CD está `Ready` e **todos** os apps do
Argo CD estão `Synced · Healthy`. `status` imprime a matriz completa.

## O que você vê

```text
verify: OK — cluster up, argocd ready, 18 apps Synced · Healthy
```

Se algo divergiu, o verify falha e imprime os apps problemáticos, ex.: `api-b: OutOfSync · Degraded`.

## Por que importa (SRE)

Health checks são o detector de incidente mais barato e o portão de toda mudança. Ter isso num
único comando (e não num runbook de `kubectl`) faz humanos e CI validarem o *mesmo* contrato —
exatamente o que o workflow E2E faz (`up → verify → down`).

## Resultado esperado

Um verify verde antes de começar um drill, e de novo após a recuperação.

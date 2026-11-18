---
title: "Política, scanning e detecção em runtime"
description: "Kyverno bloqueia um workload ruim, Trivy reporta CVEs e Falco pega uma intrusão."
weight: 6
---

**Conceito de SRE:** **defesa em profundidade** — prevenir (política de admissão), detectar
antes do deploy (scan de imagem/config) e detectar em runtime (eBPF). Confiabilidade inclui não
colocar artefatos sabidamente ruins em produção.

**Pré-requisitos:** `srectl up`, `srectl demo verify`.

## 1. Prevenir — Kyverno (política como código)

```bash
kubectl run bad -n demo --image=nginx:latest
# rejeitado por disallow-latest-tag (ValidatingPolicy em CEL)
```

O namespace `demo` também exige requests/limits e proíbe `hostPath`; `runAsNonRoot` é auditado.

## 2. Detectar antes do deploy — Trivy Operator

```bash
kubectl get vulnerabilityreport,configauditreport -n demo
kubectl get vulnerabilityreport -n demo -o jsonpath='{.items[0].report.summary}'
```

Os achados também aparecem no `srectl · Security` (severidade, top imagens, config/RBAC).

## 3. Detectar em runtime — Falco

```bash
kubectl exec -n demo deploy/hello -- cat /etc/shadow
```

O Falco (modern eBPF) dispara **"Read sensitive file untrusted"**. Veja no Loki:
`{namespace="demo"} |= "shadow"`.

## Por que importa (SRE)

Cada camada pega o que a anterior deixa passar: a política barra o erro óbvio, o scanning acha
CVEs conhecidas antes de rodarem, a detecção em runtime pega comportamentos que nenhum check
estático preveria. Juntas, encurtam a janela entre "algo ruim aconteceu" e "sabemos disso".

## Resultado esperado

Um deploy rejeitado, reports do Trivy populados e um evento do Falco correlacionado no Loki/Grafana.

---
title: "A stack"
description: "Cada componente, por que existe e como é instalado."
---

| Camada | Componente | Por quê |
| --- | --- | --- |
| Cluster | kind | Kubernetes multi-nó reproduzível no Docker |
| GitOps | Argo CD | App-of-apps, self-heal, drift |
| Ingress / TLS | ingress-nginx, cert-manager | Entrada HTTP(S), certificados automáticos |
| Métricas | Prometheus, Alertmanager, Grafana | Golden signals, RED, alertas |
| Logs | Loki + Alloy | Logs de pods com labels `namespace/pod/app/level` |
| Traces | Tempo (receptores OTLP) | Traces distribuídos api-a → api-b |
| Métricas de cluster | metrics-server, node-exporter, kube-state-metrics | `kubectl top`, saturação, estado dos objetos |
| Segurança | Kyverno, Trivy Operator, Falco | Política, scans de vuln/config, detecção em runtime |
| Caos | chaos-mesh | Drills de pod kill e stress de CPU |
| Demo | api-a, api-b, hello, fraudctl, seraph | Workloads que geram sinal |

As versões ficam em `deploy/bootstrap/apps/*.yaml` (`targetRevision` do Helm).

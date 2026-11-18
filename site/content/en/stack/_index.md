---
title: "The stack"
description: "Every component, why it exists, and how it is installed."
---

| Layer | Component | Why |
| --- | --- | --- |
| Cluster | kind | Reproducible multi-node Kubernetes in Docker |
| GitOps | Argo CD | App-of-apps, self-heal, drift management |
| Ingress / TLS | ingress-nginx, cert-manager | HTTP(S) entrypoint, automatic certificates |
| Metrics | Prometheus, Alertmanager, Grafana | Golden signals, RED, alerting |
| Logs | Loki + Alloy | Pod logs with `namespace/pod/app/level` labels |
| Traces | Tempo + OpenTelemetry | Distributed traces across api-a → api-b |
| Cluster metrics | metrics-server, node-exporter, kube-state-metrics | `kubectl top`, saturation, object state |
| Security | Kyverno, Trivy Operator, Falco | Policy, vulnerability + config scans, runtime detection |
| Chaos | chaos-mesh | Pod kill and CPU stress drills |
| Demo | api-a, api-b, hello, fraudctl | Workloads that produce signal |

Component versions live in `deploy/bootstrap/apps/*.yaml` (Helm `targetRevision`).

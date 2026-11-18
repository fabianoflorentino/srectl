---
title: "Architecture"
description: "How the platform is born: srectl → kind → Argo CD → workloads, all through GitOps."
---

Everything is declarative and reconciled by **Argo CD** (app-of-apps). `srectl` only
bootstraps the cluster and the root application; from there Git is the source of truth.

{{< mermaid >}}
flowchart TD
  U[You] -->|srectl up| K[kind cluster]
  K --> A[Argo CD<br/>app-of-apps]
  A --> APP[demo: api-a → api-b, hello, fraudctl]
  A --> OBS[LGTM: Prometheus, Loki, Tempo, Grafana, Alloy]
  A --> SEC[Kyverno, Trivy Operator, Falco]
  A --> CHAOS[chaos-mesh]
  A --> PROJ[certificate-validate, stracectl]
  U -->|srectl demo load| APP
  APP -->|metrics, logs, traces| OBS
  SEC -->|policy + findings| OBS
{{< /mermaid >}}

The lifecycle:

1. `srectl up` creates the kind cluster and installs Argo CD.
2. Argo CD applies the **root app**, which applies the child apps in
   `deploy/bootstrap/apps/` (ingress, cert-manager, observability, security, apps, projects, chaos).
3. Each child app renders Helm charts or plain manifests from this repo.
4. `srectl status` reads Argo CD's view — the same view you see in the UI.

---
title: "Arquitetura"
description: "Como a plataforma nasce: srectl → kind → Argo CD → workloads, tudo via GitOps."
---

Tudo é declarativo e reconciliado pelo **Argo CD** (app-of-apps). O `srectl` só faz o
bootstrap do cluster e da aplicação raiz; a partir daí o Git é a fonte da verdade.

{{< mermaid >}}
flowchart TD
  U[Você] -->|srectl up| K[cluster kind]
  K --> A[Argo CD<br/>app-of-apps]
  A --> APP[demo: api-a → api-b, hello, fraudctl]
  A --> OBS[LGTM: Prometheus, Loki, Tempo, Grafana, Alloy]
  A --> SEC[Kyverno, Trivy Operator, Falco]
  A --> CHAOS[chaos-mesh]
  A --> PROJ[certificate-validate, stracectl]
  U -->|srectl demo load| APP
  APP -->|métricas, logs, traces| OBS
  SEC -->|políticas + achados| OBS
{{< /mermaid >}}

O ciclo de vida:

1. `srectl up` cria o cluster kind e instala o Argo CD.
2. O Argo CD aplica o **app raiz**, que aplica os apps filhos em
   `deploy/bootstrap/apps/` (ingress, cert-manager, observabilidade, segurança, apps, projetos, caos).
3. Cada app filho renderiza charts Helm ou manifestos deste repositório.
4. `srectl status` lê a visão do Argo CD — a mesma da UI.

## Request & telemetry flow

{{< mermaid >}}
sequenceDiagram
  participant U as User / srectl
  participant N as ingress-nginx
  participant A as api-a
  participant B as api-b
  participant T as Tempo
  participant P as Prometheus
  participant L as Loki
  U->>N: GET /orders (TLS)
  N->>A: route
  A->>B: POST /pay (trace context)
  B-->>A: 200 approved
  A-->>U: 200 {order confirmed}
  A-->>T: OTLP spans
  B-->>T: OTLP spans
  A-->>P: /metrics scrape
  A-->>L: stdout logs via Alloy
{{< /mermaid >}}

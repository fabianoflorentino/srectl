# Architecture

High-level design of **srectl**: how the pieces fit together on a local kind cluster.

## Big picture

```mermaid
flowchart TB
    subgraph Host["Developer machine"]
        CLI["srectl CLI (Go task runner)"]
        KIND["kind cluster\n1 control-plane + 2 workers"]
        CLI -->|"provision + bootstrap + operate"| KIND
    end

    subgraph KIND
        direction TB

        subgraph GitOps["GitOps"]
            ARGO["Argo CD"]
        end

        subgraph IngressTLS["Edge"]
            NGINX["ingress-nginx :80/:443"]
            CERT["cert-manager (selfsigned)"]
        end

        subgraph Obs["Observability (LGTM)"]
            PROM["Prometheus"]
            AM["Alertmanager"]
            GRAF["Grafana"]
            LOKI["Loki"]
            ALLOY["Alloy"]
            TEMPO["Tempo"]
            SLO["slo-generator\n(CronJob)"]
            PUSH["Pushgateway"]
        end

        subgraph Sec["Security"]
            TRIVY["Trivy Operator"]
            FALCO["Falco"]
            KYV["Kyverno"]
        end

        subgraph Apps["Demo workloads"]
            APIA["api-a (orders)"]
            APIB["api-b (payments)"]
            FRAUD["fraudctl\n(lb + 2x api)"]
            CV["certificate-validate\n(k8s monitor)"]
            STRACE["stracectl\n(sidecar)"]
        end

        subgraph Chaos["Chaos"]
            CM["chaos-mesh"]
        end
    end

    ARGO -->|"sync deploy/"| Obs & Sec & Apps & IngressTLS & Chaos
    NGINX --> APIA & APIB & CV & GRAF
    CERT --> NGINX
    APIA -->|"tcp"| APIB
    APIA -->|"OTLP :4318"| TEMPO
    APIB -->|"OTLP :4318"| TEMPO
    ALLOY --> LOKI
    FALCO --> LOKI & AM
    PROM --> AM & GRAF
    SLO -->|"PromQL"| PROM
    SLO -->|"burn rate"| PUSH
    PUSH -->|"scrape"| PROM
    CM -.->|"inject faults"| APIA & FRAUD
```

## Control plane vs. data plane

- **Control plane (the `srectl` CLI):** a simple task runner that provisions kind,
  bootstraps Argo CD, and offers operational commands (`status`, `sync`, `open`,
  `demo`). It shells out to `kind`/`kubectl`/`helm`. It is the *only* imperative
  entry point.
- **Data plane (Argo CD):** everything inside the cluster is a GitOps resource under
  `deploy/`. Argo CD converges the cluster to the desired state.

```
srectl up ──► kind cluster ──► argo cd (bootstrap) ──► deploy/ (GitOps) ──► everything else
```

## Repository layout

```text
srectl/
├── cmd/srectl/                 # CLI entrypoint
├── internal/
│   ├── cli/                 # cobra commands
│   ├── config/              # srectl.yaml + env overrides
│   ├── kind/                # cluster lifecycle (wraps kind CLI)
│   ├── k8s/                 # kubectl/helm/port-forward wrappers
│   ├── argocd/              # bootstrap + app-of-apps
│   ├── demo/                # load gen + chaos + verify scripts
│   └── verify/              # smoke tests
├── apps/
│   ├── api-a/               # orders service (instrumented)
│   └── api-b/               # payments service (instrumented)
├── deploy/                  # GitOps source of truth (watched by Argo CD)
│   ├── bootstrap/           # argo cd install + root app
│   ├── ingress/             # ingress-nginx + cert-manager
│   ├── observability/       # LGTM stack
│   ├── security/            # trivy + falco + kyverno + netpol
│   ├── apps/                # api-a, api-b
│   └── projects/            # fraudctl, certificate-validate, stracectl
├── site/                    # Hugo bilingual site (GitHub Pages)
├── docs/
│   ├── en/                  # English docs
│   └── pt-br/               # Portuguese docs
├── config/srectl.yaml          # CLI defaults (cluster name, versions, toggles)
├── .github/workflows/ci.yml # CI (lint/test/build/e2e)
├── Makefile
└── go.mod
```

## App-of-apps pattern

Argo CD is installed imperatively once. A single **root `Application`** points at
`deploy/bootstrap/` and recursively creates child `Application`s, one per component
group:

```mermaid
flowchart LR
    ROOT["root-app\n(deploy/bootstrap)"]
    ROOT --> O["observability"]
    ROOT --> S["security"]
    ROOT --> A["apps"]
    ROOT --> P["projects"]
    ROOT --> I["ingress"]
    ROOT --> C["chaos"]
    each --> resources["Helm charts + raw manifests"]
```

## How the four services talk

```mermaid
sequenceDiagram
    participant U as User / srectl CLI
    participant NGX as ingress-nginx
    participant A as api-a
    participant B as api-b
    participant T as Tempo

    U->>NGX: GET /orders
    NGX->>A: route
    A->>B: POST /payments/charge
    B-->>A: 200
    A-->>U: 200
    A-->>T: span (api-a, OTLP)
    B-->>T: span (api-b, OTLP)
```

Both APIs emit:

- **Metrics** → Prometheus (via `/metrics` + ServiceMonitor).
- **Logs** → Alloy → Loki (stdout, `slog` JSON).
- **Traces** → Tempo (OTLP :4318, exported directly by the apps).

## Observability data flow

```mermaid
flowchart LR
    subgraph Sources
        APIA["api-a"]
        APIB["api-b"]
        FRAUD["fraudctl"]
        CV["certificate-validate"]
        STRACE["stracectl"]
        KUBE["kube-state-metrics\nnode-exporter"]
        SLO["slo-generator"]
        PUSH["Pushgateway"]
    end

    APIA -->|metrics| PROM
    APIB -->|metrics| PROM
    CV -->|metrics :9102| PROM
    STRACE -->|metrics :8080| PROM
    KUBE -->|metrics| PROM

    APIA -->|logs| ALLOY
    APIB -->|logs| ALLOY
    FRAUD -->|logs| ALLOY
    FALCO -->|events| ALLOY
    ALLOY --> LOKI

    APIA -->|spans| OTEL
    APIB -->|spans| OTEL
    OTEL --> TEMPO

    PROM --> GRAF
    LOKI --> GRAF
    TEMPO --> GRAF
    PROM --> AM
    FALCO --> AM

    SLO -->|PromQL| PROM
    SLO -->|metrics| PUSH
    PUSH -->|scrape| PROM
```

## Security data flow

```mermaid
flowchart LR
    TRIVY -->|vuln reports + metrics| GRAF
    FALCO -->|events| AM
    FALCO -->|events| LOKI
    KYV -->|policy reports| GRAF
    CERT -->|issues TLS| NGX
    CV -->|cert metrics + alerts| PROM & AM
```

## Kind networking

ingress-nginx uses kind `extraPortMappings` so `localhost:80/443` on the host map into
the cluster:

```yaml
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 80
        hostPort: 80
      - containerPort: 443
        hostPort: 443
```

Services are exposed as `*.localhost` hosts (cert-manager selfsigned covers them),
making the whole stack reachable over TLS from the host with no cloud load balancer.

## Failure and recovery model

- **Self-healing:** Argo CD prunes drift; Deployments/ReplicaSets restart crashed pods.
- **Detection:** Prometheus rules + Falco events → Alertmanager.
- **Injection:** chaos-mesh faults let us practice detection and recovery deliberately.
- **Verification:** `srectl demo verify` asserts the cluster returns to healthy.

This is the environment the SLOs and runbooks in [`OBSERVABILITY.md`](OBSERVABILITY.md)
operate against.

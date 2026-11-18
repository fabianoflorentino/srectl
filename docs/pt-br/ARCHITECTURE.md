# Arquitetura

Design de alto nível do **srectl**: como as peças se encaixam num cluster kind local.

## Visão geral

```mermaid
flowchart TB
    subgraph Host["Máquina do desenvolvedor"]
        CLI["CLI srectl (task runner em Go)"]
        KIND["cluster kind\n1 control-plane + 2 workers"]
        CLI -->|"provisiona + bootstrap + opera"| KIND
    end

    subgraph KIND
        direction TB

        subgraph GitOps["GitOps"]
            ARGO["Argo CD"]
        end

        subgraph IngressTLS["Borda"]
            NGINX["ingress-nginx :80/:443"]
            CERT["cert-manager (selfsigned)"]
        end

        subgraph Obs["Observabilidade (LGTM)"]
            PROM["Prometheus"]
            AM["Alertmanager"]
            GRAF["Grafana"]
            LOKI["Loki"]
            ALLOY["Alloy"]
            TEMPO["Tempo"]
        end

        subgraph Sec["Segurança"]
            TRIVY["Trivy Operator"]
            FALCO["Falco"]
            KYV["Kyverno"]
        end

        subgraph Apps["Workloads demo"]
            APIA["api-a (pedidos)"]
            APIB["api-b (pagamentos)"]
            FRAUD["fraudctl\n(lb + 2x api)"]
            CV["certificate-validate\n(k8s monitor)"]
            STRACE["stracectl\n(sidecar)"]
        end

        subgraph Chaos["Caos"]
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
    CM -.->|"injeta falhas"| APIA & FRAUD
```

## Plano de controle vs. plano de dados

- **Plano de controle (CLI `srectl`):** um task runner simples que provisiona o kind,
  faz o bootstrap do Argo CD e oferece comandos operacionais (`status`, `sync`, `open`,
  `demo`). Ele chama `kind`/`kubectl`/`helm`. É o *único* ponto de entrada imperativo.
- **Plano de dados (Argo CD):** tudo dentro do cluster é um recurso GitOps sob `deploy/`.
  O Argo CD converge o cluster para o estado desejado.

```
srectl up ──► cluster kind ──► argo cd (bootstrap) ──► deploy/ (GitOps) ──► todo o resto
```

## Layout do repositório

```text
srectl/
├── cmd/srectl/                 # entrypoint da CLI
├── internal/
│   ├── cli/                 # comandos cobra
│   ├── config/              # srectl.yaml + overrides por env
│   ├── kind/                # ciclo de vida do cluster (envolve o kind CLI)
│   ├── k8s/                 # wrappers de kubectl/helm/port-forward
│   ├── argocd/              # bootstrap + app-of-apps
│   ├── demo/                # scripts de carga + caos + verificação
│   └── verify/              # smoke tests
├── apps/
│   ├── api-a/               # serviço de pedidos (instrumentado)
│   └── api-b/               # serviço de pagamentos (instrumentado)
├── deploy/                  # fonte de verdade GitOps (observada pelo Argo CD)
│   ├── bootstrap/           # instalação do argo cd + root app
│   ├── ingress/             # ingress-nginx + cert-manager
│   ├── observability/       # stack LGTM
│   ├── security/            # trivy + falco + kyverno + netpol
│   ├── apps/                # api-a, api-b
│   └── projects/            # fraudctl, certificate-validate, stracectl
├── site/                    # site Hugo bilíngue (GitHub Pages)
├── docs/
│   ├── en/                  # docs em inglês
│   └── pt-br/               # docs em português
├── config/srectl.yaml          # defaults da CLI (nome do cluster, versões, toggles)
├── .github/workflows/ci.yml # CI (lint/test/build/e2e)
├── Makefile
└── go.mod
```

## Padrão app-of-apps

O Argo CD é instalado imperativamente uma única vez. Um único **root `Application`**
aponta para `deploy/bootstrap/` e cria recursivamente `Application`s filhos, um por
grupo de componentes:

```mermaid
flowchart LR
    ROOT["root-app\n(deploy/bootstrap)"]
    ROOT --> O["observability"]
    ROOT --> S["security"]
    ROOT --> A["apps"]
    ROOT --> P["projects"]
    ROOT --> I["ingress"]
    ROOT --> C["chaos"]
    each --> resources["Helm charts + manifests brutos"]
```

## Como os quatro serviços conversam

```mermaid
sequenceDiagram
    participant U as Usuário / CLI srectl
    participant NGX as ingress-nginx
    participant A as api-a
    participant B as api-b
    participant T as Tempo

    U->>NGX: GET /orders
    NGX->>A: rota
    A->>B: POST /payments/charge
    B-->>A: 200
    A-->>U: 200
    A-->>O: span (api-a)
    B-->>O: span (api-b)
    O-->>T: lote de spans
```

As duas APIs emitem:

- **Métricas** → Prometheus (via `/metrics` + ServiceMonitor).
- **Logs** → Alloy → Loki (stdout, `slog` JSON).
- **Traces** → Tempo (OTLP :4318, exported directly by the apps).

## Fluxo de dados de observabilidade

```mermaid
flowchart LR
    subgraph Sources
        APIA["api-a"]
        APIB["api-b"]
        FRAUD["fraudctl"]
        CV["certificate-validate"]
        STRACE["stracectl"]
        KUBE["kube-state-metrics\nnode-exporter"]
    end

    APIA -->|métricas| PROM
    APIB -->|métricas| PROM
    CV -->|métricas :9102| PROM
    STRACE -->|métricas :8080| PROM
    KUBE -->|métricas| PROM

    APIA -->|logs| ALLOY
    APIB -->|logs| ALLOY
    FRAUD -->|logs| ALLOY
    FALCO -->|eventos| ALLOY
    ALLOY --> LOKI

    APIA -->|spans| OTEL
    APIB -->|spans| OTEL
    OTEL --> TEMPO

    PROM --> GRAF
    LOKI --> GRAF
    TEMPO --> GRAF
    PROM --> AM
    FALCO --> AM
```

## Fluxo de dados de segurança

```mermaid
flowchart LR
    TRIVY -->|reports de vuln + métricas| GRAF
    FALCO -->|eventos| AM
    FALCO -->|eventos| LOKI
    KYV -->|reports de política| GRAF
    CERT -->|emite TLS| NGX
    CV -->|métricas + alertas de cert| PROM & AM
```

## Rede no kind

O ingress-nginx usa `extraPortMappings` do kind para que `localhost:80/443` no host
mapeiem para dentro do cluster:

```yaml
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 80
        hostPort: 80
      - containerPort: 443
        hostPort: 443
```

Os serviços são expostos como hosts `*.localhost` (o cert-manager selfsigned os cobre),
tornando toda a stack acessível via TLS do host, sem load balancer de cloud.

## Modelo de falha e recuperação

- **Self-healing:** o Argo CD faz prune de drift; Deployments/ReplicaSets reiniciam pods
  que caíram.
- **Detecção:** regras do Prometheus + eventos do Falco → Alertmanager.
- **Injeção:** as falhas do chaos-mesh nos deixam praticar detecção e recuperação de
  forma deliberada.
- **Verificação:** `srectl demo verify` valida que o cluster voltou ao saudável.

É neste ambiente que operam os SLOs e runbooks do [`OBSERVABILITY.md`](OBSERVABILITY.md).

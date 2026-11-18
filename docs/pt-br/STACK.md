# Catálogo de Componentes

Cada componente da stack, por que foi escolhido e como é instalado.

Legenda — método de instalação:

- **GitOps**: aplicado declarativamente pelo Argo CD a partir de `deploy/`.
- **Bootstrap**: instalado imperativamente uma vez pelo `srectl bootstrap` (só o Argo CD).
- **CLI**: gerenciado pela CLI `srectl` (o próprio kind).

## Plataforma

| Componente | Propósito | Instalação | Por quê |
| --- | --- | --- | --- |
| kind | Kubernetes local | CLI | Leve, multi-nó, sem cloud |
| ingress-nginx | Roteamento de borda + terminação TLS | GitOps | Controlador de ingress mais comum |
| cert-manager | Emissão/rotação automatizada de TLS | GitOps | Alimenta certificate-validate; padrão |
| Argo CD | Motor de deploy GitOps | Bootstrap | Fonte de verdade declarativa única |
| CLI `srectl` | Task runner (estilo make) | — | Envolve kind/kubectl/helm com boa UX |

## Observabilidade (LGTM)

| Componente | Propósito | Instalação | Por quê |
| --- | --- | --- | --- |
| kube-prometheus-stack | Prometheus, Alertmanager, Grafana, node-exporter, kube-state-metrics, CRDs do Operator | GitOps (Helm) | Padrão de facto, tudo incluído |
| Prometheus Operator | CRDs ServiceMonitor/PodMonitor/Alert | GitOps | Modelo de scrape padrão |
| Loki | Agregação de logs | GitOps (Helm) | Nativo do Grafana, simples |
| Alloy | Agente de logs + pipeline OTel | GitOps (DaemonSet) | Um agente, substitui Promtail/collector-side |
| Tempo | Backend de tracing distribuído | GitOps (Helm) | Traces nativos do Grafana |
| SDK OpenTelemetry (nas apps) | Traces de api-a/api-b via OTLP | embutido nas apps | sem serviço extra |
| Grafana | Dashboards + UI unificada | GitOps (Helm) | Painel único |

## Segurança

| Componente | Propósito | Instalação | Por quê |
| --- | --- | --- | --- |
| Trivy Operator | Scan de imagens/CIS + reports | GitOps (Helm) | Contínuo, baseado em CRD |
| Falco | Detecção de ameaças em runtime (eBPF) | GitOps (Helm) | Runtime security padrão CNCF |
| Kyverno | Motor de política de admissão | GitOps (Helm) | Políticas nativas do Kubernetes |
| NetworkPolicies | Controle de tráfego leste-oeste | GitOps (bruto) | Deny-by-default no ns demo |

## Caos

| Componente | Propósito | Instalação | Por quê |
| --- | --- | --- | --- |
| chaos-mesh | Injeção de falha (rede/pod/stress) | GitOps (Helm) | CRDs ricos, projeto CNCF ativo |

## Workloads

| Componente | Propósito | Instalação | Fonte |
| --- | --- | --- | --- |
| api-a | Serviço demo de pedidos (instrumentado) | GitOps | este repo |
| api-b | Serviço demo de pagamentos (instrumentado) | GitOps | este repo |
| fraudctl | API de detecção de fraude (LB + 2 réplicas) | GitOps | `github.com/fabianoflorentino/fraudctl` |
| certificate-validate | `k8s monitor` de certificados + dashboard | GitOps | `github.com/fabianoflorentino/certificate-validate` |
| stracectl | Sidecar de tracing de syscall | GitOps | `github.com/fabianoflorentino/stracectl` |

## Docs & site

| Componente | Propósito | Instalação | Por quê |
| --- | --- | --- | --- |
| Site Hugo (`site/`) | Site de estudo bilíngue | GitHub Pages (CI) | Igual aos nossos outros projetos; ensina SRE/DevOps/Platform Eng |

## Nossos projetos — como se encaixam

### fraudctl

- Implantado como **load balancer + 2 réplicas de API** (espelha a topologia do compose).
- Expõe `POST /fraud-score`; o LB distribui conexões aos workers via `SCM_RIGHTS`.
- Não tem endpoint Prometheus nativo, então é observado como **caixa-preta**:
  - logs estruturados → Alloy → Loki;
  - `pprof` para drills de profiling;
  - um sidecar exporter Prometheus (ou o middleware de telemetria) para métricas RED.
- Usado como o workload *de alta performance e opaco* para drills de carga e caos.

### certificate-validate

- Roda `k8s monitor` como DaemonSet varrendo Secrets e Ingresses TLS no cluster todo.
- Expõe `certificate_days_left`, `certificate_expired`, `certificate_revoked` em `:9102`.
- ServiceMonitor → Prometheus; `PrometheusRule` alerta em `days_left < 15`.
- Alertas webhook → um receiver in-cluster (ou Alertmanager).
- Dashboard servido atrás do ingress com certificado TLS do cert-manager — provando que
  o certificado que ele valida é, ele próprio, gerenciado pelo cert-manager.

### stracectl

- Injetado como **sidecar** na api-a (ou num pod demo dedicado) com
  `attach --serve :8080`, `shareProcessNamespace: true`, `CAP_SYS_PTRACE`.
- Expõe stats de latência por syscall + métricas Prometheus → dashboard.
- `srectl trace` embrulha `stracectl discover` / `attach` para debug interativo.
- Demonstra *observabilidade profunda*: do nível de syscall até os traces.

### seraph

- Um **status board estático** (site Hugo + snapshot `state.json`) servido por nginx.
- Empacotado como container, pinado como `fabianoflorentino/seraph:v0.1.1`; um
  `seraph sync` em background atualiza o snapshot de status oficiais a cada 5 minutos.
- Atrás do ingress (`seraph.localhost`) → `srectl open seraph`; observado pelas métricas
  RED do ingress-nginx e pelo Loki (sem `/metrics` nativo).

## Por que não ...

- **OPA/Gatekeeper em vez de Kyverno:** Kyverno é nativo do Kubernetes e mais simples de
  escrever para um ambiente de estudo.
- **Fluent Bit em vez de Alloy:** Alloy consolida o agente de log + OTel em um só,
  reduzindo peças móveis.
- **Elastic Stack em vez de LGTM:** o Grafana LGTM é mais leve e junta métricas/logs/traces
  numa UI só, sem um cluster separado de indexação.
- **Terraform/Crossplane:** não há infraestrutura externa para provisionar no kind.

## Pinning de versões

Todos os Helm charts são pinados em `deploy/*/values.yaml` (via `Chart.yaml`/`Chart.lock`
ou `targetRevision` explícito). As imagens são pinadas por digest/tag (sem `latest`) para
que o ambiente seja reproduzível e a política `disallow-latest-tag` do Kyverno passe.

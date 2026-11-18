# Plano de Implementação

Este documento é a fonte de verdade para construir o **srectl**. Ele divide o
trabalho em fases ordenadas, cada uma com tarefas concretas e critérios de aceite. Nada
neste plano é código ainda — é o roteiro que executaremos.

## Objetivos

Construir um ambiente de treinamento **SRE / DevOps / Platform Engineering** local e
reprodutível que:

1. Roda inteiramente em **kind** (Kubernetes-in-Docker) — sem cloud, sem dependências
   externas.
2. É dirigido por uma CLI Go simples, **`srectl`**, um *task runner como o Make* (veja
   abaixo).
3. Cobre toda a superfície da plataforma: **GitOps, observabilidade (LGTM), segurança,
   CI/CD, engenharia de caos**.
4. Entrega **duas APIs Go de demonstração** (instrumentadas) para exercitar a stack de
   ponta a ponta.
5. Integra três projetos nossos — **fraudctl**, **stracectl** e
   **certificate-validate** — como workloads de primeira classe.
6. Publica um site em **GitHub Pages** (Hugo, bilíngue EN/PT-BR) para explicar e
   ensinar SRE / DevOps / Platform Engineering por meio do projeto.

## Não-objetivos (por enquanto)

- Hardening de produção, HA, multi-cluster, multi-cloud.
- DNS externo / certificados públicos (usamos cert-manager self-signed).
- Criar novos produtos dentro deste repo (consumimos fraudctl/stracectl/certificate-validate).

## A CLI `srectl` é um task runner

O `srectl` é deliberadamente **simples** — o equivalente a um `Makefile` escrito em Go:

- Cada subcomando envolve uma ferramenta conhecida (`kind`, `kubectl`, `helm`) mais um
  pouco de cola em Go para boa UX (tabela de status, port-forward, progresso).
- Sem `client-go` pesado no caminho principal; sempre que possível, ele chama as
  ferramentas externas.
- A configuração fica em `config/srectl.yaml` (como um `.env`), com override por variáveis
  `SRECTL_*`.
- **Não** é um framework, nem operator, nem controller.

Pense nele como `make up`, `make status`, `make demo`, … mas descobrível via
`srectl --help`, com status colorido e defaults sensatos.

## Decisões de arquitetura (travadas)

| Decisão | Escolha | Justificativa |
| --- | --- | --- |
| Cluster | kind, 3 nós (1 control-plane + 2 workers) | Scheduling real, afinidades, tolerations |
| Modelo de deploy | **Argo CD (GitOps)** | `deploy/` é a fonte de verdade; padrão app-of-apps |
| CLI | Task runner em Go (estilo make) | Simples, descobrível, envolve kind/kubectl/helm |
| Ingress | ingress-nginx (port-mapping 80/443 no kind) | Mais comum; integração com cert-manager |
| TLS | cert-manager + ClusterIssuer selfsigned | Certificados automatizados, alimenta certificate-validate |
| Métricas | kube-prometheus-stack (Prometheus + Grafana + Alertmanager) | Padrão de facto |
| Logs | Loki + Alloy | Um único agente para logs + pipeline OTel |
| Tracing | Tempo + OpenTelemetry Collector | Backend de tracing padrão |
| Segurança | Trivy (Operator) + Falco + Kyverno | Scan, runtime, política |
| Caos | chaos-mesh | Injeção de latência/falha para drills de SRE |
| Site | Hugo, bilíngue EN/PT-BR, GitHub Pages | Mesmo tooling dos nossos outros projetos |
| Linguagem | Go 1.27 | Igual ao resto da nossa stack |

Veja o catálogo completo em [`STACK.md`](STACK.md) e diagramas em
[`ARCHITECTURE.md`](ARCHITECTURE.md).

---

## Fase 0 — Repositório & esqueleto da CLI

**Objetivo:** uma CLI Go compilável (task runner) com a superfície de comandos ligada,
e a estrutura de repo que a camada GitOps vai consumir.

Tarefas:

- [x] Inicializar o módulo Go (`github.com/fabianoflorentino/srectl`).
- [x] Montar a CLI Cobra em `cmd/srectl/` com a árvore de comandos completa (stubs).
- [x] Implementar carregamento de configuração (`config/srectl.yaml` + overrides por env,
      prefixo `SRECTL_`).
- [x] `internal/kind/` — create/delete/status do cluster kind (arquivo de config, nº de nós).
- [x] `internal/k8s/` — wrappers finos que chamam `kubectl`/`helm` + port-forward.
- [x] `internal/demo/` — scripts de carga, caos e verificação.
- [x] `Makefile` (build, test, lint, fmt, docker-build) e `lefthook.yml`.
- [x] Esqueleto do `.github/workflows/ci.yml` (lint + test + build).
- [x] `README.md` (bilíngue), scaffolding de `docs/` (en + pt-br).

**Superfície de comandos (alvo):**

```text
srectl up            cria cluster + bootstrap + aguarda saudável
srectl down          desliga tudo (cluster + kind)
srectl status        matriz de saúde dos componentes
srectl cluster       create/delete/status do cluster kind
srectl bootstrap     instala Argo CD + registra o root app
srectl sync          força o sync do Argo CD (padrão: todos)
srectl apps          deploy/restart/scale das apps demo
srectl demo          roda um cenário de demo (load, chaos, break, verify)
srectl trace         roda o stracectl contra um pod
srectl certs         dispara o scan do certificate-validate
srectl open          port-forward de um serviço (grafana, argocd, api-a, ...)
srectl logs          acompanha os logs de um workload
srectl version       info de build
```

**Critérios de aceite:**

- `srectl version` e `srectl --help` funcionam.
- `go build ./...`, `go test ./...`, `golangci-lint run` passam no CI.

---

## Fase 1 — Cluster, ingress & TLS

**Objetivo:** um cluster kind com ingress e certificados automatizados.

Tarefas:

- [x] `deploy/kind/kind-config.yaml` — 3 nós, ingress-nginx `extraPortMappings` (80/443).
- [x] ingress-nginx implantado (bootstrap imperativo; Argo CD a partir da Fase 2).
- [x] cert-manager + `ClusterIssuer` selfsigned + `ConfigMap` de bundle de confiança.
- [x] Uma app `hello` provando o round trip: ingress → TLS → pod.
- [x] `srectl up` dirige toda esta fase de forma idempotente.

**Critérios de aceite:**

- `srectl up` cria o cluster e um pod é acessível via `https://*.localhost`.
- `srectl status` reporta cluster, ingress e cert-manager como `Ready`.

---

## Fase 2 — Fundação GitOps (Argo CD)

**Objetivo:** tudo após esta fase é implantado declarativamente.

Tarefas:

- [x] Instalar o Argo CD (bootstrap imperativo — este é o único passo imperativo).
- [x] Definir o **root app** (`deploy/bootstrap/argocd/`) usando app-of-apps.
- [x] Definir CRs `ApplicationSet`/`Application` para cada grupo de componentes:
      `observability`, `security`, `apps`, `projects`, `ingress`, `chaos`.
- [x] `srectl sync` + `srectl status` integram com a API do Argo CD.
- [x] Rotação de credenciais padrão + `srectl open argocd`.

**Critérios de aceite:**

- Editar `deploy/` faz o Argo CD convergir o cluster (self-healing, prune).
- O App `argocd` mostra `Healthy + Synced` para o root app.

---

## Fase 3 — Observabilidade (LGTM)

**Objetivo:** métricas, logs, traces e alertas, unificados no Grafana.

Tarefas:

- [x] **kube-prometheus-stack**: Prometheus, Alertmanager, Grafana, node-exporter,
      kube-state-metrics, CRDs do Prometheus Operator (ServiceMonitor/PodMonitor).
- [x] **Loki** (single-binary) + **Alloy** DaemonSet (tail dos logs de pods, labels
      namespace/pod/container/app/node, `level` extraído).
- [x] **Tempo** (single-binary) — **OpenTelemetry Collector** entra com as apps demo.
- [x] Datasources do Grafana (Prometheus, Loki, Tempo) provisionados via ConfigMap.
- [ ] Dashboards baseline: visão do cluster, node/pod, RED das apps, syscall (stracectl),
      certificados (certificate-validate). (existem os dashboards `srectl · Logs` e `srectl · Golden Signals`)
- [ ] Regras de alerta baseline: alta taxa de erro, p99 alto, disco/cpu, expiração de
      certificado, pressão de nó (um `TestAlert` já prova o pipeline de alerta).
- [x] `srectl open grafana` + dashboards padrão encontráveis.

**Critérios de aceite:**

- `srectl up` deixa o Grafana com os três datasources e dashboards prontos.
- Os alvos de scrape das apps demo e dos nossos projetos aparecem como `UP` no Prometheus.
- Um erro gerado numa app demo dispara um alerta de teste no Alertmanager.

---

## Fase 4 — Segurança

**Objetivo:** scan, detecção em runtime e enforcement de política.

Tarefas:

- [x] **Trivy Operator**: reports de vulnerabilidade/config-audit/RBAC (namespace demo),
      métricas via ServiceMonitor para o Grafana.
- [x] **Falco**: detecção em runtime com probe modern eBPF; Falcosidekick → Loki.
- [x] **Kyverno**: políticas baseline como CEL ValidatingPolicy (proibir tag `latest`,
      exigir requests/limits, exigir runAsNonRoot (Audit), bloquear hostPath).
- [x] NetworkPolicies para o namespace demo (default-deny + matriz de permissão, atuando via kindnet).
- [x] cert-manager (da Fase 1) documentado como pilar de TLS.
- [x] Dashboard: visão de vulnerabilidade + política no Grafana (`srectl · Security`).

**Critérios de aceite:**

- `srectl up` deixa os reports do Trivy populados e as políticas do Kyverno atuando.
- Um workload não-conforme (ex.: tag `latest`) é rejeitado pelo Kyverno.
- O Falco produz pelo menos um evento de regra de demo (`srectl demo security`).

---

## Fase 5 — Aplicações de demonstração

**Objetivo:** o workload que torna a plataforma observável e demonstrável.

Tarefas:

- [x] **api-a** (`apps/api-a/`): serviço de pedidos/checkout.
      - `net/http` + roteamento Go 1.22; `/health`, `/ready`, `/orders`, `/slow`, `/error`.
      - SDK OTel (OTLP → Tempo), métricas Prometheus (RED), logs `slog`.
      - Chama a api-b (demonstra tracing distribuído).
- [x] **api-b** (`apps/api-b/`): serviço de pagamentos, mesma instrumentação, chamada pela api-a.
- [x] Dockerfiles + manifests em `deploy/apps/` (Deployment/Service/Ingress/ServiceMonitor).
- [x] Gerador de carga (`srectl demo load`) — um pequeno hammer em Go — para gerar sinal.

**Critérios de aceite:**

- O trace `api-a → api-b` é consultável no Tempo/Grafana.
- As métricas RED das duas APIs aparecem num dashboard compartilhado.
- Os logs das duas APIs são buscáveis no Loki.

---

## Fase 6 — Integrar nossos projetos

**Objetivo:** rodar fraudctl, stracectl e certificate-validate como workloads reais.

Tarefas:

- [x] **fraudctl** (`deploy/projects/fraudctl/`): LB + 2 workers de API num único Pod
      (SCM_RIGHTS via emptyDir compartilhado), `POST /fraud-score` via ingress; throughput
      /latência via métricas RED do ingress-nginx, logs no Loki.
- [x] **certificate-validate** (`deploy/projects/certificate-validate/`):
      DaemonSet `k8s monitor` varrendo Secrets/Ingresses TLS; `--metrics-addr=:9102`;
      ServiceMonitor → Prometheus; webhook → Alertmanager/webhook-receiver; dashboard
      atrás do ingress com TLS do cert-manager.
- [x] **stracectl** (`deploy/projects/stracectl/`): pod demo dedicado + sidecar
      (`attach --serve :8080`); métricas de syscall → Prometheus; dashboard de syscall;
      `srectl trace` faz port-forward da visão ao vivo.
- [x] Regras de alerta: `certificate_days_left < 15` (CertificateExpiringSoon/Expired);
      contadores de anomalia de syscall adiados.
- [ ] `srectl demo projects` — script ponta a ponta que toca os três.

**Critérios de aceite:**

- As métricas `certificate_days_left` aparecem no Prometheus com labels k8s.
- `srectl trace` abre uma visão ao vivo do stracectl na api-a.
- `srectl demo load` mostra throughput/latência do fraudctl no seu dashboard.

---

## Fase 7 — Drills de SRE (caos + SLOs)

**Objetivo:** transformar observabilidade em ação.

Tarefas:

- [ ] Instalar **chaos-mesh** + CRDs (NetworkChaos, PodChaos, StressChaos, HTTPChaos).
- [ ] Cenários `srectl demo chaos`: injeção de latência na api-a, kill de pod no fraudctl,
      falha de DNS, perda de pacotes.
- [ ] Definições de SLO/SLI em `docs/*/OBSERVABILITY.md` + recording rules no Prometheus
      (objetivos de disponibilidade e latência).
- [ ] Runbooks (`docs/*/runbooks/`) para os principais alertas.
- [ ] `srectl demo verify` — valida que a stack se recupera e os alertas disparam.

**Critérios de aceite:**

- Injetar latência na api-a eleva o p99 no dashboard e dispara um alerta.
- Matar os pods do fraudctl se recupera dentro do SLO definido (self-healing).

---

## Fase 8 — Site de documentação (GitHub Pages)

**Objetivo:** um site bilíngue que ensina SRE / DevOps / Platform Engineering por meio
deste projeto.

Tarefas:

- [ ] Projeto Hugo em `site/`, bilíngue (**EN** padrão + **PT-BR** em `/pt-br/`),
      igual aos nossos outros projetos.
- [ ] Seções de conteúdo: visão geral, arquitetura, a stack e trilhas de aprendizado por
      tema (**SRE**, **DevOps**, **Platform Engineering**).
- [ ] Tutoriais práticos dirigidos pelos cenários de `srectl demo` (observabilidade,
      segurança, caos).
- [ ] Runbooks + páginas de SLO gerados a partir dos docs.
- [ ] Diagramas (Mermaid) renderizados no site.
- [ ] Deploy no GitHub Pages via GitHub Actions (`site/` → Pages).

**Critérios de aceite:**

- `make site-dev` serve o site bilíngue localmente.
- Toda página está disponível em EN e PT-BR.
- O GitHub Actions publica o site no merge para `main`.

---

## Fase 9 — CI/CD & polimento

**Objetivo:** um pipeline de CI que valida tudo de forma reprodutível.

Tarefas:

- [ ] GitHub Actions: lint, test, build da CLI + das duas APIs.
- [ ] Build & publicação de imagens (api-a, api-b e mirrors dos três projetos).
- [ ] Job E2E: `srectl up` num cluster kind de CI → `srectl demo verify` → `srectl down`.
- [ ] Versionamento de imagens para pinar no Argo CD (sem `latest` no GitOps).
- [ ] Finalizar docs: dashboards, runbooks, guia de contribuição (bilíngue).

**Critérios de aceite:**

- Um run de CI verde em cada PR.
- `srectl up && srectl demo verify && srectl down` completa sem intervenção no CI.
- O site builda e publica no GitHub Pages automaticamente.

---

## Definição de pronto (por fase)

- [ ] Código mesclado (inglês, commits convencionais).
- [ ] Testes + lint verdes.
- [ ] Documentado em `docs/en/*.md` **e** `docs/pt-br/*.md`.
- [ ] Reproduzível via um único comando `srectl` ou alvo `make`.

## Justificativa do sequenciamento

- As Fases 0–2 são o esqueleto da plataforma; nada mais é demonstrável sem elas.
- A Fase 3 (observabilidade) entra **antes** das apps para que elas nasçam observáveis.
- A Fase 5 (apps) entra antes da Fase 6 (projetos) para que os projetos tenham um
  cluster instrumentado para observar.
- A Fase 7 (caos) é a última porque assume um cluster saudável, observável e seguro para
  quebrar.
- A Fase 8 (site) roda em paralelo com as Fases 3–7 assim que a plataforma existir.

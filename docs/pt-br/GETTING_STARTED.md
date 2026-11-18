# Começando

> **Status: em andamento.** As Fases 0–2 estão implementadas: o esqueleto da CLI,
> o cluster kind com ingress/TLS e a fundação GitOps (app-of-apps do Argo CD). O
> roteiro está em [`PLAN.md`](PLAN.md).

## Pré-requisitos

- Go 1.27+
- Docker (rodando)
- kind, kubectl, helm no `PATH`
- `srectl` buildado: `make build` (ou `go install ./cmd/srectl`)
- O repo publicado num git que o Argo CD consiga clonar (veja `argocd.repo` abaixo)

## Um comando para subir

```bash
srectl up          # cluster kind + Argo CD + tudo (via GitOps)
srectl status      # matriz de saúde (cluster + apps)
srectl open argocd # port-forward da UI do Argo CD
```

## O que funciona hoje (Fases 0–2)

```text
srectl up            cria cluster + bootstrap do Argo CD (idempotente)
srectl down          derruba o cluster
srectl status        matriz de saúde: cluster + cada app do Argo CD
srectl cluster       create/delete/status do cluster kind
srectl bootstrap     instala o Argo CD + registra o root app
srectl sync [app]    força o Argo CD a refrescar (padrão: todos)
srectl creds         mostra as credenciais de admin do Argo CD + Grafana
srectl creds rotate  rotaciona a senha de um componente (`creds rotate argocd`)
srectl open          port-forward de um serviço (argocd, grafana, prometheus, alertmanager, hello)
srectl version       info de build
```

`srectl open` bloqueia em primeiro plano (Ctrl+C para parar); use `--daemon` para rodar
em segundo plano e `--down` para parar (`srectl open argocd --daemon` → `srectl open argocd --down`).

`srectl creds` imprime os logins de Argo CD e Grafana; Loki e Tempo rodam sem
autenticação nesta stack local.

Comandos das fases seguintes (`demo`, `apps`, `certs`, `logs`) existem como
stubs e reportam "not implemented yet".

## Fluxo do dia a dia

```bash
srectl up                 # sobe a plataforma (idempotente)
srectl sync               # força o Argo CD a reconciliar
srectl demo load          # gera tráfego em api-a/api-b + fraudctl
srectl demo chaos         # injeta latência/mata pods
srectl demo verify        # valida stack + SLOs saudáveis
srectl open api-a         # port-forward de um serviço
srectl down               # derruba tudo
```

O `srectl` é um **task runner como o Make** — cada comando envolve `kind`/`kubectl`/`helm`
e adiciona uma UX melhor (tabela de status, port-forward, progresso).

## Explore a stack

| O quê | Comando / URL |
| --- | --- |
| UI do Argo CD | `srectl open argocd` |
| Grafana | `srectl open grafana` (ou `https://grafana.localhost`) |
| api-a | `srectl open api-a` → `/orders`, `/slow`, `/error`, `/metrics` |
| api-b | `srectl open api-b` |
| fraudctl | `srectl open fraudctl` → `POST /fraud-score` |
| dashboard certificate-validate | `https://certificate-validate.localhost` |
| visão ao vivo do stracectl | `srectl trace` (port-forward) |

## Logs (Loki)

O Alloy faz o tail dos logs de todos os pods e os envia ao Loki com as labels
`namespace`, `pod`, `container`, `app`, `node` e um `level` extraído. Consulte em
**Grafana → Dashboards → `srectl · Logs`** (filtros: `namespace`, `search`) ou em
**Explore → Loki** com LogQL:

```logql
{namespace="argocd"}                       # um namespace
{namespace="monitoring", container="grafana"}
{app="hello"}                              # pela label de app
{level="error"}                            # pela label level extraída
{namespace=~"argocd|monitoring"} |= "error"  # regex de namespace + grep de linha
sum by (namespace) (count_over_time({level="error"}[5m]))
```

## Golden signals (Prometheus)

O `srectl · Golden Signals` (Grafana → Dashboards) cobre os quatro golden signals,
alimentado pelo ingress-nginx (tráfego/latência/erros) e node-exporter +
kube-state-metrics (saturação):

- **Tráfego** — `sum(rate(nginx_ingress_controller_requests[5m]))`
- **Erros** — `100 * sum(rate(nginx_ingress_controller_requests{status=~"5.."}[5m])) / sum(rate(nginx_ingress_controller_requests[5m]))`
- **Latência** — `histogram_quantile(0.99, sum by (le) (rate(nginx_ingress_controller_request_duration_seconds_bucket[5m])))`
- **Saturação** — % de CPU / memória / filesystem dos nós e top namespaces por CPU / memória

Uma visão de todo o cluster fica em **`srectl · Cluster`** (nós prontos, pods rodando,
restarts, CPU/memória usados vs alocáveis, top pods). `kubectl top nodes` e
`kubectl top pods` usam o metrics-server.

## Segurança

- **Kyverno** aplica políticas baseline (CEL `ValidatingPolicy`) no namespace `demo`:
  sem tag `:latest`, requests/limits obrigatórios, sem `hostPath`, e `runAsNonRoot`
  é reportado (Audit).
  ```bash
  kubectl run bad -n demo --image=nginx:latest   # rejeitado por disallow-latest-tag
  ```
- **Trivy Operator** escaneia os workloads de demo (vulnerabilidades, config audit e
  RBAC): `kubectl get vulnerabilityreport,configauditreport -n demo`.
- **Falco** detecta ameaças em runtime com probe modern eBPF e envia eventos ao Loki
  (via Falcosidekick). Teste:
  ```bash
  kubectl exec -n demo deploy/hello -- cat /etc/shadow   # "Read sensitive file untrusted"
  ```
- **NetworkPolicies** aplicam default-deny no namespace `demo` e liberam só DNS,
  ingress-nginx, probes do kubelet e tráfego intra-namespace (atuando via kindnet).

Dashboard: **Grafana → Dashboards → `srectl · Security`**.

## Apps demo (api-a / api-b)

O workload demo produz os sinais RED, traces distribuídos e logs:

- `api-a` (pedidos) — `https://api-a.localhost` → `/health`, `/ready`, `/orders`, `/slow`, `/error`.
  `/orders` chama a `api-b` (pagamentos), então o trace cobre os dois serviços.
- `api-b` (pagamentos) — `POST /pay` interno, chamado pela api-a.

```bash
srectl demo load --duration 30s --rps 20   # gera tráfego (mix de /orders, /slow, /error)
```

- **Traces**: Grafana → Explore → Tempo (serviços `api-a` e `api-b`, mesmo traceID).
- **Métricas**: dashboard `srectl · App RED` (taxa/erros/latência por serviço).
- **Logs**: Loki `{namespace="demo", app="api-a"}` — cada linha carrega `trace_id`.

## Cenários de demo

```text
srectl demo load      # hammer k6/go → dashboards RED acendem
srectl demo error     # força erros → alerta HighErrorRate dispara
srectl demo chaos     # latência na api-a, mata pods do fraudctl
srectl demo security  # report do trivy + violação do kyverno + evento do falco
srectl demo certs     # scan do certificate-validate + alerta de cert expirando
srectl demo projects  # toca fraudctl + stracectl + certificate-validate
```

## Estude a stack (site)

O site bilíngue (GitHub Pages) é a superfície de aprendizado — veja [`SITE.md`](SITE.md).

```bash
make site-dev      # http://localhost:1313 (EN + PT-BR)
```

## Configuração

`config/srectl.yaml` (override com variáveis `SRECTL_*`):

```yaml
cluster:
  name: srectl
  nodes: 3            # 1 control-plane + 2 workers

components:
  argocd: true
  observability: true
  security: true
  chaos: true
  apps: true
  projects: true

argocd:
  namespace: argocd
  version: "10.18.0"  # versão do chart argo/argo-cd
  repo: https://github.com/fabianoflorentino/srectl
  revision: main

registry: localhost:5000   # registry de imagens para builds locais
```

> `argocd.repo` deve apontar para um repositório git que o Argo CD consiga clonar
> (ex.: este repo publicado no GitHub). Para uma demo totalmente local, sirva o repo
> com `git http-backend` e ajuste `SRECTL_ARGOCD_REPO`.

## Índice de documentação

- [`PLAN.md`](PLAN.md) — roteiro de implementação.
- [`ARCHITECTURE.md`](ARCHITECTURE.md) — diagramas e topologia.
- [`STACK.md`](STACK.md) — catálogo de componentes e justificativas.
- [`OBSERVABILITY.md`](OBSERVABILITY.md) — métricas/logs/traces + SLOs.
- [`SECURITY.md`](SECURITY.md) — pilares de segurança.
- [`CI_CD.md`](CI_CD.md) — design do pipeline.
- [`GITOPS.md`](GITOPS.md) — modelo do Argo CD.
- [`SITE.md`](SITE.md) — site do GitHub Pages.

> As versões em inglês estão em [`../en/`](../en/).

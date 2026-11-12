# Começando

> **Status: planejamento.** Este projeto ainda não foi implementado — veja
> [`PLAN.md`](PLAN.md). Os comandos abaixo são a UX alvo quando as Fases 0–9 forem
> concluídas.

## Pré-requisitos

- Go 1.27+
- Docker (rodando)
- kind, kubectl, helm no `PATH`
- `srectl` buildado: `make build` (ou `go install ./cmd/srectl`)

## Um comando para subir

```bash
srectl up          # cluster kind + argo cd + tudo (via GitOps)
srectl status      # matriz de saúde
srectl open grafana
```

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
| visão ao vivo do stracectl | `srectl trace api-a` |

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

registry: localhost:5000   # registry de imagens para builds locais
```

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

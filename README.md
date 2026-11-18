# srectl

A local, reproducible **SRE / DevOps / Platform Engineering** training environment on
[kind](https://kind.sigs.k8s.io/), driven by a single Go CLI.

> **Status: implemented (phases 0–9).** Cluster, GitOps, LGTM observability, security,
> demo apps, project integrations, SRE drills, the docs site and CI are all in place —
> see [`docs/en/PLAN.md`](docs/en/PLAN.md).

## What it is

One command (`srectl up`) brings up a complete platform on a local kind cluster:

- **GitOps** — Argo CD (app-of-apps).
- **Observability (LGTM)** — Prometheus + Grafana + Alertmanager, Loki + Alloy (logs),
  Tempo (traces via OTLP), and metrics-server (cluster metrics / `kubectl top`).
- **Security** — Trivy Operator, Falco, Kyverno, cert-manager, NetworkPolicies.
- **Chaos** — chaos-mesh for SRE drills.
- **Demo apps** — two instrumented Go APIs (api-a → api-b) for distributed tracing,
  RED metrics and structured logs.
- **Our projects, integrated** — [`fraudctl`](https://github.com/fabianoflorentino/fraudctl),
  [`stracectl`](https://github.com/fabianoflorentino/stracectl),
  [`certificate-validate`](https://github.com/fabianoflorentino/certificate-validate)
  and [`seraph`](https://github.com/fabianoflorentino/seraph) running as
  first-class workloads.
- **Study site** — a bilingual (EN/PT-BR) Hugo site on GitHub Pages.

The `srectl` CLI is a **simple task runner like Make** — it wraps `kind`/`kubectl`/`helm`
and adds nicer UX.

## Quick start

```bash
make build && srectl up
srectl open grafana
srectl demo load      # generate traffic
srectl demo chaos     # inject faults
srectl down
```

## Documentation (bilingual)

| Doc | EN | PT-BR |
| --- | --- | --- |
| Implementation plan | [`en/PLAN.md`](docs/en/PLAN.md) | [`pt-br/PLAN.md`](docs/pt-br/PLAN.md) |
| Architecture | [`en/ARCHITECTURE.md`](docs/en/ARCHITECTURE.md) | [`pt-br/ARCHITECTURE.md`](docs/pt-br/ARCHITECTURE.md) |
| Component catalog | [`en/STACK.md`](docs/en/STACK.md) | [`pt-br/STACK.md`](docs/pt-br/STACK.md) |
| Observability | [`en/OBSERVABILITY.md`](docs/en/OBSERVABILITY.md) | [`pt-br/OBSERVABILITY.md`](docs/pt-br/OBSERVABILITY.md) |
| Security | [`en/SECURITY.md`](docs/en/SECURITY.md) | [`pt-br/SECURITY.md`](docs/pt-br/SECURITY.md) |
| CI/CD | [`en/CI_CD.md`](docs/en/CI_CD.md) | [`pt-br/CI_CD.md`](docs/pt-br/CI_CD.md) |
| GitOps | [`en/GITOPS.md`](docs/en/GITOPS.md) | [`pt-br/GITOPS.md`](docs/pt-br/GITOPS.md) |
| Site (GitHub Pages) | [`en/SITE.md`](docs/en/SITE.md) | [`pt-br/SITE.md`](docs/pt-br/SITE.md) |
| Getting started | [`en/GETTING_STARTED.md`](docs/en/GETTING_STARTED.md) | [`pt-br/GETTING_STARTED.md`](docs/pt-br/GETTING_STARTED.md) |

## Repository layout

```text
cmd/srectl/        CLI entrypoint
internal/       cli, config, kind, k8s, argocd, demo, chaos, verify, appobs
apps/           demo APIs (api-a, api-b)
deploy/         GitOps source of truth (Argo CD)
site/           bilingual Hugo site (GitHub Pages)
docs/           en/ + pt-br/ documentation
config/         srectl.yaml defaults
.github/        CI workflows
```

## License

To be determined (see LICENSE once added).

---

# srectl

Um ambiente de treinamento **SRE / DevOps / Platform Engineering** local e reprodutível
em [kind](https://kind.sigs.k8s.io/), dirigido por uma única CLI em Go.

> **Status: implementado (fases 0–9).** Cluster, GitOps, observabilidade LGTM, segurança,
> apps demo, integração dos projetos, drills de SRE, site de docs e CI estão prontos —
> veja [`docs/pt-br/PLAN.md`](docs/pt-br/PLAN.md).

## O que é

Um comando (`srectl up`) sobe uma plataforma completa num cluster kind local:

- **GitOps** — Argo CD (app-of-apps).
- **Observabilidade (LGTM)** — Prometheus + Grafana + Alertmanager, Loki + Alloy (logs),
  Tempo (traces via OTLP) e metrics-server (métricas de cluster / `kubectl top`).
- **Segurança** — Trivy Operator, Falco, Kyverno, cert-manager, NetworkPolicies.
- **Caos** — chaos-mesh para drills de SRE.
- **Apps demo** — duas APIs Go instrumentadas (api-a → api-b) para tracing distribuído,
  métricas RED e logs estruturados.
- **Nossos projetos, integrados** — [`fraudctl`](https://github.com/fabianoflorentino/fraudctl),
  [`stracectl`](https://github.com/fabianoflorentino/stracectl),
  [`certificate-validate`](https://github.com/fabianoflorentino/certificate-validate)
  e [`seraph`](https://github.com/fabianoflorentino/seraph) rodando como
  workloads de primeira classe.
- **Site de estudo** — um site Hugo bilíngue (EN/PT-BR) no GitHub Pages.

A CLI `srectl` é um **task runner simples como o Make** — envolve `kind`/`kubectl`/`helm`
e adiciona uma UX melhor.

## Começo rápido

```bash
make build && srectl up
srectl open grafana
srectl demo load      # gera tráfego
srectl demo chaos     # injeta falhas
srectl down
```

## Documentação (bilíngue)

| Doc | EN | PT-BR |
| --- | --- | --- |
| Plano de implementação | [`en/PLAN.md`](docs/en/PLAN.md) | [`pt-br/PLAN.md`](docs/pt-br/PLAN.md) |
| Arquitetura | [`en/ARCHITECTURE.md`](docs/en/ARCHITECTURE.md) | [`pt-br/ARCHITECTURE.md`](docs/pt-br/ARCHITECTURE.md) |
| Catálogo de componentes | [`en/STACK.md`](docs/en/STACK.md) | [`pt-br/STACK.md`](docs/pt-br/STACK.md) |
| Observabilidade | [`en/OBSERVABILITY.md`](docs/en/OBSERVABILITY.md) | [`pt-br/OBSERVABILITY.md`](docs/pt-br/OBSERVABILITY.md) |
| Segurança | [`en/SECURITY.md`](docs/en/SECURITY.md) | [`pt-br/SECURITY.md`](docs/pt-br/SECURITY.md) |
| CI/CD | [`en/CI_CD.md`](docs/en/CI_CD.md) | [`pt-br/CI_CD.md`](docs/pt-br/CI_CD.md) |
| GitOps | [`en/GITOPS.md`](docs/en/GITOPS.md) | [`pt-br/GITOPS.md`](docs/pt-br/GITOPS.md) |
| Site (GitHub Pages) | [`en/SITE.md`](docs/en/SITE.md) | [`pt-br/SITE.md`](docs/pt-br/SITE.md) |
| Começando | [`en/GETTING_STARTED.md`](docs/en/GETTING_STARTED.md) | [`pt-br/GETTING_STARTED.md`](docs/pt-br/GETTING_STARTED.md) |

## Layout do repositório

```text
cmd/srectl/        entrypoint da CLI
internal/       cli, config, kind, k8s, argocd, demo, chaos, verify, appobs
apps/           APIs demo (api-a, api-b)
deploy/         fonte de verdade GitOps (Argo CD)
site/           site Hugo bilíngue (GitHub Pages)
docs/           documentação en/ + pt-br/
config/         defaults srectl.yaml
.github/        workflows de CI
```

## Licença

A definir (veja o LICENSE quando adicionado).

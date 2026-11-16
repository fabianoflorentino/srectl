# Getting Started

> **Status: in progress.** Phases 0–2 are implemented: the CLI skeleton, the kind
> cluster with ingress/TLS, and the GitOps foundation (Argo CD app-of-apps). The
> roadmap lives in [`PLAN.md`](PLAN.md).

## Prerequisites

- Go 1.27+
- Docker (running)
- kind, kubectl, helm on `PATH`
- `srectl` built: `make build` (or `go install ./cmd/srectl`)
- The repo pushed to a git server Argo CD can reach (see `argocd.repo` below)

## One command up

```bash
srectl up          # kind cluster + Argo CD + everything (via GitOps)
srectl status      # health matrix (cluster + apps)
srectl open argocd # port-forward the Argo CD UI
```

## What works today (Phases 0–2)

```text
srectl up            create cluster + bootstrap Argo CD (idempotent)
srectl down          tear down the cluster
srectl status        health matrix: cluster + every Argo CD app
srectl cluster       create/delete/status the kind cluster
srectl bootstrap     install Argo CD + register the root app
srectl sync [app]    force Argo CD to refresh (default: all)
srectl creds         show Argo CD + Grafana admin credentials
srectl creds rotate  rotate a component's password (`creds rotate argocd`)
srectl open          port-forward a service (argocd, grafana, prometheus, alertmanager, hello)
srectl version       build info
```

`srectl open` blocks in the foreground (Ctrl+C to stop); add `--daemon` to run it in
the background and `--down` to stop it (`srectl open argocd --daemon` → `srectl open argocd --down`).

`srectl creds` prints the Argo CD and Grafana logins; Loki and Tempo run without
authentication in this local stack.

Commands for later phases (`demo`, `apps`, `trace`, `certs`, `logs`) exist as stubs
and report "not implemented yet".

## Everyday flow

```bash
srectl up                 # bring the platform up (idempotent)
srectl sync               # force Argo CD to reconcile
srectl demo load          # generate traffic on api-a/api-b + fraudctl
srectl demo chaos         # inject latency/kill pods
srectl demo verify        # assert stack + SLOs healthy
srectl open api-a         # port-forward a service
srectl down               # tear down
```

`srectl` is a **task runner like Make** — each command wraps `kind`/`kubectl`/`helm` and
adds nicer UX (status table, port-forward, progress).

## Explore the stack

| What | Command / URL |
| --- | --- |
| Argo CD UI | `srectl open argocd` |
| Grafana | `srectl open grafana` (or `https://grafana.localhost`) |
| api-a | `srectl open api-a` → `/orders`, `/slow`, `/error`, `/metrics` |
| api-b | `srectl open api-b` |
| fraudctl | `srectl open fraudctl` → `POST /fraud-score` |
| certificate-validate dashboard | `https://certificate-validate.localhost` |
| stracectl live view | `srectl trace api-a` |

## Demo scenarios

```text
srectl demo load      # k6/go hammer → RED dashboards light up
srectl demo error     # force errors → HighErrorRate alert fires
srectl demo chaos     # latency on api-a, kill fraudctl pods
srectl demo security  # trivy report + kyverno violation + falco event
srectl demo certs     # certificate-validate scan + expiring-cert alert
srectl demo projects  # touch fraudctl + stracectl + certificate-validate
```

## Study the stack (site)

The bilingual site (GitHub Pages) is the learning surface — see [`SITE.md`](SITE.md).

```bash
make site-dev      # http://localhost:1313 (EN + PT-BR)
```

## Configuration

`config/srectl.yaml` (override with `SRECTL_*` env vars):

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
  version: "10.18.0"  # argo/argo-cd Helm chart version
  repo: https://github.com/fabianoflorentino/srectl
  revision: main

registry: localhost:5000   # image registry for local builds
```

> `argocd.repo` must point at a git repository Argo CD can clone (e.g. this repo
> pushed to GitHub). For a fully local demo you can serve the repo with
> `git http-backend` and set `SRECTL_ARGOCD_REPO` accordingly.

## Documentation index

- [`PLAN.md`](PLAN.md) — implementation roadmap.
- [`ARCHITECTURE.md`](ARCHITECTURE.md) — diagrams and topology.
- [`STACK.md`](STACK.md) — component catalog and rationale.
- [`OBSERVABILITY.md`](OBSERVABILITY.md) — metrics/logs/traces + SLOs.
- [`SECURITY.md`](SECURITY.md) — security pillars.
- [`CI_CD.md`](CI_CD.md) — pipeline design.
- [`GITOPS.md`](GITOPS.md) — Argo CD model.
- [`SITE.md`](SITE.md) — GitHub Pages site.

> Portuguese versions live in [`../pt-br/`](../pt-br/).

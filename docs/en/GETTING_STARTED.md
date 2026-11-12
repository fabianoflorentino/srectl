# Getting Started

> **Status: planning.** This project is not implemented yet — see
> [`PLAN.md`](PLAN.md). The commands below are the target UX once Phases 0–9 land.

## Prerequisites

- Go 1.27+
- Docker (running)
- kind, kubectl, helm on `PATH`
- `srectl` built: `make build` (or `go install ./cmd/srectl`)

## One command up

```bash
srectl up          # kind cluster + argo cd + everything (via GitOps)
srectl status      # health matrix
srectl open grafana
```

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

registry: localhost:5000   # image registry for local builds
```

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

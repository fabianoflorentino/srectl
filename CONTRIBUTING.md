# Contributing to srectl

_Leia em [Português](CONTRIBUTING.pt-br.md)._

Thanks for helping! `srectl` is a task runner that stands up a full SRE / DevOps /
Platform Engineering training platform on kind, driven by GitOps.

## Prerequisites

- Go **1.27+**, Docker, [kind](https://kind.sigs.k8s.io/), `kubectl`, `helm`
- [`golangci-lint`](https://golangci-lint.run/) v2.x
- Optional: [Hugo](https://gohugo.io/) (for the site)

## Quick start

```bash
make build                 # build ./bin/srectl
./bin/srectl up            # cluster + Argo CD + every GitOps app
./bin/srectl status        # Synced · Healthy matrix
./bin/srectl demo verify   # assert the whole stack is healthy
```

## Development workflow

```bash
make            # help
make test       # go test -race
make coverage   # enforce >= 80% statement coverage
make lint       # golangci-lint
make fmt        # gofmt + golangci-lint fmt
make site-dev   # local docs site
```

Go rules for this repo:

- **Coverage must stay >= 80%** (`make coverage` fails below it).
- Tests are executable specifications — assert public behavior, not internals.
- Follow the Go agent skills referenced in `AGENTS.md`.

## Commits

[Conventional Commits](https://www.conventionalcommits.org/), written in English, e.g.
`feat(demo): add chaos latency scenario`, `fix(observability): honorLabels`.

## Adding a GitOps app

1. Create `deploy/bootstrap/apps/<name>-app.yaml` (Helm source or a repo path).
2. `git commit && git push`.
3. `kubectl annotate application root-app -n argocd argocd.argoproj.io/refresh=hard --overwrite`
   (or `./bin/srectl sync`).
4. Confirm with `./bin/srectl status`.

Pin image tags — **never `latest`**; CI rejects it in `deploy/`.

## Tests

- Unit: `make test` (race) and `make coverage`.
- Integration: `make test-integration` (needs a running cluster).
- E2E (CI): `.github/workflows/e2e.yml` — `srectl up` → `demo verify` → `down`.

## Pull request checklist

- [ ] Tests and lint green (`make test lint coverage`).
- [ ] New behavior documented in `docs/en/*` **and** `docs/pt-br/*`.
- [ ] Conventional commit message (English).
- [ ] No `latest` tags or secrets committed.

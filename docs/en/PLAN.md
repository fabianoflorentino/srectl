# Implementation Plan

This document is the source of truth for building **srectl**. It breaks the work
into ordered phases, each with concrete tasks and acceptance criteria. Nothing in this
plan is code yet — it is the roadmap we execute against.

## Goals

Build a local, reproducible SRE / DevOps / Platform Engineering training environment
that:

1. Runs entirely on **kind** (Kubernetes-in-Docker) — no cloud, no external dependencies.
2. Is driven by a simple Go CLI, **`srectl`**, a *task runner like Make* (see below).
3. Covers the full platform surface: **GitOps, observability (LGTM), security, CI/CD,
   chaos engineering**.
4. Ships **two demo Go APIs** (instrumented) to exercise the stack end-to-end.
5. Integrates three of our own projects — **fraudctl**, **stracectl** and
   **certificate-validate** — as first-class workloads.
6. Publishes a **GitHub Pages** site (Hugo, bilingual EN/PT-BR) to explain and teach
   SRE / DevOps / Platform Engineering through the project.

## Non-goals (for now)

- Production hardening, HA, multi-cluster, multi-cloud.
- External DNS / public certificates (we use cert-manager self-signed).
- Building new products inside this repo (we consume fraudctl/stracectl/certificate-validate).

## The `srectl` CLI is a task runner

`srectl` is deliberately **simple** — the equivalent of a `Makefile` written in Go:

- Each subcommand wraps one well-known tool (`kind`, `kubectl`, `helm`) plus a little
  Go glue for nice UX (status table, port-forward, progress output).
- No heavyweight Kubernetes client-go in the hot path; it shells out where possible.
- Configuration lives in `config/srectl.yaml` (like a `.env`), overridable by `SRECTL_*` env vars.
- It is **not** a framework, an operator, or a controller.

Think of it as `make up`, `make status`, `make demo`, … but discoverable via
`srectl --help`, with colored status output and sane defaults.

## Architecture decisions (locked)

| Decision | Choice | Rationale |
| --- | --- | --- |
| Cluster | kind, 3 nodes (1 control-plane + 2 workers) | Real node scheduling, affinities, tolerations |
| Deploy model | **Argo CD (GitOps)** | `deploy/` is the single source of truth; app-of-apps pattern |
| CLI | Go task runner (make-like) | Simple, discoverable, wraps kind/kubectl/helm |
| Ingress | ingress-nginx (kind port-mapping 80/443) | Most common; cert-manager integration |
| TLS | cert-manager + selfsigned ClusterIssuer | Automated certs, feeds certificate-validate |
| Metrics | kube-prometheus-stack (Prometheus + Grafana + Alertmanager) | De-facto standard |
| Logs | Loki + Alloy | Single agent for logs + OTel pipelines |
| Traces | Tempo (OTLP receivers) | Standard tracing backend |
| Security | Trivy (Operator) + Falco + Kyverno | Scan, runtime, policy |
| Chaos | chaos-mesh | Latency/failure injection for SRE drills |
| Site | Hugo, bilingual EN/PT-BR, GitHub Pages | Same tooling as our other projects |
| Language | Go 1.27 | Same as the rest of our stack |

See [`STACK.md`](STACK.md) for the full component catalog and [`ARCHITECTURE.md`](ARCHITECTURE.md)
for diagrams.

---

## Phase 0 — Repository & CLI skeleton

**Objective:** a compiling Go CLI (task runner) with the command surface wired up, and
the repo structure the GitOps layer will consume.

Tasks:

- [x] Initialize the Go module (`github.com/fabianoflorentino/srectl`).
- [x] Stand up the Cobra CLI in `cmd/srectl/` with the full command tree (stubs).
- [x] Implement configuration loading (`config/srectl.yaml` + env overrides, `SRECTL_` prefix).
- [x] `internal/kind/` — kind cluster create/delete/status (config file, node count).
- [x] `internal/k8s/` — thin wrappers that shell out to `kubectl`/`helm` + port-forward.
- [x] `internal/demo/` — load generator, chaos and verify task scripts.
- [x] `Makefile` (build, test, lint, fmt, docker-build) and `lefthook.yml`.
- [x] `.github/workflows/ci.yml` skeleton (lint + test + build).
- [x] `README.md` (bilingual), `docs/` scaffolding (en + pt-br).

**Command surface (target):**

```text
srectl up            create cluster + bootstrap + wait healthy
srectl down          teardown (cluster + kind)
srectl status        component health matrix
srectl cluster       create/delete/status the kind cluster
srectl bootstrap     install Argo CD + register the root app
srectl sync          force Argo CD sync of an app (default: all)
srectl apps          deploy/restart/scale the demo apps
srectl demo          run a demo scenario (load, chaos, break, verify)
srectl trace         run stracectl against a pod
srectl certs         trigger certificate-validate scan
srectl open          port-forward a service (grafana, argocd, api-a, ...)
srectl logs          tail logs of a workload
srectl version       build info
```

**Acceptance criteria:**

- `srectl version` and `srectl --help` work.
- `go build ./...`, `go test ./...`, `golangci-lint run` all pass in CI.

---

## Phase 1 — Cluster, ingress & TLS

**Objective:** a kind cluster with ingress and automated certificates.

Tasks:

- [x] `deploy/kind/kind-config.yaml` — 3 nodes, ingress-nginx `extraPortMappings` (80/443).
- [x] ingress-nginx deployed (via imperative bootstrap; Argo CD from Phase 2).
- [x] cert-manager + selfsigned `ClusterIssuer` + trust bundle `ConfigMap`.
- [x] A `hello` echo app proving: ingress → TLS → pod round trip.
- [x] `srectl up` drives this whole phase idempotently.

**Acceptance criteria:**

- `srectl up` creates the cluster and a pod is reachable via `https://*.localhost`.
- `srectl status` reports cluster, ingress and cert-manager as `Ready`.

---

## Phase 2 — GitOps foundation (Argo CD)

**Objective:** everything after this phase is deployed declaratively.

Tasks:

- [x] Install Argo CD (imperative bootstrap — this is the only imperative step).
- [x] Define the **root app** (`deploy/bootstrap/argocd/`) using app-of-apps.
- [x] Define `ApplicationSet`/`Application` CRs for each component group:
      `observability`, `security`, `apps`, `projects`, `ingress`, `chaos`.
- [x] `srectl sync` + `srectl status` integrate with the Argo CD API.
- [x] Default credentials rotation + `srectl open argocd`.

**Acceptance criteria:**

- Editing `deploy/` triggers Argo CD to converge the cluster (self-healing, prune).
- The `argocd` App shows `Healthy + Synced` for the root app.

---

## Phase 3 — Observability (LGTM)

**Objective:** metrics, logs, traces and alerting, unified in Grafana.

Tasks:

- [x] **kube-prometheus-stack**: Prometheus, Alertmanager, Grafana, node-exporter,
      kube-state-metrics, Prometheus Operator CRDs (ServiceMonitor/PodMonitor).
- [x] **Loki** (single-binary) + **Alloy** DaemonSet (pod-log tailing, labels for
      namespace/pod/container/app/node, extracted `level`).
- [x] **Tempo** (single-binary) — apps export OTLP directly to Tempo (no separate collector).
- [x] Grafana datasources (Prometheus, Loki, Tempo) provisioned via ConfigMap.
- [ ] Baseline dashboards: cluster overview, node/pod, app RED, syscall (stracectl),
      certificates (certificate-validate). (`srectl · Logs` and `srectl · Golden Signals` dashboards exist)
- [ ] Baseline alert rules: high error rate, high p99, disk/cpu, certificate
      expiry, node pressure (a `TestAlert` already proves the alerting pipeline).
- [x] `srectl open grafana` + default dashboards discoverable.

**Acceptance criteria:**

- `srectl up` leaves Grafana with the three datasources and dashboards ready.
- Scraping targets for the demo apps and our projects show as `UP` in Prometheus.
- A generated error in a demo app fires a test alert in Alertmanager.

---

## Phase 4 — Security

**Objective:** scan, runtime detection and policy enforcement.

Tasks:

- [x] **Trivy Operator**: vulnerability/config-audit/RBAC reports (demo namespace),
      ServiceMonitor metrics for Grafana.
- [x] **Falco**: runtime detection with the modern eBPF probe; Falcosidekick → Loki.
- [x] **Kyverno**: baseline policies as CEL ValidatingPolicy (disallow latest tag,
      require requests/limits, require runAsNonRoot (Audit), block hostPath).
- [x] NetworkPolicies for the demo namespace (default-deny + allow matrix, enforced by kindnet).
- [x] cert-manager (from Phase 1) documented as the TLS pillar.
- [x] Dashboard: vulnerability + policy view in Grafana (`srectl · Security`).

**Acceptance criteria:**

- `srectl up` leaves Trivy reports populated and Kyverno policies enforcing.
- A non-compliant workload (e.g. `latest` tag) is rejected by Kyverno.
- Falco produces at least one demo rule event (`srectl demo security`).

---

## Phase 5 — Demo applications

**Objective:** the workload that makes the platform observable and demonstrable.

Tasks:

- [x] **api-a** (`apps/api-a/`): orders/checkout service.
      - `net/http` + Go 1.22 routing; `/health`, `/ready`, `/orders`, `/slow`, `/error`.
      - OTel SDK (OTLP → Tempo), Prometheus metrics (RED), `slog` structured logs.
      - Calls api-b (demonstrates distributed tracing).
- [x] **api-b** (`apps/api-b/`): payments service, same instrumentation, called by api-a.
- [x] Dockerfiles + `deploy/apps/` manifests (Deployment/Service/Ingress/ServiceMonitor).
- [x] Load generator (`srectl demo load`) — a small Go hammer — to produce signal.

**Acceptance criteria:**

- Trace through `api-a → api-b` is queryable in Tempo/Grafana.
- RED metrics for both APIs visible on a shared dashboard.
- Logs from both APIs searchable in Loki.

---

## Phase 6 — Integrate our projects

**Objective:** run fraudctl, stracectl and certificate-validate as real workloads.

Tasks:

- [x] **fraudctl** (`deploy/projects/fraudctl/`): LB + 2 API workers in one Pod
      (SCM_RIGHTS over a shared emptyDir), `POST /fraud-score` via ingress; throughput
      /latency via ingress-nginx RED metrics, logs in Loki.
- [x] **certificate-validate** (`deploy/projects/certificate-validate/`):
      `k8s monitor` DaemonSet scanning TLS Secrets/Ingresses; `--metrics-addr=:9102`;
      ServiceMonitor → Prometheus; webhook → Alertmanager/webhook-receiver; dashboard
      behind ingress with cert-manager TLS.
- [x] **stracectl** (`deploy/projects/stracectl/`): dedicated demo pod + sidecar
      (`attach --serve :8080`), syscall metrics → Prometheus, syscall dashboard;
      `srectl trace` port-forwards the live view.
- [x] **seraph** (`deploy/projects/seraph/`): static status board (pinned image) behind
      ingress; `srectl open seraph`.
- [x] Alert rules: `certificate_days_left < 15` (CertificateExpiringSoon/Expired);
      syscall anomaly counters deferred.
- [ ] `srectl demo projects` — an end-to-end script that touches all three.

**Acceptance criteria:**

- `certificate_days_left` metrics appear in Prometheus with k8s labels.
- `srectl trace` opens a live stracectl view of api-a.
- `srectl demo load` shows fraudctl throughput/latency on its dashboard.

---

## Phase 7 — SRE drills (chaos + SLOs)

**Objective:** turn observability into action.

Tasks:

- [x] **chaos-mesh** install + CRDs. PodChaos + StressChaos validated; NetworkChaos/
      DNSChaos (iptables/ipset) do not work on kind's nested containers.
- [x] `srectl demo chaos [latency|pod-kill]` (embedded manifests; latency = CPU stress on
      api-a, pod-kill = fraudctl). DNS/packet-loss need NetworkChaos, unsupported on kind.
- [x] SLO/SLI recording rules + `ApiAHighErrorRate` / `ApiAHighLatency` alerts
      (`deploy/observability/slo-rules.yaml`).
- [x] Runbooks (`docs/en/runbooks/`, `docs/pt-br/runbooks/`) for the top alerts.
- [x] `srectl demo verify` — asserts cluster up, Argo CD ready and every app Synced·Healthy.

**Acceptance criteria:**

- Injecting latency on api-a raises p99 on the dashboard and fires an alert.
- Killing fraudctl pods recovers within the defined SLO (self-healing).

---

## Phase 8 — Documentation site (GitHub Pages)

**Objective:** a bilingual site that teaches SRE / DevOps / Platform Engineering
through this project.

Tasks:

- [x] Hugo project in `site/`, bilingual (**EN** default + **PT-BR** at `/pt-br/`),
      custom layouts (no theme), matching our other projects.
- [x] Content sections: overview, architecture, the stack, and per-theme learning paths
      (**SRE**, **DevOps**, **Platform Engineering**).
- [x] Hands-on tutorials driven by `srectl demo` scenarios (observability, security, chaos).
- [x] Runbooks + SLO pages (site summaries; full runbooks live in `docs/*/runbooks/`).
- [x] Diagrams (Mermaid) rendered in the site (architecture page).
- [x] GitHub Pages deploy via GitHub Actions (`.github/workflows/site.yaml`).

**Acceptance criteria:**

- `make site-dev` serves the bilingual site locally.
- Every page is available in EN and PT-BR.
- GitHub Actions publishes the site on merge to `main`.

---

## Phase 9 — CI/CD & polish

**Objective:** a CI pipeline that validates everything reproducibly.

Tasks:

- [x] GitHub Actions: lint, race + coverage test, build the CLI + both APIs,
      and a GitOps guard against `:latest`.
- [x] Build & publish images: `images.yml` pushes `srectl-api-a`/`srectl-api-b` to GHCR
      pinned by SHA/tag; the three projects are published by their own CI.
- [x] E2E job (`e2e.yml`, nightly/manual): `srectl up` on kind → `demo verify` → `down`.
- [x] Image versioning pinned by tag/SHA; CI fails on any `:latest` in `deploy/`.
- [x] Finalize docs: dashboards, runbooks and a bilingual contribution guide
      (`CONTRIBUTING.md` / `CONTRIBUTING.pt-br.md`).

**Acceptance criteria:**

- A green CI run on every PR.
- `srectl up && srectl demo verify && srectl down` completes unattended in CI.
- The site builds and deploys to GitHub Pages automatically.

---

## Definition of done (per phase)

- [ ] Code merged (English, conventional commits).
- [ ] Tests + lint green.
- [ ] Documented in the relevant `docs/en/*.md` **and** `docs/pt-br/*.md`.
- [ ] Reproducible via a single `srectl` command or `make` target.

## Sequencing rationale

- Phases 0–2 are the platform skeleton; nothing else can be demonstrated without them.
- Phase 3 (observability) lands **before** the apps so the apps are born observable.
- Phase 5 (apps) lands before Phase 6 (projects) so the projects have an instrumented
  cluster to observe.
- Phase 7 (chaos) is last because it assumes a healthy, observable, secured cluster to
  break.
- Phase 8 (site) runs in parallel with 3–7 once the platform exists.

# Component Catalog

Every component in the stack, why it was chosen, and how it is installed.

Legend — install method:

- **GitOps**: applied declaratively by Argo CD from `deploy/`.
- **Bootstrap**: installed imperatively once by `srectl bootstrap` (only Argo CD).
- **CLI**: managed by the `srectl` CLI (kind itself).

## Platform

| Component | Purpose | Install | Why |
| --- | --- | --- | --- |
| kind | Local Kubernetes | CLI | Lightweight, multi-node, no cloud |
| ingress-nginx | Edge routing + TLS termination | GitOps | Most common ingress controller |
| cert-manager | Automated TLS issuance/rotation | GitOps | Feeds certificate-validate; standard |
| Argo CD | GitOps deployment engine | Bootstrap | Declarative single source of truth |
| `srectl` CLI | Task runner (make-like) | — | Wraps kind/kubectl/helm for a nice UX |

## Observability (LGTM)

| Component | Purpose | Install | Why |
| --- | --- | --- | --- |
| kube-prometheus-stack | Prometheus, Alertmanager, Grafana, node-exporter, kube-state-metrics, Operator CRDs | GitOps (Helm) | De-facto standard, batteries included |
| Prometheus Operator | ServiceMonitor/PodMonitor/Alert CRDs | GitOps | Standard scraping model |
| Loki | Log aggregation | GitOps (Helm) | Grafana-native, simple |
| Alloy | Log tailing + OTel pipeline agent | GitOps (DaemonSet) | One agent, replaces Promtail/collector-side |
| Tempo | Distributed tracing backend | GitOps (Helm) | Grafana-native traces |
| OpenTelemetry SDK (in-app) | Traces from api-a/api-b via OTLP | built into the apps | no extra service |
| Grafana | Dashboards + unified UI | GitOps (Helm) | Single pane of glass |

## Security

| Component | Purpose | Install | Why |
| --- | --- | --- | --- |
| Trivy Operator | Image/CIS scanning + reports | GitOps (Helm) | Continuous, CRD-based |
| Falco | Runtime threat detection (eBPF) | GitOps (Helm) | Standard CNCF runtime security |
| Kyverno | Admission policy engine | GitOps (Helm) | Kubernetes-native policies |
| NetworkPolicies | East-west traffic control | GitOps (raw) | Deny-by-default in demo ns |

## Chaos

| Component | Purpose | Install | Why |
| --- | --- | --- | --- |
| chaos-mesh | Fault injection (network/pod/stress) | GitOps (Helm) | Rich CRDs, active CNCF project |

## Workloads

| Component | Purpose | Install | Source |
| --- | --- | --- | --- |
| api-a | Demo orders service (instrumented) | GitOps | this repo |
| api-b | Demo payments service (instrumented) | GitOps | this repo |
| fraudctl | Fraud-detection API (LB + 2 replicas) | GitOps | `github.com/fabianoflorentino/fraudctl` |
| certificate-validate | `k8s monitor` cert scanning + dashboard | GitOps | `github.com/fabianoflorentino/certificate-validate` |
| stracectl | Syscall tracing sidecar | GitOps | `github.com/fabianoflorentino/stracectl` |

## Docs & site

| Component | Purpose | Install | Why |
| --- | --- | --- | --- |
| Hugo site (`site/`) | Bilingual study site | GitHub Pages (CI) | Matches our other projects; teaches SRE/DevOps/Platform Eng |

## Our projects — how they plug in

### fraudctl

- Deployed as **load balancer + 2 API replicas** (mirrors its compose topology).
- Exposes `POST /fraud-score`; the LB distributes connections to workers via `SCM_RIGHTS`.
- It has no native Prometheus endpoint, so it is observed as a **black box**:
  - structured logs → Alloy → Loki;
  - `pprof` for profiling drills;
  - a Prometheus exporter sidecar (or the telemetry middleware) for RED metrics.
- Used as the *high-performance, opaque* workload for load + chaos drills.

### certificate-validate

- Runs `k8s monitor` as a DaemonSet scanning TLS Secrets and Ingresses cluster-wide.
- Exposes `certificate_days_left`, `certificate_expired`, `certificate_revoked` on `:9102`.
- ServiceMonitor → Prometheus; `PrometheusRule` alerts on `days_left < 15`.
- Webhook alerts → an in-cluster receiver (or Alertmanager).
- Dashboard served behind ingress with a cert-manager TLS certificate — proving the
  cert it validates is itself managed by cert-manager.

### stracectl

- Injected as a **sidecar** into api-a (or a dedicated demo pod) with
  `attach --serve :8080`, `shareProcessNamespace: true`, `CAP_SYS_PTRACE`.
- Exposes per-syscall latency stats + Prometheus metrics → dashboard.
- `srectl trace` wraps `stracectl discover` / `attach` for interactive debugging.
- Demonstrates *deep observability*: from syscall level up to traces.

## Why not ...

- **OPA/Gatekeeper instead of Kyverno:** Kyverno is Kubernetes-native and simpler to
  author for a study environment.
- **Fluent Bit instead of Alloy:** Alloy consolidates the log + OTel agent into one,
  reducing moving parts.
- **Elastic Stack instead of LGTM:** Grafana LGTM is lighter and ties metrics/logs/traces
  in one UI without a separate indexer cluster.
- **Terraform/Crossplane:** no external infrastructure to provision on kind.

## Version pinning

All Helm charts are pinned in `deploy/*/values.yaml` (via `Chart.yaml`/`Chart.lock` or
explicit `targetRevision`). Images are pinned to digests/tags (no `latest`) so the
environment is reproducible and Kyverno's `disallow-latest-tag` policy passes.

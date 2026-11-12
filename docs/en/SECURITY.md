# Security

The security pillars of srectl: scan before run, detect at runtime, enforce by
policy, and secure traffic in transit.

## Pillars

| Pillar | Tool | What it catches |
| --- | --- | --- |
| Supply chain / images | Trivy Operator | CVE + misconfig + CIS reports |
| Runtime detection | Falco | Suspicious syscalls/behavior (eBPF) |
| Policy enforcement | Kyverno | Admission-time violations |
| East-west control | NetworkPolicies | Unauthorized pod-to-pod traffic |
| TLS automation | cert-manager | Expired/mis-issued certificates |
| Cert monitoring | certificate-validate | Cert expiry/revocation alerts |

## Scan (Trivy Operator)

- Continuously scans every image in the cluster (including fraudctl, stracectl,
  certificate-validate and our demo apps).
- Produces `VulnerabilityReport` CRDs, exposed as metrics to Grafana.
- `srectl demo security` triggers a report and shows the dashboard.

## Runtime (Falco)

- eBPF probe, default + custom rules:
  - shell spawned in a container,
  - write below `/etc`, read of sensitive files (`/etc/shadow`),
  - unexpected outbound network from the demo namespace.
- Events stream to Loki (searchable) and Alertmanager (actionable).
- stracectl complements Falco for *ad-hoc* deep dives (`srectl trace`).

## Policy (Kyverno)

Baseline policies (deny unless allowed):

```yaml
require-requests-limits      # every container sets requests+limits
disallow-latest-tag          # no :latest images (forces pinned digests)
require-run-as-non-root      # no uid 0 (stracectl sidecar is an exception via namespace)
block-hostpath               # no hostPath volumes
require-labels               # app/owner labels for every resource
```

Policies are applied to the demo/apps namespaces; infra namespaces (`kube-system`,
`monitoring`, `argocd`, `cert-manager`) are exempt via Kyverno `matchExclude`/`Exclude`.

## Network policies

The `apps` namespace is **deny-by-default**:

- api-a → api-b (port 8080) allowed.
- ingress → api-a, api-b allowed.
- api-a → OTel collector, api-b → OTel collector allowed.
- everything else blocked (including fraudctl ↔ apps unless explicitly needed).

## TLS (cert-manager + certificate-validate)

- A selfsigned `ClusterIssuer` issues certs for all `*.localhost` ingress hosts.
- **certificate-validate** `k8s monitor` watches every TLS Secret + Ingress and alerts
  when any certificate drops below the expiry threshold — closing the loop on TLS.

## Secrets management

- No secrets are committed to GitOps.
- Credentials (Grafana admin, Argo CD admin, webhook URLs) live in Kubernetes Secrets
  managed out-of-band (generated at bootstrap, or via `srectl`).
- The demo uses `sops`/age-encrypted values only if/when needed — out of scope for v1.

## Demo / drills

```text
srectl demo security   # trivy report + kyverno violation + falco event, end to end
```

Each drill proves one pillar and is verified by `srectl demo verify`.

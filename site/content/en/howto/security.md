---
title: "Policy, scanning & runtime detection"
description: "Kyverno blocks a bad workload, Trivy reports CVEs and Falco catches an intrusion."
weight: 6
---

**SRE concept:** **defense in depth** — prevent (admission policy), detect pre-deploy
(image/config scanning) and detect at runtime (eBPF). Reliability includes not shipping known
bad artifacts.

**Prerequisites:** `srectl up`, `srectl demo verify`.

## 1. Prevent — Kyverno (policy as code)

```bash
kubectl run bad -n demo --image=nginx:latest
# denied by disallow-latest-tag (CEL ValidatingPolicy)
```

The `demo` namespace also enforces requests/limits and no `hostPath`; `runAsNonRoot` is audited.

## 2. Detect pre-deploy — Trivy Operator

```bash
kubectl get vulnerabilityreport,configauditreport -n demo
kubectl get vulnerabilityreport -n demo -o jsonpath='{.items[0].report.summary}'
```

Findings also land in `srectl · Security` (severity breakdown, top images, config/RBAC issues).

## 3. Detect at runtime — Falco

```bash
kubectl exec -n demo deploy/hello -- cat /etc/shadow
```

Falco (modern eBPF) fires **"Read sensitive file untrusted"**. See it in Loki:
`{namespace="demo"} |= "shadow"`.

## Why it matters (SRE)

Each layer catches what the previous one misses: policy stops the obvious mistake, scanning
finds known CVEs before they run, runtime detection catches behaviour that no static check
could predict. Together they shrink the window between "bad thing happens" and "we know".

## Expected outcome

A rejected deployment, populated Trivy reports, and a Falco event correlated in Loki/Grafana.

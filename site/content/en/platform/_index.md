---
title: "Platform Engineering"
description: "Self-service, policy as code and observability as a platform."
---

A platform team builds the *paved road* so product teams ship safely.

## Self-service

`srectl` is the self-service entrypoint: bring the platform up, open a service, run a
drill. Everything it does is a thin wrapper over `kind`/`kubectl`/`helm` that you can
inspect and extend.

## Policy as code

Kyverno ships baseline policies (CEL `ValidatingPolicy`, scoped to `demo`):

- disallow the `latest` image tag;
- require CPU/memory requests and limits;
- require `runAsNonRoot` (Audit);
- block `hostPath` volumes.

```bash
kubectl run bad -n demo --image=nginx:latest   # rejected
```

## Observability as a platform

Dashboards are GitOps artifacts (tag `sre`): `App RED`, `Golden Signals`, `Cluster`,
`Logs`, `Security`, `Certificates`, `Syscall`. Security findings (Trivy, Falco) and cert
health flow into the same Grafana.

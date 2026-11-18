---
title: "Runbooks"
description: "What each top alert means, how to confirm and how to mitigate."
---

The full runbooks live in the repository (`docs/en/runbooks/`). Summaries:

## ApiAHighErrorRate

`slo:api_a_availability:ratio5m < 0.95` for 1m.

1. Check `srectl · App RED` (5xx) and `srectl · Logs` `{app="api-a"} |= "error"`.
2. api-a calls api-b on `/orders` — check `{app="api-b"}` and its pods.
3. Trace a failing request in Tempo (service `api-a`).
4. Mitigate: restart/scale api-b, or roll back with `srectl sync`.

## ApiAHighLatency

`slo:api_a_latency_p95:5m > 20ms` for 30s. Check `srectl · Cluster` for CPU saturation,
`srectl · Syscall` if stracectl is attached, and the upstream api-b.

## CertificateExpiringSoon / CertificateExpired

`certificate_days_left < 15` / `<= 0`. Open `srectl · Certificates`, inspect with
`kubectl get certificate -A`, and rely on cert-manager renewal (CA issuer `srectl-ca` in `demo`).

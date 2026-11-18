# Runbooks

Operational runbooks for the top srectl alerts. Each one answers: what it means,
how to confirm, and how to mitigate.

## ApiAHighErrorRate

**Alert:** `slo:api_a_availability:ratio5m < 0.95` for 1m (warning).

1. Confirm the scope: Grafana → `srectl · App RED` (5xx ratio) and
   `srectl · Logs` → `{namespace="demo", app="api-a"} |= "error"`.
2. Check the upstream: `kubectl get pods -n demo -l app=api-b` and
   `srectl · Logs` `{app="api-b"}`. api-a calls api-b on `/orders`.
3. Follow a trace: Grafana → Explore → Tempo (service `api-a`, `/orders`).
4. Mitigate: if api-b is down, `kubectl rollout restart deploy/api-b -n demo`
   (or let Argo CD self-heal). If a bad deploy, `srectl sync api-b`/rollback.

## ApiAHighLatency

**Alert:** `slo:api_a_latency_p95:5m > 20ms` for 30s (warning).

1. Confirm on `srectl · App RED` (latency p50/p95/p99) and `srectl · Cluster`
   (CPU saturation, throttling).
2. Common causes: CPU saturation (chaos stress), api-b slowdown, or a slow
   dependency. Check `srectl · Syscall` if stracectl is attached.
3. Mitigate: scale api-a (`kubectl scale deploy/api-a -n demo --replicas=3`) or
   relieve the stressor; if upstream, restart/scale api-b.

## CertificateExpiringSoon / CertificateExpired

**Alert:** `certificate_days_left < 15` (warning) / `<= 0` (critical).

1. `srectl open` → Grafana `srectl · Certificates` to find the expiring secret.
2. Inspect: `kubectl get certificate -A` and
   `kubectl describe certificate -n <ns> <name>`.
3. cert-manager renews automatically; if it is stuck, check
   `kubectl get certificaterequest -A` and the `cert-manager` logs.
4. Force a renewal: annotate the Certificate with
   `cert-manager.io/issue-temporary-certificate` or delete the pending
   CertificateRequest. The CA Issuer is `srectl-ca` in namespace `demo`.

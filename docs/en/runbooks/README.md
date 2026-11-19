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

## ErrorBudgetBurn (slo-generator)

**Signal:** `error_budget_burn_rate{service_name,slo_name}` above
`alerting_burn_rate_threshold` (1h page at 14.4x, 6h page at 6x, 3d freeze at 1x).

1. Confirm the scope: `srectl slo status` (or the Grafana dashboard
   **srectl · SLO / Error Budget**) to see which service/SLO is burning.
2. Confirm it is real, not a query bug: open `srectl · App RED`/`srectl · Golden
   Signals` and check that the SLI itself dropped (5xx or latency).
3. Check the policy step that fired — 1h/6h means **page**; 3d means **freeze
   releases** unless reliability/security related.
4. Mitigate like any SLO incident: follow the `ApiAHighErrorRate` /
   `ApiAHighLatency` runbook, or `kubectl rollout undo` the suspect deploy.
5. Recompute and re-check: `srectl slo compute && srectl slo status`. The burn
   rate must fall below the threshold before closing the incident.

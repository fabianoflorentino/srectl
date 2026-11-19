# Runbooks

Runbooks operacionais para os principais alertas do srectl. Cada um responde: o
que significa, como confirmar e como mitigar.

## ApiAHighErrorRate

**Alerta:** `slo:api_a_availability:ratio5m < 0.95` por 1m (warning).

1. Confirme o escopo: Grafana → `srectl · App RED` (razão 5xx) e
   `srectl · Logs` → `{namespace="demo", app="api-a"} |= "error"`.
2. Verifique o upstream: `kubectl get pods -n demo -l app=api-b` e
   `srectl · Logs` `{app="api-b"}`. A api-a chama a api-b em `/orders`.
3. Siga um trace: Grafana → Explore → Tempo (serviço `api-a`, `/orders`).
4. Mitigue: se a api-b caiu, `kubectl rollout restart deploy/api-b -n demo`
   (ou deixe o Argo CD self-heal). Se for deploy ruim, `srectl sync api-b`/rollback.

## ApiAHighLatency

**Alerta:** `slo:api_a_latency_p95:5m > 20ms` por 30s (warning).

1. Confirme no `srectl · App RED` (latência p50/p95/p99) e no `srectl · Cluster`
   (saturação de CPU, throttling).
2. Causas comuns: saturação de CPU (stress de chaos), api-b lenta ou dependência
   lenta. Veja `srectl · Syscall` se o stracectl estiver anexado.
3. Mitigue: escale a api-a (`kubectl scale deploy/api-a -n demo --replicas=3`) ou
   remova o stress; se for upstream, reinicie/escale a api-b.

## CertificateExpiringSoon / CertificateExpired

**Alerta:** `certificate_days_left < 15` (warning) / `<= 0` (critical).

1. `srectl open` → Grafana `srectl · Certificates` para achar o secret expirando.
2. Inspecione: `kubectl get certificate -A` e
   `kubectl describe certificate -n <ns> <name>`.
3. O cert-manager renova sozinho; se travou, veja
   `kubectl get certificaterequest -A` e os logs do `cert-manager`.
4. Force renovação: anote o Certificate com
   `cert-manager.io/issue-temporary-certificate` ou apague o CertificateRequest
   pendente. O Issuer CA é o `srectl-ca` no namespace `demo`.

## ErrorBudgetBurn (slo-generator)

**Sinal:** `error_budget_burn_rate{service_name,slo_name}` acima de
`alerting_burn_rate_threshold` (page de 1h em 14,4x, page de 6h em 6x, freeze de
3d em 1x).

1. Confirme o escopo: `srectl slo status` (ou o dashboard
   **srectl · SLO / Error Budget**) para ver qual serviço/SLO está queimando.
2. Confirme que é real, não bug de consulta: abra `srectl · App RED` /
   `srectl · Golden Signals` e veja se a SLI caiu (5xx ou latência).
3. Veja qual passo da política disparou — 1h/6h significa **page**; 3d significa
   **congelar releases** exceto por confiabilidade/segurança.
4. Mitigue como qualquer incidente de SLO: siga o runbook de
   `ApiAHighErrorRate` / `ApiAHighLatency`, ou `kubectl rollout undo` do deploy
   suspeito.
5. Recalcule e confira: `srectl slo compute && srectl slo status`. O burn rate
   deve cair abaixo do limiar antes de encerrar o incidente.

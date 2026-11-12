# Observabilidade

Como métricas, logs e traces fluem pelo srectl e como os transformamos em SLOs e
alertas acionáveis.

## Os três sinais

| Sinal | Pipeline | Backend | Consulta |
| --- | --- | --- | --- |
| Métricas | app `/metrics` → Prometheus (scrape via ServiceMonitor) | Prometheus | PromQL |
| Logs | app `slog` stdout → Alloy → Loki | Loki | LogQL |
| Traces | SDK OTel → OTel Collector → Tempo | Tempo | TraceQL |

O Grafana é provisionado com os três datasources para que uma única consulta possa
navegar entre sinais (trace → logs → métricas).

## Contrato de instrumentação (APIs demo)

Todo workload nosso emite os sinais do método **RED** mais alguns extras:

```text
http_requests_total{service, method, route, status}
http_request_duration_seconds{service, method, route, status}  # histograma
http_errors_total{service, route, status}
process_cpu_seconds_total
process_resident_memory_bytes
```

- **api-a** e **api-b** usam o OpenTelemetry Go SDK + `prometheus/client_golang`.
- Logs estruturados com `log/slog` (JSON no stdout) e campos `trace_id`/`span_id` para
  correlação log→trace.
- Toda requisição de entrada é um span, e a chamada da api-a para a api-b propaga o
  contexto de trace (W3C `traceparent`).

## Métricas dos nossos projetos

- **certificate-validate** (`k8s monitor`, `:9102`):
  `certificate_days_left{namespace,name,kind}`, `certificate_expired`,
  `certificate_revoked`.
- **stracectl** (`--serve :8080`): contagens de chamadas por syscall, taxas de erro,
  histogramas de latência P50/P95/P99.
- **fraudctl**: observado via sidecar exporter / logs de telemetria (sem `/metrics` nativo).

## Dashboards essenciais

1. **Visão do cluster** — CPU/mem/disco dos nós, contagem de pods, restarts, saúde do
   control-plane.
2. **RED das apps** — taxa de requisições, taxa de erro, percentis de latência por
   serviço/rota.
3. **Explorador de trace distribuído** — busca no Tempo + service graph.
4. **Syscall (stracectl)** — top syscalls, percentis de latência, anomalias.
5. **Certificados (certificate-validate)** — dias restantes por namespace, lista de
   vencimento próximo.
6. **Postura de segurança** — vulnerabilidades do Trivy, violações do Kyverno, eventos do
   Falco.

## Alertas

As `PrometheusRule`s (via Operator) roteiam para o Alertmanager, que distribui para o
Grafana + receivers webhook. O Falco também emite eventos para o Alertmanager.

Regras baseline:

```text
HighErrorRate       http_errors_total / http_requests_total > 0.05  (5m)
HighLatencyP99      histogram_quantile(0.99, ...) > 500ms          (5m)
CertExpiring        certificate_days_left < 15
CertRevoked         certificate_revoked == 1
NodePressure        pressão de node_filesystem/node_memory
PodCrashLooping     kube_pod_container_status_waiting reason=CrashLoopBackOff
SyscallAnomaly      pico de taxa de erro de syscall (stracectl)
```

## Definições de SLO / SLI

| Serviço | SLI | SLO |
| --- | --- | --- |
| api-a `/orders` | disponibilidade = good/(good+bad) em 5m | 99,5% (mensal) |
| api-a `/orders` | latência p99 | < 300ms |
| fraudctl `/fraud-score` | disponibilidade + latência p99 | 99,5% / < 20ms |
| Certs TLS | dias restantes > 0 | 100% (sem expiração) |

Recording rules pré-calculam as séries de SLI; dashboards de SLO no Grafana mostram o
burn-rate.

## Loop caos → observabilidade

O chaos-mesh injeta falhas; a stack de observabilidade é como as *detectamos*:

1. Injete 200ms de latência na api-a (`NetworkChaos`).
2. Veja o p99 subir além do limiar do SLO no dashboard RED.
3. Confirme que `HighLatencyP99` dispara no Alertmanager.
4. Mate os pods do fraudctl; confirme que o self-healing restaura a disponibilidade
   dentro do orçamento.

## Acesso local

| Superfície | Comando |
| --- | --- |
| Grafana | `srectl open grafana` (admin padrão do secret) |
| Prometheus | `srectl open prometheus` |
| Alertmanager | `srectl open alertmanager` |
| Métricas da app | `srectl open api-a` → `/metrics` |

Tudo também é acessível via ingress (`grafana.localhost`, `prometheus.localhost`, ...)
assim que o ingress + cert-manager estiverem de pé.

#!/usr/bin/env bash
# End-to-end verification of a running srectl platform: Argo CD apps, pods,
# ingress endpoints, in-cluster services, metrics, SLOs and (optionally) chaos.
#
# Usage: scripts/e2e-verify.sh
# Environment:
#   SRECTL_BIN     srectl binary (default: ./bin/srectl)
#   E2E_TIMEOUT    seconds to wait for pods/metrics (default: 600)
#   E2E_SKIP_CHAOS set to "true" to skip the chaos drill
set -euo pipefail

BIN="${SRECTL_BIN:-./bin/srectl}"
TIMEOUT="${E2E_TIMEOUT:-600}"
SKIP_CHAOS="${E2E_SKIP_CHAOS:-false}"

log() { printf '\n==> %s\n' "$*"; }
fail() {
  printf '::error::%s\n' "$*" >&2
  exit 1
}

http_code() {
  curl -ksS -o /dev/null -w '%{http_code}' --max-time 10 "$1" || true
}

check_http() {
  local url="$1" want="${2:-200}" code=""
  for _ in $(seq 1 40); do
    code="$(http_code "$url")"
    if [ "$code" = "$want" ]; then
      printf 'ok    %-55s %s\n' "$url" "$code"
      return 0
    fi
    sleep 3
  done
  fail "GET $url returned ${code:-no response}, want $want"
}

check_proxy() {
  local path="$1"
  for _ in $(seq 1 40); do
    if kubectl get --raw "$path" >/dev/null 2>&1; then
      printf 'ok    service proxy %s\n' "$path"
      return 0
    fi
    sleep 3
  done
  fail "service proxy check failed: $path"
}

prom_query() {
  local query="$1"
  kubectl -n monitoring get --raw \
    "/api/v1/namespaces/monitoring/services/http:kube-prometheus-stack-prometheus:9090/proxy/api/v1/query?query=$(jq -rn --arg q "$query" '$q|@uri')"
}

check_metric() {
  local name="$1" query="$2"
  for _ in $(seq 1 40); do
    if prom_query "$query" | jq -e '.data.result | length > 0' >/dev/null 2>&1; then
      printf 'ok    metric %s\n' "$name"
      return 0
    fi
    sleep 3
  done
  fail "metric not found in Prometheus: $name ($query)"
}

check_pods() {
  local deadline=$((SECONDS + TIMEOUT)) bad=""
  while ((SECONDS < deadline)); do
    bad="$(kubectl get pods -A -o json | jq -r '
      [ .items[]
        | select(.metadata.deletionTimestamp == null)
        | select(.metadata.labels.app != "slo-generator")
        | select(
            .status.phase == "Failed" or
            .status.phase == "Pending" or
            (.status.phase == "Running" and
              ([.status.containerStatuses[]? | select(.ready != true)] | length > 0))
          )
        | "\(.metadata.namespace)/\(.metadata.name) phase=\(.status.phase)"
      ] | .[]')"
    if [ -z "$bad" ]; then
      printf 'ok    all pods Ready\n'
      return 0
    fi
    printf '      waiting for pods:\n%s\n' "$bad"
    sleep 10
  done
  fail "pods did not become Ready within ${TIMEOUT}s"
}

log "Argo CD applications"
"$BIN" demo verify || fail "Argo CD applications are not Synced/Healthy"

log "Pods in every namespace"
check_pods

log "Ingress endpoints"
check_http https://hello.localhost/ 200
check_http https://api-a.localhost/health 200
check_http https://api-a.localhost/ready 200
check_http https://api-a.localhost/orders 200
check_http http://fraudctl.localhost/ready 200
check_http http://seraph.localhost/ 200

fraud_body='{"amount":100,"installments":1,"requested_at":"2026-01-01T10:00:00Z","customer":{"avg_amount":100,"tx_count_24h":5,"known_merchants":["m1"]},"merchant":{"id":"m1","mcc":"5411","avg_amount":50},"terminal":{"is_online":false,"card_present":true,"km_from_home":10},"id":"e2e-tx-1"}'
fraud_code="$(curl -ksS -o /tmp/srectl-fraud.json -w '%{http_code}' --max-time 10 \
  -X POST -H 'Content-Type: application/json' --data "$fraud_body" \
  http://fraudctl.localhost/fraud-score || true)"
[ "$fraud_code" = "200" ] || fail "POST /fraud-score returned ${fraud_code:-no response}"
grep -q approved /tmp/srectl-fraud.json || fail "POST /fraud-score response has no verdict"
printf 'ok    %-55s %s\n' "http://fraudctl.localhost/fraud-score" "$fraud_code"

log "In-cluster services"
check_proxy /api/v1/namespaces/monitoring/services/http:kube-prometheus-stack-prometheus:9090/proxy/-/ready
check_proxy /api/v1/namespaces/monitoring/services/http:kube-prometheus-stack-alertmanager:9093/proxy/-/ready
check_proxy /api/v1/namespaces/monitoring/services/http:kube-prometheus-stack-grafana:80/proxy/api/health
check_proxy /api/v1/namespaces/monitoring/services/http:loki:3100/proxy/ready
check_proxy /api/v1/namespaces/monitoring/services/http:tempo:3200/proxy/ready
check_proxy /api/v1/namespaces/demo/services/http:api-b:80/proxy/health
check_proxy /api/v1/namespaces/projects/services/http:stracectl:8080/proxy/healthz
check_proxy /api/v1/namespaces/projects/services/http:cert-validate-monitor:9102/proxy/metrics

kubectl get --raw \
  /api/v1/namespaces/projects/services/http:cert-validate-monitor:9102/proxy/metrics \
  | grep -q certificate_ || fail "certificate-validate exposes no certificate_* metric"
printf 'ok    certificate-validate metrics\n'

log "Metrics in Prometheus"
check_metric "api-a requests" 'count(api_a_http_requests_total)'
check_metric "api-b requests" 'count(api_b_http_requests_total)'
check_metric "ingress requests" 'count(nginx_ingress_controller_requests)'
check_metric "certificate_days_left" 'count(certificate_days_left)'
check_metric "stracectl metrics" 'count({__name__=~"stracectl_.+"})'

log "SLO computation (slo-generator)"
"$BIN" slo compute || fail "srectl slo compute failed"
for _ in $(seq 1 40); do
  if prom_query 'count(error_budget_burn_rate)' | jq -e '.data.result | length > 0' >/dev/null 2>&1; then
    break
  fi
  sleep 3
done
"$BIN" slo status || fail "srectl slo status failed"
"$BIN" slo status | grep -q 'api-a' || fail "slo status contains no api-a row"
printf 'ok    SLO metrics exported\n'

if [ "$SKIP_CHAOS" != "true" ]; then
  log "Chaos smoke test (pod-kill fraudctl)"
  "$BIN" demo chaos pod-kill --duration 20s || fail "chaos drill failed"
  kubectl -n projects rollout status deployment/fraudctl --timeout=180s \
    || fail "fraudctl did not recover after pod-kill"
  check_http http://fraudctl.localhost/ready 200
fi

printf '\nE2E OK: cluster, services, metrics and SLOs are healthy\n'

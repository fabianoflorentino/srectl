---
title: "Runbooks"
description: "O que cada alerta principal significa, como confirmar e como mitigar."
---

Os runbooks completos ficam no repositório (`docs/pt-br/runbooks/`). Resumos:

## ApiAHighErrorRate

`slo:api_a_availability:ratio5m < 0.95` por 1m.

1. Veja `srectl · App RED` (5xx) e `srectl · Logs` `{app="api-a"} |= "error"`.
2. A api-a chama a api-b em `/orders` — confira `{app="api-b"}` e os pods dela.
3. Siga uma requisição com falha no Tempo (serviço `api-a`).
4. Mitigue: reinicie/escale a api-b, ou faça rollback com `srectl sync`.

## ApiAHighLatency

`slo:api_a_latency_p95:5m > 20ms` por 30s. Veja `srectl · Cluster` (saturação de CPU),
`srectl · Syscall` se o stracectl estiver anexado, e a api-b upstream.

## CertificateExpiringSoon / CertificateExpired

`certificate_days_left < 15` / `<= 0`. Abra `srectl · Certificates`, inspecione com
`kubectl get certificate -A`, e conte com a renovação do cert-manager (Issuer CA `srectl-ca` no `demo`).

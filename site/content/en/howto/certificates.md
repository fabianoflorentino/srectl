---
title: "Certificate lifecycle"
description: "Watch TLS certificates' days-left and the expiry alert."
weight: 7
---

**SRE concept:** **TLS hygiene as a leading indicator.** Certificate expiry is a classic
self-inflicted outage; days-left is a metric you alert on *long before* it becomes a pager at
03:00.

**Prerequisites:** `srectl up`, `srectl demo verify`.

## Run it

```bash
srectl open grafana        # → srectl · Certificates
kubectl get certificate -A
```

`certificate-validate` runs `k8s monitor` as a DaemonSet, scanning every TLS Secret and
Ingress cluster-wide and exposing `certificate_days_left` on `:9102`.

## What you'll see

- The `srectl · Certificates` dashboard: days left per certificate (`namespace/name`), plus
  expiring/expired counters.
- The SLO-style metric `certificate_days_left{namespace,kind,name}` in Prometheus.
- Alerts `CertificateExpiringSoon` (`< 15d`) and `CertificateExpired` (`<= 0`).

## Why it matters (SRE)

Expiry is deterministic and predictable — the textbook **leading indicator**. cert-manager
already rotates automatically (CA issuer `srectl-ca` in `demo`); the monitor verifies that it
actually did, and the alert fires with enough lead time to act. This is "measure the thing that
kills you, before it does".

## Expected outcome

Green days-left on all certificates, with an alerting path for anything under 15 days.

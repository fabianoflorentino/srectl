---
title: "Health verification"
description: "Assert the cluster, Argo CD and every app are healthy in one command."
weight: 4
---

**SRE concept:** **service health checks / smoke tests.** Before and after every drill you
need a fast, deterministic verdict: is the platform in the state we expect?

**Prerequisites:** `srectl up`.

## Run it

```bash
srectl demo verify
srectl status
```

`demo verify` checks: the kind cluster is up, Argo CD is `Ready`, and **every** Argo CD app is
`Synced · Healthy`. `status` prints the full matrix.

## What you'll see

```text
verify: OK — cluster up, argocd ready, 18 apps Synced · Healthy
```

If something drifted, verify fails and prints the offending apps, e.g. `api-b: OutOfSync · Degraded`.

## Why it matters (SRE)

Health checks are the cheapest incident detector and the gate for every change. Making them a
single command (not a runbook of `kubectl` incantations) means humans and CI assert the *same*
contract — which is exactly what the E2E workflow does (`up → verify → down`).

## Expected outcome

A green verify before you start a drill, and again after recovery.

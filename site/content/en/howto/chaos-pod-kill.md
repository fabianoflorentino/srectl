---
title: "Pod kill & self-healing"
description: "Kill a fraudctl pod and watch Kubernetes and GitOps restore it."
weight: 3
---

**SRE concept:** **self-healing and blast radius.** A resilient platform tolerates the loss
of a pod without manual intervention. Mean time to recovery (MTTR) is a property of the
system, not of a hero with a keyboard.

**Prerequisites:** `srectl up`, `srectl demo verify`.

## Run it

```bash
srectl demo chaos pod-kill --duration 30s
kubectl get pods -n projects -l app=fraudctl -w
```

The PodChaos kills one `fraudctl` pod; the Deployment recreates it.

## What you'll see

- The old pod disappears; a new one is created with a new name.
- The `fraudctl` deployment returns to `3/3` (its LB container restarts briefly until the
  worker sockets exist, then settles).
- Argo CD keeps the Deployment's desired state (replicas, image, probes) — GitOps self-heal.

## Why it matters (SRE)

Replicas + controllers turn "a pod died" from an incident into a non-event. The drill makes
the recovery observable: you measure how long the service is degraded and whether any request
failed — the **blast radius**. That is the input to your availability SLO.

## Expected outcome

The pod is replaced and the service is healthy again within the SLO, with no human action.

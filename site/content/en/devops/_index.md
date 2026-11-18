---
title: "DevOps"
description: "GitOps with Argo CD, image pipelines and CI — build, ship and roll back."
---

## GitOps

Argo CD reconciles everything from this repository. The root app points at
`deploy/bootstrap/apps/`; touch a file, `srectl sync`, and the cluster converges. Drift is
corrected by `selfHeal`, and removed resources are pruned.

```bash
srectl sync            # hard refresh of every app
srectl status          # Synced · Healthy matrix
```

## Images

Demo services ship multi-stage Dockerfiles (non-root, distroless/alpine) and are loaded
into kind for local iteration:

```bash
docker build -f apps/api-a/Dockerfile -t srectl/api-a:dev .
kind load docker-image srectl/api-a:dev --name srectl
```

Images are pinned by tag — **never `latest`** — and Kyverno enforces it in the `demo`
namespace.

## CI

GitHub Actions compiles, lints, tests (≥80% coverage) and builds images; the site is
published to GitHub Pages. Argo CD — not CI — is the only deployer.

---
title: "DevOps"
description: "GitOps com Argo CD, pipelines de imagem e CI — build, deploy e rollback."
---

## GitOps

O Argo CD reconcilia tudo deste repositório. O app raiz aponta para
`deploy/bootstrap/apps/`; edite um arquivo, rode `srectl sync`, e o cluster converge. O
drift é corrigido pelo `selfHeal` e recursos removidos são podados.

```bash
srectl sync            # refresh forte de todos os apps
srectl status          # matriz Synced · Healthy
```

## Imagens

Os serviços demo têm Dockerfiles multi-stage (non-root, distroless/alpine) e são
carregados no kind para iteração local:

```bash
docker build -f apps/api-a/Dockerfile -t srectl/api-a:dev .
kind load docker-image srectl/api-a:dev --name srectl
```

Imagens são fixadas por tag — **nunca `latest`** — e o Kyverno impõe isso no namespace `demo`.

## CI

O GitHub Actions compila, faz lint, testa (≥80% de cobertura) e builda imagens; o site é
publicado no GitHub Pages. O Argo CD — não a CI — é o único deployer.

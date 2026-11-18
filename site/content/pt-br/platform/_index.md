---
title: "Plataforma"
description: "Self-service, política como código e observabilidade como plataforma."
---

Um time de plataforma constrói a *estrada pavimentada* para os times de produto entregarem com segurança.

## Self-service

O `srectl` é o ponto de entrada self-service: sobe a plataforma, abre um serviço, roda um
drill. Tudo é um wrapper fino sobre `kind`/`kubectl`/`helm` que você pode inspecionar e estender.

## Política como código

O Kyverno aplica políticas baseline (CEL `ValidatingPolicy`, escopo `demo`):

- proibir a tag `latest`;
- exigir requests e limits de CPU/memória;
- exigir `runAsNonRoot` (Audit);
- bloquear volumes `hostPath`.

```bash
kubectl run bad -n demo --image=nginx:latest   # rejeitado
```

## Observabilidade como plataforma

Os dashboards são artefatos GitOps (tag `sre`): `App RED`, `Golden Signals`, `Cluster`,
`Logs`, `Security`, `Certificates`, `Syscall`. Achados de segurança (Trivy, Falco) e a
saúde dos certificados chegam no mesmo Grafana.

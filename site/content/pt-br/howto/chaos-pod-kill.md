---
title: "Pod kill e self-healing"
description: "Mate um pod do fraudctl e veja o Kubernetes e o GitOps restaurarem."
weight: 3
---

**Conceito de SRE:** **self-healing e blast radius.** Uma plataforma resiliente tolera a perda
de um pod sem intervenção manual. O MTTR (tempo médio de recuperação) é propriedade do sistema,
não de um herói no teclado.

**Pré-requisitos:** `srectl up`, `srectl demo verify`.

## Rodar

```bash
srectl demo chaos pod-kill --duration 30s
kubectl get pods -n projects -l app=fraudctl -w
```

O PodChaos mata um pod do `fraudctl`; o Deployment recria.

## O que você vê

- O pod antigo some; um novo é criado com outro nome.
- O `fraudctl` volta a `3/3` (o container do LB reinicia brevemente até os sockets dos workers
  existirem, depois estabiliza).
- O Argo CD mantém o estado desejado do Deployment (réplicas, imagem, probes) — self-heal GitOps.

## Por que importa (SRE)

Réplicas + controllers transformam "um pod morreu" de incidente em não-evento. O drill torna a
recuperação observável: você mede quanto tempo o serviço ficou degradado e se alguma requisição
falhou — o **blast radius**. Esse é o insumo do seu SLO de disponibilidade.

## Resultado esperado

O pod é substituído e o serviço volta saudável dentro do SLO, sem ação humana.

---
title: "Ciclo de vida de certificados"
description: "Acompanhe os dias-restantes dos certificados TLS e o alerta de expiração."
weight: 7
---

**Conceito de SRE:** **higiene de TLS como indicador antecedente.** Expiração de certificado é
um clássico incidente autoinfligido; dias-restantes é uma métrica que você alerta *muito antes*
de virar páginas às 3h da manhã.

**Pré-requisitos:** `srectl up`, `srectl demo verify`.

## Rodar

```bash
srectl open grafana        # → srectl · Certificates
kubectl get certificate -A
```

O `certificate-validate` roda `k8s monitor` como DaemonSet, varrendo todo Secret e Ingress TLS
do cluster e expondo `certificate_days_left` em `:9102`.

## O que você vê

- O dashboard `srectl · Certificates`: dias restantes por certificado (`namespace/name`) e
  contadores de expirando/expirado.
- A métrica `certificate_days_left{namespace,kind,name}` no Prometheus.
- Os alertas `CertificateExpiringSoon` (`< 15d`) e `CertificateExpired` (`<= 0`).

## Por que importa (SRE)

Expiração é determinística e previsível — o **indicador antecedente** por excelência. O
cert-manager já rotaciona sozinho (Issuer CA `srectl-ca` no `demo`); o monitor verifica se de
fato rotacionou, e o alerta dispara com antecedência suficiente para agir. É "medir o que te
derruba, antes de te derrubar".

## Resultado esperado

Dias-restantes verdes em todos os certificados, com caminho de alerta para qualquer um abaixo de 15 dias.

# Segurança

Os pilares de segurança do srectl: escanear antes de rodar, detectar em runtime,
impor por política e proteger o tráfego em trânsito.

## Pilares

| Pilar | Ferramenta | O que detecta |
| --- | --- | --- |
| Supply chain / imagens | Trivy Operator | CVEs + misconfig + reports CIS |
| Detecção em runtime | Falco | Syscalls/comportamentos suspeitos (eBPF) |
| Enforcement de política | Kyverno | Violações na admissão |
| Controle leste-oeste | NetworkPolicies | Tráfego pod-a-pod não autorizado |
| Automação de TLS | cert-manager | Certificados expirados/mal-emitidos |
| Monitoramento de certs | certificate-validate | Alertas de expiração/revogação |

## Scan (Trivy Operator)

- Escaneia continuamente toda imagem do cluster (incluindo fraudctl, stracectl,
  certificate-validate e as apps demo).
- Produz CRDs `VulnerabilityReport`, expostos como métricas ao Grafana.
- `srectl demo security` dispara um report e mostra o dashboard.

## Runtime (Falco)

- Probe eBPF, regras padrão + customizadas:
  - shell iniciado em um container,
  - escrita abaixo de `/etc`, leitura de arquivos sensíveis (`/etc/shadow`),
  - rede de saída inesperada a partir do namespace demo.
- Eventos vão para o Loki (buscáveis) e Alertmanager (acionáveis).
- O stracectl complementa o Falco em investigações *ad-hoc* (`srectl trace`).

## Política (Kyverno)

Políticas baseline (nega a menos que permitido):

```yaml
require-requests-limits      # todo container define requests+limits
disallow-latest-tag          # sem imagens :latest (força digests pinados)
require-run-as-non-root      # sem uid 0 (sidecar stracectl é exceção via namespace)
block-hostpath               # sem volumes hostPath
require-labels               # labels app/owner em todo recurso
```

As políticas se aplicam aos namespaces demo/apps; namespaces de infra (`kube-system`,
`monitoring`, `argocd`, `cert-manager`) são isentos via `matchExclude`/`Exclude` do
Kyverno.

## Network policies

O namespace `apps` é **deny-by-default**:

- api-a → api-b (porta 8080) permitido.
- ingress → api-a, api-b permitido.
- api-a → OTel collector, api-b → OTel collector permitido.
- todo o resto bloqueado (incluindo fraudctl ↔ apps, salvo necessidade explícita).

## TLS (cert-manager + certificate-validate)

- Um `ClusterIssuer` selfsigned emite certs para todos os hosts `*.localhost` do ingress.
- O **certificate-validate** `k8s monitor` observa todo Secret + Ingress TLS e alerta
  quando qualquer certificado cai abaixo do limiar de expiração — fechando o ciclo do TLS.

## Gestão de secrets

- Nenhum secret é commitado no GitOps.
- Credenciais (admin do Grafana, admin do Argo CD, URLs de webhook) vivem em Secrets do
  Kubernetes gerenciados fora da banda (gerados no bootstrap, ou via `srectl`).
- A demo usa valores criptografados com `sops`/age apenas se/quando necessário — fora do
  escopo da v1.

## Demo / drills

```text
srectl demo security   # report do trivy + violação do kyverno + evento do falco, ponta a ponta
```

Cada drill prova um pilar e é verificado por `srectl demo verify`.

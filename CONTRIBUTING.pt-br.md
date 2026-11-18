# Contribuindo com o srectl

_Read this in [English](CONTRIBUTING.md)._

Obrigado por ajudar! O `srectl` é um task runner que sobe uma plataforma completa de
treino em SRE / DevOps / Platform Engineering no kind, dirigida por GitOps.

## Pré-requisitos

- Go **1.27+**, Docker, [kind](https://kind.sigs.k8s.io/), `kubectl`, `helm`
- [`golangci-lint`](https://golangci-lint.run/) v2.x
- Opcional: [Hugo](https://gohugo.io/) (para o site)

## Início rápido

```bash
make build                 # gera ./bin/srectl
./bin/srectl up            # cluster + Argo CD + todos os apps GitOps
./bin/srectl status        # matriz Synced · Healthy
./bin/srectl demo verify   # valida a saúde da stack inteira
```

## Fluxo de desenvolvimento

```bash
make            # ajuda
make test       # go test -race
make coverage   # exige >= 80% de cobertura de statements
make lint       # golangci-lint
make fmt        # gofmt + golangci-lint fmt
make site-dev   # site de documentação local
```

Regras de Go deste repositório:

- **Cobertura deve ficar >= 80%** (`make coverage` falha abaixo disso).
- Testes são especificações executáveis — valide comportamento público, não detalhes internos.
- Siga as skills de Go referenciadas no `AGENTS.md`.

## Commits

[Conventional Commits](https://www.conventionalcommits.org/), em inglês, ex.:
`feat(demo): add chaos latency scenario`, `fix(observability): honorLabels`.

## Adicionando um app GitOps

1. Crie `deploy/bootstrap/apps/<name>-app.yaml` (fonte Helm ou um path do repositório).
2. `git commit && git push`.
3. `kubectl annotate application root-app -n argocd argocd.argoproj.io/refresh=hard --overwrite`
   (ou `./bin/srectl sync`).
4. Confirme com `./bin/srectl status`.

Fixe tags de imagem — **nunca `latest`**; a CI rejeita em `deploy/`.

## Testes

- Unitários: `make test` (race) e `make coverage`.
- Integração: `make test-integration` (precisa de um cluster no ar).
- E2E (CI): `.github/workflows/e2e.yml` — `srectl up` → `demo verify` → `down`.

## Checklist de PR

- [ ] Testes e lint verdes (`make test lint coverage`).
- [ ] Comportamento novo documentado em `docs/en/*` **e** `docs/pt-br/*`.
- [ ] Mensagem de commit convencional (inglês).
- [ ] Sem tags `latest` ou segredos commitados.

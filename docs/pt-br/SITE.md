# Site de Documentação (GitHub Pages)

O projeto publica um site bilíngue (Hugo) no **GitHub Pages** que é a superfície
principal para *aprender* SRE / DevOps / Platform Engineering por meio do srectl.

## Propósito

O site transforma a plataforma em execução em material de estudo:

- Explica **por que** cada componente existe e como se encaixam.
- Ensina os três temas — **SRE**, **DevOps**, **Platform Engineering** — como trilhas de
  aprendizado.
- Hospeda tutoriais práticos dirigidos pelos cenários de `srectl demo`.
- Publica os runbooks e as definições de SLO para que um drill de incidente seja um
  exercício real.
- Documenta o ciclo de vida — *como a stack nasce* (`srectl up` → Argo CD → GitOps) —
  como uma página dedicada em `architecture/`.

## Stack & tooling

- **Hugo** (igual aos nossos outros projetos: stracectl, fraudctl, certificate-validate).
- **Bilíngue**: inglês (padrão, `/en/`) + português (`/pt-br/`).
- **Mermaid** para diagramas (os mesmos do [`ARCHITECTURE.md`](ARCHITECTURE.md)).
- Deploy via **GitHub Actions** para o GitHub Pages no merge para `main`.

## Estrutura do site

```text
site/
├── config.yaml             # idiomas, baseURL, params
├── content/
│   ├── en/
│   │   ├── _index.md       # visão geral
│   │   ├── architecture/
│   │   ├── stack/
│   │   ├── srectl/            # trilha de SRE (SLOs, incidente, caos, runbooks)
│   │   ├── devops/         # trilha de DevOps (CI/CD, GitOps, imagens)
│   │   ├── platform/       # trilha de Platform Engineering (self-service, política, obs)
│   │   ├── tutorials/      # hands-on do srectl demo
│   │   └── runbooks/
│   └── pt-br/              # conteúdo espelhado
└── themes/                 # ou um tema leve
```

## Trilhas de aprendizado

| Trilha | Público | Conteúdo |
| --- | --- | --- |
| **SRE** | Engenheiros de confiabilidade | SLO/SLI, error budgets, resposta a incidentes, drills de caos, runbooks |
| **DevOps** | Engenheiros de build/deploy | CI/CD, GitOps com Argo CD, pipelines de imagem, infra-as-code |
| **Platform Engineering** | Times de plataforma | APIs self-service, observabilidade como plataforma, política (Kyverno), segurança |

Cada trilha liga ao cluster *vivo* — por exemplo, a trilha de SRE liga aos dashboards do
Grafana e ao comando `srectl demo` correspondente.

## Mapeamento páginas → demo

| Página do site | Demo ao vivo |
| --- | --- |
| Observabilidade | `srectl demo load` → dashboards RED |
| Incidente / SLO | `srectl demo error` → alerta dispara |
| Engenharia de caos | `srectl demo chaos` |
| Segurança | `srectl demo security` |
| Certificados / TLS | `srectl demo certs` |
| Nossos projetos | `srectl demo projects` |

## Build & deploy

```bash
make site-dev        # servidor local (http://localhost:9000/srectl/) — EN + PT-BR
make site-build      # build estático
```

O Hugo é opcional: quando `hugo` não está no `PATH`, os alvos do Make usam
`ghcr.io/gohugoio/hugo:latest` via Docker. Sobrescreva com `HUGO_BIN=/caminho/hugo`
ou `HUGO_IMAGE=<imagem>` se necessário. O servidor de dev escuta e publica a porta
**9000** (`-p 9000:9000`, então a porta 1313 não é tocada). Use
`HUGO_PORT=9001 make site-dev` se ela estiver ocupada.

O GitHub Actions builda o site e faz push para o GitHub Pages (`peaceiris/actions-gh-pages`
ou o oficial `actions/deploy-pages`).

## Critérios de aceite

- O site fica acessível em `https://<user>.github.io/srectl/`.
- Toda página existe em **EN** e **PT-BR**.
- Todos os diagramas renderizam (Mermaid).
- Os tutoriais são reproduzíveis com um único comando `srectl demo`.

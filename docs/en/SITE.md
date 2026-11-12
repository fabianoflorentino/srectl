# Documentation Site (GitHub Pages)

The project publishes a bilingual website (Hugo) on **GitHub Pages** that is the main
surface for *learning* SRE / DevOps / Platform Engineering through srectl.

## Purpose

The site turns the running platform into study material:

- Explains **why** each component exists and how they fit together.
- Teaches the three themes — **SRE**, **DevOps**, **Platform Engineering** — as
  learning paths.
- Hosts hands-on tutorials driven by `srectl demo` scenarios.
- Publishes the runbooks and SLO definitions so an incident drill is a real exercise.
- Documents the lifecycle — *how the stack is born* (`srectl up` → Argo CD → GitOps) —
  as a dedicated page under `architecture/`.

## Stack & tooling

- **Hugo** (same as our other projects: stracectl, fraudctl, certificate-validate).
- **Bilingual**: English (default, `/en/`) + Portuguese (`/pt-br/`).
- **Mermaid** for diagrams (the same ones in [`ARCHITECTURE.md`](ARCHITECTURE.md)).
- Deployed via **GitHub Actions** to GitHub Pages on merge to `main`.

## Site structure

```text
site/
├── config.yaml             # languages, baseURL, params
├── content/
│   ├── en/
│   │   ├── _index.md       # overview
│   │   ├── architecture/
│   │   ├── stack/
│   │   ├── srectl/            # SRE learning path (SLOs, incident, chaos, runbooks)
│   │   ├── devops/         # DevOps path (CI/CD, GitOps, images)
│   │   ├── platform/       # Platform Engineering path (self-service, policy, obs)
│   │   ├── tutorials/      # srectl demo hands-on
│   │   └── runbooks/
│   └── pt-br/              # mirrored content
└── themes/                 # or a lightweight theme
```

## Learning paths

| Path | Audience | Content |
| --- | --- | --- |
| **SRE** | Reliability engineers | SLO/SLI, error budgets, incident response, chaos drills, runbooks |
| **DevOps** | Build/deploy engineers | CI/CD, GitOps with Argo CD, image pipelines, IaC-as-code |
| **Platform Engineering** | Platform teams | Self-service APIs, observability as a platform, policy (Kyverno), security |

Each path links to the *living* cluster — e.g. the SRE path links to Grafana dashboards
and the corresponding `srectl demo` command.

## Pages → demo mapping

| Site page | Live demo |
| --- | --- |
| Observability | `srectl demo load` → RED dashboards |
| Incident / SLO | `srectl demo error` → alert fires |
| Chaos engineering | `srectl demo chaos` |
| Security | `srectl demo security` |
| Certificates / TLS | `srectl demo certs` |
| Our projects | `srectl demo projects` |

## Build & deploy

```bash
make site-dev        # local server (http://localhost:1313) — EN + PT-BR
make site-build      # static build
```

GitHub Actions builds the site and pushes to GitHub Pages (`peaceiris/actions-gh-pages`
or the official `actions/deploy-pages`).

## Acceptance criteria

- The site is reachable at `https://<user>.github.io/srectl/`.
- Every page exists in **EN** and **PT-BR**.
- All diagrams render (Mermaid).
- Tutorials are reproducible with a single `srectl demo` command.

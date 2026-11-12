# srectl — agent instructions

Before any Go coding, review, debugging, troubleshooting, or setup task, load the `fabianoflorentino/golang-agent-skills@golang-how-to` skill first — it routes to whichever other Go skills the task needs.

## Go standards

- **Test coverage**: every phase that ships Go code must reach at least **80%** statement coverage. Run `make coverage` to verify it (it fails the build below the minimum). Do not merge a Go phase below 80%.
- Tests are executable specifications: assert on public behavior, not implementation details.
- Commits follow Conventional Commits and are written in English.

## Required Go skills

The following Go skills from `fabianoflorentino/golang-agent-skills` MUST always be applied when working on this project:

- `fabianoflorentino/golang-agent-skills@golang-cli`
- `fabianoflorentino/golang-agent-skills@golang-spf13-cobra`
- `fabianoflorentino/golang-agent-skills@golang-spf13-viper`
- `fabianoflorentino/golang-agent-skills@golang-testing`

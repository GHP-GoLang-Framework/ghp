# Contributing to GHP

Thanks for your interest in contributing. This document covers the essentials to get started — the details of each part are linked below.

## Before you start

- **Git and PR workflow**: read [`docs/git-workflow.md`](docs/git-workflow.md). Summary: branch off `main`, PR back to `main`, green CI + mandatory approval before merge. No exceptions — not even the maintainer commits directly to `main`.
- **Tests**: read [`docs/testing.md`](docs/testing.md). A minimum coverage of 90% (unit + integration) is mandatory in CI.
- **GHP syntax**: [`docs/template.ghp`](docs/template.ghp) is the reference for the language the parser/codegen will implement.

## Setting up the local environment

Prerequisites: Go (version from [`go.mod`](go.mod)). No Node.js needed — the toolchain is 100% Go.

```bash
git clone https://github.com/GHP-GoLang-Framework/GHP.git
cd GHP
git config core.hooksPath .githooks   # enables the git hooks
```

This activates `pre-commit` (gofmt + `go vet` on the staged Go files) and `commit-msg` (validates the commit message via the Go `gitlint` command).

## Making a change

```bash
git checkout main
git pull --ff-only origin main
git checkout -b feat/name-of-the-change

# ... edit, commit ...

gofmt -l ./src                    # no output = formatted
go vet ./src/...                  # static analysis
go test -short ./src/... -race    # unit tests (skips integration/e2e)
go test ./src/test/integration/... -race   # what CI will run anyway

git push -u origin feat/name-of-the-change
gh pr create --base main --title "feat(scope): description in the imperative"
```

## Commit messages

[Conventional Commits](https://www.conventionalcommits.org/en/), validated automatically by the `commit-msg` hook. Format:

```
type(optional scope): short description in the imperative
```

Accepted types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`. The PR title follows the same pattern — it is revalidated in CI and becomes the squash commit message on merge.

Recommended scope: the affected domain/package (`parser`, `codegen`, `pages`, `build`, `commitmsg`, `cli`, `docker`, `docs`...), never the file name.

## Code quality

- `gofmt` — formatting, fixed automatically by `pre-commit`.
- `go vet` — runs in `pre-commit` and in CI.
- `golangci-lint` — runs in CI (job `lint`); run `golangci-lint run ./src/...` locally if you want to get ahead.
- Tests covering whatever is added — see [`docs/testing.md`](docs/testing.md) for the conventions (table-driven tests, testing against `io.Writer`/explicit input instead of globals).

## Opening the PR

- Base is always `main`.
- Title in Conventional Commit.
- The `gate` job (lint, unit/integration/e2e tests, coverage ≥90%, build, vulnerability and secret scanning) must pass.
- Needs approval from [@castrogusttavo](https://github.com/castrogusttavo) before merge.
- Merge is by squash — the branch can have as many WIP commits as needed; only the final commit on `main` matters.

## Reporting bugs or suggesting features

Open a [GitHub issue](https://github.com/GHP-GoLang-Framework/GHP/issues) describing the problem or suggestion. If it is a bug, include a minimal example that reproduces the behavior.

## Questions

If anything in this document or its linked pages is outdated or confusing, open an issue or just fix it in a PR — documentation is code like any other part of the project.

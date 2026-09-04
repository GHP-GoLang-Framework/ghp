# AGENTS.md — GHP

Working guidelines for AI agents editing this repository. These rules are mandatory unless the user explicitly overrides them in the conversation.

## Repo overview

- Go monorepo: commands in `src/cmd` (`ghp`, `gitlint`), core under `src/internal` (`ast`, `parser`, `textutil`, `pages`, `build`, `commitmsg`, `transpiler/codegen`), integration/e2e under `src/test`.
- Source of truth for the GHP language: `docs/template.ghp`.
- Git hooks live in `.githooks/` and the repo is 100% Go (no Node): `commit-msg` runs the `gitlint` Go command (`src/cmd/gitlint`), `pre-commit` runs gofmt + `go vet`.
- VS Code extension lives under `editors/vscode` (TextMate grammar, snippets, grammar tests).
- Read `CONTRIBUTING.md`, `docs/git-workflow.md` and `docs/testing.md` before your first change.

## 1. Language: English only

- All new source code, comments, documentation, commit messages, PR titles/bodies and issue references must be in English.
- Do NOT write new PT-BR content anywhere.
- Existing PT-BR content is known debt: migrate it only in dedicated `docs:`/`refactor:` PRs, never inside feature PRs.

## 2. Granular commits, one branch per PR

- Never commit to `main`. Every change is a new branch (`<type>/<short-description>`) pushed as a PR against `main`.
- Split work into small logical commits: each commit is one self-contained unit that builds and passes tests.
- Commit messages follow Conventional Commits (enforced by the `commit-msg` git hook): imperative subject ≤ 100 chars, body explains the *why*, reference the Linear/GHP issue (`Refs: GHP-…`).
- Never force-push or rewrite pushed history; the repo merges by squash anyway.

## 3. Pull Requests

- Open a PR against `main` for every branch, filling `.github/PULL_REQUEST_TEMPLATE.md` (Resumo, Issue relacionada, Como testar, Checklist, Impacto) — in English.
- Every PR should preferably link the related GHP issue from Linear in the `Issue relacionada` section and reference it in commits (`Refs: GHP-…`). Linking is conditional on the issue already existing: creating and organizing issues is up to the project developers.
- Keep PRs small and focused. No unrelated refactors, renames or reformatting of untouched code (avoids diff noise).
- Flag breaking changes explicitly (Impacto section) and link the related issue.
- Never push directly to `main` and never bypass the CI gate.

## 4. Tests

- Do NOT create, modify or delete tests unless the user explicitly asks.
- Exception: if a change breaks existing tests or would drop CI coverage below 90%, stop and report to the user before adjusting tests.
- When tests ARE requested, follow `docs/testing.md` conventions: table-driven, explicit inputs (no `os.Stdout`/globals), integration/e2e guarded by `testing.Short()`.

## 5. Code style

- Code is gofmt-formatted and `go vet`-clean; the pre-commit hook runs both on staged files.
- Concise comments with a short example below. Comment the *why*, not the *what*; prefer one line plus an example over a paragraph:
  ```go
  // DedupePaths keeps the first occurrence of each import path.
  // Ex: {"a", "b", "a"} -> {"a", "b"}
  ```
- No comments that merely restate the code. Exported identifiers get a short English GoDoc line.
- Prefer the standard library; ask before adding new dependencies. Follow the existing package layout and naming.
- Return errors, never swallow them; no panics for expected conditions. Never log or commit secrets.

## 6. Documentation

- Update `README.md`, `docs/` or `editors/` whenever behavior or the language syntax changes (`docs/template.ghp` is the reference).
- If you change a convention, update the doc that describes it (`docs/git-workflow.md`, `docs/testing.md`, `.github/` templates).

## 7. Verify before opening the PR

- `gofmt -l ./src` (no output), `go vet ./src/...`, `go test ./src/...` (use `-short` for unit-only), and `npm test` under `editors/vscode` when the extension changed.
- Coverage must stay ≥ 90%.

## 8. Housekeeping

- Don't modify git config, hooks or CI secrets. There is no Node toolchain at the repo root; the only `npm install` that matters is under `editors/vscode`, and only run it when those dependencies change.
- Leave the working tree clean: no stray artifacts, no `coverage.out` committed.

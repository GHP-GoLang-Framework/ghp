# AGENTS.md — GHP

Guidelines for AI agents working in this repository. Follow these rules unless the user explicitly overrides them.

## Project overview

- GHP is a Go monorepo. The CLI lives in `src/cmd`; core packages live under `src/internal`, including the AST, parser, page discovery, build pipeline, and code generation.
- Integration and end-to-end tests live under `src/test`.
- `docs/template.ghp` is the source of truth for GHP template syntax.
- Git hooks live in `.github/hooks`. The hooks use POSIX shell; the repository's main toolchain is Go.
- Developer commands are defined in the root `Makefile`. Run `make help` to see the available targets.

Before making a change, read `CONTRIBUTING.md`, `docs/git-workflow.md`, and `docs/testing.md`.

## Language and documentation

- Write new source code comments, GoDoc, documentation, commit messages, pull-request titles, and pull-request descriptions in English.
- Do not write new PT-BR content anywhere. Existing PT-BR content is known debt: migrate it only in dedicated `docs:` or `refactor:` pull requests, never inside feature changes.
- Keep documentation consistent with the actual implementation. `docs/template.ghp` defines the language syntax.
- When behavior or syntax changes, update the relevant documentation, including `README.md` or `docs/` as appropriate.
- Avoid broad documentation migrations in feature changes. Keep unrelated cleanup separate.

## Git workflow

- Do not commit directly to `main`. Create a branch for each pull request, using a descriptive prefix such as `feat/`, `fix/`, `docs/`, or `refactor/`.
- Keep commits focused and self-contained. Prefer small changes that are easy to review and revert.
- Use Conventional Commits. Write an imperative subject of at most 100 characters; explain the reason in the body when it is not obvious.
- Reference an existing project issue when applicable, for example `Refs: GHP-123`. Do not invent issue IDs.
- Do not force-push or rewrite published history unless the maintainer explicitly requests it.
- Open pull requests against `main` and use `.github/PULL_REQUEST_TEMPLATE.md`.
- Keep pull requests focused. Call out breaking changes and their impact.

## Implementation guidelines

- Prefer simple, readable designs over speculative abstractions.
- Avoid duplicated logic. Reuse existing helpers and package boundaries when they fit.
- Follow the existing package layout and naming conventions.
- Prefer the standard library. Discuss new dependencies before adding them.
- Format Go code with `gofmt` and keep it clean under `go vet`.
- Write concise comments that explain why a decision exists, not what each line does. Exported Go identifiers need useful English GoDoc comments.
- Return errors instead of silently swallowing them. Do not panic for expected error conditions.
- Never log, expose, or commit credentials, tokens, or other secrets.
- Do not introduce unrelated formatting or refactoring into a focused change.

## Tests and verification

- Add or update tests when behavior changes or when a bug fix needs regression coverage. Do not remove useful coverage merely to make a change pass.
- Follow `docs/testing.md` for test conventions. Prefer table-driven tests where they improve clarity, explicit inputs, and `io.Writer` or returned values over global state and direct writes to `os.Stdout`.
- Keep integration and end-to-end tests aligned with the existing `testing.Short()` conventions.
- Preserve the repository's required coverage threshold of 90%, unless the maintainer explicitly changes that policy.
- Before submitting a change, run the relevant checks. The usual full verification is:

  ```sh
  gofmt -l ./src
  go vet ./src/...
  go test ./src/...
  ```

- Use the Makefile targets when available, since they mirror CI.
- If a check cannot be run, state that clearly and explain why. Do not claim that unrun checks passed.

## Repository hygiene

- Do not modify Git configuration, hooks, CI secrets, or release settings unless the task requires it.
- Do not commit generated binaries, coverage output, temporary files, or unrelated artifacts.
- Do not install dependencies unless needed for the change.
- Leave the working tree in a clear, reviewable state.

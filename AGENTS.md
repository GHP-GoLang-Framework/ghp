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
- Prefer the standard library. Discuss new dependencies before adding them.
- Format Go code with `gofmt` and keep it clean under `go vet`.
- Write concise comments that explain why a decision exists, not what each line does. Exported Go identifiers need useful English GoDoc comments.
- Return errors instead of silently swallowing them. Do not panic for expected error conditions.
- Never log, expose, or commit credentials, tokens, or other secrets.
- Do not introduce unrelated formatting or refactoring into a focused change.

## Code organization and modularity

- The internal packages form a strict one-way dependency graph. A package may only import packages below it in this list, never above it and never a sibling in a way that would close a cycle:
  - Leaves with no internal imports: `ast`, `colors`, `textutil`.
  - `parser` and `codegen` import `ast`; `pages` imports `textutil`.
  - `transpiler` imports `parser`, `codegen`, `pages`, and `colors`.
  - `build` imports `transpiler`, `pages`, and `colors`.
  - `cmd` imports `build`.
- Do not add an import that points upward or creates a cycle. When two packages at the same level need the same logic, move it down into a shared lower package instead of duplicating it or reaching sideways into a sibling.
- Keep one responsibility per package and keep it there. The parser turns a template into `ast` nodes and must not know how code is emitted; codegen reads `ast` nodes and must not re-parse text. The only channel between them is `ast`.
- Follow the existing package layout and naming conventions. Before writing a new helper, look for an existing one: `textutil` owns naming and normalization, and `colors` owns terminal output.
- `ast.Walk` is the traversal for observing every node with a plain callback, as `ast.CollectImports` does; reuse it instead of recursing over `Then`/`Else`/`Body` by hand. The codegen's `generateNodes` is a deliberate second traversal: it writes to a builder and returns an error per node, which `Walk`'s `func(Node)` callback cannot express. Add a third traversal only for a concrete reason, such as needing an effect or an error per node.
- When the same logic appears in a second place, extract it into the lowest package both callers can import and cover it with a test there. Prefer adding a parameter or a variant to an existing helper over copying a block and changing one detail.
- Wrap errors with `%w` and name the operation and the subject, as in `parse %s: %w`. Do not restate the underlying error's message.
- Keep package-level `var` declarations to immutable tables and interface assertions. Pass mutable state as a parameter so callers can see and test it.

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

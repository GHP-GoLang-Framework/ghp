# Contributing to GHP

Thanks for your interest in contributing to GHP. This guide covers the usual development workflow; the linked documents provide the details.

## Before you start

Read these documents before changing the code:

- [Git workflow](docs/git-workflow.md): branches, commits, pull requests, and review.
- [Testing guide](docs/testing.md): test organization, conventions, and coverage requirements.
- [GHP template syntax](docs/template.ghp): the authoritative reference for the language syntax.

## Development environment

Install the Go version declared in [`go.mod`](go.mod). The toolchain uses Go and POSIX shell.

Clone the repository and enable the Git hooks:

```sh
git clone https://github.com/GHP-GoLang-Framework/ghp.git
cd ghp
make setup
```

The hooks help catch common problems: `pre-commit` runs formatting and `go vet` checks on staged Go files, while `commit-msg` validates the commit message.

Run `make help` to see the available developer commands.

## Making a change

Start from an up-to-date `main` branch and create a focused branch:

```sh
git checkout main
git pull --ff-only origin main
git checkout -b feat/short-description
```

Make the change, add or update tests where appropriate, and run the checks relevant to it. The usual targets are:

```sh
make lint
make test
make test-integration
```

Check `make help` and the `Makefile` if a target differs in your checkout. Before opening a pull request, push your branch:

```sh
git push -u origin feat/short-description
gh pr create --base main
```

Do not commit directly to `main`.

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/). The commit-message hook validates the format:

```text
type(optional-scope): imperative summary
```

Common types include `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, and `revert`.

Use a scope for the affected domain or package when it adds clarity, for example `parser`, `codegen`, `pages`, `build`, `cli`, or `docs`. Avoid using a filename as the scope unless it genuinely describes the change.

Keep the subject concise and within the repository's enforced length limit. Explain the motivation in the body when the subject alone is not enough.

## Code quality and tests

- Run `gofmt` on changed Go files.
- Run `go vet` and the relevant tests before submitting.
- Add or update regression tests for bug fixes and tests for new behavior.
- Prefer table-driven tests when they make cases easier to compare.
- Prefer explicit inputs and returned values or `io.Writer` dependencies over global state and direct writes to `os.Stdout`.
- Follow the integration and end-to-end test conventions in [`docs/testing.md`](docs/testing.md).
- The CI pipeline currently requires at least 90% coverage. Check the testing guide for how coverage is measured.
- The `golangci-lint` job may be disabled depending on the available Go toolchain. Check the current CI configuration rather than assuming it is enabled.

## Opening a pull request

Before opening a pull request:

1. Ensure the branch is based on the latest relevant `main`.
2. Run the checks applicable to the change and report any that could not be run.
3. Update documentation when behavior, syntax, commands, or configuration changes.
4. Use the repository's pull-request template.
5. Describe the change, how it was tested, and any compatibility impact.

Pull requests target `main`. CI must pass and the required review approval must be obtained before merge. The repository uses squash merges, so keep the pull request history easy to review even if the branch contains work-in-progress commits.

## Reporting bugs and proposing features

Open an [issue on GitHub](https://github.com/GHP-GoLang-Framework/ghp/issues). For bugs, include a minimal reproduction, expected behavior, actual behavior, and relevant environment details. For feature proposals, describe the problem before prescribing a solution.

## Questions and corrections

If this guide or a linked document is unclear or outdated, open an issue or submit a focused documentation change. Documentation is part of the project and should stay accurate.

# Git and Pull Request workflow

How to contribute to GHP: branch, commit, PR, and review.

## General rule

- `main` is protected — nobody pushes directly to it.
- Every change starts as a branch from `main` and comes back as a Pull Request.
- Every PR must pass CI **and** be approved by [@castrogusttavo](https://github.com/castrogusttavo) before merge.

## Step by step

1. Clone and sync `main`:

   ```bash
   git clone https://github.com/GHP-GoLang-Framework/GHP.git
   cd GHP
   git checkout main
   git pull --ff-only origin main
   ```

2. Create a branch from `main`:

   ```bash
   git checkout -b feat/go-for-tag
   ```

   Suggested name: `<type>/<short-description>`, using the same types as commits (see below) — `feat/`, `fix/`, `docs/`, `refactor/`, `test/`, `chore/`, `ci/`, `build/`.

3. Commit on the branch following [Conventional Commits](https://www.conventionalcommits.org/en/) (enforced locally by the `commit-msg` git hook):

   ```
   feat(parser): recognize the <go:for expression> tag
   fix(codegen): emit case body before its sibling's text
   ```

   Accepted types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`.

4. Run locally before opening the PR (it is what CI will run anyway):

   ```bash
   gofmt -l ./src && go vet ./src/... && go test -short ./src/... -race
   go test ./src/test/integration/... -race
   ```

5. Push the branch and open the PR against `main`:

   ```bash
   git push -u origin feat/go-for-tag
   gh pr create --base main --title "feat(parser): recognize the <go:for expression> tag"
   ```

   The **PR title must also be a Conventional Commit** — it is validated automatically by CI.

6. Wait for CI (job `gate`: lint, unit tests, integration, e2e, minimum 90% coverage, build, vulnerability and secret scanning) and for approval from [@castrogusttavo](https://github.com/castrogusttavo).

7. Once approved and green, the PR is merged by **squash** — it becomes a single commit on `main`, with the PR title as message.

## What happens after merge

- The merge to `main` triggers CI again and, if green, publishes a Docker `edge` image to the GitHub Container Registry.
- Right after, a release is created **automatically**: a CalVer tag for the day (`YYYY.MM.DD[.N]`), a versioned Docker image, and a GitHub Release with notes generated from the commits. No manual action needed — every green merge to `main` is already a release.

## What will not work

- Direct push to `main` — blocked by branch protection.
- A PR without green CI.
- A PR without approval.
- Commit messages outside Conventional Commits — the `commit-msg` hook blocks locally even before the push reaches CI.

# Test workflow

How GHP organizes and runs tests: unit, integration, e2e, and coverage — all via native Go, no build tags.

## Where each test lives

The unit vs integration/e2e split uses no build tag: Go resolves everything with `go test ./src/...` (the Makefile at the repo root just wraps the same commands). The fast/slow split is done with the native `-short` flag (`testing.Short()`):

| Type | Where it lives | How it runs | Why |
| --- | --- | --- | --- |
| Unit | `*_test.go` next to the code (e.g.: `src/cmd/ghp/main_test.go`) | `go test -short ./src/...` | Same package as the tested code — accesses unexported functions directly, no need to export anything just to test. Always runs in short mode. |
| Integration | `src/test/integration/*_test.go` | `go test ./src/...` | External package — exercises multiple packages together or calls `go build` for real. Skipped in `-short` mode. |
| E2E | `src/test/e2e/*_test.go` | `go test ./src/...` (or `go test ./src/test/e2e/...` with `GHP_BINARY` set) | Runs against the already-compiled `ghp` binary. Skipped in `-short` mode. |

There is no `src/test/unit/` folder (and there should not be one) — putting a unit test far from the code it tests goes against idiomatic Go conventions.

The pattern in every integration/e2e test is:

```go
if testing.Short() {
	t.Skip("skipped in short mode (go test -short)")
}
```

So `go test -short ./src/...` runs only the fast ones and `go test ./src/...` runs everything, without needing a build tag.

## Running locally

Everything CI runs is wrapped in the repo [`Makefile`](../Makefile) — `make help` lists the targets. The flat Go commands are:

```bash
gofmt -l ./src                    # no output = formatted
go vet ./src/...                  # static analysis
go test -short ./src/... -race    # only fast tests (skips integration/e2e)
go test ./src/test/integration/... -race
go test ./src/test/e2e/...        # needs the compiled binary (see below)
go build -o bin/ghp ./src/cmd/ghp # builds the binary used by the e2e tests
```

Or, through the Makefile:

```bash
make lint       # gofmt check + go vet
make test       # = make test-short, the fast unit tests
make test-full  # everything, including integration/e2e
make test-e2e   # builds bin/ghp first, then runs the e2e tests
make coverage   # full run + enforces the 90% minimum
```

The e2e tests call the compiled `ghp` binary (via `GHP_BINARY`), not the source directly — hence `go build -o bin/ghp ./src/cmd/ghp` first. Today they are placeholders that always skip (GHP-14/15).

## Coverage

- Minimum required: **90%**, computed over the whole `./src/...` (`-coverpkg=./src/...` guarantees coverage counts even when the test lives in an external package like `src/test/integration`).
- Local command (also wrapped as `make coverage`, which enforces the minimum):
  `go test ./src/... -coverprofile=coverage.out -covermode=atomic -coverpkg=./src/...`
  then check with `go tool cover -func=coverage.out | tail -1` — the same calculation the `coverage` job in `ci.yml` does.
- The report also goes up to Codecov (`codecov.yml` defines the same 90% as target for `project` and `patch`).

```bash
go tool cover -func=coverage.out   # see coverage per function
go tool cover -html=coverage.out   # see coverage per line, in the browser
```

## In CI (`.github/workflows/ci.yml`)

Each test type runs in a separate job, all required by the `gate` job:

- `unit-tests` — `go test -short ./src/... -race` (integration/e2e skip via `-short`)
- `integration-tests` — `go test ./src/test/integration/... -race`
- `e2e-tests` — builds the binary, then `go test ./src/test/e2e/...` with `GHP_BINARY`
- `coverage` — runs `go test ./src/...` with `-coverprofile`, checks the 90% minimum, uploads to Codecov

## Conventions when writing a test

- **Table-driven tests** are the standard for multiple cases in the same function (see `src/cmd/ghp/main_test.go` → `TestRun`) — a `[]struct{...}` with `name`, inputs, and expected outputs, iterated with `t.Run(tt.name, ...)`.
- Test against `io.Writer`/explicit input, never against `os.Stdout`/`os.Args` globals directly — that is what makes the function testable without hacks (see `run(args []string, stdout io.Writer) int` in `src/cmd/ghp/cli.go`).
- `t.Helper()` in every test helper function, to point the error at the right line.
- Integration/e2e tests **never run in `-short` mode**: they start with the guard `if testing.Short() { t.Skip(...) }`.
- Integration/e2e tests still **do not really exist** — the files in `src/test/integration/` and `src/test/e2e/` today only have a `TestPlaceholder` with `t.Skip(...)`, pointing to the Linear issue that will replace the skip with a real test (GHP-13, GHP-14, GHP-15).

## Related

- [`git-workflow.md`](./git-workflow.md) — every PR needs the `gate` job (which includes the four test types above) green before merge.

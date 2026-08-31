# GHP

[![CI](https://github.com/GHP-GoLang-Framework/GHP/actions/workflows/ci.yml/badge.svg)](https://github.com/GHP-GoLang-Framework/GHP/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/GHP-GoLang-Framework/GHP/branch/main/graph/badge.svg)](https://codecov.io/gh/GHP-GoLang-Framework/GHP)
[![Go Version](https://img.shields.io/github/go-mod/go-version/GHP-GoLang-Framework/GHP)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

PHP-style templates, backed by real Go.

GHP compiles `.ghp` files — HTML with real embedded Go — straight into `net/http` handlers, with no runtime template engine. Everything between the tags is real Go: it compiles with `go build`, errors point to the right line of the `.ghp`, and any package (standard, external, or from your own module) can be imported.

## Status: active rewrite

GHP is being rewritten from scratch with a new syntax focused on DX. Today the repository only has the skeleton: the CLI is still a stub (`ghp dev`/`ghp build` do nothing real yet) and the parser/codegen for the syntax below does not exist yet. The syntax is already defined — see [`docs/template.ghp`](docs/template.ghp) — and the implementation is in progress.

## The syntax (target)

```html
<go:import ("fmt")/>

<go
    items := []string{"café", "chá", "suco"}
/>

<!doctype html>
<html lang="en">
<head>
  <title><go= fmt.Sprintf("Menu (%d items)", len(items)) /></title>
</head>
<body>
  <ul>
    <go:for _, item := range items/>
      <li><go= item /></li>
    <go:endfor/>
  </ul>

  <go:if len(items) == 0/>
    <p>Nothing on the menu yet.</p>
  <go:else/>
    <p>Enjoy your meal!</p>
  <go:endif/>
</body>
</html>
```

Every GHP tag is self-closing (ends with `/>`). Blocks with a body — `<go:if>`, `<go:switch>`, `<go:for>` — are closed by the `<go:endif/>`, `<go:endswitch/>`, and `<go:endfor/>` tags.

| Tag | What it does |
| --- | --- |
| `<go:import (...)/>` | Imports one or more packages — standard, external, or from your module. |
| `<go .../>` | Block of Go code (statement) — can open a scope between HTML chunks. |
| `<go= expression/>` | Renders an expression's value into the HTML, with automatic escaping. |
| `<go:if/>` / `<go:else/>` / `<go:endif/>` | Conditional, using Go's native operators. |
| `<go:switch/>` / `<go:case/>` / `<go:default/>` / `<go:endswitch/>` | Switch. |
| `<go:for/>` / `<go:endfor/>` | Loop — any form of `for`/`range` valid in Go. |

Routing is per file: `pages/index.ghp` becomes `/`, `pages/blog/[slug].ghp` becomes `/blog/{slug}`.

## Installation

There is no published binary or `go install` yet — that is part of the work in progress. For now, build from source:

```bash
git clone https://github.com/GHP-GoLang-Framework/GHP.git
cd GHP
go build -o bin/ghp ./src/cmd/ghp
```

A Docker image is also published on every merge to `main`: `edge` is the continuous build (the tip of development), and a versioned CalVer tag (`YYYY.MM.DD[.N]`) + `latest` are created automatically right after — every green merge is already a release.

```bash
docker pull ghcr.io/ghp-golang-framework/ghp:latest
docker run --rm ghcr.io/ghp-golang-framework/ghp:latest help
```

## Developing GHP

```bash
git clone https://github.com/GHP-GoLang-Framework/GHP.git
cd GHP
git config core.hooksPath .githooks  # enables the git hooks (no Node needed)
gofmt -l ./src  # no output = formatted
go vet ./src/...
go test -short ./src/... -race     # unit (skips integration/e2e)
go test ./src/... -race            # everything, including integration/e2e
go build -o bin/ghp ./src/cmd/ghp  # builds the binary
```

Coverage (minimum required: 90%): `go test ./src/... -coverprofile=coverage.out -covermode=atomic -coverpkg=./src/...` and then `go tool cover -func=coverage.out`. Details in [`docs/testing.md`](docs/testing.md).

## Documentation

- [`docs/template.ghp`](docs/template.ghp) — full syntax reference.
- [`docs/testing.md`](docs/testing.md) — how the tests (unit/integration/e2e) are organized and run.
- [`docs/git-workflow.md`](docs/git-workflow.md) — branch, commit, and Pull Request flow.
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — how to contribute.

## License

[MIT](LICENSE).

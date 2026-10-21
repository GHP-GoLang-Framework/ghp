# GHP — Golang Hyper Processor

[![CI](https://github.com/GHP-GoLang-Framework/ghp/actions/workflows/ci.yml/badge.svg)](https://github.com/GHP-GoLang-Framework/ghp/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/GHP-GoLang-Framework/ghp)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**PHP-style templates, backed by real Go.**

GHP compiles `.ghp` files into Go HTTP handlers. A page combines HTML with embedded Go code, and the generated handlers are compiled into the application rather than interpreted by a runtime template engine.

GHP is under active development. For the supported syntax and current behavior, use [`docs/template.ghp`](docs/template.ghp) as the language reference.

## How it works

The build pipeline discovers `.ghp` pages, parses each page into an AST, generates Go source code, and builds the resulting application.

- CLI: `src/cmd`
- Parser and AST: `src/internal`
- Code generation: `src/internal/transpiler/codegen`
- Integration and end-to-end tests: `src/test`

The generated Go code uses `net/http` handlers, and GHP adds the imports those handlers need — see [Template tags](#template-tags) for the rules.

## Example

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
  <go:elif len(items) == 1/>
    <p>One item on the menu.</p>
  <go:else/>
    <p>Enjoy your meal!</p>
  <go:endif/>
</body>
</html>
```

This example illustrates the intended template syntax. Refer to [`docs/template.ghp`](docs/template.ghp) for the canonical grammar and current implementation status.

### Template tags

| Tag | Purpose |
| --- | --- |
| `<go:import (...)/>` | Imports packages available to the page. |
| `<go .../>` | Executes a Go statement or block of statements. |
| `<go= expression/>` | Evaluates a Go expression and writes its value to the response. |
| `<go:if/>`, `<go:elif/>`, `<go:else/>`, `<go:endif/>` | Conditional control flow. |
| `<go:switch/>`, `<go:case/>`, `<go:default/>`, `<go:endswitch/>` | Switch control flow. |
| `<go:for/>`, `<go:endfor/>` | Loop control flow. |

`net/http` is always imported, and `fmt`/`html`/`io` are added when the page
actually echoes or prints — no need to declare them. Importing one of them
under an alias is rejected, since the generated code calls them by their
default name.

Routing is per file: `pages/index.ghp` becomes `/`, `pages/blog/[slug].ghp` becomes `/blog/{slug}`.

## Installation

GHP publishes release artifacts and container images through its release workflow. Check the [GitHub Releases page](https://github.com/GHP-GoLang-Framework/ghp/releases) for the latest available artifacts and installation instructions.

| Channel | Details |
| --- | --- |
| [Docker image](https://github.com/GHP-GoLang-Framework/ghp/pkgs/container/ghp) | `ghcr.io/ghp-golang-framework/ghp` |
| [AUR package](https://aur.archlinux.org/packages/ghp) | Arch Linux package, if available in the current release |
| [GitHub Releases](https://github.com/GHP-GoLang-Framework/ghp/releases) | Binary archives and checksums, plus platform packages when published |

For the full installation procedure, see [`docs/installation.md`](docs/installation.md).

## Developing GHP

```sh
git clone https://github.com/GHP-GoLang-Framework/ghp.git
cd ghp
make setup
make lint
make test
make build
```

Use `make help` to list all available commands. The testing guide explains test organization, integration and end-to-end tests, and the project's coverage requirement.

## Documentation

- [`docs/template.ghp`](docs/template.ghp): GHP syntax reference.
- [`docs/installation.md`](docs/installation.md): installation and release channels.
- [`docs/testing.md`](docs/testing.md): test organization and conventions.
- [`docs/git-workflow.md`](docs/git-workflow.md): branches, commits, and pull requests.
- [`CONTRIBUTING.md`](CONTRIBUTING.md): contribution guide.

## License

GHP is licensed under the [MIT License](LICENSE).

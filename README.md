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

### Source positions

Generated `.go` files carry `//line /abs/path/page.ghp:N` directives (each at column 1, before every statement, echo, text write, `if`/`elif`, `switch`/`case` and `for` that comes from template line N). Compiler errors, `go vet` output, stack traces and Delve/DWARF breakpoints therefore point at the `.ghp` file instead of the staged `.go` file. Directives are always on for `ghp build`, and they name the page in your project, not the staged copy that is deleted after the build; programmatically, `codegen.AssembleFile` takes the `.ghp` path and `codegen.Assemble` emits none. Columns and a separate source-map file are not emitted yet.

### Debugging

`ghp build --debug` (alias `-g`) builds for a debugger: optimizations and inlining are off (`-gcflags=all=-N -l`), the staged dir is mapped back to the project in the recorded paths, and the build never waits for Enter, so editors can run it as a prelaunch step. Breakpoints can then be set on lines of the `.ghp` pages and of the project's own `.go` files:

```sh
ghp build --debug -o .debug/app .
dlv dap --listen=127.0.0.1:0            # DAP launch: {"mode": "exec", "program": ".debug/app"}
# or, interactively:
dlv exec .debug/app                     # (dlv) break /abs/path/index.ghp:9
```

A page is an HTTP handler, so a breakpoint fires when a request reaches the route, not at startup. Only lines that start a statement, echo or text write carry code (a `<go ... />` block counts as the line of its opening tag); a breakpoint on any other line is reported as not verified by the debugger.

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

## Roadmap

| Status | Item |
| --- | --- |
| ✅ Done | `//line` mapping in codegen, `ghp build --debug`, and correct lines in multi-line `<go ... />` blocks |
| ✅ Done | `ghp-lsp` language server: diagnostics, completion, hover, formatting, and code actions |
| ✅ Done | Lumen editor support: `.ghp` highlighting and Delve debugging with breakpoints in templates |
| 🚧 In progress | Windows installer with self-update (#64) |
| ⏳ Next | All HTTP methods working: `GET`, `POST`, `PUT`, `PATCH`, `DELETE` |
| ⏳ Next | `require` for reusing `.ghp` files |
| ⏳ Next | AI-generated documentation |
| ⏳ Later | `references` and `signatureHelp` in `ghp-lsp`, non-blocking gopls, and aarch64 packaging tests |

See [`ROADMAP.md`](ROADMAP.md) for the full list, including known limitations.

## Documentation

- [`ROADMAP.md`](ROADMAP.md): what is delivered, in progress, and pending.
- [`docs/template.ghp`](docs/template.ghp): GHP syntax reference.
- [`docs/installation.md`](docs/installation.md): installation and release channels.
- [`docs/testing.md`](docs/testing.md): test organization and conventions.
- [`docs/git-workflow.md`](docs/git-workflow.md): branches, commits, and pull requests.
- [`CONTRIBUTING.md`](CONTRIBUTING.md): contribution guide.

## License

GHP is licensed under the [MIT License](LICENSE).

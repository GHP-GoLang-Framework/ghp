# GHP Roadmap

Last updated: 2026-11-12. Latest release: [`2026.11.12-2`](https://github.com/GHP-GoLang-Framework/ghp/releases/tag/2026.11.12-2).

Everything under **Delivered** was checked against the merged pull requests and releases of the repositories linked below.

## Ecosystem

| Repository | Purpose | State |
|---|---|---|
| [ghp](https://github.com/GHP-GoLang-Framework/ghp) | The framework and the `ghp` CLI | Active, 80 PRs, 30+ releases |
| [ghp-model](https://github.com/GHP-GoLang-Framework/ghp-model) | Example project: a complete GHP blog | Active, no releases |
| [url-shortener](https://github.com/GHP-GoLang-Framework/url-shortener) | URL shortener built with GHP, Postgres, and Redis | Active, no releases |
| [ghp-lsp](https://github.com/GHP-GoLang-Framework/ghp-lsp) | Standalone GHP language server for Neovim, Helix, Emacs, and any LSP client | Extracted from Lumen, CI green on Linux and Windows |
| [Lumen](https://github.com/castrogusttavo/lumen-text) | CPU-rendered code editor for Linux, written in Rust, with `.ghp` support, the `ghp-lsp` language server (`tools/ghp-lsp`), and Delve debugging | Active, 60+ PRs, latest release `2026.11.12.18` |
| [ghp-vscode](https://github.com/GHP-GoLang-Framework/ghp-vscode) | VS Code syntax highlighting for `.ghp` | `v0.1.0`, `.vsix` download |
| [tree-sitter-ghp](https://github.com/GHP-GoLang-Framework/tree-sitter-ghp) | Tree-sitter grammar with highlight queries | One PR merged, no releases |
| [ghp-zed](https://github.com/GHP-GoLang-Framework/ghp-zed) | Zed extension | Initial commit only, no releases |

## Delivered

### Language: parser and AST

| PR | What changed |
|---|---|
| [#17](https://github.com/GHP-GoLang-Framework/ghp/pull/17) | `.ghp` template parser and the AST node tree. |
| [#30](https://github.com/GHP-GoLang-Framework/ghp/pull/30) | Self-closing tags with `/>`. |
| [#39](https://github.com/GHP-GoLang-Framework/ghp/pull/39) | `<go:elif/>` conditional branch, across AST, parser, codegen, docs, and VS Code grammar. |
| [#42](https://github.com/GHP-GoLang-Framework/ghp/pull/42) | `<go=expr/>` recognized with no space after the equals sign. |
| [#60](https://github.com/GHP-GoLang-Framework/ghp/pull/60) | Tag payloads are validated and imports resolved at parse time. |
| [#81](https://github.com/GHP-GoLang-Framework/ghp/pull/81) | The parser reports the line where a tag payload starts, so breakpoints inside multi-line `<go ... />` blocks land on the right line. |

### Language: code generation

| PR | What changed |
|---|---|
| [#23](https://github.com/GHP-GoLang-Framework/ghp/pull/23) | Multiple imports, inline expressions, and Go generation (GHP-6, GHP-7). |
| [#24](https://github.com/GHP-GoLang-Framework/ghp/pull/24) | `go:if` and `go:else` conditionals (GHP-8). |
| [#25](https://github.com/GHP-GoLang-Framework/ghp/pull/25) | `go:switch`, `go:case`, and `go:default` (GHP-9). |
| [#26](https://github.com/GHP-GoLang-Framework/ghp/pull/26) | `go:for` loops (GHP-10). |
| [#27](https://github.com/GHP-GoLang-Framework/ghp/pull/27) | Final `.go` file assembled per page (GHP-11). |
| [#36](https://github.com/GHP-GoLang-Framework/ghp/pull/36) | `//line` directives dropped from the output, later restored by #79. |
| [#38](https://github.com/GHP-GoLang-Framework/ghp/pull/38) | Generators moved into the `gen/` subpackage. |
| [#41](https://github.com/GHP-GoLang-Framework/ghp/pull/41) | Line breaks and tabs stripped from text nodes. |
| [#79](https://github.com/GHP-GoLang-Framework/ghp/pull/79) | `//line` directives map generated Go back to `.ghp` lines, so compile errors, stack traces, and debuggers point at the template. |

### Routing, CLI, and build

| PR | What changed |
|---|---|
| [#28](https://github.com/GHP-GoLang-Framework/ghp/pull/28) | File-based routing: `pages/index.ghp` becomes `/` (GHP-12). |
| [#35](https://github.com/GHP-GoLang-Framework/ghp/pull/35) | The three `derive*` functions merged into one pass. |
| [#40](https://github.com/GHP-GoLang-Framework/ghp/pull/40) | `router.Register` dropped, routes wired in the generated `main.go`. |
| [#33](https://github.com/GHP-GoLang-Framework/ghp/pull/33) | Codegen and router moved under `transpiler/`. |
| [#34](https://github.com/GHP-GoLang-Framework/ghp/pull/34) | `build` and `dev` commands with live reload, and positional directories. |
| [#46](https://github.com/GHP-GoLang-Framework/ghp/pull/46) | Transpiler v2 pipeline, package reorganization, and full test coverage. |
| [#51](https://github.com/GHP-GoLang-Framework/ghp/pull/51) | `go.sum` staged alongside `go.mod`. |
| [#52](https://github.com/GHP-GoLang-Framework/ghp/pull/52) | `-p`, `-e`, and `-o` flags, and an owned staged-directory lifecycle. |
| [#80](https://github.com/GHP-GoLang-Framework/ghp/pull/80) | `ghp build --debug`: no optimization or inlining (`-N -l`), and `//line` points at the project's real pages. |

### Release and distribution

| PR | What changed |
|---|---|
| [#9](https://github.com/GHP-GoLang-Framework/ghp/pull/9) | Automatic CalVer tag and release on every green merge to `main`. |
| [#10](https://github.com/GHP-GoLang-Framework/ghp/pull/10) | Multi-platform binaries attached to each release. |
| [#48](https://github.com/GHP-GoLang-Framework/ghp/pull/48) | Releases through GoReleaser with AUR, apt, and COPR support. |
| [#65](https://github.com/GHP-GoLang-Framework/ghp/pull/65) | The Go suite is skipped on pull requests that do not touch Go. |
| [#67](https://github.com/GHP-GoLang-Framework/ghp/pull/67) | Change-filter defects fixed, and the release gated on real test runs. |
| [#69](https://github.com/GHP-GoLang-Framework/ghp/pull/69) | Release pipeline restored, with a guard for the permission rule that broke it. |
| [#70](https://github.com/GHP-GoLang-Framework/ghp/pull/70), [#71](https://github.com/GHP-GoLang-Framework/ghp/pull/71), [#72](https://github.com/GHP-GoLang-Framework/ghp/pull/72) | Token for the release URL, and a semver-valid same-day CalVer counter. |

### Code quality and tooling

| PR | What changed |
|---|---|
| [#1](https://github.com/GHP-GoLang-Framework/ghp/pull/1) | Bootstrap: toolchain, CI/CD, tests, and repository governance. |
| [#12](https://github.com/GHP-GoLang-Framework/ghp/pull/12), [#14](https://github.com/GHP-GoLang-Framework/ghp/pull/14) | Directory layout and native Go tests. |
| [#29](https://github.com/GHP-GoLang-Framework/ghp/pull/29) | Unused file read/write helpers removed (added in [#15](https://github.com/GHP-GoLang-Framework/ghp/pull/15)). |
| [#44](https://github.com/GHP-GoLang-Framework/ghp/pull/44), [#45](https://github.com/GHP-GoLang-Framework/ghp/pull/45) | Husky and commitlint replaced by a Go git hook; `internal/lint` merged into the command package. |
| [#47](https://github.com/GHP-GoLang-Framework/ghp/pull/47) | Developer tooling centralized in a Makefile and a POSIX shell hook. |
| [#56](https://github.com/GHP-GoLang-Framework/ghp/pull/56) | Repo docs, hooks, Docker, and CI config tidied. |
| [#73](https://github.com/GHP-GoLang-Framework/ghp/pull/73), [#75](https://github.com/GHP-GoLang-Framework/ghp/pull/75) | Import-block renderer deduplicated into `goimports`, a swallowed discover error fixed, and `fsutil` extracted. |
| [#76](https://github.com/GHP-GoLang-Framework/ghp/pull/76), [#77](https://github.com/GHP-GoLang-Framework/ghp/pull/77), [#78](https://github.com/GHP-GoLang-Framework/ghp/pull/78) | Codecov removed from CI, config, and docs. |

### Docs and governance

| PR | What changed |
|---|---|
| [#31](https://github.com/GHP-GoLang-Framework/ghp/pull/31) | PRs must link the Linear GHP issue when one exists. |
| [#32](https://github.com/GHP-GoLang-Framework/ghp/pull/32) | All pt-BR content translated to English. |
| [#59](https://github.com/GHP-GoLang-Framework/ghp/pull/59) | Project renamed to Golang Hyper Processor. |
| [#61](https://github.com/GHP-GoLang-Framework/ghp/pull/61), [#62](https://github.com/GHP-GoLang-Framework/ghp/pull/62), [#66](https://github.com/GHP-GoLang-Framework/ghp/pull/66) | AGENTS, CONTRIBUTING, and README reworded; tests expected by default; modularity and non-duplication rules. |

### VS Code extension history

| PR | What changed |
|---|---|
| [#16](https://github.com/GHP-GoLang-Framework/ghp/pull/16) | Syntax-highlighting extension for `.ghp`, with snippets and grammar tests. |
| [#53](https://github.com/GHP-GoLang-Framework/ghp/pull/53), [#64](https://github.com/GHP-GoLang-Framework/ghp/pull/64) | Extension removed from this repo, with its CI job and Dependabot entry. It now lives in [ghp-vscode](https://github.com/GHP-GoLang-Framework/ghp-vscode). |

### Dependency updates

Merged Dependabot PRs: [#2](https://github.com/GHP-GoLang-Framework/ghp/pull/2), [#3](https://github.com/GHP-GoLang-Framework/ghp/pull/3), [#4](https://github.com/GHP-GoLang-Framework/ghp/pull/4), [#5](https://github.com/GHP-GoLang-Framework/ghp/pull/5), [#6](https://github.com/GHP-GoLang-Framework/ghp/pull/6), [#7](https://github.com/GHP-GoLang-Framework/ghp/pull/7), [#8](https://github.com/GHP-GoLang-Framework/ghp/pull/8), [#18](https://github.com/GHP-GoLang-Framework/ghp/pull/18) to [#22](https://github.com/GHP-GoLang-Framework/ghp/pull/22), [#49](https://github.com/GHP-GoLang-Framework/ghp/pull/49), [#50](https://github.com/GHP-GoLang-Framework/ghp/pull/50), [#54](https://github.com/GHP-GoLang-Framework/ghp/pull/54), [#55](https://github.com/GHP-GoLang-Framework/ghp/pull/55), [#57](https://github.com/GHP-GoLang-Framework/ghp/pull/57), [#58](https://github.com/GHP-GoLang-Framework/ghp/pull/58). [#63](https://github.com/GHP-GoLang-Framework/ghp/pull/63) groups `golang.org/x` updates into one PR.

Closed without merging: [#11](https://github.com/GHP-GoLang-Framework/ghp/pull/11), [#13](https://github.com/GHP-GoLang-Framework/ghp/pull/13), [#68](https://github.com/GHP-GoLang-Framework/ghp/pull/68).

### Other repositories

| Repository | What was done |
|---|---|
| [ghp-vscode](https://github.com/GHP-GoLang-Framework/ghp-vscode) | Extension split into its own repo ([`fa73842`](https://github.com/GHP-GoLang-Framework/ghp-vscode/commit/fa73842)). Released as a downloadable `.vsix` instead of a marketplace publish ([`3ca3487`](https://github.com/GHP-GoLang-Framework/ghp-vscode/commit/3ca3487)), and `v0.1.0` is out. |
| [tree-sitter-ghp](https://github.com/GHP-GoLang-Framework/tree-sitter-ghp) | Initial grammar (2026-08-19). [PR #1](https://github.com/GHP-GoLang-Framework/tree-sitter-ghp/pull/1) added HTML tag recognition, a token-priority fix for self-closing tags, and `highlights.scm`. |
| [ghp-zed](https://github.com/GHP-GoLang-Framework/ghp-zed) | Initial Zed extension (2026-08-19). |
| [ghp-model](https://github.com/GHP-GoLang-Framework/ghp-model) | Example project moved to the new file format, and replaced by a complete GHP blog (2026-11-05). |
| [url-shortener](https://github.com/GHP-GoLang-Framework/url-shortener) | URL shortener on GHP, Postgres, and Redis. PRs [#1](https://github.com/GHP-GoLang-Framework/url-shortener/pull/1) to [#3](https://github.com/GHP-GoLang-Framework/url-shortener/pull/3): Docker image on GHCR, public deploy options, a quick tunnel over HTTP/2, an all-in-one image, and a `/health` endpoint. |

## Pending

| Priority | Item | Repo | Detail |
|---|---|---|---|
| High | Working HTTP methods (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`) | ghp | Routing, body and query parsing, and a correct response for each method. Needs per-method tests and a documentation example. |
| High | `require` | ghp | Include and reuse other `.ghp` files (layouts, partials), with path resolution and a clear error when the file is missing. |
| High | AI-generated documentation | ghp | Generate and maintain documentation from the code and templates, with human review before publishing. |
| Medium | Integration tests for documentation examples | ghp | Compile and run every example in CI so the docs do not go stale. Lands with HTTP methods and `require`. |
| Medium | Releases for the editor repos | tree-sitter-ghp, ghp-zed | Neither has a release, and `ghp-zed` has a single commit. |
| Medium | Tag a release for the example projects | ghp-model, url-shortener | Neither has a release. |
| Medium | Expose `references` and `signatureHelp` | Lumen (`ghp-lsp`) | Already in the router, only needs exposing. |
| Medium | gopls must not block the server loop | Lumen (`ghp-lsp`) | A request can currently stall the server for up to 10 s. |
| Medium | Highlight `<go= x />` inside HTML attributes | Lumen | |
| Medium | Packaging tests on Windows and Linux aarch64 | Lumen | |
| Medium | AI-written changelog summary, reviewed by a human | Lumen | Depends on an API key. |
| Low | `--debug` across `require` | ghp | Make `//line` point at the included file, not the including one. Depends on `require`. |

## Lumen (editor, `ghp-lsp`, and debugger)

[Lumen](https://github.com/castrogusttavo/lumen-text) is the editor with first-class GHP support. The `ghp-lsp` server lives in its `tools/ghp-lsp` directory and is also published as the standalone [ghp-lsp](https://github.com/GHP-GoLang-Framework/ghp-lsp) repository, for other editors.

### GHP support

| PR | What changed |
|---|---|
| [#55](https://github.com/castrogusttavo/lumen-text/pull/55) | `ghp-lsp`: tolerant lexer and parser, spanned AST, diagnostics, and `LineIndex` (UTF-16 positions). |
| [#54](https://github.com/castrogusttavo/lumen-text/pull/54) | `ghp-lsp`: gopls bridge. |
| [#58](https://github.com/castrogusttavo/lumen-text/pull/58) | `ghp-lsp`: stdio LSP server with diagnostics, completion, hover, symbols, folding, formatting, and code actions. |
| [#59](https://github.com/castrogusttavo/lumen-text/pull/59) | `ghp-lsp`: Go router, scope fallback, and incremental gopls sync. |
| [#62](https://github.com/castrogusttavo/lumen-text/pull/62) | `ghp-lsp`: hardening (cancellation, failure tests, `Cargo.lock`, package README). |
| [#56](https://github.com/castrogusttavo/lumen-text/pull/56) | `.ghp` language support in the editor: grammar, LSP wiring, and packaging, so `ghp-lsp` ships with every release package. |
| [#53](https://github.com/castrogusttavo/lumen-text/pull/53), [#57](https://github.com/castrogusttavo/lumen-text/pull/57) | Debug adapter registry with Delve, and the GHP pages adapter (`ghp build --debug` plus Delve). |
| [#60](https://github.com/castrogusttavo/lumen-text/pull/60) | Debug UI: toolbar, panel (call stack, variables, breakpoints, watch, console), conditional breakpoints, and hover-to-evaluate. |
| [#61](https://github.com/castrogusttavo/lumen-text/pull/61) | GHP debug polish, optional `launch.json`, GHP docs, and a packaging dry run. |
| [#63](https://github.com/castrogusttavo/lumen-text/pull/63) | No orphaned debuggee after Stop, Restart, or quit; program output and exit code in the Debug Console. |

### Standalone `ghp-lsp`

| Change | What changed |
|---|---|
| [ghp-lsp](https://github.com/GHP-GoLang-Framework/ghp-lsp) | `tools/ghp-lsp` extracted into its own repository in this organization (MIT), with CI on Linux and Windows and Neovim, Helix, and Emacs setup in the README. |

### Editor, releases, and platforms

| PR | What changed |
|---|---|
| [#30](https://github.com/castrogusttavo/lumen-text/pull/30), [#29](https://github.com/castrogusttavo/lumen-text/pull/29), [#33](https://github.com/castrogusttavo/lumen-text/pull/33), [#41](https://github.com/castrogusttavo/lumen-text/pull/41) | Hot exit, crash recovery, persistent undo, local history, save as administrator, file encodings, and EditorConfig. |
| [#31](https://github.com/castrogusttavo/lumen-text/pull/31), [#37](https://github.com/castrogusttavo/lumen-text/pull/37), [#38](https://github.com/castrogusttavo/lumen-text/pull/38), [#45](https://github.com/castrogusttavo/lumen-text/pull/45), [#49](https://github.com/castrogusttavo/lumen-text/pull/49) | Column selection, rulers, inline diagnostics, linked editing, replace in project, macros, zen mode, and right-to-left text. |
| [#11](https://github.com/castrogusttavo/lumen-text/pull/11), [#34](https://github.com/castrogusttavo/lumen-text/pull/34), [#36](https://github.com/castrogusttavo/lumen-text/pull/36), [#42](https://github.com/castrogusttavo/lumen-text/pull/42), [#43](https://github.com/castrogusttavo/lumen-text/pull/43), [#46](https://github.com/castrogusttavo/lumen-text/pull/46) | Git: branches, tags, worktrees, stash, commit graph, single-line staging, and a 3-way merge editor. |
| [#14](https://github.com/castrogusttavo/lumen-text/pull/14), [#21](https://github.com/castrogusttavo/lumen-text/pull/21), [#32](https://github.com/castrogusttavo/lumen-text/pull/32), [#26](https://github.com/castrogusttavo/lumen-text/pull/26) | Build and Run panel, test targets, several terminals, and no endless repaint while a dev server runs. |
| [#15](https://github.com/castrogusttavo/lumen-text/pull/15), [#17](https://github.com/castrogusttavo/lumen-text/pull/17), [#39](https://github.com/castrogusttavo/lumen-text/pull/39), [#48](https://github.com/castrogusttavo/lumen-text/pull/48) | AI panel with an IDE bridge for Claude Code, parallel sessions, and inline completions. |
| [#24](https://github.com/castrogusttavo/lumen-text/pull/24), [#25](https://github.com/castrogusttavo/lumen-text/pull/25), [#27](https://github.com/castrogusttavo/lumen-text/pull/27) | Markdown preview with TeX math, Mermaid, and alerts. |
| [#6](https://github.com/castrogusttavo/lumen-text/pull/6), [#40](https://github.com/castrogusttavo/lumen-text/pull/40), [#47](https://github.com/castrogusttavo/lumen-text/pull/47) | Language servers for many languages, project-wide Problems, image tabs, and spell checking. |
| [#19](https://github.com/castrogusttavo/lumen-text/pull/19), [#23](https://github.com/castrogusttavo/lumen-text/pull/23) | Windows build and tests green, plus macOS packaging. The `VCRUNTIME140.dll` requirement was removed in [`c563a70`](https://github.com/castrogusttavo/lumen-text/commit/c563a70). |
| [#35](https://github.com/castrogusttavo/lumen-text/pull/35), [#44](https://github.com/castrogusttavo/lumen-text/pull/44), [#52](https://github.com/castrogusttavo/lumen-text/pull/52) | Multiple windows, per-project settings, and a fix for the window handoff. Switching project resets Debug, Terminals, AI, and Git ([`b1054bb`](https://github.com/castrogusttavo/lumen-text/commit/b1054bb)). |
| [#28](https://github.com/castrogusttavo/lumen-text/pull/28), [#51](https://github.com/castrogusttavo/lumen-text/pull/51), [#65](https://github.com/castrogusttavo/lumen-text/pull/65) | CI in about 3 minutes, a site that rebuilds on every release, and an automatic changelog grouped by type. |
| [#50](https://github.com/castrogusttavo/lumen-text/pull/50) | Release V1 to V7. |

## In progress

| Item | Repo | Status |
|---|---|---|
| Windows installer (Inno Setup) with in-app self-update, [#64](https://github.com/castrogusttavo/lumen-text/pull/64) | Lumen | Open. Checks are green except the one that proves silent install, uninstall, and the N→N+1 update. |
| Roadmap with the real GHP and debug releases, [#66](https://github.com/castrogusttavo/lumen-text/pull/66) | Lumen | Open, docs. |

## Known limitations

| Limitation | Effect |
|---|---|
| Windows installer is not code-signed | SmartScreen shows a warning. |
| macOS build is not notarized | Gatekeeper asks for manual confirmation. |
| `ghp-lsp` only understands `>` in parenthesized Go expressions | Matches the reference parser. |
| A line without code inside a `<go ... />` block | The breakpoint shows as "unverified". |

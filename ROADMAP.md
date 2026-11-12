# GHP Roadmap

Last updated: 2026-11-12.

## Delivered

| Area | Item | What changed |
|---|---|---|
| Framework | #79 `//line` in codegen | Generated `.go` files map back to the `.ghp` lines. Compile errors, stack traces, and the debugger point at the template. |
| Framework | #80 `ghp build --debug` | Build without optimization or inlining (`-N -l`). `//line` points at the project's real `.ghp`, not the temporary copy. The flag no longer waits for Enter. |
| Framework | #81 correct line in multi-line `<go ... />` | The parser reports the line where the code starts. Breakpoints inside the block no longer land on the previous line. |
| ghp-lsp | Parser | Error-tolerant, with exact spans and UTF-16 positions. |
| ghp-lsp | Editing | Diagnostics, tag completion and hover, symbols, folding, formatting, and code actions (insert the missing `endif`, `endfor`, or `endswitch`). |
| ghp-lsp | Go inside tags | Goes through gopls, with a built-in fallback when it is missing. |
| ghp-lsp | Robustness | Cancellation, failure tests, `Cargo.lock`, and package README. |
| Lumen | Highlighting | `.ghp` files, with colored Go inside tags. |
| Lumen | Distribution | `ghp-lsp` ships with every release package. |
| Lumen | `.ghp` debugging | Breakpoints in the template via Delve, with variables, call stack, watch, console, and an optional `launch.json`. |
| Lumen | Debug stability | The debugged program is no longer orphaned, and the app log shows in the console. |
| Lumen | Project switching | Resets Debug, Terminals, AI, and Git. |
| Lumen | Windows | `VCRUNTIME140.dll` is no longer required. |
| Lumen | Release and site | Automatic changelog by type. The site rebuilds on every release. |

## In progress

| Item | Status | Remaining |
|---|---|---|
| Windows installer with self-update (PR #64) | Checks green, except one | Prove silent install, uninstall, and the N→N+1 update. |

## Pending

| Priority | Item | Area | Detail |
|---|---|---|---|
| High | Working HTTP methods (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`) | Framework | Routing, body and query parsing, and a correct response for each method. Needs per-method tests and a documentation example. |
| High | `require` | Framework | Include and reuse other `.ghp` files (layouts, partials), with path resolution and a clear error when the file is missing. |
| High | AI-generated documentation | Docs | Generate and maintain documentation from the code and templates, with human review before publishing. |
| Medium | `references` and `signatureHelp` in ghp-lsp | ghp-lsp | Already in the router, only needs exposing. |
| Medium | gopls must not block the server loop | ghp-lsp | A request can currently stall the server for up to 10 s. |
| Medium | Highlight `<go= x />` inside HTML attributes | Lumen | |
| Medium | Packaging tests on Windows and Linux aarch64 | CI | |
| Medium | AI-written changelog summary, reviewed by a human | Release | Depends on an API key. |
| Medium | Integration tests for documentation examples | Framework, docs | Compile and run every example in CI so the docs do not go stale. Lands with HTTP methods and `require`. |
| Low | `--debug` support across `require` | Framework | Make `//line` point at the included file, not the including one. Depends on `require`. |

## Known limitations

| Limitation | Effect |
|---|---|
| Windows installer is not code-signed | SmartScreen shows a warning. |
| macOS build is not notarized | Gatekeeper asks for manual confirmation. |
| `ghp-lsp` only understands `>` in parenthesized Go expressions | Matches the reference parser. |
| A line without code inside a `<go ... />` block | The breakpoint shows as "unverified". |

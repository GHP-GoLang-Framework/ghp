# Installing GHP

GHP is distributed from the [releases](https://github.com/GHP-GoLang-Framework/ghp/releases)
of this repository. Every merge to `main` that passes CI becomes a release
automatically (CalVer tag `YYYY.MM.DD[.N]`), and every CalVer tag gets:

- container images on `ghcr.io/ghp-golang-framework/ghp` (`<tag>` + `latest`,
  plus `edge` for the continuous build)
- GitHub Release assets: `checksums.txt`, bare-binary archives
  `ghp_<ver>_<os>_<arch>.tar.gz`/`.zip` for Linux, macOS and Windows on
  `amd64`/`arm64`, `.deb`/`.rpm` packages, and a source archive.

## Check the version

```bash
ghp version        # or: ghp --version
```

The version is the release tag. Source builds built without ldflags report
`dev`.

## Arch Linux / AUR

Two packages are published to the AUR by the release workflow:

| Package | Build type | Target |
| --- | --- | --- |
| [`ghp`](https://aur.archlinux.org/packages/ghp) | Compiles from source (tarball) | x86_64, aarch64 |
| [`ghp-bin`](https://aur.archlinux.org/packages/ghp-bin) | Prebuilt release archive | x86_64 |

```bash
# with an AUR helper
yay -S ghp            # or: ghp-bin
# or manually
git clone https://aur.archlinux.org/ghp.git
cd ghp && makepkg -si
```

Static copies of both PKGBUILDs live in `packaging/arch/`; the live versions
are generated and pushed by goreleaser on every release.

## Debian / Ubuntu (apt)

A `repository_dispatch` fires the `ghp-apt` repository on every release so it
can build a `deb`-based index. Until that repository is provisioned, install
the `.deb` attached to the release directly:

```bash
wget https://github.com/GHP-GoLang-Framework/ghp/releases/download/2026.09.04.1/ghp_2026.09.04.1_linux_amd64.deb
sudo apt install ./ghp_2026.09.04.1_linux_amd64.deb
```

## Fedora / RHEL (COPR / RPM)

A `.rpm` is attached to every release, and `packaging/fedora/ghp.spec` builds
from source for COPR. The COPR project is provisioned externally; until it is,
install the `.rpm` attached to the release directly:

```bash
sudo dnf install https://github.com/GHP-GoLang-Framework/ghp/releases/download/2026.09.04.1/ghp_2026.09.04.1_linux_amd64.rpm
```

## Docker

```bash
docker pull ghcr.io/ghp-golang-framework/ghp:latest
docker run --rm ghcr.io/ghp-golang-framework/ghp:latest help
```

`edge` is the continuous build of `main`; `latest` and the CalVer tags point
to the last release.

## From source

Requires Go 1.26+:

```bash
git clone https://github.com/GHP-GoLang-Framework/ghp.git
cd ghp
go build -ldflags "-X main.version=dev" -o bin/ghp ./src/cmd
# or use the Makefile: make build   (VERSION=... to inject a tag)
```

## Releasing / verifying a release locally

Installing goreleaser (not a Go module dependency — it runs only in CI and on
the maintainer's machine):

```bash
# from the repository root, with Go 1.26+
go install github.com/goreleaser/goreleaser/v2@latest
make release-check       # goreleaser check
make release-snapshot    # goreleaser release --snapshot --clean (dry run, no upload)
```
# Fetches a CalVer source tarball and builds GHP from source — the official
# Fedora path is COPR (docs/installation.md), which calls this spec with the
# release tag as the resolved Version. The Version in the repo is a
# placeholder: COPR substitutes 2026.09.04.1 style tags at build time.
Name:           ghp
Version:        replace-me
Release:        1%{?dist}
Summary:        GHP toolchain — PHP-style templates with real embedded Go

License:        MIT
URL:            https://github.com/GHP-GoLang-Framework/ghp
Source0:        %{url}/archive/refs/tags/%{version}.tar.gz

BuildRequires:  golang >= 1.26
Requires:       golang >= 1.26

%description
GHP compiles .ghp files — HTML with real embedded Go — straight into
net/http handlers, with no runtime template engine.

%prep
%setup -q -n ghp-%{version}

%build
export CGO_ENABLED=0
export GOFLAGS="-trimpath -mod=readonly -modcacherw"
go build -ldflags="-s -w -X main.version=%{version}" -o ghp ./src/cmd

%install
install -Dpm0755 ghp %{buildroot}%{_bindir}/ghp

%files
%{_bindir}/ghp
%license LICENSE

%changelog
* Fri Sep 04 2026 GHP maintainers <maintainers@ghp.dev> - 0.0.0-1
- Initial packaging
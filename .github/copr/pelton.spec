# Copr builds this from an SRPM whose Source0 is a pre-built binary tarball
# (see .github/workflows/release.yml, job fedora-copr) - the Go+webkit2gtk
# toolchain build already happened in GitHub Actions, with full network
# access. This spec does no compilation, only installs already-built files,
# so it needs nothing but a plain Fedora chroot to run in Copr.
Name:           pelton
Version:        %{_pelton_version}
Release:        1%{?dist}
Summary:        Open-source desktop email client
License:        GPL-3.0-or-later
URL:            https://github.com/TRC-Loop/Pelton
Source0:        %{name}-%{version}.tar.gz
BuildArch:      x86_64

Requires:       gtk3
Requires:       webkit2gtk4.1

%description
Pelton is an open-source, cross-platform desktop email client built with
Wails, Svelte and Go.

%prep
%setup -q

%install
rm -rf %{buildroot}
install -Dm0755 pelton %{buildroot}%{_bindir}/pelton
install -Dm0644 pelton.desktop %{buildroot}%{_datadir}/applications/pelton.desktop
install -Dm0644 pelton.png %{buildroot}%{_datadir}/pixmaps/pelton.png

%files
%{_bindir}/pelton
%{_datadir}/applications/pelton.desktop
%{_datadir}/pixmaps/pelton.png

%post
update-desktop-database &> /dev/null || :

%postun
update-desktop-database &> /dev/null || :

%changelog
* Thu Jan 01 1970 Pelton release automation <me@arne.sh> - 0-1
- Packaged automatically by GitHub Actions on release; see
  https://github.com/TRC-Loop/Pelton/releases for real changelogs.

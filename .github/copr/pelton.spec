# Copr builds this from an SRPM whose Source0 is a pre-built binary tarball
# (see .github/workflows/release.yml, job fedora-copr) - the Go+webkit2gtk
# toolchain build already happened in GitHub Actions, with full network
# access. This spec does no compilation, only installs already-built files,
# so it needs nothing but a plain Fedora chroot to run in Copr.

# The binary was built on the Actions runner (Ubuntu), not in the Copr
# chroot. Two rpmbuild defaults break a pre-built binary like that and made
# `dnf install pelton` fail outright:
#   * the automatic dependency generator scans the ELF and adds Requires for
#     every SONAME and glibc symbol version it links against; those come from
#     the build box's glibc/webkit and don't line up with the Fedora chroot,
#     so dnf reports "nothing provides ..." and refuses to install. Turn it
#     off and declare the real runtime deps by hand below instead.
#   * rpmbuild tries to build a -debuginfo subpackage from what is an already
#     stripped Go binary, which errors out. Disable it.
%global debug_package %{nil}
AutoReqProv:    no

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
install -Dm0644 pelton.metainfo.xml %{buildroot}%{_metainfodir}/pelton.metainfo.xml
install -Dm0644 pelton.png %{buildroot}%{_datadir}/pixmaps/pelton.png
install -Dm0644 pelton.png %{buildroot}%{_datadir}/icons/hicolor/1024x1024/apps/pelton.png

%files
%{_bindir}/pelton
%{_datadir}/applications/pelton.desktop
%{_metainfodir}/pelton.metainfo.xml
%{_datadir}/pixmaps/pelton.png
%{_datadir}/icons/hicolor/1024x1024/apps/pelton.png

%post
update-desktop-database &> /dev/null || :
touch --no-create %{_datadir}/icons/hicolor &> /dev/null || :

%postun
update-desktop-database &> /dev/null || :
if [ $1 -eq 0 ] ; then
    touch --no-create %{_datadir}/icons/hicolor &> /dev/null || :
    gtk-update-icon-cache %{_datadir}/icons/hicolor &> /dev/null || :
fi

%posttrans
gtk-update-icon-cache %{_datadir}/icons/hicolor &> /dev/null || :

%changelog
* Thu Jan 01 1970 Pelton release automation <me@arne.sh> - 0-1
- Packaged automatically by GitHub Actions on release; see
  https://github.com/TRC-Loop/Pelton/releases for real changelogs.

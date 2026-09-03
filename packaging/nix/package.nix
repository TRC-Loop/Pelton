{
  lib,
  stdenv,
  buildGoModule,
  fetchPnpmDeps,
  pnpmConfigHook,
  pnpm_10,
  nodejs_22,
  wails,
  pkg-config,
  gtk3,
  webkitgtk_4_1,
  glib-networking,
  gsettings-desktop-schemas,
  wrapGAppsHook3,
  makeWrapper,
  src ? lib.cleanSource ../..,
  version ? "2026.4",
}:

buildGoModule (finalAttrs: {
  pname = "pelton";
  inherit version src;

  __structuredAttrs = true;

  proxyVendor = true;

  vendorHash = "sha256-+Hj1Oqw0Kr2EWSkZ1QxFMAY01WfEUgg5GanGkB8vjS8=";

  overrideModAttrs = {
    preBuild = ''
      wails build -tags webkit2_41 -o pelton
    '';
  };

  nativeBuildInputs = [
    pkg-config
    pnpmConfigHook
    pnpm_10
    nodejs_22
    wails
  ]
  ++ lib.optionals stdenv.hostPlatform.isLinux [
    wrapGAppsHook3
    makeWrapper
  ];

  buildInputs = lib.optionals stdenv.hostPlatform.isLinux [
    gtk3
    webkitgtk_4_1
    glib-networking
    gsettings-desktop-schemas
  ];

  env = {
    pnpmDeps = fetchPnpmDeps {
      inherit (finalAttrs) pname version;
      src = "${finalAttrs.src}/frontend";
      pnpm = pnpm_10;
      fetcherVersion = 3;
      hash = "sha256-L/fNU53DCWv9fNyiQFD6VjiDjocx0vW61tMzQ8wiO20=";
    };
    pnpmRoot = "frontend";
  };

  buildPhase = ''
    runHook preBuild

    wails build -tags webkit2_41 -ldflags "-X main.version=${finalAttrs.version}" -o pelton

    runHook postBuild
  '';

  doCheck = false;

  installPhase = ''
    runHook preInstall

    install -Dm755 build/bin/pelton $out/bin/pelton
    install -Dm644 build/linux/pelton.desktop $out/share/applications/sh.arne.Pelton.desktop
    install -Dm644 build/linux/pelton.metainfo.xml $out/share/metainfo/sh.arne.Pelton.metainfo.xml
    install -Dm644 build/icons/pelton-512.png $out/share/icons/hicolor/512x512/apps/pelton.png

    runHook postInstall
  '';

  preFixup = lib.optionalString stdenv.hostPlatform.isLinux ''
    gappsWrapperArgs+=(
      --set GIO_EXTRA_MODULES "${glib-networking}/lib/gio/modules"
    )
  '';

  meta = {
    description = "Privacy-focused, cross-platform desktop email client";
    homepage = "https://pelton.app";
    changelog = "https://github.com/TRC-Loop/Pelton/releases/tag/v${finalAttrs.version}";
    license = lib.licenses.gpl3Plus;
    platforms = lib.platforms.linux;
    mainProgram = "pelton";
  };
})

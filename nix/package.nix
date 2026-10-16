{
  lib,
  buildGoModule,
  fetchNpmDeps,
  npmHooks,
  nodejs,
  pkg-config,
  wrapGAppsHook3,
  webkitgtk_4_1,
  gtk3,
  glib-networking,
  gsettings-desktop-schemas,
  src,
  version,
  # OAuth client credentials are baked in at compile time (see Makefile/.env.example).
  # Without them Gmail/Outlook OAuth sign-in is unavailable; plain IMAP/SMTP works.
  # Note: anything passed here ends up world-readable in the Nix store.
  googleClientId ? "",
  googleClientSecret ? "",
  microsoftClientId ? "",
  googleTestingClientId ? "",
  googleTestingClientSecret ? "",
}:

buildGoModule (finalAttrs: {
  pname = "aerion";
  inherit version;

  src = lib.cleanSourceWith {
    inherit src;
    filter =
      path: type:
      let
        base = baseNameOf path;
      in
      !(builtins.elem base [
        "node_modules"
        "dist"
        "bin"
        "result"
        ".flatpak-builder"
      ]);
  };

  vendorHash = "sha256-2NxgmklHScpcF4aOLPPzuVZZ2Qhmx/7SrCGnJtJ/x54=";

  npmDeps = fetchNpmDeps {
    src = "${finalAttrs.src}/frontend";
    hash = "sha256-a4N34fzmFMZoafX1rXUyeSAm12uV7njfhRkq3KT8h+I=";
  };
  npmRoot = "frontend";

  # Go module vendoring doesn't need the npm setup.
  overrideModAttrs = old: {
    nativeBuildInputs = lib.remove npmHooks.npmConfigHook old.nativeBuildInputs;
    preBuild = "";
  };

  nativeBuildInputs = [
    nodejs
    npmHooks.npmConfigHook
    pkg-config
    wrapGAppsHook3
  ];

  buildInputs = [
    webkitgtk_4_1
    gtk3
    glib-networking # TLS for WebKit
    gsettings-desktop-schemas
  ];

  # Only the main binary; cmd/aerion-creds is a maintainer tool.
  subPackages = [ "." ];

  # Same tags `wails build -tags webkit2_41` passes to the Go compiler.
  tags = [
    "desktop"
    "production"
    "webkit2_41"
  ];

  ldflags =
    let
      m = "github.com/hkdb/aerion/internal/oauth2";
    in
    [
      "-s"
      "-w"
      "-X '${m}.GoogleClientID=${googleClientId}'"
      "-X '${m}.GoogleClientSecret=${googleClientSecret}'"
      "-X '${m}.MicrosoftClientID=${microsoftClientId}'"
      "-X '${m}.GoogleTestingClientID=${googleTestingClientId}'"
      "-X '${m}.GoogleTestingClientSecret=${googleTestingClientSecret}'"
    ];

  # Build the Svelte frontend first; main.go embeds frontend/dist.
  preBuild = ''
    pushd frontend
    npm run build
    popd
  '';

  # Upstream tests touch the network/keyring/display; run them via `go test` in the dev shell.
  doCheck = false;

  postInstall = ''
    for sz in 32 48 64 128 256; do
      install -Dm644 "build/linux/icons/''${sz}x''${sz}/io.github.hkdb.Aerion.png" \
        "$out/share/icons/hicolor/''${sz}x''${sz}/apps/io.github.hkdb.Aerion.png"
    done
    install -Dm644 build/linux/icons/scalable/io.github.hkdb.Aerion.svg \
      $out/share/icons/hicolor/scalable/apps/io.github.hkdb.Aerion.svg
    install -Dm644 build/linux/aerion.desktop \
      $out/share/applications/io.github.hkdb.Aerion.desktop
    install -Dm644 build/flatpak/io.github.hkdb.Aerion.metainfo.xml \
      $out/share/metainfo/io.github.hkdb.Aerion.metainfo.xml
  '';

  meta = {
    description = "Lightweight, open source email client";
    homepage = "https://github.com/hkdb/aerion";
    license = lib.licenses.asl20;
    mainProgram = "aerion";
    platforms = lib.platforms.linux;
  };
})

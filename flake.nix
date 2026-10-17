{
  description = "Aerion - a lightweight, open source email client (personal fork)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});

      version = (nixpkgs.lib.importJSON ./frontend/package.json).version;

      packageArgs = {
        inherit version;
        src = self;
        # Own Azure app registration (public client, not a secret) for Outlook.com OAuth.
        microsoftClientId = "1f78976f-fb3e-442d-80df-762952279fca";
        # The app registration is personal-accounts-only, which rejects /common/.
        microsoftTenant = "consumers";
      };
    in
    {
      packages = forAllSystems (pkgs: {
        aerion = pkgs.callPackage ./nix/package.nix packageArgs;
        default = self.packages.${pkgs.stdenv.hostPlatform.system}.aerion;
      });

      overlays.default = final: _prev: {
        aerion = final.callPackage ./nix/package.nix packageArgs;
      };

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          inputsFrom = [ self.packages.${pkgs.stdenv.hostPlatform.system}.aerion ];
          packages = with pkgs; [
            go
            gopls
            golangci-lint
            nodejs
            wails
            gnumake
          ];
          # Picked up by the Makefile (make dev / make build) like a .env entry.
          MICROSOFT_CLIENT_ID = packageArgs.microsoftClientId;
          MICROSOFT_TENANT = packageArgs.microsoftTenant;
          # WebKitGTK's DMA-BUF renderer misbehaves on some GPU/driver combos
          # when run from a dev shell; uncomment if `wails dev` shows a blank window.
          # WEBKIT_DISABLE_DMABUF_RENDERER = "1";
        };
      });

      formatter = forAllSystems (pkgs: pkgs.nixfmt);
    };
}

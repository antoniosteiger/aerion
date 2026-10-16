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
    in
    {
      packages = forAllSystems (pkgs: {
        aerion = pkgs.callPackage ./nix/package.nix {
          inherit version;
          src = self;
        };
        default = self.packages.${pkgs.stdenv.hostPlatform.system}.aerion;
      });

      overlays.default = final: _prev: {
        aerion = final.callPackage ./nix/package.nix {
          inherit version;
          src = self;
        };
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
          # WebKitGTK's DMA-BUF renderer misbehaves on some GPU/driver combos
          # when run from a dev shell; uncomment if `wails dev` shows a blank window.
          # WEBKIT_DISABLE_DMABUF_RENDERER = "1";
        };
      });

      formatter = forAllSystems (pkgs: pkgs.nixfmt);
    };
}

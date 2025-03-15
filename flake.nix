{
  description = "Go development environment with Nix Flakes";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs =
    {
      nixpkgs,
      ...
    }:
    let
      system = "x86_64-linux"; # Change to "aarch64-linux" for ARM, or "x86_64-darwin" for macOS
      pkgs = import nixpkgs { inherit system; };
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [
          go
        ];
      };

      packages.${system}.default = pkgs.buildGoModule {
        pname = "distribution-patch";
        version = "0.1.0";
        src = ./.;
        vendorHash = "sha256-VTXiI77KaRZWQtbTXbWT2IHPDT9TIxklKP64Z0ip+Dc=";
      };
    };
}

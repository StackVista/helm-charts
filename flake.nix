{
  description = "StackState helm chart";

  nixConfig.bash-prompt = "STS HELM $ ";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs";

    nixpkgs-helm-docs.url = "github:NixOS/nixpkgs?rev=95732323ae530b399511ac9a4e654d2116ec079c";

    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, nixpkgs-helm-docs }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; overlays = [ ]; };
        pkgs-helm-docs = import nixpkgs-helm-docs { inherit system; overlays = [ ]; };
      in {
        devShell = pkgs.mkShell {
          buildInputs = (with pkgs; [
          ]) ++ [
            pkgs-helm-docs.helm-docs
          ];

          shellHook = ''
            echo Welcome to Helm Chart Development Console
            echo This environment is as of now incomplete, but is required to pin helm-docs to version 1.7.0
            printf "\n"
          '';
        };
      });
}

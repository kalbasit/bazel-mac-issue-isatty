let
  hostPkgs = import <nixpkgs> {};
  # Look here for information about how to generate `nixpkgs-version.json`.
  #  → https://nixos.wiki/wiki/FAQ/Pinning_Nixpkgs
  pinnedVersion = hostPkgs.lib.importJSON ./nixpkgs-version.json;
  pinnedPkgs = hostPkgs.fetchFromGitHub {
    owner = "NixOS";
    repo = "nixpkgs-channels";
    inherit (pinnedVersion) rev sha256;
  };
in

# This allows overriding nixpkgs by passing `--arg nixpkgs ...`
{ nixpkgs ? import pinnedPkgs {} }:

nixpkgs.mkShell {
  buildInputs = with nixpkgs; [
    bazel
    go
    jq
    nodejs-8_x
    python2Full
    python3Full
    yarn
  ];
}

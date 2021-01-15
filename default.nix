let
  unstableTarball = fetchTarball
    "https://github.com/NixOS/nixpkgs/archive/nixos-unstable.tar.gz";
  pkgs = import <nixpkgs> { };

in with pkgs;
pkgs.mkShell {
  name = "go";
  buildInputs = [
    entr
    go
  ];
}

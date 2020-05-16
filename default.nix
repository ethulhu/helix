{ pkgs ? import <nixpkgs> {} }:
with pkgs;

buildGoModule rec {
  name = "hello-${version}";
  version = "latest";
  goPackagePath = "github.com/ethulhu/helix";

  modSha256 = "1l22v09nx18zamn30sn4hbxh6p96k048ld2rr3gpi2jnsg5dnvvp";

  preBuild = ''
    go generate ./...
  '';

  src = ./.;
}

{ pkgs ? import <nixpkgs> {} }:
with pkgs;

buildGoModule rec {
  name = "helix-${version}";
  version = "latest";
  goPackagePath = "github.com/ethulhu/helix";

  modSha256 = "1013fhdj9ffx158g2nx86wdz77ahvk5anb5qpwj2xbngpnk6ms2p";

  preBuild = ''
    go generate ./...
  '';

  src = ./.;
}

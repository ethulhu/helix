# SPDX-FileCopyrightText: 2020 Ethel Morgan
#
# SPDX-License-Identifier: MIT

{ pkgs ? import <nixpkgs> {} }:
with pkgs;

buildGoModule rec {
  name = "helix-${version}";
  version = "latest";
  goPackagePath = "github.com/ethulhu/helix";

  modSha256 = "11a39sqsg4a9bihg44cnwdhgl39m2bn90pq2cmvh3xcyz6x5rd20";

  preBuild = ''
    go generate ./...
  '';

  src = ./.;
}

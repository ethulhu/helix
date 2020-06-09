# SPDX-FileCopyrightText: 2020 Ethel Morgan
#
# SPDX-License-Identifier: MIT

{ pkgs ? import <nixpkgs> {} }:
with pkgs;

buildGoModule rec {
  name = "helix-${version}";
  version = "latest";
  goPackagePath = "github.com/ethulhu/helix";

  modSha256 = "154wc401jkp3pwd3r06slrhb4zw9sqlsasq6alkcrvsk9kzg6jwp";

  preBuild = ''
    go generate ./...
  '';

  src = ./.;
}

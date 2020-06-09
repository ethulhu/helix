# SPDX-FileCopyrightText: 2020 Ethel Morgan
#
# SPDX-License-Identifier: MIT

{ pkgs ? import <nixpkgs> {} }:
with pkgs;

buildGoModule rec {
  name = "helix-${version}";
  version = "latest";
  goPackagePath = "github.com/ethulhu/helix";

  modSha256 = "0i8kxcmkv2d4fqd0cw3wgsppy33742fwlkh4wa41qywb1f97jkrj";

  preBuild = ''
    go generate ./...
  '';

  src = ./.;
}

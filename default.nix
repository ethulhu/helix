# SPDX-FileCopyrightText: 2020 Ethel Morgan
#
# SPDX-License-Identifier: MIT

{ pkgs ? import <nixpkgs> {} }:
with pkgs;

buildGoModule rec {
  name = "helix-${version}";
  version = "latest";
  goPackagePath = "github.com/ethulhu/helix";

  modSha256 = "0bk6179cvk2fnlam388npkwq71jqr3dzm3jqdgcpskih0hw7h8y5";

  preBuild = ''
    go generate ./...
  '';

  src = ./.;
}

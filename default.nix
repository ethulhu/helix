# SPDX-FileCopyrightText: 2020 Ethel Morgan
#
# SPDX-License-Identifier: MIT

{ pkgs ? import <nixpkgs> {} }:
with pkgs;

buildGoModule rec {
  name = "helix-${version}";
  version = "latest";
  goPackagePath = "github.com/ethulhu/helix";

  modSha256 = "1ys2lsyw3lssyaxxrczcjpwz1gdxxj516mmvrl8ci1ap52g959yq";

  preBuild = ''
    go generate ./...
  '';

  src = ./.;
}

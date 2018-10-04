{ nixpkgs ? import <nixpkgs> {} }:

nixpkgs.callPackage ./gomod2nix.nix {}

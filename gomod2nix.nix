{ buildGoPackage }:

buildGoPackage rec {
  name = "gomod2nix";
  src = ./.;
  goPackagePath = "github.com/kwohlfahrt/${name}";
}

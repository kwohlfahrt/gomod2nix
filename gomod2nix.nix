{ stdenv, buildGoPackage, makeWrapper, go, nix-prefetch-git }:

buildGoPackage rec {
  name = "gomod2nix";
  goPackagePath = "github.com/kwohlfahrt/${name}";

  src = ./.;

  goDeps = ./deps.nix;
  nativeBuildInputs = [ makeWrapper ];

  postInstall = ''
    wrapProgram $bin/bin/gomod2nix \
      --prefix PATH : ${stdenv.lib.makeBinPath [ nix-prefetch-git go ]}
  '';

  # All tests currently require network connectivity
  doCheck = false;

  # Need to run `go list`
  allowGoReference = true;
}

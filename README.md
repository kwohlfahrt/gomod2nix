# gomod2nix

[![Build Status](https://travis-ci.org/kwohlfahrt/gomod2nix.svg?branch=master)](https://travis-ci.org/kwohlfahrt/gomod2nix)

A tool to create a `.nix` file with locked dependencies from a go project with a
`go.mod`. Very similar to `vgo2nix`.

## Usage

    gomod2nix /path/to/project > deps.nix

Then, in your `package.nix`, you should include `goDeps = ./deps.nix`. See this
project itself for an example.

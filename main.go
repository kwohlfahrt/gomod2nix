package main

import (
	"fmt"
	"flag"
	"os"
)

func main() {
	arguments := flag.NewFlagSet("", flag.ExitOnError)
	output := arguments.String("output", "./deps.nix", "The .nix file to write dependencies to.")
	arguments.Usage = func() {
		fmt.Fprintln(arguments.Output(), "Usage:", os.Args[0], "[OPTIONS] [path]")
		fmt.Println("OPTIONS:")
		arguments.PrintDefaults()
	}

	arguments.Parse(os.Args[1:])

	path := "./go.mod"

	switch len(arguments.Args()) {
	case 0:
	case 1:
		path = arguments.Arg(0)
	default:
		arguments.Usage()
		os.Exit(1)
	}

	fmt.Println("Input:", path, "Output", *output)
}

package main

import (
	"fmt"
	"flag"
	"bufio"
	"os"
	"strings"
	"regexp"
	"os/exec"
	"golang.org/x/tools/go/vcs"
)

func main() {
	// No options at the moment
	arguments := flag.NewFlagSet("", flag.ExitOnError)
	arguments.Usage = func() {
		fmt.Fprintln(arguments.Output(), "Usage:", os.Args[0], "[path]")
	}
	arguments.Parse(os.Args[1:])

	path := "./."
	switch len(arguments.Args()) {
	case 0:
	case 1:
		path = arguments.Arg(0)
	default:
		arguments.Usage()
		os.Exit(1)
	}

	pseudoVersionRegex := regexp.MustCompile("v[0-9.]+.-[0-9]+-([0-9a-f]+)")

	cmd := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}} {{.Standard}} {{ if .Module }}{{ .Module.Version }}{{ end }}", path)
	stdout, err := cmd.StdoutPipe()
	if (err != nil) {
		panic(err)
	}
	scanner := bufio.NewScanner(stdout)

	if err := cmd.Start(); err != nil {
		panic(err)
	}
	for scanner.Scan() {
		components := strings.Split(scanner.Text(), " ")
		standard := components[1] == "true"
		if standard {
			continue
		}

		packagePath := components[0]

		repoRoot, err := vcs.RepoRootForImportPath(packagePath, false)
		if err != nil {
			panic(err)
		}

		version := components[2]
		if version == "" {
			// Not in a module
			continue
		}

		match := pseudoVersionRegex.FindStringSubmatch(version)
		rev := version
		if len(match) > 0 {
			rev = match[1]
		}

		fmt.Println(repoRoot.Repo, rev)
	}
	if err := cmd.Wait(); err != nil {
		switch err := err.(type) {
		case *exec.ExitError:
			fmt.Println(err.Stderr)
		}
		panic(err)
	}
}

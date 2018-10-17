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
	"encoding/json"
)

type Package struct {
	GoPackagePath string
	URL           string
	Rev           string
	Sha256        string
}

const packageFmt = `{
  goPackagePath = "%s";
  fetch = {
    type = "git";
    url = "%s";
    rev = "%s";
    sha256 = "%s";
  };
}`

const depFmt = "{{ .ImportPath }} {{ .Standard }} {{ .DepOnly }} {{ if .Module }}{{ .Module.Path }} {{ .Module.Version }}{{ end }}"

func (pkg Package) String() string {
	return fmt.Sprintf(packageFmt, pkg.GoPackagePath, pkg.URL, pkg.Rev, pkg.Sha256)
}

type Prefetch struct {
	URL string
	Rev string
	Sha256 string
}


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

	cmd := exec.Command("go", "list", "-deps", "-f", depFmt, path)
	stdout, err := cmd.StdoutPipe()
	if (err != nil) {
		panic(err)
	}
	scanner := bufio.NewScanner(stdout)

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	var packages []Package
	for scanner.Scan() {
		components := strings.Split(scanner.Text(), " ")
		importPath, standard, depOnly := components[0], components[1] == "true", components[2] == "true"
		if standard || !depOnly {
			continue
		} else if len(components) < 5 {
			fmt.Fprintln(os.Stderr, importPath, "is neither a module nor in the standard library")
			os.Exit(1)
		}

		packagePath, version := components[3], components[4]
		repoRoot, err := vcs.RepoRootForImportPath(packagePath, false)
		if err != nil {
			panic(err)
		}

		match := pseudoVersionRegex.FindStringSubmatch(version)
		rev := version
		if len(match) > 0 {
			rev = match[1]
		}

		prefetchOut, err := exec.Command("nix-prefetch-git", "--quiet", repoRoot.Repo, "--rev", rev).Output()
		if (err != nil) {
			panic(err)
		}

		var prefetch Prefetch
		json.Unmarshal([]byte(prefetchOut), &prefetch)
		packages = append(packages, Package {
			GoPackagePath: packagePath,
			URL: prefetch.URL,
			Sha256: prefetch.Sha256,
			Rev: prefetch.Rev,
		})
	}

	if err := cmd.Wait(); err != nil {
		switch err := err.(type) {
		case *exec.ExitError:
			fmt.Println(err.Stderr)
		}
		panic(err)
	}

	fmt.Print("[")
	for _, pkg := range packages {
		fmt.Print(pkg)
	}
	fmt.Println("]")
}

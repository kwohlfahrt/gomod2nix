package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/tools/go/vcs"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
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

func (pkg Package) String() string {
	return fmt.Sprintf(packageFmt, pkg.GoPackagePath, pkg.URL, pkg.Rev, pkg.Sha256)
}

type Prefetch struct {
	URL    string
	Rev    string
	Sha256 string
}

type Dependency struct {
	Path    string
	Version string
}

func (dep Dependency) String() string {
	return fmt.Sprintf("%s %s", dep.Path, dep.Version)
}

func DepsForPath(path string) map[Dependency]struct{} {
	const depFmt = "{{ .ImportPath }} {{ .Standard }} {{ .DepOnly }} {{ if .Module }}{{ .Module.Path }} {{ .Module.Version }}{{ end }}"

	cmd := exec.Command("go", "list", "-deps", "-f", depFmt, path)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(stdout)
	deps := make(map[Dependency]struct{})
	for scanner.Scan() {
		components := strings.Split(scanner.Text(), " ")
		importPath, standard, depOnly := components[0], components[1] == "true", components[2] == "true"
		if standard || !depOnly {
			continue
		}

		if len(components) < 5 {
			panic(fmt.Sprintf("%s is neither a module nor in the standard library", importPath))
		}

		packagePath, version := components[3], components[4]
		if version == "" {
			// A package in the current directory
			continue
		}
		deps[Dependency{
			Path:    packagePath,
			Version: version,
		}] = struct{}{}
	}

	if err := cmd.Wait(); err != nil {
		switch err := err.(type) {
		case *exec.ExitError:
			fmt.Fprintln(os.Stderr, err.Stderr)
		}
		panic(err)
	}

	return deps
}

var pseudoVersionRegex = regexp.MustCompile("v[0-9.]+.-[0-9]+-([0-9a-f]+)")

func PrefetchDependency(dep Dependency) Package {
	repoRoot, err := vcs.RepoRootForImportPath(dep.Path, false)
	if err != nil {
		panic(err)
	}

	rev := dep.Version
	match := pseudoVersionRegex.FindStringSubmatch(rev)
	if len(match) > 0 {
		rev = match[1]
	}

	prefetchOut, err := exec.Command("nix-prefetch-git", "--quiet", repoRoot.Repo, "--rev", rev).Output()
	if err != nil {
		panic(err)
	}

	var prefetch Prefetch
	json.Unmarshal([]byte(prefetchOut), &prefetch)

	return Package{
		GoPackagePath: dep.Path,
		URL:           prefetch.URL,
		Sha256:        prefetch.Sha256,
		Rev:           prefetch.Rev,
	}
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

	deps := DepsForPath(path)
	queue := make(chan Package)

	wg := sync.WaitGroup{}
	wg.Add(len(deps))
	go func() {
		wg.Wait()
		close(queue)
	}()

	for dep, _ := range deps {
		go func(dep Dependency) {
			defer wg.Done()
			queue <- PrefetchDependency(dep)
		}(dep)
	}

	fmt.Print("[")
	for dep := range queue {
		fmt.Print(dep)
	}
	fmt.Println("]")
}

// godeps is a small tool to generate a list of dependencies for a set of Go packages.
// The output format can directly be integrated in a Makefile rule.
//
// Usage:
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

var usage = `
Usage:
	godeps [-flags FLAGS] [-pkgdir DIR] [-include-tests] PACKAGE [PACKAGE …]

Options:
	-flags            Build flags to consider (as given to "go build" command)
	-pkgdir           Root of the package (if multi-module build)
	-include-tests    Also generate dependencies for test files (usually not required)

The most common error for this package is the "godeps only accepts main package".
If you have given an actual main package name, this is usually because the package cannot
be found. Make sure that pkgdir is set to the right directory, this can be checked by:

	go list -f '{{.Name}}' PACKAGE
`

func main() {
	buildFlags := []string{}
	pkgDir, _ := os.Getwd()

	flag.Func("flags", "build flags to include", func(s string) error {
		buildFlags = strings.Split(s, ",")
		return nil
	})

	flag.StringVar(&pkgDir, "pkgdir", "", "Load packages from dir instead of current directory")
	flag.Bool("include-tests", false, "Include related test packages")
	outspec := flag.String("o", "-", "Destination of the dependencies (stdout by default)")

	flag.Parse()

	ctx := context.Background()

	pkgDir, _ = filepath.Abs(pkgDir)

	dst := os.Stdout
	if *outspec != "-" {
		var err error
		dst, err = os.Create(*outspec)
		if err != nil {
			log.Fatalf("creating output %s: %s", *outspec, err)
		}
	}

	cfg := packages.Config{
		Context:    ctx,
		Dir:        pkgDir,
		Mode:       packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps | packages.NeedModule,
		BuildFlags: buildFlags,
	}

	pkgs, err := packages.Load(&cfg, flag.Args()...)
	if err != nil {
		log.Fatal("error loading packages", flag.Args(), err)
	}

	for _, p := range pkgs {
		if p.Name != "main" {
			log.Fatalf("godeps only accepts main packages [ran in %s]: got %s", pkgDir, p.Name)
		}
	}
	for _, p := range pkgs {
		fmt.Fprint(dst, p.Module.GoMod, " ")
	}

	cmod := pkgs[0].Module.Path

	packages.Visit(pkgs, func(p *packages.Package) bool {
		if p.Module == nil || p.Module.Path != cmod {
			return false
		}

		files := [][]string{p.GoFiles, p.EmbedFiles, p.OtherFiles}
		for _, fs := range files {
			if len(fs) > 0 {
				fmt.Fprintf(dst, "%s ", strings.Join(fs, " "))
			}
		}
		return true
	}, nil)
}

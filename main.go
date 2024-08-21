// godeps is a small tool to generate a list of dependencies for a set of Go packages.
// The output format can directly be integrated in a Makefile rule.
package main

// need package name to binary mapping
//
// All sources files must be in the same directory.
// => update to directory or files in directory -> re-generate dependencies
//
// https://go.dev/ref/spec#Package_clause

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/tools/go/packages"
)

func main() {
	timeout := 10 * time.Second
	buildFlags := []string{}
	pkgDir, _ := os.Getwd()

	flag.Var((*FDuration)(&timeout), "load-timeout", "Maximum time to wait before killing the tool")
	flag.Var((*CSV)(&buildFlags), "tags", "Build tags to include")
	flag.StringVar(&pkgDir, "pkgdir", "", "Load packages from dir instead of current directory")
	flag.Bool("include-tests", false, "Include related test packages")

	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pkgDir, _ = filepath.Abs(pkgDir)

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
			log.Fatal("only run godeps on top-level (main) packages")
		}
	}

	for _, p := range pkgs {
		fmt.Printf(".INTERMEDIATE: %s\n", p.PkgPath)
	}

	// import and reversed dependencies
	//  import name -> files in package
	ideps := make(map[string][]string)
	//  import name -> packages importing it
	rdeps := make(map[string][]string)

	fmt.Println() // empty line

	for _, p := range pkgs {
		rpath(pkgDir, p)
		fmt.Printf("%s: %s\n", p.PkgPath, strings.Join(p.GoFiles, " "))
		for _, dep := range p.Imports {
			// only include dependencies in current modules, go.mod does the rest
			if dep.Module == nil || dep.Module.Path != p.Module.Path {
				continue
			}
			if _, known := ideps[dep.PkgPath]; !known {
				rpath(pkgDir, dep)
				ideps[dep.PkgPath] = dep.GoFiles
			}
			rdeps[dep.PkgPath] = append(rdeps[dep.PkgPath], p.PkgPath)
		}
	}

	for n, p := range rdeps {
		fmt.Printf("%s: %s\n", strings.Join(p, " "), strings.Join(ideps[n], " "))
	}
}

// rewrite files path so they are relative to base
func rpath(base string, p *packages.Package) {
	for i, f := range p.GoFiles {
		p.GoFiles[i], _ = filepath.Rel(base, f)
	}
}

// godeps is a small tool to generate a list of dependencies for a set of Go packages.
// The output format can directly be integrated in a Makefile rule.
//
// Usage:
//
//	godeps [-tags go build tags] [-pkgdir directory] [-include-tests] package package …
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

func main() {
	buildFlags := []string{}
	pkgDir, _ := os.Getwd()

	flag.Var((*CSV)(&buildFlags), "tags", "Build tags to include")
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
			log.Fatalf("godeps only accepts main packages [ran in %s]", p)
		}
	}

	for _, p := range pkgs {
		fmt.Fprintf(dst, ".INTERMEDIATE: %s\n", p.PkgPath)
	}

	fmt.Fprintln(dst)

	for _, p := range pkgs {
		rp, _ := filepath.Rel(pkgDir, p.Module.GoMod)
		fmt.Fprintf(dst, "%s: %s\n", p.PkgPath, rp)
	}

	// import and reversed dependencies
	//  import name -> files in package
	ideps := make(map[string][]string)
	//  import name -> packages importing it
	rdeps := make(map[string][]string)

	fmt.Fprintln(dst) // empty line

	for _, p := range pkgs {
		rpath(pkgDir, p)
		fmt.Fprintf(dst, "%s: %s\n", p.PkgPath, strings.Join(p.GoFiles, " "))
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
		fmt.Fprintf(dst, "%s: %s\n", strings.Join(p, " "), strings.Join(ideps[n], " "))
	}
}

// rewrite files path so they are relative to base
func rpath(base string, p *packages.Package) {
	for i, f := range p.GoFiles {
		p.GoFiles[i], _ = filepath.Rel(base, f)
	}
}

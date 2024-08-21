# Go dependencies for Makefiles

When using Go as part of a multi-step, multi-languages project, a naive makefile calling `go build` in rules without prerequisite is wasteful: each invocation of `go build` creates a new binary, which in turn triggers all dependent steps (e.g. binary compression, inclusion in a disk image, …).

Instead, we need to inform `make` about our precise dependencies to only trigger the build when needed.

This tool is a small script which, given a list of main packages, creates a correct and minimal dependency tree, ready for inclusion in a larger Makefile.

## Example Makefile

Running from the directory containing both prog1 and prog2:

```
PACKAGES := example.com/prog1 example.com/prog2

deps.mkd: $(wildcard **/*.go)
	godeps $(PACKAGES) > deps.mkd

include deps.mkd

bin/prog1: example.com/prog1 go.mod
	go build $@ $<

bin/prog2: example.com/prog2 go.mod
	go build $@ $<

```

A few things of note in this makefile:
 - packages are defined using a fully qualified name (so no conflict)
 - dependencies are re-generated whenever any Go file changes (tracking possible import changes)
 - module versioning is handled through the `go` command-line tool for accuracy

## Rationale on the generated makefile

The output is a set of rules defining dependencies between Go packages and files on disk:
https://www.gnu.org/software/make/manual/html_node/Multiple-Rules.html

If a dependency is included twice, a single rule is created for both targets:
https://www.gnu.org/software/make/manual/html_node/Multiple-Targets.html

In order to use Go packages as rules (without a corresponding file on disk), they are created as intermediate rules (so the transitive dependencies are used to rebuild):
https://www.gnu.org/software/make/manual/html_node/Chained-Rules.html
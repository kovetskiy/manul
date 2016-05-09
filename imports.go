package main

import (
	"go/build"
	"os"
)

func recursiveParseImports(
	imports map[string]bool, path string, cwd string,
) error {
	if path == "C" {
		return nil
	}

	pkg, err := build.Import(path, cwd, build.IgnoreVendor)
	if err != nil {
		return err
	}

	if path != "." {
		standard := false

		// this condition copied from cmd/go/pkg.go
		if pkg.Goroot && pkg.ImportPath != "" {
			standard = true
		}

		imports[pkg.ImportPath] = standard
	}

	for _, importing := range pkg.Imports {
		_, ok := imports[importing]
		if !ok {
			err = recursiveParseImports(imports, importing, cwd)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func parseImports() ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	imports := map[string]bool{}
	err = recursiveParseImports(imports, ".", cwd)
	if err != nil {
		return nil, err
	}

	notstandard := []string{}
	for pkg, standard := range imports {
		if !standard {
			notstandard = append(notstandard, pkg)
		}
	}

	return notstandard, nil
}

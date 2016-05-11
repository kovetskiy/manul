package main

import (
	"go/build"
	"os"
	"sort"
	"strings"
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

		if strings.HasPrefix(pkg.ImportPath, "golang.org/") ||
			(pkg.Goroot && pkg.ImportPath != "") {
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

	var (
		allImports = map[string]bool{}
		imports    = []string{}
	)

	err = recursiveParseImports(allImports, ".", cwd)
	if err != nil {
		return nil, err
	}

	for importing, standard := range allImports {
		if !standard {
			importpath, err := getRootImportpath(importing)
			if err != nil {
				continue
			}

			found := false
			for _, imported := range imports {
				if importpath == imported {
					found = true
					break
				}
			}

			if found {
				continue
			}

			imports = append(imports, importpath)
		}
	}

	sort.Strings(imports)

	return imports, nil
}

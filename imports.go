package main

import (
	"go/build"
	"log"
	"os"
	"path/filepath"
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

func parseImports(recursive bool) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	var (
		allImports = map[string]bool{}
		imports    = []string{}
	)

	filepath.Walk(
		cwd, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if filepath.Base(path) == ".git" ||
				filepath.Dir(path) == filepath.Join(cwd, "vendor") {
				return filepath.SkipDir
			}

			if path == filepath.Join(cwd, "vendor") {
				return nil
			}

			if !info.IsDir() {
				return nil
			}

			err = recursiveParseImports(
				allImports,
				".",
				path,
			)
			if _, ok := err.(*build.NoGoError); ok {
				return nil
			}

			if err != nil {
				log.Println(err)
			}

			return nil
		},
	)

	for importing, standard := range allImports {
		if !standard {
			importpath, err := getRootImportpath(importing)
			if err != nil {
				continue
			}

			if isOwnPackage(importpath, cwd) {
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

func isOwnPackage(path, cwd string) bool {
	for _, gopath := range filepath.SplitList(os.Getenv("GOPATH")) {
		if strings.HasPrefix(filepath.Join(gopath, "src", path), cwd) {
			return true
		}
	}
	return false
}

package main

import (
	"encoding/json"
	"errors"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/reconquest/karma-go"
)

type golistOutput struct {
	Imports      []string
	Deps         []string
	TestImports  []string
	XTestImports []string
}

func parseImports(recursive bool, testDependencies bool) ([]string, error) {
	var imports []string
	packages, err := listPackages()
	if err != nil {
		return imports, karma.Format(
			err, "unable to list packages",
		)
	}

	// Ensuring our dependencies exists isn't a strict requirement, therefore
	// only print a message to stderr rather then completely failing.

	err = ensureDependenciesExist(packages, true)
	if err != nil {
		logger.Warning(err)
	}

	imports, err = calculateDependencies(packages, recursive, testDependencies)
	if err != nil {
		return imports, err
	}

	imports = filterPackages(imports, build.IgnoreVendor)
	sort.Strings(imports)
	return imports, nil
}

func calculateDependencies(
	packages []string,
	recursive,
	testDependencies bool,
) ([]string, error) {
	var deps, testImports, testDeps []string

	data, err := golist(packages...)
	if err != nil {
		return nil, karma.Format(
			err, "unable to list dependencies for some packages in: %v", packages,
		)
	}

	for _, pkgMetaData := range data {
		if recursive {
			deps = append(deps, pkgMetaData.Deps...)
		} else {
			deps = append(deps, pkgMetaData.Imports...)
		}

		if testDependencies {
			testImports = append(testImports, pkgMetaData.TestImports...)
			testImports = append(testImports, pkgMetaData.XTestImports...)
		}
	}

	testImports = unique(testImports)

	if recursive {
		testData, err := golist(testImports...)
		if err != nil {
			return nil, karma.Format(
				err, "unable to list dependencies for some test packages in: %v", testImports)
		}

		for _, pkgMetaData := range testData {
			testDeps = append(testDeps, pkgMetaData.Deps...)
		}

		deps = append(deps, testDeps...)
	}

	deps = append(deps, testImports...)
	deps = unique(deps)

	return deps, nil
}

func filterPackages(packages []string, mode build.ImportMode) []string {
	var imports []string

	for _, importing := range packages {
		if importing == "C" {
			continue
		}

		pkg, err := build.Import(importing, "", mode)
		if err != nil {
			continue
		}

		// must skip packages in GOROOT dirs
		if pkg.Goroot {
			continue
		}

		importpath, err := getRootImportpath(pkg, importing)
		if err != nil {
			continue
		}

		if testing {
			importpath = strings.Replace(
				importpath, "__blankd__", "localhost:60001", -1,
			)
		}

		if isOwnPackage(importpath) {
			continue
		}

		// Skip anything that is already vendored upstream
		if strings.Contains(importpath, "/vendor/") {
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

	return imports
}

func ensureDependenciesExist(packages []string, withTests bool) error {
	if packages == nil {
		return errors.New("packages list cannot be empty")
	}

	args := []string{"get", "-d"} // -d for download only

	if withTests {
		args = append(args, "-t")
	}

	for _, pkg := range packages {
		args = append(args, pkg)
	}

	_, err := execute(exec.Command("go", args...))
	if err != nil {
		return karma.Format(
			err,
			"unable to go get dependencies for one of the packages in %v",
			packages,
		)
	}

	return nil
}

func golist(packages ...string) ([]golistOutput, error) {
	result := []golistOutput{}

	if packages == nil {
		return result, errors.New("packages list cannot be empty")
	}

	args := []string{"list", "-e", "-json"}

	for _, pkg := range packages {
		args = append(args, pkg)
	}

	jsonStream, err := execute(exec.Command("go", args...))
	if err != nil {
		return result, err
	}

	decoder := json.NewDecoder(strings.NewReader(jsonStream))

	for {
		pkgMetaData := golistOutput{}

		err := decoder.Decode(&pkgMetaData)
		if err != nil {
			return result, karma.Format(err, "failed to decode go list output")
		}

		result = append(result, pkgMetaData)

		if !decoder.More() {
			break
		}
	}

	return result, nil
}

func listPackages() ([]string, error) {
	var packages []string

	out, err := execute(exec.Command("go", "list", "-e", "./..."))
	if err != nil {
		return packages, err
	}

	for _, pkg := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.Contains(pkg, "/vendor/") {
			continue
		}

		packages = append(packages, pkg)
	}

	return packages, nil
}

func isOwnPackage(path string) bool {
	// workdir is without trailing slash, to make sure we don't exclude
	// packages that have the same folder prefix, suffix with dir separator
	wdir := workdir + string(os.PathSeparator)

	for _, gopath := range filepath.SplitList(os.Getenv("GOPATH")) {
		p := filepath.Join(gopath, "src", path)

		// this is an own package if the path without separator is equal to path
		// or when path has the same prefix with separator
		if p == workdir || strings.HasPrefix(p, wdir) {
			return true
		}
	}

	return false
}

func unique(input []string) []string {
	var (
		list  = make([]string, 0, len(input))
		table = make(map[string]bool)
	)

	for _, value := range input {
		if _, ok := table[value]; !ok {
			table[value] = true
			list = append(list, value)
		}
	}

	return list
}

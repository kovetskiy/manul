package main

import (
	"encoding/json"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/reconquest/karma-go"
)

type golistOutput struct {
	Imports     []string
	Deps        []string
	TestImports []string
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
	err = ensureDependenciesExist(packages, testDependencies)
	if err != nil {
		logger.Error(err)
	}

	imports, err = calculateDependencies(packages, recursive, testDependencies)
	if err != nil {
		return imports, err
	}

	imports = filterPackages(imports)

	sort.Strings(imports)

	return imports, nil
}

func calculateDependencies(packages []string, recursive,
	testDependencies bool) ([]string, error) {
	var deps, imports, testImports, testDeps []string

	for _, pkg := range packages {
		data, err := golist(pkg)
		if err != nil {
			return imports, karma.Format(
				err, "unable to list dependecies for package: %s", pkg,
			)
		}

		if recursive {
			deps = append(deps, data.Deps...)
		} else {
			deps = append(deps, data.Imports...)
		}

		if testDependencies {
			testImports = append(testImports, data.TestImports...)
		}
	}

	testImports = unique(testImports)
	if recursive {
		for _, testImport := range testImports {
			testData, err := golist(testImport)
			if err != nil {
				return imports, karma.Format(
					err, "unable to list dependencies for package: %s",
					testImport,
				)
			}

			testDeps = append(testDeps, testData.Deps...)
		}

		deps = append(deps, testDeps...)
	}

	deps = append(deps, testImports...)
	deps = unique(deps)

	return deps, nil
}

func filterPackages(packages []string) []string {
	var imports []string

	for _, importing := range packages {
		pkg, err := build.Import(importing, "", build.IgnoreVendor)
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
	args := []string{"get", "-d"} // -d for download only

	if withTests {
		args = append(args, "-t")
	}

	args = append(args, packages...)

	_, err := execute(exec.Command("go", args...))
	if err != nil {
		return karma.Format(
			err,
			"unable to go get dependencies: %s",
			strings.Join(packages, ", "),
		)
	}

	return nil
}

func golist(pkg string) (golistOutput, error) {
	data := golistOutput{}

	out, err := execute(exec.Command("go", "list", "-e", "-json", pkg))
	if err != nil {
		return data, err
	}

	err = json.Unmarshal([]byte(out), &data)
	if err != nil {
		return data, karma.Format(err, "unable to decode `go list` JSON output")
	}

	return data, nil
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
	for _, gopath := range filepath.SplitList(os.Getenv("GOPATH")) {
		if strings.HasPrefix(filepath.Join(gopath, "src", path), workdir) {
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

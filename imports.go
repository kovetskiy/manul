package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type listOutput struct {
	Imports     []string
	Deps        []string
	TestImports []string
}

func parseImports(recursive bool, testDependencies bool) ([]string, error) {
	var imports []string
	packages, err := listPackages()
	if err != nil {
		return imports, fmt.Errorf("error listing packages: %s", err)
	}

	// Ensuring our dependencies exists isn't a strict requirement, therefore
	// only print a message to stderr rather then completely failing.
	err = ensureDependenciesExist(packages, testDependencies)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
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
		data, err := list(pkg)
		if err != nil {
			return imports,
				fmt.Errorf("failed to list dependencies for %s: %s", pkg, err)
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
			testData, err := list(testImport)
			if err != nil {
				return imports, fmt.Errorf(
					"failed to list dependencies for testing package %s: %s",
					testImport, err)
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
		importpath, err := getRootImportpath(importing)
		if err != nil {
			continue
		}

		if inTests {
			importpath = strings.Replace(importpath,
				"__blankd__", "localhost:60001", -1)
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

func ensureDependenciesExist(packages []string, includeTestDeps bool) error {
	args := []string{"get"}
	if includeTestDeps {
		args = append(args, "-t")
	}
	args = append(args, packages...)

	out, err := execute(exec.Command("go", args...))
	if err != nil {
		return fmt.Errorf("can't 'go get' dependencies:\n%s", out)
	}

	return nil
}

func list(pkg string) (listOutput, error) {
	data := listOutput{}

	out, err := execute(exec.Command("go", "list", "-e", "-json", pkg))
	if err != nil {
		return data, err
	}

	err = json.Unmarshal([]byte(out), &data)
	if err != nil {
		return data,
			fmt.Errorf("can't unmarshal go list JSON output: %s\n%q", err, out)
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
	uniqueList := make([]string, 0, len(input))
	uniqueMap := make(map[string]bool)

	for _, val := range input {
		if _, ok := uniqueMap[val]; !ok {
			uniqueMap[val] = true
			uniqueList = append(uniqueList, val)
		}
	}

	return uniqueList
}

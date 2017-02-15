package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func parseImports(recursive bool, testDependencies bool) ([]string, error) {
	var imports []string
	packages, err := listPackages()
	if err != nil {
		return imports, fmt.Errorf("error listing packages: %s", err)
	}

	ensureDependenciesExist(packages, testDependencies)

	imports, err = calculateDependencies(packages, recursive, testDependencies)
	if err != nil {
		return imports, err
	}

	imports = filterPackages(imports)

	sort.Strings(imports)

	return imports, nil

}

func calculateDependencies(packages []string, recursive, testDependencies bool) ([]string, error) {
	var deps, imports, testImports, testDeps []string

	for _, pkg := range packages {
		data, err := list(pkg)
		if err != nil {
			return imports, nil
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
		for _, i := range testImports {
			testData, err := list(i)
			if err != nil {
				return imports, nil
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
	cwd, _ := os.Getwd()

	for _, importing := range packages {
		importpath, err := getRootImportpath(importing)
		if err != nil {
			continue
		}

		if inTests {
			importpath = strings.Replace(importpath, "__blankd__", "localhost:60001", -1)
		}

		if isOwnPackage(importpath, cwd) {
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

func ensureDependenciesExist(packages []string, includeTestingDependencies bool) {
	args := []string{"get"}
	if includeTestingDependencies {
		args = append(args, "-t")
	}
	args = append(args, packages...)

	out, err := execute(exec.Command("go", args...))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", out)
	}
}

func list(pkg string) (listOutput, error) {
	data := listOutput{}

	out, err := execute(exec.Command("go", "list", "-e", "-json", pkg))
	if err != nil {
		return data, err
	}

	err = json.Unmarshal([]byte(out), &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func listPackages() ([]string, error) {
	var packages []string

	out, err := execute(exec.Command("go", "list", "-e", "./..."))
	if err != nil {
		return packages, err
	}

	r := strings.NewReader(out)
	s := bufio.NewScanner(r)
	for s.Scan() {
		pkg := s.Text()
		if strings.Contains(pkg, "/vendor/") {
			continue
		}

		packages = append(packages, pkg)
	}

	return packages, nil
}

func isOwnPackage(path, cwd string) bool {
	for _, gopath := range filepath.SplitList(os.Getenv("GOPATH")) {
		if strings.HasPrefix(filepath.Join(gopath, "src", path), cwd) {
			return true
		}
	}
	return false
}

func unique(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}

type listOutput struct {
	Imports     []string
	Deps        []string
	TestImports []string
}

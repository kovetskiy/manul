package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/docopt/docopt.go"
)

const (
	version = `manul 1.0`
	usage   = version + `

manul it is utility for vendoring dependencies using git submodule technology.

Usage:
    manul -S | --sync
    manul -C | --clean
    manul -U | --update <dependency>...
    manul -Q | --query
    manul -h | --help
    manul -v | --version

Options:
    -S --sync     Find all dependencies and add git submodule into vendor directory.
    -C --clean    Find all unused vendor submodules and remove it.
    -U --update   Update specified already-vendored dependencies.
                      If you don't specify any vendored dependency, manul will
                      update all already-vendored dependencies.
    -Q --query    List all vendored dependencies.
    -h --help     Show help message.
    -v --version  Show version.
`
)

func main() {
	args, err := docopt.Parse(usage, nil, true, version, true, true)
	if err != nil {
		panic(err)
	}

	var (
		modeSync   = args["--sync"].(bool)
		modeClean  = args["--clean"].(bool)
		modeUpdate = args["--update"].(bool)
		modeQuery  = args["--query"].(bool)
	)

	imports := []string{}
	if modeSync || modeClean {
		imports, err = parseImports()
		if err != nil {
			log.Fatal(err)
		}
	}

	submodules, err := getVendorSubmodules()
	if err != nil {
		log.Fatal(err)
	}

	switch {
	case modeSync:
		err = handleSync(imports, submodules)

	case modeClean:
		err = handleClean(imports, submodules)

	case modeUpdate:
		imports, _ := args["<dependency>"].([]string)
		if len(imports) == 0 {
			imports, err = parseImports()
			if err != nil {
				log.Fatal(err)
			}
		}

		err = handleUpdate(submodules, imports)

	case modeQuery:
		err = handleQuery(submodules)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func handleSync(imports []string, submodules map[string]string) error {
	added := 0
	for _, importpath := range imports {
		found := false
		for submodule, _ := range submodules {
			if importpath == submodule {
				found = true
				break
			}
		}

		if found {
			continue
		}

		log.Printf("adding submodule for %s", importpath)

		err := addVendorSubmodule(importpath)
		if err != nil {
			return err
		}

		added++
	}

	if added > 0 {
		fmt.Printf("added %d submodules\n", added)
	} else {
		fmt.Printf("all dependencies already vendored\n")
	}

	return nil
}

func handleClean(imports []string, submodules map[string]string) error {
	removed := 0
	for submodule, _ := range submodules {
		found := false
		for _, importpath := range imports {
			if importpath == submodule {
				found = true
				break
			}
		}

		if found {
			continue
		}

		log.Printf("removing unused vendor %s", submodule)

		err := removeVendorSubmodule(submodule)
		if err != nil {
			return err
		}

		removed++
	}

	if removed > 0 {
		fmt.Printf("removed %d unused vendors\n", removed)
	} else {
		fmt.Printf("nothing to remove\n")
	}

	return nil
}

func handleUpdate(submodules map[string]string, imports []string) error {
	updated := 0
	for _, importpath := range imports {
		if _, ok := submodules[importpath]; !ok {
			return fmt.Errorf("unexpected dependency %s", importpath)
		}

		log.Printf("updating dependency %s", importpath)
		err := updateVendorSubmodule(importpath)
		if err != nil {
			return err
		}

		updated++
	}

	if updated > 0 {
		fmt.Printf("updated %d dependencies\n", updated)
	} else {
		fmt.Printf("nothing to update\n")
	}

	return nil
}

func handleQuery(submodules map[string]string) error {
	maxlength := 0
	for submodule, _ := range submodules {
		length := len(submodule)
		if length > maxlength {
			maxlength = length
		}
	}

	format := "%-" + strconv.Itoa(maxlength) + "s %s\n"

	for submodule, commit := range submodules {
		fmt.Printf(format, submodule, commit)
	}

	return nil
}

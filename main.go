package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/docopt/docopt-go"
)

const (
	version = `manul 1.0`
	usage   = version + `

manul is the tool for vendoring dependencies using git submodule technology.

Usage:
    manul -S
    manul -C
    manul -U [<dependency>...]
    manul -Q [-o]
    manul -h
    manul -v

Options:
    -S --sync     Find all dependencies and add git submodule into the vendor
	              directory.
    -C --clean    Find all unused vendor submodules and remove it.
    -U --update   Update specified already vendored dependencies.
                      If you don't specify any vendored dependency, manul will
                      update all already vendored dependencies.
    -Q --query    List all dependencies.
        -o        List only already vendored dependencies.
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

	switch {
	case modeSync:
		err = handleSync()

	case modeClean:
		err = handleClean()

	case modeUpdate:
		imports, _ := args["<dependency>"].([]string)
		err = handleUpdate(imports)

	case modeQuery:
		onlyVendored := args["-o"].(bool)
		err = handleQuery(onlyVendored)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func handleSync() error {
	imports, err := parseImports()
	if err != nil {
		return err
	}

	submodules, err := getVendorSubmodules()
	if err != nil {
		return err
	}

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

func handleClean() error {
	imports, err := parseImports()
	if err != nil {
		return err
	}

	submodules, err := getVendorSubmodules()
	if err != nil {
		return err
	}

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

func handleUpdate(imports []string) error {
	var err error
	if len(imports) == 0 {
		imports, err = parseImports()
		if err != nil {
			return err
		}
	}

	submodules, err := getVendorSubmodules()
	if err != nil {
		return err
	}

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

func handleQuery(onlyVendored bool) error {
	submodules, err := getVendorSubmodules()
	if err != nil {
		return err
	}

	if onlyVendored {
		maxlength := getMaxLength(getKeys(submodules))
		format := "%-" + strconv.Itoa(maxlength) + "s %s\n"

		for submodule, commit := range submodules {
			fmt.Printf(format, submodule, commit)
		}
	} else {
		imports, err := parseImports()
		if err != nil {
			return err
		}

		maxlength := 0
		if len(submodules) > 0 {
			maxlength = getMaxLength(
				append(getKeys(submodules), imports...),
			)
		}

		vendoredFormat := "%-" + strconv.Itoa(maxlength) + "s  %s\n"

		for _, importpath := range imports {
			commit, vendored := submodules[importpath]
			if vendored {
				fmt.Printf(vendoredFormat, importpath, commit)
			} else {
				fmt.Println(importpath)
			}
		}

	}

	return nil
}

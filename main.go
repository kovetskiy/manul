package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/docopt/docopt-go"
)

const (
	version = `manul 1.0`
	usage   = version + `

manul is the tool for vendoring dependencies using git submodule technology.

Usage:
    manul -I [<dependency>...]
    manul -U [<dependency>...]
    manul -R [<dependency>...]
    manul -Q [-o]
    manul -C
    manul -h
    manul -v

Options:
    -I --install  Detect all dependencies and add git submodule into vendor directory.
                      If you don't specify any dependency, manul will
                      install all detected dependencies.
    -U --update   Update specified already-vendored dependencies.
                      If you don't specify any vendored dependency, manul will
                      update all already-vendored dependencies.
    -R --remove   Stop vendoring of specified dependencies.
                      If you don't specify any dependency, manul will
                      remove all vendored dependencies.
    -Q --query    List all dependencies.
        -o        List only already-vendored dependencies.
    -C --clean    Detect all unused vendored dependencies and remove it.
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
		modeInstall = args["--install"].(bool)
		modeUpdate  = args["--update"].(bool)
		modeRemove  = args["--remove"].(bool)
		modeQuery   = args["--query"].(bool)
		modeClean   = args["--clean"].(bool)

		dependencies, _ = args["<dependency>"].([]string)
	)

	switch {
	case modeInstall:
		err = handleInstall(dependencies)

	case modeUpdate:
		err = handleUpdate(dependencies)

	case modeQuery:
		onlyVendored := args["-o"].(bool)
		err = handleQuery(onlyVendored)

	case modeRemove:
		err = handleRemove(dependencies)

	case modeClean:
		err = handleClean()

	}

	if err != nil {
		log.Fatal(err)
	}
}

func handleInstall(dependencies []string) error {
	imports, err := parseImports()
	if err != nil {
		return err
	}

	installAll := len(dependencies) == 0
	if installAll {
		dependencies = imports
	}

	submodules, err := getVendorSubmodules()
	if err != nil {
		return err
	}

	added := 0
	for _, dependency := range dependencies {
		if !installAll {
			found := false
			for _, importpath := range imports {
				// there is HasPrefix for handling subpackages
				if strings.HasPrefix(importpath, dependency) {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("unknown dependency %s", dependency)
			}
		}

		vendored := false
		for submodule, _ := range submodules {
			if dependency == submodule {
				vendored = true
				break
			}
		}

		if vendored {
			log.Printf("skipping %s, already vendored", dependency)
			continue
		}

		log.Printf("adding submodule for %s", dependency)

		err := addVendorSubmodule(dependency)
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

func handleUpdate(dependencies []string) error {
	var err error
	if len(dependencies) == 0 {
		dependencies, err = parseImports()
		if err != nil {
			return err
		}
	}

	submodules, err := getVendorSubmodules()
	if err != nil {
		return err
	}

	updated := 0
	for _, importpath := range dependencies {
		if _, ok := submodules[importpath]; !ok {
			return fmt.Errorf("unknown dependency %s", importpath)
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

func handleRemove(dependencies []string) error {
	submodules, err := getVendorSubmodules()
	if err != nil {
		return err
	}

	removeAll := len(dependencies) == 0
	if removeAll {
		dependencies = getKeys(submodules)
	}

	for _, dependency := range dependencies {
		if !removeAll {
			_, found := submodules[dependency]
			if !found {
				return fmt.Errorf("unknown dependency %s", dependency)
			}
		}

		log.Printf("removing vendor %s", dependency)
		err := removeVendorSubmodule(dependency)
		if err != nil {
			return err
		}
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

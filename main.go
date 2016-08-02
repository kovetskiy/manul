package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/kovetskiy/godocs"
)

const (
	version = `manul 1.3`
	usage   = version + `

manul is the tool for vendoring dependencies using git submodule technology.

Usage:
    manul [options] -I [<dependency>...]
    manul [options] -U [<dependency>...]
    manul [options] -R [<dependency>...]
    manul [options] -Q [-o]
    manul [options] -C
    manul -h
    manul --version

Options:
    -I --install    Detect all dependencies and add git submodule into vendor directory.
                     If you don't specify any dependency, manul will
                     install all detected dependencies.
    -U --update     Update specified already-vendored dependencies.
                     If you don't specify any vendored dependency, manul will
                     update all already-vendored dependencies.
    -R --remove     Stop vendoring of specified dependencies.
                     If you don't specify any dependency, manul will
                     remove all vendored dependencies.
    -Q --query      List all dependencies.
        -o          List only already-vendored dependencies.
    -C --clean      Detect all unused vendored dependencies and remove it.
    -r --recursive  Be recursive.
    -h --help       Show help message.
    -v --verbose    Be verbose.
    --version       Show version.
`
)

var (
	verbose bool
	inTests bool
)

func init() {
	for i, arg := range os.Args {
		if arg == "--integration-test" {
			inTests = true
			os.Args = append(os.Args[:i], os.Args[i+1:]...)
		} else if arg == "--insecure-skip-verify" {
			http.DefaultClient.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
			os.Args = append(os.Args[:i], os.Args[i+1:]...)
		}
	}
}

func main() {
	args := godocs.MustParse(usage, version)

	var (
		modeInstall = args["--install"].(bool)
		modeUpdate  = args["--update"].(bool)
		modeRemove  = args["--remove"].(bool)
		modeQuery   = args["--query"].(bool)
		modeClean   = args["--clean"].(bool)

		dependencies, _ = args["<dependency>"].([]string)

		recursive = args["--recursive"].(bool)
	)

	verbose = args["--verbose"].(bool)

	var err error
	switch {
	case modeInstall:
		err = handleInstall(recursive, dependencies)

	case modeUpdate:
		err = handleUpdate(recursive, dependencies)

	case modeQuery:
		onlyVendored := args["-o"].(bool)
		err = handleQuery(recursive, onlyVendored)

	case modeRemove:
		err = handleRemove(dependencies)

	case modeClean:
		err = handleClean(recursive)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func handleInstall(recursive bool, dependencies []string) error {
	imports, err := parseImports(recursive)
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
		for submodule := range submodules {
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

func handleUpdate(recursive bool, dependencies []string) error {
	var err error
	if len(dependencies) == 0 {
		dependencies, err = parseImports(recursive)
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

func handleQuery(recursive, onlyVendored bool) error {
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
		imports, err := parseImports(recursive)
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

func handleClean(recursive bool) error {
	imports, err := parseImports(recursive)
	if err != nil {
		return err
	}

	submodules, err := getVendorSubmodules()
	if err != nil {
		return err
	}

	removed := 0
	for submodule := range submodules {
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

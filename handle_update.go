package main

import (
	"fmt"
	"strings"
)

func handleUpdate(
	recursive bool,
	withTests bool,
	dependencies []string,
) error {
	var err error
	if len(dependencies) == 0 {
		dependencies, err = parseImports(recursive, withTests)
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
		parts := strings.Split(importpath, "=")
		if len(parts) > 2 {
			return fmt.Errorf("too many `=` delimiters: %s", importpath)
		}

		var version string
		if len(parts) == 2 {
			importpath = parts[0]
			version = parts[1]
		}

		if _, ok := submodules[importpath]; !ok {
			return fmt.Errorf("unknown dependency %s", importpath)
		}

		if version != "" {
			logger.Infof("updating vendor submodule %s to %s", importpath, version)
		} else {
			logger.Infof("updating vendor submodule %s", importpath)
		}

		err := updateVendorSubmodule(importpath, version)
		if err != nil {
			return err
		}

		updated++
	}

	if updated > 0 {
		if updated == 1 {
			logger.Infof("updated 1 dependency submodule")
		} else {
			logger.Infof("updated %d dependencies submodules", updated)
		}
	} else {
		logger.Infof("nothing to update")
	}

	return nil
}

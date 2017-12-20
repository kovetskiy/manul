package main

import (
	"fmt"
	"strings"

	"github.com/reconquest/karma-go"
)

func handleInstall(recursive bool, withTests bool,
	dependencies []string) error {
	imports, err := parseImports(recursive, withTests)
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
		parts := strings.Split(dependency, "=")
		if len(parts) > 2 {
			return fmt.Errorf("too many `=` delimiters: %s", dependency)
		}

		var version string
		if len(parts) == 2 {
			dependency, version = parts[0], parts[1]
		}

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
				return fmt.Errorf("unknown dependency: %s", dependency)
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
			logger.Debugf("skipping %s, already vendored", dependency)
			continue
		}

		logger.Infof("adding submodule for %s", dependency)

		errs := addVendorSubmodule(dependency, version)
		if errs != nil {
			top := fmt.Errorf("unable to add submodule for %s", dependency)
			for _, err := range errs {
				top = karma.Push(top, err)
			}
			return top
		}

		added++
	}

	if added > 0 {
		if added == 1 {
			logger.Infof("added 1 submodule")
		} else {
			logger.Infof("added %d submodules", added)
		}
	} else {
		logger.Infof("all dependencies already vendored\n")
	}

	return nil
}

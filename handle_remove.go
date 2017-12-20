package main

import "fmt"

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

		logger.Infof("removing vendor %s", dependency)

		err := removeVendorSubmodule(dependency)
		if err != nil {
			return err
		}
	}

	return nil
}

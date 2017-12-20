package main

import "github.com/reconquest/karma-go"

func handleClean(recursive, withTests bool) error {
	imports, err := parseImports(recursive, withTests)
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

		logger.Infof("removing unused vendor submodule %s", submodule)

		err := removeVendorSubmodule(submodule)
		if err != nil {
			return karma.Format(
				err, "unable to remove vendor submodule: %s", submodule,
			)
		}

		removed++
	}

	if removed > 0 {
		if removed == 1 {
			logger.Infof("removed 1 unused vendor submodule")
		} else {
			logger.Infof("removed %d unused vendor submodules", removed)
		}
	} else {
		logger.Infof("nothing to remove")
	}

	return nil
}

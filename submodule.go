package main

import (
	"errors"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func getVendorSubmodules() (map[string]string, error) {
	output, err := execute(
		exec.Command("git", "submodule", "status"),
	)
	if err != nil {
		return nil, err
	}

	vendors := map[string]string{}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(strings.TrimLeft(line, " -"), " ")
		if len(parts) >= 2 {
			path := parts[1]
			commit := parts[0]
			if strings.HasPrefix(path, "vendor/") {
				path = strings.TrimPrefix(path, "vendor/")
				vendors[path] = commit
			}
		}
	}

	return vendors, nil
}

func addVendorSubmodule(importpath string) error {
	var (
		target   = "vendor/" + importpath
		prefixes = []string{
			"git://",
			"https://",
			"git+ssh://",
		}

		errs []string
	)

	for _, prefix := range prefixes {
		url := prefix + importpath

		_, err := execute(
			exec.Command("git", "submodule", "add", "-f", url, target),
		)
		if err == nil {
			return nil
		}

		errs = append(errs, err.Error())
	}

	return errors.New(strings.Join(errs, "\n"))
}

func removeVendorSubmodule(importpath string) error {
	vendor := "vendor/" + importpath

	_, err := execute(
		exec.Command("git", "submodule", "deinit", "-f", vendor),
	)
	if err != nil {
		return fmt.Errorf(
			"can't deinit submodule: %s", err,
		)
	}

	_, err = execute(
		exec.Command("git", "rm", vendor),
	)
	if err != nil {
		return fmt.Errorf(
			"can't remove submodule directory: %s", err,
		)
	}

	_, err = execute(
		exec.Command("rm", "-r", filepath.Join(".git", "modules", vendor)),
	)
	if err != nil {
		return fmt.Errorf(
			"can't remove submodule directory in .git/modules: %s", err,
		)
	}

	return nil
}

func updateVendorSubmodule(importpath string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "pull", "origin", "master")
	cmd.Dir = filepath.Join(cwd, "vendor", importpath)

	_, err = execute(cmd)

	return err
}

func getRootImportpath(importpath string) (string, error) {
	pkg, err := build.Import(importpath, "", build.IgnoreVendor)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = filepath.Join(pkg.SrcRoot, importpath)

	rootdir, err := execute(cmd)
	if err != nil {
		return "", err
	}

	return strings.Trim(
		strings.TrimSpace(strings.TrimPrefix(rootdir, pkg.SrcRoot)),
		"/",
	), nil
}

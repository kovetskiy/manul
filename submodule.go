package main

import (
	"errors"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// NOTE: This list is copied from
// https://github.com/golang/go/blob/10538a8f9e2e718a47633ac5a6e90415a2c3f5f1/src/cmd/go/vcs.go#L821-L861
var wellKnownSites = []string{
	"github.com/",
	"bitbucket.org/",
	"hub.jazz.net/git/",
	"git.apache.org/",
	"git.openstack.org/",
}

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
			"https://",
			"git+ssh://",
			"git://",
		}

		errs []string
	)

	for _, prefix := range prefixes {
		var url string
		if prefix == "https://" {
			var err error
			url, err = getHttpsURLForImportPath(importpath)
			if err != nil {
				errs = append(errs, err.Error())
				continue
			}
		} else {
			url = prefix + importpath
		}

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

func getHttpsURLForImportPath(importpath string) (url string, err error) {
	url = "https://" + importpath
	for _, site := range wellKnownSites {
		if strings.HasPrefix(importpath, site) {
			return url, nil
		}
	}

	// NOTE: Parse <meta name="go-import" content="import-prefix vcs repo-root">
	// For detail, see the output of "go help importpath"
	var doc *goquery.Document
	doc, err = goquery.NewDocument(url)
	if err != nil {
		return
	}
	doc.Find("meta[name=go-import]").Each(func(_ int, selection *goquery.Selection) {
		if err != nil {
			return
		}
		content, exists := selection.Attr("content")
		if !exists {
			err = fmt.Errorf(`"content" attribute not found in meta name="go-import" at %s`, url)
			return
		}
		terms := strings.Fields(content)
		if len(terms) != 3 {
			err = fmt.Errorf(`invalid formatted "content" attribute in meta name="go-import" at %s`, url)
			return
		}
		prefix := terms[0]
		vcs := terms[1]
		repoRoot := terms[2]
		if strings.HasPrefix(importpath, prefix) && vcs == "git" {
			url = repoRoot
		}
	})
	if err != nil {
		return "", err
	}
	return url, nil
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
		exec.Command("git", "rm", "--force", vendor),
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

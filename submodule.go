package main

import (
	"fmt"
	"go/build"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/reconquest/ser-go"
)

const tagMetaGoImport = "meta[name=go-import]"

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
		return nil, ser.Errorf(
			err, "unable to get submodules status",
		)
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

func addVendorSubmodule(importpath string, version string) []error {
	var (
		target   = "vendor/" + importpath
		prefixes = []string{
			"https://",
			"git+ssh://",
			"git://",
		}

		errs []error
	)

	for _, prefix := range prefixes {
		var url string
		if prefix == "https://" {
			var err error
			url, err = getHttpsURLForImportPath(importpath)
			if err != nil {
				errs = append(errs, err)
				continue
			}
		} else {
			url = prefix + importpath
		}

		_, err := execute(
			exec.Command("git", "submodule", "add", "-f", url, target),
		)
		if err == nil {
			if version != "" {
				_, err = execute(
					exec.Command("git", "-C", target, "checkout", version),
				)
				if err != nil {
					return []error{err}
				}
			}

			return nil
		}

		errs = append(errs, err)
	}

	return errs
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
	doc, err = goquery.NewDocument(url + "?go-get=1")
	if err != nil {
		return "", err
	}

	doc.Find(tagMetaGoImport).Each(func(_ int, selection *goquery.Selection) {
		if err != nil {
			return
		}

		content, ok := selection.Attr("content")
		if !ok {
			err = fmt.Errorf(
				`"content" attribute not found in `+
					`meta name="go-import" at %s`,
				url,
			)
			return
		}

		terms := strings.Fields(content)
		if len(terms) != 3 {
			err = fmt.Errorf(
				`invalid formatted "content" attribute in `+
					`meta name="go-import" at %s`, url,
			)
			return
		}

		var (
			prefix   = terms[0]
			vcs      = terms[1]
			repoRoot = terms[2]
		)

		if strings.HasPrefix(importpath, prefix) && vcs == "git" {
			url = repoRoot
		}
	})

	return url, err
}

func removeVendorSubmodule(importpath string) error {
	vendor := "vendor/" + importpath

	_, err := execute(
		exec.Command("git", "submodule", "deinit", "-f", vendor),
	)
	if err != nil {
		return ser.Errorf(
			err, "unable to deinit vendor submodule: %s", vendor,
		)
	}

	_, err = execute(
		exec.Command("git", "rm", "--force", vendor),
	)
	if err != nil {
		return ser.Errorf(
			err, "unable to remove vendor directory: %s", vendor,
		)
	}

	_, err = execute(
		exec.Command("rm", "-r", filepath.Join(".git", "modules", vendor)),
	)
	if err != nil {
		return ser.Errorf(
			err, "unable to remove .git/modules/%s directory", vendor,
		)
	}

	return nil
}

func updateVendorSubmodule(importpath string, version string) error {
	cwd := filepath.Join(workdir, "vendor", importpath)
	if version == "" {
		cmd := exec.Command(
			"git", "-C", cwd, "pull", "origin", "master")

		_, err := execute(cmd)
		return err
	}

	_, err := execute(exec.Command("git", "-C", cwd, "show", version))
	if err != nil {
		_, err := execute(exec.Command("git", "-C", cwd, "remote", "update"))
		if err != nil {
			return err
		}
	}

	_, err = execute(
		exec.Command("git", "-C", cwd, "checkout", version),
	)
	return err
}

func getRootImportpath(pkg *build.Package, importpath string) (string, error) {
	cmd := exec.Command(
		"git",
		"-C", filepath.Join(pkg.SrcRoot, importpath),
		"rev-parse", "--show-toplevel",
	)

	rootdir, err := execute(cmd)
	if err != nil {
		return "", err
	}

	vendorPath := strings.TrimPrefix(workdir, pkg.SrcRoot+"/") + "/vendor/"

	return strings.TrimPrefix(
		strings.Trim(
			strings.TrimSpace(
				strings.TrimPrefix(
					rootdir,
					pkg.SrcRoot,
				),
			), "/",
		),
		vendorPath,
	), nil
}

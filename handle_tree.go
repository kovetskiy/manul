package main

import (
	"fmt"
	"strings"

	"github.com/reconquest/karma-go"
)

type Tree struct {
	Package string
	Nested  []*Tree
}

func handleTree(withTests bool, usePath bool) error {
	packages, err := listPackages()
	if err != nil {
		return karma.Format(
			err,
			"unable to list packages",
		)
	}

	var inRoot bool
	if len(packages) > 1 {
		inRoot = true
		for i, pkg := range packages {
			if i == 0 {
				continue
			}

			if !strings.HasPrefix(pkg, packages[0]) {
				inRoot = false
				break
			}
		}
	}

	var root *Tree

	cache := map[string]*Tree{}
	for i, pkg := range packages {
		pkgTree := getTree(pkg, withTests, cache, usePath)

		if !inRoot {
			fmt.Println(formatTree(pkgTree))
		} else {
			if i == 0 {
				root = pkgTree
			} else {
				root.Nested = append(root.Nested, pkgTree)
			}
		}
	}

	if inRoot {
		fmt.Println(formatTree(root))
	}

	return nil
}

func formatTree(tree *Tree) karma.Reason {
	var root karma.Reason

	root = tree.Package

	if len(tree.Nested) == 0 {
		return root
	}

	var branches karma.Reason
	for i, subtree := range tree.Nested {
		var source karma.Reason
		if i == 0 {
			source = root
		} else {
			source = branches
		}

		branches = karma.Push(source, formatTree(subtree))
	}

	return branches
}

func getTree(
	pkg string,
	withTests bool,
	cache map[string]*Tree,
	usePath bool,
) *Tree {
	if cached, ok := cache[pkg]; ok {
		return cached
	}

	logger.Debugf("tree for %s", pkg)

	tree := &Tree{
		Package: pkg,
		Nested:  []*Tree{},
	}

	list, err := golist(pkg)
	if err != nil {
		logger.Fatal(err)
	}

	var imports []string
	var listing []string
	if usePath {
		listing = list.Deps
	} else {
		listing = list.Imports
	}

	for _, importing := range listing {
		imports = append(imports, removeVendorPrefix(importing))
	}

	imports = filterPackages(imports, 0)

	logger.Debugf("%s -> %s", pkg, imports)

	for _, imported := range imports {
		if strings.HasPrefix(imported, pkg) {
			continue
		}

		tree.Nested = append(
			tree.Nested,
			getTree(imported, withTests, cache, usePath),
		)
	}

	if withTests {
		testImports := filterPackages(list.TestImports, 0)
		for _, imported := range testImports {
			tree.Nested = append(
				tree.Nested,
				getTree(imported, withTests, cache, usePath),
			)
		}
	}

	cache[pkg] = tree

	return tree
}

func removeVendorPrefix(path string) string {
	index := strings.Index(path, "/vendor/")
	if index > 0 {
		return path[index+len("/vendor/"):]
	}

	return path
}

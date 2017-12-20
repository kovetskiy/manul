package main

import (
	"crypto/tls"
	"net/http"
	"os"

	"github.com/kovetskiy/godocs"
	"github.com/kovetskiy/lorg"
	"github.com/reconquest/hierr-go"
)

const (
	version = `manul 1.5`
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
                     You can specify commit-ish that will be used as target to
                     instal: -I golang.org/x/net=34a235h1
    -U --update     Update specified already-vendored dependencies.
                     If you don't specify any vendored dependency, manul will
                     update all already-vendored dependencies.
                     You can specify commit-ish that will be used as target to
                     update: -U golang.org/x/net=34a235h1
    -R --remove     Stop vendoring of specified dependencies.
                     If you don't specify any dependency, manul will
                     remove all vendored dependencies.
    -Q --query      List all dependencies.
        -o          List only already-vendored dependencies.
    -C --clean      Detect all unused vendored dependencies and remove it.
    -t --testing    Include dependencies from tests.
    -r --recursive  Be recursive.
    -h --help       Show help message.
    -v --verbose    Be verbose.
    --version       Show version.
`
)

var (
	verbose bool
	testing bool
	workdir string
	logger  = lorg.NewLog()
)

func init() {
	var err error

	newArgs := make([]string, 0, len(os.Args))
	for _, arg := range os.Args {
		if arg == "--integration-test" {
			testing = true
		} else if arg == "--insecure-skip-verify" {
			http.DefaultClient.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
		} else {
			newArgs = append(newArgs, arg)
		}
	}
	os.Args = newArgs

	workdir, err = os.Getwd()
	if err != nil {
		hierr.Fatalf(
			err,
			"unable to get current working directory",
		)
	}

	logger.SetIndentLines(true)
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

		recursive       = args["--recursive"].(bool)
		includeTestDeps = args["--testing"].(bool)

		verbose = args["--verbose"].(bool)
	)

	if verbose {
		logger.SetLevel(lorg.LevelDebug)
	}

	var err error
	switch {
	case modeInstall:
		err = handleInstall(recursive, includeTestDeps, dependencies)

	case modeUpdate:
		err = handleUpdate(recursive, includeTestDeps, dependencies)

	case modeQuery:
		onlyVendored := args["-o"].(bool)
		err = handleQuery(recursive, includeTestDeps, onlyVendored)

	case modeRemove:
		err = handleRemove(dependencies)

	case modeClean:
		err = handleClean(recursive, includeTestDeps)
	}

	if err != nil {
		logger.Fatal(err)
	}
}

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
	version = `manul 1.6`
	usage   = version + `

manul is the tool for vendoring dependencies using git submodule technology.

Usage:
    manul [options] -I [<dependency>...]
    manul [options] -U [<dependency>...]
    manul [options] -R [<dependency>...]
    manul [options] -Q [-o]
    manul [options] -C
    manul [options] -T
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
    -T --tree       Show dependencies tree.
    -t --testing    Include dependencies from tests.
    -r --recursive  Be recursive.
    -h --help       Show help message.
    -v --verbose    Be verbose.
    --trace         Be very verbose.
    --version       Show version.
`
)

var (
	verbose bool
	tracing bool
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
		dependencies, _ = args["<dependency>"].([]string)
		recursive       = args["--recursive"].(bool)
		withTests       = args["--testing"].(bool)
	)

	if args["--verbose"].(bool) {
		verbose = true
		logger.SetLevel(lorg.LevelDebug)
	}

	if args["--trace"].(bool) {
		logger.SetLevel(lorg.LevelTrace)
	}

	var err error
	switch {
	case args["--tree"].(bool):
		err = handleTree(withTests)

	case args["--install"].(bool):
		err = handleInstall(recursive, withTests, dependencies)

	case args["--update"].(bool):
		err = handleUpdate(recursive, withTests, dependencies)

	case args["--query"].(bool):
		onlyVendored := args["-o"].(bool)
		err = handleQuery(recursive, withTests, onlyVendored)

	case args["--remove"].(bool):
		err = handleRemove(dependencies)

	case args["--clean"].(bool):
		err = handleClean(recursive, withTests)
	}

	if err != nil {
		logger.Fatal(err)
	}
}

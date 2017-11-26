# MANUL [![Build Status](https://travis-ci.org/kovetskiy/manul.svg?branch=master)](https://travis-ci.org/kovetskiy/manul) [![License](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/kovetskiy/manul/master/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/kovetskiy/manul)](https://goreportcard.com/report/github.com/kovetskiy/manul)

![madness](https://cloud.githubusercontent.com/assets/8445924/10410421/ccca8b24-6f30-11e5-9952-9e5be5c4d792.png)

Manul is a vendoring utility for Go programs.

## Installation

**manul** can be obtained using `go get`:

```
go get github.com/kovetskiy/manul
```

Packages are also available for [Ubuntu/Debian][pkg-debian] and [Arch Linux][pkg-archlinux].

[pkg-debian]: https://github.com/kovetskiy/manul/tree/pkg-debian#readme
[pkg-archlinux]: https://github.com/kovetskiy/manul/tree/pkg-archlinux#readme

## What's the reason for yet another utility?

Because all other vendor utilities suffer from the following:

- Some wrap the `go` binary and spoof the `GOPATH` env variable.
    You will have a non-go-gettable project which needs additional 
    software in order to *compile and run*;

- Some copy the source code of dependencies into the vendor directory:
    * It will be nearly impossible to find anything using GitHub Search,
        because you will get many false-positive results;
    * Updating dependencies will require manual intervention and committing
        a lot of modified lines straight into the main repo;
    * You will not be able to tell what version of dependency your project is
        using will by looking at repository; you have to keep versions in the
        additional ambiguous file with vendors associated with commits.

- Various architecture problems:
    * Impossible to update all or specific vendored dependencies;
    * Impossible to rollback vendored dependencies to specific version;
    * Impossible to remove unused vendored dependencies;
    * Impossible to lock version of vendored dependency.

## Solution

We all love git, it's a very powerful instrument. Why don't we use its
power for vendoring dependencies using an awesome feature, which is called
**git submodule**?

With **git submodule** you will have a git repository for each dependency.
They can be managed in the same way as main project by `git`.

Pros:

- No need for additional software for building/running your Go project;

- No need for additional JSON/TOML/YAML file for storing dependencies;

- Update vendored dependencies directly from remote origins;

- Rollback changes in dependencies;

- Go-gettable

**git submodule** might look like a **Silver Bullet**, but it's still clumsy to
work with manually. We want to have a powerful yet simple interface for
vendoring dependencies using this technology.

**manul** can do it for us.

## Usage

### Who needs a documentation when there are GIFs?

First of all, we should request dependencies which we have in our project.
To do this, just run manul with `-Q` (query) flag. It will output all the
project imports (dependencies), like this:

![first query](https://cloud.githubusercontent.com/assets/8445924/10285714/9e840e76-6b79-11e5-821f-636729ce4467.gif)

For example, we have six dependencies, let's lock versions of critical
dependencies by adding submodules: in our case it's `zhash` and `blackfriday`
packages.

For locking versions (installing dependencies) we should use `-I` (install)
flag and specify dependencies, which we wish to install:

![install two dependencies](https://cloud.githubusercontent.com/assets/8445924/10285715/a0e85302-6b79-11e5-904f-051929fe472b.gif)

After installation we can have a look for vendored and non-vendored
dependencies by using flag `-Q`. After previous step we should see git commits
along with two already vendored dependencies (`zhash` and `blackfriday`):

![query after install](https://cloud.githubusercontent.com/assets/8445924/10285719/a39282e4-6b79-11e5-8877-7fba19e0d8c0.gif)

Let's install submodules for remaining dependencies, go the limit! Just run
**manul** with flag `-I` without specifying any dependencies, manul will
install all detected dependencies with skipping already vendored:

![install all dependencies](https://cloud.githubusercontent.com/assets/8445924/10285722/a63d1e6e-6b79-11e5-9f1e-1e606f3819dc.gif)

Wow, that was crazy! Now, to update some vendored dependencies, for example,
`docopt-go` package, manul should be invoked with the flag `-U` and import path
(`github.com/docopt/docopt-go`):

![update docopt](https://cloud.githubusercontent.com/assets/8445924/10285723/a8ce9f18-6b79-11e5-87ef-2caca393328c.gif)

**manul** can be used to remove specified submodules of vendored dependencies
by using `-R` (remove) flag and specifying dependencies import path:

![removing](https://cloud.githubusercontent.com/assets/8445924/10285727/ab587b50-6b79-11e5-9b5b-b7c7ff264506.gif)

By the way, manul can detect and remove unused vendored dependencies using `-C`
(clean) flag:

![unused dependencies](https://cloud.githubusercontent.com/assets/8445924/10285731/ae1d0270-6b79-11e5-9e97-151b7d77402a.gif)

**manul** can also get you a specific version of a dependency by using a commit-ish, for example:
- `manul -I golang.org/x/foo=34a235h1` will install `foo` at the specified commit
- `manul -U github.com/x/bar=this-tag` will update it to `this-tag` version.

Let's summarize:

- `-I [<dependency>...]` - install git submodules for specified/all dependencies;
- `-U [<dependency>...]` - update specified/all already vendored dependencies;
- `-R [<dependency>...]` - remove git submodules for specified/all dependencies;
- `-Q [<dependency>...]` - list all used dependencies;
- `-C` - detect and remove all git submodules for unused vendored dependencies.

You can see similar help message by passing `-h` or `--help` flag.

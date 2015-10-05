# MANUL

![madness](https://cloud.githubusercontent.com/assets/8445924/10410421/ccca8b24-6f30-11e5-9952-9e5be5c4d792.png)

Manul is the madness dependencies vendoring utility for Golang programs.

## Why yet another utility

Because all other vendor utilities sucks by following reasons:

- Wraps `go` binary and spoof `GOPATH` env variable.
    You will have not go-gettable project and should install
    additional software for *compile and run* project.

- Plain copying source code of dependencies to vendor directory.
    * You will not be able to find anything using GitHub Search, because you
        will get false-positive results.
    * After manually updating libraries you will get weird commits with
        fantastic line adding/removing counters.
    * You will not be able to say what version of dependency you use, or you
        will have additional ambigious file with associative of vendor and
        commit.

- Various architecture problems:
    * Can't update all or specific vendored dependencies
    * Can't rollback vendored dependencies changes
    * Can't remove unused vendored dependencies
    * Can't keep version of vendored dependency

## Solution

We are all love git, it's very powerful thing, so why don't we use power of
this thing for vendoring dependencies while there is awesome utility, which
called **git submodule**?

With **git submodule** you wil have a git repository for all of your
dependencies, which can be managed by same utility, which you main project
(git).

Features:
- You should not install any specific software for building/running your
   Golang project

- Apply changes from a remote repository into the vendored dependency
   repositroy

- Rollback changes

- *Almost go-gettable*, [Go have a little trouble with
    submodules](https://github.com/golang/go/issues/12573) and
     I will talk you how to bypass this.


Okay, **git submodule** looks like as **Silver Bullet**, so, we want to have a
powerful interface for vendoring dependencies using this technology, **manul**
can do it for us.

## Usage

### Who needs a documentation, when there are gifs?

First of all, we should have a look for dependencies which we have in our
project, for this run manul with `-Q` (query) flag, in output will be
all project dependencies, like this:

![first query](https://cloud.githubusercontent.com/assets/8445924/10285714/9e840e76-6b79-11e5-821f-636729ce4467.gif)

Okay, for example, we have six dependencies, let's add submodules for important
dependencies, in our case it's `zhash` and `blackfriday` packages.

For installation vendors we should use `-I` (install) flag with specifying
dependencies, which we wish to install.

![install two dependencies](https://cloud.githubusercontent.com/assets/8445924/10285715/a0e85302-6b79-11e5-904f-051929fe472b.gif)

After installation we can have a look for vendored or non-vendored dependencies
by using flag `-Q`. At this moment we should see git commits around two already
vendored dependencies (`zhash` and `blackfriday`).

![query after install](https://cloud.githubusercontent.com/assets/8445924/10285719/a39282e4-6b79-11e5-8877-7fba19e0d8c0.gif)

Let's install submodules for remaining dependencies, go the limit! Just run
**manul** with flag `-I` without specifying any dependencies, manul will
install all detected dependencies with skipping already-vendored.

![install all dependencies](https://cloud.githubusercontent.com/assets/8445924/10285722/a63d1e6e-6b79-11e5-9f1e-1e606f3819dc.gif)

Wow, it was crazy. Now update some vendored dependencies, for example,
`docopt-go` package, by running manul with flag `-U` and specifying import path
(`github.com/docopt/docopt-go`).

![update docopt](https://cloud.githubusercontent.com/assets/8445924/10285723/a8ce9f18-6b79-11e5-87ef-2caca393328c.gif)

**manul** can remove specified submodules of vendored dependencies using flag
`-R` (remove) and specifying dependencies import path.

![removing](https://cloud.githubusercontent.com/assets/8445924/10285727/ab587b50-6b79-11e5-9b5b-b7c7ff264506.gif)


By the way, manul can detect and remove unused vendored dependencies using `-C`
(clean) flag.

![unused dependencies](https://cloud.githubusercontent.com/assets/8445924/10285731/ae1d0270-6b79-11e5-9e97-151b7d77402a.gif)

Let's summarize.

- `-I [<dependency>...]` - install git submodules for specified or all dependencies
- `-U [<dependency>...]` - update specified or all already-vendored dependencies.
- `-R [<dependency>...]` - remove git submodules of specified or all
    dependencies.
- `-Q [<dependency>...]` - list all used dependencies.
- `-C` - detect and remove all git submodules of unused vendored dependencies.

You can see similar help message by passing `-h` or `--help` flag.

## What is it mean «almost go-gettable»

This means that while [that issue in
golang/go](https://github.com/golang/go/issues/12612) is not solved, you will
not be able to use `go get` with your project.

BTW, I think it's not trouble, because end-user of your project should not
install `manul` for compilation from source code.

Now projects, which uses the git submodules for vendoring dependencies, can be
builded by running this commands:
```bash
git clone --recursive <git-repository> $GOPATH/src/<project-name>
cd $GOPATH/src/<project-name>
go install
```

Although, I think it's bad practice when maintainers doesn't creates packages
for most popular distribs.

## Installation

#### Ubuntu/Debian:

```bash
git clone --branch pkg-debian git://github.com/kovetskiy/manul /tmp/manul
cd /tmp/manul
./build.sh
dpkg -i *.deb
```

#### Arch Linux:
```bash
git clone --branch pkg-archlinux git://github.com/kovetskiy/manul /tmp/manul
cd /tmp/manul
makepkg
pacman -U *.xz
```

Also package for Arch Linux available in the AUR:
https://aur4.archlinux.org/packages/manul

#### Other distros

**manul** can be builded using `go get`:

```
go get github.com/kovetskiy/manul
```

#!/bin/bash

set -e -u

SRCURL="github.com/kovetskiy/manul"
PKGDIR="manul-deb"
SRCROOT="src"
SRCDIR=${SRCROOT}/${SRCURL}

mkdir -p $PKGDIR/usr/bin
rm -rf $SRCROOT

export GOPATH=`pwd`
go get -v $SRCURL
pushd $SRCDIR
go build -o manul

count=$(git rev-list HEAD| wc -l)
commit=$(git rev-parse --short HEAD)
VERSION="${count}.$commit"
popd

sed -i 's/\$VERSION\$/'$VERSION'/g' $PKGDIR/DEBIAN/control

cp -f bin/manul $PKGDIR/usr/bin/manul

dpkg -b $PKGDIR manul-${VERSION}_amd64.deb

# restore version placeholder
git checkout $PKGDIR/DEBIAN/control

:project "main.go" <<GO
package main

import "github.com/kovetskiy/manul-test-foo"

func main() {
    bar.A()
}
GO

:lib "github.com/kovetskiy/manul-test-foo"

tests:ensure :manul -I github.com/kovetskiy/manul-test-foo
tests:assert-no-diff stdout <<VENDORS
added 1 submodules
VENDORS

tests:assert-stderr "adding submodule for github.com/kovetskiy/manul-test-foo"

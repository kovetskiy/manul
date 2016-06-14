:lib github.com/kovetskiy/manul-test-foo
:lib github.com/kovetskiy/manul-test-bar

:project "main.go" <<GO
package main

import "github.com/kovetskiy/manul-test-foo"
import "github.com/kovetskiy/manul-test-bar"

func main() {
    foo.Foo()
    bar.Bar()
}
GO

tests:ensure :manul -I

tests:ensure :manul -R
tests:assert-stderr "removing vendor github.com/kovetskiy/manul-test-bar"
tests:assert-stderr "removing vendor github.com/kovetskiy/manul-test-foo"

tests:ensure :manul -Q \| sort -n
tests:assert-no-diff stdout <<VENDORS
github.com/kovetskiy/manul-test-bar
github.com/kovetskiy/manul-test-foo
VENDORS

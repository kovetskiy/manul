:project "main.go" <<GO
package main

import "github.com/kovetskiy/manul-test-foo"
import "github.com/kovetskiy/manul-test-bar"
import "strings"

func main() {
    foo.Foo()
    bar.Bar()
    strings.Fields("aa")
}
GO

:lib "github.com/kovetskiy/manul-test-foo"
:lib "github.com/kovetskiy/manul-test-bar"

tests:ensure :manul -Q
tests:assert-no-diff stdout <<VENDORS
github.com/kovetskiy/manul-test-bar
github.com/kovetskiy/manul-test-foo
VENDORS

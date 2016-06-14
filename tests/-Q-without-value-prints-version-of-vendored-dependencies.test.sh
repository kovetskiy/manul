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

tests:ensure :manul -I github.com/kovetskiy/manul-test-foo
tests:ensure :manul -Q \| sort -n

tests:assert-no-diff stdout <<VENDORS
github.com/kovetskiy/manul-test-bar
github.com/kovetskiy/manul-test-foo  3c2b599a01a493064b9a9ea4e63f4b4fd99f0397
VENDORS

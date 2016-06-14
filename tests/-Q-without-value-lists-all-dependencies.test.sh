:project "main.go" <<GO
package main

import "github.com/kovetskiy/manul-test-foo"

func main() {
    bar.A()
}
GO

:lib "github.com/kovetskiy/manul-test-foo"


tests:ensure :manul -Q
tests:assert-no-diff stdout <<VENDORS
github.com/kovetskiy/manul-test-foo
VENDORS

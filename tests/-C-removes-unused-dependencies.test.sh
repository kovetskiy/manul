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

:project "main.go" <<GO
package main

import "github.com/kovetskiy/manul-test-bar"

func main() {
    bar.Bar()
}
GO

tests:ensure :manul -C

tests:ensure :manul -Q
tests:assert-no-diff stdout <<VENDORS
github.com/kovetskiy/manul-test-bar  9a5d4e050e8660fe7b616ce503e7c80a04e1e2db
VENDORS

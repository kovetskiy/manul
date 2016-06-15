:project "dir1/main.go" <<GO
package main

import "github.com/kovetskiy/manul-test-foo"

func main() {
    foo.A()
}
GO

:project "dir2/main.go" <<GO
package main

import "github.com/kovetskiy/manul-test-bar"
import "github.com/kovetskiy/manul-test-foo"

func main() {
    foo.A()
    bar.A()
}
GO

:lib "github.com/kovetskiy/manul-test-foo"
:lib "github.com/kovetskiy/manul-test-bar"


tests:ensure :manul -Q --recursive
tests:assert-no-diff stdout <<VENDORS
github.com/kovetskiy/manul-test-bar
github.com/kovetskiy/manul-test-foo
VENDORS

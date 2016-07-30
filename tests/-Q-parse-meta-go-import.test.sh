:project "main.go" <<GO
package main

import foo "manul-test-foo.appspot.com/a/b"

func main() {
    foo.Foo()
}
GO

# NOTE: See https://github.com/hnakamur/google-app-engine-manul-test-foo
# for source code of application manul-test-foo
:lib "manul-test-foo.appspot.com/a/b"

tests:ensure :manul -Q
tests:assert-no-diff stdout <<VENDORS
manul-test-foo.appspot.com/a/b
VENDORS

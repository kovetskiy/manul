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
tests:ensure :manul -Q \| sort -n

tests:assert-no-diff stdout <<VENDORS
github.com/kovetskiy/manul-test-bar  9a5d4e050e8660fe7b616ce503e7c80a04e1e2db
github.com/kovetskiy/manul-test-foo  3c2b599a01a493064b9a9ea4e63f4b4fd99f0397
VENDORS

tests:cd-tmp-dir go/src/project/vendor/github.com/kovetskiy/manul-test-bar/
tests:ensure git checkout db5bf508ab9ffad0e490c83555fec43d272e2b13

tests:ensure :manul -Q \| sort -n

tests:assert-no-diff stdout <<VENDORS
github.com/kovetskiy/manul-test-bar  +db5bf508ab9ffad0e490c83555fec43d272e2b13
github.com/kovetskiy/manul-test-foo  3c2b599a01a493064b9a9ea4e63f4b4fd99f0397
VENDORS

tests:ensure :manul -U
tests:ensure :manul -Q \| sort -n

tests:assert-no-diff stdout <<VENDORS
github.com/kovetskiy/manul-test-bar  9a5d4e050e8660fe7b616ce503e7c80a04e1e2db
github.com/kovetskiy/manul-test-foo  3c2b599a01a493064b9a9ea4e63f4b4fd99f0397
VENDORS

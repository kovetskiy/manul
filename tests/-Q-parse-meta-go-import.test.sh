_process=""

:project "main.go" <<GO
package main

import foo "__blankd__/kovetskiy/manul-test-foo"

func main() {
    bar.A()
}
GO

tests:put server <<SRV
#!/bin/bash

cat <<HTTP
200 OK

<meta name="go-import" content="localhost:60001/kovetskiy/manul-test-foo git https://github.com/kovetskiy/manul-test-foo" />
HTTP
SRV
tests:ensure chmod +x $(tests:get-tmp-dir)/server


:lib "github.com/kovetskiy/blankd"
:lib "github.com/kovetskiy/manul-test-foo"
tests:ensure  mv \
    $(tests:get-tmp-dir)/go/src/github.com \
    $(tests:get-tmp-dir)/go/src/__blankd__

tests:ensure $(tests:get-tmp-dir)/go/bin/blankd \
    -l localhost:60001 \
    -e $(tests:get-tmp-dir)/server \
    -o $(tests:get-tmp-dir)/blankd.log \
    --tls
tests:value _process cat $(tests:get-stdout-file)
:stop_blankd() {
    if [[ "$_process" ]]; then
        tests:eval kill "$_process"
    fi
}
trap :stop_blankd EXIT

tests:ensure :manul --integration-test --insecure-skip-verify -Q
tests:assert-no-diff stdout <<VENDORS
localhost:60001/kovetskiy/manul-test-foo
VENDORS

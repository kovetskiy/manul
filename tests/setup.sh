tests:clone manul.test bin/

tests:make-tmp-dir go go/pkg go/src go/src/project

:lib() {
    local name="$1"

    GOPATH=$(tests:get-tmp-dir)/go tests:ensure go get -v "$name"
}

:project() {
    local filename="$1"
    local _dir=$(dirname "$filename")
    tests:make-tmp-dir "go/src/project/$_dir"
    tests:put "go/src/project/$filename"

    tests:cd-tmp-dir "go/src/project/"
    tests:ensure git init
}

:manul() {
    cd $(tests:get-tmp-dir)/go/src/project

    GOPATH=$(tests:get-tmp-dir)/go tests:eval manul.test "$@"

    cat $(tests:get-stderr-file) >&2
    cat $(tests:get-stdout-file) >&1

    return $(tests:get-exitcode)
}

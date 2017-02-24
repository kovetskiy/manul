test:
	./tests/run_tests

manul:
	git submodule update --init --recursive
	go build

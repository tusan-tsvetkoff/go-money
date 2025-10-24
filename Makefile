.PHONY: test test-fuzz

PACKAGE ?= "./..."

test:
	go test -v -race ${PACKAGE}

FUZZ_TIME    ?= 30s
FUZZ_TEST    ?= ^$$
FUZZ_PACKAGE ?= ./parser

test-fuzz:
	go test -run=${FUZZ_TEST} -fuzz=Fuzz -fuzztime=${FUZZ_TIME} ${FUZZ_PACKAGE}

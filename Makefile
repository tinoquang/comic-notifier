export APP=notifier
export MODULE=comic-notifier

MAIN=cmd/main.go
BINARY=bin/${APP}
GEN=./pkg/api/

run: build start
.PHONY: run

re-build: clean run
.PHONY: re-build

build: ${APP}
.PHONY: build

${APP}:
	go build -v -o ${BINARY} ${MAIN}

start:
	./${BINARY}
.PHONY: start

gen: 
	go generate -x ${GEN}
.PHONY: gen

clean:
	rm -f ${BINARY}
.PHONY: clean

re-test: 
	go clean -testcache
	go test -v -cover ./...
.PHONY: re-test

test: 
	go test -v -cover ./...
.PHONY: test
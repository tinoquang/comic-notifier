export APP=notifier
export MODULE=comic-notifier

MAIN=cmd/main.go
BINARY=bin/${APP}
GEN=./pkg/api/

run: build start
.PHONY: run

build: ${APP}

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
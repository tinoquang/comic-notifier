include .env
export

MAIN=cmd/main.go
BINARY=bin/${APP}


run: build start
.PHONY: run

build: ${APP}

${APP}:
	go build -o ${BINARY} ${MAIN}

start:
	./${BINARY}
.PHONY: start



clean:
	rm -f ${BINARY}
.PHONY: clean
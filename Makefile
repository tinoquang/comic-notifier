include .env
export

MAIN=cmd/main.go
BINARY=bin/${APP}


build: ${APP}

${APP}:
	go build -o ${BINARY} ${MAIN}

start:
	./${BINARY}
.PHONY: start


run: build start
.PHONY: run

clean:
	rm -f ${BINARY}
.PHONY: clean
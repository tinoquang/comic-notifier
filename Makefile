include .env
export

MAIN=cmd/main.go
BINARY=bin/${APP}


build: ${APP}

${APP}:
	go build -o ${BINARY} ${MAIN}

run:
	./${BINARY}
.PHONY: run


restart: build run
.PHONY: restart

clean:
	rm -f ${BINARY}
.PHONY: clean
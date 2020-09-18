export APP=notifier
export MODULE=comic-notifier

MAIN=cmd/main.go
BINARY=bin/${APP}


run: build start
.PHONY: run

build: ${APP}

${APP}:
	go build -v -o ${BINARY} ${MAIN}

start:
	./${BINARY}
.PHONY: start



clean:
	rm -f ${BINARY}
.PHONY: clean
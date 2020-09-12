include .env
export

BINRARY=cmd/main.go



build: notifier

${APP}:
	go build -o $@ ${BINRARY}

clean:
	rm -f ${APP}
.PHONY: clean
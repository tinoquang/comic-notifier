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

local:
	docker-compose up --build
.PHONY: local

gen: 
	go generate -x ${GEN}
	sqlc generate
.PHONY: gen


test:
	go clean -testcache
	go test -v -cover ./...
.PHONY: test

test-cover:
	go clean -testcache
	go test -v -coverprofile cover.out ./...
	go tool cover -html=cover.out
.PHONY: test-cover


clean:
	rm -f ${BINARY}
	docker-compose down -v
.PHONY: clean
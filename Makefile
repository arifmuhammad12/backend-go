APP_NAME := backend-go
APP_PORT := 4000

.PHONY: run test build docker-build

run:
	APP_PORT=$(APP_PORT) go run main.go

test:
	go test -v -cover ./...

build:
	CGO_ENABLED=0 go build -ldflags "-w -s" -o bin/$(APP_NAME) main.go

docker-build:
	docker build -t $(APP_NAME):local .

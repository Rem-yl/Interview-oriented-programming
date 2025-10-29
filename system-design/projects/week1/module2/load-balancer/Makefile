APP_NAME=load-balancer
VERSION=$(shell git describe --always --tags)
BUILD_TIME=$(shell date +%Y-%m-%d_%H:%M:%S)
LDFLAGS="-X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'"

.PHONY: build run clean

build:
	go build -ldflags=$(LDFLAGS) -o bin/$(APP_NAME) cmd/lb/main.go 

run: 
	./bin/$(APP_NAME)

clean:
	rm -f bin/$(APP_NAME)
.PHONY: build test

default: build

build:
	go build -v -o ./bin/kamaji-dispatcher ./cmd/kamaji-dispatcher
	go build -v -o ./bin/kamaji-worker ./cmd/kamaji-node
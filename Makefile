build:
	go mod download
	go build -o main

debug: build 
	@mkdir -p mount
	./main mount

run: build 
	@mkdir -p mount
	./main mount

upgrade:
	go mod download
	go get -u -v
	go mod tidy
	go mod verify

test:
	go test

default: build

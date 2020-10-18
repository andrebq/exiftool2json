.PHONY: watch build test default install-modd

default: build

test:
	go test ./...

build:
	go build .

watch:
	modd

install-modd:
	env GO111MODULE=on go get github.com/cortesi/modd/cmd/modd

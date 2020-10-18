.PHONY: watch build test default install-modd docker-build docker-run

IMAGE_TAG?=latest

default: build

test:
	go test ./...

build:
	go build .

watch:
	modd

install-modd:
	env GO111MODULE=on go get github.com/cortesi/modd/cmd/modd


docker-build:
	docker build -t andrebq/exiftool2json:$(IMAGE_TAG) .

docker-run:
	docker run --name exiftool2json -p 8080:8080 --rm andrebq/exiftool2json:$(IMAGE_TAG)

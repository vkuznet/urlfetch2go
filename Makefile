GOPATH:=$(PWD)

all: build install

build:
	go build

install:
	go install

clean:
	go clean

test:
	cd src/urlfetch; go test


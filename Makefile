GOPATH:=$(PWD)

all: build install

build:
	go clean; rm -rf pkg; go build

install:
	go install

clean:
	go clean; rm -rf pkg

test:
	cd src/urlfetch; go test

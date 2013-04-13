default: build test 

build:
	go build

test:	*.go
	cd src/dendrite && go test -v

fmt:
	go fmt .
	cd src/dendrite && go fmt

clean:
	rm -f dendrite
	rm -rf dist
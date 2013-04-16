default: build test 

build:
	go build
	
release: crosscompile
	s3cmd sync dist/. s3://dendrite-binaries
	
crosscompile:
	./build-release.sh

test:	*.go
	go test -v
	cd src/dendrite && go test -v

fmt:
	go fmt .
	cd src/dendrite && go fmt

clean:
	rm -f dendrite
	rm -rf dist
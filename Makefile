default: build test 

build:
	cd cmd/dendrite && go build && cp dendrite ../../
	
release: crosscompile
	s3cmd sync dist/. s3://dendrite-binaries
	
crosscompile:
	./build-release.sh

test:	*.go
	go test -v
	cd cmd/dendrite && go test -v

fmt:
	go fmt .
	cd cmd/dendrite && go fmt

clean:
	rm -f dendrite
	rm -rf dist
default: test

test:	*.go
	cd src/dendrite && go test -v

fmt:
	go fmt .
	cd src/dendrite && go fmt

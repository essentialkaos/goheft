########################################################################################

.PHONY = fmt all clean deps

########################################################################################

all: goheft

goheft:
	go build goheft.go

deps:
	go get -v pkg.re/essentialkaos/ek.v7

fmt:
	find . -name "*.go" -exec gofmt -s -w {} \;

clean:
	rm -f goheft

########################################################################################


.PHONY: test test-race bench lint cover cover-html vet build clean

GO ?= go

build:
	$(GO) build ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

bench:
	$(GO) test -bench=. -benchmem -run=^$$ ./...

lint:
	golangci-lint run

cover:
	$(GO) test -coverprofile=cover.out -covermode=atomic ./...
	$(GO) tool cover -func=cover.out | tail -1

cover-html: cover
	$(GO) tool cover -html=cover.out

clean:
	rm -f cover.out

BINARY := ccopy
PKG    := github.com/abhignan-rakshith/ccopy/cmd/ccopy
BIN    := ./bin/$(BINARY)

.PHONY: build install test run vet fmt clean

build:
	@mkdir -p bin
	go build -o $(BIN) $(PKG)

install:
	go install $(PKG)

test:
	go test ./...

run: build
	$(BIN)

vet:
	go vet ./...

fmt:
	gofmt -w .

clean:
	rm -rf bin

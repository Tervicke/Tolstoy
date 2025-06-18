LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

.PHONY : benchmark
.PHONY : proto
build:
	go build -o Tolstoy
runbroker:
	go run main.go -config broker/config.yaml 

publisher:
	go run examples/publisher/publisher.go -addr "localhost:8080" -topic "mytopic"

subscriber:
	go run examples/subscriber/subscriber.go -addr "localhost:8080" -topic "mytopic"

benchmark:
	go run benchmark/*.go

testbroker:
	go test -v -cover ./broker

inttest:
	go test -v -cover ./test

testall:
	go test -v -cover ./broker

proto:
	protoc --go_out=. internal/proto/packet.proto
	protoc --go_out=. internal/proto/record.proto

VERSION := $(shell git describe --tags --abbrev=0)
TODAY := $(shell date -u "+%d %B %Y")

release:
	@echo "🚀 Releasing Tolstoy Version $(VERSION)..."
	@mkdir -p dist/Tolstoy_$(VERSION)_linux_amd64
	GOOS=linux GOARCH=amd64 go build -ldflags "-X 'main.VERSION=$(VERSION)' -X 'main.DATE=$(TODAY)' " -o dist/Tolstoy_$(VERSION)_linux_amd64/Tolstoy .
	cd dist && zip -r Tolstoy_$(VERSION)_linux_amd64.zip Tolstoy_$(VERSION)_linux_amd64
	@echo "✅ Done: dist/Tolstoy_$(VERSION)_linux_amd64.zip"

clean:
	rm -rf dist/


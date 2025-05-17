.PHONY : benchmark
.PHONY : proto
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
	go test -v -cover ./test

proto:
	protoc --go_out=. proto/packet.proto

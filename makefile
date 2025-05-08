.PHONY : benchmark
runbroker:
	go run $(shell find broker -name '*.go' ! -name '*_test.go') -config broker/config.yaml

publisher:
	go run examples/publisher/publisher.go -addr "localhost:8080" -topic "mytopic"

subscriber:
	go run examples/subscriber/subscriber.go -addr "localhost:8080" -topic "mytopic"

benchmark:
	go run benchmark/*.go

testbroker:
	go test -v ./broker

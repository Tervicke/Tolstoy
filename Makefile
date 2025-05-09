.PHONY : benchmark
runbroker:
	go run main.go -config broker/config.yaml 

publisher:
	go run examples/publisher/publisher.go -addr "localhost:8080" -topic "mytopic"

subscriber:
	go run examples/subscriber/subscriber.go -addr "localhost:8080" -topic "mytopic"

benchmark:
	go run benchmark/*.go

testbroker:
	gotestsum --format=short-verbose -- -cover ./broker

inttest:
	gotestsum --format=short-verbose -- -cover ./test

testall:
	make testbroker
	make inttest

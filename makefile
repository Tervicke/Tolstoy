runbroker:
	go run broker/*.go -config broker/config.yaml

publisher:
	go run examples/publisher/publisher.go -addr "localhost:8080" -topic "mytopic"

subscriber:
	go run examples/subscriber/subscriber.go -addr "localhost:8080" -topic "mytopic"

testlatencypub:
	go run testing/latencytesting/publisher/publisher.go

testlatencysub:
	go run testing/latencytesting/subscriber/subscriber.go

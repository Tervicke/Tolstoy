runbroker:
	go run broker/*.go

publisher:
	go run examples/publisher/publisher.go -addr "localhost:8080" -topic "mytopic"

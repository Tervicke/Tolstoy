runbroker:
	go run broker/*.go

publisher:
	go run publisher.go -addr "localhost:8080" -topic "mytopic"

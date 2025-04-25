runbroker:
	go run broker/*.go

test:
	go run test.go -addr "localhost:8080" -topic "mytopic"

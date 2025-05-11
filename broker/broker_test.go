package broker

import (
	"net"
	"testing"
	"time"
)


func TestStartServer(t *testing.T){
	go StartServer("testdata/correctconfig1.yaml")
	time.Sleep(100 * time.Millisecond)
	conn , err := net.Dial("tcp" , "localhost:8080")
	if err != nil {
		t.Fatal("Error occured when trying to connect to the server")
	}
	conn.Close()
}

func TestHandleConnection(t *testing.T){
}

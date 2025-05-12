package broker

import (
	"bytes"
	"log"
	"net"
	"os"
	"strings"
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

	var packet [2049]byte

	//invalid packet
	packet[0] = 100
	sendTestPacket(t,packet, "Packet recieved by unverified connection" ,false)

	//valid verification packet
	packet[0] = 7
	sendTestPacket(t,packet, "New verification packet recieved",false)
	
	//invalid packet type after verification
	packet[0] = 100
	sendTestPacket(t,packet,"Recieved Invalid packet type",true)

}


func sendTestPacket(t *testing.T , packet [2049]byte , expected string , verified bool) {

	var logbuf bytes.Buffer
	log.SetOutput(&logbuf)
	defer log.SetOutput(os.Stderr)
	serverconn , clientconn := net.Pipe()
	defer serverconn.Close()
	defer clientconn.Close()
	
	//if the user has added verified flag add the connection to mock its verified
	if verified	{
		activeconnmutex.Lock()
		ActiveConnections[serverconn] = struct{}{}
		activeconnmutex.Unlock()
	}

	go handleConnection(serverconn)

	_,err := clientconn.Write(packet[:])

	if err != nil {
		t.Fatalf("failed to write the packet")
	}
	
	time.Sleep(100 * time.Millisecond)
	logOutput := logbuf.String()
	if !strings.Contains(logOutput,expected) {
		t.Errorf("expected string not found in logoutput expected %q found\n %q",expected,logOutput)
	}

}


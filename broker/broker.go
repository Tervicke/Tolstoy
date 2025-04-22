package main

import (
	"fmt"
	"net"
)

const addr string = "localhost:8080";
//make a map of net Conn and struct (because 0 bytes) //simulate a set
var ActiveConnections = make(map[net.Conn]struct{})
var Topics = make(map[string]map[net.Conn]struct{})

func handleConnection(curCon net.Conn){
	defer curCon.Close()
	ActiveConnections[curCon] = struct{}{}
	for {
		buf := make([]byte,2049)
		totalread := 0;
		for totalread < 2049 {
			n , err := curCon.Read(buf[totalread:])
			if err != nil{
				delete(ActiveConnections ,curCon)
				return;
			}
			totalread += n
		}
		newpacket := newPacket([2049]byte(buf),curCon);
		if handlepacket, ok := handlers[newpacket.Type]; ok {
			//todo: implement the Acknowledge packet and send it after handlerfunc returns a no error 
			fmt.Println("packet recieved....handling")
			fmt.Println(newpacket.Payload)
			handlepacket(newpacket)
		}else{
			//specify error code and and send it accordingly 
			fmt.Println("invalid packet type")
		}
	}
}

func main() {
	broker , err := net.Listen("tcp" , addr);
	if err != nil{
		panic("Failed to start the broker")
	}
	for true {
		conn , err := broker.Accept();
		if err != nil {
			fmt.Println("failed to accept a request");
			fmt.Println(err);
		}

		go handleConnection(conn);
	}

}

package main

import (
	"fmt"
	"net"
)

const addr string = "localhost:8080";
//make a map of net Conn and struct (because 0 bytes) //simulate a set
var ActiveConnections = make(map[net.Conn]struct{})

func handleConnection(curCon net.Conn){
	defer curCon.Close()
	fmt.Println("connection accepted")
	ActiveConnections[curCon] = struct{}{}
	fmt.Printf("Total no of clients are now %d\n",len(ActiveConnections));
	for {
		buf := make([]byte,2049)
		totalread := 0;
		for totalread < 2049 {
			n , err := curCon.Read(buf[totalread:])
			if err != nil{
				fmt.Println("A client just disconnected")
				delete(ActiveConnections ,curCon)
				fmt.Printf("Total no of clients are now %d\n",len(ActiveConnections));
				return;
			}
			totalread += n
		}
		newpacket := newPacket([2049]byte(buf),curCon);
		newpacket.Print();
		if handlepacket, ok := handlers[newpacket.Type]; ok {
			//todo: implement the Acknowledge packet and send it after handlerfunc returns a no error 
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

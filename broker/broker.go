package main

import (
	"fmt"
	"net"
	"strings"
)

const addr string = "localhost:8080";
//make a map of net Conn and struct (because 0 bytes) //simulate a set
var ActiveConnections = make(map[net.Conn]struct{})


func handleConnection(curCon net.Conn){
	defer curCon.Close()
	fmt.Println("connection accepted")
	fmt.Printf("Total no of clients are now %d\n",len(ActiveConnections));
	ActiveConnections[curCon] = struct{}{}
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
		packet_type := int(buf[0]);
		fmt.Println("Packet Type: ",packet_type);
		topic := strings.Trim(string(buf[1:1025]),"\x00");
		fmt.Println("Topic:",topic);
		payload := strings.Trim(string(buf[1025:2049]),"\x00");
		fmt.Println("Paylod:",payload);
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

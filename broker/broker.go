package main

import (
	"fmt"
	"net"
	"strings"
)

const addr string = "localhost:8080";
var Activeconnection []net.Conn;

func handleConnection(curCon net.Conn){
	defer curCon.Close()
	Activeconnection = append(Activeconnection,curCon);
	fmt.Printf("connection accepted \n Total no of clients are now %d\n",len(Activeconnection));
	buf :=	make([]byte,2049);
	packet_size,err := curCon.Read(buf);
	if err != nil{
		fmt.Println("Error reading the buffer");
		return;
	}
	if packet_size < 2049{
		fmt.Println("invalid packet recieved");
	}
	packet_type := int(buf[0]);
	fmt.Println("Packet Type: ",packet_type);
	topic := strings.Trim(string(buf[1:1025]),"\x00");
	fmt.Println("Topic:",topic);
	payload := strings.Trim(string(buf[1025:2049]),"\x00");
	fmt.Println("Paylod:",payload);
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

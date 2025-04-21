package main

import (
	"fmt"
	"net"
)

const addr string = "localhost:8080";
var Activeconnection []net.Conn;

func handleConnection(curCon net.Conn){
	Activeconnection = append(Activeconnection,curCon);
	fmt.Printf("connection accepted \n Total no of clients are now %d\n",len(Activeconnection));
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

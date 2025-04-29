package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

const addr string = "localhost:8080";
//make a map of net Conn and struct (because 0 bytes) //simulate a set
var ActiveConnections = make(map[net.Conn]struct{})
var Topics = make(map[string]map[net.Conn]struct{})

func handleConnection(curCon net.Conn){
	fmt.Println("agent joined")
	defer curCon.Close()
	ActiveConnections[curCon] = struct{}{}
	for {
		buf := make([]byte,2049)
		totalread := 0;
		for totalread < 2049 {
			n , err := curCon.Read(buf[totalread:])
			if err != nil{
				fmt.Println("agent left")
				delete(ActiveConnections ,curCon)
				return;
			}
			totalread += n
		}
		newpacket := newPacket([2049]byte(buf),curCon);
		if handlepacket, ok := handlers[newpacket.Type]; ok {
			fmt.Println("packet recieved....acknowledging and handling")
			if handlepacket(newpacket) {
				fmt.Println("handled")
				//newpacket.acknowledge()
				//fmt.Println("acknowledged")
			}else{
				fmt.Println("could not acknowledge , error")
			}
		}else{
			//specify error code and and send it accordingly 
			fmt.Println("invalid packet type")
		}
	}
}

func main() {
	handleCrash()
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

func handleCrash(){

	c := make(chan os.Signal , 1)
	signal.Notify(c,os.Interrupt,syscall.SIGTERM)
	go func(){
		 <-c
		log.Printf("Shutting Down..sending disconnection packets to all the %d agents",len(ActiveConnections))
		dpacket := newDisconnectionPacket()
		for conn := range ActiveConnections{
			conn.Write(dpacket.toBytes())
		}
		log.Printf("Sent disconnection packets to all the %d agents",len(ActiveConnections))
		os.Exit(0)
	}()
}

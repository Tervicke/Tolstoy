package main

import "fmt"

type packetHandler func(packet Packet)

var handlers = map[int]packetHandler{
	2:handleConnectionPacket, //handles the upcoming connections and adds them to the ActiveConnections map
	3:handleDisconnectionPacket,
}

func handleConnectionPacket(packet Packet) {
	fmt.Println("handling connection");
	//to be implemented
}


func handleDisconnectionPacket(packet Packet) { //handles the disconnection request and removes the connection from the handler  
	fmt.Println("handling disconnection"); 
	//to be implemented
}

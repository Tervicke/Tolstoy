package main

import "fmt"

type packetHandler func(packet Packet)

var handlers = map[uint8]packetHandler{
	2:handleConnectionPacket, //handles the upcoming connections and adds them to the ActiveConnections map
	3:handleDisconnectionPacket, //handles the disconnection request
}

func handleConnectionPacket(packet Packet) {
	fmt.Println("handling connection");
}


func handleDisconnectionPacket(packet Packet) { //handles the disconnection request and removes the connection from the handler  
	fmt.Println("handling disconnection"); 
	delete(ActiveConnections , packet.Conn)
}

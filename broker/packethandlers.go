package main

import (
	"fmt"
	"net"
)

type packetHandler func(packet Packet)

var handlers = map[uint8]packetHandler{
	3:handlePublishPacket, //handles the publish packet
	4:handleSubscribePacket, //handles the subscribe packet
	5:handleSubscribeCreatePacket,
}
func handlePublishPacket(packet Packet){
	clients , exists := Topics[packet.Topic]
	if exists {
		for client := range clients{
			packet.Type = 2 // change the packet type to indicating its a server packet
			client.Write( packet.toBytes()  )
		}
		fmt.Println("message published")
	}else{
		fmt.Println("topic not exist")
		fmt.Println("Topic-",packet.Topic)
		// TODO: Implement Error
	}
}

func handleSubscribePacket(packet Packet){
}

func handleSubscribeCreatePacket(packet Packet){
	_ , exists := Topics[packet.Topic]
	if !exists{
		Topics[packet.Topic] = make(map[net.Conn]struct{})
	}
	Topics[packet.Topic][packet.Conn] = struct{}{}
	fmt.Println("subscriber added")
	fmt.Println("Topic-",packet.Topic)
}

func handleConnectionPacket(packet Packet) {
	fmt.Println("handling connection");
}


func handleDisconnectionPacket(packet Packet) {   
	fmt.Println("handling disconnection"); 
	delete(ActiveConnections , packet.Conn)
}

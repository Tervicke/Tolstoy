package main

import (
	"fmt"
	"net"
)

type packetHandler func(packet Packet)

var handlers = map[uint8]packetHandler{
	4:handlePublishPacket, //handles the publish packet
	5:handleSubscribePacket, //handles the subscribe packet
}
func handlePublishPacket(packet Packet){
	clients , exists := Topics[packet.Topic]
	//if the topic doesnt exist create it 
	if !exists{
		Topics[packet.Topic] = make(map[net.Conn]struct{})
	}
	if exists {
		for client := range clients{
			packet.Type = 2 // change the packet type to indicating its a server packet
			client.Write( packet.toBytes()  )
		}
		fmt.Println("message published")
	}
}

func handleSubscribePacket(packet Packet){
	_ , exists := Topics[packet.Topic]
	if !exists{
		//send error that topic doesnt exist
	}
	//add subscriber
	Topics[packet.Topic][packet.Conn] = struct{}{}
	fmt.Println("subscriber added")
	fmt.Println("Topic-",packet.Topic)
}

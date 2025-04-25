package main

import (
	"fmt"
	"net"
)

type packetHandler func(packet Packet) bool

var handlers = map[uint8]packetHandler{
	4:handlePublishPacket, //handles the publish packet
	5:handleSubscribePacket, //handles the subscribe packet
}
func handlePublishPacket(packet Packet) bool{
	clients , exists := Topics[packet.Topic]
	//if the topic doesnt exist create it 
	if !exists{
		Topics[packet.Topic] = make(map[net.Conn]struct{})
	}
	packet.acknowledge(10)
	fmt.Println("acknowledged")
	for client := range clients{
		packet.Type = 3 // change the packet type to indicating its a server sent packet
		client.Write( packet.toBytes()  )
	}
	//publish ack_code = 10
	return true
}

func handleSubscribePacket(packet Packet) bool {
	_ , exists := Topics[packet.Topic]
	if !exists{
		errorpacket := newErrPacket("Topic does not exist");
		packet.Conn.Write(errorpacket[:]);
		return false;
	}
	//add subscriber
	Topics[packet.Topic][packet.Conn] = struct{}{}
	fmt.Println("subscriber added")
	fmt.Println("Topic-",packet.Topic)
	packet.acknowledge(11) //ack code - 11 for successful subscribe
	return true
}

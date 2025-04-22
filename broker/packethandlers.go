package main

import "fmt"

type packetHandler func(packet Packet)

var handlers = map[uint8]packetHandler{
	2:handleConnectionPacket, //handles the upcoming connections and adds them to the ActiveConnections map
	3:handleDisconnectionPacket, //handles the disconnection request
	4:handlePublishPacket, //handles the publish packet
	5: handleSubscribePacket, //handles the subscribe packet
}
func handlePublishPacket(packet Packet){
	clients , exists := Topics[packet.Topic]
	if exists {
		for client := range clients{
			packet.Type = 2 // change the packet type to indicating its a server packet
			client.Write( packet.toBytes()  )
		}
	}else{
		// TODO: Implement Error
	}
}

func handleSubscribePacket(packet Packet){

}

func handleConnectionPacket(packet Packet) {
	fmt.Println("handling connection");
}


func handleDisconnectionPacket(packet Packet) {   
	fmt.Println("handling disconnection"); 
	delete(ActiveConnections , packet.Conn)
}

package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

type packetHandler func(packet Packet) bool

var handlers = map[uint8]packetHandler{
	4:handlePublishPacket, //handles the publish packet
	5:handleSubscribePacket, //handles the subscribe packet
	6:handleUnsubscribePacket,
}
func handleUnsubscribePacket(packet Packet) bool{
	topicConnections , exists := Topics[packet.Topic]
	if !exists{
		return true
	}
	delete(topicConnections , packet.Conn)
	packet.acknowledge(12) //subscribe ack
	return true;//since no error
}

func handlePublishPacket(packet Packet) bool{
	clients , exists := Topics[packet.Topic]
	//if the topic doesnt exist create it 
	if !exists{
		Topics[packet.Topic] = make(map[net.Conn]struct{})
	}
	//write it to the log file 
	WriteMessage(packet.Payload, packet.Topic)
	packet.acknowledge(10)
	log.Println("Acknowledged")
	for client := range clients{
		packet.Type = 3 // change the packet type to indicating its a server sent packet
		packetbytes := packet.toBytes()
		client.Write( packetbytes[:] )
	}
	//publish ack_code = 10
	return true
}

func handleSubscribePacket(packet Packet) bool {
	_ , exists := Topics[packet.Topic]

	if !exists{
		errorpacket := newErrPacket("Topic does not exist");
		log.Printf("Sent Error packet for no topic %v\n",packet.Topic)
		packet.Conn.Write(errorpacket[:]);
		return false;
	}
	//add subscriber
	Topics[packet.Topic][packet.Conn] = struct{}{}
	log.Printf("New subscriber added to %s | count = %d\n",packet.Topic,len(Topics[packet.Topic]))
	packet.acknowledge(11) //ack code - 11 for successful subscribe
	return true
}

func WriteMessage(payload string , topic_name string){
	if !brokerSettings.Persistence.Enabled {
		return
	}
	filename := getFilePath(topic_name)
	file,err := os.OpenFile(filename , os.O_APPEND|os.O_CREATE|os.O_WRONLY , 0644)
	if err != nil{
		log.Printf("Error writing payload to log file %v",err)
	}

	_,err = file.Write([]byte(payload + "\n"))

	if err != nil{
		log.Println("Error writing to a file")
	}

}
func getFilePath(topic_name string) string {

	year , month , day := time.Now().Date()
	date := strconv.Itoa(year) + "-" + month.String() + "-"  +  strconv.Itoa(day);
	return (brokerSettings.Persistence.Directory + date + "-" + topic_name + ".log")
}
